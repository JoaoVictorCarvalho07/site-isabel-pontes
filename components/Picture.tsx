import { endereco } from "@/lib/layout/imagens";

type Props = { base: string; w: number; h: number; alt: string; className: string };

export function Picture({ base, w, h, alt, className }: Props) {
  return (
    <picture>
      <source type="image/avif" srcSet={endereco(`${base}.avif`)} />
      <img src={endereco(`${base}.webp`)} width={w} height={h} alt={alt} loading="lazy" decoding="async" className={className} />
    </picture>
  );
}
