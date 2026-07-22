package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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

// Arquivo e uma imagem de referencia anexada a um briefing.
type Arquivo struct {
	ID           int64
	SubmissionID int64
	Campo        string
	Caminho      string // relativo ao dir de uploads; NUNCA veio do visitante
	NomeOriginal string
	Mime         string
	Tamanho      int64
}

// SalvarArquivo registra um anexo ja gravado em disco.
func (s *Store) SalvarArquivo(ctx context.Context, a Arquivo) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO submission_files (submission_id, campo, caminho, nome_original, mime, tamanho)
		 VALUES (?,?,?,?,?,?)`,
		a.SubmissionID, a.Campo, a.Caminho, a.NomeOriginal, a.Mime, a.Tamanho)
	return err
}

// ArquivosDe devolve os anexos de varios briefings de uma vez, indexados pelo
// id da submission. Recebe a lista inteira para o painel nao disparar uma query
// por card (N+1) ao listar 200 briefings.
func (s *Store) ArquivosDe(ctx context.Context, ids []int64) (map[int64][]Arquivo, error) {
	porSubmission := map[int64][]Arquivo{}
	if len(ids) == 0 {
		return porSubmission, nil
	}
	// IN (?,?,?) montado pela QUANTIDADE de ids; os valores seguem como
	// parametros. A string SQL nunca recebe dado, so placeholders.
	marcadores := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	linhas, err := s.db.QueryContext(ctx,
		`SELECT id, submission_id, campo, caminho, nome_original, mime, tamanho
		 FROM submission_files WHERE submission_id IN (`+marcadores+`) ORDER BY id`, args...)
	if err != nil {
		return nil, err
	}
	defer linhas.Close()

	for linhas.Next() {
		var a Arquivo
		if err := linhas.Scan(&a.ID, &a.SubmissionID, &a.Campo, &a.Caminho,
			&a.NomeOriginal, &a.Mime, &a.Tamanho); err != nil {
			return nil, err
		}
		porSubmission[a.SubmissionID] = append(porSubmission[a.SubmissionID], a)
	}
	return porSubmission, linhas.Err()
}

// Arquivo busca um anexo por id -- usado pela rota autenticada que serve a
// imagem para a Isa.
func (s *Store) Arquivo(ctx context.Context, id int64) (Arquivo, error) {
	var a Arquivo
	err := s.db.QueryRowContext(ctx,
		`SELECT id, submission_id, campo, caminho, nome_original, mime, tamanho
		 FROM submission_files WHERE id = ?`, id).
		Scan(&a.ID, &a.SubmissionID, &a.Campo, &a.Caminho, &a.NomeOriginal, &a.Mime, &a.Tamanho)
	return a, err
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
