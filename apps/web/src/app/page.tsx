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
import { useCurrentUser } from "@/lib/useCurrentUser";

interface FeedResponse {
  posts: Post[];
  next_cursor?: string;
}

function encodeFeedCursor(createdAt: string, id: string): string {
  try {
    const nano = new Date(createdAt).getTime() * 1_000_000;
    const raw = `${nano}:${id}`;
    if (typeof window !== "undefined") {
      return btoa(raw).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
    }
  } catch {}
  return "";
}

export default function HomePage() {
  const queryClient = useQueryClient();
  const { currentUser, logout: handleLogout, updateUser: setCurrentUser } = useCurrentUser();
  const [activeTab, setActiveTab] = useState<"for_you" | "following">("for_you");
  const [authModalOpen, setAuthModalOpen] = useState(false);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [composerModalOpen, setComposerModalOpen] = useState(false);
  const [desktopPostModalOpen, setDesktopPostModalOpen] = useState(false);

  // Real-time new posts detection state
  const [unseenCount, setUnseenCount] = useState<number>(0);
  const [pendingPosts, setPendingPosts] = useState<Post[]>([]);

  // Restore authenticated session & persisted tab preference
  useEffect(() => {
    // Restore tab from URL search param or localStorage
    if (typeof window !== "undefined") {
      const params = new URLSearchParams(window.location.search);
      const tabParam = params.get("tab");
      if (tabParam === "following") {
        setActiveTab("following");
      } else {
        const savedTab = localStorage.getItem("x_home_tab");
        if (savedTab === "following") {
          setActiveTab("following");
        }
      }
    }

    const handleCreated = (e: Event) => {
      const customEvent = e as CustomEvent<Post>;
      if (customEvent.detail) {
        handlePostCreated(customEvent.detail);
      }
    };

    const handleRemoteFeedPost = (e: Event) => {
      const customEvent = e as CustomEvent<Post>;
      const newPost = customEvent.detail;
      if (!newPost || !newPost.id) return;

      const userToken = localStorage.getItem("x_session_user");
      let myId: string | null = null;
      if (userToken) {
        try {
          myId = JSON.parse(userToken).id;
        } catch {}
      }
      if (myId && newPost.author_id === myId) return;

      setPendingPosts((prev) => {
        if (prev.some((p) => p.id === newPost.id)) return prev;
        return [newPost, ...prev];
      });
      setUnseenCount((prev) => prev + 1);
    };

    window.addEventListener("x_post_created", handleCreated);
    window.addEventListener("x_feed_post_created", handleRemoteFeedPost);
    return () => {
      window.removeEventListener("x_post_created", handleCreated);
      window.removeEventListener("x_feed_post_created", handleRemoteFeedPost);
    };
  }, []);

  const handleTabSwitch = (tab: "for_you" | "following") => {
    if (tab === "following" && !currentUser) {
      setAuthModalOpen(true);
      return;
    }
    setActiveTab(tab);
    setUnseenCount(0);
    setPendingPosts([]);
    if (typeof window !== "undefined") {
      localStorage.setItem("x_home_tab", tab);
      const url = tab === "following" ? "/?tab=following" : "/";
      window.history.replaceState(null, "", url);
    }
  };

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

  // Extract distinct author avatars for the floating indicator (matching X UI)
  const pendingAuthors = useMemo(() => {
    const map = new Map<string, { id: string; name: string; avatar_url?: string }>();
    for (const p of pendingPosts) {
      if (p.author && !map.has(p.author.id)) {
        map.set(p.author.id, {
          id: p.author.id,
          name: p.author.display_name || p.author.username || "User",
          avatar_url: p.author.avatar_url,
        });
      }
    }
    return Array.from(map.values()).slice(0, 3);
  }, [pendingPosts]);

  // Real-time Detection of New Eligible Posts (Feature B)
  useEffect(() => {
    if (!posts || posts.length === 0) return;
    const newestPost = posts[0];
    const sinceCursor = encodeFeedCursor(newestPost.created_at, newestPost.id);
    if (!sinceCursor) return;

    let isCancelled = false;

    const checkNewPosts = async () => {
      // Suspend polling when tab is inactive or hidden
      if (typeof document !== "undefined" && document.visibilityState === "hidden") {
        return;
      }

      try {
        const res = await apiRequest<{
          count: number;
          has_new: boolean;
          posts?: Post[];
        }>(`/feed/new-posts?type=${activeTab}&since=${encodeURIComponent(sinceCursor)}&limit=20`);

        if (!isCancelled && res.count > 0) {
          setUnseenCount(res.count);
          if (res.posts && res.posts.length > 0) {
            setPendingPosts(res.posts);
          }
        }
      } catch (err) {
        // Silently tolerate transient failures
      }
    };

    // Initial check after 10s, then poll every 20s
    const timer = setTimeout(checkNewPosts, 10000);
    const interval = setInterval(checkNewPosts, 20000);

    const handleVisibilityChange = () => {
      if (document.visibilityState === "visible") {
        checkNewPosts();
      }
    };
    document.addEventListener("visibilitychange", handleVisibilityChange);

    return () => {
      isCancelled = true;
      clearTimeout(timer);
      clearInterval(interval);
      document.removeEventListener("visibilitychange", handleVisibilityChange);
    };
  }, [activeTab, posts[0]?.id]);

  const handleLoadNewPosts = () => {
    if (pendingPosts.length > 0) {
      queryClient.setQueryData(
        feedKeys.list(activeTab),
        (oldData: { pages: FeedResponse[]; pageParams: string[] } | undefined) => {
          if (!oldData || !oldData.pages || oldData.pages.length === 0) {
            return oldData;
          }
          const firstPage = oldData.pages[0];
          const existingIDs = new Set(firstPage.posts.map((p) => p.id));
          const newUnique = pendingPosts.filter((p) => !existingIDs.has(p.id));

          return {
            ...oldData,
            pages: [
              {
                ...firstPage,
                posts: [...newUnique, ...firstPage.posts],
              },
              ...oldData.pages.slice(1),
            ],
          };
        }
      );
    } else {
      refetch();
    }

    setUnseenCount(0);
    setPendingPosts([]);
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

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


  const handleLike = (_postId: string) => {
    if (!currentUser) {
      setAuthModalOpen(true);
    }
  };

  const handleRepost = (_postId: string) => {
    if (!currentUser) {
      setAuthModalOpen(true);
    }
  };

  const handleBookmark = (_postId: string) => {
    if (!currentUser) {
      setAuthModalOpen(true);
    }
  };

  return (
    <>
      {/* Center Feed Timeline (600px max) */}
      <main
          className={cn(
            "w-full sm:w-[600px] border-r border-border min-h-screen flex flex-col pb-16 sm:pb-0 relative",
            !currentUser && "border-l"
          )}
        >
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
                onClick={() => handleTabSwitch("for_you")}
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
                onClick={() => handleTabSwitch("following")}
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

          {/* Floating Blue Pill: Real-Time New Posts Indicator (Exact X Style with Overlapping Avatars) */}
          {unseenCount > 0 && (
            <button
              onClick={handleLoadNewPosts}
              className="fixed top-20 sm:top-24 left-1/2 -translate-x-1/2 z-30 bg-[#1d9bf0] hover:bg-[#1a8cd8] active:scale-95 text-white px-3.5 py-1.5 rounded-full shadow-[0_4px_16px_rgba(29,155,240,0.45)] flex items-center gap-2 text-[14px] font-bold transition-all cursor-pointer animate-in fade-in slide-in-from-top-2 duration-200 select-none"
            >
              <ArrowUp className="w-4 h-4 stroke-[2.75]" />
              {pendingAuthors.length > 0 && (
                <div className="flex -space-x-2 items-center">
                  {pendingAuthors.map((author) => (
                    <div
                      key={author.id}
                      className="w-5 h-5 rounded-full overflow-hidden border-2 border-[#1d9bf0] bg-[#16181c] shrink-0"
                    >
                      {author.avatar_url ? (
                        <img
                          src={author.avatar_url}
                          alt={author.name}
                          className="w-full h-full object-cover"
                        />
                      ) : (
                        <div className="w-full h-full flex items-center justify-center text-[9px] font-bold text-white bg-slate-700">
                          {author.name.charAt(0)}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}
              <span>posted</span>
            </button>
          )}

          {/* Desktop Post Composer (Hidden on mobile and unauthenticated) */}
          {currentUser && (
            <div className="hidden sm:block">
              <PostComposer
                currentUser={currentUser}
                onPostSuccess={handlePostCreated}
                onOpenAuth={() => setAuthModalOpen(true)}
              />
            </div>
          )}

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
        <RightSidebar
          currentUser={currentUser}
          onOpenAuth={() => setAuthModalOpen(true)}
        />

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

        {/* Auth Modal */}
        <AuthModal
          isOpen={authModalOpen}
          onClose={() => setAuthModalOpen(false)}
          onSuccess={(user) => {
            setCurrentUser(user);
            refetch();
          }}
        />
    </>
  );
}
