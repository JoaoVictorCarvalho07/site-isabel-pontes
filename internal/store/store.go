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

const schemaVersion = 1

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
	atual := s.metaInt("schema_version", 0)
	// (nenhuma migracao incremental ainda; o proximo case entra aqui)
	// switch { case atual < 2: ...; }
	if atual < schemaVersion {
		if err := s.setMeta("schema_version", strconv.Itoa(schemaVersion)); err != nil {
			return err
		}
	}
	return nil
}

// --- meta ---

func (s *Store) metaInt(k string, def int) int {
	var v string
	err := s.db.QueryRow("SELECT v FROM meta WHERE k = ?", k).Scan(&v)
	if err != nil {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func (s *Store) setMeta(k, v string) error {
	_, err := s.db.Exec(
		"INSERT INTO meta(k,v) VALUES(?,?) ON CONFLICT(k) DO UPDATE SET v=excluded.v",
		k, v)
	return err
}

// DataVersion incrementa a cada sync; entra no ETag para invalidar cache.
func (s *Store) DataVersion() int { return s.metaInt("data_version", 0) }

func (s *Store) bumpDataVersion(ctx context.Context, tx *sql.Tx) error {
	var v int
	_ = tx.QueryRowContext(ctx, "SELECT v FROM meta WHERE k='data_version'").Scan(&v)
	_, err := tx.ExecContext(ctx,
		"INSERT INTO meta(k,v) VALUES('data_version',?) ON CONFLICT(k) DO UPDATE SET v=excluded.v",
		strconv.Itoa(v+1))
	return err
}
