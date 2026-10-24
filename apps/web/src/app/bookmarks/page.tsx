"use client";

import React, { useEffect, useState, useCallback } from "react";
import { Bookmark, MoreHorizontal, Loader2 } from "lucide-react";
import { LeftNavRail } from "@/components/layout/LeftNavRail";
import { RightSidebar } from "@/components/layout/RightSidebar";
import { PostCard } from "@/components/feed/PostCard";
import { Post } from "@/types";
import { apiRequest } from "@/lib/api";
import { useCurrentUser } from "@/lib/useCurrentUser";
import { MobileBottomNav } from "@/components/layout/MobileBottomNav";
import { MobileDrawer } from "@/components/layout/MobileDrawer";
import { MobileComposerModal } from "@/components/composer/MobileComposerModal";
import { AuthModal } from "@/components/auth/AuthModal";

export default function BookmarksPage() {
  const { currentUser, updateUser: setCurrentUser, logout: handleLogout } = useCurrentUser();
  const [posts, setPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [nextCursor, setNextCursor] = useState<string | null>(null);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [composerModalOpen, setComposerModalOpen] = useState(false);
  const [authModalOpen, setAuthModalOpen] = useState(false);

  const fetchBookmarks = useCallback(async () => {
    if (!currentUser) {
      setLoading(false);
      return;
    }
    setLoading(true);
    try {
      const res = await apiRequest<{ posts: Post[]; next_cursor?: string }>("/bookmarks?limit=20");
      setPosts(res.posts || []);
      setNextCursor(res.next_cursor || null);
    } catch (err) {
      console.error("Failed to load bookmarks", err);
    } finally {
      setLoading(false);
    }
  }, [currentUser]);

  useEffect(() => {
    fetchBookmarks();
  }, [fetchBookmarks]);

  const handleLoadMore = async () => {
    if (!nextCursor || loadingMore) return;
    setLoadingMore(true);
    try {
      const res = await apiRequest<{ posts: Post[]; next_cursor?: string }>(
        `/bookmarks?cursor=${encodeURIComponent(nextCursor)}&limit=20`
      );
      if (res.posts && res.posts.length > 0) {
        setPosts((prev) => {
          const ids = new Set(prev.map((p) => p.id));
          const unique = res.posts.filter((p) => !ids.has(p.id));
          return [...prev, ...unique];
        });
      }
      setNextCursor(res.next_cursor || null);
    } catch (err) {
      console.error("Failed to load more bookmarks", err);
    } finally {
      setLoadingMore(false);
    }
  };

  const handleBookmarkToggle = (postId: string) => {
    // If user unbookmarked from the bookmarks page, remove it from list
    setPosts((prev) => prev.filter((p) => p.id !== postId));
  };

  const handlePostDeleted = (postId: string) => {
    setPosts((prev) => prev.filter((p) => p.id !== postId));
  };

  return (
    <>
      <main className="w-full sm:w-[600px] border-r border-border min-h-screen flex flex-col pb-16 sm:pb-0">
          <header className="sticky top-0 bg-background/80 backdrop-blur-md border-b border-border z-20 flex items-center justify-between px-4 h-[53px]">
            <div className="flex items-center gap-3">
              <button
                onClick={() => setDrawerOpen(true)}
                className="sm:hidden w-8 h-8 rounded-full overflow-hidden border border-border flex items-center justify-center shrink-0"
              >
                {currentUser?.avatar_url ? (
                  <img
                    src={currentUser.avatar_url}
                    alt={currentUser.display_name}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="w-full h-full bg-[#16181c] flex items-center justify-center text-xs font-bold text-foreground">
                    {currentUser?.display_name?.charAt(0) || "X"}
                  </div>
                )}
              </button>
              <div>
                <h1 className="font-extrabold text-[20px] leading-6">Bookmarks</h1>
                <span className="text-[13px] text-muted leading-4">
                  @{currentUser?.username || "user"}
                </span>
              </div>
            </div>
            <button className="p-2 hover:bg-subtle rounded-full text-foreground transition-colors" title="More">
              <MoreHorizontal className="w-5 h-5" />
            </button>
          </header>

          {loading ? (
            <div className="flex justify-center items-center py-20">
              <Loader2 className="w-8 h-8 text-brand animate-spin" />
            </div>
          ) : !currentUser ? (
            <div className="text-center py-20 px-8">
              <div className="w-16 h-16 rounded-full bg-subtle flex items-center justify-center mx-auto mb-4 text-brand">
                <Bookmark className="w-8 h-8" />
              </div>
              <h2 className="text-[28px] font-extrabold mb-2">Save posts for later</h2>
              <p className="text-muted text-[15px] max-w-sm mx-auto mb-6">
                Don’t let the good ones fly away! Log in to bookmark posts and easily find them again.
              </p>
              <button
                onClick={() => setAuthModalOpen(true)}
                className="px-6 py-2.5 rounded-full bg-foreground text-background font-bold text-[15px] hover:opacity-90 transition-opacity"
              >
                Log in
              </button>
            </div>
          ) : posts.length === 0 ? (
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
            <div className="flex flex-col">
              {posts.map((post) => (
                <PostCard
                  key={post.id}
                  post={post}
                  currentUser={currentUser}
                  onBookmark={handleBookmarkToggle}
                  onPostDeleted={handlePostDeleted}
                />
              ))}

              {nextCursor && (
                <div className="p-4 flex justify-center">
                  <button
                    onClick={handleLoadMore}
                    disabled={loadingMore}
                    className="px-6 py-2 text-sm font-bold text-brand hover:bg-brand/10 rounded-full transition-colors disabled:opacity-50 flex items-center gap-2"
                  >
                    {loadingMore ? (
                      <>
                        <Loader2 className="w-4 h-4 animate-spin" />
                        Loading...
                      </>
                    ) : (
                      "Load more bookmarks"
                    )}
                  </button>
                </div>
              )}
            </div>
          )}
        </main>

        <RightSidebar currentUser={currentUser} />

      <MobileBottomNav onOpenComposer={() => setComposerModalOpen(true)} />

      {drawerOpen && currentUser && (
        <MobileDrawer
          isOpen={drawerOpen}
          onClose={() => setDrawerOpen(false)}
          currentUser={currentUser}
          onOpenAuth={() => setAuthModalOpen(true)}
          onLogout={handleLogout}
        />
      )}

      {/* Mobile Composer Modal */}
      <MobileComposerModal
        isOpen={composerModalOpen}
        onClose={() => setComposerModalOpen(false)}
        currentUser={currentUser}
        onPostSuccess={() => fetchBookmarks()}
        onOpenAuth={() => setAuthModalOpen(true)}
      />

      {authModalOpen && (
        <AuthModal
          isOpen={authModalOpen}
          onClose={() => setAuthModalOpen(false)}
          onSuccess={(u) => setCurrentUser(u)}
        />
      )}
    </>
  );
}
