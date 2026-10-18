"use client";

import React, { useEffect, useState, useMemo } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { MoreHorizontal, CheckCircle2, X, Loader2 } from "lucide-react";
import { signIn } from "next-auth/react";
import { useQuery } from "@tanstack/react-query";
import { formatNumber } from "@/lib/utils";
import { SearchBarDropdown } from "@/components/search/SearchBarDropdown";
import { apiRequest, setStoredToken, setStoredUser } from "@/lib/api";
import { TrendingTopic, UserRecommendation, User, AuthResponse } from "@/types";
import { AuthModal } from "@/components/auth/AuthModal";
import { useCurrentUser } from "@/lib/useCurrentUser";

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
  const { currentUser: authUser, updateUser } = useCurrentUser();
  const user = propUser !== undefined ? propUser : authUser;
  const [localAuthOpen, setLocalAuthOpen] = useState(false);
  const [authModalMode, setAuthModalMode] = useState<"login" | "register">("login");
  const [followingMap, setFollowingMap] = useState<Record<string, boolean>>({});

  // Inline sidebar login state
  const [sidebarIdentifier, setSidebarIdentifier] = useState("");
  const [sidebarPassword, setSidebarPassword] = useState("");
  const [sidebarLoading, setSidebarLoading] = useState(false);
  const [sidebarGoogleLoading, setSidebarGoogleLoading] = useState(false);
  const [sidebarError, setSidebarError] = useState<string | null>(null);

  // Cached Trends
  const { data: trends = [], isLoading: isLoadingTrends } = useQuery<TrendingTopic[]>({
    queryKey: ["explore", "trending_sidebar"],
    queryFn: async () => {
      const res = await apiRequest<{ trending: TrendingTopic[] }>("/explore/trending?limit=5");
      return res.trending || [];
    },
    staleTime: 5 * 60 * 1000,
    gcTime: 15 * 60 * 1000,
    refetchOnWindowFocus: false,
  });

  // Cached Recommendations
  const { data: fetchedRecs = [], isLoading: isLoadingRecs } = useQuery<UserRecommendation[]>({
    queryKey: ["users", "recommendations_sidebar", user?.id || "guest"],
    queryFn: async () => {
      const res = await apiRequest<{ recommendations: UserRecommendation[] }>("/users/recommendations?limit=4");
      return res.recommendations || [];
    },
    staleTime: 5 * 60 * 1000,
    gcTime: 15 * 60 * 1000,
    refetchOnWindowFocus: false,
  });

  const [dismissedUserIds, setDismissedUserIds] = useState<Set<string>>(new Set());

  const recommendations = useMemo(() => {
    return fetchedRecs.filter((r) => !dismissedUserIds.has(r.user.id));
  }, [fetchedRecs, dismissedUserIds]);

  const handleTriggerAuth = (mode: "login" | "register" = "login") => {
    setAuthModalMode(mode);
    if (onOpenAuth) {
      onOpenAuth();
    } else {
      setLocalAuthOpen(true);
    }
  };

  const handleSidebarGoogleSignIn = async () => {
    try {
      setSidebarGoogleLoading(true);
      setSidebarError(null);
      await signIn("google", { callbackUrl: window.location.href });
    } catch (err: any) {
      setSidebarError(err?.message || "Failed to initialize Google authentication");
      setSidebarGoogleLoading(false);
    }
  };

  const handleSidebarLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setSidebarError(null);
    setSidebarLoading(true);

    try {
      const res = await apiRequest<AuthResponse>("/auth/login", {
        method: "POST",
        body: JSON.stringify({ identifier: sidebarIdentifier, password: sidebarPassword }),
      });
      setStoredToken(res.session_token);
      setStoredUser(res.user);
      window.dispatchEvent(new Event("x_auth_state_change"));
    } catch (err: any) {
      setSidebarError(err?.message || "Invalid credentials");
    } finally {
      setSidebarLoading(false);
    }
  };

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
    setDismissedUserIds((prev) => {
      const next = new Set(prev);
      next.add(targetUserID);
      return next;
    });

    try {
      await apiRequest(`/users/recommendations/${targetUserID}/dismiss`, { method: "POST" });
    } catch (err) {
      console.error("Failed to record recommendation dismissal", err);
    }
  };

  const renderSignInCard = () => (
    <div className="bg-transparent rounded-2xl p-5 border border-border mb-4 flex flex-col items-center">
      {/* X Logo */}
      <div className="mb-4 text-foreground">
        <svg viewBox="0 0 24 24" aria-hidden="true" className="w-8 h-8 fill-current">
          <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z" />
        </svg>
      </div>

      <h2 className="text-[22px] font-extrabold mb-5 text-center leading-tight text-foreground">
        Sign in to X
      </h2>

      {/* Google Sign-In Button */}
      <div className="w-full mb-3.5">
        <button
          type="button"
          onClick={handleSidebarGoogleSignIn}
          disabled={sidebarGoogleLoading || sidebarLoading}
          className="w-full flex items-center justify-center gap-3 bg-white hover:bg-neutral-200 text-[#0f1419] font-semibold text-[14px] h-[40px] px-4 rounded-full transition-colors shadow-sm select-none disabled:opacity-50"
        >
          {sidebarGoogleLoading ? (
            <Loader2 className="w-4 h-4 animate-spin text-neutral-600" />
          ) : (
            <svg className="w-4 h-4 shrink-0" viewBox="0 0 24 24">
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
          )}
          <span>Sign in with Google</span>
        </button>
      </div>

      {/* Horizontal Divider with OR */}
      <div className="w-full flex items-center mb-3.5">
        <div className="flex-1 h-px bg-border" />
        <span className="px-3 text-xs text-muted font-normal uppercase">OR</span>
        <div className="flex-1 h-px bg-border" />
      </div>

      {sidebarError && (
        <div className="w-full mb-3 p-2.5 rounded-lg bg-red-500/10 border border-red-500/30 text-red-500 text-xs text-center">
          {sidebarError}
        </div>
      )}

      {/* Credentials Form */}
      <form onSubmit={handleSidebarLogin} className="w-full flex flex-col gap-3">
        <div>
          <label className="text-xs text-muted font-medium mb-1 block">Phone, email, or username</label>
          <input
            type="text"
            required
            value={sidebarIdentifier}
            onChange={(e) => setSidebarIdentifier(e.target.value)}
            placeholder="Username or email"
            className="w-full bg-transparent border border-border focus:border-brand rounded-lg p-2.5 text-[14px] outline-none text-foreground"
          />
        </div>
        <div>
          <label className="text-xs text-muted font-medium mb-1 block">Password</label>
          <input
            type="password"
            required
            value={sidebarPassword}
            onChange={(e) => setSidebarPassword(e.target.value)}
            placeholder="Password"
            className="w-full bg-transparent border border-border focus:border-brand rounded-lg p-2.5 text-[14px] outline-none text-foreground"
          />
        </div>
        <button
          type="submit"
          disabled={sidebarLoading || sidebarGoogleLoading}
          className="w-full mt-1 bg-foreground text-background font-bold text-[15px] py-2.5 rounded-full hover:opacity-90 transition-opacity flex items-center justify-center disabled:opacity-50"
        >
          {sidebarLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : "Sign In"}
        </button>
      </form>

      {/* Footer */}
      <div className="mt-4 text-[13px] text-muted text-center">
        Don&apos;t have an account?{" "}
        <button
          type="button"
          onClick={() => handleTriggerAuth("register")}
          className="text-brand hover:underline font-medium"
        >
          Sign up
        </button>
      </div>
    </div>
  );

  // If user is explicitly not logged in (null) AND on a post detail page with relevantUser (matching reference screenshot)
  if (user === null && !!relevantUser) {
    const focal = relevantUser;

    return (
      <aside className="hidden lg:block w-[350px] xl:w-[420px] px-6 py-3 select-none shrink-0">
        <div className="sticky top-3 space-y-4 pb-12">
          {/* Top Auth Card: Sign in to X (matching image 2) */}
          {renderSignInCard()}

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
                  onClick={() => handleTriggerAuth("login")}
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
          onSuccess={(u) => updateUser(u)}
        />
      </aside>
    );
  }

  return (
    <aside className="hidden lg:block w-[350px] xl:w-[420px] px-6 py-3 select-none shrink-0">
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
        renderSignInCard()
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
            {trends.length > 0 && (
              <Link
                href="/explore"
                className="px-4 py-3 text-brand text-[15px] hover:bg-subtle transition-colors block font-medium"
              >
                Show more
              </Link>
            )}
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

        {/* Scan to get the app QR code card (matches reference image 2) */}
        {!user && (
          <div className="bg-transparent rounded-2xl border border-border p-4 flex flex-col items-center">
            <p className="text-[14px] font-bold text-foreground mb-3 text-center">
              Scan to get the app
            </p>
            <div className="relative p-2.5 bg-white rounded-xl flex items-center justify-center shadow-sm">
              <img
                src="https://api.qrserver.com/v1/create-qr-code/?size=140x140&data=https%3A%2F%2Flostmedia.pro"
                alt="Scan to get X"
                className="w-[140px] h-[140px]"
              />
              <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
                <div className="w-8 h-8 bg-black rounded-full flex items-center justify-center text-white border-2 border-white shadow-md">
                  <svg viewBox="0 0 24 24" aria-hidden="true" className="w-4 h-4 fill-current">
                    <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z" />
                  </svg>
                </div>
              </div>
            </div>
          </div>
        )}

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

      <AuthModal
        isOpen={localAuthOpen}
        onClose={() => setLocalAuthOpen(false)}
        onSuccess={(u) => updateUser(u)}
        initialMode={authModalMode}
      />
    </aside>
  );
}
