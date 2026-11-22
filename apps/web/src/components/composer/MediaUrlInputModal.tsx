"use client";

import React, { useState } from "react";
import { X, Image as ImageIcon, Link as LinkIcon, AlertCircle, HelpCircle, ExternalLink, Sparkles, CheckCircle2 } from "lucide-react";
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
  const [showImageKitGuide, setShowImageKitGuide] = useState(false);

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
    <div className="fixed inset-0 z-[200] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in duration-150">
      <div
        className="relative w-full max-w-[500px] max-h-[90vh] overflow-y-auto bg-[#000000] border border-border text-foreground rounded-2xl shadow-2xl flex flex-col"
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

          {/* ImageKit Suggestion Box */}
          <div className="flex items-center justify-between p-3 rounded-xl bg-brand/10 border border-brand/20">
            <div className="flex items-center gap-2.5">
              <div className="w-8 h-8 rounded-lg bg-brand/20 text-brand flex items-center justify-center shrink-0">
                <Sparkles className="w-4 h-4" />
              </div>
              <div className="text-left">
                <p className="text-[13px] font-semibold text-foreground">Need a free image host?</p>
                <p className="text-[11px] text-muted">Upload for free to ImageKit.io & paste URL</p>
              </div>
            </div>
            <button
              type="button"
              onClick={() => setShowImageKitGuide(true)}
              className="px-3 py-1.5 text-[12px] font-semibold bg-brand text-white hover:bg-brand-hover active:scale-95 rounded-lg transition-all flex items-center gap-1.5 shadow-sm shrink-0"
            >
              <HelpCircle className="w-3.5 h-3.5" />
              <span>Guide</span>
            </button>
          </div>

          <div>
            <div className="flex items-center justify-between mb-1.5">
              <label className="text-[13px] font-semibold text-muted">
                Media URL
              </label>
              <button
                type="button"
                onClick={() => setShowImageKitGuide(true)}
                className="text-[11px] text-brand hover:underline flex items-center gap-1"
              >
                <span>How to use ImageKit.io</span>
                <ExternalLink className="w-3 h-3" />
              </button>
            </div>
            <input
              type="url"
              value={url}
              onChange={(e) => handleUrlChange(e.target.value)}
              placeholder="https://ik.imagekit.io/... or https://example.com/image.jpg"
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

      {/* ImageKit Guide Popup Modal */}
      {showImageKitGuide && (
        <div
          className="fixed inset-0 z-[210] flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm animate-in fade-in duration-150"
          onClick={() => setShowImageKitGuide(false)}
        >
          <div
            className="relative w-full max-w-[480px] bg-[#000000] border border-border text-foreground rounded-2xl shadow-2xl p-5 flex flex-col gap-4 animate-in zoom-in-95 duration-150"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-center justify-between border-b border-border pb-3">
              <div className="flex items-center gap-2">
                <div className="w-7 h-7 rounded-lg bg-brand/20 text-brand flex items-center justify-center">
                  <ExternalLink className="w-4 h-4" />
                </div>
                <h4 className="font-bold text-[16px]">How to upload via ImageKit.io</h4>
              </div>
              <button
                type="button"
                onClick={() => setShowImageKitGuide(false)}
                className="p-1.5 hover:bg-white/10 rounded-full transition-colors text-muted hover:text-foreground"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <div className="space-y-3.5 text-[13px] text-foreground/90">
              <p className="text-muted leading-relaxed">
                ImageKit provides fast, free CDN-backed image & media hosting. Follow these simple steps:
              </p>

              <div className="space-y-3">
                <div className="flex items-start gap-3 p-2.5 rounded-xl bg-card border border-border">
                  <div className="w-6 h-6 rounded-full bg-brand/20 text-brand flex items-center justify-center font-bold text-xs shrink-0 mt-0.5">
                    1
                  </div>
                  <div>
                    <span className="font-semibold text-foreground">Open ImageKit.io</span>
                    <p className="text-[12px] text-muted mt-0.5">
                      Visit{" "}
                      <a
                        href="https://imagekit.io/"
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-brand hover:underline font-medium inline-flex items-center gap-0.5"
                      >
                        imagekit.io
                        <ExternalLink className="w-3 h-3" />
                      </a>{" "}
                      and log in or create a free account.
                    </p>
                  </div>
                </div>

                <div className="flex items-start gap-3 p-2.5 rounded-xl bg-card border border-border">
                  <div className="w-6 h-6 rounded-full bg-brand/20 text-brand flex items-center justify-center font-bold text-xs shrink-0 mt-0.5">
                    2
                  </div>
                  <div>
                    <span className="font-semibold text-foreground">Upload your Image or Video</span>
                    <p className="text-[12px] text-muted mt-0.5">
                      Go to the <strong>Media Library</strong> and drag & drop your media files to upload.
                    </p>
                  </div>
                </div>

                <div className="flex items-start gap-3 p-2.5 rounded-xl bg-card border border-border">
                  <div className="w-6 h-6 rounded-full bg-brand/20 text-brand flex items-center justify-center font-bold text-xs shrink-0 mt-0.5">
                    3
                  </div>
                  <div>
                    <span className="font-semibold text-foreground">Copy Direct File URL</span>
                    <p className="text-[12px] text-muted mt-0.5">
                      Click on the uploaded file and copy its direct public URL (e.g., <code className="bg-subtle text-foreground px-1 py-0.5 rounded text-[11px]">https://ik.imagekit.io/your_id/photo.jpg</code>).
                    </p>
                  </div>
                </div>

                <div className="flex items-start gap-3 p-2.5 rounded-xl bg-card border border-border">
                  <div className="w-6 h-6 rounded-full bg-brand/20 text-brand flex items-center justify-center font-bold text-xs shrink-0 mt-0.5">
                    4
                  </div>
                  <div>
                    <span className="font-semibold text-foreground">Paste into Media URL</span>
                    <p className="text-[12px] text-muted mt-0.5">
                      Paste the link into the URL field here, verify the live preview, and click <strong>Add Media</strong>!
                    </p>
                  </div>
                </div>
              </div>
            </div>

            <div className="flex items-center justify-between pt-3 border-t border-border mt-1">
              <a
                href="https://imagekit.io/"
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1.5 text-xs text-brand hover:underline font-semibold"
              >
                <span>Visit ImageKit.io</span>
                <ExternalLink className="w-3.5 h-3.5" />
              </a>
              <button
                type="button"
                onClick={() => setShowImageKitGuide(false)}
                className="px-4 py-1.5 text-[13px] font-bold bg-brand text-white hover:bg-brand-hover active:scale-95 rounded-full transition-all shadow"
              >
                Got it
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
