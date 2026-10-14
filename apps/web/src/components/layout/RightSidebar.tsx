"use client";

import React, { useState } from "react";
import Link from "next/link";
import { Search, MoreHorizontal, CheckCircle2 } from "lucide-react";
import { formatNumber } from "@/lib/utils";

interface RightSidebarProps {
  onSearch?: (query: string) => void;
}

export function RightSidebar({ onSearch }: RightSidebarProps) {
  const [searchTerm, setSearchTerm] = useState("");

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

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter" && onSearch) {
      onSearch(searchTerm);
    }
  };

  return (
    <aside className="hidden lg:block w-[350px] xl:w-[420px] px-6 py-3 select-none">
      {/* Search Bar */}
      <div className="sticky top-0 bg-background/80 backdrop-blur-md pt-1 pb-3 z-10">
        <div className="relative flex items-center bg-card rounded-full border border-transparent focus-within:border-brand focus-within:bg-background transition-colors">
          <Search className="w-5 h-5 text-muted absolute left-4 pointer-events-none" />
          <input
            type="text"
            placeholder="Search"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            onKeyDown={handleKeyDown}
            className="w-full bg-transparent py-3 pl-12 pr-4 text-[15px] outline-none placeholder:text-muted rounded-full"
          />
        </div>
      </div>

      {/* Subscribe to Premium Card */}
      <div className="bg-card rounded-2xl p-4 border border-border mb-4">
        <h2 className="font-extrabold text-[20px] leading-6 mb-2">Subscribe to Premium</h2>
        <p className="text-[15px] leading-5 text-muted mb-3">
          Subscribe to unlock new features and if eligible, receive a creator revenue share.
        </p>
        <button className="bg-brand hover:bg-brand-hover text-white font-bold py-2 px-5 rounded-full text-[15px] transition-colors">
          Subscribe
        </button>
      </div>

      {/* What's Happening / Trends */}
      <div className="bg-card rounded-2xl border border-border mb-4 overflow-hidden">
        <h2 className="font-extrabold text-[20px] leading-6 p-4">What's happening</h2>
        <div className="flex flex-col">
          {trends.map((t) => (
            <div
              key={t.topic}
              className="px-4 py-3 hover:bg-subtle cursor-pointer transition-colors flex justify-between items-start"
            >
              <div>
                <span className="text-[13px] text-muted leading-4 block">{t.category}</span>
                <span className="text-[15px] font-bold leading-5 block">{t.topic}</span>
                <span className="text-[13px] text-muted leading-4 block">{formatNumber(t.posts)} posts</span>
              </div>
              <button className="text-muted hover:text-foreground p-1 hover:bg-subtle rounded-full">
                <MoreHorizontal className="w-4 h-4" />
              </button>
            </div>
          ))}
        </div>
      </div>

      {/* Who to Follow */}
      <div className="bg-card rounded-2xl border border-border mb-4 overflow-hidden">
        <h2 className="font-extrabold text-[20px] leading-6 p-4">Who to follow</h2>
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
                    <span className="font-bold text-[15px] leading-5 hover:underline">{u.name}</span>
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
    </aside>
  );
}
