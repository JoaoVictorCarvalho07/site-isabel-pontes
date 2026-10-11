import { Galeria, metadadosGaleria, type Busca } from "@/components/paginas/Galeria";

type Props = { searchParams: Busca };

export const generateMetadata = ({ searchParams }: Props) => metadadosGaleria("en", searchParams);

export default function Pagina({ searchParams }: Props) {
  return <Galeria idioma="en" busca={searchParams} />;
}
