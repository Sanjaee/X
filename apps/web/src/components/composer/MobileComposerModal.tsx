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
} from "lucide-react";
import { User, Post, MediaObject } from "@/types";
import { apiRequest } from "@/lib/api";

interface MobileComposerModalProps {
  isOpen: boolean;
  onClose: () => void;
  currentUser: User | null;
  onPostSuccess: (newPost: Post) => void;
  onOpenAuth?: () => void;
}

interface StagedAttachment {
  id: string;
  file: File;
  previewUrl: string;
  type: "image" | "video" | "gif";
}

export function MobileComposerModal({
  isOpen,
  onClose,
  currentUser,
  onPostSuccess,
  onOpenAuth,
}: MobileComposerModalProps) {
  const [content, setContent] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [stagedFiles, setStagedFiles] = useState<StagedAttachment[]>([]);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  // Focus textarea when modal opens
  useEffect(() => {
    if (isOpen) {
      setTimeout(() => textareaRef.current?.focus(), 150);
      document.body.style.overflow = "hidden";
    } else {
      document.body.style.overflow = "";
    }
    return () => {
      document.body.style.overflow = "";
    };
  }, [isOpen]);

  // Clean object URLs
  useEffect(() => {
    return () => {
      stagedFiles.forEach((f) => URL.revokeObjectURL(f.previewUrl));
    };
  }, [stagedFiles]);

  if (!isOpen) return null;

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files || []);
    if (!files.length) return;

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
    if ((!content.trim() && stagedFiles.length === 0) || isSubmitting) return;

    if (!currentUser) {
      onClose();
      onOpenAuth?.();
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

      const res = await apiRequest<Post>("/posts", {
        method: "POST",
        body: JSON.stringify({
          content: content.trim(),
          media_ids: uploadedMediaIds,
        }),
      });

      const completePost: Post = {
        ...res,
        author: res.author || currentUser,
      };

      stagedFiles.forEach((f) => URL.revokeObjectURL(f.previewUrl));
      setContent("");
      setStagedFiles([]);
      onPostSuccess(completePost);
      onClose();
    } catch (err: any) {
      alert(err.message || "Gagal memposting");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 bg-background text-foreground flex flex-col sm:hidden animate-in slide-in-from-bottom duration-200">
      {/* Top Header Bar */}
      <header className="flex items-center justify-between px-4 h-[53px] border-b border-border">
        <button
          onClick={onClose}
          className="text-foreground text-[16px] font-medium hover:opacity-80 transition-opacity"
        >
          Cancel
        </button>

        <button
          onClick={handleSubmit}
          disabled={(!content.trim() && stagedFiles.length === 0) || isSubmitting}
          className={`font-bold text-[14px] px-5 py-1.5 rounded-full transition-opacity flex items-center gap-1.5 shadow ${
            (!content.trim() && stagedFiles.length === 0) || isSubmitting
              ? "bg-[#333639] text-[#71767b] cursor-not-allowed"
              : "bg-[#eff3f4] text-[#0f1419] hover:bg-[#d7dbdc]"
          }`}
        >
          {isSubmitting && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
          <span>Post</span>
        </button>
      </header>

      {/* Main Composer Area */}
      <div className="flex-1 overflow-y-auto p-4 flex gap-3">
        <img
          src={
            currentUser?.avatar_url ||
            "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"
          }
          alt={currentUser?.display_name || "User"}
          className="w-10 h-10 rounded-full object-cover shrink-0 border border-border"
        />

        <div className="flex-1 flex flex-col">
          {/* Audience Pill */}
          <div className="inline-flex items-center gap-1 text-brand text-xs font-bold border border-brand/40 px-2.5 py-1 rounded-full w-fit mb-3">
            <Globe className="w-3 h-3" />
            <span>Everyone can reply</span>
          </div>

          <textarea
            ref={textareaRef}
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder="What is happening?!"
            rows={5}
            className="w-full bg-transparent text-[17px] text-foreground placeholder:text-muted outline-none resize-none leading-relaxed"
          />

          {/* Staged Media Attachments */}
          {stagedFiles.length > 0 && (
            <div
              className={`grid gap-2 mt-3 mb-2 rounded-2xl overflow-hidden border border-border ${
                stagedFiles.length === 1 ? "grid-cols-1" : "grid-cols-2"
              }`}
            >
              {stagedFiles.map((file, idx) => (
                <div key={file.id} className="relative aspect-video bg-black/30">
                  {file.type === "video" ? (
                    <video
                      src={file.previewUrl}
                      className="w-full h-full object-cover"
                      muted
                      autoPlay
                      loop
                    />
                  ) : (
                    <img
                      src={file.previewUrl}
                      alt="attachment preview"
                      className="w-full h-full object-cover"
                    />
                  )}
                  <button
                    onClick={() => handleRemoveStaged(idx)}
                    className="absolute top-2 right-2 w-7 h-7 rounded-full bg-black/70 hover:bg-black text-white flex items-center justify-center transition-colors"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Bottom Media Toolbar */}
      <footer className="border-t border-border px-4 py-3 flex items-center justify-between pb-safe">
        <div className="flex items-center gap-3 text-brand">
          <input
            type="file"
            ref={fileInputRef}
            onChange={handleFileChange}
            accept="image/*,video/*"
            multiple
            className="hidden"
          />
          <button
            onClick={() => fileInputRef.current?.click()}
            className="p-2 hover:bg-brand/10 rounded-full transition-colors"
            title="Tambah foto atau video"
          >
            <ImageIcon className="w-5 h-5" />
          </button>
          <button
            className="p-2 hover:bg-brand/10 rounded-full transition-colors"
            title="Tambah GIF"
          >
            <Smile className="w-5 h-5" />
          </button>
          <button
            className="p-2 hover:bg-brand/10 rounded-full transition-colors"
            title="Jadwalkan"
          >
            <Calendar className="w-5 h-5" />
          </button>
          <button
            className="p-2 hover:bg-brand/10 rounded-full transition-colors"
            title="Lokasi"
          >
            <MapPin className="w-5 h-5" />
          </button>
        </div>

        <div className="text-xs text-muted">
          {currentUser?.is_verified ? "4000" : "280"}
        </div>
      </footer>
    </div>
  );
}
