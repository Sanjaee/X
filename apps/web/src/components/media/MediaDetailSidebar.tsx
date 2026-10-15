"use client";

import React, { useState } from "react";
import Link from "next/link";
import {
  MessageCircle,
  Repeat2,
  Heart,
  Bookmark,
  Share,
  BarChart2,
  MoreHorizontal,
  Loader2,
  Sparkles,
} from "lucide-react";
import { Post, User } from "@/types";
import { VerifiedBadge } from "@/components/common/VerifiedBadge";
import { formatRelativeTime, formatNumber, cn } from "@/lib/utils";
import { PostCard } from "@/components/feed/PostCard";
import { apiRequest } from "@/lib/api";

interface MediaDetailSidebarProps {
  post: Post;
  currentUser: User | null;
  isLiked: boolean;
  likesCount: number;
  isReposted: boolean;
  repostsCount: number;
  isBookmarked: boolean;
  onLike: () => void;
  onRepost: () => void;
  onBookmark: () => void;
  replies: Post[];
  isLoadingReplies?: boolean;
  onReplySuccess: (reply: Post) => void;
  onOpenAuth?: () => void;
}

export function MediaDetailSidebar({
  post,
  currentUser,
  isLiked,
  likesCount,
  isReposted,
  repostsCount,
  isBookmarked,
  onLike,
  onRepost,
  onBookmark,
  replies,
  isLoadingReplies,
  onReplySuccess,
  onOpenAuth,
}: MediaDetailSidebarProps) {
  const [replyContent, setReplyContent] = useState("");
  const [isSubmittingReply, setIsSubmittingReply] = useState(false);
  const [copied, setCopied] = useState(false);

  const author = post.author || {
    id: post.author_id,
    username: "user",
    display_name: "User",
    avatar_url: "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png",
    is_verified: false,
    verified_type: undefined,
  };

  const handleShare = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (typeof window !== "undefined") {
      navigator.clipboard.writeText(window.location.href);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const handleSendReply = async () => {
    if (!replyContent.trim() || isSubmittingReply) return;
    if (!currentUser) {
      if (onOpenAuth) onOpenAuth();
      return;
    }

    setIsSubmittingReply(true);
    try {
      const res = await apiRequest<Post>("/posts", {
        method: "POST",
        body: JSON.stringify({
          content: replyContent.trim(),
          reply_to_id: post.id,
        }),
      });

      const completeReply: Post = {
        ...res,
        author: res.author || currentUser,
      };

      setReplyContent("");
      onReplySuccess(completeReply);
    } catch (err: any) {
      alert(err.message || "Failed to post reply");
    } finally {
      setIsSubmittingReply(false);
    }
  };

  const dateObj = new Date(post.created_at || Date.now());
  const timeStr = dateObj.toLocaleTimeString("id-ID", {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  });
  const dateStr = dateObj.toLocaleDateString("id-ID", {
    day: "numeric",
    month: "short",
    year: "numeric",
  });
  const viewsCount = post.views_count || Math.max(120, (likesCount + 1) * 35);

  return (
    <aside className="w-full lg:w-[380px] xl:w-[420px] bg-background text-foreground border-t lg:border-t-0 lg:border-l border-border flex flex-col h-full overflow-y-auto shrink-0 select-text">
      {/* 1. Author Info Bar */}
      <div className="p-4 flex items-start justify-between gap-3">
        <Link
          href={`/${author.username}`}
          className="flex items-center gap-3 group shrink-0"
        >
          <img
            src={
              author.avatar_url ||
              "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"
            }
            alt={author.display_name}
            className="w-10 h-10 rounded-full object-cover border border-border group-hover:opacity-90 transition-opacity"
          />
          <div className="flex flex-col">
            <div className="flex items-center gap-1">
              <span className="font-bold text-[15px] leading-snug group-hover:underline">
                {author.display_name}
              </span>
              {author.is_verified && <VerifiedBadge type={author.verified_type} />}
            </div>
            <span className="text-muted text-[13px] leading-tight">
              @{author.username}
            </span>
          </div>
        </Link>

        <button
          className="p-2 hover:bg-subtle rounded-full text-muted hover:text-foreground transition-colors"
          title="More options"
        >
          <MoreHorizontal className="w-5 h-5" />
        </button>
      </div>

      {/* 2. Tweet Main Content */}
      <div className="px-4 pb-3">
        {post.content && (
          <p className="text-[16px] leading-relaxed whitespace-pre-line text-foreground font-normal break-words">
            {post.content}
          </p>
        )}

        {/* Translation Helper Banner (As observable in X screenshot) */}
        <div className="mt-2.5 flex items-center gap-1.5 text-xs text-muted cursor-pointer hover:underline">
          <Sparkles className="w-3.5 h-3.5 text-brand" />
          <span>Diterjemahkan dari Inggris · Tampilkan yang asli</span>
        </div>
      </div>

      {/* 3. Timestamp & Impressions (Observable X Layout) */}
      <div className="px-4 py-3 text-[14px] text-muted border-t border-border flex items-center gap-1.5 flex-wrap">
        <span>{timeStr}</span>
        <span>·</span>
        <span>{dateStr}</span>
        <span>·</span>
        <span className="font-bold text-foreground">{formatNumber(viewsCount)}</span>
        <span>Tayangan</span>
      </div>

      {/* 4. Metric Counts Bar */}
      <div className="px-4 py-2.5 border-t border-border flex items-center gap-5 text-[14px] text-muted overflow-x-auto">
        <div className="flex items-center gap-1">
          <span className="font-bold text-foreground">
            {formatNumber(post.replies_count || replies.length)}
          </span>
          <span>Balasan</span>
        </div>
        <div className="flex items-center gap-1">
          <span className="font-bold text-foreground">
            {formatNumber(repostsCount)}
          </span>
          <span>Repost</span>
        </div>
        <div className="flex items-center gap-1">
          <span className="font-bold text-foreground">
            {formatNumber(likesCount)}
          </span>
          <span>Suka</span>
        </div>
        <div className="flex items-center gap-1">
          <span className="font-bold text-foreground">
            {formatNumber(post.views_count ? Math.round(post.views_count / 100) : 103)}
          </span>
          <span>Markah</span>
        </div>
      </div>

      {/* 5. Interactive Action Icons Row */}
      <div className="px-6 py-2 border-t border-b border-border flex items-center justify-between text-muted">
        <button
          onClick={() => {
            const input = document.getElementById("media-reply-input");
            input?.focus();
          }}
          className="p-2 hover:bg-brand/10 hover:text-brand rounded-full transition-colors"
          title="Balas"
        >
          <MessageCircle className="w-5 h-5" />
        </button>

        <button
          onClick={onRepost}
          className={cn(
            "p-2 hover:bg-green-500/10 hover:text-green-500 rounded-full transition-colors",
            isReposted && "text-green-500"
          )}
          title="Repost"
        >
          <Repeat2 className="w-5 h-5" />
        </button>

        <button
          onClick={onLike}
          className={cn(
            "p-2 hover:bg-pink-500/10 hover:text-pink-500 rounded-full transition-colors",
            isLiked && "text-pink-500"
          )}
          title="Suka"
        >
          <Heart className={cn("w-5 h-5", isLiked && "fill-current")} />
        </button>

        <button
          onClick={onBookmark}
          className={cn(
            "p-2 hover:bg-brand/10 hover:text-brand rounded-full transition-colors",
            isBookmarked && "text-brand"
          )}
          title="Markah"
        >
          <Bookmark className={cn("w-5 h-5", isBookmarked && "fill-current")} />
        </button>

        <button
          onClick={handleShare}
          className="p-2 hover:bg-brand/10 hover:text-brand rounded-full transition-colors relative"
          title="Bagikan"
        >
          <Share className="w-5 h-5" />
          {copied && (
            <span className="absolute -top-7 left-1/2 -translate-x-1/2 bg-black text-white text-[11px] px-2 py-0.5 rounded shadow">
              Tautan disalin!
            </span>
          )}
        </button>
      </div>

      {/* 6. Inline Reply Composer Box (Identical to X) */}
      <div className="p-4 border-b border-border flex gap-3 items-start bg-subtle/20">
        <img
          src={
            currentUser?.avatar_url ||
            "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"
          }
          alt={currentUser?.display_name || "User"}
          className="w-10 h-10 rounded-full object-cover shrink-0 border border-border"
        />
        <div className="flex-1 flex flex-col gap-2">
          <textarea
            id="media-reply-input"
            value={replyContent}
            onChange={(e) => setReplyContent(e.target.value)}
            placeholder="Posting balasan"
            rows={2}
            className="w-full bg-transparent resize-none outline-none text-[15px] placeholder:text-muted leading-relaxed"
          />
          <div className="flex items-center justify-between pt-1">
            <span className="text-xs text-muted">
              Membalas <span className="text-brand">@{author.username}</span>
            </span>
            <button
              onClick={handleSendReply}
              disabled={!replyContent.trim() || isSubmittingReply}
              className="bg-brand hover:bg-brand-hover text-white text-sm font-bold py-1.5 px-4 rounded-full disabled:opacity-50 transition-colors flex items-center gap-1.5"
            >
              {isSubmittingReply && <Loader2 className="w-4 h-4 animate-spin" />}
              <span>Balas</span>
            </button>
          </div>
        </div>
      </div>

      {/* 7. Conversation Thread / Replies List */}
      <div className="flex-1 divide-y divide-border">
        {isLoadingReplies ? (
          <div className="p-8 flex items-center justify-center text-muted">
            <Loader2 className="w-6 h-6 animate-spin text-brand" />
          </div>
        ) : replies.length === 0 ? (
          <div className="p-8 text-center text-muted text-sm">
            Belum ada balasan. Jadilah yang pertama membalas!
          </div>
        ) : (
          replies.map((reply) => (
            <PostCard key={reply.id} post={reply} />
          ))
        )}
      </div>
    </aside>
  );
}
