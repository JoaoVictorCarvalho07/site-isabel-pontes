import { Contato, metadadosContato } from "@/components/paginas/Contato";

export const generateMetadata = () => metadadosContato("en");

export default function Pagina() {
  return <Contato idioma="en" />;
}
