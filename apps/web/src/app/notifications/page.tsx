"use client";

import React, { useEffect, useState } from "react";
import Link from "next/link";
import { Settings, Heart, Repeat2, UserPlus, MessageCircle, Loader2 } from "lucide-react";
import { LeftNavRail } from "@/components/layout/LeftNavRail";
import { RightSidebar } from "@/components/layout/RightSidebar";
import { AuthModal } from "@/components/auth/AuthModal";
import { Notification, User } from "@/types";
import { apiRequest } from "@/lib/api";
import { cn } from "@/lib/utils";

export default function NotificationsPage() {
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<"all" | "verified" | "mentions">("all");
  const [authModalOpen, setAuthModalOpen] = useState(false);

  useEffect(() => {
    apiRequest<User>("/me")
      .then((u) => {
        setCurrentUser(u);
        return apiRequest<{ notifications: Notification[] }>("/notifications");
      })
      .then((res) => {
        setNotifications(res.notifications || []);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const getIconForType = (type: string) => {
    switch (type) {
      case "like":
        return <Heart className="w-7 h-7 fill-like text-like" />;
      case "repost":
        return <Repeat2 className="w-7 h-7 text-repost" />;
      case "follow":
        return <UserPlus className="w-7 h-7 fill-brand text-brand" />;
      case "reply":
      case "mention":
        return <MessageCircle className="w-7 h-7 fill-brand text-brand" />;
      default:
        return <Heart className="w-7 h-7 fill-brand text-brand" />;
    }
  };

  return (
    <div className="min-h-screen bg-background text-foreground flex justify-center">
      <div className="w-full max-w-[1280px] flex">
        <LeftNavRail
          currentUser={currentUser}
          onOpenAuth={() => setAuthModalOpen(true)}
        />

        <main className="w-full sm:w-[600px] border-r border-border min-h-screen flex flex-col">
          {/* Header */}
          <header className="sticky top-0 bg-background/80 backdrop-blur-md border-b border-border z-20 flex flex-col">
            <div className="flex items-center justify-between px-4 h-[53px]">
              <h1 className="font-extrabold text-[20px]">Notifications</h1>
              <button className="p-2 hover:bg-subtle rounded-full text-foreground transition-colors">
                <Settings className="w-5 h-5" />
              </button>
            </div>

            {/* Tabs */}
            <div className="flex w-full h-[53px]">
              {(["all", "verified", "mentions"] as const).map((tab) => (
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
          </header>

          {/* Notifications List */}
          <div className="flex-1 flex flex-col">
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
            ) : notifications.length === 0 ? (
              <div className="text-center py-16 px-6">
                <h3 className="font-bold text-[22px] mb-2">Nothing to see here — yet</h3>
                <p className="text-muted text-[15px] max-w-sm mx-auto">
                  From likes to reposts and a whole lot more, this is where all the action about your posts and account will happen.
                </p>
              </div>
            ) : (
              <div className="flex flex-col divide-y divide-border">
                {notifications.map((n) => (
                  <div
                    key={n.id}
                    className="flex gap-4 p-4 hover:bg-subtle/50 transition-colors cursor-pointer"
                  >
                    <div className="shrink-0 pt-0.5">{getIconForType(n.type)}</div>
                    <div className="flex-1">
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
                      </p>
                      {n.post && (
                        <p className="text-[14px] text-muted mt-1 line-clamp-2">
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

        <RightSidebar />

        <AuthModal
          isOpen={authModalOpen}
          onClose={() => setAuthModalOpen(false)}
          onSuccess={(u) => setCurrentUser(u)}
        />
      </div>
    </div>
  );
}
