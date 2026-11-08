"use client";

import React, { useState, useEffect } from "react";
import { Shield, Check, Copy, Loader2, KeyRound, ShieldCheck, ShieldOff, AlertTriangle } from "lucide-react";
import { apiRequest } from "@/lib/api";
import { TOTPSetupResponse, User } from "@/types";
import { QRCodeSVG } from "@/lib/qrcode";

interface TwoFactorManagerProps {
  currentUser: User;
  onUserUpdated?: (updated: User) => void;
  compact?: boolean;
}

export function TwoFactorManager({ currentUser, onUserUpdated, compact = false }: TwoFactorManagerProps) {
  const [totpEnabled, setTotpEnabled] = useState<boolean>(!!currentUser.totp_enabled);
  const [setupData, setSetupData] = useState<TOTPSetupResponse | null>(null);
  const [code, setCode] = useState("");
  const [disableInput, setDisableInput] = useState("");
  const [showDisableForm, setShowDisableForm] = useState(false);
  const [loading, setLoading] = useState(false);
  const [copied, setCopied] = useState(false);
  const [copiedSeed, setCopiedSeed] = useState(false);
  const [copiedBackupCodes, setCopiedBackupCodes] = useState(false);
  // One-time reveal: Only populated immediately upon activation, never shown again afterwards
  const [oneTimeCodes, setOneTimeCodes] = useState<{ seed_phrase?: string; backup_codes?: string[] } | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);

  useEffect(() => {
    setTotpEnabled(!!currentUser.totp_enabled);
  }, [currentUser.totp_enabled]);

  useEffect(() => {
    apiRequest<{ totp_enabled: boolean }>("/me/totp/status")
      .then((res) => {
        setTotpEnabled(!!res.totp_enabled);
      })
      .catch(() => {});
  }, []);

  const syncUserState = (enabled: boolean) => {
    setTotpEnabled(enabled);
    const updatedUser: User = { ...currentUser, totp_enabled: enabled };
    if (typeof window !== "undefined") {
      try {
        const stored = localStorage.getItem("x_session_user");
        if (stored) {
          const parsed = JSON.parse(stored);
          localStorage.setItem("x_session_user", JSON.stringify({ ...parsed, totp_enabled: enabled }));
        }
      } catch {}
    }
    onUserUpdated?.(updatedUser);
  };

  const handleStartSetup = async (regenerate = false) => {
    setLoading(true);
    setError(null);
    setSuccessMsg(null);
    setOneTimeCodes(null);
    try {
      const endpoint = regenerate ? "/me/totp/setup?regenerate=true" : "/me/totp/setup";
      const res = await apiRequest<TOTPSetupResponse>(endpoint, {
        method: "POST",
      });
      setSetupData(res);
      setCode("");
    } catch (err: any) {
      setError(err.message || "Failed to initialize authenticator setup");
    } finally {
      setLoading(false);
    }
  };

  const handleEnable = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    const cleaned = code.trim();
    if (!cleaned) {
      setError("Please enter the 6-digit code from your authenticator app");
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const res = await apiRequest<{ totp_enabled: boolean; seed_phrase?: string; backup_codes?: string[] }>("/me/totp/enable", {
        method: "POST",
        body: JSON.stringify({ code: cleaned }),
      });

      setSetupData(null);
      setCode("");
      syncUserState(true);
      // Present one-time codes immediately upon activation
      setOneTimeCodes({
        seed_phrase: res.seed_phrase,
        backup_codes: res.backup_codes && res.backup_codes.length > 0 ? res.backup_codes : setupData?.backup_codes,
      });
      setSuccessMsg("Two-factor authentication is now active on your account.");
    } catch (err: any) {
      setError(err.message || "Invalid verification code. Please check your app and try again.");
    } finally {
      setLoading(false);
    }
  };

  const handleDisable = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    const trimmed = disableInput.trim();
    if (!trimmed) {
      setError("Enter your 6-digit authenticator code or account password to disable 2FA");
      return;
    }

    setLoading(true);
    setError(null);
    try {
      const isSixDigits = /^\d{6}$/.test(trimmed);
      await apiRequest("/me/totp/disable", {
        method: "POST",
        body: JSON.stringify(isSixDigits ? { code: trimmed } : { password: trimmed }),
      });
      setShowDisableForm(false);
      setDisableInput("");
      setOneTimeCodes(null);
      syncUserState(false);
      setSuccessMsg("Two-factor authentication has been turned off.");
    } catch (err: any) {
      setError(err.message || "Failed to disable two-factor authentication");
    } finally {
      setLoading(false);
    }
  };

  const handleCopySecret = () => {
    if (!setupData?.secret) return;
    navigator.clipboard.writeText(setupData.secret);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleCopySeed = (seedText?: string) => {
    const text = seedText || oneTimeCodes?.seed_phrase;
    if (!text) return;
    navigator.clipboard.writeText(text);
    setCopiedSeed(true);
    setTimeout(() => setCopiedSeed(false), 2000);
  };

  const handleCopyBackupCodes = (codes: string[]) => {
    if (!codes || codes.length === 0) return;
    navigator.clipboard.writeText(codes.join("\n"));
    setCopiedBackupCodes(true);
    setTimeout(() => setCopiedBackupCodes(false), 2000);
  };

  return (
    <div className={compact ? "pt-3 border-t border-neutral-900" : "border border-border rounded-2xl p-4 sm:p-5 bg-subtle/10"}>
      <div className="flex items-start justify-between gap-4">
        <div className="flex items-start gap-3">
          <div className="p-2 rounded-full bg-brand/10 text-brand shrink-0 mt-0.5">
            {totpEnabled ? <ShieldCheck className="w-5 h-5 text-green-500" /> : <Shield className="w-5 h-5" />}
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h3 className="text-[15px] font-bold text-foreground">
                Two-factor authentication (TOTP)
              </h3>
              {totpEnabled ? (
                <span className="text-[11px] font-bold px-2 py-0.5 rounded-full bg-green-500/15 text-green-500 border border-green-500/30">
                  Enabled
                </span>
              ) : (
                <span className="text-[11px] font-medium px-2 py-0.5 rounded-full bg-neutral-800 text-muted">
                  Off
                </span>
              )}
            </div>
            <p className="text-[13px] text-muted mt-0.5 leading-snug">
              Protect your account by requiring a 6-digit verification code from an authenticator app (Google Authenticator, Authy, 1Password) when signing in.
            </p>
          </div>
        </div>

        {!setupData && !totpEnabled && (
          <button
            type="button"
            onClick={() => handleStartSetup(false)}
            disabled={loading}
            className="shrink-0 bg-foreground text-background hover:opacity-90 font-bold text-[13px] px-4 py-1.5 rounded-full transition-opacity disabled:opacity-50 flex items-center gap-1.5"
          >
            {loading && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
            <span>Set up TOTP</span>
          </button>
        )}

        {totpEnabled && !showDisableForm && (
          <button
            type="button"
            onClick={() => {
              setShowDisableForm(true);
              setError(null);
              setSuccessMsg(null);
            }}
            className="shrink-0 border border-red-500/40 text-red-500 hover:bg-red-500/10 font-bold text-[13px] px-3.5 py-1.5 rounded-full transition-colors flex items-center gap-1.5"
          >
            <ShieldOff className="w-3.5 h-3.5" />
            <span>Turn off</span>
          </button>
        )}
      </div>

      {error && (
        <div className="mt-3 p-2.5 rounded-xl bg-red-500/10 border border-red-500/30 text-red-500 text-xs font-medium">
          {error}
        </div>
      )}

      {successMsg && !oneTimeCodes && (
        <div className="mt-3 p-2.5 rounded-xl bg-green-500/10 border border-green-500/30 text-green-500 text-xs font-medium flex items-center gap-2">
          <Check className="w-4 h-4 shrink-0" />
          <span>{successMsg}</span>
        </div>
      )}

      {/* QR Code Setup Panel (Only visible during initial setup, no codes shown yet) */}
      {setupData && !totpEnabled && (
        <div className="mt-4 p-4 rounded-2xl border border-neutral-800 bg-neutral-950/80 space-y-4">
          <div className="flex flex-col sm:flex-row items-center gap-4">
            <div className="p-2.5 bg-white rounded-xl shrink-0 shadow-md">
              <QRCodeSVG value={setupData.otpauth_url} size={168} />
            </div>

            <div className="flex-1 space-y-2.5 text-left w-full">
              <div className="text-[14px] font-bold text-foreground">
                1. Scan QR or use your Authenticator app
              </div>
              <p className="text-xs text-muted leading-relaxed">
                Open Google Authenticator, Authy, or your preferred TOTP app. If you cannot scan the QR code, enter this secret key manually:
              </p>

              <div className="flex items-center justify-between gap-2 bg-neutral-900 border border-neutral-800 rounded-lg px-3 py-2">
                <code className="text-xs font-mono text-brand tracking-wider break-all select-all">
                  {setupData.secret}
                </code>
                <div className="flex items-center gap-1 shrink-0">
                  <button
                    type="button"
                    onClick={handleCopySecret}
                    className="p-1.5 rounded-md hover:bg-neutral-800 text-muted hover:text-foreground transition-colors"
                    title="Copy setup key"
                  >
                    {copied ? <Check className="w-4 h-4 text-green-500" /> : <Copy className="w-4 h-4" />}
                  </button>
                </div>
              </div>
              <div className="flex items-center justify-between text-[11px] text-muted">
                <span>Scan QR or copy setup key into Authenticator</span>
                <button
                  type="button"
                  onClick={() => handleStartSetup(true)}
                  disabled={loading}
                  className="text-brand hover:underline font-semibold"
                >
                  Generate new QR key
                </button>
              </div>
            </div>
          </div>

          <div className="pt-2 border-t border-neutral-800">
            <label className="block text-xs font-bold text-foreground mb-1.5">
              2. Enter the 6-digit verification code from your Authenticator app
            </label>
            <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-2.5">
              <input
                type="text"
                autoComplete="one-time-code"
                value={code}
                onChange={(e) => setCode(e.target.value)}
                placeholder="000000"
                maxLength={8}
                className="flex-1 bg-transparent border border-neutral-700 focus:border-brand rounded-lg px-3.5 py-2 text-[15px] font-mono text-center outline-none text-foreground"
                onKeyDown={(e) => {
                  if (e.key === "Enter") {
                    e.preventDefault();
                    handleEnable();
                  }
                }}
              />
              <div className="flex items-center gap-2">
                <button
                  type="button"
                  onClick={() => handleEnable()}
                  disabled={loading || !code.trim()}
                  className="bg-brand hover:bg-brand/90 text-white font-bold text-[14px] px-5 py-2 rounded-full transition-colors disabled:opacity-50 flex items-center justify-center gap-1.5"
                >
                  {loading && <Loader2 className="w-4 h-4 animate-spin" />}
                  <span>Verify &amp; Activate</span>
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setSetupData(null);
                    setCode("");
                    setError(null);
                  }}
                  disabled={loading}
                  className="border border-neutral-700 hover:bg-neutral-800 text-muted hover:text-foreground font-bold text-[13px] px-4 py-2 rounded-full transition-colors"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ONE-TIME ONLY DISPLAY: Show backup codes & seed phrase immediately after activation, prompt user to save */}
      {totpEnabled && oneTimeCodes && (
        <div className="mt-4 p-4 rounded-2xl border-2 border-emerald-500/40 bg-neutral-950 space-y-4 animate-in fade-in">
          {/* Important Notice */}
          <div className="p-3.5 rounded-xl bg-amber-500/10 border border-amber-500/30 flex items-start gap-3">
            <AlertTriangle className="w-5 h-5 text-amber-400 shrink-0 mt-0.5" />
            <div>
              <h4 className="text-xs font-bold text-amber-300">
                Save your backup codes and recovery seed phrase now
              </h4>
              <p className="text-[11px] text-zinc-300 mt-1 leading-relaxed">
                For security reasons, this is the <strong>only time</strong> your backup codes and seed phrase will ever be shown. Make sure to copy them to a password manager or safe storage. Once you dismiss this screen, they will never be displayed again.
              </p>
            </div>
          </div>

          {/* 10 Backup Codes Card */}
          {oneTimeCodes.backup_codes && oneTimeCodes.backup_codes.length > 0 && (
            <div className="p-3.5 rounded-xl bg-emerald-950/20 border border-emerald-500/20 space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-xs font-bold text-emerald-400">
                  10 Backup Codes (e.g. 58B5-A215)
                </span>
                <button
                  type="button"
                  onClick={() => handleCopyBackupCodes(oneTimeCodes.backup_codes!)}
                  className="px-2.5 py-1 rounded-full bg-emerald-500/20 text-emerald-300 hover:bg-emerald-500/30 text-[11px] font-semibold flex items-center gap-1 transition-colors"
                >
                  {copiedBackupCodes ? <Check className="w-3.5 h-3.5" /> : <Copy className="w-3.5 h-3.5" />}
                  <span>{copiedBackupCodes ? "Copied" : "Copy All Codes"}</span>
                </button>
              </div>
              <div className="grid grid-cols-2 gap-1.5 p-2.5 rounded-xl bg-black/60 border border-emerald-500/20 font-mono text-xs text-emerald-300 text-center select-all">
                {oneTimeCodes.backup_codes.map((c, i) => (
                  <span key={i} className="py-1 px-1.5 bg-white/5 rounded border border-emerald-500/10">{c}</span>
                ))}
              </div>
            </div>
          )}

          {/* Account Recovery Seed Phrase Card */}
          {oneTimeCodes.seed_phrase && (
            <div className="p-3.5 rounded-xl bg-amber-500/10 border border-amber-500/30 space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-xs font-bold text-amber-400">
                  Account Recovery Seed Phrase
                </span>
                <button
                  type="button"
                  onClick={() => handleCopySeed(oneTimeCodes.seed_phrase)}
                  className="px-2.5 py-1 rounded-full bg-amber-500/20 text-amber-300 hover:bg-amber-500/30 text-[11px] font-semibold flex items-center gap-1 transition-colors"
                >
                  {copiedSeed ? <Check className="w-3.5 h-3.5" /> : <Copy className="w-3.5 h-3.5" />}
                  <span>{copiedSeed ? "Copied" : "Copy Seed Phrase"}</span>
                </button>
              </div>
              <div className="p-2.5 rounded-xl bg-black/60 border border-amber-500/20 text-amber-200 font-mono text-xs select-all break-words leading-relaxed">
                {oneTimeCodes.seed_phrase}
              </div>
            </div>
          )}

          {/* Dismiss / Acknowledge Button */}
          <button
            type="button"
            onClick={() => {
              setOneTimeCodes(null);
              setSuccessMsg("Two-factor authentication is active. Backup codes saved.");
            }}
            className="w-full py-2.5 px-4 rounded-full font-bold bg-white hover:bg-neutral-200 text-black text-sm transition-colors flex items-center justify-center gap-2 shadow-md cursor-pointer"
          >
            <Check className="w-4 h-4 text-black" />
            <span>I have saved my backup codes &amp; seed phrase</span>
          </button>
        </div>
      )}

      {/* Disable TOTP Confirmation Form */}
      {totpEnabled && showDisableForm && (
        <div className="mt-4 p-4 rounded-2xl border border-neutral-800 bg-neutral-950/80 space-y-3">
          <div className="flex items-center gap-2 text-sm font-bold text-foreground">
            <KeyRound className="w-4 h-4 text-red-500" />
            <span>Confirm turning off two-factor authentication</span>
          </div>
          <p className="text-xs text-muted">
            Enter your current 6-digit authenticator code or your account password to disable TOTP protection.
          </p>
          <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-2.5">
            <input
              type="text"
              value={disableInput}
              onChange={(e) => setDisableInput(e.target.value)}
              placeholder="6-digit TOTP code or password"
              className="flex-1 bg-transparent border border-neutral-700 focus:border-red-500 rounded-lg px-3.5 py-2 text-[14px] outline-none text-foreground"
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  handleDisable();
                }
              }}
            />
            <div className="flex items-center gap-2">
              <button
                type="button"
                onClick={() => handleDisable()}
                disabled={loading || !disableInput.trim()}
                className="bg-red-600 hover:bg-red-700 text-white font-bold text-[13px] px-4 py-2 rounded-full transition-colors disabled:opacity-50 flex items-center gap-1.5"
              >
                {loading && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
                <span>Disable 2FA</span>
              </button>
              <button
                type="button"
                onClick={() => {
                  setShowDisableForm(false);
                  setDisableInput("");
                }}
                disabled={loading}
                className="border border-neutral-700 hover:bg-neutral-800 text-muted hover:text-foreground font-bold text-[13px] px-4 py-2 rounded-full transition-colors"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
