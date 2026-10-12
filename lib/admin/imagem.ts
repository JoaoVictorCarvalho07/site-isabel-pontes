import sharp from "sharp";

export type Variante = { pasta: string; largura: number; qualidade: number };

export const VARIANTES: Variante[] = [
  { pasta: "grid400", largura: 400, qualidade: 50 },
  { pasta: "grid800", largura: 800, qualidade: 55 },
  { pasta: "full1600", largura: 1600, qualidade: 60 },
];

export type Processada = { w: number; h: number; lqip: string; variantes: { pasta: string; bytes: Buffer }[] };

export async function processarImagem(original: Uint8Array): Promise<Processada> {
  const abrir = () => sharp(original, { failOn: "error" }).rotate();
  const { width = 0, height = 0, orientation = 1 } = await sharp(original).metadata();
  const deitada = orientation >= 5;
  const variantes = [];
  for (const { pasta, largura, qualidade } of VARIANTES) {
    const bytes = await abrir().resize({ width: largura, withoutEnlargement: true }).avif({ quality: qualidade, effort: 4 }).toBuffer();
    variantes.push({ pasta, bytes });
  }
  const miniatura = await abrir().resize({ width: 24 }).webp({ quality: 40 }).toBuffer();
  return {
    w: deitada ? height : width,
    h: deitada ? width : height,
    lqip: `data:image/webp;base64,${miniatura.toString("base64")}`,
    variantes,
  };
}
