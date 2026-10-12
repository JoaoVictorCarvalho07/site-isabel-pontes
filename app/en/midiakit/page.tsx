import { Midiakit, metadadosMidiakit, type BuscaMidiakit } from "@/components/paginas/Midiakit";

export const generateMetadata = () => metadadosMidiakit("en");

export default function Pagina({ searchParams }: { searchParams: BuscaMidiakit }) {
  return <Midiakit idioma="en" busca={searchParams} />;
}
