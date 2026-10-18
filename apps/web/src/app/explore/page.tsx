"use client";

import React, { useState } from "react";
import { Search, Settings, MoreHorizontal } from "lucide-react";
import { LeftNavRail } from "@/components/layout/LeftNavRail";
import { RightSidebar } from "@/components/layout/RightSidebar";
import { PostCard } from "@/components/feed/PostCard";
import { Post, User } from "@/types";
import { apiRequest } from "@/lib/api";
import { formatNumber, cn } from "@/lib/utils";
import { useCurrentUser } from "@/lib/useCurrentUser";

export default function ExplorePage() {
  const { currentUser } = useCurrentUser();
  const [query, setQuery] = useState("");
  const [activeTab, setActiveTab] = useState<"for_you" | "trending" | "news" | "sports">("for_you");
  const [searchResults, setSearchResults] = useState<Post[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [trendingTopics, setTrendingTopics] = useState<Array<{ category: string; topic: string; posts_count: number }>>([]);

  React.useEffect(() => {
    apiRequest<{ trending: Array<{ category: string; topic: string; posts_count: number }> }>("/explore/trending")
      .then((res) => {
        setTrendingTopics(res.trending || []);
      })
      .catch((e) => console.error("Failed to load trending topics", e));
  }, []);

  const handleSearch = async (e?: React.FormEvent, customQuery?: string) => {
    if (e) e.preventDefault();
    const q = customQuery !== undefined ? customQuery : query;
    if (!q.trim()) return;
    setIsSearching(true);
    try {
      const res = await apiRequest<{ posts: Post[] }>(`/search?q=${encodeURIComponent(q)}`);
      setSearchResults(res.posts || []);
    } catch (e) {
      console.error(e);
    } finally {
      setIsSearching(false);
    }
  };

  return (
    <>
      <main
        className={cn(
          "w-full sm:w-[600px] border-r border-border min-h-screen flex flex-col",
          !currentUser && "border-l"
        )}
      >
          {/* Top Search Header */}
          <header className="sticky top-0 bg-background/80 backdrop-blur-md border-b border-border z-20 flex flex-col">
            <div className="flex items-center gap-3 px-4 h-[53px]">
              <form onSubmit={handleSearch} className="flex-1 relative flex items-center bg-transparent rounded-full border border-border focus-within:border-brand focus-within:bg-background transition-colors">
                <Search className="w-5 h-5 text-muted absolute left-4 pointer-events-none" />
                <input
                  type="text"
                  placeholder="Search"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  className="w-full bg-transparent py-2 pl-12 pr-4 text-[15px] outline-none placeholder:text-muted rounded-full text-foreground"
                />
              </form>
              <button className="p-2 hover:bg-subtle rounded-full text-foreground transition-colors">
                <Settings className="w-5 h-5" />
              </button>
            </div>

            {/* Tabs */}
            <div className="flex w-full h-[53px]">
              {(["for_you", "trending", "news", "sports"] as const).map((tab) => (
                <button
                  key={tab}
                  onClick={() => setActiveTab(tab)}
                  className="flex-1 flex justify-center items-center hover:bg-subtle transition-colors relative capitalize text-[15px]"
                >
                  <span
                    className={cn(
                      "font-bold py-3.5 relative",
                      activeTab === tab ? "text-foreground" : "text-muted"
                    )}
                  >
                    {tab.replace("_", " ")}
                    {activeTab === tab && (
                      <div className="absolute bottom-0 left-0 right-0 h-1 bg-brand rounded-full" />
                    )}
                  </span>
                </button>
              ))}
            </div>
          </header>

          {/* Results or Trends */}
          {searchResults.length > 0 ? (
            <div className="flex flex-col">
              {searchResults.map((post) => (
                <PostCard key={post.id} post={post} />
              ))}
            </div>
          ) : (
            <div className="flex flex-col divide-y divide-border">
              {trendingTopics.map((t) => (
                <div
                  key={t.topic}
                  onClick={() => {
                    setQuery(t.topic);
                    handleSearch(undefined, t.topic);
                  }}
                  className="px-4 py-3.5 hover:bg-subtle cursor-pointer transition-colors flex justify-between items-start"
                >
                  <div>
                    <span className="text-[13px] text-muted leading-4 block">{t.category}</span>
                    <span className="text-[16px] font-bold leading-5 block mt-0.5">{t.topic}</span>
                    <span className="text-[13px] text-muted leading-4 block mt-0.5">
                      {formatNumber(t.posts_count || 0)} posts
                    </span>
                  </div>
                  <button className="text-muted hover:text-foreground p-1 hover:bg-subtle rounded-full">
                    <MoreHorizontal className="w-4 h-4" />
                  </button>
                </div>
              ))}
            </div>
          )}
        </main>

        <RightSidebar currentUser={currentUser} />
    </>
  );
}
