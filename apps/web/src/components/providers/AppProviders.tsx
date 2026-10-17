"use client";

import React, { useEffect, useState } from "react";
import { SessionProvider, useSession } from "next-auth/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { RealtimeProvider } from "@/lib/useRealtime";
import { setStoredToken } from "@/lib/api";
import { MediaViewerProvider } from "@/components/media/MediaViewerContext";
import { PhotoViewerModal } from "@/components/media/PhotoViewerModal";

function NextAuthSync() {
  const { data: session, status } = useSession();

  useEffect(() => {
    if (status === "authenticated" && (session as { sessionToken?: string })?.sessionToken) {
      const token = (session as { sessionToken?: string }).sessionToken;
      const current = typeof window !== "undefined" ? localStorage.getItem("x_session_token") : null;
      if (token && current !== token) {
        setStoredToken(token);
        window.dispatchEvent(new Event("x_auth_state_change"));
      }
    }
  }, [session, status]);

  return null;
}

import { ComposerProvider } from "@/components/composer/ComposerContext";

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
              {children}
              <PhotoViewerModal />
            </ComposerProvider>
          </MediaViewerProvider>
        </RealtimeProvider>
      </QueryClientProvider>
    </SessionProvider>
  );
}
