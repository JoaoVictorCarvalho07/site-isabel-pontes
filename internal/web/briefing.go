package web

import (
	"errors"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Formularios de briefing. Ver docs/plano-juncao-midiakit.md §8.
//
// O fluxo e progressive enhancement de verdade: o <form> tem action/method
// reais, entao ele FUNCIONA sem JavaScript (POST normal -> pagina de obrigado).
// Com o HTMX carregado, o mesmo POST volta como fragmento e substitui o form no
// lugar. Um so handler serve os dois casos -- e a mesma ideia do filtro da
// galeria e do portfolio de videos.

// campoHoneypot e um input escondido via CSS. Humano nao ve, entao nao preenche;
// bot que preenche formulario por heuristica de name= cai nele. Custa uma linha
// e pega a esmagadora maioria do spam, sem CAPTCHA de terceiro (que quebraria a
// promessa de "nenhum JS de terceiro" e a acessibilidade).
const campoHoneypot = "website"

// briefingView e o ViewModel da pagina /briefing e do partial.
type briefingView struct {
	pageData
	Form *formBriefing
	// Erros: chave da pergunta -> mensagem. Vazio no GET.
	Erros map[string]string
	// Valores digitados, devolvidos no erro para o visitante nao reescrever tudo.
	Valores map[string]string
	// Enviado liga a tela de sucesso no lugar do formulario.
	Enviado bool
	// Action e o destino do POST. Fica no ViewModel (e nao fixo no template)
	// porque o mesmo partial e usado em /briefing e embutido no /midiakit.
	Action string
}

// --- Rate limit --------------------------------------------------------------

// limitador guarda o instante do ultimo envio por IP. Um map com mutex basta:
// sao alguns envios por semana, e um limitador em memoria perde o estado no
// restart -- o que aqui e aceitavel (o pior caso e um bot conseguir um envio a
// mais logo apos um deploy).
type limitador struct {
	mu      sync.Mutex
	ultimo  map[string]time.Time
	janela  time.Duration
	limpeza time.Time
}

func novoLimitador(janela time.Duration) *limitador {
	return &limitador{ultimo: map[string]time.Time{}, janela: janela, limpeza: time.Now()}
}

// permitir devolve false se o IP enviou algo dentro da janela.
func (l *limitador) permitir(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	agora := time.Now()

	// Faxina preguicosa: sem isto o map cresceria para sempre num processo que
	// roda por meses.
	if agora.Sub(l.limpeza) > 10*l.janela {
		for chave, quando := range l.ultimo {
			if agora.Sub(quando) > l.janela {
				delete(l.ultimo, chave)
			}
		}
		l.limpeza = agora
	}

	if anterior, existe := l.ultimo[ip]; existe && agora.Sub(anterior) < l.janela {
		return false
	}
	l.ultimo[ip] = agora
	return true
}

// ipDoRequest extrai o IP para o rate limit. Confia no X-Forwarded-For apenas
// porque em producao o Caddy fica na frente e o reescreve; exposto direto, um
// atacante poderia forjar o header e escapar do limite -- por isso ADDR passa a
// 127.0.0.1 quando o proxy entrar (ver make.ps1).
func ipDoRequest(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if primeiro, _, achou := strings.Cut(xff, ","); achou {
			return strings.TrimSpace(primeiro)
		}
		return strings.TrimSpace(xff)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// --- Validacao ---------------------------------------------------------------

// tamanhoMaxCampo corta textos absurdos antes de tocar o banco. Generoso para
// prosa real (a pergunta da motivacao merece espaco), restritivo para payload
// de ataque.
const tamanhoMaxCampo = 4000

// validar extrai e valida as respostas do form. Devolve os valores limpos e um
// map de erros por chave (vazio = tudo certo).
func validar(f *formBriefing, r *http.Request, lang string) (map[string]string, map[string]string) {
	valores := make(map[string]string, len(f.Perguntas))
	erros := map[string]string{}

	for _, p := range f.Perguntas {
		// Anexo nao tem valor de texto: ele viaja pelo multipart e vira linha
		// em submission_files (ver upload.go).
		if p.Tipo == campoArquivo {
			continue
		}
		valor := strings.TrimSpace(r.PostFormValue(p.Chave))
		if len(valor) > tamanhoMaxCampo {
			valor = valor[:tamanhoMaxCampo]
		}
		valores[p.Chave] = valor

		if p.Obrigatoria && valor == "" {
			erros[p.Chave] = t(lang, "form.obrigatorio")
			continue
		}
		if valor == "" {
			continue // opcional em branco: nada a validar
		}
		if p.Tipo == campoEmail && !reEmail.MatchString(valor) {
			erros[p.Chave] = t(lang, "form.emailInvalido")
		}
		if p.Tipo == campoContato && !contatoValido(valor) {
			erros[p.Chave] = t(lang, "form.contatoInvalido")
		}
		// Select: a resposta precisa ser uma das opcoes oferecidas. Sem isto, um
		// POST cru grava qualquer string num campo que a Isa le como categoria.
		if p.Tipo == campoSelect && !opcaoValida(p, valor) {
			erros[p.Chave] = t(lang, "form.opcaoInvalida")
		}
	}
	return valores, erros
}

// Validacao de contato. Deliberadamente PERMISSIVA na forma e ESTRITA no
// essencial: nao importa se a pessoa escreveu (41) 99999-8888, 41999998888 ou
// +55 41 9 9999-8888 -- importa que dê para responder. Rejeitar um formato
// valido custa um cliente; aceitar um numero incompleto custa o mesmo, so que
// sem ninguem perceber.
var (
	// Exige um ponto no dominio. Nao e o RFC 5322 completo (que aceita coisas
	// que nenhum provedor real emite) -- e o suficiente para pegar erro de
	// digitacao, que e o caso que acontece de verdade.
	reEmail     = regexp.MustCompile(`^[^@\s]+@[^@\s.]+(\.[^@\s.]+)+$`)
	reNaoDigito = regexp.MustCompile(`\D`)
)

// contatoValido aceita e-mail OU telefone brasileiro em qualquer formatacao.
func contatoValido(valor string) bool {
	if strings.Contains(valor, "@") {
		return reEmail.MatchString(valor)
	}
	return telefoneBRValido(reNaoDigito.ReplaceAllString(valor, ""))
}

// telefoneBRValido decide pelo COMPRIMENTO, nunca removendo um "55" do inicio.
// A pegadinha: 55 e um DDI (Brasil) mas tambem um DDD valido (Santa Maria/RS).
// Em "5511987654321" o 55 e pais; em "5599887766" e DDD. So o tamanho desfaz a
// ambiguidade -- por isso nada de TrimPrefix aqui.
func telefoneBRValido(digitos string) bool {
	switch len(digitos) {
	case 10, 11: // DDD + numero (fixo com 8, celular com 9)
		return true
	case 12, 13: // DDI 55 + DDD + numero
		return strings.HasPrefix(digitos, "55")
	}
	return false
}

// opcaoValida confere o valor de um select contra as opcoes. Compara nos DOIS
// idiomas: o visitante pode trocar de idioma com o form ja preenchido.
func opcaoValida(p pergunta, valor string) bool {
	for _, o := range p.Opcoes {
		if valor == o.pt || valor == o.en {
			return true
		}
	}
	return false
}

// --- Handlers ----------------------------------------------------------------

// briefing serve a pagina dedicada do briefing de fotografia (GET /briefing).
func (s *Server) briefing(w http.ResponseWriter, r *http.Request) {
	lang := idioma(r)
	f := &briefingFotografia

	titulo := "Briefing — Isabel Pontes"
	descr := "Conte sobre o ensaio que você quer fazer: tema, data, referências e o que te motivou."
	if lang == "en" {
		titulo = "Briefing — Isabel Pontes"
		descr = "Tell me about the shoot you have in mind: theme, date, references and what inspired you."
	}

	s.render.Page(w, r, "briefing", http.StatusOK, briefingView{
		pageData: s.novaPagina(lang, titulo, descr, "/briefing"),
		Form:     f,
		Action:   caminhoEnvio(lang, f.Slug),
		Valores:  map[string]string{},
		Erros:    map[string]string{},
	})
}

// anexarImagens grava os anexos de todos os campos de arquivo do formulario e
// devolve quantos foram salvos (o e-mail so precisa saber que existem).
func (s *Server) anexarImagens(r *http.Request, f *formBriefing, submissionID int64) int {
	total := 0
	if r.MultipartForm == nil {
		return total // formulario sem anexo (urlencoded) -- caminho normal
	}
	for _, p := range f.Perguntas {
		if p.Tipo != campoArquivo {
			continue
		}
		enviados := r.MultipartForm.File[p.Chave]
		if len(enviados) == 0 {
			continue
		}
		for _, arquivo := range s.salvarAnexos(submissionID, p.Chave, enviados) {
			if err := s.store.SalvarArquivo(r.Context(), arquivo); err != nil {
				logger.Error("briefing: registrar anexo falhou", "id", submissionID, "erro", err)
				continue
			}
			total++
		}
	}
	return total
}

// caminhoEnvio monta o destino do POST no idioma corrente, para o envio sem JS
// nao jogar um visitante de /en para a versao PT.
func caminhoEnvio(lang, slug string) string {
	if lang == "en" {
		return "/en/briefing/" + slug
	}
	return "/briefing/" + slug
}

// enviarBriefing trata o POST dos dois formularios (o slug vem da rota).
func (s *Server) enviarBriefing(w http.ResponseWriter, r *http.Request) {
	lang := idioma(r)
	f, existe := formsPorSlug[r.PathValue("slug")]
	if !existe {
		s.notFound(w, r)
		return
	}

	responder := func(v briefingView, status int) {
		v.pageData = s.novaPagina(lang, "Briefing — Isabel Pontes", "", "/briefing")
		v.Form = f
		v.Action = caminhoEnvio(lang, f.Slug)

		// Com HTMX, devolve so o formulario (swap no lugar). Sem JS, a pagina
		// inteira -- o mesmo ViewModel serve os dois.
		//
		// O fragmento sai SEMPRE 200, ignorando o `status`: o HTMX so troca o
		// DOM em resposta 2xx, entao devolver 422 com os erros de validacao
		// faria o visitante clicar em "enviar" e nao ver absolutamente nada
		// acontecer. Erro de preenchimento e desfecho normal da interacao, nao
		// falha de transporte -- o HTML devolvido JA comunica o erro. O status
		// semantico continua valendo no caminho sem JS, que e o que um cliente
		// automatizado veria.
		if r.Header.Get("HX-Request") == "true" {
			s.render.Fragment(w, "briefing", "form_briefing", v)
			return
		}
		s.render.Page(w, r, "briefing", status, v)
	}

	// Teto no corpo ANTES de qualquer parsing: o ParseMultipartForm faz spool
	// em disco, entao sem isto um POST gigante enche a VM antes de o codigo
	// chegar a validar coisa alguma.
	r.Body = http.MaxBytesReader(w, r.Body, maxCorpoUpload)

	// ParseMultipartForm trata os dois content-types (multipart e urlencoded),
	// entao serve tambem para os formularios sem anexo.
	if err := r.ParseMultipartForm(2 << 20); err != nil && !errors.Is(err, http.ErrNotMultipart) {
		var excedeu *http.MaxBytesError
		chave := "form.erroGeral"
		if errors.As(err, &excedeu) {
			chave = "form.arquivoGrande"
		}
		logger.Warn("briefing: corpo recusado", "form", f.Slug, "erro", err)
		responder(briefingView{Erros: map[string]string{"_geral": t(lang, chave)}}, http.StatusBadRequest)
		return
	}

	// Honeypot: responde SUCESSO ao bot. Um erro ensinaria o autor do bot a
	// contornar o campo; um "obrigada" faz ele registrar vitoria e ir embora.
	if strings.TrimSpace(r.PostFormValue(campoHoneypot)) != "" {
		logger.Info("briefing: honeypot", "form", f.Slug, "ip", ipDoRequest(r))
		responder(briefingView{Enviado: true}, http.StatusOK)
		return
	}

	valores, erros := validar(f, r, lang)
	if len(erros) > 0 {
		responder(briefingView{Valores: valores, Erros: erros}, http.StatusUnprocessableEntity)
		return
	}

	if !s.limiteBriefing.permitir(ipDoRequest(r)) {
		responder(briefingView{
			Valores: valores,
			Erros:   map[string]string{"_geral": t(lang, "form.muitosEnvios")},
		}, http.StatusTooManyRequests)
		return
	}

	id, err := s.store.SalvarSubmission(r.Context(), f.Slug, lang, valores,
		ipDoRequest(r), r.UserAgent())
	if err != nil {
		// Falha de banco e nossa, nao do visitante: registra com o conteudo para
		// a resposta nao se perder junto com o erro.
		logger.Error("briefing: salvar falhou", "form", f.Slug, "erro", err, "respostas", valores)
		responder(briefingView{
			Valores: valores,
			Erros:   map[string]string{"_geral": t(lang, "form.erroGeral")},
		}, http.StatusInternalServerError)
		return
	}
	// Anexos DEPOIS do insert: eles precisam do id, e um erro aqui nao pode
	// custar a resposta em texto, que ja esta a salvo.
	anexos := s.anexarImagens(r, f, id)

	logger.Info("briefing recebido", "form", f.Slug, "id", id, "lang", lang)

	// E-mail em goroutine: o Resend pode estar fora do ar ou lento, e o visitante
	// nao pode esperar por isso -- a resposta JA esta salva, que e o que importa.
	go s.notificar(f, lang, valores, id, anexos)

	responder(briefingView{Enviado: true}, http.StatusOK)
}
