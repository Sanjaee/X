"use client";

import React, { useState, useEffect, useRef, useCallback } from "react";
import { useRouter } from "next/navigation";
import { Search, X, Clock, Loader2, ArrowUpRight } from "lucide-react";
import { apiRequest } from "@/lib/api";
import { User } from "@/types";
import { VerifiedBadge } from "@/components/common/VerifiedBadge";

interface RecentSearchItem {
  id?: string;
  query: string;
}

interface SuggestionData {
  queries: string[];
  users: User[];
}

interface SearchBarDropdownProps {
  initialQuery?: string;
  onSearchSubmit?: (query: string) => void;
  autoFocus?: boolean;
  className?: string;
}

export function SearchBarDropdown({
  initialQuery = "",
  onSearchSubmit,
  autoFocus = false,
  className = "",
}: SearchBarDropdownProps) {
  const router = useRouter();
  const [query, setQuery] = useState(initialQuery);
  const [isOpen, setIsOpen] = useState(false);
  const [recentSearches, setRecentSearches] = useState<RecentSearchItem[]>([]);
  const [suggestions, setSuggestions] = useState<SuggestionData>({ queries: [], users: [] });
  const [isLoading, setIsLoading] = useState(false);
  const [selectedIndex, setSelectedIndex] = useState<number>(-1);

  const containerRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const abortControllerRef = useRef<AbortController | null>(null);

  // Sync external initialQuery if updated
  useEffect(() => {
    setQuery(initialQuery);
  }, [initialQuery]);

  // Load recent searches from backend or localStorage
  const loadRecentSearches = useCallback(async () => {
    try {
      const res = await apiRequest<{ data: { id: string; query: string }[] }>("/search/recent");
      if (res && res.data && res.data.length > 0) {
        setRecentSearches(res.data);
        return;
      }
    } catch {
      // Fallback to local storage if unauthenticated
    }

    if (typeof window !== "undefined") {
      const local = localStorage.getItem("x_recent_searches");
      if (local) {
        try {
          const parsed = JSON.parse(local) as string[];
          setRecentSearches(parsed.map((q) => ({ query: q })));
        } catch {}
      }
    }
  }, []);

  // Save to recent searches
  const saveRecentSearch = useCallback((searchQuery: string) => {
    const trimmed = searchQuery.trim();
    if (!trimmed) return;

    if (typeof window !== "undefined") {
      const local = localStorage.getItem("x_recent_searches");
      let list: string[] = [];
      if (local) {
        try {
          list = JSON.parse(local);
        } catch {}
      }
      list = [trimmed, ...list.filter((q) => q.toLowerCase() !== trimmed.toLowerCase())].slice(0, 10);
      localStorage.setItem("x_recent_searches", JSON.stringify(list));
    }
  }, []);

  // Clear single recent search
  const deleteRecentItem = async (e: React.MouseEvent, item: RecentSearchItem) => {
    e.stopPropagation();
    if (item.id) {
      try {
        await apiRequest(`/search/recent/${item.id}`, { method: "DELETE" });
      } catch {}
    }
    setRecentSearches((prev) => prev.filter((s) => s.query !== item.query));

    if (typeof window !== "undefined") {
      const local = localStorage.getItem("x_recent_searches");
      if (local) {
        try {
          const list = JSON.parse(local) as string[];
          const updated = list.filter((q) => q !== item.query);
          localStorage.setItem("x_recent_searches", JSON.stringify(updated));
        } catch {}
      }
    }
  };

  // Clear all recent searches
  const clearAllRecent = async (e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await apiRequest("/search/recent", { method: "DELETE" });
    } catch {}
    setRecentSearches([]);
    if (typeof window !== "undefined") {
      localStorage.removeItem("x_recent_searches");
    }
  };

  // Debounced Autocomplete with AbortController (200ms)
  useEffect(() => {
    const trimmed = query.trim();
    if (!trimmed) {
      setSuggestions({ queries: [], users: [] });
      setIsLoading(false);
      return;
    }

    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }
    const controller = new AbortController();
    abortControllerRef.current = controller;

    setIsLoading(true);
    const timer = setTimeout(async () => {
      try {
        const data = await apiRequest<SuggestionData>(
          `/search/suggestions?q=${encodeURIComponent(trimmed)}`,
          { signal: controller.signal }
        );
        setSuggestions(data || { queries: [], users: [] });
      } catch (err: unknown) {
        if ((err as Error)?.name !== "AbortError") {
          setSuggestions({ queries: [], users: [] });
        }
      } finally {
        setIsLoading(false);
      }
    }, 200);

    return () => {
      clearTimeout(timer);
      controller.abort();
    };
  }, [query]);

  // Click outside listener
  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const handleNavigateSearch = (targetQuery: string) => {
    const trimmed = targetQuery.trim();
    if (!trimmed) return;
    saveRecentSearch(trimmed);
    setIsOpen(false);
    if (onSearchSubmit) {
      onSearchSubmit(trimmed);
    } else {
      router.push(`/search?q=${encodeURIComponent(trimmed)}&tab=top`);
    }
  };

  // Total selectable items in dropdown for keyboard navigation
  const allSelectableItems: { type: "recent" | "query" | "user"; value: string; user?: User }[] = [
    ...(!query.trim()
      ? recentSearches.map((r) => ({ type: "recent" as const, value: r.query }))
      : [
          ...suggestions.queries.map((q) => ({ type: "query" as const, value: q })),
          ...suggestions.users.map((u) => ({ type: "user" as const, value: u.username, user: u })),
        ]),
  ];

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "ArrowDown") {
      e.preventDefault();
      if (!isOpen) {
        setIsOpen(true);
        loadRecentSearches();
        return;
      }
      setSelectedIndex((prev) => (prev < allSelectableItems.length - 1 ? prev + 1 : 0));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setSelectedIndex((prev) => (prev > 0 ? prev - 1 : allSelectableItems.length - 1));
    } else if (e.key === "Enter") {
      e.preventDefault();
      if (selectedIndex >= 0 && selectedIndex < allSelectableItems.length) {
        const selected = allSelectableItems[selectedIndex];
        if (selected.type === "user" && selected.user) {
          router.push(`/${selected.user.username}`);
          setIsOpen(false);
        } else {
          handleNavigateSearch(selected.value);
        }
      } else {
        handleNavigateSearch(query);
      }
    } else if (e.key === "Escape") {
      setIsOpen(false);
      inputRef.current?.blur();
    }
  };

  return (
    <div ref={containerRef} className={`relative w-full ${className}`}>
      {/* Search Input Bar */}
      <div
        className={`relative flex items-center rounded-full border transition-colors ${
          isOpen
            ? "border-brand bg-background shadow-md ring-1 ring-brand"
            : "border-border bg-subtle/50 hover:bg-subtle"
        }`}
      >
        <Search className={`w-5 h-5 absolute left-4 pointer-events-none ${isOpen ? "text-brand" : "text-muted"}`} />
        <input
          ref={inputRef}
          type="text"
          placeholder="Search"
          autoFocus={autoFocus}
          value={query}
          onFocus={() => {
            setIsOpen(true);
            loadRecentSearches();
          }}
          onChange={(e) => {
            setQuery(e.target.value);
            setIsOpen(true);
            setSelectedIndex(-1);
          }}
          onKeyDown={handleKeyDown}
          className="w-full bg-transparent py-2.5 pl-12 pr-10 text-[15px] outline-none placeholder:text-muted rounded-full text-foreground"
        />
        {query && (
          <button
            onClick={() => {
              setQuery("");
              setSuggestions({ queries: [], users: [] });
              inputRef.current?.focus();
            }}
            className="absolute right-3 p-1 rounded-full bg-muted/20 hover:bg-muted/40 text-muted hover:text-foreground transition-colors"
          >
            <X className="w-3.5 h-3.5" />
          </button>
        )}
      </div>

      {/* Autocomplete & Recent Searches Dropdown */}
      {isOpen && (
        <div className="absolute top-12 left-0 right-0 bg-background border border-border shadow-2xl rounded-2xl overflow-hidden z-50 py-2 max-h-[460px] overflow-y-auto">
          {/* Loading Indicator */}
          {isLoading && (
            <div className="flex items-center justify-center py-4 text-brand">
              <Loader2 className="w-5 h-5 animate-spin" />
            </div>
          )}

          {/* When Query is empty: Show Recent Searches */}
          {!query.trim() && (
            <div>
              <div className="flex items-center justify-between px-4 py-2">
                <span className="font-extrabold text-[17px] text-foreground">Recent</span>
                {recentSearches.length > 0 && (
                  <button
                    onClick={clearAllRecent}
                    className="text-brand hover:underline text-[13px] font-bold"
                  >
                    Clear all
                  </button>
                )}
              </div>

              {recentSearches.length === 0 ? (
                <div className="px-4 py-6 text-center text-muted text-[14px]">
                  Try searching for people, topics, or keywords
                </div>
              ) : (
                <div className="flex flex-col">
                  {recentSearches.map((item, idx) => (
                    <div
                      key={item.query}
                      onClick={() => handleNavigateSearch(item.query)}
                      className={`px-4 py-3 flex items-center justify-between hover:bg-subtle cursor-pointer transition-colors ${
                        selectedIndex === idx ? "bg-subtle" : ""
                      }`}
                    >
                      <div className="flex items-center gap-3">
                        <Clock className="w-4 h-4 text-muted shrink-0" />
                        <span className="text-[15px] font-medium text-foreground">{item.query}</span>
                      </div>
                      <button
                        onClick={(e) => deleteRecentItem(e, item)}
                        className="p-1 hover:bg-muted/20 rounded-full text-muted hover:text-foreground"
                      >
                        <X className="w-4 h-4" />
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* When Query is active: Show Suggestions & User Matches */}
          {query.trim() && !isLoading && (
            <div>
              {/* Direct query submit shortcut */}
              <div
                onClick={() => handleNavigateSearch(query)}
                className="px-4 py-3 flex items-center justify-between hover:bg-subtle cursor-pointer border-b border-border/40"
              >
                <div className="flex items-center gap-3">
                  <Search className="w-4 h-4 text-brand shrink-0" />
                  <span className="text-[15px] text-foreground">
                    Search for <span className="font-bold">"{query}"</span>
                  </span>
                </div>
                <ArrowUpRight className="w-4 h-4 text-muted" />
              </div>

              {/* Matching Users Section */}
              {suggestions.users.length > 0 && (
                <div className="py-2 border-b border-border/40">
                  <div className="px-4 py-1 text-[13px] font-bold text-muted uppercase tracking-wider">
                    People
                  </div>
                  {suggestions.users.map((u, idx) => {
                    const itemIndex = suggestions.queries.length + idx;
                    return (
                      <div
                        key={u.id}
                        onClick={() => {
                          router.push(`/${u.username}`);
                          setIsOpen(false);
                        }}
                        className={`px-4 py-2.5 flex items-center gap-3 hover:bg-subtle cursor-pointer transition-colors ${
                          selectedIndex === itemIndex ? "bg-subtle" : ""
                        }`}
                      >
                        <img
                          src={u.avatar_url || "https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=100"}
                          alt={u.display_name}
                          className="w-10 h-10 rounded-full object-cover shrink-0 border border-border"
                        />
                        <div className="flex flex-col min-w-0">
                          <div className="flex items-center gap-1">
                            <span className="font-bold text-[15px] text-foreground truncate">
                              {u.display_name}
                            </span>
                            {u.is_verified && <VerifiedBadge type={u.verified_type} />}
                          </div>
                          <span className="text-[13px] text-muted truncate">@{u.username}</span>
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}

              {/* Matching Queries / Hashtags Section */}
              {suggestions.queries.length > 0 && (
                <div className="py-2">
                  <div className="px-4 py-1 text-[13px] font-bold text-muted uppercase tracking-wider">
                    Keywords
                  </div>
                  {suggestions.queries.map((q, idx) => (
                    <div
                      key={q}
                      onClick={() => handleNavigateSearch(q)}
                      className={`px-4 py-2.5 flex items-center gap-3 hover:bg-subtle cursor-pointer transition-colors ${
                        selectedIndex === idx ? "bg-subtle" : ""
                      }`}
                    >
                      <Search className="w-4 h-4 text-muted shrink-0" />
                      <span className="text-[15px] text-foreground font-medium">{q}</span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
