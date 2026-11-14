"use client";

import React, { useState, useRef, useEffect } from "react";
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
  BadgeCheck,
  Shield,
  Check,
  Plus,
} from "lucide-react";
import { User } from "@/types";
import { cn } from "@/lib/utils";
import { UserNameWithRole } from "@/components/ui/UserNameWithRole";
import { ConfirmModal } from "@/components/common/ConfirmModal";
import { useRealtime } from "@/lib/useRealtime";
import { useComposer } from "@/components/composer/ComposerContext";
import { useCurrentUser } from "@/lib/useCurrentUser";
import { useUpgradeModal } from "@/components/subscription/UpgradeModalContext";

interface LeftNavRailProps {
  currentUser?: User | null;
  onOpenComposer?: () => void;
  onOpenAuth?: () => void;
  onLogout?: () => void;
}

export function LeftNavRail({
  currentUser: propUser,
  onOpenComposer,
  onOpenAuth,
  onLogout: propLogout,
}: LeftNavRailProps) {
  const pathname = usePathname();
  const { openComposer } = useComposer();
  const { openUpgradeModal } = useUpgradeModal();
  const { currentUser: authUser, logout: authLogout } = useCurrentUser();
  const currentUser = propUser !== undefined ? propUser : authUser;
  const onLogout = propLogout || authLogout;

  const [menuOpen, setMenuOpen] = useState(false);
  const [logoutModalOpen, setLogoutModalOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen(false);
      }
    };
    if (menuOpen) {
      document.addEventListener("mousedown", handleClickOutside);
    }
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [menuOpen]);

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

  if (!currentUser) {
    return (
      <header className="hidden sm:flex sticky top-0 h-screen flex-col justify-start items-end xl:items-start px-2 sm:px-4 py-3 select-none w-[68px] xl:w-[275px] shrink-0">
        <div className="w-full flex flex-col items-center xl:items-start">
          {/* LM Logo */}
          <Link
            href="/"
            className="p-2 hover:bg-subtle rounded-full transition-colors flex items-center justify-center mb-1 text-foreground"
            aria-label="Home"
          >
            <img src="/LM.png" alt="LM Logo" className="w-9 h-9 object-contain" />
          </Link>
        </div>
      </header>
    );
  }

  return (
    <>
      <header className="hidden sm:flex sticky top-0 h-screen flex-col justify-between items-end xl:items-start px-2 sm:px-4 py-3 select-none border-r border-border w-[68px] xl:w-[275px] shrink-0">
        <div className="w-full flex flex-col items-center xl:items-start">
          {/* LM Logo */}
          <Link
            href="/"
            className="p-2 hover:bg-subtle rounded-full transition-colors flex items-center justify-center mb-1 text-foreground"
            aria-label="Home"
          >
            <img src="/LM.png" alt="LM Logo" className="w-9 h-9 object-contain" />
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

              if (item.label === "Premium") {
                return (
                  <button
                    key={item.label}
                    type="button"
                    onClick={() => openUpgradeModal()}
                    className="flex items-center gap-4 p-2.5 xl:py-3 xl:px-4 rounded-full hover:bg-subtle transition-colors group text-foreground w-full text-left"
                  >
                    <div className="relative flex items-center justify-center">
                      <Icon className="w-7 h-7 stroke-[2]" />
                    </div>
                    <span className="hidden xl:inline text-[20px] leading-6">{item.label}</span>
                  </button>
                );
              }

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

        {/* Account / User Switcher with More Options Popover */}
        {currentUser ? (
          <div className="relative w-full" ref={menuRef}>
            {/* Popover Menu */}
            {menuOpen && (
              <div className="absolute bottom-full left-0 mb-3 w-[300px] bg-black border border-[#2f3336] rounded-2xl shadow-[0_0_20px_rgba(0,0,0,0.8),0_0_15px_rgba(255,255,255,0.08)] py-3 z-[100] animate-in fade-in zoom-in-95 duration-150 select-none">
                {/* Active Account Info */}
                <div className="px-4 py-2.5 flex items-center justify-between border-b border-[#2f3336]">
                  <div className="flex items-center gap-3 min-w-0">
                    <img
                      src={
                        currentUser.avatar_url ||
                        "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"
                      }
                      alt={currentUser.display_name}
                      className="w-10 h-10 rounded-full object-cover border border-border shrink-0"
                    />
                    <div className="flex flex-col min-w-0 text-left">
                      <div className="text-[15px] font-bold text-foreground truncate flex items-center">
                        <UserNameWithRole
                          displayName={currentUser.display_name || currentUser.username}
                          role={currentUser.role}
                          nameStyle={currentUser.name_style}
                        />
                      </div>
                      <span className="text-muted text-[13px] truncate">@{currentUser.username}</span>
                    </div>
                  </div>
                  <Check className="w-5 h-5 text-brand shrink-0 ml-2" />
                </div>

                {/* Add existing account */}
                <button
                  type="button"
                  onClick={() => {
                    setMenuOpen(false);
                    if (onOpenAuth) onOpenAuth();
                  }}
                  className="w-full px-4 py-3 text-left text-[15px] font-bold hover:bg-subtle/70 transition-colors text-foreground flex items-center justify-between"
                >
                  <span>Add an existing account</span>
                </button>

                {/* Log out */}
                <button
                  type="button"
                  onClick={() => {
                    setMenuOpen(false);
                    setLogoutModalOpen(true);
                  }}
                  className="w-full px-4 py-3 text-left text-[15px] font-bold hover:bg-subtle/70 transition-colors text-foreground truncate"
                >
                  Log out @{currentUser.username}
                </button>
              </div>
            )}

            {/* Profile trigger button */}
            <div
              onClick={() => setMenuOpen((prev) => !prev)}
              className="w-full flex items-center justify-between p-2 hover:bg-subtle rounded-full cursor-pointer transition-colors group"
            >
              <div className="flex items-center gap-3 min-w-0">
                <img
                  src={
                    currentUser.avatar_url ||
                    "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"
                  }
                  alt={currentUser.display_name}
                  className="w-10 h-10 rounded-full object-cover border border-border shrink-0"
                />
                <div className="hidden xl:flex flex-col text-left min-w-0">
                  <div className="text-[15px] leading-5 truncate flex items-center">
                    <UserNameWithRole
                      displayName={currentUser.display_name || currentUser.username}
                      role={currentUser.role}
                      nameStyle={currentUser.name_style}
                    />
                  </div>
                  <span className="text-muted text-[14px] leading-4 truncate">
                    @{currentUser.username}
                  </span>
                </div>
              </div>

              {/* Three dots icon */}
              <div className="hidden xl:flex items-center justify-center p-1 text-foreground">
                <MoreHorizontal className="w-5 h-5 text-muted group-hover:text-foreground transition-colors" />
              </div>
            </div>
          </div>
        ) : null}
      </header>

      {/* Logout Confirmation Modal */}
      <ConfirmModal
        isOpen={logoutModalOpen}
        onClose={() => setLogoutModalOpen(false)}
        onConfirm={async () => {
          setLogoutModalOpen(false);
          if (onLogout) await onLogout();
        }}
        title={`Log out of @${currentUser?.username || "account"}?`}
        description="You can always log back in at any time. If you just want to switch accounts, you can add an existing account."
        confirmText="Log out"
        cancelText="Cancel"
        variant="destructive"
      />
    </>
  );
}
