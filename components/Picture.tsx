import { asset } from "@/lib/asset";

type Props = { base: string; w: number; h: number; alt: string; className: string };

export function Picture({ base, w, h, alt, className }: Props) {
  return (
    <picture>
      <source type="image/avif" srcSet={asset(`${base}.avif`)} />
      <img src={asset(`${base}.webp`)} width={w} height={h} alt={alt} loading="lazy" decoding="async" className={className} />
    </picture>
  );
}
