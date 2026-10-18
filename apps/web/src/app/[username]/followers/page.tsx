"use client";

import React, { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Loader2, UserPlus } from "lucide-react";
import { LeftNavRail } from "@/components/layout/LeftNavRail";
import { RightSidebar } from "@/components/layout/RightSidebar";
import { AuthModal } from "@/components/auth/AuthModal";
import { VerifiedBadge } from "@/components/common/VerifiedBadge";
import { User } from "@/types";
import { apiRequest } from "@/lib/api";
import { cn } from "@/lib/utils";
import { useCurrentUser } from "@/lib/useCurrentUser";

export default function FollowersPage() {
  const params = useParams();
  const router = useRouter();
  const username = params.username as string;

  const [targetUser, setTargetUser] = useState<User | null>(null);
  const { currentUser, updateUser: setCurrentUser } = useCurrentUser();
  const [followers, setFollowers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [authModalOpen, setAuthModalOpen] = useState(false);
  const [hoveredUserId, setHoveredUserId] = useState<string | null>(null);

  // Load target user profile and followers list
  useEffect(() => {
    if (!username) return;
    setLoading(true);

    Promise.all([
      apiRequest<User>(`/users/${username}`),
      apiRequest<{ users: User[] }>(`/users/${username}/followers`),
    ])
      .then(([user, followersRes]) => {
        setTargetUser(user);
        setFollowers(followersRes.users || []);
      })
      .catch((err) => {
        console.error("Failed to load followers", err);
      })
      .finally(() => {
        setLoading(false);
      });
  }, [username]);

  const handleToggleFollow = async (user: User) => {
    if (!currentUser) {
      setAuthModalOpen(true);
      return;
    }

    const prevFollowing = user.is_following;

    // Optimistic state update
    setFollowers((prev) =>
      prev.map((u) =>
        u.id === user.id ? { ...u, is_following: !prevFollowing } : u
      )
    );

    try {
      if (prevFollowing) {
        await apiRequest(`/users/${user.username}/follow`, { method: "DELETE" });
      } else {
        await apiRequest(`/users/${user.username}/follow`, { method: "POST" });
      }
    } catch (e) {
      console.error("Follow toggle failed", e);
      // Revert optimistic update
      setFollowers((prev) =>
        prev.map((u) =>
          u.id === user.id ? { ...u, is_following: prevFollowing } : u
        )
      );
    }
  };

  return (
    <>
      {/* Main Feed Column */}
      <main className="flex-1 max-w-[600px] border-r border-border min-h-screen pb-20">
          {/* Top Sticky Header */}
          <header className="sticky top-0 bg-background/80 backdrop-blur-md border-b border-border z-20 flex items-center gap-6 px-4 h-[53px]">
            <button
              onClick={() => router.push(`/${username}`)}
              className="p-2 hover:bg-subtle rounded-full text-foreground transition-colors"
              title="Back"
            >
              <ArrowLeft className="w-5 h-5" />
            </button>
            <div className="flex flex-col">
              <h1 className="font-extrabold text-[18px] sm:text-[20px] leading-tight flex items-center gap-1">
                <span>{targetUser?.display_name || username}</span>
                {targetUser?.is_verified && (
                  <VerifiedBadge type={targetUser.verified_type} />
                )}
              </h1>
              <span className="text-muted text-[13px]">@{username}</span>
            </div>
          </header>

          {/* Navigation Tabs (Standard X) */}
          <div className="flex border-b border-border text-[15px] font-bold">
            <Link
              href={`/${username}/followers`}
              className="flex-1 flex justify-center py-3.5 hover:bg-subtle/50 transition-colors relative"
            >
              <span className="text-foreground">Followers</span>
              <div className="absolute bottom-0 h-1 w-16 bg-brand rounded-full" />
            </Link>
            <Link
              href={`/${username}/following`}
              className="flex-1 flex justify-center py-3.5 hover:bg-subtle/50 transition-colors text-muted hover:text-foreground"
            >
              <span>Following</span>
            </Link>
          </div>

          {/* Followers List Content */}
          {loading ? (
            <div className="flex justify-center p-12 text-brand">
              <Loader2 className="w-7 h-7 animate-spin" />
            </div>
          ) : followers.length === 0 ? (
            <div className="text-center py-16 px-6">
              <UserPlus className="w-12 h-12 mx-auto mb-3 text-muted/50" />
              <h2 className="text-[20px] font-bold text-foreground mb-1">
                Looking for followers?
              </h2>
              <p className="text-muted text-[15px] max-w-sm mx-auto">
                When someone follows @{username}, they’ll show up here.
              </p>
            </div>
          ) : (
            <div className="divide-y divide-border">
              {followers.map((fUser) => {
                const isSelf = currentUser?.id === fUser.id;
                const isHovered = hoveredUserId === fUser.id;

                return (
                  <div
                    key={fUser.id}
                    className="p-4 hover:bg-subtle/40 transition-colors flex items-start justify-between gap-3"
                  >
                    <div className="flex items-start gap-3 min-w-0">
                      <Link href={`/${fUser.username}`} className="shrink-0">
                        <img
                          src={
                            fUser.avatar_url ||
                            "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"
                          }
                          alt={fUser.display_name}
                          className="w-10 h-10 rounded-full object-cover border border-border hover:opacity-90 transition-opacity"
                        />
                      </Link>

                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-1 flex-wrap">
                          <Link
                            href={`/${fUser.username}`}
                            className="font-bold text-[15px] hover:underline truncate"
                          >
                            {fUser.display_name}
                          </Link>
                          {fUser.is_verified && (
                            <VerifiedBadge type={fUser.verified_type} />
                          )}
                        </div>
                        <div className="flex items-center gap-1.5 text-[14px] text-muted">
                          <span>@{fUser.username}</span>
                        </div>
                        {fUser.bio && (
                          <p className="text-[14px] text-foreground mt-1 whitespace-pre-wrap break-words leading-snug">
                            {fUser.bio}
                          </p>
                        )}
                      </div>
                    </div>

                    {/* Follow / Unfollow Button (Standard X Style) */}
                    {!isSelf && (
                      <button
                        onClick={() => handleToggleFollow(fUser)}
                        onMouseEnter={() => setHoveredUserId(fUser.id)}
                        onMouseLeave={() => setHoveredUserId(null)}
                        className={cn(
                          "rounded-full font-bold text-[14px] px-4 py-1.5 transition-colors shrink-0",
                          fUser.is_following
                            ? isHovered
                              ? "bg-red-500/10 text-red-500 border border-red-500/30"
                              : "border border-border text-foreground hover:bg-subtle"
                            : "bg-foreground text-background hover:bg-foreground/90"
                        )}
                      >
                        {fUser.is_following
                          ? isHovered
                            ? "Unfollow"
                            : "Following"
                          : "Follow"}
                      </button>
                    )}
                  </div>
                );
              })}
            </div>
          )}
        </main>

        <RightSidebar />

        <AuthModal
          isOpen={authModalOpen}
          onClose={() => setAuthModalOpen(false)}
          onSuccess={(u) => setCurrentUser(u)}
        />
    </>
  );
}
