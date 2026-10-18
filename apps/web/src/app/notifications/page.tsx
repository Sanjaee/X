"use client";

import React, { useEffect, useState, useCallback } from "react";
import Link from "next/link";
import {
  Settings,
  Heart,
  Repeat2,
  UserPlus,
  MessageCircle,
  Mail,
  UserCheck,
  CheckCheck,
  Check,
  Loader2,
} from "lucide-react";
import { LeftNavRail } from "@/components/layout/LeftNavRail";
import { RightSidebar } from "@/components/layout/RightSidebar";
import { MobileDrawer } from "@/components/layout/MobileDrawer";
import { MobileBottomNav } from "@/components/layout/MobileBottomNav";
import { MobileComposerModal } from "@/components/composer/MobileComposerModal";
import { AuthModal } from "@/components/auth/AuthModal";
import { Notification, User, FollowRequest } from "@/types";
import { apiRequest } from "@/lib/api";
import { cn } from "@/lib/utils";
import { useRealtime } from "@/lib/useRealtime";
import { useCurrentUser } from "@/lib/useCurrentUser";

export default function NotificationsPage() {
  const { currentUser, updateUser: setCurrentUser } = useCurrentUser();
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [followRequests, setFollowRequests] = useState<FollowRequest[]>([]);
  const [requestStates, setRequestStates] = useState<Record<string, "accepted" | "rejected">>({});
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<"all" | "verified" | "mentions">("all");
  const [authModalOpen, setAuthModalOpen] = useState(false);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [composerModalOpen, setComposerModalOpen] = useState(false);
  const { setUnreadNotifCount } = useRealtime();

  const fetchNotifs = useCallback(() => {
    return apiRequest<{ notifications: Notification[] }>("/notifications?limit=50")
      .then((res) => {
        setNotifications(res.notifications || []);
      })
      .catch(() => {});
  }, []);

  const fetchFollowRequests = useCallback(() => {
    return apiRequest<{ requests: FollowRequest[] }>("/follow-requests")
      .then((res) => {
        setFollowRequests(res.requests || []);
      })
      .catch(() => {});
  }, []);

  useEffect(() => {
    Promise.all([fetchNotifs(), fetchFollowRequests()]).finally(() => setLoading(false));

    // Automatically mark as read on visit
    apiRequest("/notifications/read", { method: "PATCH" })
      .then(() => setUnreadNotifCount(0))
      .catch(() => {});

    // Listen for live incoming notifications
    const handleNewNotif = () => {
      fetchNotifs();
      fetchFollowRequests();
    };

    window.addEventListener("x_new_notification", handleNewNotif);
    return () => {
      window.removeEventListener("x_new_notification", handleNewNotif);
    };
  }, [setUnreadNotifCount, fetchNotifs, fetchFollowRequests]);

  const handleMarkAllRead = async () => {
    try {
      await apiRequest("/notifications/read", { method: "PATCH" });
      setNotifications((prev) => prev.map((n) => ({ ...n, is_read: true })));
      setUnreadNotifCount(0);
    } catch (e) {
      console.error("Failed to mark all as read", e);
    }
  };

  const handleRespondFollowRequest = async (n: Notification, action: "accept" | "reject") => {
    // Optimistic state
    setRequestStates((prev) => ({ ...prev, [n.id]: action === "accept" ? "accepted" : "rejected" }));

    try {
      // 1. Try by entity_id (the follow_request UUID)
      if (n.entity_id) {
        await apiRequest(`/follow-requests/${n.entity_id}/respond`, {
          method: "POST",
          body: JSON.stringify({ action }),
        });
      } else if (n.actor?.username) {
        // 2. Fallback to responding by username
        await apiRequest(`/users/${n.actor.username}/follow-request/respond`, {
          method: "POST",
          body: JSON.stringify({ action }),
        });
      }
      fetchFollowRequests();
    } catch (e) {
      console.error("Failed to respond to follow request", e);
    }
  };

  const getIconForType = (type: string) => {
    switch (type) {
      case "like":
        return <Heart className="w-7 h-7 fill-rose-500 text-rose-500" />;
      case "repost":
        return <Repeat2 className="w-7 h-7 text-emerald-500" />;
      case "follow":
        return <UserPlus className="w-7 h-7 fill-brand text-brand" />;
      case "follow_request":
        return <UserCheck className="w-7 h-7 text-amber-500" />;
      case "follow_accept":
        return <UserCheck className="w-7 h-7 fill-emerald-500 text-emerald-500" />;
      case "dm":
        return <Mail className="w-7 h-7 fill-blue-400 text-blue-400" />;
      case "reply":
      case "mention":
        return <MessageCircle className="w-7 h-7 fill-brand text-brand" />;
      default:
        return <Heart className="w-7 h-7 fill-brand text-brand" />;
    }
  };

  const filteredNotifications = notifications.filter((n) => {
    // Direct messages are dedicated to the Messages tab and badge
    if (n.type === "dm") return false;

    if (activeTab === "verified") {
      return n.actor?.is_verified;
    }
    if (activeTab === "mentions") {
      return n.type === "mention" || n.type === "reply";
    }
    return true;
  });

  return (
    <>
      <main className="w-full sm:w-[600px] border-r border-border min-h-screen flex flex-col pb-16 sm:pb-0">
          {/* Header */}
          <header className="sticky top-0 bg-background/80 backdrop-blur-md border-b border-border z-20 flex flex-col">
            <div className="flex items-center justify-between px-4 h-[53px]">
              <div className="flex items-center gap-3">
                <button
                  onClick={() => setDrawerOpen(true)}
                  className="sm:hidden w-8 h-8 rounded-full overflow-hidden border border-border flex items-center justify-center shrink-0"
                >
                  {currentUser?.avatar_url ? (
                    <img
                      src={currentUser.avatar_url}
                      alt={currentUser.display_name}
                      className="w-full h-full object-cover"
                    />
                  ) : (
                    <div className="w-full h-full bg-[#16181c] flex items-center justify-center text-xs font-bold text-foreground">
                      {currentUser?.display_name?.charAt(0) || "X"}
                    </div>
                  )}
                </button>
                <h1 className="font-extrabold text-[20px]">Notifications</h1>
              </div>

              <div className="flex items-center gap-2">
                {currentUser && (
                  <button
                    onClick={handleMarkAllRead}
                    title="Mark all as read"
                    className="p-2 hover:bg-subtle rounded-full text-muted hover:text-foreground transition-colors"
                  >
                    <CheckCheck className="w-5 h-5" />
                  </button>
                )}
                <button
                  className="p-2 hover:bg-subtle rounded-full text-foreground transition-colors"
                  title="Settings"
                >
                  <Settings className="w-5 h-5" />
                </button>
              </div>
            </div>

            {/* Tabs: All & Mentions */}
            <div className="flex w-full h-[53px]">
              {[
                { id: "all", label: "All" },
                { id: "mentions", label: "Mentions" },
              ].map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id as "all" | "mentions")}
                  className="flex-1 flex justify-center items-center hover:bg-subtle transition-colors relative text-[15px]"
                >
                  <span
                    className={cn(
                      "font-bold py-3.5 relative",
                      activeTab === tab.id ? "text-foreground" : "text-muted"
                    )}
                  >
                    {tab.label}
                    {activeTab === tab.id && (
                      <div className="absolute bottom-0 left-0 right-0 h-1 bg-brand rounded-full" />
                    )}
                  </span>
                </button>
              ))}
            </div>
          </header>

          {/* Notifications List */}
          <div className="flex-1 flex flex-col">
            {/* Pending Follow Requests Banner */}
            {followRequests.length > 0 && (
              <div className="p-4 border-b border-border bg-subtle/20 flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-full bg-amber-500/15 flex items-center justify-center text-amber-500 shrink-0">
                    <UserCheck className="w-5 h-5" />
                  </div>
                  <div>
                    <h4 className="font-bold text-[15px] leading-5 text-foreground">Follow requests</h4>
                    <p className="text-[13px] text-muted">
                      {followRequests.length} pending {followRequests.length === 1 ? "request" : "requests"}
                    </p>
                  </div>
                </div>
                <div className="flex -space-x-2 overflow-hidden">
                  {followRequests.slice(0, 3).map((fr) => (
                    <img
                      key={fr.id}
                      src={fr.requester?.avatar_url || "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"}
                      alt={fr.requester?.display_name || "User"}
                      className="inline-block h-7 w-7 rounded-full ring-2 ring-background object-cover"
                    />
                  ))}
                </div>
              </div>
            )}

            {loading ? (
              <div className="flex justify-center items-center p-16 text-brand">
                <Loader2 className="w-8 h-8 animate-spin" />
              </div>
            ) : !currentUser ? (
              <div className="text-center py-16 px-6">
                <h3 className="font-bold text-[22px] mb-2">Join to see notifications</h3>
                <p className="text-muted text-[15px] max-w-sm mx-auto mb-6">
                  Sign in to view likes, replies, mentions, and follows from people you know.
                </p>
                <button
                  onClick={() => setAuthModalOpen(true)}
                  className="bg-brand hover:bg-brand-hover text-white font-bold py-2.5 px-6 rounded-full text-[15px]"
                >
                  Sign in
                </button>
              </div>
            ) : filteredNotifications.length === 0 ? (
              <div className="text-center py-16 px-6">
                <h3 className="font-bold text-[22px] mb-2">Nothing to see here — yet</h3>
                <p className="text-muted text-[15px] max-w-sm mx-auto">
                  {activeTab === "verified"
                    ? "Likes, mentions, and other interactions from verified users will appear here."
                    : activeTab === "mentions"
                    ? "Conversations, replies, and mentions will appear here."
                    : "From likes to reposts and a whole lot more, this is where all the action about your posts and account will happen."}
                </p>
              </div>
            ) : (
              <div className="flex flex-col divide-y divide-border">
                {filteredNotifications.map((n) => (
                  <div
                    key={n.id}
                    className={cn(
                      "flex gap-4 p-4 hover:bg-subtle/50 transition-colors",
                      !n.is_read && "bg-brand/5"
                    )}
                  >
                    <div className="shrink-0 pt-0.5">{getIconForType(n.type)}</div>
                    <div className="flex-1 min-w-0">
                      <Link href={`/${n.actor?.username || "user"}`} className="block mb-2">
                        <img
                          src={
                            n.actor?.avatar_url ||
                            "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"
                          }
                          alt={n.actor?.display_name || "User"}
                          className="w-8 h-8 rounded-full object-cover border border-border"
                        />
                      </Link>
                      <p className="text-[15px] leading-5 text-foreground">
                        <Link
                          href={`/${n.actor?.username || "user"}`}
                          className="font-bold hover:underline"
                        >
                          {n.actor?.display_name || n.actor?.username}
                        </Link>{" "}
                        {n.type === "like" && "liked your post"}
                        {n.type === "repost" && "reposted your post"}
                        {n.type === "reply" && "replied to your post"}
                        {n.type === "follow" && "followed you"}
                        {n.type === "follow_request" && "requested to follow you"}
                        {n.type === "follow_accept" && "accepted your follow request"}
                        {n.type === "dm" && "sent you a direct message"}
                      </p>

                      {/* Follow Request Action Buttons */}
                      {n.type === "follow_request" && (
                        <div className="flex items-center gap-2 mt-3">
                          {requestStates[n.id] === "accepted" ? (
                            <span className="text-emerald-500 font-bold text-[13px] flex items-center gap-1.5 bg-emerald-500/10 px-3.5 py-1.5 rounded-full">
                              <Check className="w-4 h-4" /> Request accepted
                            </span>
                          ) : requestStates[n.id] === "rejected" ? (
                            <span className="text-muted font-bold text-[13px] bg-subtle px-3.5 py-1.5 rounded-full">
                              Request declined
                            </span>
                          ) : (
                            <>
                              <button
                                onClick={(e) => {
                                  e.preventDefault();
                                  e.stopPropagation();
                                  handleRespondFollowRequest(n, "accept");
                                }}
                                className="bg-foreground text-background hover:opacity-90 font-bold px-4 py-1.5 rounded-full text-[13px] transition-all shadow-sm active:scale-95 cursor-pointer"
                              >
                                Accept
                              </button>
                              <button
                                onClick={(e) => {
                                  e.preventDefault();
                                  e.stopPropagation();
                                  handleRespondFollowRequest(n, "reject");
                                }}
                                className="border border-border hover:bg-subtle text-foreground font-bold px-4 py-1.5 rounded-full text-[13px] transition-all active:scale-95 cursor-pointer"
                              >
                                Decline
                              </button>
                            </>
                          )}
                        </div>
                      )}

                      {/* Direct Message Link */}
                      {n.type === "dm" && (
                        <div className="mt-1.5">
                          <Link
                            href={n.entity_id ? `/messages/${n.entity_id}` : "/messages"}
                            className="inline-flex items-center gap-1 text-[13px] text-brand hover:underline font-semibold"
                          >
                            Open message →
                          </Link>
                        </div>
                      )}

                      {/* Post Snippet */}
                      {n.post && (
                        <p className="text-[14px] text-muted mt-1.5 line-clamp-2 bg-subtle/30 p-2.5 rounded-xl border border-border/50">
                          {n.post.content}
                        </p>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </main>

        <RightSidebar
          currentUser={currentUser}
          onOpenAuth={() => setAuthModalOpen(true)}
        />

        {/* Mobile Navigation Drawer */}
        <MobileDrawer
          isOpen={drawerOpen}
          onClose={() => setDrawerOpen(false)}
          currentUser={currentUser}
          onOpenAuth={() => setAuthModalOpen(true)}
          onLogout={() => {
            setCurrentUser(null);
            setNotifications([]);
          }}
        />

        {/* Mobile Composer Modal */}
        <MobileComposerModal
          isOpen={composerModalOpen}
          onClose={() => setComposerModalOpen(false)}
          currentUser={currentUser}
          onPostSuccess={() => fetchNotifs()}
          onOpenAuth={() => setAuthModalOpen(true)}
        />

        <AuthModal
          isOpen={authModalOpen}
          onClose={() => setAuthModalOpen(false)}
          onSuccess={(u) => setCurrentUser(u)}
        />
    </>
  );
}
