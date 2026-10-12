import { Briefing, metadadosBriefing, type BuscaBriefing } from "@/components/paginas/Briefing";

export const generateMetadata = () => metadadosBriefing("pt");

export default function Pagina({ searchParams }: { searchParams: BuscaBriefing }) {
  return <Briefing idioma="pt" busca={searchParams} />;
}
