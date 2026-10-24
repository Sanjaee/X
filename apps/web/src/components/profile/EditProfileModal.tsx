"use client";

import React, { useState, useRef } from "react";
import { X, Camera, Sparkles, ChevronRight, Loader2 } from "lucide-react";
import { User } from "@/types";
import { apiRequest } from "@/lib/api";

interface EditProfileModalProps {
  isOpen: boolean;
  onClose: () => void;
  currentUser: User;
  onSuccess: (updatedUser: User) => void;
}

export function EditProfileModal({
  isOpen,
  onClose,
  currentUser,
  onSuccess,
}: EditProfileModalProps) {
  const [displayName, setDisplayName] = useState(currentUser.display_name || "");
  const [bio, setBio] = useState(currentUser.bio || "");
  const [location, setLocation] = useState(currentUser.location || "");
  const [website, setWebsite] = useState(currentUser.website || "");
  const [avatarUrl, setAvatarUrl] = useState(currentUser.avatar_url || "");
  const [bannerUrl, setBannerUrl] = useState(currentUser.banner_url || "");

  const [saving, setSaving] = useState(false);
  const [uploadingAvatar, setUploadingAvatar] = useState(false);
  const [uploadingBanner, setUploadingBanner] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const avatarInputRef = useRef<HTMLInputElement>(null);
  const bannerInputRef = useRef<HTMLInputElement>(null);

  if (!isOpen) return null;

  const handleFileUpload = async (file: File, isAvatar: boolean) => {
    try {
      if (isAvatar) setUploadingAvatar(true);
      else setUploadingBanner(true);
      setError(null);

      const formData = new FormData();
      formData.append("file", file);

      const token = localStorage.getItem("x_session_token");
      const res = await fetch("/api/v1/media/upload", {
        method: "POST",
        headers: token ? { Authorization: `Bearer ${token}` } : {},
        body: formData,
      });

      if (!res.ok) {
        throw new Error("Failed to upload image");
      }

      const data = await res.json();
      const uploadedUrl = data.url || data.media_url || data.media?.url;
      if (uploadedUrl) {
        if (isAvatar) {
          setAvatarUrl(uploadedUrl);
        } else {
          setBannerUrl(uploadedUrl);
        }
      }
    } catch (err: any) {
      console.error(err);
      setError(err.message || "Failed to upload image");
    } finally {
      if (isAvatar) setUploadingAvatar(false);
      else setUploadingBanner(false);
    }
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!displayName.trim()) {
      setError("Name cannot be empty");
      return;
    }

    setSaving(true);
    setError(null);

    try {
      const updated = await apiRequest<User>("/me/profile", {
        method: "PATCH",
        body: JSON.stringify({
          display_name: displayName.trim(),
          bio: bio.trim(),
          location: location.trim(),
          website: website.trim(),
          avatar_url: avatarUrl,
          banner_url: bannerUrl,
        }),
      });

      // Update localStorage cached session user
      if (typeof window !== "undefined") {
        const stored = localStorage.getItem("x_session_user");
        if (stored) {
          try {
            const parsed = JSON.parse(stored);
            localStorage.setItem("x_session_user", JSON.stringify({ ...parsed, ...updated }));
          } catch {}
        }
      }

      onSuccess(updated);
      onClose();
    } catch (err: any) {
      console.error(err);
      setError(err.message || "Failed to save profile");
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-4 bg-black/60 backdrop-blur-sm select-none animate-in fade-in duration-200">
      <div
        className="w-full max-w-[600px] max-h-[90vh] bg-black text-foreground rounded-2xl border border-neutral-800 flex flex-col overflow-hidden shadow-2xl relative"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="h-[53px] px-4 flex items-center justify-between border-b border-neutral-800 sticky top-0 bg-black/90 backdrop-blur-md z-30">
          <div className="flex items-center gap-6">
            <button
              type="button"
              onClick={onClose}
              disabled={saving}
              className="p-1.5 -ml-1 rounded-full hover:bg-neutral-800 transition-colors text-foreground"
              aria-label="Close"
            >
              <X className="w-5 h-5" />
            </button>
            <h2 className="font-extrabold text-[20px] text-foreground">Edit profile</h2>
          </div>

          <button
            type="button"
            onClick={handleSave}
            disabled={saving || uploadingAvatar || uploadingBanner || !displayName.trim()}
            className="bg-white hover:bg-neutral-200 text-black font-bold text-[14px] px-5 py-1.5 rounded-full transition-all active:scale-95 disabled:opacity-50 flex items-center gap-2"
          >
            {saving && <Loader2 className="w-3.5 h-3.5 animate-spin text-black" />}
            <span>Save</span>
          </button>
        </div>

        {/* Scrollable Form Body */}
        <div className="flex-1 overflow-y-auto no-scrollbar">
          {error && (
            <div className="bg-red-500/10 border-b border-red-500/20 px-4 py-2 text-red-500 text-xs">
              {error}
            </div>
          )}

          {/* Banner Area */}
          <div className="h-[180px] sm:h-[200px] w-full bg-neutral-800 relative flex items-center justify-center group">
            {bannerUrl ? (
              <img
                src={bannerUrl}
                alt="Banner"
                className="w-full h-full object-cover"
              />
            ) : (
              <div className="w-full h-full bg-[#1e2732]" />
            )}

            <div className="absolute inset-0 bg-black/30 flex items-center justify-center gap-4">
              <button
                type="button"
                onClick={() => bannerInputRef.current?.click()}
                className="p-3 bg-black/60 hover:bg-black/80 rounded-full text-white transition-transform active:scale-95"
                title="Add banner photo"
              >
                {uploadingBanner ? (
                  <Loader2 className="w-5 h-5 animate-spin" />
                ) : (
                  <Camera className="w-5 h-5" />
                )}
              </button>

              {bannerUrl && (
                <button
                  type="button"
                  onClick={() => setBannerUrl("")}
                  className="p-3 bg-black/60 hover:bg-black/80 rounded-full text-white transition-transform active:scale-95"
                  title="Remove banner photo"
                >
                  <X className="w-5 h-5" />
                </button>
              )}
            </div>

            <input
              ref={bannerInputRef}
              type="file"
              accept="image/*"
              className="hidden"
              onChange={(e) => {
                const f = e.target.files?.[0];
                if (f) handleFileUpload(f, false);
              }}
            />
          </div>

          {/* Avatar and Imagine Banner */}
          <div className="px-4 pb-4">
            <div className="flex items-end justify-between -mt-16 sm:-mt-20 mb-3 gap-3">
              {/* Circular Avatar */}
              <div className="relative w-[96px] h-[96px] sm:w-[112px] sm:h-[112px] rounded-full border-4 border-black bg-neutral-900 overflow-hidden group shrink-0">
                <img
                  src={
                    avatarUrl ||
                    currentUser.avatar_url ||
                    "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"
                  }
                  alt={displayName}
                  className="w-full h-full object-cover"
                />

                <div className="absolute inset-0 bg-black/40 flex items-center justify-center">
                  <button
                    type="button"
                    onClick={() => avatarInputRef.current?.click()}
                    className="p-2.5 bg-black/60 hover:bg-black/80 rounded-full text-white transition-transform active:scale-95"
                    title="Change profile photo"
                  >
                    {uploadingAvatar ? (
                      <Loader2 className="w-5 h-5 animate-spin" />
                    ) : (
                      <Camera className="w-5 h-5" />
                    )}
                  </button>
                </div>

                <input
                  ref={avatarInputRef}
                  type="file"
                  accept="image/*"
                  className="hidden"
                  onChange={(e) => {
                    const f = e.target.files?.[0];
                    if (f) handleFileUpload(f, true);
                  }}
                />
              </div>
            </div>

            {/* Form Fields: Floating label style matching X reference */}
            <form onSubmit={handleSave} className="space-y-4 pt-1">
              {/* Name Field */}
              <div className="relative border border-neutral-800 focus-within:border-brand rounded-lg p-2.5 transition-colors bg-transparent">
                <label className="block text-[12px] text-muted font-medium leading-none mb-1">
                  Name
                </label>
                <input
                  type="text"
                  required
                  maxLength={50}
                  value={displayName}
                  onChange={(e) => setDisplayName(e.target.value)}
                  placeholder="Name"
                  className="w-full bg-transparent text-[16px] text-foreground outline-none border-none p-0"
                />
              </div>

              {/* Bio Field */}
              <div className="relative border border-neutral-800 focus-within:border-brand rounded-lg p-2.5 transition-colors bg-transparent">
                <label className="block text-[12px] text-muted font-medium leading-none mb-1">
                  Bio
                </label>
                <textarea
                  rows={3}
                  maxLength={160}
                  value={bio}
                  onChange={(e) => setBio(e.target.value)}
                  placeholder="Bio"
                  className="w-full bg-transparent text-[16px] text-foreground outline-none border-none p-0 resize-none"
                />
              </div>

              {/* Location Field */}
              <div className="relative border border-neutral-800 focus-within:border-brand rounded-lg p-2.5 transition-colors bg-transparent">
                <label className="block text-[12px] text-muted font-medium leading-none mb-1">
                  Location
                </label>
                <input
                  type="text"
                  maxLength={100}
                  value={location}
                  onChange={(e) => setLocation(e.target.value)}
                  placeholder="Location"
                  className="w-full bg-transparent text-[16px] text-foreground outline-none border-none p-0"
                />
              </div>

              {/* Website Field */}
              <div className="relative border border-neutral-800 focus-within:border-brand rounded-lg p-2.5 transition-colors bg-transparent">
                <label className="block text-[12px] text-muted font-medium leading-none mb-1">
                  Website
                </label>
                <input
                  type="text"
                  maxLength={100}
                  value={website}
                  onChange={(e) => setWebsite(e.target.value)}
                  placeholder="Website"
                  className="w-full bg-transparent text-[16px] text-foreground outline-none border-none p-0"
                />
              </div>

              {/* Birth date row (matching reference screenshot) */}
              <div className="py-2.5 px-1 flex items-center justify-between text-muted hover:text-foreground cursor-pointer transition-colors border-t border-neutral-900">
                <div>
                  <div className="text-[15px] font-bold text-foreground leading-tight">Birth date</div>
                  <div className="text-[13px] text-muted leading-tight mt-0.5">October 18, 2000</div>
                </div>
                <ChevronRight className="w-4 h-4 text-muted" />
              </div>

              {/* Switch to professional row (matching reference screenshot) */}
              <div className="py-2.5 px-1 flex items-center justify-between text-muted hover:text-foreground cursor-pointer transition-colors border-t border-neutral-900 pb-2">
                <div className="text-[15px] font-bold text-foreground">Switch to professional</div>
                <ChevronRight className="w-4 h-4 text-muted" />
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>
  );
}
