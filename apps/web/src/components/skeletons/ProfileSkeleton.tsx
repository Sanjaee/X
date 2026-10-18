"use client";

import React from "react";
import { PostCardSkeleton } from "./FeedSkeleton";

export function ProfileSkeleton() {
  return (
    <div
      role="status"
      aria-label="Loading profile"
      className="w-full sm:w-[600px] border-r border-border min-h-screen flex flex-col"
    >
      {/* Header bar */}
      <div className="sticky top-0 bg-background/80 backdrop-blur-md border-b border-border h-[53px] flex items-center px-4 gap-6 z-10">
        <div className="w-5 h-5 bg-muted/20 rounded-full shrink-0" />
        <div className="flex flex-col gap-1.5">
          <div className="w-28 h-4 bg-muted/25 rounded-md" />
          <div className="w-16 h-2.5 bg-muted/15 rounded-md" />
        </div>
      </div>

      {/* Banner */}
      <div className="w-full h-[200px] bg-muted/15 relative" />

      {/* Profile info section */}
      <div className="px-4 pb-4 border-b border-border">
        {/* Avatar and Edit button row */}
        <div className="flex justify-between items-end relative -mt-[67px] mb-3">
          <div className="w-[134px] h-[134px] rounded-full border-4 border-background bg-muted/25" />
          <div className="w-24 h-9 rounded-full bg-muted/20 mb-2" />
        </div>

        {/* Name and handle */}
        <div className="flex flex-col gap-1.5 mb-3">
          <div className="w-36 h-5 bg-muted/25 rounded-md" />
          <div className="w-24 h-3.5 bg-muted/15 rounded-md" />
        </div>

        {/* Bio */}
        <div className="flex flex-col gap-2 mb-3">
          <div className="w-full h-3.5 bg-muted/15 rounded-md" />
          <div className="w-3/4 h-3.5 bg-muted/15 rounded-md" />
        </div>

        {/* Joined date & location */}
        <div className="flex items-center gap-4 mb-3">
          <div className="w-24 h-3 bg-muted/15 rounded-md" />
          <div className="w-28 h-3 bg-muted/15 rounded-md" />
        </div>

        {/* Following & Followers */}
        <div className="flex items-center gap-4">
          <div className="w-20 h-3.5 bg-muted/20 rounded-md" />
          <div className="w-20 h-3.5 bg-muted/20 rounded-md" />
        </div>
      </div>

      {/* Profile Tabs */}
      <div className="flex border-b border-border h-[53px]">
        {["Posts", "Replies", "Reposts", "Media"].map((t, idx) => (
          <div key={t} className="flex-1 flex justify-center items-center">
            <div className={`h-4 bg-muted/20 rounded-md ${idx === 0 ? "w-12 bg-muted/30" : "w-10"}`} />
          </div>
        ))}
      </div>

      {/* Feed placeholder items */}
      <PostCardSkeleton />
      <PostCardSkeleton />
      <PostCardSkeleton />
      <span className="sr-only">Loading profile...</span>
    </div>
  );
}
