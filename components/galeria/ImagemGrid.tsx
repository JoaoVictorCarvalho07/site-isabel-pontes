"use client";

type Props = { grid: string; grid2x: string; w: number; h: number; alt: string };

function revelar(img: HTMLImageElement | null) {
  if (!img) return;
  if (img.complete) img.classList.add("is-loaded");
  else img.addEventListener("load", () => img.classList.add("is-loaded"), { once: true });
}

export function ImagemGrid({ grid, grid2x, w, h, alt }: Props) {
  return (
    <img
      ref={revelar}
      className="masonry__img"
      src={grid}
      srcSet={`${grid} 400w, ${grid2x} 800w`}
      sizes="(max-width:480px) 100vw, (max-width:750px) 50vw, (max-width:1100px) 33vw, 25vw"
      width={w}
      height={h}
      alt={alt}
      loading="lazy"
      decoding="async"
    />
  );
}
