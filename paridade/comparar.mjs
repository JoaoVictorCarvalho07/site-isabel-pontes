import { existsSync, mkdirSync, readdirSync, readFileSync, writeFileSync } from "node:fs";
import { dirname, join, relative } from "node:path";
import { parseArgs } from "node:util";
import pixelmatch from "pixelmatch";
import { PNG } from "pngjs";

const { values } = parseArgs({
  options: {
    referencia: { type: "string", default: "paridade/referencia" },
    atual: { type: "string", default: "paridade/atual" },
    diff: { type: "string", default: "paridade/diff" },
  },
});

function listar(pasta) {
  return readdirSync(pasta, { withFileTypes: true, recursive: true })
    .filter((item) => item.isFile() && item.name.endsWith(".png"))
    .map((item) => relative(pasta, join(item.parentPath, item.name)));
}

let diferentes = 0;
const arquivos = listar(values.atual);

for (const arquivo of arquivos) {
  const caminhoReferencia = join(values.referencia, arquivo);
  if (!existsSync(caminhoReferencia)) {
    console.log(`SEM REFERENCIA  ${arquivo}`);
    diferentes++;
    continue;
  }
  const referencia = PNG.sync.read(readFileSync(caminhoReferencia));
  const atual = PNG.sync.read(readFileSync(join(values.atual, arquivo)));
  if (referencia.width !== atual.width || referencia.height !== atual.height) {
    console.log(`TAMANHO  ${arquivo}  ${referencia.width}x${referencia.height} -> ${atual.width}x${atual.height}`);
    diferentes++;
    continue;
  }
  const mapa = new PNG({ width: atual.width, height: atual.height });
  const pixels = pixelmatch(referencia.data, atual.data, mapa.data, atual.width, atual.height, { threshold: 0 });
  if (pixels === 0) {
    console.log(`IGUAL  ${arquivo}`);
    continue;
  }
  diferentes++;
  const destino = join(values.diff, arquivo);
  mkdirSync(dirname(destino), { recursive: true });
  writeFileSync(destino, PNG.sync.write(mapa));
  console.log(`DIFERENTE  ${arquivo}  ${pixels} px (${((100 * pixels) / (atual.width * atual.height)).toFixed(4)}%)`);
}

console.log(`${arquivos.length - diferentes} de ${arquivos.length} identicas`);
process.exit(diferentes ? 1 : 0);
