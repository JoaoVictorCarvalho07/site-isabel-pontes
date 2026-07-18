// Package config carrega a configuracao do ambiente e falha no boot se algo
// essencial faltar -- em vez de descobrir no primeiro request.
package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Addr         string // porta HTTP
	DBPath       string // caminho do SQLite
	ManifestPath string // manifest de midia que o `sync` ingere
	BaseURL      string // URL canonica; entra no og:url, canonical e sitemap
	Dev          bool   // reparse de template por request

	// R2: usado somente pelo `site sync`, nunca ao servir request.
	R2PublicBase string
	R2AccountID  string
	R2AccessKey  string
	R2SecretKey  string
	R2Bucket     string
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// Load le o ambiente. `needR2` so e true no comando sync: servir o site nao
// precisa de credencial nenhuma, e exigir isso no boot do servidor faria o
// deploy carregar segredo que ele nao usa.
func Load(needR2 bool) (*Config, error) {
	cfg := &Config{
		Addr:         env("ADDR", ":8080"),
		DBPath:       env("DB_PATH", "data/site.db"),
		ManifestPath: env("MANIFEST_PATH", "data/manifest.json"),
		BaseURL:      strings.TrimSuffix(env("BASE_URL", "http://localhost:8080"), "/"),
		Dev:          env("DEV", "") != "",
		R2PublicBase: strings.TrimSuffix(os.Getenv("R2_PUBLIC_BASE"), "/"),
		R2AccountID:  os.Getenv("R2_ACCOUNT_ID"),
		R2AccessKey:  os.Getenv("R2_ACCESS_KEY_ID"),
		R2SecretKey:  os.Getenv("R2_SECRET_ACCESS_KEY"),
		R2Bucket:     os.Getenv("R2_BUCKET"),
	}

	if needR2 {
		faltando := []string{}
		for nomeVar, valor := range map[string]string{
			"R2_PUBLIC_BASE":       cfg.R2PublicBase,
			"R2_ACCOUNT_ID":        cfg.R2AccountID,
			"R2_ACCESS_KEY_ID":     cfg.R2AccessKey,
			"R2_SECRET_ACCESS_KEY": cfg.R2SecretKey,
			"R2_BUCKET":            cfg.R2Bucket,
		} {
			if valor == "" {
				faltando = append(faltando, nomeVar)
			}
		}
		if len(faltando) > 0 {
			return nil, fmt.Errorf("faltam variaveis de ambiente: %s (veja .env.example)", strings.Join(faltando, ", "))
		}
	}

	return cfg, nil
}
