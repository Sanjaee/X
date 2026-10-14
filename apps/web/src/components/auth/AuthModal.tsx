"use client";

import React, { useState } from "react";
import { X, Loader2 } from "lucide-react";
import { signIn } from "next-auth/react";
import { apiRequest, setStoredToken } from "@/lib/api";
import { User, AuthResponse } from "@/types";

interface AuthModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: (user: User) => void;
}

export function AuthModal({ isOpen, onClose, onSuccess }: AuthModalProps) {
  const [mode, setMode] = useState<"login" | "register">("login");
  const [identifier, setIdentifier] = useState("");
  const [password, setPassword] = useState("");
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [googleLoading, setGoogleLoading] = useState(false);

  if (!isOpen) return null;

  const handleGoogleSignIn = async () => {
    try {
      setGoogleLoading(true);
      setError(null);
      await signIn("google", { callbackUrl: window.location.href });
    } catch (err: any) {
      setError(err.message || "Failed to initialize Google authentication");
      setGoogleLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      if (mode === "login") {
        const res = await apiRequest<AuthResponse>("/auth/login", {
          method: "POST",
          body: JSON.stringify({ identifier, password }),
        });
        setStoredToken(res.session_token);
        onSuccess(res.user);
        onClose();
      } else {
        const res = await apiRequest<AuthResponse>("/auth/register", {
          method: "POST",
          body: JSON.stringify({
            username,
            email,
            password,
            display_name: displayName || username,
          }),
        });
        setStoredToken(res.session_token);
        onSuccess(res.user);
        onClose();
      }
    } catch (err: any) {
      setError(err.message || "Authentication failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
      <div className="relative w-full max-w-[440px] bg-background border border-border rounded-2xl shadow-2xl p-6 sm:p-8 flex flex-col items-center">
        {/* Close Button */}
        <button
          onClick={onClose}
          className="absolute top-4 left-4 p-2 hover:bg-subtle rounded-full text-foreground transition-colors"
        >
          <X className="w-5 h-5" />
        </button>

        {/* X Logo */}
        <div className="mb-6 text-foreground">
          <svg viewBox="0 0 24 24" aria-hidden="true" className="w-9 h-9 fill-current">
            <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z" />
          </svg>
        </div>

        <h2 className="text-[26px] font-extrabold mb-6 text-center leading-tight">
          {mode === "login" ? "Sign in to X" : "Join X today"}
        </h2>

        {error && (
          <div className="w-full mb-4 p-3 rounded-lg bg-red-500/10 border border-red-500/30 text-red-500 text-sm text-center">
            {error}
          </div>
        )}

        {/* X-Standard Google Sign-In Button */}
        <div className="w-full mb-4">
          <button
            type="button"
            onClick={handleGoogleSignIn}
            disabled={googleLoading || loading}
            className="w-full flex items-center justify-center gap-3 bg-white hover:bg-neutral-200 text-[#0f1419] font-semibold text-[15px] h-[42px] px-4 rounded-full border border-neutral-300 transition-colors shadow-sm select-none disabled:opacity-50"
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
            <span>{mode === "login" ? "Sign in with Google" : "Sign up with Google"}</span>
          </button>
        </div>

        {/* X Horizontal Divider with "or" */}
        <div className="w-full flex items-center mb-4">
          <div className="flex-1 h-px bg-border" />
          <span className="px-3 text-xs text-muted font-normal uppercase">or</span>
          <div className="flex-1 h-px bg-border" />
        </div>

        {/* Traditional Credentials Form */}
        <form onSubmit={handleSubmit} className="w-full flex flex-col gap-3.5">
          {mode === "login" ? (
            <>
              <div>
                <label className="text-xs text-muted font-medium mb-1 block">Phone, email, or username</label>
                <input
                  type="text"
                  required
                  value={identifier}
                  onChange={(e) => setIdentifier(e.target.value)}
                  placeholder="Username or email"
                  className="w-full bg-transparent border border-border focus:border-brand rounded-lg p-3 text-[15px] outline-none text-foreground"
                />
              </div>
              <div>
                <label className="text-xs text-muted font-medium mb-1 block">Password</label>
                <input
                  type="password"
                  required
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="Password"
                  className="w-full bg-transparent border border-border focus:border-brand rounded-lg p-3 text-[15px] outline-none text-foreground"
                />
              </div>
            </>
          ) : (
            <>
              <div>
                <label className="text-xs text-muted font-medium mb-1 block">Display Name</label>
                <input
                  type="text"
                  value={displayName}
                  onChange={(e) => setDisplayName(e.target.value)}
                  placeholder="Display Name"
                  className="w-full bg-transparent border border-border focus:border-brand rounded-lg p-3 text-[15px] outline-none text-foreground"
                />
              </div>
              <div>
                <label className="text-xs text-muted font-medium mb-1 block">Username</label>
                <input
                  type="text"
                  required
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  placeholder="Username"
                  className="w-full bg-transparent border border-border focus:border-brand rounded-lg p-3 text-[15px] outline-none text-foreground"
                />
              </div>
              <div>
                <label className="text-xs text-muted font-medium mb-1 block">Email</label>
                <input
                  type="email"
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="Email"
                  className="w-full bg-transparent border border-border focus:border-brand rounded-lg p-3 text-[15px] outline-none text-foreground"
                />
              </div>
              <div>
                <label className="text-xs text-muted font-medium mb-1 block">Password (min 8 chars)</label>
                <input
                  type="password"
                  required
                  minLength={8}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="Password"
                  className="w-full bg-transparent border border-border focus:border-brand rounded-lg p-3 text-[15px] outline-none text-foreground"
                />
              </div>
            </>
          )}

          <button
            type="submit"
            disabled={loading || googleLoading}
            className="w-full mt-2 bg-foreground text-background font-bold text-[16px] py-3 rounded-full hover:opacity-90 transition-opacity flex items-center justify-center disabled:opacity-50"
          >
            {loading ? <Loader2 className="w-5 h-5 animate-spin" /> : mode === "login" ? "Sign In" : "Sign Up"}
          </button>
        </form>

        <div className="mt-6 text-sm text-muted text-center">
          {mode === "login" ? (
            <p>
              Don&apos;t have an account?{" "}
              <button
                onClick={() => {
                  setMode("register");
                  setError(null);
                }}
                className="text-brand hover:underline font-medium"
              >
                Sign up
              </button>
            </p>
          ) : (
            <p>
              Have an account already?{" "}
              <button
                onClick={() => {
                  setMode("login");
                  setError(null);
                }}
                className="text-brand hover:underline font-medium"
              >
                Log in
              </button>
            </p>
          )}
        </div>
      </div>
    </div>
  );
}
