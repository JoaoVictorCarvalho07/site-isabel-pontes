// Comando site: um binario, todos os subcomandos.
//
//	site serve   sobe o servidor HTTP
//	site sync    le o bucket R2 e atualiza o SQLite
//
// Um binario em vez de scripts separados porque `sync` e `serve` compartilham
// config e store. No projeto React isso era um scripts/generate-manifest.ts que
// rodava no build e exigia as credenciais R2 no ambiente da Vercel para o site
// conseguir buildar.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	site "github.com/JoaoVictorCarvalho07/site-da-bebel-htmx"
	"github.com/JoaoVictorCarvalho07/site-da-bebel-htmx/internal/config"
	"github.com/JoaoVictorCarvalho07/site-da-bebel-htmx/internal/web"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
}

func run() error {
	cmd := "serve"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "serve":
		return serve()
	case "sync":
		// Fase 1.
		return errors.New("`sync` ainda nao foi implementado (Fase 1)")
	default:
		return fmt.Errorf("subcomando desconhecido: %q (use: serve | sync)", cmd)
	}
}

func serve() error {
	cfg, err := config.Load(false) // servir nao precisa de credencial R2
	if err != nil {
		return err
	}

	srv, err := web.New(cfg, site.Assets)
	if err != nil {
		return err
	}

	httpSrv := &http.Server{
		Addr:    cfg.Addr,
		Handler: srv.Handler(),
		// Sem timeout, um cliente lento segura a conexao para sempre.
		// O default do net/http e nao ter nenhum, o que e uma armadilha.
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Shutdown limpo: sem isto, um deploy corta requests no meio.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	parou := make(chan struct{})
	go func() {
		<-ctx.Done()
		slog.Info("desligando...")
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shutCtx)
		close(parou)
	}()

	slog.Info("no ar", "addr", cfg.Addr, "dev", cfg.Dev, "base_url", cfg.BaseURL)
	// ListenAndServe sempre devolve erro nao-nil; ErrServerClosed e o caso feliz.
	if err := httpSrv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	<-parou // deixa o Shutdown drenar as conexoes antes de sair
	return nil
}
