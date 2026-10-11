import { Home, metadadosHome } from "@/components/paginas/Home";

export const generateMetadata = () => metadadosHome("pt");

export default function Pagina() {
  return <Home idioma="pt" />;
}
