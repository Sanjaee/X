"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import {
  Users,
  Lock,
  Globe,
  Shield,
  ArrowLeft,
  Settings,
  UserCheck,
  UserPlus,
  Clock,
  Pin,
  CheckCircle,
  XCircle,
  AlertCircle,
} from "lucide-react";
import { apiRequest } from "@/lib/api";
import { Community, CommunityRule, CommunityMember, CommunityJoinRequest, Post } from "@/types";
import { useCurrentUser } from "@/lib/useCurrentUser";
import { PostCard } from "@/components/feed/PostCard";
import { cn } from "@/lib/utils";

export default function CommunityDetailPage() {
  const params = useParams();
  const slug = params?.slug as string;
  const router = useRouter();
  const { currentUser } = useCurrentUser();

  const [community, setCommunity] = useState<Community | null>(null);
  const [rules, setRules] = useState<CommunityRule[]>([]);
  const [posts, setPosts] = useState<Post[]>([]);
  const [members, setMembers] = useState<CommunityMember[]>([]);
  const [joinRequests, setJoinRequests] = useState<CommunityJoinRequest[]>([]);
  const [activeTab, setActiveTab] = useState<"feed" | "about" | "members" | "requests">("feed");
  const [isLoading, setIsLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState("");

  const loadCommunityData = async () => {
    if (!slug) return;
    setIsLoading(true);
    try {
      const res = await apiRequest<{ community: Community; rules: CommunityRule[] }>(
        `/communities/${slug}`
      );
      setCommunity(res.community);
      setRules(res.rules || []);

      // Load posts
      const postsRes = await apiRequest<{ posts: Post[] }>(`/communities/${slug}/posts?limit=30`);
      setPosts(postsRes.posts || []);

      // If user is moderator or owner, load pending join requests
      const role = res.community.viewer_role;
      if (role === "owner" || role === "admin" || role === "moderator") {
        try {
          const reqRes = await apiRequest<{ requests: CommunityJoinRequest[] }>(
            `/communities/${slug}/requests`
          );
          setJoinRequests(reqRes.requests || []);
        } catch (_) {}
      }
    } catch (err: any) {
      setErrorMsg(err.message || "Failed to load community");
    } finally {
      setIsLoading(false);
    }
  };

  const loadMembers = async () => {
    if (!slug) return;
    try {
      const res = await apiRequest<{ members: CommunityMember[]; total: number }>(
        `/communities/${slug}/members?limit=50`
      );
      setMembers(res.members || []);
    } catch (err) {
      console.error("Failed to load members", err);
    }
  };

  useEffect(() => {
    loadCommunityData();
  }, [slug, currentUser]);

  useEffect(() => {
    if (activeTab === "members") {
      loadMembers();
    }
  }, [activeTab]);

  const handleJoinOrLeave = async () => {
    if (!community || !currentUser) {
      router.push("/login");
      return;
    }

    setActionLoading(true);
    try {
      if (community.viewer_status === "active") {
        // Leave
        await apiRequest(`/communities/${slug}/leave`, { method: "POST" });
        setCommunity({
          ...community,
          viewer_role: "",
          viewer_status: "",
          members_count: Math.max(0, community.members_count - 1),
        });
      } else {
        // Join or Request
        const res = await apiRequest<{ status: string; message: string }>(
          `/communities/${slug}/join`,
          { method: "POST" }
        );
        if (res.status === "active") {
          setCommunity({
            ...community,
            viewer_role: "member",
            viewer_status: "active",
            members_count: community.members_count + 1,
          });
        } else if (res.status === "pending") {
          setCommunity({
            ...community,
            viewer_status: "pending",
          });
        }
      }
    } catch (err: any) {
      alert(err.message || "Action failed");
    } finally {
      setActionLoading(false);
    }
  };

  const handleReviewRequest = async (requestId: string, status: "approved" | "rejected") => {
    try {
      await apiRequest(`/communities/${slug}/requests/${requestId}/review`, {
        method: "POST",
        body: JSON.stringify({ status }),
      });
      setJoinRequests(joinRequests.filter((r) => r.id !== requestId));
      if (status === "approved" && community) {
        setCommunity({
          ...community,
          members_count: community.members_count + 1,
        });
      }
    } catch (err: any) {
      alert(err.message || "Failed to review request");
    }
  };

  if (isLoading) {
    return (
      <main className="w-full sm:w-[600px] border-r border-border min-h-screen flex items-center justify-center p-8 text-muted-foreground">
        Loading community...
      </main>
    );
  }

  if (errorMsg || !community) {
    return (
      <main className="w-full sm:w-[600px] border-r border-border min-h-screen p-8 text-center">
        <AlertCircle className="w-12 h-12 text-destructive mx-auto mb-3" />
        <h2 className="text-xl font-bold mb-1">Community Not Found</h2>
        <p className="text-sm text-muted-foreground mb-4">
          This community may have been removed or does not exist.
        </p>
        <Link
          href="/communities"
          className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-brand text-white text-sm font-bold"
        >
          <ArrowLeft className="w-4 h-4" /> Back to Communities
        </Link>
      </main>
    );
  }

  const isModOrOwner =
    community.viewer_role === "owner" ||
    community.viewer_role === "admin" ||
    community.viewer_role === "moderator";

  return (
    <main
      className={cn(
        "w-full sm:w-[600px] border-r border-border min-h-screen flex flex-col bg-background text-foreground",
        !currentUser && "border-l"
      )}
    >
      {/* Top Sticky Header */}
      <header className="sticky top-0 z-20 bg-background/80 backdrop-blur-md border-b border-border px-4 py-2.5 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <button
            onClick={() => router.push("/communities")}
            className="p-2 -ml-2 rounded-full hover:bg-subtle transition-colors"
            aria-label="Back"
          >
            <ArrowLeft className="w-5 h-5 text-foreground" />
          </button>
          <div>
            <h1 className="text-lg font-bold leading-tight truncate max-w-[340px]">
              {community.name}
            </h1>
            <p className="text-xs text-muted-foreground">
              {community.members_count} {community.members_count === 1 ? "member" : "members"}
            </p>
          </div>
        </div>
      </header>

      {/* Banner / Header Image */}
      <div className="h-36 sm:h-44 w-full bg-gradient-to-r from-brand/20 via-brand/10 to-subtle relative overflow-hidden">
        {community.banner_url ? (
          <img
            src={community.banner_url}
            alt={community.name}
            className="w-full h-full object-cover"
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center opacity-10">
            <Users className="w-24 h-24 text-brand" />
          </div>
        )}
      </div>

      {/* Community Info Header */}
      <div className="px-4 pb-4 border-b border-border">
        <div className="flex items-end justify-between -mt-12 mb-3">
          <div className="w-20 h-20 sm:w-24 sm:h-24 rounded-3xl bg-background border-4 border-background overflow-hidden shadow-lg flex items-center justify-center text-brand font-bold text-2xl shrink-0">
            {community.avatar_url ? (
              <img
                src={community.avatar_url}
                alt={community.name}
                className="w-full h-full object-cover"
              />
            ) : (
              community.name.slice(0, 2).toUpperCase()
            )}
          </div>

          <div className="flex items-center gap-2">
            {community.viewer_status === "active" ? (
              <button
                onClick={handleJoinOrLeave}
                disabled={actionLoading || community.viewer_role === "owner"}
                className={cn(
                  "px-4 py-1.5 rounded-full text-sm font-bold border transition-all active:scale-95",
                  community.viewer_role === "owner"
                    ? "border-border text-muted-foreground cursor-default"
                    : "border-border hover:border-red-500/50 hover:bg-red-500/10 hover:text-red-500"
                )}
              >
                {community.viewer_role === "owner" ? "Owner" : "Joined"}
              </button>
            ) : community.viewer_status === "pending" ? (
              <button
                disabled
                className="px-4 py-1.5 rounded-full text-sm font-bold border border-border text-muted-foreground flex items-center gap-1.5 bg-subtle"
              >
                <Clock className="w-3.5 h-3.5" /> Pending Approval
              </button>
            ) : (
              <button
                onClick={handleJoinOrLeave}
                disabled={actionLoading}
                className="px-5 py-1.5 rounded-full bg-foreground text-background text-sm font-bold hover:opacity-90 active:scale-95 transition-all shadow"
              >
                {community.access_model === "restricted" ? "Request to Join" : "Join"}
              </button>
            )}
          </div>
        </div>

        <div>
          <div className="flex items-center gap-2">
            <h2 className="text-xl font-bold">{community.name}</h2>
            {community.access_model === "restricted" ? (
              <span className="flex items-center gap-1 text-[11px] font-medium text-amber-500 bg-amber-500/10 px-2 py-0.5 rounded-full">
                <Lock className="w-3 h-3" /> Restricted
              </span>
            ) : (
              <span className="flex items-center gap-1 text-[11px] font-medium text-emerald-500 bg-emerald-500/10 px-2 py-0.5 rounded-full">
                <Globe className="w-3 h-3" /> Open
              </span>
            )}
          </div>

          <p className="text-xs text-muted-foreground mt-0.5">/communities/{community.slug}</p>

          {community.description && (
            <p className="text-sm mt-3 leading-relaxed text-foreground/90 whitespace-pre-line">
              {community.description}
            </p>
          )}

          <div className="flex items-center gap-4 text-xs text-muted-foreground mt-3 pt-3 border-t border-border/50">
            <span>
              <strong className="text-foreground">{community.members_count}</strong> Members
            </span>
            <span>
              <strong className="text-foreground">{rules.length}</strong> Rules
            </span>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex border-b border-border bg-background sticky top-[53px] z-10">
        <button
          onClick={() => setActiveTab("feed")}
          className={cn(
            "flex-1 py-3 text-sm font-bold transition-colors relative text-center",
            activeTab === "feed" ? "text-foreground" : "text-muted-foreground hover:bg-subtle/50"
          )}
        >
          Timeline
          {activeTab === "feed" && (
            <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-12 h-1 bg-brand rounded-full" />
          )}
        </button>

        <button
          onClick={() => setActiveTab("about")}
          className={cn(
            "flex-1 py-3 text-sm font-bold transition-colors relative text-center",
            activeTab === "about" ? "text-foreground" : "text-muted-foreground hover:bg-subtle/50"
          )}
        >
          About & Rules
          {activeTab === "about" && (
            <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-12 h-1 bg-brand rounded-full" />
          )}
        </button>

        <button
          onClick={() => setActiveTab("members")}
          className={cn(
            "flex-1 py-3 text-sm font-bold transition-colors relative text-center",
            activeTab === "members" ? "text-foreground" : "text-muted-foreground hover:bg-subtle/50"
          )}
        >
          Members
          {activeTab === "members" && (
            <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-12 h-1 bg-brand rounded-full" />
          )}
        </button>

        {isModOrOwner && (
          <button
            onClick={() => setActiveTab("requests")}
            className={cn(
              "flex-1 py-3 text-sm font-bold transition-colors relative text-center flex items-center justify-center gap-1.5",
              activeTab === "requests" ? "text-foreground" : "text-muted-foreground hover:bg-subtle/50"
            )}
          >
            <span>Requests</span>
            {joinRequests.length > 0 && (
              <span className="w-4 h-4 rounded-full bg-brand text-white text-[10px] font-black flex items-center justify-center">
                {joinRequests.length}
              </span>
            )}
            {activeTab === "requests" && (
              <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-16 h-1 bg-brand rounded-full" />
            )}
          </button>
        )}
      </div>

      {/* Tab Panels */}
      <div className="flex-1">
        {activeTab === "feed" && (
          <div className="divide-y divide-border">
            {posts.length === 0 ? (
              <div className="p-12 text-center text-muted-foreground">
                <Users className="w-12 h-12 mx-auto mb-3 opacity-30" />
                <p className="font-semibold text-foreground">No posts yet</p>
                <p className="text-sm mt-1">Be the first to post in this community!</p>
              </div>
            ) : (
              posts.map((post) => <PostCard key={post.id} post={post} />)
            )}
          </div>
        )}

        {activeTab === "about" && (
          <div className="p-5 space-y-6">
            <div>
              <h3 className="font-bold text-base mb-2">About this Community</h3>
              <p className="text-sm leading-relaxed text-foreground/80">
                {community.description || "No description provided for this community."}
              </p>
            </div>

            <div>
              <h3 className="font-bold text-base mb-3">Community Rules</h3>
              {rules.length === 0 ? (
                <p className="text-xs text-muted-foreground italic">
                  No specific rules established yet. Standard platform rules apply.
                </p>
              ) : (
                <div className="space-y-3">
                  {rules.map((rule, idx) => (
                    <div
                      key={rule.id || idx}
                      className="p-3.5 rounded-xl border border-border bg-subtle/40"
                    >
                      <h4 className="font-bold text-sm text-foreground flex items-center gap-2">
                        <span className="w-5 h-5 rounded-full bg-brand/10 text-brand text-xs flex items-center justify-center font-black">
                          {rule.rule_number || idx + 1}
                        </span>
                        {rule.title}
                      </h4>
                      {rule.description && (
                        <p className="text-xs text-muted-foreground mt-1.5 pl-7 leading-relaxed">
                          {rule.description}
                        </p>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        )}

        {activeTab === "members" && (
          <div className="divide-y divide-border">
            {members.length === 0 ? (
              <div className="p-8 text-center text-muted-foreground">Loading members...</div>
            ) : (
              members.map((m) => (
                <div key={m.id} className="p-4 flex items-center justify-between hover:bg-subtle/30">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-full bg-subtle overflow-hidden flex items-center justify-center font-bold text-brand">
                      {m.user?.avatar_url ? (
                        <img
                          src={m.user.avatar_url}
                          alt={m.user.username}
                          className="w-full h-full object-cover"
                        />
                      ) : (
                        m.user?.username?.slice(0, 1).toUpperCase() || "U"
                      )}
                    </div>
                    <div>
                      <p className="font-bold text-sm text-foreground">
                        {m.user?.display_name || m.user?.username}
                      </p>
                      <p className="text-xs text-muted-foreground">@{m.user?.username}</p>
                    </div>
                  </div>

                  <div>
                    <span
                      className={cn(
                        "text-[10px] font-extrabold uppercase px-2.5 py-1 rounded-full",
                        m.role === "owner"
                          ? "bg-amber-500/10 text-amber-500"
                          : m.role === "moderator" || m.role === "admin"
                          ? "bg-brand/10 text-brand"
                          : "bg-subtle text-muted-foreground"
                      )}
                    >
                      {m.role}
                    </span>
                  </div>
                </div>
              ))
            )}
          </div>
        )}

        {activeTab === "requests" && isModOrOwner && (
          <div className="divide-y divide-border">
            {joinRequests.length === 0 ? (
              <div className="p-12 text-center text-muted-foreground">
                <CheckCircle className="w-12 h-12 text-emerald-500 mx-auto mb-2 opacity-50" />
                <p className="font-semibold text-foreground">All caught up!</p>
                <p className="text-sm mt-0.5">No pending membership requests to review.</p>
              </div>
            ) : (
              joinRequests.map((req) => (
                <div key={req.id} className="p-4 flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-full bg-subtle overflow-hidden flex items-center justify-center font-bold text-brand">
                      {req.user?.avatar_url ? (
                        <img
                          src={req.user.avatar_url}
                          alt={req.user.username}
                          className="w-full h-full object-cover"
                        />
                      ) : (
                        req.user?.username?.slice(0, 1).toUpperCase() || "U"
                      )}
                    </div>
                    <div>
                      <p className="font-bold text-sm text-foreground">
                        {req.user?.display_name || req.user?.username}
                      </p>
                      <p className="text-xs text-muted-foreground">@{req.user?.username}</p>
                    </div>
                  </div>

                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => handleReviewRequest(req.id, "approved")}
                      className="px-3 py-1 rounded-full bg-emerald-500 hover:bg-emerald-600 text-white text-xs font-bold transition-all"
                    >
                      Approve
                    </button>
                    <button
                      onClick={() => handleReviewRequest(req.id, "rejected")}
                      className="px-3 py-1 rounded-full border border-border text-xs font-bold hover:bg-subtle transition-all"
                    >
                      Reject
                    </button>
                  </div>
                </div>
              ))
            )}
          </div>
        )}
      </div>
    </main>
  );
}
