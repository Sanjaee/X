"use client";

import React, { useEffect } from "react";
import { SessionProvider, useSession } from "next-auth/react";
import { RealtimeProvider } from "@/lib/useRealtime";
import { setStoredToken } from "@/lib/api";

function NextAuthSync() {
  const { data: session, status } = useSession();

  useEffect(() => {
    if (status === "authenticated" && (session as any)?.sessionToken) {
      const token = (session as any).sessionToken;
      const current = typeof window !== "undefined" ? localStorage.getItem("x_session_token") : null;
      if (current !== token) {
        setStoredToken(token);
        window.dispatchEvent(new Event("x_auth_state_change"));
      }
    }
  }, [session, status]);

  return null;
}

export function AppProviders({ children }: { children: React.ReactNode }) {
  return (
    <SessionProvider>
      <NextAuthSync />
      <RealtimeProvider>{children}</RealtimeProvider>
    </SessionProvider>
  );
}
