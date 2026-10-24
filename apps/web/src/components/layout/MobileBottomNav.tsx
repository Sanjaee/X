"use client";

import React from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  Home,
  Search,
  Compass,
  Bell,
  MessageCircle,
  Feather,
} from "lucide-react";
import { useRealtime } from "@/lib/useRealtime";
import { cn } from "@/lib/utils";

interface MobileBottomNavProps {
  onOpenComposer: () => void;
}

export function MobileBottomNav({ onOpenComposer }: MobileBottomNavProps) {
  const pathname = usePathname();
  const { unreadNotifCount, unreadMessageCount } = useRealtime();

  const isHome = pathname === "/";
  const isSearch = pathname.startsWith("/search") || pathname.startsWith("/explore");
  const isNotifs = pathname.startsWith("/notifications");
  const isMessages = pathname.startsWith("/messages");

  return (
    <>
      {/* Floating Action Button (FAB - Post Composer) matching X */}
      <button
        onClick={onOpenComposer}
        aria-label="New post"
        className="fixed bottom-[68px] right-4 z-40 sm:hidden w-[54px] h-[54px] rounded-full bg-brand hover:bg-brand-hover active:scale-95 text-white flex items-center justify-center shadow-lg transition-transform focus:outline-none"
      >
        <Feather className="w-6 h-6 stroke-[2.2]" />
      </button>

      {/* Fixed Bottom Tab Bar matching X App */}
      <nav className="fixed bottom-0 left-0 right-0 z-40 bg-background/95 backdrop-blur-md border-t border-border flex items-center justify-around h-[53px] sm:hidden pb-safe select-none">
        {/* 1. Home */}
        <Link
          href="/"
          className="flex-1 flex flex-col items-center justify-center h-full relative text-foreground hover:opacity-80 transition-opacity"
          aria-label="Home"
        >
          <div className="relative">
            <Home
              className={cn(
                "w-[26px] h-[26px]",
                isHome ? "stroke-[2.8] text-foreground" : "stroke-[1.8] text-foreground/80"
              )}
            />
            {isHome && (
              <span className="absolute -top-0.5 -right-0.5 w-2 h-2 rounded-full bg-brand" />
            )}
          </div>
        </Link>

        {/* 2. Search / Explore */}
        <Link
          href="/search"
          className="flex-1 flex flex-col items-center justify-center h-full relative text-foreground hover:opacity-80 transition-opacity"
          aria-label="Search"
        >
          <Search
            className={cn(
              "w-[26px] h-[26px]",
              isSearch ? "stroke-[2.8] text-foreground" : "stroke-[1.8] text-foreground/80"
            )}
          />
        </Link>


        {/* 4. Notifications */}
        <Link
          href="/notifications"
          className="flex-1 flex flex-col items-center justify-center h-full relative text-foreground hover:opacity-80 transition-opacity"
          aria-label="Notifications"
        >
          <div className="relative">
            <Bell
              className={cn(
                "w-[26px] h-[26px]",
                isNotifs ? "stroke-[2.8] text-foreground" : "stroke-[1.8] text-foreground/80"
              )}
            />
            {unreadNotifCount > 0 && (
              <span className="absolute -top-1.5 -right-2 bg-brand text-white text-[10px] font-extrabold px-1.5 py-0.2 rounded-full min-w-[16px] text-center shadow">
                {unreadNotifCount > 9 ? "9+" : unreadNotifCount}
              </span>
            )}
          </div>
        </Link>

        {/* 5. Messages / DMs */}
        <Link
          href="/messages"
          className="flex-1 flex flex-col items-center justify-center h-full relative text-foreground hover:opacity-80 transition-opacity"
          aria-label="Messages"
        >
          <div className="relative">
            <MessageCircle
              className={cn(
                "w-[26px] h-[26px]",
                isMessages ? "stroke-[2.8] text-foreground" : "stroke-[1.8] text-foreground/80"
              )}
            />
            {unreadMessageCount > 0 && (
              <span className="absolute -top-1.5 -right-2 bg-brand text-white text-[10px] font-extrabold px-1.5 py-0.2 rounded-full min-w-[16px] text-center shadow">
                {unreadMessageCount > 9 ? "9+" : unreadMessageCount}
              </span>
            )}
          </div>
        </Link>
      </nav>
    </>
  );
}
