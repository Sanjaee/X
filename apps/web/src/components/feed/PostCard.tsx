"use client";

import React, { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  MessageCircle,
  Repeat2,
  Heart,
  BarChart2,
  Bookmark,
  Share,
  CheckCircle2,
  MoreHorizontal,
} from "lucide-react";
import { Post } from "@/types";
import { formatRelativeTime, formatNumber, cn } from "@/lib/utils";
import { VerifiedBadge } from "@/components/common/VerifiedBadge";
import { useMediaViewer } from "@/components/media/MediaViewerContext";
import { VideoPlayer } from "@/components/media/VideoPlayer";

interface PostCardProps {
  post: Post;
  onLike?: (postId: string) => void;
  onRepost?: (postId: string) => void;
  onBookmark?: (postId: string) => void;
  onFeedback?: (postId: string, feedbackType: 'not_interested' | 'hide') => void;
}

export function PostCard({ post, onLike, onRepost, onBookmark, onFeedback }: PostCardProps) {
  const router = useRouter();
  const { openMedia } = useMediaViewer();
  const [isLiked, setIsLiked] = useState(post.is_liked);
  const [likesCount, setLikesCount] = useState(post.likes_count);
  const [isReposted, setIsReposted] = useState(post.is_reposted);
  const [repostsCount, setRepostsCount] = useState(post.reposts_count);
  const [isBookmarked, setIsBookmarked] = useState(post.is_bookmarked);
  const [menuOpen, setMenuOpen] = useState(false);

  const handleCardClick = (e: React.MouseEvent) => {
    // Avoid triggering card navigation if interactive child clicked
    const target = e.target as HTMLElement;
    if (target.closest("button") || target.closest("a")) return;
    router.push(`/${post.author?.username || "user"}/status/${post.id}`);
  };

  const handleLike = (e: React.MouseEvent) => {
    e.stopPropagation();
    setIsLiked(!isLiked);
    setLikesCount(isLiked ? Math.max(0, likesCount - 1) : likesCount + 1);
    if (onLike) onLike(post.id);
  };

  const handleRepost = (e: React.MouseEvent) => {
    e.stopPropagation();
    setIsReposted(!isReposted);
    setRepostsCount(isReposted ? Math.max(0, repostsCount - 1) : repostsCount + 1);
    if (onRepost) onRepost(post.id);
  };

  const handleBookmark = (e: React.MouseEvent) => {
    e.stopPropagation();
    setIsBookmarked(!isBookmarked);
    if (onBookmark) onBookmark(post.id);
  };

  return (
    <article
      data-post-id={post.id}
      onClick={handleCardClick}
      className="flex gap-3 px-4 py-3 border-b border-border hover:bg-subtle/50 transition-colors cursor-pointer select-none"
    >
      {/* Left Avatar */}
      <Link
        href={`/${post.author?.username || "user"}`}
        onClick={(e) => e.stopPropagation()}
        className="shrink-0"
      >
        <img
          src={post.author?.avatar_url || "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"}
          alt={post.author?.display_name || "User"}
          className="w-10 h-10 rounded-full object-cover hover:opacity-90 transition-opacity border border-border"
        />
      </Link>

      {/* Content Column */}
      <div className="flex-1 min-w-0">
        {/* Author Header */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-1.5 flex-wrap text-[15px] leading-5">
            <Link
              href={`/${post.author?.username || "user"}`}
              onClick={(e) => e.stopPropagation()}
              className="font-bold text-foreground hover:underline line-clamp-1"
            >
              {post.author?.display_name || post.author?.username || "User"}
            </Link>
            {post.author?.is_verified && (
              <VerifiedBadge type={post.author.verified_type} />
            )}

            <span className="text-muted">@{post.author?.username || "user"}</span>
            <span className="text-muted">·</span>
            <span className="text-muted hover:underline">{formatRelativeTime(post.created_at)}</span>
          </div>

          <div className="relative">
            <button
              onClick={(e) => {
                e.stopPropagation();
                setMenuOpen(!menuOpen);
              }}
              className="p-1 hover:bg-subtle rounded-full text-muted hover:text-foreground transition-colors"
            >
              <MoreHorizontal className="w-4 h-4" />
            </button>

            {menuOpen && (
              <>
                <div
                  className="fixed inset-0 z-30"
                  onClick={(e) => {
                    e.stopPropagation();
                    setMenuOpen(false);
                  }}
                />
                <div className="absolute right-0 top-7 w-52 bg-background border border-border shadow-xl rounded-xl py-1.5 z-40 text-[14px]">
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      setMenuOpen(false);
                      if (onFeedback) onFeedback(post.id, 'not_interested');
                    }}
                    className="w-full text-left px-3.5 py-2.5 hover:bg-subtle flex items-center gap-2.5 text-foreground transition-colors"
                  >
                    <span>Not interested in this post</span>
                  </button>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      setMenuOpen(false);
                      if (onFeedback) onFeedback(post.id, 'hide');
                    }}
                    className="w-full text-left px-3.5 py-2.5 hover:bg-subtle flex items-center gap-2.5 text-foreground transition-colors"
                  >
                    <span>Hide post</span>
                  </button>
                </div>
              </>
            )}
          </div>
        </div>

        {/* Post Text */}
        <p className="text-[15px] leading-5 text-foreground mt-1 whitespace-pre-wrap break-words">
          {post.content}
        </p>

        {/* Media Attachments with CLS-Protected Aspect Ratio & Photo Viewer */}
        {post.media && post.media.length > 0 && (
          <div className="mt-3 rounded-2xl overflow-hidden border border-border bg-muted/20">
            {post.media.length === 1 && (
              post.media[0].media_type === "video" ? (
                <div
                  onClick={(e) => e.stopPropagation()}
                  className="w-full aspect-video bg-black flex items-center justify-center"
                >
                  <VideoPlayer
                    src={post.media[0].url}
                    mode="in_feed"
                    onOpenViewer={() => openMedia(post, 0)}
                  />
                </div>
              ) : (
                <div
                  onClick={(e) => {
                    e.stopPropagation();
                    openMedia(post, 0);
                  }}
                  className="w-full bg-muted/30 relative overflow-hidden cursor-pointer group"
                  style={{
                    aspectRatio:
                      post.media[0].width && post.media[0].height
                        ? `${post.media[0].width}/${post.media[0].height}`
                        : "16/9",
                    maxHeight: "510px",
                  }}
                >
                  <img
                    src={post.media[0].url}
                    alt="Attachment"
                    loading="lazy"
                    decoding="async"
                    className="w-full h-full object-cover group-hover:opacity-95 transition-opacity"
                  />
                </div>
              )
            )}
            {post.media.length > 1 && (
              <div className="grid grid-cols-2 gap-0.5 aspect-video bg-muted/20">
                {post.media.slice(0, 4).map((m, idx) => (
                  <div
                    key={m.id || idx}
                    onClick={(e) => {
                      e.stopPropagation();
                      openMedia(post, idx);
                    }}
                    className="relative w-full h-full overflow-hidden bg-muted/30 cursor-pointer group"
                  >
                    {m.media_type === "video" ? (
                      <VideoPlayer
                        src={m.url}
                        mode="in_feed"
                        onOpenViewer={() => openMedia(post, idx)}
                      />
                    ) : (
                      <img
                        src={m.url}
                        alt="Attachment"
                        loading="lazy"
                        decoding="async"
                        className="w-full h-full object-cover group-hover:opacity-95 transition-opacity"
                      />
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* Engagement Action Bar */}
        <div className="flex items-center justify-between text-muted mt-3 max-w-[425px]">
          {/* Reply */}
          <button
            onClick={(e) => e.stopPropagation()}
            className="group flex items-center gap-1.5 hover:text-brand transition-colors text-[13px]"
          >
            <div className="p-2 rounded-full group-hover:bg-brand/10 transition-colors">
              <MessageCircle className="w-4 h-4" />
            </div>
            <span>{post.replies_count > 0 ? formatNumber(post.replies_count) : ""}</span>
          </button>

          {/* Repost */}
          <button
            onClick={handleRepost}
            className={cn(
              "group flex items-center gap-1.5 transition-colors text-[13px]",
              isReposted ? "text-repost" : "hover:text-repost"
            )}
          >
            <div className="p-2 rounded-full group-hover:bg-repost/10 transition-colors">
              <Repeat2 className="w-4 h-4" />
            </div>
            <span>{repostsCount > 0 ? formatNumber(repostsCount) : ""}</span>
          </button>

          {/* Like */}
          <button
            onClick={handleLike}
            className={cn(
              "group flex items-center gap-1.5 transition-colors text-[13px]",
              isLiked ? "text-like" : "hover:text-like"
            )}
          >
            <div className="p-2 rounded-full group-hover:bg-like/10 transition-colors">
              <Heart className={cn("w-4 h-4 transition-transform active:scale-125", isLiked && "fill-current")} />
            </div>
            <span>{likesCount > 0 ? formatNumber(likesCount) : ""}</span>
          </button>

          {/* Views */}
          <div className="group flex items-center gap-1.5 text-[13px] hover:text-brand transition-colors">
            <div className="p-2 rounded-full group-hover:bg-brand/10 transition-colors">
              <BarChart2 className="w-4 h-4" />
            </div>
            <span>{post.views_count > 0 ? formatNumber(post.views_count) : "0"}</span>
          </div>

          {/* Bookmark & Share */}
          <div className="flex items-center">
            <button
              onClick={handleBookmark}
              className={cn(
                "p-2 rounded-full hover:bg-brand/10 hover:text-brand transition-colors",
                isBookmarked && "text-brand"
              )}
            >
              <Bookmark className={cn("w-4 h-4", isBookmarked && "fill-current")} />
            </button>
            <button
              onClick={(e) => e.stopPropagation()}
              className="p-2 rounded-full hover:bg-brand/10 hover:text-brand transition-colors"
            >
              <Share className="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>
    </article>
  );
}
