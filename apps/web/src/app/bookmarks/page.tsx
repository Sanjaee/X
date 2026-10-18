"use client";

import React, { useState } from "react";
import { Bookmark, MoreHorizontal } from "lucide-react";
import { RightSidebar } from "@/components/layout/RightSidebar";
import { PostCard } from "@/components/feed/PostCard";
import { Post } from "@/types";
import { useCurrentUser } from "@/lib/useCurrentUser";

export default function BookmarksPage() {
  const { currentUser } = useCurrentUser();
  const [posts] = useState<Post[]>([]);

  return (
    <>
      <main className="w-full sm:w-[600px] border-r border-border min-h-screen flex flex-col">
        <header className="sticky top-0 bg-background/80 backdrop-blur-md border-b border-border z-20 flex items-center justify-between px-4 h-[53px]">
          <div>
            <h1 className="font-extrabold text-[20px] leading-6">Bookmarks</h1>
            <span className="text-[13px] text-muted leading-4">
              @{currentUser?.username || "user"}
            </span>
          </div>
          <button className="p-2 hover:bg-subtle rounded-full text-foreground transition-colors">
            <MoreHorizontal className="w-5 h-5" />
          </button>
        </header>

        {posts.length === 0 ? (
          <div className="text-center py-20 px-8">
            <div className="w-16 h-16 rounded-full bg-subtle flex items-center justify-center mx-auto mb-4 text-brand">
              <Bookmark className="w-8 h-8" />
            </div>
            <h2 className="text-[28px] font-extrabold mb-2">Save posts for later</h2>
            <p className="text-muted text-[15px] max-w-sm mx-auto">
              Don’t let the good ones fly away! Bookmark posts to easily find them again in the future.
            </p>
          </div>
        ) : (
          posts.map((post) => <PostCard key={post.id} post={post} />)
        )}
      </main>

      <RightSidebar currentUser={currentUser} />
    </>
  );
}
