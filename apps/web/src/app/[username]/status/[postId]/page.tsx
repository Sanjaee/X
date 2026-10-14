"use client";

import React, { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import {
  ArrowLeft,
  CheckCircle2,
  Heart,
  Repeat2,
  Bookmark,
  Share,
  MessageCircle,
  MoreHorizontal,
  Loader2,
} from "lucide-react";
import { LeftNavRail } from "@/components/layout/LeftNavRail";
import { RightSidebar } from "@/components/layout/RightSidebar";
import { PostComposer } from "@/components/composer/PostComposer";
import { PostCard } from "@/components/feed/PostCard";
import { AuthModal } from "@/components/auth/AuthModal";
import { User, Post } from "@/types";
import { apiRequest } from "@/lib/api";
import { formatFullTimestamp, formatNumber, cn } from "@/lib/utils";

export default function PostDetailPage() {
  const params = useParams();
  const router = useRouter();
  const postId = params.postId as string;

  const [post, setPost] = useState<Post | null>(null);
  const [replies, setReplies] = useState<Post[]>([]);
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [authModalOpen, setAuthModalOpen] = useState(false);

  const [isLiked, setIsLiked] = useState(false);
  const [likesCount, setLikesCount] = useState(0);
  const [isReposted, setIsReposted] = useState(false);
  const [repostsCount, setRepostsCount] = useState(0);
  const [isBookmarked, setIsBookmarked] = useState(false);

  useEffect(() => {
    apiRequest<User>("/me")
      .then((user) => setCurrentUser(user))
      .catch(() => setCurrentUser(null));
  }, []);

  useEffect(() => {
    if (!postId) return;
    setLoading(true);

    Promise.all([
      apiRequest<Post>(`/posts/${postId}`),
      apiRequest<{ posts: Post[] }>(`/posts/${postId}/replies`),
    ])
      .then(([postData, repliesData]) => {
        setPost(postData);
        setIsLiked(postData.is_liked);
        setLikesCount(postData.likes_count);
        setIsReposted(postData.is_reposted);
        setRepostsCount(postData.reposts_count);
        setIsBookmarked(postData.is_bookmarked);
        setReplies(repliesData.posts || []);
      })
      .catch((err) => {
        console.error("Failed to load post detail", err);
      })
      .finally(() => {
        setLoading(false);
      });
  }, [postId]);

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
    } catch (e) {
      setIsLiked(prev);
      setLikesCount(likesCount);
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
    } catch (e) {
      setIsReposted(prev);
      setRepostsCount(repostsCount);
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
    } catch (e) {
      setIsBookmarked(prev);
    }
  };

  const handleReplyCreated = (newReply: Post) => {
    setReplies((prev) => [newReply, ...prev]);
  };

  return (
    <div className="min-h-screen bg-background text-foreground flex justify-center">
      <div className="w-full max-w-[1280px] flex">
        <LeftNavRail
          currentUser={currentUser}
          onOpenAuth={() => setAuthModalOpen(true)}
        />

        <main className="w-full sm:w-[600px] border-r border-border min-h-screen flex flex-col">
          {/* Header */}
          <header className="sticky top-0 bg-background/80 backdrop-blur-md border-b border-border z-20 flex items-center gap-6 px-4 h-[53px]">
            <button
              onClick={() => router.back()}
              className="p-2 hover:bg-subtle rounded-full text-foreground transition-colors"
            >
              <ArrowLeft className="w-5 h-5" />
            </button>
            <h1 className="font-extrabold text-[20px] leading-6">Post</h1>
          </header>

          {loading ? (
            <div className="flex justify-center items-center p-16 text-brand">
              <Loader2 className="w-8 h-8 animate-spin" />
            </div>
          ) : !post ? (
            <div className="text-center py-20 px-6">
              <h2 className="text-2xl font-bold mb-2">Post not found</h2>
              <p className="text-muted">The post may have been deleted.</p>
            </div>
          ) : (
            <div className="flex flex-col">
              {/* Focal Post Container */}
              <div className="p-4 border-b border-border">
                {/* Author Info */}
                <div className="flex items-center justify-between mb-4">
                  <Link
                    href={`/${post.author?.username || "user"}`}
                    className="flex items-center gap-3 group"
                  >
                    <img
                      src={
                        post.author?.avatar_url ||
                        "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"
                      }
                      alt={post.author?.display_name || "User"}
                      className="w-10 h-10 rounded-full object-cover border border-border"
                    />
                    <div className="flex flex-col">
                      <div className="flex items-center gap-1 font-bold text-[15px] leading-5 group-hover:underline">
                        <span>{post.author?.display_name || "User"}</span>
                        {post.author?.is_verified && (
                          <CheckCircle2 className="w-4 h-4 fill-brand text-background inline" />
                        )}
                      </div>
                      <span className="text-[14px] text-muted leading-4">
                        @{post.author?.username || "user"}
                      </span>
                    </div>
                  </Link>

                  <button className="p-2 hover:bg-subtle rounded-full text-muted hover:text-foreground transition-colors">
                    <MoreHorizontal className="w-5 h-5" />
                  </button>
                </div>

                {/* Focal Content Text */}
                <p className="text-[19px] leading-6 text-foreground mb-4 whitespace-pre-wrap break-words">
                  {post.content}
                </p>

                {/* Media gallery */}
                {post.media && post.media.length > 0 && (
                  <div className="mb-4 rounded-2xl overflow-hidden border border-border">
                    <img
                      src={post.media[0].url}
                      alt="Attachment"
                      className="w-full max-h-[510px] object-cover"
                    />
                  </div>
                )}

                {/* Full Timestamp & Views */}
                <div className="flex items-center gap-1 text-[15px] text-muted py-3 border-y border-border">
                  <span>{formatFullTimestamp(post.created_at)}</span>
                  <span>·</span>
                  <span className="font-bold text-foreground">
                    {formatNumber(post.views_count || 1)}
                  </span>
                  <span>Views</span>
                </div>

                {/* Engagement Counters Row */}
                <div className="flex items-center gap-6 py-3 border-b border-border text-[14px]">
                  {repostsCount > 0 && (
                    <div>
                      <span className="font-bold text-foreground">{formatNumber(repostsCount)}</span>{" "}
                      <span className="text-muted">Reposts</span>
                    </div>
                  )}
                  {likesCount > 0 && (
                    <div>
                      <span className="font-bold text-foreground">{formatNumber(likesCount)}</span>{" "}
                      <span className="text-muted">Likes</span>
                    </div>
                  )}
                  <div>
                    <span className="font-bold text-foreground">{post.replies_count}</span>{" "}
                    <span className="text-muted">Replies</span>
                  </div>
                </div>

                {/* Primary Action Bar */}
                <div className="flex items-center justify-around py-2 border-b border-border text-muted">
                  <button className="p-2 hover:bg-brand/10 hover:text-brand rounded-full transition-colors">
                    <MessageCircle className="w-5 h-5" />
                  </button>
                  <button
                    onClick={handleRepost}
                    className={cn(
                      "p-2 hover:bg-repost/10 hover:text-repost rounded-full transition-colors",
                      isReposted && "text-repost"
                    )}
                  >
                    <Repeat2 className="w-5 h-5" />
                  </button>
                  <button
                    onClick={handleLike}
                    className={cn(
                      "p-2 hover:bg-like/10 hover:text-like rounded-full transition-colors",
                      isLiked && "text-like"
                    )}
                  >
                    <Heart className={cn("w-5 h-5", isLiked && "fill-current")} />
                  </button>
                  <button
                    onClick={handleBookmark}
                    className={cn(
                      "p-2 hover:bg-brand/10 hover:text-brand rounded-full transition-colors",
                      isBookmarked && "text-brand"
                    )}
                  >
                    <Bookmark className={cn("w-5 h-5", isBookmarked && "fill-current")} />
                  </button>
                  <button className="p-2 hover:bg-brand/10 hover:text-brand rounded-full transition-colors">
                    <Share className="w-5 h-5" />
                  </button>
                </div>
              </div>

              {/* Reply Composer */}
              <div className="border-b border-border">
                <PostComposer
                  currentUser={currentUser}
                  placeholder="Post your reply"
                  replyToId={post.id}
                  onPostSuccess={handleReplyCreated}
                  onOpenAuth={() => setAuthModalOpen(true)}
                />
              </div>

              {/* Descendant Replies Thread */}
              <div className="flex flex-col">
                {replies.map((reply) => (
                  <PostCard key={reply.id} post={reply} />
                ))}
              </div>
            </div>
          )}
        </main>

        <RightSidebar />

        <AuthModal
          isOpen={authModalOpen}
          onClose={() => setAuthModalOpen(false)}
          onSuccess={(user) => setCurrentUser(user)}
        />
      </div>
    </div>
  );
}
