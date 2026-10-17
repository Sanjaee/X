"use client";

import React, { useEffect, useState, useCallback } from "react";
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
  BarChart2,
  PanelRightClose,
  PanelRightOpen,
} from "lucide-react";
import { Post, User } from "@/types";
import { apiRequest } from "@/lib/api";
import { formatNumber, cn } from "@/lib/utils";
import { MediaDetailSidebar } from "./MediaDetailSidebar";
import { AuthModal } from "@/components/auth/AuthModal";

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

  const [replies, setReplies] = useState<Post[]>([]);
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [loadingReplies, setLoadingReplies] = useState(false);
  const [showSidebar, setShowSidebar] = useState(true);
  const [authModalOpen, setAuthModalOpen] = useState(false);
  const [copied, setCopied] = useState(false);

  const [isLiked, setIsLiked] = useState(false);
  const [likesCount, setLikesCount] = useState(0);
  const [isReposted, setIsReposted] = useState(false);
  const [repostsCount, setRepostsCount] = useState(0);
  const [isBookmarked, setIsBookmarked] = useState(false);

  // Sync post state when opened
  useEffect(() => {
    if (!isOpen || !post) return;
    setIsLiked(post.is_liked);
    setLikesCount(post.likes_count);
    setIsReposted(post.is_reposted);
    setRepostsCount(post.reposts_count);
    setIsBookmarked(post.is_bookmarked);

    apiRequest<User>("/me")
      .then((user) => setCurrentUser(user))
      .catch(() => setCurrentUser(null));

    setLoadingReplies(true);
    apiRequest<{ posts: Post[] }>(`/posts/${post.id}/replies`)
      .then((res) => setReplies(res.posts || []))
      .catch(() => {})
      .finally(() => setLoadingReplies(false));
  }, [isOpen, post]);

  // Escape key handler
  useEffect(() => {
    if (!isOpen) return;

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
        closeMedia();
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isOpen, closeMedia]);

  if (!isOpen || !post || !activeMedia) return null;

  const isVideo = activeMedia.media_type === "video";
  const viewsCount = post.views_count || Math.max(120, (likesCount + 1) * 35);

  const handleLike = async () => {
    if (!currentUser) {
      setAuthModalOpen(true);
      return;
    }
    const prev = isLiked;
    setIsLiked(!prev);
    setLikesCount(prev ? Math.max(0, likesCount - 1) : likesCount + 1);
    try {
      await apiRequest(`/posts/${post.id}/likes`, {
        method: prev ? "DELETE" : "POST",
      });
    } catch {
      setIsLiked(prev);
    }
  };

  const handleRepost = async () => {
    if (!currentUser) {
      setAuthModalOpen(true);
      return;
    }
    const prev = isReposted;
    setIsReposted(!prev);
    setRepostsCount(prev ? Math.max(0, repostsCount - 1) : repostsCount + 1);
    try {
      await apiRequest(`/posts/${post.id}/reposts`, {
        method: prev ? "DELETE" : "POST",
      });
    } catch {
      setIsReposted(prev);
    }
  };

  const handleBookmark = async () => {
    if (!currentUser) {
      setAuthModalOpen(true);
      return;
    }
    const prev = isBookmarked;
    setIsBookmarked(!prev);
    try {
      await apiRequest(`/posts/${post.id}/bookmarks`, {
        method: prev ? "DELETE" : "POST",
      });
    } catch {
      setIsBookmarked(prev);
    }
  };

  const handleShare = () => {
    if (typeof window !== "undefined") {
      navigator.clipboard.writeText(window.location.href);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-label="Media viewer"
      className="fixed inset-0 z-50 bg-black flex flex-col lg:flex-row select-none animate-in fade-in duration-150 overflow-hidden"
    >
      {/* LEFT / CENTER MEDIA STAGE (Observable X Canvas) */}
      <div className="relative flex-1 h-full flex flex-col bg-black overflow-hidden">
        {/* Top-Left Close Button (X) */}
        <div className="absolute top-3 left-3 z-40 flex items-center gap-2">
          <button
            onClick={(e) => {
              e.stopPropagation();
              closeMedia();
            }}
            title="Close (Esc)"
            aria-label="Close media viewer"
            className="w-10 h-10 rounded-full bg-black/60 hover:bg-black/90 active:bg-black/40 text-white flex items-center justify-center transition-all shadow-md backdrop-blur-md border border-white/10 group"
          >
            <X className="w-5 h-5 group-hover:scale-110 transition-transform" />
          </button>
        </div>

        {/* Top-Right Toggle Sidebar Button */}
        <div className="absolute top-3 right-3 z-40 hidden lg:flex items-center gap-2">
          <button
            onClick={() => setShowSidebar(!showSidebar)}
            title={showSidebar ? "Hide details" : "Show details"}
            className="w-10 h-10 rounded-full bg-black/60 hover:bg-black/90 active:bg-black/40 text-white flex items-center justify-center transition-all shadow-md backdrop-blur-md border border-white/10"
          >
            {showSidebar ? (
              <PanelRightClose className="w-5 h-5" />
            ) : (
              <PanelRightOpen className="w-5 h-5" />
            )}
          </button>
        </div>

        {/* Main Canvas Area */}
        <div
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
        </div>

        {/* Floating Bottom Engagement Bar (Observable X Layout) */}
        <div
          onClick={(e) => e.stopPropagation()}
          className="relative w-full py-3 px-6 z-30 flex items-center justify-center select-none bg-gradient-to-t from-black via-black/80 to-transparent"
        >
          <div className="flex items-center justify-between w-full max-w-lg text-white/80 px-4">
            {/* Reply */}
            <button
              onClick={() => {
                if (!showSidebar) setShowSidebar(true);
                const input = document.getElementById("media-reply-input");
                input?.focus();
              }}
              className="flex items-center gap-2 hover:text-brand transition-colors text-xs font-medium group"
              title="Reply"
            >
              <div className="p-2 rounded-full group-hover:bg-brand/20 transition-colors">
                <MessageCircle className="w-4 h-4" />
              </div>
              <span>{formatNumber(post.replies_count || replies.length)}</span>
            </button>

            {/* Repost */}
            <button
              onClick={handleRepost}
              className={cn(
                "flex items-center gap-2 hover:text-green-500 transition-colors text-xs font-medium group",
                isReposted && "text-green-500"
              )}
              title="Repost"
            >
              <div className="p-2 rounded-full group-hover:bg-green-500/20 transition-colors">
                <Repeat2 className="w-4 h-4" />
              </div>
              <span>{formatNumber(repostsCount)}</span>
            </button>

            {/* Like */}
            <button
              onClick={handleLike}
              className={cn(
                "flex items-center gap-2 hover:text-pink-500 transition-colors text-xs font-medium group",
                isLiked && "text-pink-500"
              )}
              title="Like"
            >
              <div className="p-2 rounded-full group-hover:bg-pink-500/20 transition-colors">
                <Heart className={cn("w-4 h-4", isLiked && "fill-current")} />
              </div>
              <span>{formatNumber(likesCount)}</span>
            </button>

            {/* Views */}
            <div className="flex items-center gap-2 text-xs font-medium text-white/60">
              <div className="p-2">
                <BarChart2 className="w-4 h-4" />
              </div>
              <span>{formatNumber(viewsCount)}</span>
            </div>

            {/* Bookmark */}
            <button
              onClick={handleBookmark}
              className={cn(
                "p-2 rounded-full hover:bg-brand/20 hover:text-brand transition-colors",
                isBookmarked && "text-brand"
              )}
              title="Bookmark"
            >
              <Bookmark className={cn("w-4 h-4", isBookmarked && "fill-current")} />
            </button>

            {/* Share */}
            <button
              onClick={handleShare}
              className="p-2 rounded-full hover:bg-brand/20 hover:text-brand transition-colors relative"
              title="Share"
            >
              <Share className="w-4 h-4" />
              {copied && (
                <span className="absolute -top-7 left-1/2 -translate-x-1/2 bg-white text-black text-[11px] px-2 py-0.5 rounded shadow font-bold">
                  Link copied!
                </span>
              )}
            </button>
          </div>
        </div>
      </div>

      {/* RIGHT SIDEBAR (Observable X Layout) */}
      {showSidebar && (
        <MediaDetailSidebar
          post={post}
          currentUser={currentUser}
          isLiked={isLiked}
          likesCount={likesCount}
          isReposted={isReposted}
          repostsCount={repostsCount}
          isBookmarked={isBookmarked}
          onLike={handleLike}
          onRepost={handleRepost}
          onBookmark={handleBookmark}
          replies={replies}
          isLoadingReplies={loadingReplies}
          onReplySuccess={(newReply) => setReplies((prev) => [newReply, ...prev])}
          onOpenAuth={() => setAuthModalOpen(true)}
        />
      )}

      {/* Auth Modal for Unauthenticated Users */}
      <AuthModal
        isOpen={authModalOpen}
        onClose={() => setAuthModalOpen(false)}
        onSuccess={(u) => setCurrentUser(u)}
      />
    </div>
  );
}
