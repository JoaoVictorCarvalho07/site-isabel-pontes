// Package web monta o servidor HTTP.
package web

import (
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/JoaoVictorCarvalho07/site-da-bebel-htmx/internal/config"
	"github.com/JoaoVictorCarvalho07/site-da-bebel-htmx/internal/store"
)

var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

type Server struct {
	cfg    *config.Config
	render *Renderer
	static fs.FS
	snap   *store.Snapshot // fotos em RAM; carregado no boot
}

// New monta o servidor. `assets` e o FS embutido com templates/ e static/;
// `snap` e o snapshot da galeria carregado do SQLite no boot.
func New(cfg *config.Config, assets fs.FS, snap *store.Snapshot) (*Server, error) {
	hasher := newAssetHasher(assets)
	funcs := template.FuncMap{
		"asset": hasher.url,
		// svg embute um arquivo SVG inline no HTML. Precisa ser inline (nao
		// <img>) para o fill=currentColor herdar a cor do texto -- e o que
		// deixa o logo mudar de cor por pagina e no dark mode sem trocar
		// arquivo. Vem do //go:embed, entao e conteudo nosso, seguro.
		"svg": func(caminho string) (template.HTML, error) {
			b, err := fs.ReadFile(assets, "static/"+caminho)
			if err != nil {
				return "", err
			}
			return template.HTML(b), nil //nolint:gosec // asset proprio, embutido
		},
	}
	r, err := NewRenderer(assets, cfg.Dev, funcs)
	if err != nil {
		return nil, err
	}
	static, err := fs.Sub(assets, "static")
	if err != nil {
		return nil, err
	}
	return &Server{cfg: cfg, render: r, static: static, snap: snap}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// Os padroes com metodo ("GET /sobre") sao do ServeMux do Go 1.22+.
	// E o que torna chi/echo desnecessario num site de 8 rotas.
	mux.HandleFunc("GET /{$}", s.home) // {$} = so a raiz exata, nao um catch-all
	mux.HandleFunc("GET /sobre", s.todo("sobre"))
	mux.HandleFunc("GET /contato", s.todo("contato"))
	mux.HandleFunc("GET /galeria", s.galeria)               // pagina completa
	mux.HandleFunc("GET /galeria/fotos", s.galeriaFragment) // fragmento HTMX
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// max-age longo + immutable so e seguro porque o caminho tem hash de conteudo.
	// Ver AssetURL: trocar o arquivo troca a URL, entao nao ha cache velho.
	mux.Handle("GET /static/", http.StripPrefix("/static/",
		cacheForever(http.FileServer(http.FS(s.static)))))

	// Sem padrao registrado -> 404 de verdade, com status 404.
	// A SPA devolvia 200 em qualquer URL inexistente, o que faz o Google
	// indexar paginas de erro como se fossem conteudo.
	mux.HandleFunc("GET /", s.notFound)

	return recoverer(requestLog(gzipMW(mux)))
}

// pageData e o ViewModel base. Sempre struct, nunca map[string]any: um campo
// inexistente num struct e erro de execucao que o teste pega; num map ele
// renderiza "<no value>" em silencio, direto para o visitante. Essa e a unica
// defesa real contra a falta de type safety do html/template. Nao abra excecao.
type pageData struct {
	Titulo    string
	Descricao string
	Canonical string
	BodyClass string
	Ano       int
}

// novaPagina preenche o que vale para toda pagina.
func (s *Server) novaPagina(titulo, descricao, path string) pageData {
	return pageData{
		Titulo:    titulo,
		Descricao: descricao,
		Canonical: s.cfg.BaseURL + path,
		// O React tinha "© 2026" cravado no JSX, que envelhece sozinho em
		// janeiro. Aqui o rodape nunca fica errado.
		Ano: time.Now().Year(),
	}
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	d := s.novaPagina(
		"Isabel Pontes — Fotografia artística e direção criativa em Curitiba",
		"Ensaios conceituais, fotografia artística e direção criativa em Curitiba.",
		"/",
	)
	// O hero e escuro, entao a logo tem que ser a branca nos dois temas.
	// No React isto era useState + matchMedia + dois useEffect, e causava um
	// flash de logo errada ate o React hidratar. Aqui o servidor ja decide.
	d.BodyClass = "hero-escuro"
	s.render.Page(w, r, "home", http.StatusOK, d)
}

func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	s.render.Page(w, r, "404", http.StatusNotFound,
		s.novaPagina("Página não encontrada — Isabel Pontes", "", r.URL.Path))
}

// todo é um placeholder honesto: a rota existe e responde 501, em vez de 404
// (que mentiria dizendo que a página não vai existir) ou de uma página vazia
// (que fingiria estar pronta). Sai conforme cada fase entrega a página real.
func (s *Server) todo(nome string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusNotImplemented)
		_, _ = w.Write([]byte("501 — /" + nome + " ainda não foi portada.\n"))
	}
}
