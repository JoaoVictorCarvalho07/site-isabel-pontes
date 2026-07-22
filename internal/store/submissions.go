package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Submission e uma resposta de formulario de briefing ja materializada.
//
// Respostas e um map porque as perguntas sao editaveis: ver o comentario da
// tabela em schema.sql. O painel itera sobre o formulario ATUAL e busca a chave
// aqui, entao pergunta removida some da tela mas continua no banco -- nada se
// perde, e nenhuma migracao e necessaria para mexer no questionario.
type Submission struct {
	ID        int64
	Form      string
	Lang      string
	Respostas map[string]string
	Lida      bool
	CriadoEm  time.Time
}

// SalvarSubmission grava uma resposta. Sincrono de proposito: se falhar, o
// visitante precisa ver o erro e poder tentar de novo -- o oposto de enfileirar
// e perder em silencio.
func (s *Store) SalvarSubmission(ctx context.Context, form, lang string, respostas map[string]string, ip, ua string) (int64, error) {
	payload, err := json.Marshal(respostas)
	if err != nil {
		return 0, fmt.Errorf("serializar respostas: %w", err)
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO submissions (form, lang, payload, ip, user_agent) VALUES (?,?,?,?,?)`,
		form, lang, string(payload), ip, ua)
	if err != nil {
		return 0, fmt.Errorf("inserir submission: %w", err)
	}
	return res.LastInsertId()
}

// ListarSubmissions devolve as mais recentes primeiro, para o painel /admin.
// `form` vazio traz os dois formularios.
func (s *Store) ListarSubmissions(ctx context.Context, form string, limite int) ([]Submission, error) {
	if limite <= 0 {
		limite = 100
	}
	// Duas queries em vez de um WHERE com OR: o filtro e opcional e montar SQL
	// por concatenacao para isso convida injecao sem ganhar nada.
	consulta := `SELECT id, form, lang, payload, lida, created_at FROM submissions
	             ORDER BY created_at DESC, id DESC LIMIT ?`
	args := []any{limite}
	if form != "" {
		consulta = `SELECT id, form, lang, payload, lida, created_at FROM submissions
		            WHERE form = ? ORDER BY created_at DESC, id DESC LIMIT ?`
		args = []any{form, limite}
	}

	linhas, err := s.db.QueryContext(ctx, consulta, args...)
	if err != nil {
		return nil, err
	}
	defer linhas.Close()

	var lista []Submission
	for linhas.Next() {
		var sub Submission
		var payload, criadoEm string
		if err := linhas.Scan(&sub.ID, &sub.Form, &sub.Lang, &payload, &sub.Lida, &criadoEm); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(payload), &sub.Respostas); err != nil {
			// Uma linha corrompida nao pode derrubar o painel inteiro: a Isa
			// precisa conseguir ler as outras.
			sub.Respostas = map[string]string{"_erro": "payload ilegivel: " + payload}
		}
		// O SQLite guarda datetime('now') em UTC, sem timezone no texto.
		if t, err := time.Parse("2006-01-02 15:04:05", criadoEm); err == nil {
			sub.CriadoEm = t.UTC()
		}
		lista = append(lista, sub)
	}
	return lista, linhas.Err()
}

// MarcarLida registra que a Isa ja tratou aquele briefing.
func (s *Store) MarcarLida(ctx context.Context, id int64, lida bool) error {
	_, err := s.db.ExecContext(ctx, `UPDATE submissions SET lida = ? WHERE id = ?`, lida, id)
	return err
}

// ContarNaoLidas alimenta o badge do painel.
func (s *Store) ContarNaoLidas(ctx context.Context) int {
	var n int
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM submissions WHERE lida = 0`).Scan(&n)
	return n
}
