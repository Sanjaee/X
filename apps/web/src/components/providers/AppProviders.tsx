"use client";

import React, { useEffect, useState } from "react";
import { SessionProvider, useSession } from "next-auth/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { RealtimeProvider } from "@/lib/useRealtime";
import { setStoredToken, setStoredUser } from "@/lib/api";
import { MediaViewerProvider } from "@/components/media/MediaViewerContext";
import { PhotoViewerModal } from "@/components/media/PhotoViewerModal";
import { ComposerProvider } from "@/components/composer/ComposerContext";
import { UpgradeModalProvider } from "@/components/subscription/UpgradeModalContext";
import { UpgradeModal } from "@/components/subscription/UpgradeModal";
import { AuthModal } from "@/components/auth/AuthModal";

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
    if (status === "authenticated") {
      const sess = session as any;
      const currentToken = typeof window !== "undefined" ? localStorage.getItem("x_session_token") : null;

      // If user already has an active backend session token, do not prompt for TOTP or overwrite session
      if (currentToken) {
        return;
      }

      // If user has 2FA enabled, backend returned requires_totp and an mfa_token
      if (sess?.requiresTotp && sess?.mfaToken) {
        const handledToken = typeof sessionStorage !== "undefined" ? sessionStorage.getItem("x_handled_mfa_token") : null;
        if (handledToken === sess.mfaToken) {
          return;
        }

        setTimeout(() => {
          if (typeof window !== "undefined" && localStorage.getItem("x_session_token")) {
            return;
          }
          window.dispatchEvent(
            new CustomEvent("x_totp_challenge_required", {
              detail: {
                mfa_token: sess.mfaToken,
                username: sess.username || "Google User",
              },
            })
          );
        }, 100);
        return;
      }

      // If user does not require TOTP, backend directly returned sessionToken
      if (sess?.sessionToken) {
        const token = sess.sessionToken;
        setStoredToken(token);
        if (sess?.backendUser) {
          setStoredUser(sess.backendUser);
        }
        window.dispatchEvent(new Event("x_auth_state_change"));
      }
    }
  }, [session, status]);

  return null;
}

function GlobalAuthModal() {
  const [isOpen, setIsOpen] = useState(false);
  const [initialMode, setInitialMode] = useState<"login" | "register">("login");
  const [mfaToken, setMfaToken] = useState<string | null>(null);

  useEffect(() => {
    const handleOpen = (e: Event) => {
      // NEVER show login popup if user is already logged in
      if (typeof window !== "undefined" && localStorage.getItem("x_session_token")) {
        return;
      }
      const mode = (e as CustomEvent)?.detail?.mode || "login";
      setInitialMode(mode);
      setMfaToken(null);
      setIsOpen(true);
    };

    const handleChallenge = (e: Event) => {
      // NEVER show login popup if user is already logged in
      if (typeof window !== "undefined" && localStorage.getItem("x_session_token")) {
        return;
      }
      const detail = (e as CustomEvent)?.detail;
      if (detail?.mfa_token) {
        setMfaToken(detail.mfa_token);
      }
      setInitialMode("login");
      setIsOpen(true);
    };

    window.addEventListener("x_open_auth_modal", handleOpen);
    window.addEventListener("x_totp_challenge_required", handleChallenge);
    return () => {
      window.removeEventListener("x_open_auth_modal", handleOpen);
      window.removeEventListener("x_totp_challenge_required", handleChallenge);
    };
  }, []);

  if (!isOpen) return null;

  return (
    <AuthModal
      isOpen={isOpen}
      initialMfaToken={mfaToken}
      onClose={() => {
        setIsOpen(false);
        setMfaToken(null);
      }}
      onSuccess={() => {
        setIsOpen(false);
        setMfaToken(null);
        window.dispatchEvent(new Event("x_auth_state_change"));
      }}
      initialMode={initialMode}
    />
  );
}

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
                <GlobalAuthModal />
              </UpgradeModalProvider>
            </ComposerProvider>
          </MediaViewerProvider>
        </RealtimeProvider>
      </QueryClientProvider>
    </SessionProvider>
  );
}
