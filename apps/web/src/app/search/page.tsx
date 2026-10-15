"use client";

import React, { useState, useEffect, useMemo, Suspense } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { useInfiniteQuery, useQueryClient } from "@tanstack/react-query";
import {
  ArrowLeft,
  SlidersHorizontal,
  Search,
  CheckCircle2,
  X,
  Loader2,
  Calendar,
  Heart,
  User as UserIcon,
} from "lucide-react";
import { LeftNavRail } from "@/components/layout/LeftNavRail";
import { RightSidebar } from "@/components/layout/RightSidebar";
import { SearchBarDropdown } from "@/components/search/SearchBarDropdown";
import { VirtualizedFeed } from "@/components/virtualized-list/VirtualizedFeed";
import { VerifiedBadge } from "@/components/common/VerifiedBadge";
import { AuthModal } from "@/components/auth/AuthModal";
import { User, Post } from "@/types";
import { apiRequest } from "@/lib/api";
import { searchKeys } from "@/lib/query-keys";
import { cn, formatNumber } from "@/lib/utils";

type SearchTab = "top" | "latest" | "people" | "media";

interface SearchPostsResponse {
  data: Post[];
  page_info: {
    next_cursor?: string;
    has_next_page: boolean;
  };
}

interface SearchUsersResponse {
  data: User[];
  page_info: {
    next_cursor?: string;
    has_next_page: boolean;
  };
}

function SearchPageContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const queryClient = useQueryClient();

  const rawQuery = searchParams.get("q") || "";
  const activeTab = (searchParams.get("tab") as SearchTab) || "top";

  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [authModalOpen, setAuthModalOpen] = useState(false);
  const [filtersOpen, setFiltersOpen] = useState(false);

  // Advanced Filters State
  const [filterFromUser, setFilterFromUser] = useState("");
  const [filterToUser, setFilterToUser] = useState("");
  const [filterMinLikes, setFilterMinLikes] = useState("");
  const [filterSinceDate, setFilterSinceDate] = useState("");
  const [filterUntilDate, setFilterUntilDate] = useState("");

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

  // Update URL state
  const handleTabChange = (tab: SearchTab) => {
    const params = new URLSearchParams(searchParams.toString());
    params.set("tab", tab);
    router.push(`/search?${params.toString()}`);
  };

  const handleSearchSubmit = (newQuery: string) => {
    const params = new URLSearchParams();
    params.set("q", newQuery);
    params.set("tab", activeTab);
    router.push(`/search?${params.toString()}`);
  };

  // 1. Post Query (for Top, Latest, Media)
  const isPostTab = activeTab === "top" || activeTab === "latest" || activeTab === "media";
  const {
    data: postsData,
    isLoading: postsLoading,
    isFetchingNextPage: postsFetchingNext,
    hasNextPage: postsHasNext,
    fetchNextPage: postsFetchNext,
    isError: postsIsError,
    error: postsError,
    refetch: postsRefetch,
  } = useInfiniteQuery({
    queryKey: searchKeys.posts(rawQuery, activeTab),
    queryFn: async ({ pageParam }) => {
      if (!rawQuery.trim()) {
        return { data: [], page_info: { next_cursor: undefined, has_next_page: false } };
      }
      const cursorParam = pageParam ? `&cursor=${encodeURIComponent(pageParam as string)}` : "";
      const url = `/search/posts?q=${encodeURIComponent(rawQuery)}&tab=${activeTab}&limit=20${cursorParam}`;
      return apiRequest<SearchPostsResponse>(url);
    },
    initialPageParam: "",
    getNextPageParam: (lastPage) => lastPage?.page_info?.next_cursor || undefined,
    enabled: isPostTab && rawQuery.trim().length > 0,
    maxPages: 5, // Evicts distant pages from browser RAM!
    staleTime: 60 * 1000,
  });

  // Deduplicated posts across pages
  const posts = useMemo(() => {
    if (!postsData?.pages) return [];
    const seen = new Set<string>();
    const result: Post[] = [];

    for (const page of postsData.pages) {
      if (page.data) {
        for (const p of page.data) {
          if (!seen.has(p.id)) {
            seen.add(p.id);
            result.push(p);
          }
        }
      }
    }
    return result;
  }, [postsData]);

  // 2. People Query (for People tab)
  const isPeopleTab = activeTab === "people";
  const {
    data: usersData,
    isLoading: usersLoading,
    isFetchingNextPage: usersFetchingNext,
    hasNextPage: usersHasNext,
    fetchNextPage: usersFetchNext,
    isError: usersIsError,
    refetch: usersRefetch,
  } = useInfiniteQuery({
    queryKey: searchKeys.users(rawQuery),
    queryFn: async ({ pageParam }) => {
      if (!rawQuery.trim()) {
        return { data: [], page_info: { next_cursor: undefined, has_next_page: false } };
      }
      const cursorParam = pageParam ? `&cursor=${encodeURIComponent(pageParam as string)}` : "";
      const url = `/search/users?q=${encodeURIComponent(rawQuery)}&limit=20${cursorParam}`;
      return apiRequest<SearchUsersResponse>(url);
    },
    initialPageParam: "",
    getNextPageParam: (lastPage) => lastPage?.page_info?.next_cursor || undefined,
    enabled: isPeopleTab && rawQuery.trim().length > 0,
    maxPages: 5,
    staleTime: 60 * 1000,
  });

  const users = useMemo(() => {
    if (!usersData?.pages) return [];
    const seen = new Set<string>();
    const result: User[] = [];
    for (const page of usersData.pages) {
      if (page.data) {
        for (const u of page.data) {
          if (!seen.has(u.id)) {
            seen.add(u.id);
            result.push(u);
          }
        }
      }
    }
    return result;
  }, [usersData]);

  // Apply Advanced Filters to Query
  const applyAdvancedFilters = () => {
    let constructed = rawQuery;
    if (filterFromUser.trim()) {
      constructed += ` from:${filterFromUser.trim()}`;
    }
    if (filterToUser.trim()) {
      constructed += ` to:${filterToUser.trim()}`;
    }
    if (filterMinLikes.trim() && Number(filterMinLikes) > 0) {
      constructed += ` min_faves:${filterMinLikes.trim()}`;
    }
    if (filterSinceDate) {
      constructed += ` since:${filterSinceDate}`;
    }
    if (filterUntilDate) {
      constructed += ` until:${filterUntilDate}`;
    }

    setFiltersOpen(false);
    handleSearchSubmit(constructed.trim());
  };

  const handleFollowToggle = async (targetUser: User) => {
    if (!currentUser) {
      setAuthModalOpen(true);
      return;
    }

    try {
      if (targetUser.is_following) {
        await apiRequest(`/users/${targetUser.username}/follow`, { method: "DELETE" });
      } else {
        await apiRequest(`/users/${targetUser.username}/follow`, { method: "POST" });
      }
      usersRefetch();
    } catch (e) {
      console.error("Follow action failed", e);
    }
  };

  return (
    <div className="min-h-screen bg-background text-foreground flex justify-center">
      <div className="w-full max-w-[1280px] flex">
        {/* Left Navigation Rail */}
        <LeftNavRail
          currentUser={currentUser}
          onOpenComposer={() => router.push("/")}
          onOpenAuth={() => setAuthModalOpen(true)}
          onLogout={() => {
            queryClient.clear();
            setCurrentUser(null);
          }}
        />

        {/* Center Search Container (600px) */}
        <main className="w-full sm:w-[600px] border-r border-border min-h-screen flex flex-col">
          {/* Sticky Header with Search Bar and Filter Toggle */}
          <header className="sticky top-0 bg-background/80 backdrop-blur-md border-b border-border z-20 flex flex-col">
            <div className="flex items-center gap-3 px-4 py-2.5">
              <button
                onClick={() => router.back()}
                className="p-2 hover:bg-subtle rounded-full text-foreground transition-colors shrink-0"
              >
                <ArrowLeft className="w-5 h-5" />
              </button>

              <div className="flex-1">
                <SearchBarDropdown
                  initialQuery={rawQuery}
                  onSearchSubmit={handleSearchSubmit}
                />
              </div>

              <button
                onClick={() => setFiltersOpen(!filtersOpen)}
                className={cn(
                  "p-2.5 rounded-full transition-colors shrink-0",
                  filtersOpen ? "bg-brand text-white" : "hover:bg-subtle text-foreground"
                )}
                title="Advanced filters"
              >
                <SlidersHorizontal className="w-5 h-5" />
              </button>
            </div>

            {/* Search Tabs: Top, Latest, People, Media */}
            <div className="flex w-full h-[53px]">
              {(["top", "latest", "people", "media"] as SearchTab[]).map((tab) => {
                const label = tab.charAt(0).toUpperCase() + tab.slice(1);
                const isActive = activeTab === tab;
                return (
                  <button
                    key={tab}
                    onClick={() => handleTabChange(tab)}
                    className="flex-1 flex justify-center items-center hover:bg-subtle transition-colors relative"
                  >
                    <span
                      className={cn(
                        "text-[15px] font-bold py-3.5 relative",
                        isActive ? "text-foreground" : "text-muted"
                      )}
                    >
                      {label}
                      {isActive && (
                        <div className="absolute bottom-0 left-0 right-0 h-1 bg-brand rounded-full" />
                      )}
                    </span>
                  </button>
                );
              })}
            </div>
          </header>

          {/* Advanced Filters Drawer / Modal */}
          {filtersOpen && (
            <div className="bg-subtle/30 border-b border-border p-4 transition-all">
              <div className="flex items-center justify-between mb-3">
                <span className="font-bold text-[16px]">Search filters</span>
                <button
                  onClick={() => setFiltersOpen(false)}
                  className="text-muted hover:text-foreground p-1"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-[14px]">
                <div>
                  <label className="text-muted text-[12px] block mb-1">From user (@username)</label>
                  <input
                    type="text"
                    placeholder="e.g. elonmusk"
                    value={filterFromUser}
                    onChange={(e) => setFilterFromUser(e.target.value)}
                    className="w-full bg-background border border-border rounded-lg px-3 py-1.5 outline-none focus:border-brand"
                  />
                </div>

                <div>
                  <label className="text-muted text-[12px] block mb-1">To user (@username)</label>
                  <input
                    type="text"
                    placeholder="e.g. jack"
                    value={filterToUser}
                    onChange={(e) => setFilterToUser(e.target.value)}
                    className="w-full bg-background border border-border rounded-lg px-3 py-1.5 outline-none focus:border-brand"
                  />
                </div>

                <div>
                  <label className="text-muted text-[12px] block mb-1">Minimum Likes</label>
                  <input
                    type="number"
                    placeholder="e.g. 100"
                    value={filterMinLikes}
                    onChange={(e) => setFilterMinLikes(e.target.value)}
                    className="w-full bg-background border border-border rounded-lg px-3 py-1.5 outline-none focus:border-brand"
                  />
                </div>

                <div>
                  <label className="text-muted text-[12px] block mb-1">Since Date</label>
                  <input
                    type="date"
                    value={filterSinceDate}
                    onChange={(e) => setFilterSinceDate(e.target.value)}
                    className="w-full bg-background border border-border rounded-lg px-3 py-1.5 outline-none focus:border-brand"
                  />
                </div>
              </div>

              <div className="flex justify-end gap-2 mt-4">
                <button
                  onClick={() => {
                    setFilterFromUser("");
                    setFilterToUser("");
                    setFilterMinLikes("");
                    setFilterSinceDate("");
                    setFilterUntilDate("");
                  }}
                  className="px-3 py-1.5 text-muted hover:text-foreground text-[13px] font-bold"
                >
                  Reset
                </button>
                <button
                  onClick={applyAdvancedFilters}
                  className="bg-brand hover:bg-brand-hover text-white text-[13px] font-bold px-4 py-1.5 rounded-full transition-colors"
                >
                  Apply filters
                </button>
              </div>
            </div>
          )}

          {/* Search Content */}
          <div className="flex-1 flex flex-col">
            {!rawQuery.trim() ? (
              <div className="text-center py-20 px-6">
                <div className="w-16 h-16 rounded-full bg-subtle flex items-center justify-center mx-auto mb-4 text-brand">
                  <Search className="w-8 h-8" />
                </div>
                <h3 className="font-extrabold text-[22px] mb-2">Search X</h3>
                <p className="text-muted text-[15px] max-w-sm mx-auto">
                  Search for topics, people, keywords, or operators like <span className="text-brand">from:username</span>.
                </p>
              </div>
            ) : isPeopleTab ? (
              /* People Tab View */
              <div className="flex flex-col divide-y divide-border">
                {usersLoading && users.length === 0 ? (
                  <div className="flex justify-center items-center py-16 text-brand">
                    <Loader2 className="w-8 h-8 animate-spin" />
                  </div>
                ) : users.length === 0 ? (
                  <div className="text-center py-20 px-6">
                    <h3 className="font-bold text-[20px] mb-2">No people found</h3>
                    <p className="text-muted text-[15px]">
                      No accounts matched the query "{rawQuery}".
                    </p>
                  </div>
                ) : (
                  <>
                    {users.map((u) => (
                      <div
                        key={u.id}
                        onClick={() => router.push(`/${u.username}`)}
                        className="p-4 hover:bg-subtle/50 cursor-pointer transition-colors flex items-start justify-between gap-3"
                      >
                        <div className="flex items-start gap-3 min-w-0">
                          <img
                            src={u.avatar_url || "https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=100"}
                            alt={u.display_name}
                            className="w-12 h-12 rounded-full object-cover shrink-0 border border-border"
                          />
                          <div className="flex flex-col min-w-0">
                            <div className="flex items-center gap-1">
                              <span className="font-bold text-[15px] text-foreground hover:underline truncate">
                                {u.display_name}
                              </span>
                              {u.is_verified && <VerifiedBadge type={u.verified_type} />}
                            </div>
                            <span className="text-[13px] text-muted leading-4">@{u.username}</span>
                            {u.bio && (
                              <p className="text-[14px] text-foreground mt-2 line-clamp-2">
                                {u.bio}
                              </p>
                            )}
                            <div className="flex items-center gap-4 mt-2 text-[12px] text-muted">
                              <span><strong className="text-foreground">{formatNumber(u.following_count)}</strong> Following</span>
                              <span><strong className="text-foreground">{formatNumber(u.followers_count)}</strong> Followers</span>
                            </div>
                          </div>
                        </div>

                        {currentUser && currentUser.id !== u.id && (
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              handleFollowToggle(u);
                            }}
                            className={cn(
                              "font-bold text-[14px] py-1.5 px-4 rounded-full transition-all shrink-0",
                              u.is_following
                                ? "border border-border text-foreground hover:border-red-500/50 hover:text-red-500"
                                : "bg-foreground text-background hover:opacity-90"
                            )}
                          >
                            {u.is_following ? "Following" : "Follow"}
                          </button>
                        )}
                      </div>
                    ))}

                    {/* Pagination for People */}
                    {usersHasNext && (
                      <div className="p-4 flex justify-center">
                        <button
                          onClick={() => usersFetchNext()}
                          disabled={usersFetchingNext}
                          className="w-full py-2.5 hover:bg-subtle rounded-xl text-brand font-bold text-[14px] flex items-center justify-center gap-2 border border-border/50"
                        >
                          {usersFetchingNext ? (
                            <>
                              <Loader2 className="w-4 h-4 animate-spin" />
                              <span>Loading more people...</span>
                            </>
                          ) : (
                            <span>Load more</span>
                          )}
                        </button>
                      </div>
                    )}
                  </>
                )}
              </div>
            ) : (
              /* Top, Latest, Media Virtualized Feed */
              <VirtualizedFeed
                posts={posts}
                isLoading={postsLoading}
                isFetchingNextPage={postsFetchingNext}
                hasNextPage={postsHasNext}
                fetchNextPage={postsFetchNext}
                isError={postsIsError}
                error={postsError as Error}
                refetch={postsRefetch}
                storageKey={`search_${activeTab}_${rawQuery}`}
                emptyState={
                  <div className="text-center py-20 px-6">
                    <h3 className="font-bold text-[20px] mb-2">No results for "{rawQuery}"</h3>
                    <p className="text-muted text-[15px] max-w-sm mx-auto">
                      Try searching for something else, or check your search terms for spelling errors.
                    </p>
                  </div>
                }
              />
            )}
          </div>
        </main>

        {/* Right Sidebar (with search bar hidden on search page) */}
        <RightSidebar hideSearchBar />

        {/* Auth Modal */}
        <AuthModal
          isOpen={authModalOpen}
          onClose={() => setAuthModalOpen(false)}
          onSuccess={(user) => {
            setCurrentUser(user);
            if (isPostTab) postsRefetch();
            if (isPeopleTab) usersRefetch();
          }}
        />
      </div>
    </div>
  );
}

export default function SearchPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen bg-background flex justify-center items-center">
          <Loader2 className="w-8 h-8 animate-spin text-brand" />
        </div>
      }
    >
      <SearchPageContent />
    </Suspense>
  );
}
