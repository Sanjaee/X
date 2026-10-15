"use client";

import React, { useEffect, useCallback } from "react";
import { useMediaViewer } from "./MediaViewerContext";
import { PhotoCanvas } from "./PhotoCanvas";
import { PhotoNavigation } from "./PhotoNavigation";
import { VideoPlayer } from "./VideoPlayer";
import {
  X,
  MessageCircle,
  Repeat2,
  Heart,
  Bookmark,
  Share,
} from "lucide-react";
import Link from "next/link";
import { formatRelativeTime, formatNumber } from "@/lib/utils";

export function PhotoViewerModal() {
  const {
    isOpen,
    post,
    activeIndex,
    activeMedia,
    totalMedia,
    closeMedia,
    nextMedia,
    prevMedia,
  } = useMediaViewer();

  // Escape key handler
  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        e.preventDefault();
        closeMedia();
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isOpen, closeMedia]);

  if (!isOpen || !post || !activeMedia) return null;

  const isVideo = activeMedia.media_type === "video";

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-label="Media viewer"
      className="fixed inset-0 z-50 bg-black/95 flex flex-col select-none animate-in fade-in duration-150"
    >
      {/* Top Header Bar */}
      <header className="relative w-full h-14 flex items-center justify-between px-4 z-40 bg-gradient-to-b from-black/80 to-transparent">
        {/* Close Button (X) */}
        <button
          onClick={(e) => {
            e.stopPropagation();
            closeMedia();
          }}
          title="Close (Esc)"
          aria-label="Close media viewer"
          className="w-10 h-10 rounded-full hover:bg-white/10 active:bg-white/20 text-white flex items-center justify-center transition-colors"
        >
          <X className="w-5 h-5" />
        </button>

        {/* Post Author Info Header */}
        {post.author && (
          <div className="flex items-center gap-2.5">
            <Link
              href={`/${post.author.username}`}
              onClick={(e) => e.stopPropagation()}
              className="flex items-center gap-2 hover:opacity-90 transition-opacity"
            >
              <img
                src={
                  post.author.avatar_url ||
                  "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"
                }
                alt={post.author.display_name}
                className="w-8 h-8 rounded-full object-cover border border-white/10"
              />
              <div className="flex flex-col text-left">
                <span className="text-white text-xs font-bold leading-tight hover:underline">
                  {post.author.display_name}
                </span>
                <span className="text-white/60 text-[11px] leading-tight">
                  @{post.author.username} · {formatRelativeTime(post.created_at)}
                </span>
              </div>
            </Link>
          </div>
        )}

        {/* Spacer for symmetry */}
        <div className="w-10" />
      </header>

      {/* Main Canvas Area */}
      <main
        onClick={closeMedia}
        className="relative flex-1 w-full h-full flex items-center justify-center overflow-hidden"
      >
        <div
          onClick={(e) => e.stopPropagation()}
          className="relative w-full h-full flex items-center justify-center"
        >
          {isVideo ? (
            <div className="w-full h-full max-h-[85vh] max-w-[85vw] flex items-center justify-center">
              <VideoPlayer src={activeMedia.url} mode="viewer" autoPlay />
            </div>
          ) : (
            <PhotoCanvas media={activeMedia} />
          )}

          {/* Gallery Navigation Chevrons */}
          <PhotoNavigation
            currentIndex={activeIndex}
            total={totalMedia}
            onPrev={prevMedia}
            onNext={nextMedia}
          />
        </div>
      </main>

      {/* Bottom Footer & Post Context Action Bar */}
      <footer
        onClick={(e) => e.stopPropagation()}
        className="relative w-full bg-gradient-to-t from-black via-black/80 to-transparent pt-4 pb-4 px-6 z-40 flex flex-col items-center gap-2"
      >
        {/* Post Text Caption (if present) */}
        {post.content && (
          <p className="text-white text-sm max-w-2xl text-center line-clamp-2 text-white/90 font-normal">
            {post.content}
          </p>
        )}

        {/* Social Action Bar */}
        <div className="flex items-center justify-center gap-8 text-white/70 max-w-md w-full pt-1">
          {/* Reply */}
          <button
            title="Reply"
            className="flex items-center gap-1.5 hover:text-brand transition-colors text-xs"
          >
            <MessageCircle className="w-4 h-4" />
            <span>{post.replies_count > 0 ? formatNumber(post.replies_count) : ""}</span>
          </button>

          {/* Repost */}
          <button
            title="Repost"
            className={`flex items-center gap-1.5 transition-colors text-xs ${
              post.is_reposted ? "text-green-500" : "hover:text-green-500"
            }`}
          >
            <Repeat2 className="w-4 h-4" />
            <span>{post.reposts_count > 0 ? formatNumber(post.reposts_count) : ""}</span>
          </button>

          {/* Like */}
          <button
            title="Like"
            className={`flex items-center gap-1.5 transition-colors text-xs ${
              post.is_liked ? "text-pink-500" : "hover:text-pink-500"
            }`}
          >
            <Heart
              className={`w-4 h-4 ${post.is_liked ? "fill-pink-500 text-pink-500" : ""}`}
            />
            <span>{post.likes_count > 0 ? formatNumber(post.likes_count) : ""}</span>
          </button>

          {/* Bookmark */}
          <button
            title="Bookmark"
            className={`flex items-center gap-1.5 transition-colors text-xs ${
              post.is_bookmarked ? "text-brand" : "hover:text-brand"
            }`}
          >
            <Bookmark
              className={`w-4 h-4 ${post.is_bookmarked ? "fill-brand text-brand" : ""}`}
            />
          </button>

          {/* Share */}
          <button
            title="Share"
            onClick={() => {
              if (navigator.clipboard) {
                navigator.clipboard.writeText(window.location.href);
              }
            }}
            className="hover:text-brand transition-colors text-xs"
          >
            <Share className="w-4 h-4" />
          </button>
        </div>
      </footer>
    </div>
  );
}
