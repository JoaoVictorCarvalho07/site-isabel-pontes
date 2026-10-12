import type { NextRequest } from "next/server";
import { responderEnvio } from "@/lib/briefing/rota";

export async function POST(requisicao: NextRequest, { params }: { params: Promise<{ slug: string }> }) {
  return responderEnvio(requisicao, (await params).slug, "en");
}
