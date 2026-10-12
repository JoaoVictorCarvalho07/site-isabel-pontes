const JANELA_MS = 30_000;

const ultimoEnvio = new Map<string, number>();
let ultimaLimpeza = Date.now();

export function permitir(ip: string, agora = Date.now()): boolean {
  if (agora - ultimaLimpeza > 10 * JANELA_MS) {
    for (const [chave, quando] of ultimoEnvio) {
      if (agora - quando > JANELA_MS) ultimoEnvio.delete(chave);
    }
    ultimaLimpeza = agora;
  }
  const anterior = ultimoEnvio.get(ip);
  if (anterior !== undefined && agora - anterior < JANELA_MS) return false;
  ultimoEnvio.set(ip, agora);
  return true;
}
