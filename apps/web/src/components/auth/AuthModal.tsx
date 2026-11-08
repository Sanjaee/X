"use client";

import React, { useState, useEffect } from "react";
import { createPortal } from "react-dom";
import { X, Loader2, ShieldCheck } from "lucide-react";
import { signIn } from "next-auth/react";
import { apiRequest, setStoredToken, setStoredUser } from "@/lib/api";
import { User, AuthResponse } from "@/types";

interface AuthModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: (user: User) => void;
  initialMode?: "login" | "register";
  initialMfaToken?: string | null;
}

export function AuthModal({
  isOpen,
  onClose,
  onSuccess,
  initialMfaToken = null,
}: AuthModalProps) {
  const [mfaToken, setMfaToken] = useState<string | null>(initialMfaToken);
  const [totpCode, setTotpCode] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [googleLoading, setGoogleLoading] = useState(false);

  useEffect(() => {
    if (isOpen) {
      setMfaToken(initialMfaToken);
      setTotpCode("");
      setError(null);
    }
  }, [isOpen, initialMfaToken]);

  // Listen for TOTP challenge required from Google OAuth flow
  useEffect(() => {
    const handleChallenge = (e: Event) => {
      const detail = (e as CustomEvent).detail;
      if (detail?.mfa_token) {
        setMfaToken(detail.mfa_token);
        setError(null);
      }
    };
    window.addEventListener("x_totp_challenge_required", handleChallenge);
    return () => window.removeEventListener("x_totp_challenge_required", handleChallenge);
  }, []);

  if (!isOpen || typeof document === "undefined") return null;

  const handleGoogleSignIn = async () => {
    try {
      setGoogleLoading(true);
      setError(null);
      const cleanCallback = window.location.origin + (window.location.pathname === "/" ? "" : window.location.pathname);
      await signIn("google", { callbackUrl: cleanCallback || "/" });
    } catch (err: any) {
      setError(err?.message || "Failed to initialize Google authentication");
      setGoogleLoading(false);
    }
  };

  const handleMfaPaste = (e: React.ClipboardEvent<HTMLInputElement>) => {
    const pasted = e.clipboardData.getData("text");
    if (pasted) {
      // If user pasted multiple lines or tokens, pick the first valid code token
      const parts = pasted.split(/[\r\n,]+/).map((s) => s.trim()).filter(Boolean);
      if (parts.length > 0) {
        e.preventDefault();
        setTotpCode(parts[0]);
      }
    }
  };

  const handleVerifyMfa = async (e: React.FormEvent) => {
    e.preventDefault();
    const cleaned = totpCode.trim();
    if (!cleaned || !mfaToken) return;

    setError(null);
    setLoading(true);

    try {
      const res = await apiRequest<AuthResponse>("/auth/login/totp", {
        method: "POST",
        body: JSON.stringify({
          mfa_token: mfaToken,
          code: cleaned,
        }),
      });

      if (res.session_token && res.user) {
        if (typeof sessionStorage !== "undefined") {
          sessionStorage.setItem("x_handled_mfa_token", mfaToken);
        }
        setStoredToken(res.session_token);
        setStoredUser(res.user);
        onSuccess(res.user);
        onClose();
      }
    } catch (err: any) {
      setError(err?.message || "Invalid verification code or backup code.");
    } finally {
      setLoading(false);
    }
  };

  const handleCancelMfa = () => {
    setMfaToken(null);
    setTotpCode("");
    setError(null);
  };

  return createPortal(
    <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/75 backdrop-blur-sm p-4 animate-in fade-in">
      <div
        style={{ backgroundColor: "var(--background, #000000)" }}
        className="relative w-full max-w-[440px] bg-[#000000] border border-zinc-800 rounded-3xl shadow-2xl p-6 sm:p-8 flex flex-col items-center"
      >
        {/* Close Button */}
        <button
          onClick={onClose}
          aria-label="Close modal"
          className="absolute top-4 left-4 p-2 hover:bg-zinc-800/60 rounded-full text-foreground transition-colors"
        >
          <X className="w-5 h-5 text-zinc-300" />
        </button>

        {/* 2FA MFA CHALLENGE SCREEN */}
        {mfaToken ? (
          <div className="w-full flex flex-col items-center pt-2">
            <h2 className="text-2xl sm:text-[26px] font-extrabold text-white text-center leading-tight">
              Welcome back
            </h2>
            <p className="text-xs sm:text-sm text-zinc-400 text-center mt-1.5 mb-6">
              Verify your account security
            </p>

            {/* Green Shield Information Card */}
            <div className="w-full mb-5 p-4 rounded-2xl bg-emerald-950/20 border border-emerald-500/20 text-left space-y-1.5">
              <div className="flex items-center gap-2">
                <ShieldCheck className="w-4 h-4 text-emerald-400 shrink-0" />
                <span className="text-xs font-bold text-white">Two-Step Verification</span>
              </div>
              <p className="text-[11px] text-zinc-400 leading-relaxed">
                Enter the 6-digit code from <strong className="text-zinc-200">Google Authenticator</strong>, or use one of your 10 <strong className="text-zinc-200">Backup Codes</strong> (e.g.{" "}
                <code className="text-emerald-400 font-mono font-semibold">58B5-A215</code>) or your 12-word seed phrase.
              </p>
            </div>

            {error && (
              <div className="w-full mb-4 p-3 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-xs text-center">
                {error}
              </div>
            )}

            <form onSubmit={handleVerifyMfa} className="w-full flex flex-col space-y-4">
              <div>
                <label className="text-xs font-semibold text-zinc-300 block mb-2">
                  TOTP Code or Backup Code
                </label>
                <input
                  type="text"
                  autoFocus
                  required
                  autoComplete="one-time-code"
                  value={totpCode}
                  onChange={(e) => setTotpCode(e.target.value)}
                  onPaste={handleMfaPaste}
                  placeholder="123456 OR 58B5-A215"
                  className="w-full text-center text-base font-mono tracking-wider px-3.5 py-3 bg-zinc-900/80 border border-zinc-800 rounded-xl text-white placeholder:text-zinc-600 focus:outline-none focus:border-zinc-500 uppercase transition-colors"
                />
                <p className="text-[10px] text-zinc-500 text-center mt-2">
                  *Pasting a list of codes will automatically take the first code.
                </p>
              </div>

              <button
                type="submit"
                disabled={loading || !totpCode.trim()}
                className={`w-full py-3.5 px-4 rounded-xl font-bold transition-all flex items-center justify-center gap-2 shadow-sm ${
                  totpCode.trim()
                    ? "bg-white text-black hover:bg-neutral-200 active:scale-[0.99] cursor-pointer"
                    : "bg-[#71767b]/50 text-zinc-400 cursor-not-allowed opacity-40"
                }`}
              >
                {loading ? (
                  <Loader2 className={`w-4 h-4 animate-spin ${totpCode.trim() ? "text-black" : "text-zinc-400"}`} />
                ) : (
                  <ShieldCheck className={`w-4 h-4 ${totpCode.trim() ? "text-black" : "text-zinc-400"}`} />
                )}
                <span className="text-sm">Verify &amp; Log In</span>
              </button>

              <button
                type="button"
                onClick={handleCancelMfa}
                className="text-xs text-zinc-400 hover:text-white hover:underline transition-colors text-center pt-1"
              >
                Cancel and return to sign in
              </button>
            </form>

            <div className="mt-6 text-xs text-zinc-400 text-center">
              Don&apos;t have an account?{" "}
              <button
                type="button"
                onClick={handleGoogleSignIn}
                className="text-white font-bold hover:underline"
              >
                Sign up
              </button>
            </div>
          </div>
        ) : (
          /* STANDARD GOOGLE SIGN-IN SCREEN (No email/password credentials) */
          <div className="w-full flex flex-col items-center">
            {/* X Logo */}
            <div className="mb-5 text-white">
              <svg viewBox="0 0 24 24" aria-hidden="true" className="w-9 h-9 fill-current">
                <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z" />
              </svg>
            </div>

            <h2 className="text-2xl sm:text-[26px] font-extrabold mb-2 text-center text-white leading-tight">
              Sign in to X
            </h2>
            <p className="text-xs sm:text-sm text-zinc-400 text-center mb-6 max-w-xs">
              Welcome to X. Please sign in with your Google account to continue.
            </p>

            {error && (
              <div className="w-full mb-4 p-3 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-xs text-center">
                {error}
              </div>
            )}

            {/* Official Google Sign-In Button */}
            <div className="w-full mb-5">
              <button
                type="button"
                onClick={handleGoogleSignIn}
                disabled={googleLoading}
                className="w-full flex items-center justify-center gap-3 bg-white hover:bg-neutral-100 active:scale-[0.99] text-[#0f1419] font-bold text-[15px] h-[46px] px-5 rounded-full transition-all shadow-md select-none disabled:opacity-50"
              >
                {googleLoading ? (
                  <Loader2 className="w-5 h-5 animate-spin text-neutral-600" />
                ) : (
                  <svg className="w-5 h-5 shrink-0" viewBox="0 0 24 24">
                    <path
                      fill="#4285F4"
                      d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                    />
                    <path
                      fill="#34A853"
                      d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                    />
                    <path
                      fill="#FBBC05"
                      d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.06H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.94l2.85-2.22.81-.63z"
                    />
                    <path
                      fill="#EA4335"
                      d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.06l3.66 2.84c.87-2.6 3.3-4.52 6.16-4.52z"
                    />
                  </svg>
                )}
                <span>Sign in with Google</span>
              </button>
            </div>

            <p className="text-[11px] text-zinc-500 text-center leading-relaxed px-2">
              By signing in, you agree to the Terms of Service and Privacy Policy, including Cookie Use.
            </p>
          </div>
        )}
      </div>
    </div>,
    document.body
  );
}
