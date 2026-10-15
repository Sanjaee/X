"use client";

import React from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { MoreHorizontal, CheckCircle2 } from "lucide-react";
import { formatNumber } from "@/lib/utils";
import { SearchBarDropdown } from "@/components/search/SearchBarDropdown";

interface RightSidebarProps {
  onSearch?: (query: string) => void;
  hideSearchBar?: boolean;
}

export function RightSidebar({ onSearch, hideSearchBar = false }: RightSidebarProps) {
  const router = useRouter();

  const trends = [
    { category: "Technology · Trending", topic: "#NextJS15", posts: 142000 },
    { category: "Space · Trending", topic: "Artemis III", posts: 89400 },
    { category: "AI · Trending", topic: "DeepSeek", posts: 320000 },
    { category: "Software Engineering", topic: "Golang 1.25", posts: 45100 },
    { category: "Trending in Web", topic: "Tailwind CSS v4", posts: 12500 },
  ];

  const suggestedUsers = [
    {
      name: "NASA",
      username: "NASA",
      avatar: "https://images.unsplash.com/photo-1614728894747-a83421e2b9c9?w=150&auto=format&fit=crop&q=80",
      verified: true,
      verifiedType: "government",
    },
    {
      name: "Go Lang",
      username: "golang",
      avatar: "https://images.unsplash.com/photo-1517694712202-14dd9538aa97?w=150&auto=format&fit=crop&q=80",
      verified: true,
      verifiedType: "blue",
    },
  ];

  return (
    <aside className="hidden lg:block w-[350px] xl:w-[420px] px-6 py-3 select-none">
      {/* Search Bar with Autocomplete Dropdown */}
      {!hideSearchBar && (
        <div className="sticky top-0 bg-background/80 backdrop-blur-md pt-1 pb-3 z-20">
          <SearchBarDropdown onSearchSubmit={onSearch} />
        </div>
      )}

      {/* Subscribe to Premium Card */}
      <div className="bg-transparent rounded-2xl p-4 border border-border mb-4">
        <h2 className="font-extrabold text-[20px] leading-6 mb-1 text-foreground">Subscribe to Premium</h2>
        <p className="text-[15px] leading-5 text-foreground mb-3 font-normal">
          Subscribe to unlock new features and if eligible, receive a creator revenue share.
        </p>
        <Link
          href="/premium"
          className="inline-block bg-brand hover:bg-brand-hover text-white font-bold py-2 px-5 rounded-full text-[15px] transition-colors"
        >
          Subscribe
        </Link>
      </div>

      {/* Sticky Container for What's happening & Who to follow (matches X behavior) */}
      <div className="sticky top-16 space-y-4 pb-12">
        {/* What's Happening / Trends */}
        <div className="bg-transparent rounded-2xl border border-border overflow-hidden">
          <h2 className="font-extrabold text-[20px] leading-6 px-4 pt-3 pb-2 text-foreground">What's happening</h2>
          <div className="flex flex-col">
            {trends.map((t) => (
              <div
                key={t.topic}
                onClick={() => router.push(`/search?q=${encodeURIComponent(t.topic)}`)}
                className="px-4 py-3 hover:bg-subtle cursor-pointer transition-colors flex justify-between items-start"
              >
                <div>
                  <span className="text-[13px] text-muted leading-4 block">{t.category}</span>
                  <span className="text-[15px] font-bold text-foreground leading-5 block mt-0.5">{t.topic}</span>
                  <span className="text-[13px] text-muted leading-4 block mt-0.5">{formatNumber(t.posts)} posts</span>
                </div>
                <button className="text-muted hover:text-foreground p-1 hover:bg-subtle rounded-full">
                  <MoreHorizontal className="w-4 h-4" />
                </button>
              </div>
            ))}
            <Link
              href="/explore"
              className="px-4 py-3 text-brand text-[15px] hover:bg-subtle transition-colors block"
            >
              Show more
            </Link>
          </div>
        </div>

        {/* Who to Follow */}
        <div className="bg-transparent rounded-2xl border border-border overflow-hidden">
          <h2 className="font-extrabold text-[20px] leading-6 px-4 pt-3 pb-2 text-foreground">Who to follow</h2>
          <div className="flex flex-col">
            {suggestedUsers.map((u) => (
              <div
                key={u.username}
                className="px-4 py-3 hover:bg-subtle cursor-pointer transition-colors flex items-center justify-between"
              >
                <div className="flex items-center gap-3">
                  <img src={u.avatar} alt={u.name} className="w-10 h-10 rounded-full object-cover border border-border" />
                  <div className="flex flex-col">
                    <div className="flex items-center gap-1">
                      <span className="font-bold text-[15px] leading-5 text-foreground hover:underline">{u.name}</span>
                      {u.verified && (
                        <CheckCircle2 className="w-4 h-4 fill-brand text-background inline" />
                      )}
                    </div>
                    <span className="text-[14px] text-muted leading-4">@{u.username}</span>
                  </div>
                </div>
                <button className="bg-foreground text-background font-bold text-[14px] py-1.5 px-4 rounded-full hover:opacity-90 transition-opacity">
                  Follow
                </button>
              </div>
            ))}
            <Link
              href="/explore"
              className="px-4 py-3 text-brand text-[15px] hover:bg-subtle transition-colors block"
            >
              Show more
            </Link>
          </div>
        </div>

        {/* Footer Links */}
        <footer className="text-[13px] text-muted px-4 leading-5 flex flex-wrap gap-x-3 gap-y-1">
          <Link href="#" className="hover:underline">Terms of Service</Link>
          <Link href="#" className="hover:underline">Privacy Policy</Link>
          <Link href="#" className="hover:underline">Cookie Policy</Link>
          <Link href="#" className="hover:underline">Accessibility</Link>
          <Link href="#" className="hover:underline">Ads info</Link>
          <span>© 2026 X Corp. (Reconstructed)</span>
        </footer>
      </div>
    </aside>
  );
}
