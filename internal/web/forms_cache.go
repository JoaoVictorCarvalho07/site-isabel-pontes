package web

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/JoaoVictorCarvalho07/site-da-bebel-htmx/internal/store"
)

// formsCache e a copia em RAM dos formularios de briefing. Espelha o Snapshot da
// galeria: o boot carrega do SQLite, os requests leem daqui, e uma edicao no
// /admin chama reload(). O caminho quente (visitante preenchendo) nunca toca o
// banco. Ver docs §8b.
type formsCache struct {
	mu    sync.RWMutex
	banco *store.Store

	// publico: forms VISIVEIS, so perguntas nao-ocultas -- e o que o visitante ve.
	publico map[string]*formBriefing
	// completo: TODAS as perguntas de TODOS os forms (incl. ocultas e forms
	// invisiveis). O painel usa para rotular respostas antigas de perguntas que
	// a Isa escondeu -- sem isso, esconder uma pergunta sumiria com as respostas
	// dela da tela.
	completo map[string][]pergunta
	// ordem estavel dos slugs, para listagens deterministicas.
	slugs []string
}

// novoFormsCache semeia (se vazio) e carrega. `banco` nil deixa o cache vazio --
// os handlers de briefing entao respondem 404, sem panico (usado por testes que
// nao exercitam formulario).
func novoFormsCache(banco *store.Store) (*formsCache, error) {
	c := &formsCache{banco: banco}
	if banco == nil {
		c.publico = map[string]*formBriefing{}
		c.completo = map[string][]pergunta{}
		return c, nil
	}
	if err := banco.SeedFormsIfEmpty(context.Background(), seedDefs()); err != nil {
		return nil, err
	}
	if err := c.reload(); err != nil {
		return nil, err
	}
	return c, nil
}

// reload relê tudo do banco e troca o conteudo do cache de uma vez. Chamado no
// boot e apos cada edicao do painel (Fase B+).
func (c *formsCache) reload() error {
	defs, err := c.banco.LoadForms(context.Background())
	if err != nil {
		return err
	}
	publico := make(map[string]*formBriefing, len(defs))
	completo := make(map[string][]pergunta, len(defs))
	slugs := make([]string, 0, len(defs))

	for _, d := range defs {
		slugs = append(slugs, d.Slug)

		todas := make([]pergunta, 0, len(d.Perguntas))
		visiveis := make([]pergunta, 0, len(d.Perguntas))
		for _, q := range d.Perguntas {
			p := perguntaFromDef(q)
			todas = append(todas, p)
			if !q.Oculta {
				visiveis = append(visiveis, p)
			}
		}
		completo[d.Slug] = todas

		if d.Visivel {
			f := formFromDef(d)
			f.Perguntas = visiveis
			publico[d.Slug] = f
		}
	}

	c.mu.Lock()
	c.publico, c.completo, c.slugs = publico, completo, slugs
	c.mu.Unlock()
	return nil
}

// get devolve o formulario publico de um slug (nil, false se invisivel/inexistente).
func (c *formsCache) get(slug string) (*formBriefing, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	f, ok := c.publico[slug]
	return f, ok
}

// existe diz se o slug corresponde a algum form conhecido (visivel ou nao).
func (c *formsCache) existe(slug string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.completo[slug]
	return ok
}

// perguntasDe devolve TODAS as perguntas de um form (incl. ocultas), para o
// painel casar respostas antigas com seus rotulos.
func (c *formsCache) perguntasDe(slug string) ([]pergunta, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	ps, ok := c.completo[slug]
	return ps, ok
}

// --- Mapeamento store.FormDef <-> tipos do web -------------------------------

func formFromDef(d store.FormDef) *formBriefing {
	return &formBriefing{
		Slug:      d.Slug,
		Titulo:    ml{d.TituloPT, d.TituloEN},
		Subtitulo: ml{d.SubtituloPT, d.SubtituloEN},
		Enviar:    ml{d.EnviarPT, d.EnviarEN},
		Sucesso:   ml{d.SucessoPT, d.SucessoEN},
		// Perguntas preenchidas pelo chamador (publico usa so as visiveis).
	}
}

func perguntaFromDef(q store.QuestionDef) pergunta {
	return pergunta{
		Chave:       q.Chave,
		Label:       ml{q.LabelPT, q.LabelEN},
		Ajuda:       ml{q.AjudaPT, q.AjudaEN},
		Tipo:        tipoCampo(q.Tipo),
		Obrigatoria: q.Obrigatoria,
		Opcoes:      opcoesFromJSON(q.OpcoesJSON),
	}
}

// opcaoJSON e o shape exportado que atravessa o JSON -- `ml` tem campos
// minusculos e o encoding/json nao os alcanca.
type opcaoJSON struct {
	PT string `json:"pt"`
	EN string `json:"en"`
}

func opcoesFromJSON(s string) []ml {
	if s == "" {
		return nil
	}
	var brutas []opcaoJSON
	if err := json.Unmarshal([]byte(s), &brutas); err != nil {
		logger.Error("forms: opcoes_json ilegivel", "erro", err, "json", s)
		return nil
	}
	opcoes := make([]ml, len(brutas))
	for i, o := range brutas {
		opcoes[i] = ml{o.PT, o.EN}
	}
	return opcoes
}

func opcoesToJSON(opcoes []ml) string {
	if len(opcoes) == 0 {
		return ""
	}
	brutas := make([]opcaoJSON, len(opcoes))
	for i, o := range opcoes {
		brutas[i] = opcaoJSON{o.pt, o.en}
	}
	b, err := json.Marshal(brutas)
	if err != nil {
		return ""
	}
	return string(b)
}
