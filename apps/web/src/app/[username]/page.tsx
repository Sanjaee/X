"use client";

import React, { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import {
  ArrowLeft,
  Calendar,
  MapPin,
  Link as LinkIcon,
  Loader2,
  Mail,
  MoreHorizontal,
  Lock,
  VolumeX,
  Ban,
} from "lucide-react";
import { LeftNavRail } from "@/components/layout/LeftNavRail";
import { RightSidebar } from "@/components/layout/RightSidebar";
import { PostCard } from "@/components/feed/PostCard";
import { AuthModal } from "@/components/auth/AuthModal";
import { VerifiedBadge } from "@/components/common/VerifiedBadge";
import { User, Post, SocialRelationship } from "@/types";
import { apiRequest } from "@/lib/api";
import { formatNumber, cn } from "@/lib/utils";

export default function UserProfilePage() {
  const params = useParams();
  const router = useRouter();
  const username = params.username as string;

  const [profileUser, setProfileUser] = useState<User | null>(null);
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [relationship, setRelationship] = useState<SocialRelationship | null>(null);
  const [mutuals, setMutuals] = useState<User[]>([]);
  const [posts, setPosts] = useState<Post[]>([]);
  const [activeTab, setActiveTab] = useState<"posts" | "replies" | "media" | "likes">("posts");
  const [loading, setLoading] = useState(true);
  const [authModalOpen, setAuthModalOpen] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const [followersCount, setFollowersCount] = useState(0);

  useEffect(() => {
    apiRequest<User>("/me")
      .then((user) => setCurrentUser(user))
      .catch(() => setCurrentUser(null));
  }, []);

  useEffect(() => {
    if (!username) return;
    setLoading(true);

    Promise.all([
      apiRequest<User>(`/users/${username}`),
      apiRequest<{ posts: Post[] }>(`/users/${username}/posts?tab=${activeTab}`),
    ])
      .then(([user, postsRes]) => {
        setProfileUser(user);
        setFollowersCount(user.followers_count || 0);
        setPosts(postsRes.posts || []);
      })
      .catch((err) => {
        console.error("Failed to load profile", err);
      })
      .finally(() => {
        setLoading(false);
      });
  }, [username, activeTab]);

  // Load relationship and mutual connections if authenticated
  useEffect(() => {
    if (currentUser && profileUser && currentUser.id !== profileUser.id) {
      apiRequest<SocialRelationship>(`/users/${username}/relationship`)
        .then((rel) => setRelationship(rel))
        .catch(() => {});

      apiRequest<{ mutuals: User[] }>(`/users/${username}/mutuals`)
        .then((res) => setMutuals(res.mutuals || []))
        .catch(() => {});
    }
  }, [currentUser, profileUser, username]);

  const handleFollowAction = async () => {
    if (!currentUser) {
      setAuthModalOpen(true);
      return;
    }

    if (!relationship) return;

    // Case 1: Cancel pending follow request
    if (relationship.has_pending_follow_request) {
      try {
        await apiRequest(`/users/${username}/follow-request`, { method: "DELETE" });
        setRelationship((prev) => prev ? { ...prev, has_pending_follow_request: false } : null);
      } catch (e) {
        console.error("Failed to cancel request", e);
      }
      return;
    }

    // Case 2: Request follow to private account
    if (relationship.is_private && !relationship.is_following) {
      try {
        await apiRequest(`/users/${username}/follow-request`, { method: "POST" });
        setRelationship((prev) => prev ? { ...prev, has_pending_follow_request: true } : null);
      } catch (e) {
        console.error("Failed to submit follow request", e);
      }
      return;
    }

    // Case 3: Public follow / unfollow
    const prevFollowing = relationship.is_following;
    const prevCount = followersCount;

    setRelationship((prev) => prev ? { ...prev, is_following: !prevFollowing } : null);
    setFollowersCount(prevFollowing ? Math.max(0, prevCount - 1) : prevCount + 1);

    try {
      const res = await apiRequest<{ is_following: boolean; followers_count: number }>(
        `/users/${username}/follow`,
        { method: prevFollowing ? "DELETE" : "POST" }
      );
      setRelationship((prev) => prev ? { ...prev, is_following: res.is_following } : null);
      setFollowersCount(res.followers_count);
    } catch (e) {
      setRelationship((prev) => prev ? { ...prev, is_following: prevFollowing } : null);
      setFollowersCount(prevCount);
    }
  };

  const handleBlockToggle = async () => {
    if (!relationship) return;
    const isBlocked = relationship.is_blocked;
    try {
      await apiRequest(`/users/${username}/block`, { method: isBlocked ? "DELETE" : "POST" });
      setRelationship((prev) => prev ? { ...prev, is_blocked: !isBlocked, is_following: false } : null);
      setMenuOpen(false);
    } catch (e) {
      console.error("Block action failed", e);
    }
  };

  const handleMuteToggle = async () => {
    if (!relationship) return;
    const isMuted = relationship.is_muted;
    try {
      await apiRequest(`/users/${username}/mute`, { method: isMuted ? "DELETE" : "POST" });
      setRelationship((prev) => prev ? { ...prev, is_muted: !isMuted } : null);
      setMenuOpen(false);
    } catch (e) {
      console.error("Mute action failed", e);
    }
  };

  const isSelf = currentUser && profileUser && currentUser.id === profileUser.id;

  return (
    <div className="min-h-screen bg-background text-foreground flex justify-center">
      <div className="w-full max-w-[1280px] flex">
        <LeftNavRail
          currentUser={currentUser}
          onOpenAuth={() => setAuthModalOpen(true)}
        />

        <main className="w-full sm:w-[600px] border-r border-border min-h-screen flex flex-col">
          {/* Header */}
          <header className="sticky top-0 bg-background/80 backdrop-blur-md border-b border-border z-20 flex items-center px-4 h-[53px] gap-6">
            <button
              onClick={() => router.back()}
              className="p-2 hover:bg-subtle rounded-full transition-colors text-foreground"
            >
              <ArrowLeft className="w-5 h-5" />
            </button>
            <div className="flex flex-col">
              <h1 className="font-extrabold text-[19px] leading-tight flex items-center gap-1.5">
                <span>{profileUser?.display_name || username}</span>
                {profileUser?.is_verified && (
                  <VerifiedBadge type={profileUser.verified_type} size="sm" />
                )}
              </h1>
              <span className="text-xs text-muted">
                {profileUser ? `${formatNumber(profileUser.posts_count || 0)} posts` : ""}
              </span>
            </div>
          </header>

          {loading && !profileUser ? (
            <div className="flex justify-center items-center p-16 text-brand">
              <Loader2 className="w-8 h-8 animate-spin" />
            </div>
          ) : !profileUser ? (
            <div className="text-center py-20 px-6">
              <h2 className="text-2xl font-bold mb-2">This account doesn’t exist</h2>
              <p className="text-muted">Try searching for another.</p>
            </div>
          ) : (
            <div className="flex flex-col">
              {/* Banner */}
              <div className="w-full h-[200px] bg-muted/20 relative">
                {profileUser.banner_url ? (
                  <img
                    src={profileUser.banner_url}
                    alt="Banner"
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="w-full h-full bg-gradient-to-r from-neutral-800 to-neutral-900" />
                )}
              </div>

              {/* Profile Details Container */}
              <div className="px-4 pb-4 border-b border-border">
                {/* Avatar and Action button row */}
                <div className="flex justify-between items-end relative -mt-[68px] mb-3">
                  <img
                    src={
                      profileUser.avatar_url ||
                      "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"
                    }
                    alt={profileUser.display_name}
                    className="w-[136px] h-[136px] rounded-full object-cover border-4 border-background bg-background shrink-0"
                  />

                  {isSelf ? (
                    <button className="border border-border hover:bg-subtle font-bold text-[15px] px-4 py-1.5 rounded-full transition-colors mb-2">
                      Edit profile
                    </button>
                  ) : (
                    <div className="flex items-center gap-2 mb-2 relative">
                      {/* More actions dropdown */}
                      <div className="relative">
                        <button
                          onClick={() => setMenuOpen(!menuOpen)}
                          className="p-2 border border-border hover:bg-subtle rounded-full text-foreground transition-colors"
                        >
                          <MoreHorizontal className="w-5 h-5" />
                        </button>
                        {menuOpen && (
                          <div className="absolute right-0 top-11 z-30 bg-card border border-border rounded-2xl shadow-2xl py-2 w-48 text-sm">
                            <button
                              onClick={handleMuteToggle}
                              className="w-full px-4 py-2 text-left hover:bg-subtle flex items-center gap-2"
                            >
                              <VolumeX className="w-4 h-4" />
                              <span>{relationship?.is_muted ? `Unmute @${username}` : `Mute @${username}`}</span>
                            </button>
                            <button
                              onClick={handleBlockToggle}
                              className="w-full px-4 py-2 text-left hover:bg-subtle flex items-center gap-2 text-red-500 font-medium"
                            >
                              <Ban className="w-4 h-4" />
                              <span>{relationship?.is_blocked ? `Unblock @${username}` : `Block @${username}`}</span>
                            </button>
                          </div>
                        )}
                      </div>

                      {/* Direct Message button */}
                      <Link
                        href={`/messages`}
                        className="p-2 border border-border hover:bg-subtle rounded-full text-foreground transition-colors"
                        title="Send Message"
                      >
                        <Mail className="w-5 h-5" />
                      </Link>

                      {/* Follow / Following / Requested button */}
                      {relationship?.is_blocked ? (
                        <button
                          onClick={handleBlockToggle}
                          className="bg-red-600 text-white font-bold text-[15px] px-5 py-1.5 rounded-full hover:bg-red-700 transition-colors"
                        >
                          Blocked
                        </button>
                      ) : (
                        <button
                          onClick={handleFollowAction}
                          className={cn(
                            "font-bold text-[15px] px-5 py-1.5 rounded-full transition-colors",
                            relationship?.is_following
                              ? "border border-border text-foreground hover:border-red-600 hover:text-red-600 hover:bg-red-600/10"
                              : relationship?.has_pending_follow_request
                              ? "border border-border text-muted hover:border-red-600 hover:text-red-600"
                              : "bg-foreground text-background hover:opacity-90"
                          )}
                        >
                          {relationship?.is_following
                            ? "Following"
                            : relationship?.has_pending_follow_request
                            ? "Requested"
                            : "Follow"}
                        </button>
                      )}
                    </div>
                  )}
                </div>

                {/* Display Name and Handle */}
                <div className="flex items-center gap-1.5 font-extrabold text-[20px] leading-6">
                  <span>{profileUser.display_name}</span>
                  {profileUser.is_verified && (
                    <VerifiedBadge type={profileUser.verified_type} size="md" />
                  )}
                </div>

                <div className="flex items-center gap-2 mb-3">
                  <span className="text-muted text-[15px]">@{profileUser.username}</span>
                  {relationship?.is_followed_by && (
                    <span className="bg-subtle text-muted text-xs px-1.5 py-0.5 rounded font-normal">
                      Follows you
                    </span>
                  )}
                </div>

                {/* Bio */}
                {profileUser.bio && (
                  <p className="text-[15px] leading-5 text-foreground mb-3 whitespace-pre-wrap">
                    {profileUser.bio}
                  </p>
                )}

                {/* Meta details */}
                <div className="flex items-center gap-x-4 gap-y-1 flex-wrap text-[15px] text-muted mb-3">
                  {profileUser.location && (
                    <div className="flex items-center gap-1">
                      <MapPin className="w-4 h-4 shrink-0" />
                      <span>{profileUser.location}</span>
                    </div>
                  )}
                  {profileUser.website && (
                    <div className="flex items-center gap-1">
                      <LinkIcon className="w-4 h-4 shrink-0" />
                      <a
                        href={profileUser.website}
                        target="_blank"
                        rel="noreferrer"
                        className="text-brand hover:underline"
                      >
                        {profileUser.website.replace(/^https?:\/\//, "")}
                      </a>
                    </div>
                  )}
                  <div className="flex items-center gap-1">
                    <Calendar className="w-4 h-4 shrink-0" />
                    <span>
                      Joined{" "}
                      {new Date(profileUser.created_at).toLocaleDateString("en-US", {
                        month: "long",
                        year: "numeric",
                      })}
                    </span>
                  </div>
                </div>

                {/* Social Counters */}
                <div className="flex items-center gap-5 text-[14px] mb-3">
                  <Link href={`/${username}/following`} className="hover:underline">
                    <span className="font-bold text-foreground">
                      {formatNumber(profileUser.following_count)}
                    </span>{" "}
                    <span className="text-muted">Following</span>
                  </Link>
                  <Link href={`/${username}/followers`} className="hover:underline">
                    <span className="font-bold text-foreground">
                      {formatNumber(followersCount)}
                    </span>{" "}
                    <span className="text-muted">Followers</span>
                  </Link>
                </div>

                {/* Mutual connections section */}
                {mutuals.length > 0 && (
                  <div className="flex items-center gap-2 text-xs text-muted pt-2 border-t border-border">
                    <div className="flex -space-x-1.5 overflow-hidden">
                      {mutuals.slice(0, 3).map((m) => (
                        <img
                          key={m.id}
                          src={m.avatar_url || "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"}
                          alt={m.display_name}
                          className="inline-block h-5 w-5 rounded-full ring-2 ring-background object-cover"
                        />
                      ))}
                    </div>
                    <span>
                      Followed by{" "}
                      <span className="text-foreground font-medium">
                        {mutuals.slice(0, 2).map((m) => m.display_name).join(", ")}
                      </span>
                      {mutuals.length > 2 ? ` and ${mutuals.length - 2} others you follow` : ""}
                    </span>
                  </div>
                )}
              </div>

              {/* Private account protection or blocked state banner */}
              {relationship?.is_blocked ? (
                <div className="text-center py-16 px-4">
                  <Ban className="w-12 h-12 text-muted mx-auto mb-2" />
                  <h3 className="font-bold text-lg mb-1">@{username} is blocked</h3>
                  <p className="text-muted text-sm max-w-sm mx-auto">
                    You won’t see their posts or receive notifications from them.
                  </p>
                </div>
              ) : relationship?.is_private && !relationship?.is_following && !isSelf ? (
                <div className="text-center py-16 px-4">
                  <Lock className="w-12 h-12 text-muted mx-auto mb-2" />
                  <h3 className="font-bold text-lg mb-1">These posts are protected</h3>
                  <p className="text-muted text-sm max-w-sm mx-auto">
                    Only confirmed followers have access to @{username}’s posts. Send a follow request to follow them.
                  </p>
                </div>
              ) : (
                <>
                  {/* Profile Sub-tabs */}
                  <div className="flex border-b border-border h-[53px]">
                    {(["posts", "replies", "media", "likes"] as const).map((tab) => (
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
                          {tab}
                          {activeTab === tab && (
                            <div className="absolute bottom-0 left-0 right-0 h-1 bg-brand rounded-full" />
                          )}
                        </span>
                      </button>
                    ))}
                  </div>

                  {/* Tab Content List */}
                  <div className="flex flex-col">
                    {posts.length === 0 ? (
                      <div className="text-center py-12 text-muted text-[15px]">
                        No {activeTab} yet.
                      </div>
                    ) : (
                      posts.map((post) => <PostCard key={post.id} post={post} />)
                    )}
                  </div>
                </>
              )}
            </div>
          )}
        </main>

        <RightSidebar onSearch={(q) => router.push(`/search?q=${encodeURIComponent(q)}`)} />

        <AuthModal
          isOpen={authModalOpen}
          onClose={() => setAuthModalOpen(false)}
          onSuccess={(user) => setCurrentUser(user)}
        />
      </div>
    </div>
  );
}
