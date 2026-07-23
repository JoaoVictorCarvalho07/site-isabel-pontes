package store

import (
	"context"
	"database/sql"
	"fmt"
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
