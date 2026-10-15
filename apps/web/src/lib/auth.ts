import { NextAuthOptions } from "next-auth";
import GoogleProvider from "next-auth/providers/google";

export const authOptions: NextAuthOptions = {
  providers: [
    GoogleProvider({
      clientId: process.env.GOOGLE_CLIENT_ID || "dummy_google_client_id",
      clientSecret: process.env.GOOGLE_CLIENT_SECRET || "dummy_google_client_secret",
    }),
  ],
  secret: process.env.NEXTAUTH_SECRET || "D8D3DA7A75F61ACD5A4CD579EDBBC",
  callbacks: {
    async signIn({ user, account }) {
      if (account?.provider === "google" && user?.email) {
        try {
          // Sync with Go backend API to register/login user and generate backend session_token
          const candidates = [
            process.env.INTERNAL_API_URL,
            process.env.INTERNAL_BACKEND_URL,
            process.env.BACKEND_URL,
            "http://api:8080/api/v1",
            "http://localhost:8080/api/v1",
          ].filter(Boolean) as string[];

          let response: Response | null = null;
          for (const baseUrl of candidates) {
            try {
              const url = baseUrl.endsWith("/api/v1")
                ? `${baseUrl}/auth/google`
                : `${baseUrl}/api/v1/auth/google`;
              const res = await fetch(url, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                  email: user.email,
                  display_name: user.name || user.email.split("@")[0],
                  avatar_url: user.image || "",
                  google_id: account.providerAccountId,
                }),
              });
              if (res.ok) {
                response = res;
                break;
              }
            } catch (err) {
              // Try next candidate url
            }
          }

          if (response && response.ok) {
            const data = await response.json();
            (user as any).sessionToken = data.session_token;
            (user as any).backendUser = data.user;
            return true;
          }
        } catch (error) {
          console.error("Backend Google auth exchange error:", error);
        }
      }
      return true;
    },
    async jwt({ token, user, account }) {
      if (user && (user as any).sessionToken) {
        token.sessionToken = (user as any).sessionToken;
        token.backendUser = (user as any).backendUser;
      }
      if (!token.sessionToken && account?.provider === "google" && token?.email) {
        try {
          const candidates = [
            process.env.INTERNAL_API_URL,
            process.env.INTERNAL_BACKEND_URL,
            process.env.BACKEND_URL,
            "http://api:8080/api/v1",
            "http://localhost:8080/api/v1",
          ].filter(Boolean) as string[];

          for (const baseUrl of candidates) {
            try {
              const url = baseUrl.endsWith("/api/v1")
                ? `${baseUrl}/auth/google`
                : `${baseUrl}/api/v1/auth/google`;
              const res = await fetch(url, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                  email: token.email,
                  display_name: token.name || (token.email as string).split("@")[0],
                  avatar_url: token.picture || "",
                  google_id: account.providerAccountId,
                }),
              });
              if (res.ok) {
                const data = await res.json();
                token.sessionToken = data.session_token;
                token.backendUser = data.user;
                break;
              }
            } catch (e) {}
          }
        } catch (e) {}
      }
      return token;
    },
    async session({ session, token }) {
      if (token) {
        (session as any).sessionToken = token.sessionToken;
        (session as any).backendUser = token.backendUser;
      }
      return session;
    },
  },
  pages: {
    signIn: "/",
    error: "/",
  },
};
