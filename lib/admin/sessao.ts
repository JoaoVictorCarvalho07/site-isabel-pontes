import { redirect } from "next/navigation";
import { auth } from "@/auth";
import { ehAdmin } from "./permitidos";

export async function emailAdmin(): Promise<string | null> {
  const sessao = await auth();
  const email = sessao?.user?.email;
  return ehAdmin(email) ? (email as string) : null;
}

export async function exigirAdmin(): Promise<string> {
  const email = await emailAdmin();
  if (!email) redirect("/admin/entrar");
  return email;
}
