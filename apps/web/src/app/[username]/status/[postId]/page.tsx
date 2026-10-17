"use client";

import React, { useEffect, useState, useRef } from "react";
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
  ChevronDown,
  ChevronRight,
  Image as ImageIcon,
  Smile,
  BarChart2,
  Calendar,
  MapPin,
  Pencil,
  Check,
  X,
  RotateCcw,
  Settings,
} from "lucide-react";
import { LeftNavRail } from "@/components/layout/LeftNavRail";
import { RightSidebar } from "@/components/layout/RightSidebar";
import { PostModal } from "@/components/composer/PostModal";
import { PostCard } from "@/components/feed/PostCard";
import { AuthModal } from "@/components/auth/AuthModal";
import { User, Post, MediaObject } from "@/types";
import { apiRequest } from "@/lib/api";
import { formatFullTimestamp, formatNumber, cn } from "@/lib/utils";
import { useMediaViewer } from "@/components/media/MediaViewerContext";
import { VideoPlayer } from "@/components/media/VideoPlayer";
import {
  X_ALLOWED_MEDIA_ACCEPT,
  filterValidMediaFiles,
} from "@/lib/mediaValidation";

interface StagedFile {
  id: string;
  file: File;
  previewUrl: string;
  type: "image" | "video";
}

export default function PostDetailPage() {
  const params = useParams();
  const router = useRouter();
  const { openMedia } = useMediaViewer();
  const postId = params.postId as string;

  const [post, setPost] = useState<Post | null>(null);
  const [replies, setReplies] = useState<Post[]>([]);
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [authModalOpen, setAuthModalOpen] = useState(false);
  const [replyModalOpen, setReplyModalOpen] = useState(false);
  const [quoteModalOpen, setQuoteModalOpen] = useState(false);
  const [repostMenuOpen, setRepostMenuOpen] = useState(false);

  // Engagement states
  const [isLiked, setIsLiked] = useState(false);
  const [likesCount, setLikesCount] = useState(0);
  const [isReposted, setIsReposted] = useState(false);
  const [repostsCount, setRepostsCount] = useState(0);
  const [isBookmarked, setIsBookmarked] = useState(false);
  const [bookmarksCount, setBookmarksCount] = useState(0);

  // Reply sorting state
  const [replySort, setReplySort] = useState<"relevant" | "recent" | "likes">("relevant");
  const [sortMenuOpen, setSortMenuOpen] = useState(false);
  const [loadingReplies, setLoadingReplies] = useState(false);

  // Quotes modal state
  const [quotesModalOpen, setQuotesModalOpen] = useState(false);
  const [quotePosts, setQuotePosts] = useState<Post[]>([]);
  const [loadingQuotes, setLoadingQuotes] = useState(false);

  // Inline Reply Composer state
  const [replyContent, setReplyContent] = useState("");
  const [stagedFiles, setStagedFiles] = useState<StagedFile[]>([]);
  const [isSubmittingReply, setIsSubmittingReply] = useState(false);
  const [composerFocused, setComposerFocused] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const composerContainerRef = useRef<HTMLDivElement>(null);

  // Collapse inline reply composer when clicking outside and input is empty
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (
        composerContainerRef.current &&
        !composerContainerRef.current.contains(e.target as Node) &&
        !replyContent.trim() &&
        stagedFiles.length === 0
      ) {
        setComposerFocused(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [replyContent, stagedFiles.length]);

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
      apiRequest<{ posts: Post[] }>(`/posts/${postId}/replies?sort=${replySort}`),
    ])
      .then(([postData, repliesData]) => {
        setPost(postData);
        setIsLiked(postData.is_liked);
        setLikesCount(postData.likes_count || 0);
        setIsReposted(postData.is_reposted);
        setRepostsCount(postData.reposts_count || 0);
        setIsBookmarked(postData.is_bookmarked);
        setBookmarksCount(postData.is_bookmarked ? 1 : 0);
        setReplies(repliesData.posts || []);
      })
      .catch((err) => {
        console.error("Failed to load post detail", err);
      })
      .finally(() => {
        setLoading(false);
      });
  }, [postId]);

  // Re-fetch replies when sort changes
  useEffect(() => {
    if (!postId || loading) return;
    setLoadingReplies(true);
    apiRequest<{ posts: Post[] }>(`/posts/${postId}/replies?sort=${replySort}`)
      .then((res) => {
        setReplies(res.posts || []);
      })
      .catch((e) => {
        console.error("Failed to sort replies", e);
      })
      .finally(() => {
        setLoadingReplies(false);
      });
  }, [replySort, postId]);

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

  const handleToggleRepost = async () => {
    if (!currentUser) {
      setAuthModalOpen(true);
      return;
    }
    const prev = isReposted;
    setIsReposted(!prev);
    setRepostsCount(prev ? Math.max(0, repostsCount - 1) : repostsCount + 1);
    setRepostMenuOpen(false);

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
    setBookmarksCount(prev ? Math.max(0, bookmarksCount - 1) : bookmarksCount + 1);
    try {
      await apiRequest(`/posts/${postId}/bookmarks`, {
        method: prev ? "DELETE" : "POST",
      });
    } catch (e) {
      setIsBookmarked(prev);
      setBookmarksCount(bookmarksCount);
    }
  };

  // Open Quotes modal and fetch quote posts
  const handleOpenQuotes = async () => {
    setQuotesModalOpen(true);
    setLoadingQuotes(true);
    try {
      const res = await apiRequest<{ posts: Post[] }>(`/posts/${postId}/quotes`);
      setQuotePosts(res.posts || []);
    } catch (err) {
      console.error("Failed to load quotes", err);
      setQuotePosts([]);
    } finally {
      setLoadingQuotes(false);
    }
  };

  // Inline reply file staging with strict X allowed file types validation
  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const rawFiles = e.target.files;
    if (!rawFiles || rawFiles.length === 0) return;

    const { validFiles, rejectedCount } = filterValidMediaFiles(rawFiles);
    if (rejectedCount > 0) {
      alert(
        `Selected unsupported file format(s). Only valid image and video files are accepted.`
      );
    }
    if (validFiles.length === 0) return;

    const totalAllowed = 4 - stagedFiles.length;
    const filesToStage = validFiles.slice(0, totalAllowed);

    const newStaged: StagedFile[] = filesToStage.map((f) => ({
      id: Math.random().toString(36).substring(2),
      file: f,
      previewUrl: URL.createObjectURL(f),
      type: f.type.startsWith("video/") || f.name.endsWith(".m4v") || f.name.endsWith(".mp4") || f.name.endsWith(".mov")
        ? "video"
        : "image",
    }));

    setStagedFiles((prev) => [...prev, ...newStaged]);
    if (fileInputRef.current) fileInputRef.current.value = "";
  };

  const handleRemoveStaged = (idx: number) => {
    setStagedFiles((prev) => {
      const copy = [...prev];
      const removed = copy.splice(idx, 1);
      if (removed[0]) URL.revokeObjectURL(removed[0].previewUrl);
      return copy;
    });
  };

  // Submit inline reply
  const handleSubmitInlineReply = async () => {
    if (!currentUser) {
      setAuthModalOpen(true);
      return;
    }
    if (!replyContent.trim() && stagedFiles.length === 0) return;

    setIsSubmittingReply(true);
    try {
      const uploadedMedia: MediaObject[] = [];
      for (const staged of stagedFiles) {
        const formData = new FormData();
        formData.append("file", staged.file);
        const mediaRes = await apiRequest<MediaObject>("/media/upload", {
          method: "POST",
          body: formData,
        });
        if (mediaRes && mediaRes.id) {
          uploadedMedia.push(mediaRes);
        }
      }

      const newReply = await apiRequest<Post>("/posts", {
        method: "POST",
        body: JSON.stringify({
          content: replyContent.trim(),
          reply_to_id: postId,
          media_ids: uploadedMedia.map((m) => m.id),
        }),
      });

      const completeReply: Post = {
        ...newReply,
        author: newReply.author || { ...currentUser },
        media: uploadedMedia,
      };

      setReplies((prev) => [completeReply, ...prev]);
      setPost((prev) =>
        prev ? { ...prev, replies_count: (prev.replies_count || 0) + 1 } : null
      );

      setReplyContent("");
      stagedFiles.forEach((f) => URL.revokeObjectURL(f.previewUrl));
      setStagedFiles([]);
      setComposerFocused(false);
    } catch (err: any) {
      alert(err.message || "Failed to post reply. Please try again.");
    } finally {
      setIsSubmittingReply(false);
    }
  };

  const handleModalReplyCreated = (newReply: Post) => {
    const completeReply: Post = {
      ...newReply,
      author: newReply.author || (currentUser ? { ...currentUser } : undefined),
    };
    setReplies((prev) => [completeReply, ...prev]);
    setPost((prev) =>
      prev ? { ...prev, replies_count: (prev.replies_count || 0) + 1 } : null
    );
  };

  const sortLabels = {
    relevant: "Most relevant",
    recent: "Recent",
    likes: "Likes",
  };

  const isComposerActive = composerFocused || replyContent.trim().length > 0 || stagedFiles.length > 0;

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
                <div className="flex items-center justify-between mb-3">
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
                          <CheckCircle2 className="w-4 h-4 fill-amber-400 text-black inline" />
                        )}
                        <span className="text-[12px] px-1 py-0.2 bg-muted/30 rounded font-black">𝕏</span>
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

                {/* Translation hint banner */}
                <div className="flex items-center gap-1.5 text-xs text-muted mb-3 flex-wrap">
                  <RotateCcw className="w-3.5 h-3.5" />
                  <span>Translate post</span>
                  <button className="text-brand hover:underline font-medium">
                    Show original
                  </button>
                  <Settings className="w-3.5 h-3.5 text-muted hover:text-foreground cursor-pointer" />
                </div>

                {/* Focal Content Text */}
                <p className="text-[19px] leading-6 text-foreground mb-3 whitespace-pre-wrap break-words">
                  {post.content}
                </p>

                {/* Media gallery with Photo & Video Viewer */}
                {post.media && post.media.length > 0 && (
                  <div className="mb-4 rounded-2xl overflow-hidden border border-border bg-muted/20">
                    {post.media.length === 1 && (
                      post.media[0].media_type === "video" ? (
                        <div className="w-full aspect-video bg-black flex items-center justify-center">
                          <VideoPlayer
                            src={post.media[0].url}
                            mode="in_feed"
                            onOpenViewer={() => openMedia(post, 0)}
                          />
                        </div>
                      ) : (
                        <div
                          onClick={() => openMedia(post, 0)}
                          className="w-full bg-muted/30 relative overflow-hidden cursor-pointer group"
                        >
                          <img
                            src={post.media[0].url}
                            alt="Attachment"
                            className="w-full max-h-[510px] object-cover group-hover:opacity-95 transition-opacity"
                          />
                        </div>
                      )
                    )}
                    {post.media.length > 1 && (
                      <div className="grid grid-cols-2 gap-0.5 aspect-video bg-muted/20">
                        {post.media.slice(0, 4).map((m, idx) => (
                          <div
                            key={m.id || idx}
                            onClick={() => openMedia(post, idx)}
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
                                className="w-full h-full object-cover group-hover:opacity-95 transition-opacity"
                              />
                            )}
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                )}

                {/* Quoted Post Card Preview if present */}
                {post.quote_post && (
                  <div
                    onClick={(e) => {
                      e.stopPropagation();
                      router.push(
                        `/${post.quote_post?.author?.username || "user"}/status/${post.quote_post?.id}`
                      );
                    }}
                    className="mb-4 p-3 rounded-2xl border border-border bg-card hover:bg-subtle/30 transition-colors cursor-pointer flex flex-col gap-2"
                  >
                    <div className="flex items-center gap-1.5 text-[14px]">
                      <img
                        src={
                          post.quote_post.author?.avatar_url ||
                          "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"
                        }
                        alt="Author"
                        className="w-5 h-5 rounded-full object-cover"
                      />
                      <span className="font-bold text-foreground">
                        {post.quote_post.author?.display_name || "User"}
                      </span>
                      <span className="text-muted">
                        @{post.quote_post.author?.username || "user"}
                      </span>
                    </div>
                    {post.quote_post.content && (
                      <p className="text-[14px] leading-5 text-foreground whitespace-pre-wrap break-words">
                        {post.quote_post.content}
                      </p>
                    )}
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

                {/* Primary Action Bar with Icon & Counter Side-by-Side */}
                <div className="flex items-center justify-between py-3 border-b border-border text-muted text-[13px] font-medium px-2">
                  {/* Reply */}
                  <button
                    onClick={() => {
                      if (!currentUser) setAuthModalOpen(true);
                      else setReplyModalOpen(true);
                    }}
                    className="flex items-center gap-2 hover:text-brand transition-colors group"
                  >
                    <div className="p-1.5 rounded-full group-hover:bg-brand/10 transition-colors">
                      <MessageCircle className="w-5 h-5" />
                    </div>
                    <span>{post.replies_count > 0 ? post.replies_count : ""}</span>
                  </button>

                  {/* Repost Popover */}
                  <div className="relative">
                    <button
                      onClick={() => {
                        if (!currentUser) {
                          setAuthModalOpen(true);
                          return;
                        }
                        setRepostMenuOpen(!repostMenuOpen);
                      }}
                      className={cn(
                        "flex items-center gap-2 transition-colors group",
                        isReposted ? "text-repost" : "hover:text-repost"
                      )}
                    >
                      <div className="p-1.5 rounded-full group-hover:bg-repost/10 transition-colors">
                        <Repeat2 className="w-5 h-5" />
                      </div>
                      <span>{repostsCount > 0 ? repostsCount : ""}</span>
                    </button>

                    {repostMenuOpen && (
                      <>
                        <div
                          className="fixed inset-0 z-30"
                          onClick={() => setRepostMenuOpen(false)}
                        />
                        <div className="absolute left-0 top-9 w-44 bg-background border border-border shadow-2xl rounded-2xl py-1.5 z-40 text-[14px] font-bold">
                          <button
                            onClick={handleToggleRepost}
                            className="w-full text-left px-4 py-3 hover:bg-subtle flex items-center gap-3 text-foreground transition-colors"
                          >
                            <Repeat2 className="w-4 h-4" />
                            <span>{isReposted ? "Undo repost" : "Repost"}</span>
                          </button>
                          <button
                            onClick={() => {
                              setRepostMenuOpen(false);
                              setQuoteModalOpen(true);
                            }}
                            className="w-full text-left px-4 py-3 hover:bg-subtle flex items-center gap-3 text-foreground transition-colors"
                          >
                            <Pencil className="w-4 h-4" />
                            <span>Quote</span>
                          </button>
                        </div>
                      </>
                    )}
                  </div>

                  {/* Like */}
                  <button
                    onClick={handleLike}
                    className={cn(
                      "flex items-center gap-2 transition-colors group",
                      isLiked ? "text-like" : "hover:text-like"
                    )}
                  >
                    <div className="p-1.5 rounded-full group-hover:bg-like/10 transition-colors">
                      <Heart
                        className={cn("w-5 h-5 transition-transform active:scale-125", isLiked && "fill-current text-like")}
                      />
                    </div>
                    <span>{likesCount > 0 ? likesCount : ""}</span>
                  </button>

                  {/* Bookmark */}
                  <button
                    onClick={handleBookmark}
                    className={cn(
                      "flex items-center gap-2 transition-colors group",
                      isBookmarked ? "text-brand" : "hover:text-brand"
                    )}
                  >
                    <div className="p-1.5 rounded-full group-hover:bg-brand/10 transition-colors">
                      <Bookmark
                        className={cn("w-5 h-5", isBookmarked && "fill-current text-brand")}
                      />
                    </div>
                    <span>{bookmarksCount > 0 ? bookmarksCount : ""}</span>
                  </button>

                  {/* Share */}
                  <button
                    onClick={() => {
                      if (navigator.share) {
                        navigator.share({
                          title: `Post by ${post.author?.display_name || "user"}`,
                          url: window.location.href,
                        }).catch(() => {});
                      } else {
                        navigator.clipboard.writeText(window.location.href);
                        alert("Link copied to clipboard!");
                      }
                    }}
                    className="p-1.5 rounded-full hover:bg-brand/10 hover:text-brand transition-colors"
                  >
                    <Share className="w-5 h-5" />
                  </button>
                </div>

                {/* Sub-bar: Sort dropdown (Left) and Quotes link (Right) */}
                <div className="flex items-center justify-between pt-3 text-[14px]">
                  {/* Left: Sort dropdown */}
                  <div className="relative">
                    <button
                      onClick={() => setSortMenuOpen(!sortMenuOpen)}
                      className="flex items-center gap-1 text-muted hover:text-foreground font-medium transition-colors py-1 px-1.5 rounded hover:bg-subtle"
                    >
                      <span>{sortLabels[replySort]}</span>
                      <ChevronDown className="w-4 h-4 text-muted" />
                    </button>

                    {sortMenuOpen && (
                      <>
                        <div
                          className="fixed inset-0 z-30"
                          onClick={() => setSortMenuOpen(false)}
                        />
                        <div className="absolute left-0 top-8 w-48 bg-card border border-border shadow-2xl rounded-2xl py-2 z-40">
                          <div className="px-4 py-1.5 text-xs font-bold text-muted border-b border-border/50">
                            Sort replies
                          </div>
                          {(["relevant", "recent", "likes"] as const).map((mode) => (
                            <button
                              key={mode}
                              onClick={() => {
                                setReplySort(mode);
                                setSortMenuOpen(false);
                              }}
                              className="w-full px-4 py-2.5 text-left text-sm font-semibold flex items-center justify-between hover:bg-subtle text-foreground transition-colors"
                            >
                              <span>{sortLabels[mode]}</span>
                              {replySort === mode && (
                                <Check className="w-4 h-4 text-brand" />
                              )}
                            </button>
                          ))}
                        </div>
                      </>
                    )}
                  </div>

                  {/* Right: View quotes > */}
                  <button
                    onClick={handleOpenQuotes}
                    className="flex items-center gap-1 text-muted hover:text-foreground font-medium transition-colors py-1 px-1.5 rounded hover:bg-subtle"
                  >
                    <span>View quotes</span>
                    {(post.quotes_count ?? 0) > 0 && (
                      <span className="text-muted">({post.quotes_count})</span>
                    )}
                    <ChevronRight className="w-4 h-4 text-muted" />
                  </button>
                </div>
              </div>

              {/* Inline Reply Composer (Only when logged in: Inactive: single compact row; Active: expands with toolbar) */}
              {currentUser && (
                <div ref={composerContainerRef} className="p-4 border-b border-border">
                  {!isComposerActive ? (
                    /* Inactive State: Avatar + Placeholder + Reply button on right (No action icons) */
                    <div className="flex items-center gap-3">
                      <img
                        src={
                          currentUser?.avatar_url ||
                          "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"
                        }
                        alt={currentUser?.display_name || "User"}
                        className="w-10 h-10 rounded-full object-cover shrink-0 border border-border"
                      />
                      <div
                        onClick={() => setComposerFocused(true)}
                        className="flex-1 cursor-text py-2 text-[16px] text-muted font-normal select-none"
                      >
                        Post your reply
                      </div>
                      <button
                        disabled
                        className="bg-brand/50 text-white/60 px-4 py-1.5 rounded-full font-bold text-[14px] cursor-not-allowed"
                      >
                        Reply
                      </button>
                    </div>
                  ) : (
                    /* Active State: Replying to @author + Expanded Textarea + Action Icons + Reply button */
                    <div>
                      {/* Replying context banner */}
                      <div className="text-xs text-muted mb-2 ml-13">
                        <span>Replying to </span>
                        <span className="text-brand font-medium">@{post.author?.username || "user"}</span>
                      </div>

                      <div className="flex gap-3 items-start">
                        <img
                          src={
                            currentUser?.avatar_url ||
                            "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"
                          }
                          alt={currentUser?.display_name || "User"}
                          className="w-10 h-10 rounded-full object-cover shrink-0 border border-border mt-1"
                        />

                        <div className="flex-1 min-w-0 flex flex-col">
                          <textarea
                            autoFocus
                            value={replyContent}
                            onChange={(e) => setReplyContent(e.target.value)}
                            placeholder="Post your reply"
                            rows={3}
                            className="w-full bg-transparent text-[16px] text-foreground placeholder:text-muted outline-none resize-none leading-normal font-normal"
                          />

                          {/* Staged file previews */}
                          {stagedFiles.length > 0 && (
                            <div className="grid grid-cols-2 gap-2 my-2">
                              {stagedFiles.map((staged, idx) => (
                                <div
                                  key={staged.id}
                                  className="relative rounded-xl overflow-hidden border border-border bg-black max-h-[160px] group"
                                >
                                  {staged.type === "video" ? (
                                    <video
                                      src={staged.previewUrl}
                                      controls
                                      className="w-full h-full object-cover max-h-[160px]"
                                    />
                                  ) : (
                                    <img
                                      src={staged.previewUrl}
                                      alt="attachment"
                                      className="w-full h-full object-cover max-h-[160px]"
                                    />
                                  )}
                                  <button
                                    onClick={() => handleRemoveStaged(idx)}
                                    className="absolute top-1 right-1 p-1 rounded-full bg-black/70 hover:bg-black/90 text-white transition-colors"
                                  >
                                    <X className="w-3.5 h-3.5" />
                                  </button>
                                </div>
                              ))}
                            </div>
                          )}

                          {/* Action Icons Toolbar & Reply Button Row */}
                          <div className="flex items-center justify-between pt-2 mt-1">
                            <div className="flex items-center gap-0.5 text-brand">
                              {/* Validated file input */}
                              <input
                                ref={fileInputRef}
                                type="file"
                                accept={X_ALLOWED_MEDIA_ACCEPT}
                                multiple
                                onChange={handleFileSelect}
                                className="hidden"
                                id="inline-reply-media"
                              />
                              <label
                                htmlFor="inline-reply-media"
                                className="p-2 hover:bg-brand/10 rounded-full cursor-pointer transition-colors"
                                title="Media"
                              >
                                <ImageIcon className="w-4 h-4" />
                              </label>

                              <button
                                type="button"
                                className="p-2 hover:bg-brand/10 rounded-full transition-colors opacity-70 hover:opacity-100"
                                title="GIF"
                              >
                                <span className="text-[11px] font-black border border-brand rounded px-1 py-0.5 leading-none">
                                  GIF
                                </span>
                              </button>
                              <button
                                type="button"
                                className="p-2 hover:bg-brand/10 rounded-full transition-colors opacity-70 hover:opacity-100"
                                title="Poll"
                              >
                                <BarChart2 className="w-4 h-4" />
                              </button>
                              <button
                                type="button"
                                className="p-2 hover:bg-brand/10 rounded-full transition-colors opacity-70 hover:opacity-100"
                                title="Emoji"
                              >
                                <Smile className="w-4 h-4" />
                              </button>
                              <button
                                type="button"
                                className="p-2 hover:bg-brand/10 rounded-full transition-colors opacity-70 hover:opacity-100"
                                title="Schedule"
                              >
                                <Calendar className="w-4 h-4" />
                              </button>
                              <button
                                type="button"
                                className="p-2 hover:bg-brand/10 rounded-full transition-colors opacity-70 hover:opacity-100"
                                title="Location"
                              >
                                <MapPin className="w-4 h-4" />
                              </button>
                            </div>

                            <button
                              onClick={handleSubmitInlineReply}
                              disabled={isSubmittingReply || (!replyContent.trim() && stagedFiles.length === 0)}
                              className={cn(
                                "px-4 py-1.5 rounded-full font-bold text-[14px] transition-colors flex items-center gap-1.5",
                                replyContent.trim() || stagedFiles.length > 0
                                  ? "bg-brand text-white hover:bg-brand/90"
                                  : "bg-brand/50 text-white/60 cursor-not-allowed"
                              )}
                            >
                              {isSubmittingReply && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
                              <span>Reply</span>
                            </button>
                          </div>
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              )}

              {/* Descendant Replies Thread */}
              <div className="flex flex-col">
                {loadingReplies ? (
                  <div className="flex justify-center p-8 text-brand">
                    <Loader2 className="w-6 h-6 animate-spin" />
                  </div>
                ) : replies.length === 0 ? (
                  currentUser ? (
                    <div className="p-8 text-center text-muted text-sm">
                      No replies yet. Be the first to reply!
                    </div>
                  ) : (
                    <div className="py-8 px-4 flex flex-col items-center justify-center text-center border-t border-border">
                      <p className="text-muted text-sm mb-3">No replies yet.</p>
                      <h3 className="text-[20px] font-extrabold text-foreground mb-4">
                        See all the replies
                      </h3>
                      <button
                        onClick={() => setAuthModalOpen(true)}
                        className="w-full max-w-[560px] bg-white hover:bg-neutral-200 text-black font-bold py-3 px-6 rounded-full text-[15px] transition-all flex items-center justify-center gap-2 shadow-md active:scale-[0.99]"
                      >
                        <span>Continue to 𝕏</span>
                      </button>
                    </div>
                  )
                ) : (
                  <>
                    {(currentUser ? replies : replies.slice(0, 2)).map((reply) => (
                      <PostCard
                        key={reply.id}
                        post={reply}
                        currentUser={currentUser}
                        onOpenAuth={() => setAuthModalOpen(true)}
                      />
                    ))}

                    {/* Unauthenticated barrier callout matching screenshot */}
                    {!currentUser && (
                      <div className="py-8 px-4 flex flex-col items-center justify-center text-center border-t border-border bg-gradient-to-t from-background to-transparent">
                        <h3 className="text-[20px] font-extrabold text-foreground mb-4">
                          See all the replies
                        </h3>
                        <button
                          onClick={() => setAuthModalOpen(true)}
                          className="w-full max-w-[560px] bg-white hover:bg-neutral-200 text-black font-bold py-3 px-6 rounded-full text-[15px] transition-all flex items-center justify-center gap-2 shadow-md active:scale-[0.99]"
                        >
                          <span>Continue to 𝕏</span>
                        </button>
                      </div>
                    )}
                  </>
                )}
              </div>
            </div>
          )}
        </main>

        <RightSidebar
          currentUser={currentUser}
          relevantUser={post?.author}
          onOpenAuth={() => setAuthModalOpen(true)}
        />

        <AuthModal
          isOpen={authModalOpen}
          onClose={() => setAuthModalOpen(false)}
          onSuccess={(user) => setCurrentUser(user)}
        />

        {/* Reply Modal Popup */}
        {replyModalOpen && post && (
          <PostModal
            isOpen={replyModalOpen}
            onClose={() => setReplyModalOpen(false)}
            currentUser={currentUser}
            replyToPost={post}
            onOpenAuth={() => setAuthModalOpen(true)}
            onPostSuccess={(newReply) => {
              handleModalReplyCreated(newReply);
              setReplyModalOpen(false);
            }}
          />
        )}

        {/* Quote Modal Popup */}
        {quoteModalOpen && post && (
          <PostModal
            isOpen={quoteModalOpen}
            onClose={() => setQuoteModalOpen(false)}
            currentUser={currentUser}
            quotedPost={post}
            onOpenAuth={() => setAuthModalOpen(true)}
            onPostSuccess={() => {
              setQuoteModalOpen(false);
              setPost((prev) =>
                prev ? { ...prev, quotes_count: (prev.quotes_count || 0) + 1 } : null
              );
            }}
          />
        )}

        {/* Quotes Viewer Modal */}
        {quotesModalOpen && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
            <div className="bg-background border border-border rounded-2xl w-full max-w-lg max-h-[85vh] flex flex-col shadow-2xl overflow-hidden">
              <div className="flex items-center justify-between px-4 py-3 border-b border-border">
                <h3 className="font-bold text-[18px]">Quotes</h3>
                <button
                  onClick={() => setQuotesModalOpen(false)}
                  className="p-1.5 hover:bg-subtle rounded-full text-muted hover:text-foreground transition-colors"
                >
                  <X className="w-5 h-5" />
                </button>
              </div>

              <div className="flex-1 overflow-y-auto">
                {loadingQuotes ? (
                  <div className="flex justify-center p-12 text-brand">
                    <Loader2 className="w-6 h-6 animate-spin" />
                  </div>
                ) : quotePosts.length === 0 ? (
                  <div className="p-12 text-center text-muted">
                    No quote posts yet for this post.
                  </div>
                ) : (
                  quotePosts.map((qp) => <PostCard key={qp.id} post={qp} />)
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
