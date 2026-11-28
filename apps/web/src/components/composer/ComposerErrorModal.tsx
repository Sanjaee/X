"use client";

import React from "react";
import { AlertCircle, X } from "lucide-react";

interface ComposerErrorModalProps {
  message: string | null;
  onClose: () => void;
}

export function ComposerErrorModal({ message, onClose }: ComposerErrorModalProps) {
  if (!message) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4"
      onClick={onClose}
    >
      <div
        className="w-full max-w-sm rounded-2xl bg-background border border-border shadow-2xl p-5 space-y-4 animate-in fade-in zoom-in-95 duration-150"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-start justify-between gap-3">
          <div className="flex items-center gap-2.5 text-destructive">
            <AlertCircle className="w-5 h-5 shrink-0" />
            <h3 className="font-bold text-base text-foreground">Unable to Post</h3>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="p-1 rounded-full hover:bg-subtle text-muted-foreground hover:text-foreground transition-colors"
            aria-label="Close"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        <p className="text-sm text-muted-foreground leading-relaxed">{message}</p>

        <div className="flex justify-end pt-1">
          <button
            type="button"
            onClick={onClose}
            className="px-5 py-2 rounded-full bg-brand hover:bg-brand/90 text-white text-sm font-bold transition-all active:scale-95"
          >
            OK
          </button>
        </div>
      </div>
    </div>
  );
}
