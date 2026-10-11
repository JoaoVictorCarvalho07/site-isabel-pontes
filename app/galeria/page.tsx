import { Galeria, metadadosGaleria, type Busca } from "@/components/paginas/Galeria";

type Props = { searchParams: Busca };

export const generateMetadata = ({ searchParams }: Props) => metadadosGaleria("pt", searchParams);

export default function Pagina({ searchParams }: Props) {
  return <Galeria idioma="pt" busca={searchParams} />;
}
