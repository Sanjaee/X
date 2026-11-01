import NextAuth from "next-auth";
import { authOptions } from "@/lib/auth";
import { NextRequest } from "next/server";

function getDynamicAuthOptions(req: NextRequest) {
  const host = req.headers.get("x-forwarded-host") || req.headers.get("host") || "";
  const forwardedProto = req.headers.get("x-forwarded-proto");
  const isLocal = host.includes("localhost") || host.includes("127.0.0.1");
  const proto = forwardedProto || (isLocal ? "http" : "https");
  
  // Resolve strict canonical origin based on host:
  // If local, origin is strictly http://localhost (or current local port)
  // If prod, origin is strictly https://lostmedia.pro
  const origin = isLocal
    ? `${proto}://${host}`
    : (process.env.NEXTAUTH_URL && !process.env.NEXTAUTH_URL.includes("localhost")
        ? process.env.NEXTAUTH_URL
        : "https://lostmedia.pro");

  // Keep NextAuth runtime environment in sync for this request
  process.env.NEXTAUTH_URL = origin;
  process.env.AUTH_URL = origin;

  return {
    ...authOptions,
    callbacks: {
      ...authOptions.callbacks,
      async redirect({ url, baseUrl }: { url: string; baseUrl: string }) {
        if (url.startsWith("/")) {
          return `${origin}${url}`;
        }
        try {
          const parsedUrl = new URL(url);
          const parsedOrigin = new URL(origin);
          if (parsedUrl.origin === parsedOrigin.origin) {
            return url;
          }
        } catch {}
        return origin;
      },
    },
  };
}

const handler = async (req: NextRequest, ctx: any) => {
  return NextAuth(req, ctx, getDynamicAuthOptions(req));
};

export { handler as GET, handler as POST };
