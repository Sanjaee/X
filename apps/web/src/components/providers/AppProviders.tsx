"use client";

import React, { useEffect, useState } from "react";
import { SessionProvider, useSession } from "next-auth/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { RealtimeProvider } from "@/lib/useRealtime";
import { setStoredToken, setStoredUser } from "@/lib/api";
import { MediaViewerProvider } from "@/components/media/MediaViewerContext";
import { PhotoViewerModal } from "@/components/media/PhotoViewerModal";

function NextAuthSync() {
  const { data: session, status } = useSession();

  // Clean stale error query params from previous failed OAuth redirects
  useEffect(() => {
    if (typeof window !== "undefined" && window.location.search.includes("error=OAuthCallback")) {
      try {
        const url = new URL(window.location.href);
        url.searchParams.delete("error");
        url.searchParams.delete("callbackUrl");
        window.history.replaceState({}, document.title, url.pathname + (url.search ? url.search : ""));
      } catch {}
    }
  }, []);

  useEffect(() => {
    if (status === "authenticated" && (session as { sessionToken?: string })?.sessionToken) {
      const token = (session as { sessionToken?: string }).sessionToken;
      const current = typeof window !== "undefined" ? localStorage.getItem("x_session_token") : null;
      if (token && current !== token) {
        setStoredToken(token);
        if ((session as any)?.backendUser) {
          setStoredUser((session as any).backendUser);
        }
        window.dispatchEvent(new Event("x_auth_state_change"));
      }
    }
  }, [session, status]);

  return null;
}

import { ComposerProvider } from "@/components/composer/ComposerContext";
import { UpgradeModalProvider } from "@/components/subscription/UpgradeModalContext";
import { UpgradeModal } from "@/components/subscription/UpgradeModal";

export function AppProviders({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 60 * 1000,
            gcTime: 5 * 60 * 1000,
            retry: 2,
            refetchOnWindowFocus: false,
          },
        },
      })
  );

  return (
    <SessionProvider>
      <QueryClientProvider client={queryClient}>
        <NextAuthSync />
        <RealtimeProvider>
          <MediaViewerProvider>
            <ComposerProvider>
              <UpgradeModalProvider>
                {children}
                <PhotoViewerModal />
                <UpgradeModal />
              </UpgradeModalProvider>
            </ComposerProvider>
          </MediaViewerProvider>
        </RealtimeProvider>
      </QueryClientProvider>
    </SessionProvider>
  );
}
