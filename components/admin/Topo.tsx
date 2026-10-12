import { sair } from "@/lib/admin/acoes";

const LINKS = [
  { href: "/admin/fotos", rotulo: "Fotos" },
  { href: "/admin/temas", rotulo: "Temas" },
  { href: "/admin/lixeira", rotulo: "Lixeira" },
];

export function Topo({ atual }: { atual: string }) {
  return (
    <nav className="adm-topo" aria-label="Painel">
      {LINKS.map((link) => (
        <a key={link.href} href={link.href} aria-current={link.href === atual ? "page" : undefined}>
          {link.rotulo}
        </a>
      ))}
      <a href="/galeria" target="_blank" rel="noopener">
        Ver o site
      </a>
      <form action={sair} className="adm-topo__sair">
        <button type="submit">Sair</button>
      </form>
    </nav>
  );
}
