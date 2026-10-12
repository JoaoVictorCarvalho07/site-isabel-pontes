export function ehAdmin(email: string | null | undefined): boolean {
  if (!email) return false;
  const permitidos = (process.env.ADMIN_EMAILS ?? "")
    .split(",")
    .map((item) => item.trim().toLowerCase())
    .filter(Boolean);
  return permitidos.includes(email.toLowerCase());
}
