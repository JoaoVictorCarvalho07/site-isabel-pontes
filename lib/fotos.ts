import { banco } from "./banco";

export type Foto = {
  id: number;
  tema: string;
  w: number;
  h: number;
  grid: string;
  grid2x: string;
  full: string;
  lqip: string;
  alt: string;
};

export function fotosPublicadas(): Foto[] {
  return banco()
    .prepare(
      `SELECT id, tema_slug AS tema, w, h, grid, grid2x, full, lqip, alt
         FROM photos
        WHERE published = 1
        ORDER BY tema_slug, sort_order IS NULL, sort_order, id`,
    )
    .all() as Foto[];
}
