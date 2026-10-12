export const TAMANHO_MAXIMO_ARQUIVO = 5 * 1024 * 1024;
export const MAXIMO_ARQUIVOS = 4;

const ACEITOS: Record<string, string> = {
  "image/jpeg": ".jpg",
  "image/png": ".png",
  "image/webp": ".webp",
  "image/gif": ".gif",
  "image/avif": ".avif",
  "image/heic": ".heic",
  "image/heif": ".heif",
};

const MARCAS_FTYP: Record<string, string> = {
  heic: "image/heic",
  heix: "image/heic",
  hevc: "image/heic",
  heim: "image/heic",
  heis: "image/heic",
  mif1: "image/heif",
  msf1: "image/heif",
  avif: "image/avif",
  avis: "image/avif",
};

function comeca(bytes: Uint8Array, assinatura: number[], inicio = 0): boolean {
  return assinatura.every((byte, i) => bytes[inicio + i] === byte);
}

function texto(bytes: Uint8Array, inicio: number, fim: number): string {
  return String.fromCharCode(...bytes.subarray(inicio, fim));
}

export function farejarTipo(bytes: Uint8Array): { mime: string; extensao: string } | null {
  let mime = "";
  if (comeca(bytes, [0xff, 0xd8, 0xff])) mime = "image/jpeg";
  else if (comeca(bytes, [0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a])) mime = "image/png";
  else if (texto(bytes, 0, 6) === "GIF87a" || texto(bytes, 0, 6) === "GIF89a") mime = "image/gif";
  else if (texto(bytes, 0, 4) === "RIFF" && texto(bytes, 8, 14) === "WEBPVP") mime = "image/webp";
  else if (bytes.length >= 12 && texto(bytes, 4, 8) === "ftyp") mime = MARCAS_FTYP[texto(bytes, 8, 12)] ?? "";
  const extensao = ACEITOS[mime];
  return extensao ? { mime, extensao } : null;
}

export function nomeParaExibir(nome: string): string {
  const base = (nome.replaceAll("\\", "/").split("/").pop() ?? "").replace(/[\x00-\x1f\x7f]/g, "").slice(0, 120);
  return base === "" || base === "." || base === ".." ? "imagem" : base;
}
