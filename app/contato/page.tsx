import { Contato, metadadosContato } from "@/components/paginas/Contato";

export const generateMetadata = () => metadadosContato("pt");

export default function Pagina() {
  return <Contato idioma="pt" />;
}
