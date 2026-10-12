import { Briefing, metadadosBriefing, type BuscaBriefing } from "@/components/paginas/Briefing";

export const generateMetadata = () => metadadosBriefing("en");

export default function Pagina({ searchParams }: { searchParams: BuscaBriefing }) {
  return <Briefing idioma="en" busca={searchParams} />;
}
