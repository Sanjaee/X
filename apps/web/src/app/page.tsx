"use client";

import React, { useEffect, useState, useMemo } from "react";
import { useInfiniteQuery, useQueryClient } from "@tanstack/react-query";
import { LeftNavRail } from "@/components/layout/LeftNavRail";
import { RightSidebar } from "@/components/layout/RightSidebar";
import { PostComposer } from "@/components/composer/PostComposer";
import { VirtualizedFeed } from "@/components/virtualized-list/VirtualizedFeed";
import { AuthModal } from "@/components/auth/AuthModal";
import { User, Post } from "@/types";
import { apiRequest, setStoredToken } from "@/lib/api";
import { feedKeys } from "@/lib/query-keys";
import { cn } from "@/lib/utils";
import { MobileDrawer } from "@/components/layout/MobileDrawer";
import { MobileBottomNav } from "@/components/layout/MobileBottomNav";
import { MobileComposerModal } from "@/components/composer/MobileComposerModal";
import { PostModal } from "@/components/composer/PostModal";
import { Plus, Radio, ArrowUp } from "lucide-react";
import Link from "next/link";

interface FeedResponse {
  posts: Post[];
  next_cursor?: string;
}

export default function HomePage() {
  const queryClient = useQueryClient();
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [activeTab, setActiveTab] = useState<"for_you" | "following">("for_you");
  const [authModalOpen, setAuthModalOpen] = useState(false);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [composerModalOpen, setComposerModalOpen] = useState(false);
  const [desktopPostModalOpen, setDesktopPostModalOpen] = useState(false);
  const [showNewPostsPill, setShowNewPostsPill] = useState(true);

  // Restore authenticated session
  useEffect(() => {
    const syncUser = () => {
      apiRequest<User>("/me")
        .then((user) => setCurrentUser(user))
        .catch(() => setCurrentUser(null));
    };
    syncUser();

    window.addEventListener("x_auth_state_change", syncUser);
    return () => window.removeEventListener("x_auth_state_change", syncUser);
  }, []);

  // TanStack Query with Bounded Cache (maxPages: 5)
  const {
    data,
    isLoading,
    isFetchingNextPage,
    hasNextPage,
    fetchNextPage,
    isError,
    error,
    refetch,
  } = useInfiniteQuery({
    queryKey: feedKeys.list(activeTab),
    queryFn: async ({ pageParam }) => {
      const cursorParam = pageParam ? `&cursor=${encodeURIComponent(pageParam as string)}` : "";
      const res = await apiRequest<FeedResponse>(`/feed?type=${activeTab}&limit=20${cursorParam}`);

      // Record exposure impressions for ranking feedback
      if (res.posts && res.posts.length > 0) {
        apiRequest("/feed/impressions", {
          method: "POST",
          body: JSON.stringify({
            impressions: res.posts.map((p) => ({
              post_id: p.id,
              feed_type: activeTab,
              dwell_time_ms: 1200,
            })),
          }),
        }).catch(() => {});
      }

      return res;
    },
    initialPageParam: "",
    getNextPageParam: (lastPage) => lastPage?.next_cursor || undefined,
    maxPages: 5, // Automatically evicts distant pages from browser memory!
    staleTime: 60 * 1000,
  });

  // Client-side post deduplication across pages
  const posts = useMemo(() => {
    if (!data?.pages) return [];
    const seen = new Set<string>();
    const result: Post[] = [];

    for (const page of data.pages) {
      if (page.posts) {
        for (const p of page.posts) {
          if (!seen.has(p.id)) {
            seen.add(p.id);
            result.push(p);
          }
        }
      }
    }
    return result;
  }, [data]);

  const handlePostCreated = (newPost: Post) => {
    // Ensure post author is populated even before refetching
    const postWithAuthor: Post = {
      ...newPost,
      author: newPost.author || (currentUser ? { ...currentUser } : undefined),
    };

    queryClient.setQueryData(
      feedKeys.list(activeTab),
      (oldData: { pages: FeedResponse[]; pageParams: string[] } | undefined) => {
        if (!oldData || !oldData.pages || oldData.pages.length === 0) {
          return {
            pages: [{ posts: [postWithAuthor], next_cursor: undefined }],
            pageParams: [""],
          };
        }
        const firstPage = oldData.pages[0];
        // Deduplicate in case it already exists
        const filteredPosts = firstPage.posts.filter((p) => p.id !== postWithAuthor.id);
        const updatedFirstPage = {
          ...firstPage,
          posts: [postWithAuthor, ...filteredPosts],
        };
        return {
          ...oldData,
          pages: [updatedFirstPage, ...oldData.pages.slice(1)],
        };
      }
    );

    // Reset scroll position to top and clear anchors
    if (typeof window !== "undefined") {
      sessionStorage.removeItem("x_feed_scroll_anchor");
      sessionStorage.removeItem("x_last_media_anchor");
      window.scrollTo({ top: 0, behavior: "smooth" });
    }
  };

  const handleFeedback = async (postId: string, feedbackType: "not_interested" | "hide") => {
    // Optimistically update query cache
    queryClient.setQueryData(
      feedKeys.list(activeTab),
      (oldData: { pages: FeedResponse[]; pageParams: string[] } | undefined) => {
        if (!oldData) return oldData;
        return {
          ...oldData,
          pages: oldData.pages.map((page) => ({
            ...page,
            posts: page.posts.filter((p) => p.id !== postId),
          })),
        };
      }
    );

    try {
      await apiRequest("/feed/feedback", {
        method: "POST",
        body: JSON.stringify({ post_id: postId, feedback_type: feedbackType }),
      });
    } catch (e) {
      console.error("Failed to record feed feedback", e);
    }
  };

  const handleLogout = async () => {
    try {
      await apiRequest("/auth/logout", { method: "POST" });
    } catch {}
    try {
      const { signOut } = await import("next-auth/react");
      await signOut({ redirect: false });
    } catch {}
    setStoredToken(null);
    setCurrentUser(null);
    queryClient.clear();
  };

  const handleLike = async (postId: string) => {
    if (!currentUser) {
      setAuthModalOpen(true);
      return;
    }
    try {
      await apiRequest(`/posts/${postId}/likes`, { method: "POST" });
    } catch (e) {
      console.error(e);
    }
  };

  const handleRepost = async (postId: string) => {
    if (!currentUser) {
      setAuthModalOpen(true);
      return;
    }
    try {
      await apiRequest(`/posts/${postId}/reposts`, { method: "POST" });
    } catch (e) {
      console.error(e);
    }
  };

  const handleBookmark = async (postId: string) => {
    if (!currentUser) {
      setAuthModalOpen(true);
      return;
    }
    try {
      await apiRequest(`/posts/${postId}/bookmarks`, { method: "POST" });
    } catch (e) {
      console.error(e);
    }
  };

  return (
    <div className="min-h-screen bg-background text-foreground flex justify-center">
      <div className="w-full max-w-[1280px] flex">
        {/* Left Navigation Rail (88px / 275px) */}
        <LeftNavRail
          currentUser={currentUser}
          onOpenComposer={() => setDesktopPostModalOpen(true)}
          onOpenAuth={() => setAuthModalOpen(true)}
          onLogout={handleLogout}
        />

        {/* Center Feed Timeline (600px max) */}
        <main className="w-full sm:w-[600px] border-r border-border min-h-screen flex flex-col pb-16 sm:pb-0 relative">
          {/* Sticky Header with Mobile Top Bar and Feed Tabs */}
          <header className="sticky top-0 bg-background/80 backdrop-blur-md border-b border-border z-20 flex flex-col">
            {/* Mobile Top Bar: Avatar, X logo, Subscribe */}
            <div className="flex h-[53px] items-center justify-between px-4 sm:hidden">
              <button
                onClick={() => setDrawerOpen(true)}
                className="w-8 h-8 rounded-full overflow-hidden border border-border flex items-center justify-center shrink-0"
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

              {/* Centered X Logo */}
              <div className="flex items-center justify-center">
                <svg viewBox="0 0 24 24" aria-hidden="true" className="w-5 h-5 fill-current">
                  <g>
                    <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 24.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z" />
                  </g>
                </svg>
              </div>

              {/* Subscribe Pill Button */}
              <Link
                href="/premium"
                className="text-xs font-bold px-3.5 py-1.5 rounded-full border border-border text-foreground hover:bg-subtle transition-colors"
              >
                Subscribe
              </Link>
            </div>

            {/* Feed Tabs: For you & Following + Add List Button */}
            <div className="flex w-full h-[53px] items-center">
              <button
                onClick={() => setActiveTab("for_you")}
                className="flex-1 flex justify-center items-center hover:bg-subtle transition-colors relative h-full"
              >
                <span
                  className={cn(
                    "text-[15px] font-bold py-3.5 relative",
                    activeTab === "for_you" ? "text-foreground" : "text-muted"
                  )}
                >
                  For you
                  {activeTab === "for_you" && (
                    <div className="absolute bottom-0 left-0 right-0 h-1 bg-brand rounded-full" />
                  )}
                </span>
              </button>

              <button
                onClick={() => {
                  if (!currentUser) {
                    setAuthModalOpen(true);
                    return;
                  }
                  setActiveTab("following");
                }}
                className="flex-1 flex justify-center items-center hover:bg-subtle transition-colors relative h-full"
              >
                <span
                  className={cn(
                    "text-[15px] font-bold py-3.5 relative",
                    activeTab === "following" ? "text-foreground" : "text-muted"
                  )}
                >
                  Following
                  {activeTab === "following" && (
                    <div className="absolute bottom-0 left-0 right-0 h-1 bg-brand rounded-full" />
                  )}
                </span>
              </button>

              <div className="px-3 sm:hidden">
                <button
                  aria-label="Add feed"
                  className="p-1.5 hover:bg-subtle rounded-full text-muted hover:text-foreground transition-colors"
                >
                  <Plus className="w-4 h-4" />
                </button>
              </div>
            </div>
          </header>

          {/* Floating Pill: New Posts Indicator (↑ 👥 New posts) */}
          {showNewPostsPill && (
            <button
              onClick={() => {
                window.scrollTo({ top: 0, behavior: "smooth" });
                setShowNewPostsPill(false);
                setTimeout(() => setShowNewPostsPill(true), 30000);
              }}
              className="fixed top-24 left-1/2 -translate-x-1/2 z-30 bg-[#1d9bf0] hover:bg-[#1a8cd8] active:scale-95 text-white px-3.5 py-1.5 rounded-full shadow-lg flex items-center gap-2 text-xs font-bold transition-transform cursor-pointer"
            >
              <ArrowUp className="w-3.5 h-3.5 stroke-[2.5]" />
              <div className="flex -space-x-1.5 overflow-hidden">
                <img
                  src="https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=50"
                  alt="user"
                  className="w-4 h-4 rounded-full ring-1 ring-white object-cover"
                />
                <img
                  src="https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=50"
                  alt="user"
                  className="w-4 h-4 rounded-full ring-1 ring-white object-cover"
                />
              </div>
              <span>New posts</span>
            </button>
          )}

          {/* Desktop Post Composer (Hidden on mobile) */}
          <div className="hidden sm:block">
            <PostComposer
              currentUser={currentUser}
              onPostSuccess={handlePostCreated}
              onOpenAuth={() => setAuthModalOpen(true)}
            />
          </div>

          {/* Virtualized Infinite Feed with Bounded Cache */}
          <div className="flex-1 flex flex-col">
            <VirtualizedFeed
              posts={posts}
              isLoading={isLoading}
              isFetchingNextPage={isFetchingNextPage}
              hasNextPage={hasNextPage}
              fetchNextPage={fetchNextPage}
              isError={isError}
              error={error as Error}
              refetch={refetch}
              onLike={handleLike}
              onRepost={handleRepost}
              onBookmark={handleBookmark}
              onFeedback={handleFeedback}
              storageKey={`feed_${activeTab}`}
              emptyState={
                <div className="text-center py-16 px-6">
                  <h3 className="font-bold text-[22px] mb-2">Welcome to X!</h3>
                  <p className="text-muted text-[15px] max-w-sm mx-auto mb-6">
                    This is the best place to see what’s happening in your world. Find people and topics to follow now.
                  </p>
                  {!currentUser && (
                    <button
                      onClick={() => setAuthModalOpen(true)}
                      className="bg-brand hover:bg-brand-hover text-white font-bold py-2.5 px-6 rounded-full text-[15px]"
                    >
                      Sign in to customize
                    </button>
                  )}
                </div>
              }
            />
          </div>
        </main>

        {/* Right Sidebar (420px max) */}
        <RightSidebar />

        {/* Mobile Navigation Drawer */}
        <MobileDrawer
          isOpen={drawerOpen}
          onClose={() => setDrawerOpen(false)}
          currentUser={currentUser}
          onOpenAuth={() => setAuthModalOpen(true)}
          onLogout={handleLogout}
        />

        {/* Desktop Post Modal Popup (Matches Screenshot) */}
        <PostModal
          isOpen={desktopPostModalOpen}
          onClose={() => setDesktopPostModalOpen(false)}
          currentUser={currentUser}
          onPostSuccess={handlePostCreated}
          onOpenAuth={() => setAuthModalOpen(true)}
        />

        {/* Mobile Composer Modal (Full screen sheet on FAB click) */}
        <MobileComposerModal
          isOpen={composerModalOpen}
          onClose={() => setComposerModalOpen(false)}
          currentUser={currentUser}
          onPostSuccess={handlePostCreated}
          onOpenAuth={() => setAuthModalOpen(true)}
        />

        {/* Mobile Fixed Bottom Nav */}
        <MobileBottomNav onOpenComposer={() => setComposerModalOpen(true)} />

        {/* Auth Modal */}
        <AuthModal
          isOpen={authModalOpen}
          onClose={() => setAuthModalOpen(false)}
          onSuccess={(user) => {
            setCurrentUser(user);
            refetch();
          }}
        />
      </div>
    </div>
  );
}
