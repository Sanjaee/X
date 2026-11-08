"use client";

import React, { useEffect } from "react";
import Link from "next/link";
import {
  X,
  User as UserIcon,
  BadgeCheck,
  Bookmark,
  List,
  Mic,
  Coins,
  Settings,
  LogOut,
  Moon,
  Sparkles,
} from "lucide-react";
import { User } from "@/types";
import { UserNameWithRole } from "@/components/ui/UserNameWithRole";
import { formatNumber } from "@/lib/utils";
import { useUpgradeModal } from "@/components/subscription/UpgradeModalContext";

interface MobileDrawerProps {
  isOpen: boolean;
  onClose: () => void;
  currentUser: User | null;
  onOpenAuth: () => void;
  onLogout: () => void;
}

export function MobileDrawer({
  isOpen,
  onClose,
  currentUser,
  onOpenAuth,
  onLogout,
}: MobileDrawerProps) {
  const { openUpgradeModal } = useUpgradeModal();

  // Prevent body scroll when drawer is open
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = "hidden";
    } else {
      document.body.style.overflow = "";
    }
    return () => {
      document.body.style.overflow = "";
    };
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 sm:hidden">
      {/* Backdrop */}
      <div
        onClick={onClose}
        className="fixed inset-0 bg-black/60 backdrop-blur-sm transition-opacity animate-in fade-in duration-200"
      />

      {/* Drawer Panel */}
      <div className="fixed inset-y-0 left-0 w-[290px] max-w-[85vw] bg-background text-foreground shadow-2xl flex flex-col justify-between p-5 z-50 animate-in slide-in-from-left duration-200 border-r border-border">
        <div className="flex flex-col">
          {/* Header Row: User Avatar & Close */}
          <div className="flex items-center justify-between mb-3">
            {currentUser ? (
              <Link
                href={`/${currentUser.username}`}
                onClick={onClose}
                className="relative block"
              >
                <img
                  src={
                    currentUser.avatar_url ||
                    "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"
                  }
                  alt={currentUser.display_name}
                  className="w-11 h-11 rounded-full object-cover border border-border hover:opacity-90 transition-opacity"
                />
              </Link>
            ) : (
              <div
                onClick={() => {
                  onClose();
                  onOpenAuth();
                }}
                className="w-11 h-11 rounded-full bg-subtle flex items-center justify-center cursor-pointer"
              >
                <UserIcon className="w-6 h-6 text-muted" />
              </div>
            )}

            <button
              onClick={onClose}
              className="p-2 hover:bg-subtle rounded-full text-muted hover:text-foreground transition-colors"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          {/* User Name & Handle */}
          {currentUser ? (
            <div className="mb-4">
              <Link
                href={`/${currentUser.username}`}
                onClick={onClose}
                className="flex items-center gap-1 font-bold text-[17px] leading-tight hover:underline"
              >
                <UserNameWithRole
                  displayName={currentUser.display_name || currentUser.username}
                  role={currentUser.role}
                  nameStyle={currentUser.name_style}
                />
              </Link>
              <span className="text-muted text-[14px]">
                @{currentUser.username}
              </span>

              {/* Follower / Following Stats */}
              <div className="flex items-center gap-4 mt-3 text-[14px]">
                <Link
                  href={`/${currentUser.username}/following`}
                  onClick={onClose}
                  className="hover:underline flex items-center gap-1 text-muted"
                >
                  <span className="font-bold text-foreground">
                    {formatNumber(currentUser.following_count || 0)}
                  </span>
                  <span>Following</span>
                </Link>
                <Link
                  href={`/${currentUser.username}/followers`}
                  onClick={onClose}
                  className="hover:underline flex items-center gap-1 text-muted"
                >
                  <span className="font-bold text-foreground">
                    {formatNumber(currentUser.followers_count || 0)}
                  </span>
                  <span>Followers</span>
                </Link>
              </div>
            </div>
          ) : (
            <div className="mb-5">
              <h3 className="font-bold text-[17px] mb-1">Welcome to X</h3>
              <p className="text-muted text-[13px] mb-3">
                Sign in to interact, post, and see your personalized timeline.
              </p>
              <button
                onClick={() => {
                  onClose();
                  onOpenAuth();
                }}
                className="w-full bg-brand hover:bg-brand-hover text-white font-bold py-2 rounded-full text-sm transition-colors"
              >
                Sign In / Sign Up
              </button>
            </div>
          )}

          <div className="border-t border-border my-2" />

          {/* Menu Items matching X */}
          <nav className="flex flex-col gap-1 text-[17px] font-semibold text-foreground">
            {currentUser && (
              <Link
                href={`/${currentUser.username}`}
                onClick={onClose}
                className="flex items-center gap-4 py-3 px-2 rounded-xl hover:bg-subtle transition-colors"
              >
                <UserIcon className="w-6 h-6 stroke-[2]" />
                <span>Profile</span>
              </Link>
            )}

            <button
              type="button"
              onClick={() => {
                onClose();
                openUpgradeModal();
              }}
              className="flex items-center gap-4 py-3 px-2 rounded-xl hover:bg-subtle transition-colors w-full text-left"
            >
              <BadgeCheck className="w-6 h-6 stroke-[2] text-brand" />
              <span>Premium</span>
            </button>

            <Link
              href="/bookmarks"
              onClick={onClose}
              className="flex items-center gap-4 py-3 px-2 rounded-xl hover:bg-subtle transition-colors"
            >
              <Bookmark className="w-6 h-6 stroke-[2]" />
              <span>Bookmarks</span>
            </Link>

            <Link
              href="/explore"
              onClick={onClose}
              className="flex items-center gap-4 py-3 px-2 rounded-xl hover:bg-subtle transition-colors"
            >
              <Sparkles className="w-6 h-6 stroke-[2]" />
              <span>Explore Topics</span>
            </Link>

            <button
              type="button"
              onClick={() => {
                onClose();
                openUpgradeModal();
              }}
              className="flex items-center gap-4 py-3 px-2 rounded-xl hover:bg-subtle transition-colors w-full text-left"
            >
              <Coins className="w-6 h-6 stroke-[2]" />
              <span>Monetization</span>
            </button>
          </nav>
        </div>

        {/* Footer Area: Settings & Logout */}
        <div className="border-t border-border pt-3 flex flex-col gap-2">
          {currentUser && (
            <button
              onClick={() => {
                onClose();
                onLogout();
              }}
              className="flex items-center gap-3 py-2.5 px-2 rounded-xl hover:bg-subtle text-red-500 font-semibold text-[15px] transition-colors text-left"
            >
              <LogOut className="w-5 h-5" />
              <span>Log out @{currentUser.username}</span>
            </button>
          )}

          <div className="flex items-center justify-between px-2 pt-1 text-muted text-xs">
            <span>X for Mobile</span>
            <div className="p-2 hover:bg-subtle rounded-full text-foreground cursor-pointer">
              <Moon className="w-4 h-4" />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
