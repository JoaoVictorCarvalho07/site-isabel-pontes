package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Persistencia dos formularios de briefing editaveis (docs §8b).
//
// O store guarda um shape NEUTRO (strings pt/en soltas, opcoes como JSON): ele
// nao conhece o tipo `ml` bilingue nem `tipoCampo` do pacote web -- esses sao
// conceitos de apresentacao. O web mapeia FormDef <-> seus proprios tipos.

// FormDef e um formulario como vive no banco.
type FormDef struct {
	Slug        string
	TituloPT    string
	TituloEN    string
	SubtituloPT string
	SubtituloEN string
	EnviarPT    string
	EnviarEN    string
	SucessoPT   string
	SucessoEN   string
	Visivel     bool
	Ordem       int
	Perguntas   []QuestionDef
}

// QuestionDef e uma pergunta como vive no banco. OpcoesJSON e a lista de opcoes
// do select serializada (`[{"pt":"","en":""}]`), vazia para os demais tipos.
type QuestionDef struct {
	ID          int64
	Chave       string
	LabelPT     string
	LabelEN     string
	AjudaPT     string
	AjudaEN     string
	Tipo        string
	Obrigatoria bool
	OpcoesJSON  string
	Ordem       int
	Oculta      bool
	ENStale     bool
}

// LoadForms le todos os formularios e suas perguntas, montados. Ordena por
// sort_order para o editor e o render respeitarem a ordem escolhida pela Isa.
func (s *Store) LoadForms(ctx context.Context) ([]FormDef, error) {
	linhas, err := s.db.QueryContext(ctx, `
		SELECT slug, titulo_pt, titulo_en, subtitulo_pt, subtitulo_en,
		       enviar_pt, enviar_en, sucesso_pt, sucesso_en, visivel, sort_order
		FROM forms ORDER BY sort_order, slug`)
	if err != nil {
		return nil, fmt.Errorf("consultar forms: %w", err)
	}
	defer linhas.Close()

	var forms []FormDef
	indice := map[string]int{} // slug -> posicao em forms, para casar as perguntas
	for linhas.Next() {
		var f FormDef
		if err := linhas.Scan(&f.Slug, &f.TituloPT, &f.TituloEN, &f.SubtituloPT, &f.SubtituloEN,
			&f.EnviarPT, &f.EnviarEN, &f.SucessoPT, &f.SucessoEN, &f.Visivel, &f.Ordem); err != nil {
			return nil, err
		}
		indice[f.Slug] = len(forms)
		forms = append(forms, f)
	}
	if err := linhas.Err(); err != nil {
		return nil, err
	}

	qLinhas, err := s.db.QueryContext(ctx, `
		SELECT form_slug, id, chave, label_pt, label_en, ajuda_pt, ajuda_en,
		       tipo, obrigatoria, opcoes_json, sort_order, oculta, en_stale
		FROM form_questions ORDER BY form_slug, sort_order, id`)
	if err != nil {
		return nil, fmt.Errorf("consultar form_questions: %w", err)
	}
	defer qLinhas.Close()

	for qLinhas.Next() {
		var slug string
		var q QuestionDef
		if err := qLinhas.Scan(&slug, &q.ID, &q.Chave, &q.LabelPT, &q.LabelEN,
			&q.AjudaPT, &q.AjudaEN, &q.Tipo, &q.Obrigatoria, &q.OpcoesJSON,
			&q.Ordem, &q.Oculta, &q.ENStale); err != nil {
			return nil, err
		}
		if pos, ok := indice[slug]; ok {
			forms[pos].Perguntas = append(forms[pos].Perguntas, q)
		}
		// pergunta com form_slug orfao (nao deveria existir por causa da FK) e
		// simplesmente ignorada -- nunca derruba o load.
	}
	return forms, qLinhas.Err()
}

// LoadForm carrega UM formulario com todas as perguntas (incl. ocultas), para o
// editor -- que precisa dos ids e das flags que a view publica descarta.
func (s *Store) LoadForm(ctx context.Context, slug string) (FormDef, bool, error) {
	todos, err := s.LoadForms(ctx)
	if err != nil {
		return FormDef{}, false, err
	}
	for _, f := range todos {
		if f.Slug == slug {
			return f, true, nil
		}
	}
	return FormDef{}, false, nil
}

// SeedFormsIfEmpty popula as tabelas a partir dos defaults SO se ainda nao houver
// nenhum formulario. E como os structs Go de hoje viram a linha 1 sem nada se
// perder. Idempotente: rodar de novo com dados existentes nao faz nada.
func (s *Store) SeedFormsIfEmpty(ctx context.Context, defs []FormDef) error {
	var n int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM forms").Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }() // no-op apos Commit

	for _, f := range defs {
		if err := inserirForm(ctx, tx, f); err != nil {
			return fmt.Errorf("seed do form %q: %w", f.Slug, err)
		}
	}
	return tx.Commit()
}

func inserirForm(ctx context.Context, tx *sql.Tx, f FormDef) error {
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO forms (slug, titulo_pt, titulo_en, subtitulo_pt, subtitulo_en,
		                   enviar_pt, enviar_en, sucesso_pt, sucesso_en, visivel, sort_order)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		f.Slug, f.TituloPT, f.TituloEN, f.SubtituloPT, f.SubtituloEN,
		f.EnviarPT, f.EnviarEN, f.SucessoPT, f.SucessoEN, boolInt(f.Visivel), f.Ordem); err != nil {
		return err
	}
	for i, q := range f.Perguntas {
		ordem := q.Ordem
		if ordem == 0 {
			ordem = (i + 1) * 10 // 10,20,30... deixa espaco para inserir no meio depois
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO form_questions (form_slug, chave, label_pt, label_en, ajuda_pt, ajuda_en,
			                            tipo, obrigatoria, opcoes_json, sort_order, oculta, en_stale)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
			f.Slug, q.Chave, q.LabelPT, q.LabelEN, q.AjudaPT, q.AjudaEN,
			q.Tipo, boolInt(q.Obrigatoria), q.OpcoesJSON, ordem, boolInt(q.Oculta), boolInt(q.ENStale)); err != nil {
			return err
		}
	}
	return nil
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// --- Escrita (editor, Fase B) ------------------------------------------------

// AtualizarForm salva os textos de nivel de formulario. Recebe o FormDef inteiro
// (o chamador preserva o EN e troca so o PT que a Isa editou), entao todas as
// colunas de texto sao reescritas de uma vez.
func (s *Store) AtualizarForm(ctx context.Context, f FormDef) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE forms SET titulo_pt=?, titulo_en=?, subtitulo_pt=?, subtitulo_en=?,
		                 enviar_pt=?, enviar_en=?, sucesso_pt=?, sucesso_en=?,
		                 updated_at=datetime('now')
		WHERE slug=?`,
		f.TituloPT, f.TituloEN, f.SubtituloPT, f.SubtituloEN,
		f.EnviarPT, f.EnviarEN, f.SucessoPT, f.SucessoEN, f.Slug)
	if err != nil {
		return err
	}
	return exigeUmaLinha(res, "form", f.Slug)
}

// AtualizarPergunta salva os campos EDITAVEIS de uma pergunta na Fase B: textos,
// obrigatoriedade, opcoes e a flag en_stale. NAO toca `tipo` nem `chave` -- trocar
// tipo e Fase C (pode orfanar resposta) e a chave e imutavel por design.
func (s *Store) AtualizarPergunta(ctx context.Context, q QuestionDef) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE form_questions SET label_pt=?, label_en=?, ajuda_pt=?, ajuda_en=?,
		                          obrigatoria=?, opcoes_json=?, en_stale=?,
		                          updated_at=datetime('now')
		WHERE id=?`,
		q.LabelPT, q.LabelEN, q.AjudaPT, q.AjudaEN,
		boolInt(q.Obrigatoria), q.OpcoesJSON, boolInt(q.ENStale), q.ID)
	if err != nil {
		return err
	}
	return exigeUmaLinha(res, "pergunta", fmt.Sprint(q.ID))
}

// AdicionarPergunta insere uma pergunta nova no fim do formulario. A `chave` e
// gerada por nos (opaca, estavel) -- nunca vem da Isa. Devolve o id novo.
func (s *Store) AdicionarPergunta(ctx context.Context, slug string, q QuestionDef) (int64, error) {
	var ordem int
	// MAX+10 poe a nova no fim; COALESCE cobre o form sem perguntas.
	if err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(sort_order),0)+10 FROM form_questions WHERE form_slug=?`, slug).
		Scan(&ordem); err != nil {
		return 0, err
	}

	// Tenta algumas vezes por causa da colisao (astronomicamente improvavel) com
	// a UNIQUE(form_slug, chave).
	for tentativa := 0; tentativa < 5; tentativa++ {
		chave, err := gerarChave()
		if err != nil {
			return 0, err
		}
		res, err := s.db.ExecContext(ctx, `
			INSERT INTO form_questions (form_slug, chave, label_pt, label_en, ajuda_pt, ajuda_en,
			                            tipo, obrigatoria, opcoes_json, sort_order, en_stale)
			VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
			slug, chave, q.LabelPT, q.LabelEN, q.AjudaPT, q.AjudaEN,
			q.Tipo, boolInt(q.Obrigatoria), q.OpcoesJSON, ordem, boolInt(q.ENStale))
		if err == nil {
			return res.LastInsertId()
		}
		if !ehConflitoUnique(err) {
			return 0, err
		}
	}
	return 0, errors.New("nao consegui gerar uma chave unica apos varias tentativas")
}

// MoverPergunta troca a ordem de uma pergunta com a vizinha (cima/baixo), dentro
// do mesmo formulario. Swap de sort_order em transacao: as duas mudam ou nenhuma.
func (s *Store) MoverPergunta(ctx context.Context, slug string, id int64, paraCima bool) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Ordem atual das perguntas do form (id, sort_order).
	linhas, err := tx.QueryContext(ctx,
		`SELECT id, sort_order FROM form_questions WHERE form_slug=? ORDER BY sort_order, id`, slug)
	if err != nil {
		return err
	}
	type item struct {
		id    int64
		ordem int
	}
	var itens []item
	for linhas.Next() {
		var it item
		if err := linhas.Scan(&it.id, &it.ordem); err != nil {
			linhas.Close()
			return err
		}
		itens = append(itens, it)
	}
	linhas.Close()
	if err := linhas.Err(); err != nil {
		return err
	}

	// Acha a posicao e a vizinha na direcao pedida.
	pos := -1
	for i, it := range itens {
		if it.id == id {
			pos = i
			break
		}
	}
	if pos == -1 {
		return fmt.Errorf("pergunta %d nao esta no form %q", id, slug)
	}
	vizinho := pos + 1
	if paraCima {
		vizinho = pos - 1
	}
	if vizinho < 0 || vizinho >= len(itens) {
		return nil // ja esta na ponta; mover e no-op, nao erro
	}

	// Troca os sort_order.
	a, b := itens[pos], itens[vizinho]
	if _, err := tx.ExecContext(ctx, `UPDATE form_questions SET sort_order=? WHERE id=?`, b.ordem, a.id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE form_questions SET sort_order=? WHERE id=?`, a.ordem, b.id); err != nil {
		return err
	}
	return tx.Commit()
}

// --- Operacoes destrutivas (editor, Fase C) ----------------------------------

// ContarRespostasPorChave conta, por chave de pergunta, quantas submissions do
// formulario tem resposta NAO-vazia para ela. E o numero que o editor mostra
// antes de deixar apagar/trocar tipo -- "essa pergunta ja tem 3 respostas".
//
// Uma passada so pelos payloads do form, contada em Go: o volume e baixo e evita
// um json_extract por pergunta.
func (s *Store) ContarRespostasPorChave(ctx context.Context, slug string) (map[string]int, error) {
	linhas, err := s.db.QueryContext(ctx, `SELECT payload FROM submissions WHERE form=?`, slug)
	if err != nil {
		return nil, err
	}
	defer linhas.Close()

	contagem := map[string]int{}
	for linhas.Next() {
		var payload string
		if err := linhas.Scan(&payload); err != nil {
			return nil, err
		}
		var respostas map[string]string
		if err := json.Unmarshal([]byte(payload), &respostas); err != nil {
			continue // payload corrompido nao derruba a contagem
		}
		for chave, valor := range respostas {
			if valor != "" {
				contagem[chave]++
			}
		}
	}
	return contagem, linhas.Err()
}

// OcultarPergunta liga/desliga o soft-delete. Some do formulario mas continua
// existindo para rotular respostas antigas. Reversivel.
func (s *Store) OcultarPergunta(ctx context.Context, id int64, oculta bool) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE form_questions SET oculta=?, updated_at=datetime('now') WHERE id=?`, boolInt(oculta), id)
	if err != nil {
		return err
	}
	return exigeUmaLinha(res, "pergunta", fmt.Sprint(id))
}

// ApagarPergunta remove a pergunta de vez. O chamador (handler) SO chama isto
// quando a pergunta nao tem respostas -- apagar uma com respostas orfanaria os
// briefings antigos, e para esse caso existe OcultarPergunta.
func (s *Store) ApagarPergunta(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM form_questions WHERE id=?`, id)
	if err != nil {
		return err
	}
	return exigeUmaLinha(res, "pergunta", fmt.Sprint(id))
}

// TrocarTipoPergunta muda o tipo (e as opcoes, quando o novo tipo e lista). Marca
// en_stale porque as opcoes novas nao tem traducao. As respostas ja gravadas
// mantem o texto que era -- o painel as mostra como estao.
func (s *Store) TrocarTipoPergunta(ctx context.Context, id int64, tipo, opcoesJSON string) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE form_questions SET tipo=?, opcoes_json=?, en_stale=1, updated_at=datetime('now') WHERE id=?`,
		tipo, opcoesJSON, id)
	if err != nil {
		return err
	}
	return exigeUmaLinha(res, "pergunta", fmt.Sprint(id))
}

// gerarChave produz um identificador opaco e estavel para uma pergunta nova.
func gerarChave() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "q_" + hex.EncodeToString(b), nil
}

func exigeUmaLinha(res sql.Result, tipo, id string) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("%s %q nao encontrado", tipo, id)
	}
	return nil
}

// ehConflitoUnique detecta a violacao de UNIQUE do SQLite sem depender do driver.
func ehConflitoUnique(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE")
}
