// Package store e a camada de dados: SQLite como fonte de verdade editavel.
//
// O padrao (ver snapshot.go): sync escreve, boot carrega para RAM, request nao
// toca o banco. Para 190 fotos read-only o SQLite nao existe pela velocidade de
// query -- existe para dar labels de tema com acento, ordem curada, publicar/
// despublicar e, no futuro, ensaios, sem redeploy.
package store

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"strconv"

	_ "modernc.org/sqlite" // driver CGO-free: cross-compila p/ ARM sem toolchain C
)

//go:embed schema.sql
var schemaSQL string

// 2: tabela submissions (formularios de briefing).
// 3: tabela submission_files (imagens de referencia anexadas).
const schemaVersion = 3

type Store struct {
	db *sql.DB
}

// Open abre (ou cria) o banco e aplica as migracoes.
func Open(caminho string) (*Store, error) {
	// _busy_timeout evita "database is locked" se um sync coincidir com leitura.
	dsn := "file:" + caminho + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// modernc lida bem com 1 conexao de escrita; o pool default serve.
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) DB() *sql.DB { return s.db }

// migrate aplica o schema e versiona. Um switch com um case por versao resolve
// para sempre neste projeto -- goose/atlas seria peso morto para 190 fotos.
func (s *Store) migrate() error {
	if _, err := s.db.Exec(schemaSQL); err != nil {
		return fmt.Errorf("aplicar schema: %w", err)
	}
	versaoAtual := s.metaInt("schema_version", 0)
	// (nenhuma migracao incremental ainda; o proximo case entra aqui)
	// switch { case versaoAtual < 2: ...; }
	if versaoAtual < schemaVersion {
		if err := s.setMeta("schema_version", strconv.Itoa(schemaVersion)); err != nil {
			return err
		}
	}
	return nil
}

// --- meta (tabela chave-valor de sistema; colunas k e v no SQL) ---

func (s *Store) metaInt(chave string, padrao int) int {
	var valorStr string
	err := s.db.QueryRow("SELECT v FROM meta WHERE k = ?", chave).Scan(&valorStr)
	if err != nil {
		return padrao
	}
	numero, err := strconv.Atoi(valorStr)
	if err != nil {
		return padrao
	}
	return numero
}

func (s *Store) setMeta(chave, valor string) error {
	_, err := s.db.Exec(
		"INSERT INTO meta(k,v) VALUES(?,?) ON CONFLICT(k) DO UPDATE SET v=excluded.v",
		chave, valor)
	return err
}

// DataVersion incrementa a cada sync; entra no ETag para invalidar cache.
func (s *Store) DataVersion() int { return s.metaInt("data_version", 0) }

func (s *Store) bumpDataVersion(ctx context.Context, tx *sql.Tx) error {
	var versao int
	_ = tx.QueryRowContext(ctx, "SELECT v FROM meta WHERE k='data_version'").Scan(&versao)
	_, err := tx.ExecContext(ctx,
		"INSERT INTO meta(k,v) VALUES('data_version',?) ON CONFLICT(k) DO UPDATE SET v=excluded.v",
		strconv.Itoa(versao+1))
	return err
}
