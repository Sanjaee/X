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

interface FeedResponse {
  posts: Post[];
  next_cursor?: string;
}

export default function HomePage() {
  const queryClient = useQueryClient();
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [activeTab, setActiveTab] = useState<"for_you" | "following">("for_you");
  const [authModalOpen, setAuthModalOpen] = useState(false);

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
          onOpenComposer={() => window.scrollTo({ top: 0, behavior: "smooth" })}
          onOpenAuth={() => setAuthModalOpen(true)}
          onLogout={handleLogout}
        />

        {/* Center Feed Timeline (600px max) */}
        <main className="w-full sm:w-[600px] border-r border-border min-h-screen flex flex-col">
          {/* Sticky Header with Feed Tabs */}
          <header className="sticky top-0 bg-background/80 backdrop-blur-md border-b border-border z-20 flex flex-col">
            <div className="flex h-[53px] items-center justify-between px-4 sm:hidden">
              <span className="font-extrabold text-[18px]">Home</span>
            </div>

            <div className="flex w-full h-[53px]">
              <button
                onClick={() => setActiveTab("for_you")}
                className="flex-1 flex justify-center items-center hover:bg-subtle transition-colors relative"
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
                className="flex-1 flex justify-center items-center hover:bg-subtle transition-colors relative"
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
            </div>
          </header>

          {/* Post Composer */}
          <PostComposer
            currentUser={currentUser}
            onPostSuccess={handlePostCreated}
            onOpenAuth={() => setAuthModalOpen(true)}
          />

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
