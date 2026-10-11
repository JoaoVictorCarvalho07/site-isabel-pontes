import { readFileSync } from "node:fs";
import { PNG } from "pngjs";

for (const arquivo of process.argv.slice(2)) {
  const a = PNG.sync.read(readFileSync(`paridade/referencia/${arquivo}`));
  const b = PNG.sync.read(readFileSync(`paridade/atual/${arquivo}`));
  let x0 = Infinity, y0 = Infinity, x1 = -1, y1 = -1, total = 0;
  for (let y = 0; y < Math.min(a.height, b.height); y++) {
    for (let x = 0; x < a.width; x++) {
      const i = (y * a.width + x) * 4;
      if (a.data[i] !== b.data[i] || a.data[i + 1] !== b.data[i + 1] || a.data[i + 2] !== b.data[i + 2]) {
        total++;
        x0 = Math.min(x0, x); y0 = Math.min(y0, y); x1 = Math.max(x1, x); y1 = Math.max(y1, y);
      }
    }
  }
  console.log(`${arquivo}  ${a.width}x${a.height}  ${total} px  x ${x0}-${x1}  y ${y0}-${y1}`);
}
