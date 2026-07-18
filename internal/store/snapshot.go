package store

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

// saoPaulo fixa o fuso para o "dia" do shuffle virar a meia-noite de Brasilia,
// nao em UTC.
var saoPaulo = func() *time.Location {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return time.FixedZone("BRT", -3*3600)
	}
	return loc
}()

// Snapshot e a copia em RAM que os requests leem. O banco so e tocado no boot e
// no sync -- nunca ao servir. E o que justifica manter o SQLite sem custo no
// caminho quente.
//
// O shuffle usa uma seed por DIA, nao por request. Shuffle por request geraria
// HTML diferente a cada acesso, o que mata ETag/cache. Com seed diaria, todo
// visitante do dia ve a mesma ordem -> resposta cacheavel -> e o site "muda"
// todo dia, que e 100% do efeito percebido.
type Snapshot struct {
	mu          sync.RWMutex
	temas       []Tema
	porTema     map[string][]Photo // ja embaralhado para o dia
	todas       []Photo            // idem, todas as fotos
	dataVersion int
	dia         string // AAAA-MM-DD do shuffle atual (Sao Paulo)
}

// LoadSnapshot le tudo do banco e monta o snapshot ja embaralhado para hoje.
func (s *Store) LoadSnapshot(ctx context.Context) (*Snapshot, error) {
	temas, err := s.LoadTemas(ctx)
	if err != nil {
		return nil, err
	}
	fotos, err := s.LoadPhotos(ctx)
	if err != nil {
		return nil, err
	}
	snap := &Snapshot{dataVersion: s.DataVersion()}
	snap.rebuild(fotos, temas)
	return snap, nil
}

func diaAtual() string { return time.Now().In(saoPaulo).Format("2006-01-02") }

func daySeed() int64 {
	y, m, d := time.Now().In(saoPaulo).Date()
	return int64(y)*10000 + int64(m)*100 + int64(d)
}

// rebuild reembaralha e reindexa por tema. Fotos com sort_order definido ficam
// fixas no topo, na ordem; o resto e embaralhado com a seed do dia.
func (snap *Snapshot) rebuild(fotos []Photo, temas []Tema) {
	aleatorio := rand.New(rand.NewSource(daySeed()))

	// separa as fotos com ordem definida (ficam fixas no topo) das que entram no
	// shuffle
	fixas := make([]Photo, 0)
	soltas := make([]Photo, 0, len(fotos))
	for _, foto := range fotos {
		if foto.SortOrder != nil {
			fixas = append(fixas, foto)
		} else {
			soltas = append(soltas, foto)
		}
	}
	aleatorio.Shuffle(len(soltas), func(i, j int) {
		soltas[i], soltas[j] = soltas[j], soltas[i]
	})

	todas := append(fixas, soltas...)

	// indexa por tema para o filtro nao precisar percorrer tudo a cada request
	porTema := make(map[string][]Photo, len(temas))
	for _, foto := range todas {
		porTema[foto.TemaSlug] = append(porTema[foto.TemaSlug], foto)
	}

	snap.mu.Lock()
	snap.temas = temas
	snap.todas = todas
	snap.porTema = porTema
	snap.dia = diaAtual()
	snap.mu.Unlock()
}

// EnsureDay reembaralha se o dia virou. Barato: so compara a string do dia.
// Chamado no inicio de cada request de galeria.
func (snap *Snapshot) EnsureDay() {
	snap.mu.RLock()
	precisa := snap.dia != diaAtual()
	snap.mu.RUnlock()
	if !precisa {
		return
	}
	snap.mu.RLock()
	fotos := snap.todas
	temas := snap.temas
	snap.mu.RUnlock()
	snap.rebuild(fotos, temas)
}

// Temas devolve os temas para os botoes do filtro.
func (snap *Snapshot) Temas() []Tema {
	snap.mu.RLock()
	defer snap.mu.RUnlock()
	return snap.temas
}

// Fotos devolve as fotos de um tema (ou todas, se slug == ""), ja embaralhadas
// para o dia. A fatia e compartilhada -- o chamador nao deve modifica-la.
func (snap *Snapshot) Fotos(temaSlug string) []Photo {
	snap.mu.RLock()
	defer snap.mu.RUnlock()
	if temaSlug == "" {
		return snap.todas
	}
	return snap.porTema[temaSlug]
}

// TemaValido diz se o slug existe (para a galeria dar 404 em tema inventado em
// vez de lista vazia).
func (snap *Snapshot) TemaValido(slug string) bool {
	snap.mu.RLock()
	defer snap.mu.RUnlock()
	for _, t := range snap.temas {
		if t.Slug == slug {
			return true
		}
	}
	return false
}

// DiaSeed identifica a ordem do dia; entra no ETag junto com data_version.
func (snap *Snapshot) DiaSeed() string {
	snap.mu.RLock()
	defer snap.mu.RUnlock()
	return snap.dia
}

// DataVersion do snapshot (fixado no load; um sync exige reload para mudar).
func (snap *Snapshot) DataVersion() int {
	snap.mu.RLock()
	defer snap.mu.RUnlock()
	return snap.dataVersion
}
