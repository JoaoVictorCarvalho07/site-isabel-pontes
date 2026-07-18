package store

import "context"

// Photo espelha uma linha de `photos`. Os campos de URL/lqip vem prontos do
// manifest; o Go so os serve.
type Photo struct {
	ID        int64
	Key       string
	TemaSlug  string
	W, H      int
	Grid      string
	Grid2x    string
	Full      string
	LQIP      string
	Alt       string
	SortOrder *int // nil => entra no shuffle
}

// Tema com o label acentuado que o usuario ve.
type Tema struct {
	Slug      string
	Label     string
	SortOrder int
}

// LoadPhotos le todas as fotos publicadas, ordenadas por (tema, sort_order, id).
// Chamado uma vez no boot pelo snapshot; nao e caminho de request.
func (s *Store) LoadPhotos(ctx context.Context) ([]Photo, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, key, tema_slug, w, h, grid, grid2x, full, lqip, alt, sort_order
		FROM photos
		WHERE published = 1
		ORDER BY tema_slug, sort_order IS NULL, sort_order, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fotos []Photo
	for rows.Next() {
		var foto Photo
		if err := rows.Scan(&foto.ID, &foto.Key, &foto.TemaSlug, &foto.W, &foto.H,
			&foto.Grid, &foto.Grid2x, &foto.Full, &foto.LQIP, &foto.Alt, &foto.SortOrder); err != nil {
			return nil, err
		}
		fotos = append(fotos, foto)
	}
	return fotos, rows.Err()
}

// LoadTemas le os temas visiveis, ja ordenados para os botoes do filtro.
func (s *Store) LoadTemas(ctx context.Context) ([]Tema, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT slug, label, sort_order
		FROM temas
		WHERE visible = 1
		ORDER BY sort_order, label`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var temas []Tema
	for rows.Next() {
		var tema Tema
		if err := rows.Scan(&tema.Slug, &tema.Label, &tema.SortOrder); err != nil {
			return nil, err
		}
		temas = append(temas, tema)
	}
	return temas, rows.Err()
}
