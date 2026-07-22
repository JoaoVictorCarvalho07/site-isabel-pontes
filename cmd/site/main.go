// Comando site: um binario, todos os subcomandos.
//
//	site serve            sobe o servidor HTTP
//	site sync [caminho]   le o manifest de midia e popula o SQLite
//	site snapshot [dest]  copia consistente do banco, com o servidor no ar
//
// Um binario em vez de scripts separados porque `sync` e `serve` compartilham
// config e store. O `sync` consome o manifest.json produzido pelo
// migrate_media.py (que ja tem dimensoes, variantes e lqip prontos) em vez de
// listar o R2 -- o Go nao processa imagem.
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
	"github.com/JoaoVictorCarvalho07/site-da-bebel-htmx/internal/store"
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
		return sync()
	case "snapshot":
		return snapshot()
	default:
		return fmt.Errorf("subcomando desconhecido: %q (use: serve | sync | snapshot)", cmd)
	}
}

// snapshot grava uma copia consistente do banco, com o servidor no ar.
//
//	site snapshot [destino]     # padrao: <DB_PATH>.snapshot
//
// Existe como subcomando (e nao como linha no script de backup) porque copiar
// um SQLite em modo WAL com `cp` produz arquivo corrompido em silencio -- as
// escritas recentes estao no -wal. Ver store.Snapshot. Assim o backup nao
// depende do sqlite3 CLI estar instalado na VM.
func snapshot() error {
	cfg, err := config.Load(false)
	if err != nil {
		return err
	}
	destino := cfg.DBPath + ".snapshot"
	if len(os.Args) > 2 {
		destino = os.Args[2]
	}

	banco, err := store.Open(cfg.DBPath)
	if err != nil {
		return err
	}
	defer banco.Close()

	if err := banco.Snapshot(context.Background(), destino); err != nil {
		return err
	}
	info, err := os.Stat(destino)
	if err != nil {
		return err
	}
	slog.Info("snapshot gravado", "destino", destino, "kb", info.Size()/1024)
	return nil
}

// sync le o manifest de midia (saida do migrate_media.py) e popula o SQLite.
// Roda como job manual ou cron -- nunca no boot. Se falhar, o site continua
// servindo o snapshot anterior.
func sync() error {
	cfg, err := config.Load(false)
	if err != nil {
		return err
	}
	banco, err := store.Open(cfg.DBPath)
	if err != nil {
		return err
	}
	defer banco.Close()

	caminhoManifest := cfg.ManifestPath
	if len(os.Args) > 2 {
		caminhoManifest = os.Args[2] // `site sync <caminho>`
	}

	relatorio, err := banco.SyncFromManifest(context.Background(), caminhoManifest)
	if err != nil {
		return err
	}
	slog.Info("sync concluido",
		"fotos", relatorio.Fotos, "temas", relatorio.Temas, "data_version", relatorio.DataVersion,
		"manifest", caminhoManifest, "db", cfg.DBPath)
	return nil
}

func serve() error {
	cfg, err := config.Load(false) // servir nao precisa de credencial R2
	if err != nil {
		return err
	}

	// Abre o banco e carrega o snapshot para a RAM UMA vez, no boot. Dai em
	// diante os requests leem so da memoria. Se o banco ainda nao foi populado
	// (sem `sync`), o snapshot vem vazio e a galeria mostra "nenhuma foto" --
	// o servidor sobe do mesmo jeito.
	banco, err := store.Open(cfg.DBPath)
	if err != nil {
		return err
	}
	defer banco.Close()

	snapshot, err := banco.LoadSnapshot(context.Background())
	if err != nil {
		return err
	}
	slog.Info("snapshot carregado", "fotos", len(snapshot.Fotos("")), "temas", len(snapshot.Temas()))

	servidor, err := web.New(cfg, site.Assets, snapshot, banco)
	if err != nil {
		return err
	}

	servidorHTTP := &http.Server{
		Addr:    cfg.Addr,
		Handler: servidor.Handler(),
		// Sem timeout, um cliente lento segura a conexao para sempre.
		// O default do net/http e nao ter nenhum, o que e uma armadilha.
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Shutdown limpo: sem isto, um deploy corta requests no meio.
	ctx, pararSinal := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer pararSinal()

	desligou := make(chan struct{})
	go func() {
		<-ctx.Done()
		slog.Info("desligando...")
		ctxDesligar, cancelar := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelar()
		_ = servidorHTTP.Shutdown(ctxDesligar)
		close(desligou)
	}()

	slog.Info("no ar", "addr", cfg.Addr, "dev", cfg.Dev, "base_url", cfg.BaseURL)
	// ListenAndServe sempre devolve erro nao-nil; ErrServerClosed e o caso feliz.
	if err := servidorHTTP.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	<-desligou // deixa o Shutdown drenar as conexoes antes de sair
	return nil
}
