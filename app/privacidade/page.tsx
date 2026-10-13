import { Privacidade, metadadosPrivacidade } from "@/components/paginas/Privacidade";

export const generateMetadata = () => metadadosPrivacidade("pt");

export default function Pagina() {
  return <Privacidade idioma="pt" />;
}
