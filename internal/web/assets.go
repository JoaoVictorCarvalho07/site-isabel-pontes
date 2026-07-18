package web

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/fs"
	"sync"
)

// assetHasher resolve "/static/css/app.css" -> "/static/css/app.css?v=1a2b3c4d".
//
// Existe para tornar honesto o `Cache-Control: immutable, max-age=1 ano` do
// middleware: sem um identificador de conteudo na URL, um visitante recorrente
// ficaria preso ao CSS antigo por um ano depois de cada deploy. Com o hash, o
// arquivo novo tem URL nova e o cache velho simplesmente deixa de ser pedido.
//
// O hash e calculado sob demanda e memoizado. Os arquivos vem do //go:embed,
// entao nunca mudam durante a vida do processo -- cachear e seguro por
// construcao, sem invalidacao.
type assetHasher struct {
	fsys fs.FS
	mu   sync.RWMutex
	memo map[string]string
}

func newAssetHasher(fsys fs.FS) *assetHasher {
	return &assetHasher{fsys: fsys, memo: map[string]string{}}
}

// url e exposta aos templates como {{asset "/static/css/app.css"}}.
//
// Se o arquivo nao existir, devolve o caminho original em vez de falhar: um
// asset faltando deve degradar para um 404 daquele recurso, nao derrubar a
// pagina inteira. O log denuncia.
func (a *assetHasher) url(caminho string) string {
	a.mu.RLock()
	urlCacheada, ok := a.memo[caminho]
	a.mu.RUnlock()
	if ok {
		return urlCacheada
	}

	// o FS embutido nao usa a barra inicial ("static/..." e nao "/static/...")
	nomeNoFS := caminho
	if len(nomeNoFS) > 0 && nomeNoFS[0] == '/' {
		nomeNoFS = nomeNoFS[1:]
	}

	arquivo, err := a.fsys.Open(nomeNoFS)
	if err != nil {
		logger.Warn("asset nao encontrado", "caminho", caminho, "erro", err)
		return caminho
	}
	defer arquivo.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, arquivo); err != nil {
		logger.Warn("falha ao ler asset", "caminho", caminho, "erro", err)
		return caminho
	}
	urlComHash := caminho + "?v=" + hex.EncodeToString(hash.Sum(nil))[:8]

	a.mu.Lock()
	a.memo[caminho] = urlComHash
	a.mu.Unlock()
	return urlComHash
}
