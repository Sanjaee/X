"use client";

import React, { useRef, useEffect, useCallback } from "react";
import { useWindowVirtualizer } from "@tanstack/react-virtual";
import { PostCard } from "@/components/feed/PostCard";
import { Post } from "@/types";
import { Loader2, RefreshCw, AlertCircle } from "lucide-react";

interface VirtualizedFeedProps {
  posts: Post[];
  isLoading: boolean;
  isFetchingNextPage: boolean;
  hasNextPage?: boolean;
  fetchNextPage: () => void;
  isError?: boolean;
  error?: Error | null;
  refetch?: () => void;
  onLike?: (postId: string) => void;
  onRepost?: (postId: string) => void;
  onBookmark?: (postId: string) => void;
  onFeedback?: (postId: string, feedbackType: "not_interested" | "hide") => void;
  storageKey?: string;
  emptyState?: React.ReactNode;
}

export function VirtualizedFeed({
  posts,
  isLoading,
  isFetchingNextPage,
  hasNextPage,
  fetchNextPage,
  isError,
  error,
  refetch,
  onLike,
  onRepost,
  onBookmark,
  onFeedback,
  storageKey = "x_feed_scroll",
  emptyState,
}: VirtualizedFeedProps) {
  const listRef = useRef<HTMLDivElement>(null);
  const sentinelRef = useRef<HTMLDivElement>(null);

  // Set up window virtualizer for natural window scrolling
  const virtualizer = useWindowVirtualizer({
    count: posts.length,
    estimateSize: () => 180, // estimated tweet height
    overscan: 5,
    scrollMargin: listRef.current?.offsetTop ?? 0,
  });

  const virtualItems = virtualizer.getVirtualItems();

  // Scroll restoration: store and restore anchor post ID seamlessly
  useEffect(() => {
    if (!storageKey || typeof window === "undefined" || posts.length === 0) return;

    const mediaAnchor = sessionStorage.getItem("x_last_media_anchor");
    const savedAnchor = mediaAnchor || sessionStorage.getItem(`${storageKey}_anchor`);

    if (savedAnchor) {
      const targetIndex = posts.findIndex((p) => p.id === savedAnchor);
      if (targetIndex >= 0) {
        requestAnimationFrame(() => {
          virtualizer.scrollToIndex(targetIndex, { align: "start" });
          if (mediaAnchor) {
            sessionStorage.removeItem("x_last_media_anchor");
          }
        });
      }
    }

    const handleBeforeUnload = () => {
      // Find the top-most visible item
      if (virtualItems.length > 0) {
        const topItem = posts[virtualItems[0].index];
        if (topItem) {
          sessionStorage.setItem(`${storageKey}_anchor`, topItem.id);
        }
      }
    };

    window.addEventListener("scroll", handleBeforeUnload, { passive: true });
    return () => window.removeEventListener("scroll", handleBeforeUnload);
  }, [posts, storageKey, virtualItems, virtualizer]);

  // Bottom Sentinel IntersectionObserver with 800px prefetch margin
  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel) return;

    const observer = new IntersectionObserver(
      (entries) => {
        const entry = entries[0];
        if (entry.isIntersecting && hasNextPage && !isFetchingNextPage && !isLoading) {
          fetchNextPage();
        }
      },
      { rootMargin: "800px" }
    );

    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [hasNextPage, isFetchingNextPage, isLoading, fetchNextPage]);

  // Initial Loading Skeleton
  if (isLoading && posts.length === 0) {
    return (
      <div className="flex flex-col divide-y divide-border">
        {[1, 2, 3, 4, 5].map((i) => (
          <div key={i} className="p-4 flex gap-3 animate-pulse">
            <div className="w-10 h-10 rounded-full bg-muted/40 shrink-0" />
            <div className="flex-1 space-y-3 py-1">
              <div className="h-4 bg-muted/40 rounded w-1/4" />
              <div className="h-4 bg-muted/40 rounded w-5/6" />
              <div className="h-4 bg-muted/40 rounded w-2/3" />
              <div className="h-32 bg-muted/30 rounded-2xl w-full" />
            </div>
          </div>
        ))}
      </div>
    );
  }

  // Error State with Retry
  if (isError && posts.length === 0) {
    return (
      <div className="p-8 text-center flex flex-col items-center justify-center">
        <AlertCircle className="w-10 h-10 text-red-500 mb-3" />
        <h3 className="font-bold text-[18px] mb-1">Something went wrong</h3>
        <p className="text-muted text-[14px] mb-4 max-w-sm">
          {error?.message || "Failed to load posts. Please try again."}
        </p>
        {refetch && (
          <button
            onClick={() => refetch()}
            className="flex items-center gap-2 bg-brand hover:bg-brand-hover text-white font-bold py-2 px-5 rounded-full text-[14px] transition-colors"
          >
            <RefreshCw className="w-4 h-4" />
            <span>Retry</span>
          </button>
        )}
      </div>
    );
  }

  // Empty State
  if (posts.length === 0) {
    return emptyState ? (
      <>{emptyState}</>
    ) : (
      <div className="text-center py-16 px-6">
        <h3 className="font-bold text-[20px] mb-2">No posts found</h3>
        <p className="text-muted text-[15px] max-w-sm mx-auto">
          There are no posts to display here right now.
        </p>
      </div>
    );
  }

  return (
    <div ref={listRef} className="w-full relative">
      {/* Virtualized Container with Calculated Total Height */}
      <div
        style={{
          height: `${virtualizer.getTotalSize()}px`,
          width: "100%",
          position: "relative",
        }}
      >
        {virtualItems.map((virtualRow) => {
          const post = posts[virtualRow.index];
          if (!post) return null;

          return (
            <div
              key={post.id}
              data-index={virtualRow.index}
              data-post-id={post.id}
              ref={virtualizer.measureElement}
              style={{
                position: "absolute",
                top: 0,
                left: 0,
                width: "100%",
                transform: `translateY(${virtualRow.start - (virtualizer.options.scrollMargin ?? 0)}px)`,
              }}
            >
              <PostCard
                post={post}
                onLike={onLike}
                onRepost={onRepost}
                onBookmark={onBookmark}
                onFeedback={onFeedback}
              />
            </div>
          );
        })}
      </div>

      {/* Sentinel Trigger & Loading More Indicator */}
      <div ref={sentinelRef} className="w-full py-4 flex justify-center items-center">
        {isFetchingNextPage && (
          <div className="flex items-center gap-2 text-brand text-[14px] font-medium">
            <Loader2 className="w-5 h-5 animate-spin" />
            <span>Loading more posts...</span>
          </div>
        )}
        {!hasNextPage && posts.length > 0 && !isFetchingNextPage && (
          <p className="text-muted text-[13px] py-4 text-center">
            You're all caught up!
          </p>
        )}
      </div>
    </div>
  );
}
