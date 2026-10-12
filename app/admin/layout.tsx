import type { Metadata } from "next";
import type { ReactNode } from "react";
import "./admin.scss";

export const metadata: Metadata = { title: "Painel — Isabel Pontes", robots: { index: false, follow: false } };

export default function LayoutAdmin({ children }: { children: ReactNode }) {
  return <div className="adm">{children}</div>;
}
