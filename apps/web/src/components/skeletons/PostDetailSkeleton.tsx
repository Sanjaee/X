"use client";

import React from "react";
import { PostCardSkeleton } from "./FeedSkeleton";

export function PostDetailSkeleton() {
  return (
    <div
      role="status"
      aria-label="Loading post details"
      className="w-full sm:w-[600px] border-r border-border min-h-screen flex flex-col"
    >
      {/* Header */}
      <div className="sticky top-0 bg-background/80 backdrop-blur-md border-b border-border h-[53px] flex items-center px-4 gap-6 z-10">
        <div className="w-5 h-5 bg-muted/20 rounded-full shrink-0" />
        <div className="w-16 h-5 bg-muted/25 rounded-md" />
      </div>

      {/* Main focal post */}
      <div className="p-4 border-b border-border flex flex-col gap-3">
        {/* Author info */}
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-full bg-muted/25 shrink-0" />
          <div className="flex flex-col gap-1 flex-1">
            <div className="w-28 h-4 bg-muted/25 rounded-md" />
            <div className="w-20 h-3 bg-muted/15 rounded-md" />
          </div>
        </div>

        {/* Focal text */}
        <div className="flex flex-col gap-2 pt-1">
          <div className="w-full h-4.5 bg-muted/20 rounded-md" />
          <div className="w-full h-4.5 bg-muted/20 rounded-md" />
          <div className="w-3/5 h-4.5 bg-muted/15 rounded-md" />
        </div>

        {/* Timestamp */}
        <div className="py-2 border-b border-border/60">
          <div className="w-44 h-3 bg-muted/15 rounded-md" />
        </div>

        {/* Metrics row */}
        <div className="flex gap-4 py-2 border-b border-border/60">
          <div className="w-16 h-3.5 bg-muted/20 rounded-md" />
          <div className="w-16 h-3.5 bg-muted/20 rounded-md" />
          <div className="w-16 h-3.5 bg-muted/20 rounded-md" />
        </div>

        {/* Actions row */}
        <div className="flex justify-around py-1">
          <div className="w-5 h-5 bg-muted/15 rounded-md" />
          <div className="w-5 h-5 bg-muted/15 rounded-md" />
          <div className="w-5 h-5 bg-muted/15 rounded-md" />
          <div className="w-5 h-5 bg-muted/15 rounded-md" />
          <div className="w-5 h-5 bg-muted/15 rounded-md" />
        </div>
      </div>

      {/* Reply skeletons */}
      <PostCardSkeleton />
      <PostCardSkeleton />
      <span className="sr-only">Loading post conversation...</span>
    </div>
  );
}
