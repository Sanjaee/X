"use client";

import React, { useEffect, useState } from "react";
import { LeftNavRail } from "@/components/layout/LeftNavRail";
import { RightSidebar } from "@/components/layout/RightSidebar";
import { PostComposer } from "@/components/composer/PostComposer";
import { PostCard } from "@/components/feed/PostCard";
import { AuthModal } from "@/components/auth/AuthModal";
import { User, Post } from "@/types";
import { apiRequest, setStoredToken } from "@/lib/api";
import { cn } from "@/lib/utils";
import { Loader2 } from "lucide-react";

export default function HomePage() {
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [activeTab, setActiveTab] = useState<"for_you" | "following">("for_you");
  const [posts, setPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(true);
  const [authModalOpen, setAuthModalOpen] = useState(false);

  // Restore authenticated session
  useEffect(() => {
    const syncUser = () => {
      apiRequest<User>("/me")
        .then((user) => setCurrentUser(user))
        .catch(() => setCurrentUser(null));
    };
    syncUser();

    window.addEventListener("x_auth_state_change", syncUser);
    return () => window.removeEventListener("x_auth_state_change", syncUser);
  }, []);

  // Fetch timeline feed
  const loadFeed = async (tab: "for_you" | "following") => {
    setLoading(true);
    try {
      const res = await apiRequest<{ posts: Post[] }>(`/feed?type=${tab}`);
      setPosts(res.posts || []);
    } catch (err) {
      console.error("Failed to load feed", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadFeed(activeTab);
  }, [activeTab]);

  const handlePostCreated = (newPost: Post) => {
    setPosts((prev) => [newPost, ...prev]);
  };

  const handleLogout = async () => {
    try {
      await apiRequest("/auth/logout", { method: "POST" });
    } catch (e) {}
    try {
      const { signOut } = await import("next-auth/react");
      await signOut({ redirect: false });
    } catch (e) {}
    setStoredToken(null);
    setCurrentUser(null);
  };

  const handleLike = async (postId: string) => {
    if (!currentUser) {
      setAuthModalOpen(true);
      return;
    }
    try {
      await apiRequest(`/posts/${postId}/likes`, { method: "POST" });
    } catch (e) {
      console.error(e);
    }
  };

  const handleRepost = async (postId: string) => {
    if (!currentUser) {
      setAuthModalOpen(true);
      return;
    }
    try {
      await apiRequest(`/posts/${postId}/reposts`, { method: "POST" });
    } catch (e) {
      console.error(e);
    }
  };

  const handleBookmark = async (postId: string) => {
    if (!currentUser) {
      setAuthModalOpen(true);
      return;
    }
    try {
      await apiRequest(`/posts/${postId}/bookmarks`, { method: "POST" });
    } catch (e) {
      console.error(e);
    }
  };

  return (
    <div className="min-h-screen bg-background text-foreground flex justify-center">
      <div className="w-full max-w-[1280px] flex">
        {/* Left Navigation Rail (88px / 275px) */}
        <LeftNavRail
          currentUser={currentUser}
          onOpenComposer={() => window.scrollTo({ top: 0, behavior: "smooth" })}
          onOpenAuth={() => setAuthModalOpen(true)}
          onLogout={handleLogout}
        />

        {/* Center Feed Timeline (600px max) */}
        <main className="w-full sm:w-[600px] border-r border-border min-h-screen flex flex-col">
          {/* Sticky Header with Feed Tabs */}
          <header className="sticky top-0 bg-background/80 backdrop-blur-md border-b border-border z-20 flex flex-col">
            <div className="flex h-[53px] items-center justify-between px-4 sm:hidden">
              <span className="font-extrabold text-[18px]">Home</span>
            </div>

            <div className="flex w-full h-[53px]">
              <button
                onClick={() => setActiveTab("for_you")}
                className="flex-1 flex justify-center items-center hover:bg-subtle transition-colors relative"
              >
                <span
                  className={cn(
                    "text-[15px] font-bold py-3.5 relative",
                    activeTab === "for_you" ? "text-foreground" : "text-muted"
                  )}
                >
                  For you
                  {activeTab === "for_you" && (
                    <div className="absolute bottom-0 left-0 right-0 h-1 bg-brand rounded-full" />
                  )}
                </span>
              </button>

              <button
                onClick={() => {
                  if (!currentUser) {
                    setAuthModalOpen(true);
                    return;
                  }
                  setActiveTab("following");
                }}
                className="flex-1 flex justify-center items-center hover:bg-subtle transition-colors relative"
              >
                <span
                  className={cn(
                    "text-[15px] font-bold py-3.5 relative",
                    activeTab === "following" ? "text-foreground" : "text-muted"
                  )}
                >
                  Following
                  {activeTab === "following" && (
                    <div className="absolute bottom-0 left-0 right-0 h-1 bg-brand rounded-full" />
                  )}
                </span>
              </button>
            </div>
          </header>

          {/* Post Composer */}
          <PostComposer
            currentUser={currentUser}
            onPostSuccess={handlePostCreated}
            onOpenAuth={() => setAuthModalOpen(true)}
          />

          {/* Feed List */}
          <div className="flex-1 flex flex-col">
            {loading ? (
              <div className="flex justify-center items-center p-12 text-brand">
                <Loader2 className="w-8 h-8 animate-spin" />
              </div>
            ) : posts.length === 0 ? (
              <div className="text-center py-16 px-6">
                <h3 className="font-bold text-[22px] mb-2">Welcome to X!</h3>
                <p className="text-muted text-[15px] max-w-sm mx-auto mb-6">
                  This is the best place to see what’s happening in your world. Find people and topics to follow now.
                </p>
                {!currentUser && (
                  <button
                    onClick={() => setAuthModalOpen(true)}
                    className="bg-brand hover:bg-brand-hover text-white font-bold py-2.5 px-6 rounded-full text-[15px]"
                  >
                    Sign in to customize
                  </button>
                )}
              </div>
            ) : (
              posts.map((post) => (
                <PostCard
                  key={post.id}
                  post={post}
                  onLike={handleLike}
                  onRepost={handleRepost}
                  onBookmark={handleBookmark}
                />
              ))
            )}
          </div>
        </main>

        {/* Right Sidebar (420px max) */}
        <RightSidebar onSearch={(q) => alert(`Searching for: ${q}`)} />

        {/* Auth Modal */}
        <AuthModal
          isOpen={authModalOpen}
          onClose={() => setAuthModalOpen(false)}
          onSuccess={(user) => {
            setCurrentUser(user);
            loadFeed(activeTab);
          }}
        />
      </div>
    </div>
  );
}
