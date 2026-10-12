import NextAuth from "next-auth";
import Google from "next-auth/providers/google";
import { ehAdmin } from "@/lib/admin/permitidos";

export const { handlers, auth, signIn, signOut } = NextAuth({
  providers: [Google],
  session: { strategy: "jwt", maxAge: 60 * 60 * 24 * 30 },
  trustHost: true,
  pages: { signIn: "/admin/entrar", error: "/admin/entrar" },
  callbacks: {
    signIn: ({ profile }) => profile?.email_verified === true && ehAdmin(profile.email),
  },
});
