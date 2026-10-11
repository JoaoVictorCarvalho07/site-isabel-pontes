import { Midiakit, metadadosMidiakit } from "@/components/paginas/Midiakit";

export const generateMetadata = () => metadadosMidiakit("en");

export default function Pagina() {
  return <Midiakit idioma="en" />;
}
