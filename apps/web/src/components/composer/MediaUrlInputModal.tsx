"use client";

import React, { useState } from "react";
import { X, Image as ImageIcon, Link as LinkIcon, AlertCircle } from "lucide-react";
import { cn } from "@/lib/utils";

interface MediaUrlInputModalProps {
  isOpen: boolean;
  onClose: () => void;
  onAddMedia: (media: { url: string; type: "image" | "video" | "gif" }) => void;
}

export function MediaUrlInputModal({
  isOpen,
  onClose,
  onAddMedia,
}: MediaUrlInputModalProps) {
  const [url, setUrl] = useState("");
  const [previewError, setPreviewError] = useState(false);
  const [selectedType, setSelectedType] = useState<"image" | "video" | "gif">("image");

  if (!isOpen) return null;

  const detectTypeFromUrl = (inputUrl: string): "image" | "video" | "gif" => {
    const clean = inputUrl.trim().toLowerCase();
    if (clean.endsWith(".mp4") || clean.endsWith(".mov") || clean.endsWith(".webm") || clean.endsWith(".m4v")) {
      return "video";
    }
    if (clean.endsWith(".gif")) {
      return "gif";
    }
    return "image";
  };

  const handleUrlChange = (val: string) => {
    setUrl(val);
    setPreviewError(false);
    if (val.trim()) {
      setSelectedType(detectTypeFromUrl(val));
    }
  };

  const handleConfirm = () => {
    const trimmed = url.trim();
    if (!trimmed) return;
    onAddMedia({
      url: trimmed,
      type: selectedType,
    });
    setUrl("");
    setPreviewError(false);
    onClose();
  };

  const hasUrl = url.trim().length > 0;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in duration-150">
      <div
        className="relative w-full max-w-[500px] bg-[#000000] border border-border text-foreground rounded-2xl shadow-2xl overflow-hidden flex flex-col"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-border">
          <div className="flex items-center gap-2">
            <LinkIcon className="w-5 h-5 text-brand" />
            <h3 className="font-bold text-[17px]">Add Media via URL</h3>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="p-1.5 hover:bg-white/10 rounded-full transition-colors text-muted hover:text-foreground"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Content */}
        <div className="p-4 space-y-4">
          <div className="text-[13px] text-muted">
            Direct file upload is reserved for Premium and VIP accounts. Please provide a direct URL to an image, GIF, or video.
          </div>

          <div>
            <label className="block text-[13px] font-semibold text-muted mb-1.5">
              Media URL
            </label>
            <input
              type="url"
              value={url}
              onChange={(e) => handleUrlChange(e.target.value)}
              placeholder="https://example.com/image.jpg or .gif or .mp4"
              className="w-full px-3.5 py-2.5 rounded-xl bg-card border border-border text-[14px] text-foreground placeholder:text-muted/60 focus:outline-none focus:border-brand transition-colors"
              autoFocus
            />
          </div>

          {/* Media Type Selector */}
          <div>
            <label className="block text-[13px] font-semibold text-muted mb-1.5">
              Media Type
            </label>
            <div className="flex gap-2">
              {(["image", "gif", "video"] as const).map((t) => (
                <button
                  key={t}
                  type="button"
                  onClick={() => setSelectedType(t)}
                  className={cn(
                    "flex-1 py-1.5 px-3 rounded-lg text-[13px] font-medium border capitalize transition-all",
                    selectedType === t
                      ? "bg-brand/15 border-brand text-brand font-semibold"
                      : "bg-card border-border text-muted hover:text-foreground"
                  )}
                >
                  {t}
                </button>
              ))}
            </div>
          </div>

          {/* Live Preview Box */}
          <div>
            <div className="flex items-center justify-between text-[13px] font-semibold text-muted mb-1.5">
              <span>Preview</span>
              {hasUrl && !previewError && (
                <span className="text-[11px] text-green-500 font-normal">Active</span>
              )}
            </div>

            <div className="w-full h-[200px] rounded-xl border border-border bg-black/40 overflow-hidden flex items-center justify-center relative">
              {hasUrl ? (
                previewError ? (
                  <div className="flex flex-col items-center justify-center text-center p-4 text-red-400 gap-1.5">
                    <AlertCircle className="w-8 h-8" />
                    <span className="text-[13px] font-medium">Failed to load preview</span>
                    <span className="text-[11px] text-muted">Check the URL or access permissions</span>
                  </div>
                ) : selectedType === "video" ? (
                  <video
                    src={url.trim()}
                    controls
                    onError={() => setPreviewError(true)}
                    className="w-full h-full object-contain"
                  />
                ) : (
                  <img
                    src={url.trim()}
                    alt="Preview"
                    onError={() => setPreviewError(true)}
                    className="w-full h-full object-contain"
                  />
                )
              ) : (
                <div className="flex flex-col items-center justify-center text-muted gap-2">
                  <ImageIcon className="w-8 h-8 opacity-40" />
                  <span className="text-[12px]">Enter a URL above to see live preview</span>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end gap-2 px-4 py-3 border-t border-border bg-card/40">
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 text-[14px] font-medium text-muted hover:text-foreground rounded-full transition-colors"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={handleConfirm}
            disabled={!hasUrl}
            className={cn(
              "px-5 py-2 text-[14px] font-bold rounded-full transition-all",
              hasUrl
                ? "bg-brand text-white hover:bg-brand/90 active:scale-95"
                : "bg-muted/20 text-muted cursor-not-allowed"
            )}
          >
            Add Media
          </button>
        </div>
      </div>
    </div>
  );
}
