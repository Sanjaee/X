"use client";

import React, { useState, useRef, useEffect } from "react";
import {
  Image as ImageIcon,
  Smile,
  Calendar,
  MapPin,
  X,
  Loader2,
  Play,
  ArrowLeft,
  ArrowRight,
} from "lucide-react";
import { User, MediaObject } from "@/types";
import { apiRequest } from "@/lib/api";
import { cn } from "@/lib/utils";
import {
  X_ALLOWED_MEDIA_ACCEPT,
  filterValidMediaFiles,
} from "@/lib/mediaValidation";
import { useUpgradeModal } from "@/components/subscription/UpgradeModalContext";
import { checkIsPremiumUser } from "@/lib/premium";
import { MediaUrlInputModal } from "./MediaUrlInputModal";

interface PostComposerProps {
  currentUser: User | null;
  placeholder?: string;
  replyToId?: string;
  onPostSuccess?: (newPost: any) => void;
  onOpenAuth?: () => void;
}

interface StagedAttachment {
  id: string;
  file?: File;
  previewUrl: string;
  type: "image" | "video" | "gif";
}

export function PostComposer({
  currentUser,
  placeholder = "What is happening?!",
  replyToId,
  onPostSuccess,
  onOpenAuth,
}: PostComposerProps) {
  const { openUpgradeModal } = useUpgradeModal();
  const [content, setContent] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [stagedFiles, setStagedFiles] = useState<StagedAttachment[]>([]);
  const [urlModalOpen, setUrlModalOpen] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const isPremium = checkIsPremiumUser(currentUser);

  const maxChars = currentUser?.is_verified ? 4000 : 280;
  const charsRemaining = maxChars - content.length;
  const canPost = (content.trim().length > 0 || stagedFiles.length > 0) && charsRemaining >= 0 && !isSubmitting;

  // Cleanup object URLs on unmount
  useEffect(() => {
    return () => {
      stagedFiles.forEach((item) => URL.revokeObjectURL(item.previewUrl));
    };
  }, [stagedFiles]);

  // Stage files locally in state without uploading immediately (Observable X flow)
  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const rawFiles = Array.from(e.target.files || []);
    if (!rawFiles.length) return;

    if (!currentUser) {
      if (onOpenAuth) onOpenAuth();
      return;
    }

    const { valid, rejected } = filterValidMediaFiles(rawFiles);
    if (rejected.length > 0) {
      alert("Unsupported file format. Please select valid images or videos (*.pjp, *.jfif, *.jpe, *.pjpeg, *.jpeg, *.jpg, *.png, *.webp, *.gif, *.m4v, *.mp4, *.mov).");
    }

    if (!valid.length) {
      if (fileInputRef.current) fileInputRef.current.value = "";
      return;
    }

    const newAttachments: StagedAttachment[] = valid.map((file) => {
      const type = file.type.startsWith("video") || file.name.endsWith(".mp4") || file.name.endsWith(".mov") || file.name.endsWith(".m4v")
        ? "video"
        : file.type.includes("gif") || file.name.endsWith(".gif")
        ? "gif"
        : "image";

      return {
        id: Math.random().toString(36).substring(2, 9),
        file,
        previewUrl: URL.createObjectURL(file),
        type,
      };
    });

    // Limit to max 4 media attachments per post (X standard)
    setStagedFiles((prev) => [...prev, ...newAttachments].slice(0, 4));
    if (fileInputRef.current) fileInputRef.current.value = "";
  };

  const handleRemoveStaged = (index: number) => {
    setStagedFiles((prev) => {
      const target = prev[index];
      if (target?.file && target?.previewUrl) URL.revokeObjectURL(target.previewUrl);
      return prev.filter((_, i) => i !== index);
    });
  };

  const handleAddMediaUrl = (media: { url: string; type: "image" | "video" | "gif" }) => {
    setStagedFiles((prev) => [
      ...prev,
      {
        id: Math.random().toString(36).substring(2, 9),
        previewUrl: media.url,
        type: media.type,
      },
    ].slice(0, 4));
  };

  const handleMediaIconClick = () => {
    if (!currentUser) {
      if (onOpenAuth) onOpenAuth();
      return;
    }
    if (isPremium) {
      fileInputRef.current?.click();
    } else {
      setUrlModalOpen(true);
    }
  };

  const handleMoveMedia = (index: number, direction: "left" | "right") => {
    setStagedFiles((prev) => {
      const targetIdx = direction === "left" ? index - 1 : index + 1;
      if (targetIdx < 0 || targetIdx >= prev.length) return prev;
      const copy = [...prev];
      const temp = copy[index];
      copy[index] = copy[targetIdx];
      copy[targetIdx] = temp;
      return copy;
    });
  };

  // Upload queued attachments to Cloudflare R2 and publish post upon click
  const handleSubmit = async () => {
    if (!canPost) return;

    if (!currentUser) {
      if (onOpenAuth) onOpenAuth();
      return;
    }

    setIsSubmitting(true);
    try {
      const uploadedMediaIds: string[] = [];
      const inlineMedia: Array<{ url: string; media_type: string }> = [];

      if (stagedFiles.length > 0) {
        for (const staged of stagedFiles) {
          if (staged.file) {
            const formData = new FormData();
            formData.append("file", staged.file);
            const media = await apiRequest<MediaObject>("/media/upload", {
              method: "POST",
              body: formData,
            });
            uploadedMediaIds.push(media.id);
          } else if (staged.previewUrl) {
            inlineMedia.push({
              url: staged.previewUrl,
              media_type: staged.type || "image",
            });
          }
        }
      }

      const payload = {
        content: content.trim(),
        media_ids: uploadedMediaIds,
        media: inlineMedia,
        reply_to_id: replyToId || null,
      };

      const res = await apiRequest("/posts", {
        method: "POST",
        body: JSON.stringify(payload),
      });

      // Cleanup preview URLs
      stagedFiles.forEach((item) => {
        if (item.file && item.previewUrl) URL.revokeObjectURL(item.previewUrl);
      });
      setContent("");
      setStagedFiles([]);
      if (onPostSuccess) onPostSuccess(res);
    } catch (err: any) {
      alert(err.message || "Failed to post");
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!currentUser) {
    return null;
  }

  return (
    <div className="flex gap-3 px-4 py-3 border-b border-border bg-background select-none">
      {/* Avatar */}
      <img
        src={
          currentUser?.avatar_url ||
          "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"
        }
        alt="Avatar"
        className="w-10 h-10 rounded-full object-cover shrink-0 border border-border"
      />

      {/* Composer Area */}
      <div className="flex-1 min-w-0">
        <textarea
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder={placeholder}
          rows={2}
          className="w-full bg-transparent text-[18px] text-foreground placeholder:text-muted outline-none resize-none min-h-[56px]"
        />

        {/* Staged Media Previews (X Style Grid) */}
        {stagedFiles.length > 0 && (
          <div
            className={cn(
              "relative mb-3 rounded-2xl overflow-hidden border border-border gap-2 grid",
              stagedFiles.length === 1 ? "grid-cols-1" : "grid-cols-2"
            )}
          >
            {stagedFiles.map((staged, idx) => (
              <div
                key={staged.id}
                className={cn(
                  "relative group overflow-hidden bg-black/40",
                  stagedFiles.length === 1 ? "max-h-[360px]" : "h-[180px]"
                )}
              >
                {staged.type === "video" ? (
                  <video
                    src={staged.previewUrl}
                    controls
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <img
                    src={staged.previewUrl}
                    alt="Draft Attachment"
                    className="w-full h-full object-cover"
                  />
                )}

                {/* Control buttons: reorder left/right & delete */}
                <div className="absolute top-2 right-2 flex items-center gap-1.5 z-10">
                  {stagedFiles.length > 1 && (
                    <div className="flex items-center bg-black/75 rounded-full p-0.5 shadow-md">
                      <button
                        type="button"
                        onClick={() => handleMoveMedia(idx, "left")}
                        disabled={idx === 0}
                        className="p-1 hover:bg-white/20 rounded-full disabled:opacity-30 disabled:hover:bg-transparent text-white transition-colors"
                        title="Move left"
                      >
                        <ArrowLeft className="w-3.5 h-3.5" />
                      </button>
                      <button
                        type="button"
                        onClick={() => handleMoveMedia(idx, "right")}
                        disabled={idx === stagedFiles.length - 1}
                        className="p-1 hover:bg-white/20 rounded-full disabled:opacity-30 disabled:hover:bg-transparent text-white transition-colors"
                        title="Move right"
                      >
                        <ArrowRight className="w-3.5 h-3.5" />
                      </button>
                    </div>
                  )}
                  <button
                    type="button"
                    onClick={() => handleRemoveStaged(idx)}
                    className="bg-black/75 hover:bg-black text-white p-1.5 rounded-full transition-colors shadow-md"
                    aria-label="Remove media"
                    title="Remove media"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Toolbar & Submit */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-1 text-brand">
            <input
              type="file"
              ref={fileInputRef}
              onChange={handleFileChange}
              accept={X_ALLOWED_MEDIA_ACCEPT}
              multiple
              className="hidden"
            />
            <button
              type="button"
              onClick={handleMediaIconClick}
              className="p-2 hover:bg-brand/10 rounded-full transition-colors"
              title={isPremium ? "Add photos or video (Direct Upload)" : "Add media via URL"}
            >
              <ImageIcon className="w-5 h-5" />
            </button>
            <button
              type="button"
              className="p-2 hover:bg-brand/10 rounded-full transition-colors"
              title="Emoji"
            >
              <Smile className="w-5 h-5" />
            </button>
            <button
              type="button"
              className="p-2 hover:bg-brand/10 rounded-full transition-colors"
              title="Schedule"
            >
              <Calendar className="w-5 h-5" />
            </button>
            <button
              type="button"
              className="p-2 hover:bg-brand/10 rounded-full transition-colors"
              title="Location"
            >
              <MapPin className="w-5 h-5 opacity-50" />
            </button>
          </div>

          <div className="flex items-center gap-3">
            {!currentUser?.is_verified && content.length > 200 && (
              <button
                type="button"
                onClick={() => openUpgradeModal()}
                className="text-[12px] text-brand hover:underline hidden sm:inline"
              >
                Need 4,000 chars? Premium
              </button>
            )}

            {content.length > 0 && (
              <span
                className={cn(
                  "text-[13px] font-medium",
                  charsRemaining < 0
                    ? "text-red-500 font-bold"
                    : charsRemaining <= 20
                    ? "text-yellow-500"
                    : "text-muted"
                )}
              >
                {charsRemaining}
              </span>
            )}

            <button
              type="button"
              onClick={handleSubmit}
              disabled={!canPost}
              className={cn(
                "font-bold text-[15px] px-4 py-1.5 rounded-full transition-all flex items-center justify-center min-w-[70px]",
                !canPost
                  ? "bg-[#333639] text-[#71767b] cursor-not-allowed"
                  : "bg-[#eff3f4] text-[#0f1419] hover:bg-[#d7dbdc] active:scale-95 shadow-sm"
              )}
            >
              {isSubmitting ? (
                <Loader2 className="w-4 h-4 animate-spin" />
              ) : (
                "Post"
              )}
            </button>
          </div>
        </div>
      </div>

      {/* URL Media Input Modal for non-premium accounts */}
      <MediaUrlInputModal
        isOpen={urlModalOpen}
        onClose={() => setUrlModalOpen(false)}
        onAddMedia={handleAddMediaUrl}
      />
    </div>
  );
}
