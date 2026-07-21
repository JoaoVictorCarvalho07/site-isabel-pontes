// Package web monta o servidor HTTP.
package web

import (
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	// boot muda a cada restart (= cada deploy). Entra no ETag da galeria para
	// que uma troca de template/CSS/JS invalide o cache -- o data_version
	// sozinho so cobre mudanca de DADOS, nao de codigo.
	boot string
}

// New monta o servidor. `assets` e o FS embutido com templates/ e static/;
// `snap` e o snapshot da galeria carregado do SQLite no boot.
func New(cfg *config.Config, assets fs.FS, snap *store.Snapshot) (*Server, error) {
	hasher := newAssetHasher(assets)
	funcs := template.FuncMap{
		"asset": hasher.url,
		// srcset monta um srcset com hash de conteudo em cada candidato, para o
		// Cache-Control immutable de /static/ nao prender uma imagem trocada.
		//   {{ srcset "/static/img/hero/hero-land" "avif" "960 1280 1920" }}
		"srcset": func(base, ext, larguras string) string {
			var sb strings.Builder
			for i, largura := range strings.Fields(larguras) {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(hasher.url(base + "-" + largura + "." + ext))
				sb.WriteString(" " + largura + "w")
			}
			return sb.String()
		},
		// svg embute um arquivo SVG inline no HTML. Precisa ser inline (nao
		// <img>) para o fill=currentColor herdar a cor do texto -- e o que
		// deixa o logo mudar de cor por pagina e no dark mode sem trocar
		// arquivo. Vem do //go:embed, entao e conteudo nosso, seguro.
		"svg": func(caminho string) (template.HTML, error) {
			conteudo, err := fs.ReadFile(assets, "static/"+caminho)
			if err != nil {
				return "", err
			}
			return template.HTML(conteudo), nil //nolint:gosec // asset proprio, embutido
		},
		// dict monta um map inline para passar varios campos a um partial (o Go
		// template so passa um valor). Usado por {{ template "picture" (dict ...) }}.
		"dict": func(paresChaveValor ...any) (map[string]any, error) {
			if len(paresChaveValor)%2 != 0 {
				return nil, fmt.Errorf("dict: numero impar de argumentos")
			}
			resultado := make(map[string]any, len(paresChaveValor)/2)
			for i := 0; i < len(paresChaveValor); i += 2 {
				chave, ok := paresChaveValor[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict: chave %v nao e string", paresChaveValor[i])
				}
				resultado[chave] = paresChaveValor[i+1]
			}
			return resultado, nil
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
	return &Server{
		cfg: cfg, render: r, static: static, snap: snap,
		boot: strconv.FormatInt(time.Now().Unix(), 36),
	}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// Os padroes com metodo ("GET /sobre") sao do ServeMux do Go 1.22+.
	// E o que torna chi/echo desnecessario num site de 8 rotas.
	mux.HandleFunc("GET /{$}", s.home) // {$} = so a raiz exata, nao um catch-all
	mux.HandleFunc("GET /sobre", s.sobre)
	mux.HandleFunc("GET /contato", s.contato)
	mux.HandleFunc("GET /midiakit", s.midiakit)             // one-page de pitch UGC/IGC
	mux.HandleFunc("GET /galeria", s.galeria)               // pagina completa
	mux.HandleFunc("GET /galeria/fotos", s.galeriaFragment) // fragmento HTMX
	mux.HandleFunc("GET /robots.txt", s.robots)
	mux.HandleFunc("GET /sitemap.xml", s.sitemap)
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
	// SEO: URLs ABSOLUTAS. O bot do WhatsApp/Instagram que gera o card de preview
	// nao resolve caminho relativo -- e a razao de todo o porte.
	BaseURL string
	OGImage string
	OGType  string
}

// novaPagina preenche o que vale para toda pagina.
func (s *Server) novaPagina(titulo, descricao, path string) pageData {
	return pageData{
		Titulo:    titulo,
		Descricao: descricao,
		Canonical: s.cfg.BaseURL + path,
		BaseURL:   s.cfg.BaseURL,
		OGImage:   s.cfg.BaseURL + "/static/img/og-image.jpg",
		OGType:    "website",
		// O React tinha "© 2026" cravado no JSX, que envelhece sozinho em
		// janeiro. Aqui o rodape nunca fica errado.
		Ano: time.Now().Year(),
	}
}

func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	s.render.Page(w, r, "404", http.StatusNotFound,
		s.novaPagina("Página não encontrada — Isabel Pontes", "", r.URL.Path))
}
