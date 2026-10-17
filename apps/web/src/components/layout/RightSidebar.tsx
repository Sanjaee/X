"use client";

import React, { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { MoreHorizontal, CheckCircle2, X, Phone } from "lucide-react";
import { formatNumber } from "@/lib/utils";
import { SearchBarDropdown } from "@/components/search/SearchBarDropdown";
import { apiRequest } from "@/lib/api";
import { TrendingTopic, UserRecommendation, User } from "@/types";
import { AuthModal } from "@/components/auth/AuthModal";

interface RightSidebarProps {
  onSearch?: (query: string) => void;
  hideSearchBar?: boolean;
  currentUser?: User | null;
  relevantUser?: User | null;
  onOpenAuth?: () => void;
}

export function RightSidebar({
  onSearch,
  hideSearchBar = false,
  currentUser: propUser,
  relevantUser,
  onOpenAuth,
}: RightSidebarProps) {
  const router = useRouter();
  const [user, setUser] = useState<User | null | undefined>(propUser);
  const [localAuthOpen, setLocalAuthOpen] = useState(false);
  const [trends, setTrends] = useState<TrendingTopic[]>([]);
  const [recommendations, setRecommendations] = useState<UserRecommendation[]>([]);
  const [followingMap, setFollowingMap] = useState<Record<string, boolean>>({});
  const [isLoadingTrends, setIsLoadingTrends] = useState(true);
  const [isLoadingRecs, setIsLoadingRecs] = useState(true);

  const handleTriggerAuth = () => {
    if (onOpenAuth) {
      onOpenAuth();
    } else {
      setLocalAuthOpen(true);
    }
  };

  // Sync or fetch currentUser
  useEffect(() => {
    let isMounted = true;

    // If parent passed an authenticated user, adopt it immediately
    if (propUser) {
      setUser(propUser);
      return;
    }

    // Verify session from /me (supports cookies and localStorage tokens)
    const syncUser = () => {
      apiRequest<User>("/me")
        .then((u) => {
          if (isMounted) setUser(u);
        })
        .catch(() => {
          if (isMounted) setUser(null);
        });
    };

    syncUser();
    window.addEventListener("x_auth_state_change", syncUser);

    return () => {
      isMounted = false;
      window.removeEventListener("x_auth_state_change", syncUser);
    };
  }, [propUser]);

  // Load genuine trending topics & personalized recommendations
  useEffect(() => {
    let isMounted = true;

    // Fetch Trends
    apiRequest<{ trending: TrendingTopic[] }>("/explore/trending?limit=5")
      .then((res) => {
        if (isMounted && res.trending) {
          setTrends(res.trending);
        }
      })
      .catch((err) => console.error("Failed to load trending topics", err))
      .finally(() => {
        if (isMounted) setIsLoadingTrends(false);
      });

    // Fetch Who to Follow
    apiRequest<{ recommendations: UserRecommendation[] }>("/users/recommendations?limit=4")
      .then((res) => {
        if (isMounted && res.recommendations) {
          setRecommendations(res.recommendations);
          const initialFollowing: Record<string, boolean> = {};
          res.recommendations.forEach((r) => {
            if (r.user.is_following) {
              initialFollowing[r.user.id] = true;
            }
          });
          setFollowingMap(initialFollowing);
        }
      })
      .catch((err) => console.error("Failed to load account recommendations", err))
      .finally(() => {
        if (isMounted) setIsLoadingRecs(false);
      });

    return () => {
      isMounted = false;
    };
  }, []);

  const handleToggleFollow = async (e: React.MouseEvent, rec: UserRecommendation) => {
    e.stopPropagation();
    if (!user) {
      if (onOpenAuth) onOpenAuth();
      return;
    }
    const isCurrentlyFollowing = !!followingMap[rec.user.id];

    // Optimistic toggle
    setFollowingMap((prev) => ({ ...prev, [rec.user.id]: !isCurrentlyFollowing }));

    try {
      if (isCurrentlyFollowing) {
        await apiRequest(`/users/${rec.user.username}/follow`, { method: "DELETE" });
      } else {
        await apiRequest(`/users/${rec.user.username}/follow`, { method: "POST" });
      }
    } catch (err) {
      // Rollback on failure
      setFollowingMap((prev) => ({ ...prev, [rec.user.id]: isCurrentlyFollowing }));
      console.error("Failed to update follow state", err);
    }
  };

  const handleDismiss = async (e: React.MouseEvent, targetUserID: string) => {
    e.stopPropagation();
    // Remove from UI immediately
    setRecommendations((prev) => prev.filter((r) => r.user.id !== targetUserID));

    try {
      await apiRequest(`/users/recommendations/${targetUserID}/dismiss`, { method: "POST" });
    } catch (err) {
      console.error("Failed to record recommendation dismissal", err);
    }
  };

  // If user is explicitly not logged in (null) AND on a post detail page with relevantUser (matching reference screenshot)
  if (user === null && !!relevantUser) {
    const focal = relevantUser;

    return (
      <aside className="hidden lg:block w-[350px] xl:w-[420px] px-6 py-3 select-none">
        <div className="sticky top-3 space-y-4 pb-12">
          {/* Top Auth Card: Log in or sign up for X */}
          <div className="bg-transparent rounded-2xl p-4 border border-border">
            <h2 className="font-extrabold text-[20px] leading-6 mb-1 text-foreground">
              Log in or sign up for X
            </h2>
            <p className="text-[15px] leading-5 text-muted mb-4 font-normal">
              See what’s happening and join the conversation
            </p>

            <div className="space-y-3">
              {/* Continue with phone */}
              <button
                onClick={handleTriggerAuth}
                className="w-full bg-white hover:bg-neutral-200 text-black font-bold py-2.5 px-4 rounded-full text-[15px] transition-colors flex items-center justify-center gap-2.5 shadow-sm"
              >
                <Phone className="w-4 h-4 text-black fill-current" />
                <span>Continue with phone</span>
              </button>

              {/* Continue with Apple */}
              <button
                onClick={handleTriggerAuth}
                className="w-full bg-black hover:bg-white/10 text-white font-bold py-2.5 px-4 rounded-full text-[15px] transition-colors border border-border flex items-center justify-center gap-2.5"
              >
                <svg className="w-4 h-4 fill-current" viewBox="0 0 170 170">
                  <path d="M150.37 130.25c-2.45 5.66-5.35 10.87-8.71 15.66-4.58 6.53-8.33 11.05-11.22 13.56-4.48 4.12-9.28 6.23-14.42 6.35-3.69 0-8.14-1.05-13.32-3.18-5.19-2.12-9.97-3.17-14.34-3.17-4.58 0-9.49 1.05-14.75 3.17-5.26 2.13-9.5 3.24-12.74 3.35-4.35.13-9.16-1.9-14.42-6.08-3.7-3.04-7.6-7.77-11.72-14.19-6.42-9.99-11.39-21.36-14.9-34.1-3.52-12.75-5.28-24.81-5.28-36.19 0-14.02 3.36-25.9 10.09-35.63 6.72-9.74 15.15-14.68 25.29-14.83 5.44 0 11.3 1.49 17.58 4.47 6.28 2.99 10.37 4.54 12.27 4.65 1.52-.11 5.88-1.74 13.08-4.9 7.2-3.15 13.1-4.63 17.71-4.44 14.13.76 25.12 5.99 32.96 15.69-12.61 7.6-18.72 18.04-18.33 31.31.32 10.21 4.23 18.79 11.73 25.75 7.5 6.96 16.41 10.74 26.74 11.36-2.18 6.53-4.8 12.87-7.85 19.03zM119.22 31.84c0-7.72 2.76-14.9 8.27-21.56 5.51-6.65 12.28-10.28 20.31-10.88.22 1.09.33 2.18.33 3.27 0 7.61-2.82 14.89-8.47 21.84-5.65 6.96-12.49 10.74-20.52 11.34-.11-1.3-.2-2.6-.2-4.01z" />
                </svg>
                <span>Continue with Apple</span>
              </button>

              {/* Continue with Google */}
              <button
                onClick={handleTriggerAuth}
                className="w-full bg-black hover:bg-white/10 text-white font-bold py-2.5 px-4 rounded-full text-[15px] transition-colors border border-border flex items-center justify-center gap-2.5"
              >
                <svg className="w-4 h-4" viewBox="0 0 24 24">
                  <path
                    fill="#4285F4"
                    d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                  />
                  <path
                    fill="#34A853"
                    d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                  />
                  <path
                    fill="#FBBC05"
                    d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.06H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.94l2.85-2.22.81-.63z"
                  />
                  <path
                    fill="#EA4335"
                    d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.06l3.66 2.84c.87-2.6 3.3-4.52 6.16-4.52z"
                  />
                </svg>
                <span>Continue with Google</span>
              </button>

              {/* Divider: or */}
              <div className="flex items-center my-3">
                <div className="flex-1 border-t border-border" />
                <span className="px-3 text-muted text-[13px] font-normal">or</span>
                <div className="flex-1 border-t border-border" />
              </div>

              {/* Log in with username or email */}
              <button
                onClick={handleTriggerAuth}
                className="w-full text-foreground hover:bg-subtle font-bold py-2 px-4 rounded-full text-[14px] transition-colors flex items-center justify-center gap-2"
              >
                <span className="text-[15px] font-medium text-muted">@</span>
                <span>Log in with username or email</span>
              </button>
            </div>
          </div>

          {/* Relevant people */}
          <div className="bg-transparent rounded-2xl border border-border overflow-hidden">
            <h2 className="font-extrabold text-[20px] leading-6 px-4 pt-3 pb-2 text-foreground">
              Relevant people
            </h2>

            {focal ? (
              <div className="px-4 py-3 flex items-start justify-between gap-3">
                <div className="flex items-start gap-3 min-w-0 flex-1">
                  <Link href={`/${focal.username}`}>
                    <img
                      src={
                        focal.avatar_url ||
                        `https://ui-avatars.com/api/?name=${encodeURIComponent(focal.display_name)}&background=random`
                      }
                      alt={focal.display_name}
                      className="w-10 h-10 rounded-full object-cover border border-border shrink-0 hover:opacity-90"
                    />
                  </Link>
                  <div className="flex flex-col min-w-0 flex-1">
                    <Link
                      href={`/${focal.username}`}
                      className="font-bold text-[15px] leading-5 text-foreground hover:underline truncate flex items-center gap-1"
                    >
                      <span>{focal.display_name}</span>
                      {focal.is_verified && (
                        <CheckCircle2 className="w-4 h-4 fill-amber-400 text-black shrink-0 inline" />
                      )}
                    </Link>
                    <span className="text-[14px] text-muted leading-4 truncate">
                      @{focal.username}
                    </span>
                    {focal.bio && (
                      <p className="text-[14px] leading-5 text-foreground mt-2 line-clamp-3">
                        {focal.bio}
                      </p>
                    )}
                  </div>
                </div>
                <button
                  onClick={handleTriggerAuth}
                  className="bg-white hover:bg-neutral-200 text-black font-bold text-[14px] py-1.5 px-4 rounded-full shrink-0 transition-colors shadow-sm"
                >
                  Follow
                </button>
              </div>
            ) : (
              <div className="px-4 py-4 text-sm text-muted">No relevant profiles available.</div>
            )}
          </div>

          {/* Legal Footer */}
          <footer className="text-[13px] text-muted px-4 leading-5 flex flex-wrap gap-x-2 gap-y-1">
            <Link href="#" className="hover:underline">Terms</Link>
            <span>·</span>
            <Link href="#" className="hover:underline">Privacy</Link>
            <span>·</span>
            <Link href="#" className="hover:underline">Cookies</Link>
            <span>·</span>
            <Link href="#" className="hover:underline">Accessibility</Link>
            <span>·</span>
            <Link href="#" className="hover:underline">Ads Info</Link>
            <span>·</span>
            <span className="hover:underline cursor-pointer">More ...</span>
            <span>© 2026 X Corp.</span>
          </footer>
        </div>

        <AuthModal
          isOpen={localAuthOpen}
          onClose={() => setLocalAuthOpen(false)}
          onSuccess={(u) => setUser(u)}
        />
      </aside>
    );
  }

  return (
    <aside className="hidden lg:block w-[350px] xl:w-[420px] px-6 py-3 select-none">
      {/* Search Bar with Autocomplete Dropdown */}
      {!hideSearchBar && (
        <div className="sticky top-0 bg-background/80 backdrop-blur-md pt-1 pb-3 z-20">
          <SearchBarDropdown onSearchSubmit={onSearch} />
        </div>
      )}

      {/* Top Card: Subscribe to Premium (when logged in) OR Log in or sign up for X (when unauthenticated) */}
      {user ? (
        <div className="bg-transparent rounded-2xl p-4 border border-border mb-4">
          <h2 className="font-extrabold text-[20px] leading-6 mb-1 text-foreground">Subscribe to Premium</h2>
          <p className="text-[15px] leading-5 text-foreground mb-3 font-normal">
            Subscribe to unlock new features and if eligible, receive a creator revenue share.
          </p>
          <Link
            href="/premium"
            className="inline-block bg-brand hover:bg-brand-hover text-white font-bold py-2 px-5 rounded-full text-[15px] transition-colors"
          >
            Subscribe
          </Link>
        </div>
      ) : user === null ? (
        <div className="bg-transparent rounded-2xl p-4 border border-border mb-4">
          <h2 className="font-extrabold text-[20px] leading-6 mb-1 text-foreground">
            Log in or sign up for X
          </h2>
          <p className="text-[15px] leading-5 text-muted mb-4 font-normal">
            See what’s happening and join the conversation
          </p>

          <div className="space-y-3">
            <button
              onClick={handleTriggerAuth}
              className="w-full bg-white hover:bg-neutral-200 text-black font-bold py-2.5 px-4 rounded-full text-[15px] transition-colors flex items-center justify-center gap-2.5 shadow-sm"
            >
              <Phone className="w-4 h-4 text-black fill-current" />
              <span>Continue with phone</span>
            </button>

            <button
              onClick={handleTriggerAuth}
              className="w-full bg-black hover:bg-white/10 text-white font-bold py-2.5 px-4 rounded-full text-[15px] transition-colors border border-border flex items-center justify-center gap-2.5"
            >
              <svg className="w-4 h-4 fill-current" viewBox="0 0 170 170">
                <path d="M150.37 130.25c-2.45 5.66-5.35 10.87-8.71 15.66-4.58 6.53-8.33 11.05-11.22 13.56-4.48 4.12-9.28 6.23-14.42 6.35-3.69 0-8.14-1.05-13.32-3.18-5.19-2.12-9.97-3.17-14.34-3.17-4.58 0-9.49 1.05-14.75 3.17-5.26 2.13-9.5 3.24-12.74 3.35-4.35.13-9.16-1.9-14.42-6.08-3.7-3.04-7.6-7.77-11.72-14.19-6.42-9.99-11.39-21.36-14.9-34.1-3.52-12.75-5.28-24.81-5.28-36.19 0-14.02 3.36-25.9 10.09-35.63 6.72-9.74 15.15-14.68 25.29-14.83 5.44 0 11.3 1.49 17.58 4.47 6.28 2.99 10.37 4.54 12.27 4.65 1.52-.11 5.88-1.74 13.08-4.9 7.2-3.15 13.1-4.63 17.71-4.44 14.13.76 25.12 5.99 32.96 15.69-12.61 7.6-18.72 18.04-18.33 31.31.32 10.21 4.23 18.79 11.73 25.75 7.5 6.96 16.41 10.74 26.74 11.36-2.18 6.53-4.8 12.87-7.85 19.03zM119.22 31.84c0-7.72 2.76-14.9 8.27-21.56 5.51-6.65 12.28-10.28 20.31-10.88.22 1.09.33 2.18.33 3.27 0 7.61-2.82 14.89-8.47 21.84-5.65 6.96-12.49 10.74-20.52 11.34-.11-1.3-.2-2.6-.2-4.01z" />
              </svg>
              <span>Continue with Apple</span>
            </button>

            <button
              onClick={handleTriggerAuth}
              className="w-full bg-black hover:bg-white/10 text-white font-bold py-2.5 px-4 rounded-full text-[15px] transition-colors border border-border flex items-center justify-center gap-2.5"
            >
              <svg className="w-4 h-4" viewBox="0 0 24 24">
                <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" />
                <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" />
                <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.06H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.94l2.85-2.22.81-.63z" />
                <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.06l3.66 2.84c.87-2.6 3.3-4.52 6.16-4.52z" />
              </svg>
              <span>Continue with Google</span>
            </button>

            <div className="flex items-center my-3">
              <div className="flex-1 border-t border-border" />
              <span className="px-3 text-muted text-[13px] font-normal">or</span>
              <div className="flex-1 border-t border-border" />
            </div>

            <button
              onClick={handleTriggerAuth}
              className="w-full text-foreground hover:bg-subtle font-bold py-2 px-4 rounded-full text-[14px] transition-colors flex items-center justify-center gap-2"
            >
              <span className="text-[15px] font-medium text-muted">@</span>
              <span>Log in with username or email</span>
            </button>
          </div>
        </div>
      ) : null}

      {/* Sticky Container for What's happening & Who to follow */}
      <div className="sticky top-16 space-y-4 pb-12">
        {/* What's Happening / Trends */}
        <div className="bg-transparent rounded-2xl border border-border overflow-hidden">
          <h2 className="font-extrabold text-[20px] leading-6 px-4 pt-3 pb-2 text-foreground">What's happening</h2>
          <div className="flex flex-col">
            {isLoadingTrends ? (
              <div className="px-4 py-6 text-sm text-muted animate-pulse">Loading trends...</div>
            ) : trends.length > 0 ? (
              trends.map((t) => (
                <div
                  key={t.topic}
                  onClick={() => router.push(`/search?q=${encodeURIComponent(t.topic)}`)}
                  className="px-4 py-3 hover:bg-subtle cursor-pointer transition-colors flex justify-between items-start"
                >
                  <div>
                    <span className="text-[13px] text-muted leading-4 block">{t.category}</span>
                    <span className="text-[15px] font-bold text-foreground leading-5 block mt-0.5">{t.topic}</span>
                    <span className="text-[13px] text-muted leading-4 block mt-0.5">
                      {formatNumber(t.posts_count)} posts
                    </span>
                  </div>
                  <button
                    aria-label="More options"
                    onClick={(e) => {
                      e.stopPropagation();
                      router.push(`/search?q=${encodeURIComponent(t.topic)}`);
                    }}
                    className="text-muted hover:text-foreground p-1 hover:bg-subtle rounded-full"
                  >
                    <MoreHorizontal className="w-4 h-4" />
                  </button>
                </div>
              ))
            ) : (
              <div className="px-4 py-4 text-sm text-muted">No trending topics right now.</div>
            )}
            <Link
              href="/explore"
              className="px-4 py-3 text-brand text-[15px] hover:bg-subtle transition-colors block font-medium"
            >
              Show more
            </Link>
          </div>
        </div>

        {/* Who to Follow */}
        <div className="bg-transparent rounded-2xl border border-border overflow-hidden">
          <h2 className="font-extrabold text-[20px] leading-6 px-4 pt-3 pb-2 text-foreground">Who to follow</h2>
          <div className="flex flex-col">
            {isLoadingRecs ? (
              <div className="px-4 py-6 text-sm text-muted animate-pulse">Loading suggestions...</div>
            ) : recommendations.length > 0 ? (
              recommendations.map((rec) => {
                const isFollowing = !!followingMap[rec.user.id];
                return (
                  <div
                    key={rec.user.id}
                    onClick={() => router.push(`/${rec.user.username}`)}
                    className="px-4 py-3 hover:bg-subtle cursor-pointer transition-colors flex items-center justify-between group"
                  >
                    <div className="flex items-center gap-3 min-w-0 pr-2">
                      <img
                        src={
                          rec.user.avatar_url ||
                          `https://ui-avatars.com/api/?name=${encodeURIComponent(rec.user.display_name)}&background=random`
                        }
                        alt={rec.user.display_name}
                        className="w-10 h-10 rounded-full object-cover border border-border shrink-0"
                      />
                      <div className="flex flex-col min-w-0">
                        <div className="flex items-center gap-1 min-w-0">
                          <span className="font-bold text-[15px] leading-5 text-foreground hover:underline truncate">
                            {rec.user.display_name}
                          </span>
                          {rec.user.is_verified && (
                            <CheckCircle2 className="w-4 h-4 fill-brand text-background shrink-0 inline" />
                          )}
                        </div>
                        <span className="text-[14px] text-muted leading-4 truncate">@{rec.user.username}</span>
                        {rec.reason && (
                          <span className="text-[12px] text-muted/80 leading-4 mt-0.5 truncate">
                            {rec.reason}
                          </span>
                        )}
                      </div>
                    </div>
                    <div className="flex items-center gap-1.5 shrink-0">
                      <button
                        onClick={(e) => handleToggleFollow(e, rec)}
                        className={`text-[14px] font-bold py-1.5 px-4 rounded-full transition-all ${
                          isFollowing
                            ? "border border-border text-foreground hover:border-red-500 hover:text-red-500 hover:bg-red-500/10"
                            : "bg-foreground text-background hover:opacity-90"
                        }`}
                      >
                        {isFollowing ? "Following" : "Follow"}
                      </button>
                      <button
                        title="Dismiss recommendation"
                        aria-label="Dismiss"
                        onClick={(e) => handleDismiss(e, rec.user.id)}
                        className="p-1 rounded-full text-muted hover:text-foreground hover:bg-subtle opacity-0 group-hover:opacity-100 transition-opacity"
                      >
                        <X className="w-3.5 h-3.5" />
                      </button>
                    </div>
                  </div>
                );
              })
            ) : (
              <div className="px-4 py-4 text-sm text-muted">No recommendations right now.</div>
            )}
            <Link
              href="/explore"
              className="px-4 py-3 text-brand text-[15px] hover:bg-subtle transition-colors block font-medium"
            >
              Show more
            </Link>
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
      </div>
    </aside>
  );
}
