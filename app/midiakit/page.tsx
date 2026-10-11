import { Midiakit, metadadosMidiakit } from "@/components/paginas/Midiakit";

export const generateMetadata = () => metadadosMidiakit("pt");

export default function Pagina() {
  return <Midiakit idioma="pt" />;
}
