import { Home, metadadosHome } from "@/components/paginas/Home";

export const generateMetadata = () => metadadosHome("en");

export default function Pagina() {
  return <Home idioma="en" />;
}
