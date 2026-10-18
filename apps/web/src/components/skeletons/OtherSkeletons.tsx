"use client";

import React from "react";

export function NotificationsSkeleton() {
  return (
    <div
      role="status"
      aria-label="Loading notifications"
      className="w-full sm:w-[600px] border-r border-border min-h-screen flex flex-col"
    >
      {/* Header */}
      <div className="sticky top-0 bg-background/80 backdrop-blur-md border-b border-border h-[53px] flex items-center justify-between px-4 z-10">
        <div className="w-28 h-5 bg-muted/25 rounded-md" />
        <div className="w-5 h-5 bg-muted/20 rounded-full" />
      </div>

      {/* Tabs */}
      <div className="flex border-b border-border h-[53px]">
        {["All", "Verified", "Mentions"].map((t, i) => (
          <div key={t} className="flex-1 flex justify-center items-center">
            <div className={`h-4 bg-muted/20 rounded-md ${i === 0 ? "w-10 bg-muted/30" : "w-14"}`} />
          </div>
        ))}
      </div>

      {/* Notification items */}
      {Array.from({ length: 6 }).map((_, idx) => (
        <div key={idx} className="border-b border-border p-4 flex gap-3 animate-pulse">
          <div className="w-7 h-7 rounded-full bg-brand/15 shrink-0" />
          <div className="flex-1 flex flex-col gap-2">
            <div className="w-8 h-8 rounded-full bg-muted/25" />
            <div className="w-2/3 h-4 bg-muted/20 rounded-md" />
            <div className="w-1/2 h-3.5 bg-muted/15 rounded-md" />
          </div>
        </div>
      ))}
      <span className="sr-only">Loading notifications...</span>
    </div>
  );
}

export function ExploreSkeleton() {
  return (
    <div
      role="status"
      aria-label="Loading explore"
      className="w-full sm:w-[600px] border-r border-border min-h-screen flex flex-col"
    >
      {/* Search Header */}
      <div className="sticky top-0 bg-background/80 backdrop-blur-md border-b border-border h-[53px] flex items-center px-4 gap-3 z-10">
        <div className="flex-1 h-9 rounded-full bg-muted/15" />
        <div className="w-5 h-5 bg-muted/20 rounded-full" />
      </div>

      {/* Trending items */}
      {Array.from({ length: 5 }).map((_, idx) => (
        <div key={idx} className="p-4 border-b border-border flex justify-between animate-pulse">
          <div className="flex flex-col gap-1.5 flex-1">
            <div className="w-24 h-3 bg-muted/15 rounded-md" />
            <div className="w-36 h-4 bg-muted/25 rounded-md" />
            <div className="w-20 h-3 bg-muted/15 rounded-md" />
          </div>
          <div className="w-4 h-4 bg-muted/15 rounded-full" />
        </div>
      ))}
      <span className="sr-only">Loading explore content...</span>
    </div>
  );
}

export function MessagesSkeleton() {
  return (
    <div
      role="status"
      aria-label="Loading messages"
      className="flex-1 flex border-r border-border min-h-screen"
    >
      {/* Conversations column */}
      <div className="w-full sm:w-[380px] border-r border-border flex flex-col">
        <div className="h-[53px] border-b border-border flex items-center justify-between px-4">
          <div className="w-24 h-5 bg-muted/25 rounded-md" />
          <div className="w-6 h-6 bg-muted/20 rounded-full" />
        </div>
        <div className="p-3 border-b border-border">
          <div className="h-9 rounded-full bg-muted/15" />
        </div>
        {Array.from({ length: 6 }).map((_, i) => (
          <div key={i} className="p-3 border-b border-border/50 flex gap-3 animate-pulse">
            <div className="w-10 h-10 rounded-full bg-muted/25 shrink-0" />
            <div className="flex-1 flex flex-col gap-1.5">
              <div className="w-28 h-4 bg-muted/25 rounded-md" />
              <div className="w-3/4 h-3.5 bg-muted/15 rounded-md" />
            </div>
          </div>
        ))}
      </div>

      {/* Empty chat pane on desktop */}
      <div className="hidden sm:flex flex-1 items-center justify-center">
        <div className="flex flex-col items-center gap-2">
          <div className="w-40 h-5 bg-muted/20 rounded-md" />
          <div className="w-56 h-3.5 bg-muted/15 rounded-md" />
        </div>
      </div>
      <span className="sr-only">Loading messages...</span>
    </div>
  );
}
