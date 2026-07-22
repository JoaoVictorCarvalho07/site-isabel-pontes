package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"
)

// Notificacao por e-mail via Resend. E o UNICO serviço externo do projeto
// (ver docs §8) e, de proposito, o mais descartavel: sem RESEND_API_KEY o site
// funciona inteiro, a resposta e salva no SQLite e a Isa le pelo /admin. O
// e-mail e conveniencia, nunca a fonte de verdade.

// notificar dispara o e-mail de aviso. Roda em goroutine (ver enviarBriefing) e
// por isso NAO recebe o contexto do request: o cancelamento do request -- que
// acontece assim que o visitante ve "obrigada" -- mataria o envio no meio.
func (s *Server) notificar(f *formBriefing, lang string, respostas map[string]string, id int64, anexos int) {
	if s.cfg.ResendAPIKey == "" || s.cfg.MailPara == "" {
		logger.Info("briefing: e-mail desligado (sem RESEND_API_KEY/MAIL_PARA)", "id", id)
		return
	}
	ctx, cancelar := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelar()

	assunto := fmt.Sprintf("Novo briefing (%s) — %s", f.Slug, resumo(f, respostas))
	if err := s.enviarEmail(ctx, assunto, corpoEmail(f, lang, respostas, id, s.cfg.BaseURL, anexos)); err != nil {
		// Erro de e-mail nao pode assustar: a resposta ESTA salva. Log e segue.
		logger.Error("briefing: e-mail falhou (resposta salva mesmo assim)", "id", id, "erro", err)
		return
	}
	logger.Info("briefing: e-mail enviado", "id", id)
}

// resumo escolhe um identificador humano para o assunto do e-mail: o primeiro
// campo obrigatorio de texto preenchido (nome, ou marca no UGC).
func resumo(f *formBriefing, respostas map[string]string) string {
	for _, p := range f.Perguntas {
		if p.Obrigatoria && (p.Tipo == campoTexto) {
			if v := respostas[p.Chave]; v != "" {
				return v
			}
		}
	}
	return "sem identificacao"
}

// corpoEmail monta o HTML do aviso: pergunta e resposta, na ordem do formulario.
func corpoEmail(f *formBriefing, lang string, respostas map[string]string, id int64, baseURL string, anexos int) string {
	var sb strings.Builder
	sb.WriteString(`<div style="font-family:system-ui,sans-serif;font-size:15px;line-height:1.6;color:#2b2b2b">`)
	fmt.Fprintf(&sb, `<h2 style="font-weight:600">%s</h2>`, html.EscapeString(f.Titulo.T("pt")))
	fmt.Fprintf(&sb, `<p style="color:#777;font-size:13px">Briefing #%d · formulário <strong>%s</strong> · idioma %s</p>`,
		id, html.EscapeString(f.Slug), html.EscapeString(lang))

	for _, p := range f.Perguntas {
		// Anexos nao vao no e-mail: seriam imagens de terceiros trafegando por
		// um servico externo, e o link para ve-las exige autenticacao mesmo. O
		// e-mail so avisa que existem; ver a imagem e no painel.
		if p.Tipo == campoArquivo {
			continue
		}
		valor := respostas[p.Chave]
		if valor == "" {
			continue // campo opcional em branco so poluiria o e-mail
		}
		fmt.Fprintf(&sb,
			`<p style="margin:14px 0"><strong style="display:block;color:#5f5f5f;font-size:13px">%s</strong>%s</p>`,
			html.EscapeString(p.Label.T("pt")),
			// nl2br: a motivacao e as referencias costumam vir em varias linhas.
			strings.ReplaceAll(html.EscapeString(valor), "\n", "<br>"))
	}

	if anexos > 0 {
		fmt.Fprintf(&sb,
			`<p style="margin:14px 0;padding:10px 12px;background:#f4efe6;border-radius:6px">`+
				`📎 <strong>%d imagem(ns) de referência</strong> anexada(s) — veja no painel.</p>`, anexos)
	}

	painel := html.EscapeString(baseURL + "/admin")
	fmt.Fprintf(&sb, `<hr style="border:none;border-top:1px solid #e5e5e5;margin:24px 0">`+
		`<p style="font-size:13px;color:#777">Também disponível em <a href="%s">%s</a>.</p>`,
		painel, painel)
	sb.WriteString(`</div>`)
	return sb.String()
}

// enviarEmail faz o POST na API do Resend. Sem SDK: e um POST JSON: uma
// dependencia a menos para auditar e atualizar.
func (s *Server) enviarEmail(ctx context.Context, assunto, corpoHTML string) error {
	corpo, err := json.Marshal(map[string]any{
		"from":     s.cfg.MailDe,
		"to":       []string{s.cfg.MailPara},
		"subject":  assunto,
		"html":     corpoHTML,
		"reply_to": s.cfg.MailPara,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.resend.com/emails", bytes.NewReader(corpo))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.ResendAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		// Le um pedaco da resposta: o Resend explica o motivo (dominio nao
		// verificado, chave invalida) e sem isso o debug vira adivinhacao.
		var detalhe bytes.Buffer
		_, _ = detalhe.ReadFrom(http.MaxBytesReader(nil, resp.Body, 2048))
		return fmt.Errorf("resend devolveu %d: %s", resp.StatusCode, detalhe.String())
	}
	return nil
}
