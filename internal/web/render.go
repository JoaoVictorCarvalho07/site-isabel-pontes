package web

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
)

// Renderer executa os templates.
//
// Em prod os templates sao parseados uma vez no boot e vivem em memoria. Com
// --dev eles sao reparseados a cada request, o que substitui o hot reload que o
// Vite dava de graca. Sem isso voce reinicia o servidor a mao a cada `<div>`.
type Renderer struct {
	fsys  fs.FS
	dev   bool
	funcs template.FuncMap
	cache map[string]*template.Template
}

func NewRenderer(fsys fs.FS, dev bool, funcs template.FuncMap) (*Renderer, error) {
	r := &Renderer{fsys: fsys, dev: dev, funcs: funcs, cache: map[string]*template.Template{}}
	if !dev {
		// Falha no boot, nao no primeiro request: um template quebrado tem que
		// impedir o deploy de subir, nao virar 500 na cara do visitante.
		if err := r.warm(); err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (r *Renderer) warm() error {
	pages, err := fs.Glob(r.fsys, "templates/pages/*.gohtml")
	if err != nil {
		return err
	}
	if len(pages) == 0 {
		return fmt.Errorf("nenhum template em templates/pages/ -- o //go:embed pegou?")
	}
	for _, p := range pages {
		if _, err := r.parse(p); err != nil {
			return fmt.Errorf("template %s: %w", p, err)
		}
	}
	return nil
}

func (r *Renderer) parse(page string) (*template.Template, error) {
	if !r.dev {
		if t, ok := r.cache[page]; ok {
			return t, nil
		}
	}
	patterns := []string{"templates/layout.gohtml", "templates/partials/*.gohtml", page}
	t, err := template.New("layout.gohtml").Funcs(r.funcs).ParseFS(r.fsys, patterns...)
	if err != nil {
		return nil, err
	}
	if !r.dev {
		r.cache[page] = t
	}
	return t, nil
}

// Page renderiza uma pagina completa (layout + conteudo).
//
// Renderiza num buffer antes de escrever na resposta. Sem isso, um erro no meio
// do template ja teria mandado 200 e meio HTML para o cliente, e o erro viraria
// lixo no meio da pagina em vez de um 500 honesto. Nao remova.
func (r *Renderer) Page(w http.ResponseWriter, req *http.Request, name string, status int, data any) {
	t, err := r.parse("templates/pages/" + name + ".gohtml")
	if err != nil {
		r.fail(w, name, err)
		return
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "layout.gohtml", data); err != nil {
		r.fail(w, name, err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}

// Fragment renderiza so um partial, sem layout -- as respostas do HTMX.
//
// O par Page/Fragment tem que renderizar exatamente o mesmo partial com o mesmo
// struct. E isso que garante que /galeria?tema=casal colado no WhatsApp mostre o
// mesmo que o clique no filtro. Se divergirem, o SEO quebra em silencio.
func (r *Renderer) Fragment(w http.ResponseWriter, page, partial string, data any) {
	t, err := r.parse("templates/pages/" + page + ".gohtml")
	if err != nil {
		r.fail(w, partial, err)
		return
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, partial, data); err != nil {
		r.fail(w, partial, err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = buf.WriteTo(w)
}

func (r *Renderer) fail(w http.ResponseWriter, name string, err error) {
	// O erro vai para o log, nunca para o cliente: mensagem de template pode
	// vazar caminho de arquivo e estrutura interna.
	logger.Error("falha ao renderizar template", "template", name, "erro", err)
	http.Error(w, "Erro interno", http.StatusInternalServerError)
}
