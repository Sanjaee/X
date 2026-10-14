"use client";

import React, { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import {
  ArrowLeft,
  Calendar,
  MapPin,
  Link as LinkIcon,
  CheckCircle2,
  Loader2,
} from "lucide-react";
import { LeftNavRail } from "@/components/layout/LeftNavRail";
import { RightSidebar } from "@/components/layout/RightSidebar";
import { PostCard } from "@/components/feed/PostCard";
import { AuthModal } from "@/components/auth/AuthModal";
import { VerifiedBadge } from "@/components/common/VerifiedBadge";
import { User, Post } from "@/types";

import { apiRequest } from "@/lib/api";
import { formatNumber, cn } from "@/lib/utils";

export default function UserProfilePage() {
  const params = useParams();
  const router = useRouter();
  const username = params.username as string;

  const [profileUser, setProfileUser] = useState<User | null>(null);
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [posts, setPosts] = useState<Post[]>([]);
  const [activeTab, setActiveTab] = useState<"posts" | "replies" | "media" | "likes">("posts");
  const [loading, setLoading] = useState(true);
  const [authModalOpen, setAuthModalOpen] = useState(false);
  const [isFollowing, setIsFollowing] = useState(false);
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
        setIsFollowing(!!user.is_following);
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

  const handleFollowToggle = async () => {
    if (!currentUser) {
      setAuthModalOpen(true);
      return;
    }

    const previousStatus = isFollowing;
    const previousCount = followersCount;

    setIsFollowing(!previousStatus);
    setFollowersCount(previousStatus ? Math.max(0, previousCount - 1) : previousCount + 1);

    try {
      const res = await apiRequest<{ is_following: boolean; followers_count: number }>(
        `/users/${username}/follow`,
        { method: previousStatus ? "DELETE" : "POST" }
      );
      setIsFollowing(res.is_following);
      setFollowersCount(res.followers_count);
    } catch (err) {
      setIsFollowing(previousStatus);
      setFollowersCount(previousCount);
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
          <header className="sticky top-0 bg-background/80 backdrop-blur-md border-b border-border z-20 flex items-center gap-6 px-4 h-[53px]">
            <button
              onClick={() => router.back()}
              className="p-2 hover:bg-subtle rounded-full text-foreground transition-colors"
            >
              <ArrowLeft className="w-5 h-5" />
            </button>
            <div className="flex flex-col">
              <div className="flex items-center gap-1 font-extrabold text-[20px] leading-6">
                <span>{profileUser?.display_name || username}</span>
                {profileUser?.is_verified && (
                  <CheckCircle2 className="w-4 h-4 fill-brand text-background inline" />
                )}
              </div>
              <span className="text-[13px] text-muted leading-4">
                {profileUser?.posts_count || 0} posts
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
                    <button
                      onClick={handleFollowToggle}
                      className={cn(
                        "font-bold text-[15px] px-5 py-1.5 rounded-full transition-colors mb-2",
                        isFollowing
                          ? "border border-border text-foreground hover:border-red-600 hover:text-red-600 hover:bg-red-600/10"
                          : "bg-foreground text-background hover:opacity-90"
                      )}
                    >
                      {isFollowing ? "Following" : "Follow"}
                    </button>
                  )}
                </div>

                {/* Display Name and Handle */}
                <div className="flex items-center gap-1.5 font-extrabold text-[20px] leading-6">
                  <span>{profileUser.display_name}</span>
                  {profileUser.is_verified && (
                    <VerifiedBadge type={profileUser.verified_type} size="md" />
                  )}
                </div>

                <span className="text-muted text-[15px] block mb-3">@{profileUser.username}</span>

                {/* Bio */}
                {profileUser.bio && (
                  <p className="text-[15px] leading-5 text-foreground mb-3 whitespace-pre-wrap">
                    {profileUser.bio}
                  </p>
                )}

                {/* Meta details (location, website, joined date) */}
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
                <div className="flex items-center gap-5 text-[14px]">
                  <Link href="#" className="hover:underline">
                    <span className="font-bold text-foreground">
                      {formatNumber(profileUser.following_count)}
                    </span>{" "}
                    <span className="text-muted">Following</span>
                  </Link>
                  <Link href="#" className="hover:underline">
                    <span className="font-bold text-foreground">
                      {formatNumber(followersCount)}
                    </span>{" "}
                    <span className="text-muted">Followers</span>
                  </Link>
                </div>
              </div>

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
