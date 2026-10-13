import { Privacidade, metadadosPrivacidade } from "@/components/paginas/Privacidade";

export const generateMetadata = () => metadadosPrivacidade("en");

export default function Pagina() {
  return <Privacidade idioma="en" />;
}
