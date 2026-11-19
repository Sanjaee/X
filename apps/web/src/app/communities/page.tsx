"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Users, Search, Plus, Globe, Lock, Shield, ArrowRight, Check } from "lucide-react";
import { apiRequest } from "@/lib/api";
import { Community } from "@/types";
import { useCurrentUser } from "@/lib/useCurrentUser";
import { cn } from "@/lib/utils";

export default function CommunitiesPage() {
  const router = useRouter();
  const { currentUser } = useCurrentUser();
  const [activeTab, setActiveTab] = useState<"discover" | "my">("discover");
  const [search, setSearch] = useState("");
  const [communities, setCommunities] = useState<Community[]>([]);
  const [myCommunities, setMyCommunities] = useState<Community[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);

  // Create Modal State
  const [createName, setCreateName] = useState("");
  const [createSlug, setCreateSlug] = useState("");
  const [createDesc, setCreateDesc] = useState("");
  const [createAccess, setCreateAccess] = useState<"open" | "restricted">("open");
  const [createRuleTitle, setCreateRuleTitle] = useState("");
  const [createRuleDesc, setCreateRuleDesc] = useState("");
  const [isCreating, setIsCreating] = useState(false);
  const [errorMsg, setErrorMsg] = useState("");

  const loadData = async () => {
    setIsLoading(true);
    try {
      const commRes = await apiRequest<{ communities: Community[]; total: number }>(
        `/communities?q=${encodeURIComponent(search)}&limit=30`
      );
      setCommunities(commRes.communities || []);

      if (currentUser) {
        const myRes = await apiRequest<{ communities: Community[] }>("/communities/me");
        setMyCommunities(myRes.communities || []);
      }
    } catch (err) {
      console.error("Failed to load communities", err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    const timer = setTimeout(() => {
      loadData();
    }, 250);
    return () => clearTimeout(timer);
  }, [search, currentUser]);

  const handleCreateSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!createName.trim() || !createSlug.trim()) {
      setErrorMsg("Name and unique slug are required");
      return;
    }
    setIsCreating(true);
    setErrorMsg("");

    try {
      const payload: any = {
        name: createName.trim(),
        slug: createSlug.trim().toLowerCase().replace(/[^a-z0-9_-]/g, ""),
        description: createDesc.trim(),
        access_model: createAccess,
        visibility: "public",
        rules: [],
      };

      if (createRuleTitle.trim()) {
        payload.rules.push({
          title: createRuleTitle.trim(),
          description: createRuleDesc.trim(),
        });
      }

      await apiRequest("/communities", {
        method: "POST",
        body: JSON.stringify(payload),
      });

      setIsModalOpen(false);
      setCreateName("");
      setCreateSlug("");
      setCreateDesc("");
      setCreateRuleTitle("");
      setCreateRuleDesc("");
      router.push(`/communities/${payload.slug}`);
    } catch (err: any) {
      setErrorMsg(err.message || "Failed to create community. Try another slug.");
    } finally {
      setIsCreating(false);
    }
  };

  return (
    <main
      className={cn(
        "w-full sm:w-[600px] border-r border-border min-h-screen flex flex-col bg-background text-foreground",
        !currentUser && "border-l"
      )}
    >
      {/* Top Header */}
      <header className="sticky top-0 z-20 bg-background/80 backdrop-blur-md border-b border-border px-4 py-3 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-full bg-brand/10 text-brand flex items-center justify-center">
            <Users className="w-5 h-5" />
          </div>
          <div>
            <h1 className="text-xl font-bold leading-tight">Communities</h1>
            <p className="text-xs text-muted-foreground">Discover and share topics you love</p>
          </div>
        </div>

        {currentUser && (
          <button
            onClick={() => setIsModalOpen(true)}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-foreground text-background text-sm font-bold hover:opacity-90 active:scale-95 transition-all shadow"
          >
            <Plus className="w-4 h-4" />
            <span>Create</span>
          </button>
        )}
      </header>

      {/* Tabs */}
      <div className="flex border-b border-border bg-background">
        <button
          onClick={() => setActiveTab("discover")}
          className={cn(
            "flex-1 py-3 text-sm font-bold transition-colors relative text-center",
            activeTab === "discover" ? "text-foreground" : "text-muted-foreground hover:bg-subtle/50"
          )}
        >
          Discover
          {activeTab === "discover" && (
            <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-12 h-1 bg-brand rounded-full" />
          )}
        </button>

        {currentUser && (
          <button
            onClick={() => setActiveTab("my")}
            className={cn(
              "flex-1 py-3 text-sm font-bold transition-colors relative text-center",
              activeTab === "my" ? "text-foreground" : "text-muted-foreground hover:bg-subtle/50"
            )}
          >
            My Communities ({myCommunities.length})
            {activeTab === "my" && (
              <div className="absolute bottom-0 left-1/2 -translate-x-1/2 w-16 h-1 bg-brand rounded-full" />
            )}
          </button>
        )}
      </div>

      {/* Search Input */}
      <div className="p-4 border-b border-border">
        <div className="relative">
          <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search communities by name or topic..."
            className="w-full bg-subtle/60 border border-border/60 rounded-full pl-10 pr-4 py-2 text-sm focus:outline-none focus:border-brand focus:ring-1 focus:ring-brand placeholder:text-muted-foreground"
          />
        </div>
      </div>

      {/* Content Feed */}
      <div className="flex-1 divide-y divide-border">
        {isLoading ? (
          <div className="p-8 text-center text-muted-foreground">Loading communities...</div>
        ) : activeTab === "discover" ? (
          communities.length === 0 ? (
            <div className="p-12 text-center text-muted-foreground">
              <Users className="w-12 h-12 mx-auto mb-3 opacity-30" />
              <p className="font-semibold text-foreground">No communities found</p>
              <p className="text-sm mt-1">Be the first to create one!</p>
            </div>
          ) : (
            communities.map((c) => (
              <Link
                key={c.id}
                href={`/communities/${c.slug}`}
                className="p-4 flex items-start gap-4 hover:bg-subtle/30 transition-colors cursor-pointer group block"
              >
                <div className="w-14 h-14 rounded-2xl bg-subtle border border-border flex items-center justify-center shrink-0 overflow-hidden text-brand font-bold text-xl">
                  {c.avatar_url ? (
                    <img src={c.avatar_url} alt={c.name} className="w-full h-full object-cover" />
                  ) : (
                    c.name.slice(0, 2).toUpperCase()
                  )}
                </div>

                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <h2 className="font-bold text-base text-foreground group-hover:underline truncate">
                      {c.name}
                    </h2>
                    {c.access_model === "restricted" ? (
                      <span className="flex items-center gap-1 text-[11px] font-medium text-amber-500 bg-amber-500/10 px-2 py-0.5 rounded-full">
                        <Lock className="w-3 h-3" /> Restricted
                      </span>
                    ) : (
                      <span className="flex items-center gap-1 text-[11px] font-medium text-emerald-500 bg-emerald-500/10 px-2 py-0.5 rounded-full">
                        <Globe className="w-3 h-3" /> Open
                      </span>
                    )}
                  </div>

                  <p className="text-xs text-muted-foreground mt-0.5">
                    {c.members_count} {c.members_count === 1 ? "member" : "members"}
                  </p>

                  {c.description && (
                    <p className="text-sm text-foreground/80 mt-1.5 line-clamp-2 leading-relaxed">
                      {c.description}
                    </p>
                  )}
                </div>

                <div className="shrink-0 self-center">
                  <ArrowRight className="w-5 h-5 text-muted-foreground group-hover:text-foreground transition-transform group-hover:translate-x-0.5" />
                </div>
              </Link>
            ))
          )
        ) : (
          myCommunities.length === 0 ? (
            <div className="p-12 text-center text-muted-foreground">
              <Users className="w-12 h-12 mx-auto mb-3 opacity-30" />
              <p className="font-semibold text-foreground">You haven't joined any communities</p>
              <p className="text-sm mt-1">Explore communities above to find your interests.</p>
            </div>
          ) : (
            myCommunities.map((c) => (
              <Link
                key={c.id}
                href={`/communities/${c.slug}`}
                className="p-4 flex items-start gap-4 hover:bg-subtle/30 transition-colors cursor-pointer group block"
              >
                <div className="w-14 h-14 rounded-2xl bg-subtle border border-border flex items-center justify-center shrink-0 overflow-hidden text-brand font-bold text-xl">
                  {c.avatar_url ? (
                    <img src={c.avatar_url} alt={c.name} className="w-full h-full object-cover" />
                  ) : (
                    c.name.slice(0, 2).toUpperCase()
                  )}
                </div>

                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <h2 className="font-bold text-base text-foreground group-hover:underline truncate">
                      {c.name}
                    </h2>
                    {c.viewer_role && (
                      <span className="text-[10px] uppercase font-bold text-brand bg-brand/10 px-2 py-0.5 rounded-full">
                        {c.viewer_role}
                      </span>
                    )}
                  </div>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    {c.members_count} {c.members_count === 1 ? "member" : "members"}
                  </p>
                  {c.description && (
                    <p className="text-sm text-foreground/80 mt-1.5 line-clamp-2">
                      {c.description}
                    </p>
                  )}
                </div>

                <div className="shrink-0 self-center">
                  <ArrowRight className="w-5 h-5 text-muted-foreground group-hover:text-foreground transition-transform group-hover:translate-x-0.5" />
                </div>
              </Link>
            ))
          )
        )}
      </div>

      {/* Create Community Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
          <div className="w-full max-w-lg bg-card border border-border rounded-2xl p-6 shadow-2xl relative animate-in fade-in zoom-in-95 duration-150">
            <h2 className="text-xl font-bold mb-1">Create a Community</h2>
            <p className="text-xs text-muted-foreground mb-4">
              Bring people together around shared interests, passions, or knowledge.
            </p>

            {errorMsg && (
              <div className="mb-4 p-3 bg-red-500/10 border border-red-500/20 text-red-400 text-xs rounded-xl">
                {errorMsg}
              </div>
            )}

            <form onSubmit={handleCreateSubmit} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  Community Name *
                </label>
                <input
                  type="text"
                  value={createName}
                  onChange={(e) => {
                    setCreateName(e.target.value);
                    if (!createSlug) {
                      setCreateSlug(e.target.value.toLowerCase().replace(/[^a-z0-9_-]/g, ""));
                    }
                  }}
                  placeholder="e.g. Lost Media Research"
                  maxLength={50}
                  className="w-full bg-subtle border border-border rounded-xl px-3.5 py-2.5 text-sm focus:outline-none focus:border-brand placeholder:text-muted-foreground"
                  required
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  Unique Slug (URL handle) *
                </label>
                <div className="flex items-center bg-subtle border border-border rounded-xl px-3.5 py-2.5 text-sm text-muted-foreground focus-within:border-brand">
                  <span>/communities/</span>
                  <input
                    type="text"
                    value={createSlug}
                    onChange={(e) =>
                      setCreateSlug(e.target.value.toLowerCase().replace(/[^a-z0-9_-]/g, ""))
                    }
                    placeholder="lost-media"
                    maxLength={30}
                    className="flex-1 bg-transparent text-foreground focus:outline-none ml-0.5"
                    required
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  Description
                </label>
                <textarea
                  value={createDesc}
                  onChange={(e) => setCreateDesc(e.target.value)}
                  placeholder="What is this community about?"
                  rows={2}
                  maxLength={280}
                  className="w-full bg-subtle border border-border rounded-xl px-3.5 py-2 text-sm focus:outline-none focus:border-brand placeholder:text-muted-foreground resize-none"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  Membership Access Model
                </label>
                <div className="grid grid-cols-2 gap-3">
                  <button
                    type="button"
                    onClick={() => setCreateAccess("open")}
                    className={cn(
                      "p-3 rounded-xl border text-left flex flex-col gap-1 transition-all",
                      createAccess === "open"
                        ? "border-brand bg-brand/5 ring-1 ring-brand text-foreground"
                        : "border-border bg-subtle/50 text-muted-foreground hover:bg-subtle"
                    )}
                  >
                    <div className="flex items-center gap-1.5 font-bold text-sm">
                      <Globe className="w-4 h-4 text-emerald-500" />
                      <span>Open</span>
                    </div>
                    <span className="text-[11px] leading-tight">Anyone can view and join immediately.</span>
                  </button>

                  <button
                    type="button"
                    onClick={() => setCreateAccess("restricted")}
                    className={cn(
                      "p-3 rounded-xl border text-left flex flex-col gap-1 transition-all",
                      createAccess === "restricted"
                        ? "border-brand bg-brand/5 ring-1 ring-brand text-foreground"
                        : "border-border bg-subtle/50 text-muted-foreground hover:bg-subtle"
                    )}
                  >
                    <div className="flex items-center gap-1.5 font-bold text-sm">
                      <Lock className="w-4 h-4 text-amber-500" />
                      <span>Restricted</span>
                    </div>
                    <span className="text-[11px] leading-tight">Requires moderator review to join.</span>
                  </button>
                </div>
              </div>

              <div>
                <label className="block text-xs font-semibold text-muted-foreground mb-1">
                  First Rule (Optional)
                </label>
                <input
                  type="text"
                  value={createRuleTitle}
                  onChange={(e) => setCreateRuleTitle(e.target.value)}
                  placeholder="e.g. Be respectful and constructive"
                  maxLength={100}
                  className="w-full bg-subtle border border-border rounded-xl px-3.5 py-2 text-sm focus:outline-none focus:border-brand placeholder:text-muted-foreground"
                />
              </div>

              <div className="flex items-center justify-end gap-2.5 pt-3 border-t border-border">
                <button
                  type="button"
                  onClick={() => setIsModalOpen(false)}
                  disabled={isCreating}
                  className="px-4 py-2 rounded-full border border-border text-sm font-semibold hover:bg-subtle transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isCreating}
                  className="px-5 py-2 rounded-full bg-brand text-white text-sm font-bold hover:bg-brand-hover active:scale-95 transition-all shadow disabled:opacity-50"
                >
                  {isCreating ? "Creating..." : "Create Community"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </main>
  );
}
