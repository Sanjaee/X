"use client";

import React from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  Home,
  Search,
  Bell,
  Mail,
  Bookmark,
  User as UserIcon,
  Sparkles,
  MoreHorizontal,
  Feather,
  LogOut,
  BadgeCheck,
  Shield,
} from "lucide-react";
import { User } from "@/types";
import { cn } from "@/lib/utils";
import { VerifiedBadge } from "@/components/common/VerifiedBadge";
import { useRealtime } from "@/lib/useRealtime";
import { useComposer } from "@/components/composer/ComposerContext";

interface LeftNavRailProps {
  currentUser: User | null;
  onOpenComposer?: () => void;
  onOpenAuth?: () => void;
  onLogout?: () => void;
}

export function LeftNavRail({
  currentUser,
  onOpenComposer,
  onOpenAuth,
  onLogout,
}: LeftNavRailProps) {
  const pathname = usePathname();
  const { openComposer } = useComposer();

  const handlePostClick = () => {
    if (onOpenComposer) {
      onOpenComposer();
    } else {
      openComposer();
    }
  };

  const isAdmin =
    currentUser?.is_owner ||
    currentUser?.role === "owner" ||
    currentUser?.role === "god" ||
    currentUser?.role === "mod" ||
    currentUser?.email === "afrizaahmad18@gmail.com";

  const navItems = [
    { label: "Home", href: "/", icon: Home },
    { label: "Explore", href: "/explore", icon: Search },
    { label: "Notifications", href: "/notifications", icon: Bell },
    { label: "Messages", href: "/messages", icon: Mail },
    { label: "Grok", href: "/grok", icon: Sparkles },
    { label: "Bookmarks", href: "/bookmarks", icon: Bookmark },
    { label: "Premium", href: "/premium", icon: BadgeCheck },
    ...(isAdmin ? [{ label: "Admin", href: "/admin", icon: Shield }] : []),
    {
      label: "Profile",
      href: currentUser ? `/${currentUser.username}` : "/login",
      icon: UserIcon,
    },
  ];

  const { unreadNotifCount, unreadMessageCount } = useRealtime();

  return (
    <header className="hidden sm:flex sticky top-0 h-screen flex-col justify-between items-end xl:items-start px-2 sm:px-4 py-3 select-none border-r border-border w-[68px] xl:w-[275px] overflow-y-auto no-scrollbar">
      <div className="w-full flex flex-col items-center xl:items-start">
        {/* X Logo */}
        <Link
          href="/"
          className="p-3 hover:bg-subtle rounded-full transition-colors flex items-center justify-center mb-1 text-foreground"
          aria-label="X"
        >
          <svg viewBox="0 0 24 24" aria-hidden="true" className="w-8 h-8 fill-current">
            <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z" />
          </svg>
        </Link>

        {/* Navigation Items */}
        <nav className="flex flex-col gap-0.5 xl:gap-1 w-full">
          {navItems.map((item) => {
            const Icon = item.icon;
            const isActive = pathname === item.href;
            const badgeCount =
              item.label === "Notifications"
                ? unreadNotifCount
                : item.label === "Messages"
                ? unreadMessageCount
                : 0;

            return (
              <Link
                key={item.label}
                href={item.href}
                className={cn(
                  "flex items-center gap-4 p-2.5 xl:py-3 xl:px-4 rounded-full hover:bg-subtle transition-colors group text-foreground",
                  isActive && "font-bold"
                )}
              >
                <div className="relative flex items-center justify-center">
                  <Icon className={cn("w-7 h-7 stroke-[2]", isActive && "stroke-[2.5]")} />
                  {badgeCount > 0 && (
                    <span className="absolute -top-1 -right-1 bg-brand text-white text-[11px] font-extrabold px-1.5 py-0.2 rounded-full min-w-[18px] text-center shadow-md animate-pulse">
                      {badgeCount > 99 ? "99+" : badgeCount}
                    </span>
                  )}
                </div>
                <span className="hidden xl:inline text-[20px] leading-6">{item.label}</span>
              </Link>
            );
          })}
        </nav>

        {/* Post Button CTA */}
        <div className="w-full mt-2 xl:mt-4">
          {currentUser ? (
            <button
              onClick={handlePostClick}
              className="w-full bg-[#eff3f4] hover:bg-[#d7dbdc] text-[#0f1419] font-bold p-3 xl:py-3.5 xl:px-8 rounded-full shadow-md transition-all flex items-center justify-center active:scale-95"
            >
              <Feather className="w-6 h-6 xl:hidden" />
              <span className="hidden xl:inline text-[17px]">Post</span>
            </button>
          ) : (
            <button
              onClick={onOpenAuth}
              className="w-full bg-[#eff3f4] hover:bg-[#d7dbdc] text-[#0f1419] font-bold p-3 xl:py-3.5 xl:px-8 rounded-full shadow-md transition-all flex items-center justify-center active:scale-95"
            >
              <span className="hidden xl:inline text-[17px]">Sign In</span>
              <UserIcon className="w-6 h-6 xl:hidden" />
            </button>
          )}
        </div>
      </div>

      {/* Account / User Switcher */}
      {currentUser ? (
        <div className="w-full flex items-center justify-between p-2 hover:bg-subtle rounded-full cursor-pointer transition-colors group">
          <div className="flex items-center gap-3">
            <img
              src={currentUser.avatar_url || "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"}
              alt={currentUser.display_name}
              className="w-10 h-10 rounded-full object-cover border border-border"
            />
            <div className="hidden xl:flex flex-col text-left">
              <span className="font-bold text-[15px] leading-5 line-clamp-1 flex items-center">
                {currentUser.display_name}
                {currentUser.is_verified && (
                  <VerifiedBadge type={currentUser.verified_type} />
                )}
              </span>
              <span className="text-muted text-[14px] leading-4">@{currentUser.username}</span>
            </div>

          </div>
          <button
            onClick={onLogout}
            title="Log out"
            className="hidden xl:block p-1.5 hover:bg-subtle rounded-full text-muted hover:text-foreground"
          >
            <LogOut className="w-5 h-5" />
          </button>
        </div>
      ) : null}
    </header>
  );
}
