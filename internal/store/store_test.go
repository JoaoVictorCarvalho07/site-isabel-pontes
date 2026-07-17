package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// manifest de teste: 2 temas, um deles ('epoca') exercitando o label acentuado.
const manifestFixture = `[
  {"key":"epoca/epoca_1.webp","tema":"epoca","w":800,"h":1200,"grid":"g1","grid2x":"g1x","full":"f1","lqip":"data:,1"},
  {"key":"epoca/epoca_2.webp","tema":"epoca","w":800,"h":1200,"grid":"g2","grid2x":"g2x","full":"f2","lqip":"data:,2"},
  {"key":"epoca/epoca_3.webp","tema":"epoca","w":800,"h":1200,"grid":"g3","grid2x":"g3x","full":"f3","lqip":"data:,3"},
  {"key":"casal/casal_1.webp","tema":"casal","w":800,"h":1200,"grid":"c1","grid2x":"c1x","full":"cf1","lqip":"data:,c1"},
  {"key":"casal/casal_2.webp","tema":"casal","w":800,"h":1200,"grid":"c2","grid2x":"c2x","full":"cf2","lqip":"data:,c2"}
]`

func storeComManifest(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	mf := filepath.Join(dir, "manifest.json")
	if err := os.WriteFile(mf, []byte(manifestFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	st, err := Open(filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { st.Close() })

	rep, err := st.SyncFromManifest(context.Background(), mf)
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if rep.Fotos != 5 || rep.Temas != 2 {
		t.Fatalf("relatorio = %+v, quero 5 fotos / 2 temas", rep)
	}
	return st
}

func TestSyncELabelAcentuado(t *testing.T) {
	st := storeComManifest(t)
	temas, err := st.LoadTemas(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	labels := map[string]string{}
	for _, tm := range temas {
		labels[tm.Slug] = tm.Label
	}
	// o bug do React: filtro mostrava 'epoca' cru. Aqui tem que virar 'Época'.
	if labels["epoca"] != "Época" {
		t.Errorf("label de epoca = %q, quero 'Época'", labels["epoca"])
	}
	if labels["casal"] != "Casal" {
		t.Errorf("label de casal = %q, quero 'Casal'", labels["casal"])
	}
}

func TestSyncEIdempotente(t *testing.T) {
	st := storeComManifest(t)
	dv1 := st.DataVersion()
	dir := t.TempDir()
	mf := filepath.Join(dir, "m.json")
	_ = os.WriteFile(mf, []byte(manifestFixture), 0o644)

	rep, err := st.SyncFromManifest(context.Background(), mf)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Fotos != 5 {
		t.Errorf("2o sync: %d fotos, quero 5 (replace, nao acumula)", rep.Fotos)
	}
	if st.DataVersion() <= dv1 {
		t.Errorf("data_version nao subiu: %d -> %d", dv1, st.DataVersion())
	}
}

func TestSnapshotFiltraECconta(t *testing.T) {
	st := storeComManifest(t)
	snap, err := st.LoadSnapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got := len(snap.Fotos("")); got != 5 {
		t.Errorf("todas = %d, quero 5", got)
	}
	if got := len(snap.Fotos("epoca")); got != 3 {
		t.Errorf("epoca = %d, quero 3", got)
	}
	if got := len(snap.Fotos("casal")); got != 2 {
		t.Errorf("casal = %d, quero 2", got)
	}
	// tema inventado -> vazio (a galeria transforma isso em 404)
	if got := len(snap.Fotos("nao-existe")); got != 0 {
		t.Errorf("tema invalido devia ser vazio, veio %d", got)
	}
	if snap.TemaValido("nao-existe") {
		t.Error("TemaValido devia ser false para tema inventado")
	}
	if !snap.TemaValido("epoca") {
		t.Error("TemaValido devia ser true para epoca")
	}
}

func TestShuffleDeterministicoNoDia(t *testing.T) {
	st := storeComManifest(t)
	s1, _ := st.LoadSnapshot(context.Background())
	s2, _ := st.LoadSnapshot(context.Background())
	// mesma seed do dia -> mesma ordem entre dois loads. E o que torna a
	// resposta cacheavel (ETag estavel dentro do dia).
	a, b := s1.Fotos(""), s2.Fotos("")
	for i := range a {
		if a[i].Key != b[i].Key {
			t.Fatalf("ordem divergiu na pos %d: %s vs %s -- shuffle nao e estavel no dia",
				i, a[i].Key, b[i].Key)
		}
	}
}
