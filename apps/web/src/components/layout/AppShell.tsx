"use client";

import React from "react";
import { usePathname } from "next/navigation";
import { LeftNavRail } from "@/components/layout/LeftNavRail";
import { RouteTransition } from "@/components/layout/RouteTransition";
import { MobileBottomNav } from "@/components/layout/MobileBottomNav";
import { useCurrentUser } from "@/lib/useCurrentUser";
import { useComposer } from "@/components/composer/ComposerContext";
import { cn } from "@/lib/utils";

export function AppShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const { currentUser, logout } = useCurrentUser();
  const { openComposer } = useComposer();

  // Standalone routes that manage their own full-width canvas
  const isStandalone =
    pathname === "/admin" ||
    pathname.startsWith("/admin/") ||
    pathname.includes("/photo/") ||
    pathname.includes("/video/");

  if (isStandalone) {
    return <>{children}</>;
  }

  return (
    <div className="min-h-screen bg-background text-foreground flex justify-center">
      <div
        className={cn(
          "w-full max-w-[1280px] flex",
          !currentUser && "justify-center"
        )}
      >
        {/* Persistent Left Nav Rail - stays mounted across all page navigations */}
        {currentUser && (
          <LeftNavRail currentUser={currentUser} onLogout={logout} />
        )}

        {/* Dynamic page content with subtle 150ms opacity transition */}
        <RouteTransition>{children}</RouteTransition>
      </div>

      {/* Persistent Mobile Bottom Navigation on phone viewports */}
      {currentUser && (
        <MobileBottomNav onOpenComposer={openComposer} />
      )}
    </div>
  );
}
