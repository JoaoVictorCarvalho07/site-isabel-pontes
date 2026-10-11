import { NextResponse, type NextRequest } from "next/server";

export function proxy(requisicao: NextRequest) {
  const cabecalhos = new Headers(requisicao.headers);
  cabecalhos.set("x-caminho", requisicao.nextUrl.pathname);
  return NextResponse.next({ request: { headers: cabecalhos } });
}

export const config = { matcher: ["/((?!_next/|static/).*)"] };
