import { Midiakit, metadadosMidiakit, type BuscaMidiakit } from "@/components/paginas/Midiakit";

export const generateMetadata = () => metadadosMidiakit("pt");

export default function Pagina({ searchParams }: { searchParams: BuscaMidiakit }) {
  return <Midiakit idioma="pt" busca={searchParams} />;
}
