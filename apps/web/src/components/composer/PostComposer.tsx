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
} from "lucide-react";
import { User, MediaObject } from "@/types";
import { apiRequest } from "@/lib/api";
import { cn } from "@/lib/utils";

interface PostComposerProps {
  currentUser: User | null;
  placeholder?: string;
  replyToId?: string;
  onPostSuccess?: (newPost: any) => void;
  onOpenAuth?: () => void;
}

interface StagedAttachment {
  id: string;
  file: File;
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
  const [content, setContent] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [stagedFiles, setStagedFiles] = useState<StagedAttachment[]>([]);
  const fileInputRef = useRef<HTMLInputElement>(null);

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
    const files = Array.from(e.target.files || []);
    if (!files.length) return;

    if (!currentUser) {
      if (onOpenAuth) onOpenAuth();
      return;
    }

    const newAttachments: StagedAttachment[] = files.map((file) => {
      const type = file.type.startsWith("video")
        ? "video"
        : file.type.includes("gif")
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
      if (target?.previewUrl) URL.revokeObjectURL(target.previewUrl);
      return prev.filter((_, i) => i !== index);
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

      // Concurrently compress and upload staged assets
      if (stagedFiles.length > 0) {
        const uploadPromises = stagedFiles.map(async (staged) => {
          const formData = new FormData();
          formData.append("file", staged.file);
          const media = await apiRequest<MediaObject>("/media/upload", {
            method: "POST",
            body: formData,
          });
          return media.id;
        });

        const ids = await Promise.all(uploadPromises);
        uploadedMediaIds.push(...ids);
      }

      const payload = {
        content: content.trim(),
        media_ids: uploadedMediaIds,
        reply_to_id: replyToId || null,
      };

      const res = await apiRequest("/posts", {
        method: "POST",
        body: JSON.stringify(payload),
      });

      // Cleanup preview URLs
      stagedFiles.forEach((item) => URL.revokeObjectURL(item.previewUrl));
      setContent("");
      setStagedFiles([]);
      if (onPostSuccess) onPostSuccess(res);
    } catch (err: any) {
      alert(err.message || "Failed to post");
    } finally {
      setIsSubmitting(false);
    }
  };

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
                <button
                  type="button"
                  onClick={() => handleRemoveStaged(idx)}
                  className="absolute top-2 right-2 bg-black/75 hover:bg-black text-white p-1.5 rounded-full transition-colors z-10 shadow-md"
                  aria-label="Remove media"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>
            ))}
          </div>
        )}

        {/* Toolbar & Submit */}
        <div className="flex items-center justify-between pt-2">
          <div className="flex items-center gap-1 text-brand">
            <input
              type="file"
              ref={fileInputRef}
              onChange={handleFileChange}
              accept="image/*,video/*"
              multiple
              className="hidden"
            />
            <button
              type="button"
              onClick={() => fileInputRef.current?.click()}
              className="p-2 hover:bg-brand/10 rounded-full transition-colors"
              title="Add photos or video"
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
              <a
                href="/premium"
                className="text-[12px] text-brand hover:underline hidden sm:inline"
              >
                Need 4,000 chars? Premium
              </a>
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
                "bg-brand text-white font-bold text-[15px] px-4 py-1.5 rounded-full shadow-sm transition-all flex items-center justify-center min-w-[70px]",
                !canPost
                  ? "opacity-50 cursor-not-allowed"
                  : "hover:bg-brand-hover active:scale-95"
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
    </div>
  );
}
