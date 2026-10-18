"use client";

import React, { useEffect } from "react";
import { Loader2 } from "lucide-react";

interface ConfirmModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void | Promise<void>;
  title: string;
  description: string;
  confirmText?: string;
  cancelText?: string;
  variant?: "destructive" | "primary";
  isLoading?: boolean;
}

export function ConfirmModal({
  isOpen,
  onClose,
  onConfirm,
  title,
  description,
  confirmText = "Delete",
  cancelText = "Cancel",
  variant = "destructive",
  isLoading = false,
}: ConfirmModalProps) {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape" && isOpen && !isLoading) {
        onClose();
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isOpen, isLoading, onClose]);

  if (!isOpen) return null;

  return (
    <div
      onClick={(e) => {
        e.stopPropagation();
        if (!isLoading) onClose();
      }}
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 animate-in fade-in duration-150 select-none"
    >
      <div
        onClick={(e) => e.stopPropagation()}
        className="relative w-full max-w-[320px] bg-background border border-border rounded-2xl shadow-2xl p-6 sm:p-8 flex flex-col animate-in zoom-in-95 duration-150"
      >
        <h2 className="text-[20px] font-extrabold text-foreground mb-2 leading-tight">
          {title}
        </h2>
        <p className="text-[15px] leading-5 text-muted mb-6 font-normal">
          {description}
        </p>

        <button
          type="button"
          disabled={isLoading}
          onClick={async (e) => {
            e.stopPropagation();
            await onConfirm();
          }}
          className={
            variant === "destructive"
              ? "w-full bg-[#f4212e] hover:bg-[#dc1e29] text-white font-bold py-3 px-4 rounded-full text-[15px] transition-colors mb-3 flex items-center justify-center disabled:opacity-50 cursor-pointer"
              : "w-full bg-foreground hover:opacity-90 text-background font-bold py-3 px-4 rounded-full text-[15px] transition-colors mb-3 flex items-center justify-center disabled:opacity-50 cursor-pointer"
          }
        >
          {isLoading ? (
            <Loader2 className="w-4 h-4 animate-spin" />
          ) : (
            confirmText
          )}
        </button>

        <button
          type="button"
          disabled={isLoading}
          onClick={(e) => {
            e.stopPropagation();
            onClose();
          }}
          className="w-full bg-transparent hover:bg-subtle text-foreground font-bold py-3 px-4 rounded-full text-[15px] transition-colors border border-border flex items-center justify-center disabled:opacity-50 cursor-pointer"
        >
          {cancelText}
        </button>
      </div>
    </div>
  );
}
