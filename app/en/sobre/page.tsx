import { Sobre, metadadosSobre } from "@/components/paginas/Sobre";

export const generateMetadata = () => metadadosSobre("en");

export default function Pagina() {
  return <Sobre idioma="en" />;
}
