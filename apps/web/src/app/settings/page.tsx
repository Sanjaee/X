"use client";

import React, { useState } from "react";
import Link from "next/link";
import {
  Settings as SettingsIcon,
  Moon,
  Sun,
  Shield,
  KeyRound,
  User,
  Bell,
  Check,
  ChevronRight,
  Monitor,
} from "lucide-react";
import { LeftNavRail } from "@/components/layout/LeftNavRail";
import { RightSidebar } from "@/components/layout/RightSidebar";
import { useCurrentUser } from "@/lib/useCurrentUser";
import { MobileBottomNav } from "@/components/layout/MobileBottomNav";
import { MobileDrawer } from "@/components/layout/MobileDrawer";
import { MobileComposerModal } from "@/components/composer/MobileComposerModal";
import { AuthModal } from "@/components/auth/AuthModal";
import { cn } from "@/lib/utils";
import { apiRequest } from "@/lib/api";

type SettingsSection = "display" | "account" | "privacy" | "notifications";

export default function SettingsPage() {
  const { currentUser, updateUser, logout: handleLogout } = useCurrentUser();
  const [activeSection, setActiveSection] = useState<SettingsSection>("display");
  const [authModalOpen, setAuthModalOpen] = useState(false);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [composerModalOpen, setComposerModalOpen] = useState(false);

  // Display Theme state
  const [theme, setTheme] = useState<"dark" | "dim" | "light">(() => {
    if (typeof window !== "undefined") {
      const saved = localStorage.getItem("x_theme");
      if (saved === "dim" || saved === "light" || saved === "dark") return saved;
    }
    return "dark";
  });

  // Account form state
  const [displayName, setDisplayName] = useState(currentUser?.display_name || "");
  const [bio, setBio] = useState(currentUser?.bio || "");
  const [location, setLocation] = useState(currentUser?.location || "");
  const [website, setWebsite] = useState(currentUser?.website || "");
  const [savingAccount, setSavingAccount] = useState(false);
  const [saveSuccess, setSaveSuccess] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);

  const handleThemeChange = (newTheme: "dark" | "dim" | "light") => {
    setTheme(newTheme);
    if (typeof window !== "undefined") {
      localStorage.setItem("x_theme", newTheme);
      document.documentElement.classList.remove("light", "dim", "dark");
      if (newTheme === "light") {
        document.documentElement.classList.add("light");
      } else if (newTheme === "dim") {
        document.documentElement.classList.add("dim");
      } else {
        document.documentElement.classList.add("dark");
      }
    }
  };

  const handleSaveProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!currentUser) return;
    setSavingAccount(true);
    setSaveError(null);
    setSaveSuccess(false);

    try {
      const updated = await apiRequest<{ user?: typeof currentUser } | typeof currentUser>("/me/profile", {
        method: "PATCH",
        body: JSON.stringify({
          display_name: displayName.trim(),
          bio: bio.trim(),
          location: location.trim(),
          website: website.trim(),
        }),
      });
      const resolved = (updated as any).user || updated;
      updateUser(resolved);
      setSaveSuccess(true);
      setTimeout(() => setSaveSuccess(false), 3000);
    } catch (err: any) {
      setSaveError(err.message || "Failed to update profile");
    } finally {
      setSavingAccount(false);
    }
  };

  return (
    <>
      <main className="w-full sm:w-[600px] border-r border-border min-h-screen flex flex-col pb-16 sm:pb-0">
          <header className="sticky top-0 bg-background/80 backdrop-blur-md border-b border-border z-20 flex items-center px-4 h-[53px]">
            <h1 className="font-extrabold text-[20px] leading-6">Settings</h1>
          </header>

          <div className="flex flex-col">
            {/* Section tabs */}
            <div className="flex border-b border-border bg-subtle/10 overflow-x-auto no-scrollbar">
              {[
                { id: "display", label: "Display", icon: Moon },
                { id: "account", label: "Your Account", icon: User },
                { id: "privacy", label: "Privacy & Safety", icon: Shield },
                { id: "notifications", label: "Notifications", icon: Bell },
              ].map((s) => {
                const Icon = s.icon;
                const isActive = activeSection === s.id;
                return (
                  <button
                    key={s.id}
                    onClick={() => setActiveSection(s.id as SettingsSection)}
                    className={cn(
                      "flex items-center gap-2 px-4 py-3 text-[14px] font-bold border-b-2 whitespace-nowrap transition-colors",
                      isActive
                        ? "border-brand text-foreground"
                        : "border-transparent text-muted hover:text-foreground hover:bg-subtle/40"
                    )}
                  >
                    <Icon className="w-4 h-4" />
                    {s.label}
                  </button>
                );
              })}
            </div>

            {/* Display Section */}
            {activeSection === "display" && (
              <div className="p-4 sm:p-6 flex flex-col gap-6">
                <div>
                  <h2 className="text-[20px] font-extrabold mb-1">Customize your view</h2>
                  <p className="text-muted text-[14px]">
                    Manage your background and color theme. These settings affect all X accounts on this browser.
                  </p>
                </div>

                {/* Theme Selector */}
                <div className="flex flex-col gap-3">
                  <span className="text-[14px] font-bold text-muted uppercase tracking-wider">Background</span>
                  <div className="grid grid-cols-3 gap-3">
                    {[
                      { id: "dark", label: "Lights out", bg: "bg-black border-border", text: "text-white" },
                      { id: "dim", label: "Dim", bg: "bg-[#15202b] border-[#38444d]", text: "text-white" },
                      { id: "light", label: "Default", bg: "bg-white border-gray-200", text: "text-black" },
                    ].map((opt) => (
                      <button
                        key={opt.id}
                        type="button"
                        onClick={() => handleThemeChange(opt.id as any)}
                        className={cn(
                          "flex items-center justify-between p-3.5 rounded-xl border-2 transition-all font-bold text-[14px]",
                          opt.bg,
                          opt.text,
                          theme === opt.id ? "border-brand ring-2 ring-brand/30" : "opacity-80 hover:opacity-100"
                        )}
                      >
                        <span>{opt.label}</span>
                        {theme === opt.id && (
                          <div className="w-5 h-5 rounded-full bg-brand flex items-center justify-center text-white">
                            <Check className="w-3.5 h-3.5 stroke-[3]" />
                          </div>
                        )}
                      </button>
                    ))}
                  </div>
                </div>
              </div>
            )}

            {/* Account Section */}
            {activeSection === "account" && (
              <div className="p-4 sm:p-6 flex flex-col gap-6">
                <div>
                  <h2 className="text-[20px] font-extrabold mb-1">Account Information</h2>
                  <p className="text-muted text-[14px]">
                    Update your public display name, biography, location, and website.
                  </p>
                </div>

                {!currentUser ? (
                  <div className="p-6 border border-border rounded-2xl text-center">
                    <p className="text-muted mb-4 text-[14px]">You need to be logged in to edit account details.</p>
                    <button
                      onClick={() => setAuthModalOpen(true)}
                      className="px-6 py-2 bg-brand text-white font-bold rounded-full hover:bg-brand-hover text-[14px]"
                    >
                      Log in
                    </button>
                  </div>
                ) : (
                  <form onSubmit={handleSaveProfile} className="flex flex-col gap-4 max-w-md">
                    {saveSuccess && (
                      <div className="p-3 bg-green-500/10 border border-green-500/20 text-green-500 rounded-xl text-sm font-semibold flex items-center gap-2">
                        <Check className="w-4 h-4" /> Profile updated successfully!
                      </div>
                    )}
                    {saveError && (
                      <div className="p-3 bg-red-500/10 border border-red-500/20 text-red-500 rounded-xl text-sm font-semibold">
                        {saveError}
                      </div>
                    )}

                    <div>
                      <label className="block text-xs font-bold text-muted mb-1 uppercase tracking-wider">Display Name</label>
                      <input
                        type="text"
                        value={displayName}
                        onChange={(e) => setDisplayName(e.target.value)}
                        className="w-full bg-transparent border border-border rounded-xl px-4 py-2.5 text-[15px] focus:border-brand outline-none text-foreground"
                      />
                    </div>

                    <div>
                      <label className="block text-xs font-bold text-muted mb-1 uppercase tracking-wider">Bio</label>
                      <textarea
                        rows={3}
                        value={bio}
                        onChange={(e) => setBio(e.target.value)}
                        className="w-full bg-transparent border border-border rounded-xl px-4 py-2.5 text-[15px] focus:border-brand outline-none text-foreground resize-none"
                      />
                    </div>

                    <div>
                      <label className="block text-xs font-bold text-muted mb-1 uppercase tracking-wider">Location</label>
                      <input
                        type="text"
                        value={location}
                        onChange={(e) => setLocation(e.target.value)}
                        className="w-full bg-transparent border border-border rounded-xl px-4 py-2.5 text-[15px] focus:border-brand outline-none text-foreground"
                      />
                    </div>

                    <div>
                      <label className="block text-xs font-bold text-muted mb-1 uppercase tracking-wider">Website</label>
                      <input
                        type="text"
                        value={website}
                        onChange={(e) => setWebsite(e.target.value)}
                        className="w-full bg-transparent border border-border rounded-xl px-4 py-2.5 text-[15px] focus:border-brand outline-none text-foreground"
                      />
                    </div>

                    <button
                      type="submit"
                      disabled={savingAccount}
                      className="mt-2 py-2.5 px-6 bg-foreground text-background font-bold text-[15px] rounded-full hover:opacity-90 transition-opacity disabled:opacity-50 self-start"
                    >
                      {savingAccount ? "Saving..." : "Save changes"}
                    </button>
                  </form>
                )}
              </div>
            )}

            {/* Privacy Section */}
            {activeSection === "privacy" && (
              <div className="p-4 sm:p-6 flex flex-col gap-4">
                <h2 className="text-[20px] font-extrabold mb-1">Privacy and Safety</h2>
                <p className="text-muted text-[14px]">
                  Manage what information you see and share on X.
                </p>

                <div className="divide-y divide-border border-y border-border">
                  <div className="py-4 flex items-center justify-between">
                    <div>
                      <h4 className="font-bold text-[15px]">Audience and tagging</h4>
                      <p className="text-xs text-muted">Manage who can see your posts and media tags</p>
                    </div>
                    <ChevronRight className="w-5 h-5 text-muted" />
                  </div>
                  <div className="py-4 flex items-center justify-between">
                    <div>
                      <h4 className="font-bold text-[15px]">Mute and block</h4>
                      <p className="text-xs text-muted">Manage accounts and notifications you’ve muted or blocked</p>
                    </div>
                    <ChevronRight className="w-5 h-5 text-muted" />
                  </div>
                  <div className="py-4 flex items-center justify-between">
                    <div>
                      <h4 className="font-bold text-[15px]">Direct Messages</h4>
                      <p className="text-xs text-muted">Manage who can message you directly</p>
                    </div>
                    <ChevronRight className="w-5 h-5 text-muted" />
                  </div>
                </div>
              </div>
            )}

            {/* Notifications Section */}
            {activeSection === "notifications" && (
              <div className="p-4 sm:p-6 flex flex-col gap-4">
                <h2 className="text-[20px] font-extrabold mb-1">Notification Preferences</h2>
                <p className="text-muted text-[14px]">
                  Select the kinds of notifications you get about your activities, interests, and recommendations.
                </p>
                <div className="divide-y divide-border border-y border-border">
                  <div className="py-4 flex items-center justify-between">
                    <div>
                      <h4 className="font-bold text-[15px]">Filters</h4>
                      <p className="text-xs text-muted">Choose the notifications you’d like to see — and those you don’t</p>
                    </div>
                    <ChevronRight className="w-5 h-5 text-muted" />
                  </div>
                  <div className="py-4 flex items-center justify-between">
                    <div>
                      <h4 className="font-bold text-[15px]">Preferences</h4>
                      <p className="text-xs text-muted">Select your preferences by notification type</p>
                    </div>
                    <ChevronRight className="w-5 h-5 text-muted" />
                  </div>
                </div>
              </div>
            )}
          </div>
        </main>

        <RightSidebar currentUser={currentUser} />

      <MobileBottomNav onOpenComposer={() => setComposerModalOpen(true)} />

      {drawerOpen && currentUser && (
        <MobileDrawer
          isOpen={drawerOpen}
          onClose={() => setDrawerOpen(false)}
          currentUser={currentUser}
          onOpenAuth={() => setAuthModalOpen(true)}
          onLogout={handleLogout}
        />
      )}

      {/* Mobile Composer Modal */}
      <MobileComposerModal
        isOpen={composerModalOpen}
        onClose={() => setComposerModalOpen(false)}
        currentUser={currentUser}
        onPostSuccess={() => {}}
        onOpenAuth={() => setAuthModalOpen(true)}
      />

      {authModalOpen && (
        <AuthModal
          isOpen={authModalOpen}
          onClose={() => setAuthModalOpen(false)}
          onSuccess={(u) => updateUser(u)}
        />
      )}
    </>
  );
}
