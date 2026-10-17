"use client";

import React, { useState, useRef, useEffect } from "react";
import {
  X,
  Image as ImageIcon,
  Smile,
  Calendar,
  MapPin,
  Loader2,
  Globe,
  BarChart2,
} from "lucide-react";
import { User, Post, MediaObject } from "@/types";
import { apiRequest } from "@/lib/api";
import {
  X_ALLOWED_MEDIA_ACCEPT,
  filterValidMediaFiles,
} from "@/lib/mediaValidation";

interface PostModalProps {
  isOpen: boolean;
  onClose: () => void;
  currentUser: User | null;
  quotedPost?: Post | null;
  replyToPost?: Post | null;
  onPostSuccess: (newPost: Post) => void;
  onOpenAuth: () => void;
}

interface StagedAttachment {
  id: string;
  file: File;
  previewUrl: string;
  type: "image" | "video" | "gif";
}

export function PostModal({
  isOpen,
  onClose,
  currentUser,
  quotedPost,
  replyToPost,
  onPostSuccess,
  onOpenAuth,
}: PostModalProps) {
  const [content, setContent] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [stagedFiles, setStagedFiles] = useState<StagedAttachment[]>([]);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const maxChars = currentUser?.is_verified ? 4000 : 280;
  const charsRemaining = maxChars - content.length;
  const canPost =
    (content.trim().length > 0 || stagedFiles.length > 0 || Boolean(quotedPost)) &&
    charsRemaining >= 0 &&
    !isSubmitting;

  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = "hidden";
      setTimeout(() => textareaRef.current?.focus(), 50);
    } else {
      document.body.style.overflow = "";
    }
    return () => {
      document.body.style.overflow = "";
    };
  }, [isOpen]);

  useEffect(() => {
    return () => {
      stagedFiles.forEach((item) => URL.revokeObjectURL(item.previewUrl));
    };
  }, [stagedFiles]);

  if (!isOpen) return null;

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const rawFiles = Array.from(e.target.files || []);
    if (!rawFiles.length) return;

    if (!currentUser) {
      onClose();
      onOpenAuth();
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

  const handleSubmit = async () => {
    if (!canPost) return;

    if (!currentUser) {
      onClose();
      onOpenAuth();
      return;
    }

    setIsSubmitting(true);
    try {
      const uploadedMediaIds: string[] = [];

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
        quote_post_id: quotedPost ? quotedPost.id : null,
        reply_to_id: replyToPost ? replyToPost.id : null,
      };

      const res = await apiRequest<Post>("/posts", {
        method: "POST",
        body: JSON.stringify(payload),
      });

      stagedFiles.forEach((item) => URL.revokeObjectURL(item.previewUrl));
      setContent("");
      setStagedFiles([]);
      onClose();
      onPostSuccess(res);
    } catch (err: any) {
      alert(err.message || "Failed to post");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-12 sm:pt-16 p-4">
      {/* Dimmed Backdrop */}
      <div
        onClick={onClose}
        className="fixed inset-0 bg-[#5b7083]/40 dark:bg-black/60 backdrop-blur-sm transition-opacity"
      />

      {/* Modal Dialog Card */}
      <div className="relative w-full max-w-[600px] bg-background border border-border rounded-2xl shadow-2xl z-10 flex flex-col overflow-hidden animate-in fade-in zoom-in-95 duration-150">
        {/* Header: Close Button & Drafts Button (Screenshot) */}
        <div className="flex items-center justify-between px-4 py-3">
          <button
            onClick={onClose}
            className="p-2 hover:bg-subtle rounded-full text-foreground transition-colors"
            title="Close"
          >
            <X className="w-5 h-5" />
          </button>

          <button
            onClick={() => {}}
            className="text-brand font-bold text-[14px] hover:underline px-2 py-1 transition-colors"
          >
            Drafts
          </button>
        </div>

        {/* Content Body */}
        <div className="px-4 pb-3 flex flex-col">
          {/* If replying to a post: render parent post context with connecting line (Screenshot layout) */}
          {replyToPost && (
            <div className="flex gap-3 mb-2 relative">
              {/* Left Column: Avatar & Connecting Line */}
              <div className="flex flex-col items-center shrink-0">
                <img
                  src={
                    replyToPost.author?.avatar_url ||
                    "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"
                  }
                  alt={replyToPost.author?.display_name || "User"}
                  className="w-10 h-10 rounded-full object-cover border border-border"
                />
                <div className="w-0.5 grow bg-border my-1 min-h-[24px]" />
              </div>

              {/* Right Column: Author info, text snippet, and Replying to tag */}
              <div className="flex-1 min-w-0 pb-3">
                <div className="flex items-center gap-1.5 flex-wrap text-[15px] leading-5">
                  <span className="font-bold text-foreground">
                    {replyToPost.author?.display_name || replyToPost.author?.username || "User"}
                  </span>
                  {replyToPost.author?.is_verified && (
                    <span className="text-brand">●</span>
                  )}
                  <span className="text-muted">@{replyToPost.author?.username || "user"}</span>
                </div>
                {replyToPost.content && (
                  <p className="text-[15px] leading-5 text-foreground mt-1 whitespace-pre-wrap break-words line-clamp-4">
                    {replyToPost.content}
                  </p>
                )}
                <div className="text-muted text-[14px] mt-2">
                  <span>Replying to </span>
                  <span className="text-brand">@{replyToPost.author?.username || "user"}</span>
                </div>
              </div>
            </div>
          )}

          {/* User Reply Input Row */}
          <div className="flex gap-3">
            <img
              src={
                currentUser?.avatar_url ||
                "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"
              }
              alt={currentUser?.display_name || "User"}
              className="w-10 h-10 rounded-full object-cover shrink-0 border border-border"
            />

            <div className="flex-1 min-w-0">
              <textarea
                ref={textareaRef}
                value={content}
                onChange={(e) => setContent(e.target.value)}
                placeholder={replyToPost ? "Post your reply" : "What is happening?!"}
                rows={replyToPost ? 3 : 4}
                className="w-full bg-transparent text-[20px] text-foreground placeholder:text-muted outline-none resize-none leading-normal font-normal"
              />

            {/* Staged Attachments Grid */}
            {stagedFiles.length > 0 && (
              <div
                className={`grid gap-2 mb-3 ${
                  stagedFiles.length === 1 ? "grid-cols-1" : "grid-cols-2"
                }`}
              >
                {stagedFiles.map((staged, idx) => (
                  <div
                    key={staged.id}
                    className="relative rounded-2xl overflow-hidden border border-border bg-black max-h-[260px] group"
                  >
                    {staged.type === "video" ? (
                      <video
                        src={staged.previewUrl}
                        controls
                        className="w-full h-full object-cover max-h-[260px]"
                      />
                    ) : (
                      <img
                        src={staged.previewUrl}
                        alt="attachment"
                        className="w-full h-full object-cover max-h-[260px]"
                      />
                    )}
                    <button
                      onClick={() => handleRemoveStaged(idx)}
                      className="absolute top-2 right-2 p-1.5 rounded-full bg-black/70 hover:bg-black/90 text-white transition-colors"
                      title="Remove media"
                    >
                      <X className="w-4 h-4" />
                    </button>
                  </div>
                ))}
              </div>
            )}

            {/* Quoted Post Card Preview */}
            {quotedPost && (
              <div className="mb-3 p-3 rounded-2xl border border-border bg-subtle/30 flex flex-col gap-1.5 text-[14px]">
                <div className="flex items-center gap-1.5 font-bold text-foreground">
                  <img
                    src={quotedPost.author?.avatar_url || "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"}
                    alt={quotedPost.author?.display_name || "User"}
                    className="w-5 h-5 rounded-full object-cover"
                  />
                  <span>{quotedPost.author?.display_name || quotedPost.author?.username || "User"}</span>
                  <span className="font-normal text-muted">@{quotedPost.author?.username || "user"}</span>
                </div>
                {quotedPost.content && (
                  <p className="text-foreground/90 line-clamp-3 whitespace-pre-wrap">{quotedPost.content}</p>
                )}
              </div>
            )}

            {/* Audience Pill: Everyone can reply (Screenshot) */}
            <div className="flex items-center gap-1 text-brand font-bold text-[13px] hover:bg-brand/10 py-1 px-3 rounded-full w-fit cursor-pointer transition-colors -ml-3 mb-2">
              <Globe className="w-4 h-4 stroke-[2.2]" />
              <span>Everyone can reply</span>
            </div>

            {/* Bottom Toolbar & Post Action */}
            <div className="border-t border-border pt-3 flex items-center justify-between">
              {/* Media Action Icons */}
              <div className="flex items-center gap-1 -ml-2 text-brand">
                <input
                  type="file"
                  ref={fileInputRef}
                  onChange={handleFileChange}
                  multiple
                  accept={X_ALLOWED_MEDIA_ACCEPT}
                  className="hidden"
                />

                <button
                  type="button"
                  onClick={() => fileInputRef.current?.click()}
                  className="p-2 hover:bg-brand/10 rounded-full transition-colors"
                  title="Media"
                >
                  <ImageIcon className="w-5 h-5 stroke-[2]" />
                </button>

                <button
                  type="button"
                  className="p-2 hover:bg-brand/10 rounded-full transition-colors flex items-center justify-center"
                  title="GIF"
                >
                  <span className="font-extrabold text-[10px] tracking-wider border border-brand px-1 py-0.5 rounded leading-none">
                    GIF
                  </span>
                </button>

                <button
                  type="button"
                  className="p-2 hover:bg-brand/10 rounded-full transition-colors"
                  title="Poll"
                >
                  <BarChart2 className="w-5 h-5 stroke-[2] rotate-90" />
                </button>

                <button
                  type="button"
                  className="p-2 hover:bg-brand/10 rounded-full transition-colors"
                  title="Emoji"
                >
                  <Smile className="w-5 h-5 stroke-[2]" />
                </button>

                <button
                  type="button"
                  className="p-2 hover:bg-brand/10 rounded-full transition-colors"
                  title="Schedule"
                >
                  <Calendar className="w-5 h-5 stroke-[2]" />
                </button>

                <button
                  type="button"
                  className="p-2 hover:bg-brand/10 rounded-full transition-colors opacity-50 cursor-not-allowed"
                  title="Location"
                  disabled
                >
                  <MapPin className="w-5 h-5 stroke-[2]" />
                </button>
              </div>

              {/* Right: Submit Button (Screenshot styling) */}
              <div className="flex items-center gap-3">
                {content.length > 0 && (
                  <span
                    className={`text-xs ${
                      charsRemaining < 20 ? "text-red-500 font-bold" : "text-muted"
                    }`}
                  >
                    {charsRemaining}
                  </span>
                )}

                <button
                  onClick={handleSubmit}
                  disabled={!canPost}
                  className={`font-bold text-[15px] px-5 py-2 rounded-full transition-all ${
                    canPost
                      ? "bg-[#eff3f4] hover:bg-[#d7dbdc] text-[#0f1419] active:scale-95 shadow-sm"
                      : "bg-[#333639] text-[#71767b] cursor-not-allowed"
                  }`}
                >
                  {isSubmitting ? (
                    <div className="flex items-center gap-1.5">
                      <Loader2 className="w-4 h-4 animate-spin" />
                      <span>{replyToPost ? "Replying..." : "Posting..."}</span>
                    </div>
                  ) : (
                    replyToPost ? "Reply" : "Post"
                  )}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
  );
}
