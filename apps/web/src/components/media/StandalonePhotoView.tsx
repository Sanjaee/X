"use client";

import React, { useEffect, useState, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import {
  X,
  MessageCircle,
  Repeat2,
  Heart,
  Bookmark,
  Share,
  Loader2,
  AlertCircle,
  ChevronLeft,
  ChevronRight,
} from "lucide-react";
import { Post, MediaObject } from "@/types";
import { apiRequest } from "@/lib/api";
import { formatRelativeTime, formatNumber } from "@/lib/utils";
import { PhotoCanvas } from "@/components/media/PhotoCanvas";
import { VideoPlayer } from "@/components/media/VideoPlayer";

export function StandalonePhotoView() {
  const params = useParams();
  const router = useRouter();

  const postId = params.postId as string;
  const username = (params.username as string) || "user";
  const indexParam = parseInt(
    (params.photoIndex as string) || (params.videoIndex as string) || "1",
    10
  );
  const activeIndex = Math.max(0, indexParam - 1);

  const [post, setPost] = useState<Post | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!postId) return;
    setLoading(true);
    apiRequest<Post>(`/posts/${postId}`)
      .then((data) => {
        setPost(data);
        setLoading(false);
      })
      .catch((err) => {
        setError(err.message || "Failed to load post media");
        setLoading(false);
      });
  }, [postId]);

  const handleClose = useCallback(() => {
    if (window.history.length > 1) {
      router.back();
    } else {
      router.push(`/${username}/status/${postId}`);
    }
  }, [router, username, postId]);

  const handleNavigate = useCallback(
    (newIndex: number) => {
      if (!post?.media || newIndex < 0 || newIndex >= post.media.length) return;
      const mediaItem = post.media[newIndex];
      const type = mediaItem.media_type === "video" ? "video" : "photo";
      router.replace(`/${username}/status/${postId}/${type}/${newIndex + 1}`);
    },
    [router, username, postId, post]
  );

  // Keyboard controls
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (
        document.activeElement instanceof HTMLInputElement ||
        document.activeElement instanceof HTMLTextAreaElement ||
        (document.activeElement as HTMLElement)?.isContentEditable
      ) {
        return;
      }

      if (e.key === "Escape") {
        e.preventDefault();
        handleClose();
      } else if (e.key === "ArrowLeft" && post?.media && activeIndex > 0) {
        e.preventDefault();
        handleNavigate(activeIndex - 1);
      } else if (
        e.key === "ArrowRight" &&
        post?.media &&
        activeIndex < post.media.length - 1
      ) {
        e.preventDefault();
        handleNavigate(activeIndex + 1);
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [handleClose, handleNavigate, post, activeIndex]);

  if (loading) {
    return (
      <div className="fixed inset-0 bg-black flex items-center justify-center z-50">
        <Loader2 className="w-10 h-10 animate-spin text-brand" />
      </div>
    );
  }

  if (error || !post || !post.media || post.media.length === 0) {
    return (
      <div className="fixed inset-0 bg-black flex flex-col items-center justify-center p-6 text-white z-50">
        <AlertCircle className="w-12 h-12 text-red-500 mb-3" />
        <h2 className="text-xl font-bold mb-2">Media not found</h2>
        <p className="text-white/60 mb-6 text-sm">{error || "This media item does not exist or has been removed."}</p>
        <button
          onClick={handleClose}
          className="bg-brand hover:bg-brand-hover text-white font-bold py-2.5 px-6 rounded-full text-sm transition-colors"
        >
          Return to post
        </button>
      </div>
    );
  }

  const activeMedia: MediaObject = post.media[activeIndex] || post.media[0];
  const total = post.media.length;
  const isVideo = activeMedia.media_type === "video";

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-label="Direct media viewer"
      className="fixed inset-0 z-50 bg-black flex flex-col select-none"
    >
      {/* Header */}
      <header className="relative w-full h-14 flex items-center justify-between px-4 z-40 bg-gradient-to-b from-black/80 to-transparent">
        <button
          onClick={handleClose}
          title="Close (Esc)"
          aria-label="Close"
          className="w-10 h-10 rounded-full hover:bg-white/10 active:bg-white/20 text-white flex items-center justify-center transition-colors"
        >
          <X className="w-5 h-5" />
        </button>

        {post.author && (
          <Link
            href={`/${post.author.username}`}
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
        )}

        <div className="w-10" />
      </header>

      {/* Main Canvas Area */}
      <main className="relative flex-1 w-full h-full flex items-center justify-center overflow-hidden">
        {isVideo ? (
          <div className="w-full h-full max-h-[85vh] max-w-[85vw] flex items-center justify-center">
            <VideoPlayer src={activeMedia.url} mode="viewer" autoPlay />
          </div>
        ) : (
          <PhotoCanvas media={activeMedia} />
        )}

        {/* Gallery Navigation */}
        {total > 1 && (
          <>
            <div className="absolute top-4 left-1/2 -translate-x-1/2 bg-black/60 backdrop-blur-md border border-white/10 rounded-full px-3 py-1 text-white text-xs font-semibold select-none z-30">
              {activeIndex + 1} / {total}
            </div>

            {activeIndex > 0 && (
              <button
                onClick={() => handleNavigate(activeIndex - 1)}
                title="Previous photo (Arrow Left)"
                aria-label="Previous photo"
                className="absolute left-4 top-1/2 -translate-y-1/2 w-11 h-11 rounded-full bg-black/60 hover:bg-black/80 backdrop-blur-md border border-white/10 flex items-center justify-center text-white transition-all transform hover:scale-105 active:scale-95 shadow-xl z-30 group"
              >
                <ChevronLeft className="w-6 h-6 group-hover:-translate-x-0.5 transition-transform" />
              </button>
            )}

            {activeIndex < total - 1 && (
              <button
                onClick={() => handleNavigate(activeIndex + 1)}
                title="Next photo (Arrow Right)"
                aria-label="Next photo"
                className="absolute right-4 top-1/2 -translate-y-1/2 w-11 h-11 rounded-full bg-black/60 hover:bg-black/80 backdrop-blur-md border border-white/10 flex items-center justify-center text-white transition-all transform hover:scale-105 active:scale-95 shadow-xl z-30 group"
              >
                <ChevronRight className="w-6 h-6 group-hover:translate-x-0.5 transition-transform" />
              </button>
            )}
          </>
        )}
      </main>

      {/* Footer */}
      <footer className="relative w-full bg-gradient-to-t from-black via-black/80 to-transparent pt-4 pb-4 px-6 z-40 flex flex-col items-center gap-2">
        {post.content && (
          <p className="text-white text-sm max-w-2xl text-center line-clamp-2 text-white/90">
            {post.content}
          </p>
        )}

        <div className="flex items-center justify-center gap-8 text-white/70 max-w-md w-full pt-1">
          <Link
            href={`/${username}/status/${postId}`}
            className="flex items-center gap-1.5 hover:text-brand transition-colors text-xs"
          >
            <MessageCircle className="w-4 h-4" />
            <span>{post.replies_count > 0 ? formatNumber(post.replies_count) : ""}</span>
          </Link>
          <div className="flex items-center gap-1.5 text-xs">
            <Repeat2 className="w-4 h-4" />
            <span>{post.reposts_count > 0 ? formatNumber(post.reposts_count) : ""}</span>
          </div>
          <div className="flex items-center gap-1.5 text-xs">
            <Heart className="w-4 h-4" />
            <span>{post.likes_count > 0 ? formatNumber(post.likes_count) : ""}</span>
          </div>
          <div className="flex items-center gap-1.5 text-xs">
            <Bookmark className="w-4 h-4" />
          </div>
          <button
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
