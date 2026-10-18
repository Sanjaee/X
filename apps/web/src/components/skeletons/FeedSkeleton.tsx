"use client";

import React from "react";

export function PostCardSkeleton() {
  return (
    <div
      role="status"
      aria-label="Loading post"
      className="border-b border-border p-4 flex gap-3 animate-pulse"
    >
      {/* Avatar skeleton */}
      <div className="w-10 h-10 rounded-full bg-muted/20 shrink-0" />

      {/* Content skeleton */}
      <div className="flex-1 min-w-0 flex flex-col gap-2 pt-0.5">
        <div className="flex items-center gap-2">
          <div className="w-28 h-3.5 bg-muted/25 rounded-md" />
          <div className="w-16 h-3 bg-muted/20 rounded-md" />
          <div className="w-8 h-3 bg-muted/15 rounded-md" />
        </div>
        <div className="w-full h-3.5 bg-muted/20 rounded-md" />
        <div className="w-4/5 h-3.5 bg-muted/20 rounded-md" />
        <div className="w-2/5 h-3.5 bg-muted/15 rounded-md" />

        {/* Action bar skeleton */}
        <div className="flex justify-between max-w-md pt-2">
          <div className="w-5 h-4 bg-muted/15 rounded-md" />
          <div className="w-5 h-4 bg-muted/15 rounded-md" />
          <div className="w-5 h-4 bg-muted/15 rounded-md" />
          <div className="w-5 h-4 bg-muted/15 rounded-md" />
          <div className="w-5 h-4 bg-muted/15 rounded-md" />
        </div>
      </div>
    </div>
  );
}

export function FeedSkeleton({ count = 5 }: { count?: number }) {
  return (
    <div
      role="status"
      aria-label="Loading feed"
      className="w-full sm:w-[600px] border-r border-border min-h-screen flex flex-col"
    >
      {/* Header bar skeleton */}
      <div className="sticky top-0 bg-background/80 backdrop-blur-md border-b border-border h-[53px] flex items-center px-4 justify-between z-10">
        <div className="w-24 h-5 bg-muted/20 rounded-md" />
      </div>

      {/* Skeletons list */}
      {Array.from({ length: count }).map((_, i) => (
        <PostCardSkeleton key={i} />
      ))}
      <span className="sr-only">Loading feed...</span>
    </div>
  );
}
