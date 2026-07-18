package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// manifestEntry e o shape produzido por scripts/migrate_media.py.
type manifestEntry struct {
	Key    string `json:"key"`
	Tema   string `json:"tema"`
	W      int    `json:"w"`
	H      int    `json:"h"`
	Grid   string `json:"grid"`
	Grid2x string `json:"grid2x"`
	Full   string `json:"full"`
	LQIP   string `json:"lqip"`
}

// labelsTema mapeia o slug (pasta no R2) para o rotulo que o usuario ve.
// O manifest so tem o slug cru; o acento e a capitalizacao vivem aqui.
// Tema novo que apareca sem entrada aqui cai no fallback (Title-case do slug).
var labelsTema = map[string]struct {
	Label string
	Ordem int
}{
	"Eventos":       {"Eventos", 10},
	"personalizado": {"Personalizado", 20},
	"epoca":         {"Época", 30},
	"artistico":     {"Artístico", 40},
	"casal":         {"Casal", 50},
	"conceitual":    {"Conceitual", 60},
	"outros":        {"Outros", 90},
}

// SyncReport resume o que o sync fez, para o comando imprimir.
type SyncReport struct {
	Fotos       int
	Temas       int
	DataVersion int
}

// SyncFromManifest le o manifest.json e substitui o conteudo do banco por ele.
//
// Faz replace completo dentro de uma transacao: o manifest e a fonte de verdade
// da migracao, entao e mais simples e seguro reescrever do que fazer diff. As
// colunas editaveis a mao (temas.label, futuramente photos.alt curado) sao
// preservadas por merge -- ver abaixo.
func (s *Store) SyncFromManifest(ctx context.Context, caminho string) (SyncReport, error) {
	var relatorio SyncReport

	conteudo, err := os.ReadFile(caminho)
	if err != nil {
		return relatorio, fmt.Errorf("ler manifest %s: %w", caminho, err)
	}
	var entradas []manifestEntry
	if err := json.Unmarshal(conteudo, &entradas); err != nil {
		return relatorio, fmt.Errorf("parse do manifest: %w", err)
	}
	if len(entradas) == 0 {
		return relatorio, fmt.Errorf("manifest vazio -- abortando para nao esvaziar o banco")
	}

	// labels editados a mao no banco nao devem ser sobrescritos pelo sync.
	labelsExistentes, err := s.labelsAtuais(ctx)
	if err != nil {
		return relatorio, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return relatorio, err
	}
	defer func() { _ = tx.Rollback() }()

	// junta os slugs de tema presentes no manifest (sem repetir), ordenados
	temasPresentes := map[string]bool{}
	for _, entrada := range entradas {
		temasPresentes[entrada.Tema] = true
	}
	slugs := make([]string, 0, len(temasPresentes))
	for slug := range temasPresentes {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)

	if _, err := tx.ExecContext(ctx, "DELETE FROM photos"); err != nil {
		return relatorio, err
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM temas"); err != nil {
		return relatorio, err
	}

	for _, slug := range slugs {
		label, ordem := rotuloTema(slug)
		// preserva label editado a mao, se houver
		if existente, ok := labelsExistentes[slug]; ok && existente != "" {
			label = existente
		}
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO temas(slug,label,sort_order,visible) VALUES(?,?,?,1)",
			slug, label, ordem); err != nil {
			return relatorio, fmt.Errorf("inserir tema %s: %w", slug, err)
		}
	}

	inserirFoto, err := tx.PrepareContext(ctx, `
		INSERT INTO photos(key,tema_slug,w,h,grid,grid2x,full,lqip)
		VALUES(?,?,?,?,?,?,?,?)`)
	if err != nil {
		return relatorio, err
	}
	defer inserirFoto.Close()

	for _, entrada := range entradas {
		if _, err := inserirFoto.ExecContext(ctx,
			entrada.Key, entrada.Tema, entrada.W, entrada.H,
			entrada.Grid, entrada.Grid2x, entrada.Full, entrada.LQIP); err != nil {
			return relatorio, fmt.Errorf("inserir foto %s: %w", entrada.Key, err)
		}
	}

	if err := s.bumpDataVersion(ctx, tx); err != nil {
		return relatorio, err
	}
	if err := tx.Commit(); err != nil {
		return relatorio, err
	}

	relatorio.Fotos = len(entradas)
	relatorio.Temas = len(slugs)
	relatorio.DataVersion = s.DataVersion()
	return relatorio, nil
}

func (s *Store) labelsAtuais(ctx context.Context) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT slug, label FROM temas")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	labels := map[string]string{}
	for rows.Next() {
		var slug, label string
		if err := rows.Scan(&slug, &label); err != nil {
			return nil, err
		}
		labels[slug] = label
	}
	return labels, rows.Err()
}

func rotuloTema(slug string) (string, int) {
	if info, ok := labelsTema[slug]; ok {
		return info.Label, info.Ordem
	}
	// fallback: Title-case simples, ordem no fim
	return titulo(slug), 100
}

func titulo(s string) string {
	if s == "" {
		return s
	}
	// so capitaliza a primeira letra; slugs com acento ja vem do mapa labelsTema
	b := []rune(s)
	if b[0] >= 'a' && b[0] <= 'z' {
		b[0] -= 32
	}
	return string(b)
}
