"use client";

import React, { useState } from "react";
import { X, Loader2 } from "lucide-react";
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

  if (!isOpen) return null;

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
      <div className="relative w-full max-w-[500px] bg-background border border-border rounded-2xl shadow-2xl p-6 sm:p-8 flex flex-col items-center">
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

        <h2 className="text-[26px] font-extrabold mb-6 text-center">
          {mode === "login" ? "Sign in to X" : "Create your account"}
        </h2>

        {error && (
          <div className="w-full mb-4 p-3 rounded-lg bg-red-500/10 border border-red-500/30 text-red-500 text-sm text-center">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="w-full flex flex-col gap-4">
          {mode === "login" ? (
            <>
              <div>
                <label className="text-xs text-muted font-medium mb-1 block">Username or Email</label>
                <input
                  type="text"
                  required
                  value={identifier}
                  onChange={(e) => setIdentifier(e.target.value)}
                  placeholder="Username or email"
                  className="w-full bg-transparent border border-border focus:border-brand rounded-lg p-3 text-[15px] outline-none"
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
                  className="w-full bg-transparent border border-border focus:border-brand rounded-lg p-3 text-[15px] outline-none"
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
                  className="w-full bg-transparent border border-border focus:border-brand rounded-lg p-3 text-[15px] outline-none"
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
                  className="w-full bg-transparent border border-border focus:border-brand rounded-lg p-3 text-[15px] outline-none"
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
                  className="w-full bg-transparent border border-border focus:border-brand rounded-lg p-3 text-[15px] outline-none"
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
                  className="w-full bg-transparent border border-border focus:border-brand rounded-lg p-3 text-[15px] outline-none"
                />
              </div>
            </>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full mt-2 bg-foreground text-background font-bold text-[16px] py-3 rounded-full hover:opacity-90 transition-opacity flex items-center justify-center"
          >
            {loading ? <Loader2 className="w-5 h-5 animate-spin" /> : mode === "login" ? "Sign In" : "Sign Up"}
          </button>
        </form>

        <div className="mt-6 text-sm text-muted text-center">
          {mode === "login" ? (
            <p>
              Don't have an account?{" "}
              <button
                onClick={() => {
                  setMode("register");
                  setError(null);
                }}
                className="text-brand hover:underline font-semibold"
              >
                Sign up
              </button>
            </p>
          ) : (
            <p>
              Already have an account?{" "}
              <button
                onClick={() => {
                  setMode("login");
                  setError(null);
                }}
                className="text-brand hover:underline font-semibold"
              >
                Sign in
              </button>
            </p>
          )}
        </div>
      </div>
    </div>
  );
}
