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
  BarChart2,
  Loader2,
  AlertCircle,
  ChevronLeft,
  ChevronRight,
  PanelRightClose,
  PanelRightOpen,
} from "lucide-react";
import { Post, MediaObject, User } from "@/types";
import { apiRequest } from "@/lib/api";
import { formatNumber, cn } from "@/lib/utils";
import { PhotoCanvas } from "@/components/media/PhotoCanvas";
import { VideoPlayer } from "@/components/media/VideoPlayer";
import { MediaDetailSidebar } from "@/components/media/MediaDetailSidebar";
import { AuthModal } from "@/components/auth/AuthModal";

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
  const [replies, setReplies] = useState<Post[]>([]);
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadingReplies, setLoadingReplies] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [isLiked, setIsLiked] = useState(false);
  const [likesCount, setLikesCount] = useState(0);
  const [isReposted, setIsReposted] = useState(false);
  const [repostsCount, setRepostsCount] = useState(0);
  const [isBookmarked, setIsBookmarked] = useState(false);
  const [showSidebar, setShowSidebar] = useState(true);
  const [authModalOpen, setAuthModalOpen] = useState(false);
  const [copied, setCopied] = useState(false);

  // Load current user
  useEffect(() => {
    apiRequest<User>("/me")
      .then((user) => setCurrentUser(user))
      .catch(() => setCurrentUser(null));
  }, []);

  // Fetch post and replies
  useEffect(() => {
    if (!postId) return;
    setLoading(true);
    setLoadingReplies(true);

    apiRequest<Post>(`/posts/${postId}`)
      .then((data) => {
        setPost(data);
        setIsLiked(data.is_liked);
        setLikesCount(data.likes_count);
        setIsReposted(data.is_reposted);
        setRepostsCount(data.reposts_count);
        setIsBookmarked(data.is_bookmarked);
        setLoading(false);
      })
      .catch((err) => {
        setError(err.message || "Failed to load post media");
        setLoading(false);
      });

    apiRequest<{ posts: Post[] }>(`/posts/${postId}/replies`)
      .then((res) => {
        setReplies(res.posts || []);
      })
      .catch(() => {})
      .finally(() => setLoadingReplies(false));
  }, [postId]);

  const handleClose = useCallback(() => {
    if (typeof window !== "undefined" && window.history.length > 1) {
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

  const handleLike = async () => {
    if (!currentUser) {
      setAuthModalOpen(true);
      return;
    }
    const prev = isLiked;
    setIsLiked(!prev);
    setLikesCount(prev ? Math.max(0, likesCount - 1) : likesCount + 1);
    try {
      await apiRequest(`/posts/${postId}/likes`, {
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
      await apiRequest(`/posts/${postId}/reposts`, {
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
      await apiRequest(`/posts/${postId}/bookmarks`, {
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
        <p className="text-white/60 mb-6 text-sm">
          {error || "This media item does not exist or has been removed."}
        </p>
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
  const viewsCount = post.views_count || Math.max(120, (likesCount + 1) * 35);

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-label="Direct media viewer"
      className="fixed inset-0 z-50 bg-black flex flex-col lg:flex-row select-none overflow-hidden"
    >
      {/* LEFT / CENTER MEDIA STAGE (Observable X Canvas) */}
      <div className="relative flex-1 h-full flex flex-col bg-black overflow-hidden">
        {/* Top-Left Close Button (X) */}
        <div className="absolute top-3 left-3 z-40 flex items-center gap-2">
          <button
            onClick={handleClose}
            title="Close (Esc)"
            aria-label="Close"
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

        {/* Main Media Canvas Area */}
        <div className="relative flex-1 w-full h-full flex items-center justify-center overflow-hidden">
          {isVideo ? (
            <div className="w-full h-full max-h-[85vh] max-w-[85vw] flex items-center justify-center">
              <VideoPlayer src={activeMedia.url} mode="viewer" autoPlay />
            </div>
          ) : (
            <PhotoCanvas media={activeMedia} />
          )}

          {/* Gallery Navigation Chevrons */}
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
                  className="absolute left-4 top-1/2 -translate-y-1/2 w-11 h-11 rounded-full bg-black/60 hover:bg-black/90 backdrop-blur-md border border-white/10 flex items-center justify-center text-white transition-all transform hover:scale-105 active:scale-95 shadow-xl z-30 group"
                >
                  <ChevronLeft className="w-6 h-6 group-hover:-translate-x-0.5 transition-transform" />
                </button>
              )}

              {activeIndex < total - 1 && (
                <button
                  onClick={() => handleNavigate(activeIndex + 1)}
                  title="Next photo (Arrow Right)"
                  aria-label="Next photo"
                  className="absolute right-4 top-1/2 -translate-y-1/2 w-11 h-11 rounded-full bg-black/60 hover:bg-black/90 backdrop-blur-md border border-white/10 flex items-center justify-center text-white transition-all transform hover:scale-105 active:scale-95 shadow-xl z-30 group"
                >
                  <ChevronRight className="w-6 h-6 group-hover:translate-x-0.5 transition-transform" />
                </button>
              )}
            </>
          )}
        </div>

        {/* Floating Bottom Engagement Bar (Observable X Layout) */}
        <div className="relative w-full py-3 px-6 z-30 flex items-center justify-center select-none bg-gradient-to-t from-black via-black/80 to-transparent">
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

            {/* Views / Impressions */}
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

      {/* RIGHT SIDEBAR (Full Tweet Details & Live Conversation Thread) */}
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
