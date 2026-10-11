import { Sobre, metadadosSobre } from "@/components/paginas/Sobre";

export const generateMetadata = () => metadadosSobre("pt");

export default function Pagina() {
  return <Sobre idioma="pt" />;
}
