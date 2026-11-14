"use client";

import React, { useState, useRef } from "react";
import { X, Camera, Sparkles, ChevronRight, Loader2, Palette, RotateCcw, Lock } from "lucide-react";
import { User, CustomNameStyle } from "@/types";
import { apiRequest } from "@/lib/api";
import { TwoFactorManager } from "@/components/profile/TwoFactorManager";
import { UserNameWithRole, parseCustomNameStyle } from "@/components/ui/UserNameWithRole";
import { useUpgradeModal } from "@/components/subscription/UpgradeModalContext";

const NAME_STYLE_PRESETS: Array<{ label: string; style: CustomNameStyle }> = [
  {
    label: "Crimson Pink Glow",
    style: {
      enabled: true,
      color: "#a50000",
      shadow_color: "#d035c9",
      text_shadow: "2px 2px 6px #d035c9",
      bg_image: "/admin.gif",
      bg_color: "#0a0a0a",
      font_weight: "800",
      font_family: "Roboto, Arial, sans-serif",
    },
  },
  {
    label: "God Fire",
    style: {
      enabled: true,
      color: "#ff5722",
      shadow_color: "#ff5722",
      text_shadow: "0 0 8px #ff5722",
      bg_image: "/god.gif",
      bg_color: "#0a0a0a",
      font_weight: "800",
      font_family: "Roboto, Arial, sans-serif",
    },
  },
  {
    label: "MVP Sparkle",
    style: {
      enabled: true,
      color: "#55acee",
      shadow_color: "#000000",
      text_shadow: "0px 2px 3px #000000",
      bg_image: "/mvp.gif",
      bg_color: "transparent",
      font_weight: "800",
      font_family: "Roboto, Arial, sans-serif",
    },
  },
  {
    label: "VIP Matrix",
    style: {
      enabled: true,
      color: "#a855f7",
      shadow_color: "#9b30ff",
      text_shadow: "2px 2px 6px #9b30ff",
      bg_image: "/vip.gif",
      bg_color: "#0a0a0a",
      font_weight: "800",
      font_family: "Roboto, Arial, sans-serif",
    },
  },
  {
    label: "Cyber Gold",
    style: {
      enabled: true,
      color: "#ffd700",
      shadow_color: "#ff8c00",
      text_shadow: "2px 2px 6px #ff8c00",
      bg_image: "/admin.gif",
      bg_color: "#0a0a0a",
      font_weight: "900",
      font_family: "Roboto, Arial, sans-serif",
    },
  },
];

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

  const initialParsedStyle = parseCustomNameStyle(currentUser.name_style);
  const [customStyleEnabled, setCustomStyleEnabled] = useState<boolean>(!!initialParsedStyle);
  const [nameColor, setNameColor] = useState<string>(initialParsedStyle?.color || "#a50000");
  const [shadowColor, setShadowColor] = useState<string>(initialParsedStyle?.shadow_color || "#d035c9");
  const [textShadow, setTextShadow] = useState<string>(
    initialParsedStyle?.text_shadow || "2px 2px 6px #d035c9"
  );
  const [bgImage, setBgImage] = useState<string>(initialParsedStyle?.bg_image || "/admin.gif");
  const [bgColor, setBgColor] = useState<string>(initialParsedStyle?.bg_color || "#0a0a0a");
  const [fontWeight, setFontWeight] = useState<string>(initialParsedStyle?.font_weight || "800");
  const [fontFamily, setFontFamily] = useState<string>(
    initialParsedStyle?.font_family || "Roboto, Arial, sans-serif"
  );

  const currentCustomStyleObj: CustomNameStyle | null = customStyleEnabled
    ? {
        enabled: true,
        color: nameColor,
        shadow_color: shadowColor,
        text_shadow: textShadow,
        bg_image: bgImage,
        bg_color: bgColor,
        font_weight: fontWeight,
        font_family: fontFamily,
      }
    : null;

  const applyPreset = (preset: CustomNameStyle) => {
    setCustomStyleEnabled(true);
    if (preset.color) setNameColor(preset.color);
    if (preset.shadow_color) setShadowColor(preset.shadow_color);
    if (preset.text_shadow !== undefined) setTextShadow(preset.text_shadow);
    if (preset.bg_image !== undefined) setBgImage(preset.bg_image);
    if (preset.bg_color !== undefined) setBgColor(preset.bg_color);
    if (preset.font_weight) setFontWeight(preset.font_weight);
    if (preset.font_family) setFontFamily(preset.font_family);
  };

  const { openUpgradeModal } = useUpgradeModal();

  // Premium roles that have access to Username Style & Effects
  const userRole = (currentUser.role || "").toLowerCase();
  const isPremiumUser =
    Boolean(currentUser.is_owner) ||
    ["vip", "mvp", "god", "mod", "admin", "owner"].includes(userRole) ||
    Boolean(currentUser.is_verified);

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
      const serializedNameStyle = currentCustomStyleObj ? JSON.stringify(currentCustomStyleObj) : "";
      const updated = await apiRequest<User>("/me/profile", {
        method: "PATCH",
        body: JSON.stringify({
          display_name: displayName.trim(),
          bio: bio.trim(),
          location: location.trim(),
          website: website.trim(),
          avatar_url: avatarUrl,
          banner_url: bannerUrl,
          name_style: serializedNameStyle,
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

              {/* Username Style Customizer - Restricted to Premium Roles */}
              {isPremiumUser ? (
                <div className="border border-neutral-800 rounded-xl p-3.5 bg-neutral-950/60 space-y-3.5">
                  <div className="flex items-center justify-between gap-2">
                    <div className="flex items-center gap-2">
                      <Palette className="w-4 h-4 text-brand" />
                      <div>
                        <div className="text-[14px] font-bold text-foreground leading-tight">
                          Username Style &amp; Effects
                        </div>
                        <div className="text-[12px] text-muted leading-tight mt-0.5">
                          Customize text color, glow shadow, animated GIF background, and font regardless of role
                        </div>
                      </div>
                    </div>
                    <button
                      type="button"
                      onClick={() => setCustomStyleEnabled(!customStyleEnabled)}
                      className={`px-3 py-1 rounded-full text-xs font-bold transition-colors ${
                        customStyleEnabled
                          ? "bg-brand text-white"
                          : "bg-neutral-800 text-muted hover:text-foreground"
                      }`}
                    >
                      {customStyleEnabled ? "Custom Active" : "Role Default"}
                    </button>
                  </div>

                {/* Live Preview Box */}
                <div className="rounded-lg border border-neutral-800 bg-black px-3.5 py-2.5 flex items-center justify-between">
                  <div className="flex items-center gap-2.5 min-w-0">
                    <img
                      src={
                        avatarUrl ||
                        currentUser.avatar_url ||
                        "https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png"
                      }
                      alt={displayName}
                      className="w-8 h-8 rounded-full object-cover border border-neutral-800 shrink-0"
                    />
                    <div className="min-w-0">
                      <div className="text-[16px] leading-tight flex items-center gap-1">
                        <UserNameWithRole
                          displayName={displayName.trim() || currentUser.username}
                          role={currentUser.role}
                          nameStyle={currentCustomStyleObj}
                        />
                      </div>
                      <div className="text-[12px] text-muted">@{currentUser.username}</div>
                    </div>
                  </div>
                  <span className="text-[11px] uppercase tracking-wider text-muted font-semibold">
                    Live Preview
                  </span>
                </div>

                {/* Quick Style Presets */}
                <div>
                  <div className="flex items-center justify-between mb-1.5">
                    <span className="text-[11px] font-semibold uppercase tracking-wider text-muted">
                      Quick Presets
                    </span>
                    {customStyleEnabled && (
                      <button
                        type="button"
                        onClick={() => setCustomStyleEnabled(false)}
                        className="text-[11px] text-muted hover:text-foreground inline-flex items-center gap-1"
                      >
                        <RotateCcw className="w-3 h-3" />
                        Reset to Role Default
                      </button>
                    )}
                  </div>
                  <div className="flex flex-wrap gap-1.5">
                    {NAME_STYLE_PRESETS.map((preset) => (
                      <button
                        key={preset.label}
                        type="button"
                        onClick={() => applyPreset(preset.style)}
                        className="px-2.5 py-1 rounded-lg border border-neutral-800 hover:border-neutral-600 bg-neutral-900/90 text-[12px] font-semibold text-foreground transition-colors"
                      >
                        {preset.label}
                      </button>
                    ))}
                  </div>
                </div>

                {customStyleEnabled && (
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-1 border-t border-neutral-900">
                    {/* Text Color */}
                    <div className="border border-neutral-800 rounded-lg p-2 bg-black/50">
                      <label className="block text-[11px] text-muted font-medium mb-1">
                        Text Color (`color`)
                      </label>
                      <div className="flex items-center gap-2">
                        <input
                          type="color"
                          value={nameColor.startsWith("#") && nameColor.length === 7 ? nameColor : "#a50000"}
                          onChange={(e) => setNameColor(e.target.value)}
                          className="w-7 h-7 rounded cursor-pointer bg-transparent border border-neutral-700 shrink-0"
                        />
                        <input
                          type="text"
                          value={nameColor}
                          onChange={(e) => setNameColor(e.target.value)}
                          placeholder="#a50000"
                          className="w-full bg-transparent text-[13px] text-foreground outline-none font-mono"
                        />
                      </div>
                    </div>

                    {/* Text Shadow */}
                    <div className="border border-neutral-800 rounded-lg p-2 bg-black/50">
                      <label className="block text-[11px] text-muted font-medium mb-1">
                        Text Shadow (`text-shadow`)
                      </label>
                      <div className="flex items-center gap-2">
                        <input
                          type="color"
                          value={shadowColor.startsWith("#") && shadowColor.length === 7 ? shadowColor : "#d035c9"}
                          onChange={(e) => {
                            setShadowColor(e.target.value);
                            setTextShadow(`2px 2px 6px ${e.target.value}`);
                          }}
                          className="w-7 h-7 rounded cursor-pointer bg-transparent border border-neutral-700 shrink-0"
                          title="Pick glow color"
                        />
                        <input
                          type="text"
                          value={textShadow}
                          onChange={(e) => setTextShadow(e.target.value)}
                          placeholder="2px 2px 6px #d035c9"
                          className="w-full bg-transparent text-[13px] text-foreground outline-none font-mono"
                        />
                      </div>
                    </div>

                    {/* Animated Background GIF */}
                    <div className="border border-neutral-800 rounded-lg p-2 bg-black/50">
                      <label className="block text-[11px] text-muted font-medium mb-1">
                        Background GIF (`background-image`)
                      </label>
                      <select
                        value={
                          ["none", "/admin.gif", "/god.gif", "/mvp.gif", "/vip.gif"].includes(bgImage)
                            ? bgImage
                            : "custom"
                        }
                        onChange={(e) => {
                          if (e.target.value !== "custom") {
                            setBgImage(e.target.value);
                          }
                        }}
                        className="w-full bg-black text-[13px] text-foreground outline-none border border-neutral-800 rounded px-2 py-1 mb-1.5"
                      >
                        <option value="/admin.gif">/admin.gif (Sparkle Pink/Red)</option>
                        <option value="/god.gif">/god.gif (Fire Particles)</option>
                        <option value="/mvp.gif">/mvp.gif (Blue Sparkle)</option>
                        <option value="/vip.gif">/vip.gif (VIP Matrix)</option>
                        <option value="none">None (No GIF)</option>
                        <option value="custom">Custom URL...</option>
                      </select>
                      <input
                        type="text"
                        value={bgImage}
                        onChange={(e) => setBgImage(e.target.value)}
                        placeholder='/admin.gif or url("/admin.gif")'
                        className="w-full bg-transparent text-[12px] text-muted focus:text-foreground outline-none font-mono"
                      />
                    </div>

                    {/* Background Color */}
                    <div className="border border-neutral-800 rounded-lg p-2 bg-black/50">
                      <div className="flex items-center justify-between mb-1">
                        <label className="block text-[11px] text-muted font-medium">
                          Background Color (`background-color`)
                        </label>
                        <button
                          type="button"
                          onClick={() =>
                            setBgColor(bgColor === "transparent" ? "#0a0a0a" : "transparent")
                          }
                          className="text-[10px] text-brand hover:underline"
                        >
                          {bgColor === "transparent" ? "Use #0a0a0a" : "Set Transparent"}
                        </button>
                      </div>
                      <div className="flex items-center gap-2">
                        <input
                          type="color"
                          value={bgColor.startsWith("#") && bgColor.length === 7 ? bgColor : "#0a0a0a"}
                          onChange={(e) => setBgColor(e.target.value)}
                          className="w-7 h-7 rounded cursor-pointer bg-transparent border border-neutral-700 shrink-0"
                        />
                        <input
                          type="text"
                          value={bgColor}
                          onChange={(e) => setBgColor(e.target.value)}
                          placeholder="#0a0a0a or transparent"
                          className="w-full bg-transparent text-[13px] text-foreground outline-none font-mono"
                        />
                      </div>
                    </div>

                    {/* Font Weight */}
                    <div className="border border-neutral-800 rounded-lg p-2 bg-black/50">
                      <label className="block text-[11px] text-muted font-medium mb-1">
                        Font Weight (`font-weight`)
                      </label>
                      <select
                        value={fontWeight}
                        onChange={(e) => setFontWeight(e.target.value)}
                        className="w-full bg-black text-[13px] text-foreground outline-none border border-neutral-800 rounded px-2 py-1"
                      >
                        <option value="800">800 (Extra Bold - Default)</option>
                        <option value="900">900 (Black)</option>
                        <option value="700">700 (Bold)</option>
                        <option value="600">600 (Semi Bold)</option>
                        <option value="400">400 (Normal)</option>
                      </select>
                    </div>

                    {/* Font Family */}
                    <div className="border border-neutral-800 rounded-lg p-2 bg-black/50">
                      <label className="block text-[11px] text-muted font-medium mb-1">
                        Font Family (`font-family`)
                      </label>
                      <select
                        value={fontFamily}
                        onChange={(e) => setFontFamily(e.target.value)}
                        className="w-full bg-black text-[13px] text-foreground outline-none border border-neutral-800 rounded px-2 py-1"
                      >
                        <option value="Roboto, Arial, sans-serif">Roboto, Arial, sans-serif</option>
                        <option value="Inter, system-ui, sans-serif">Inter, system-ui, sans-serif</option>
                        <option value="'Georgia', serif">Georgia, serif</option>
                        <option value="'Courier New', monospace">Courier New, monospace</option>
                        <option value="Impact, Haettenschweiler, 'Arial Narrow Bold', sans-serif">
                          Impact, sans-serif
                        </option>
                      </select>
                    </div>
                  </div>
                )}
              </div>
            ) : (
              /* Non-Premium Upsell Card */
              <div className="border border-neutral-800/80 rounded-xl p-4 bg-neutral-950/40 flex items-center justify-between gap-3">
                <div className="flex items-center gap-3 min-w-0">
                  <div className="w-9 h-9 rounded-full bg-brand/10 border border-brand/20 flex items-center justify-center shrink-0">
                    <Lock className="w-4 h-4 text-brand" />
                  </div>
                  <div className="min-w-0">
                    <div className="text-[14px] font-bold text-foreground leading-tight">
                      Username Style &amp; Effects
                    </div>
                    <div className="text-[12px] text-muted leading-tight mt-0.5">
                      Upgrade to Premium or VIP to unlock custom colors, shadows, GIF effects, and fonts.
                    </div>
                  </div>
                </div>
                <button
                  type="button"
                  onClick={() => openUpgradeModal('vip')}
                  className="px-3.5 py-1.5 rounded-full bg-brand hover:bg-brand/90 text-white text-xs font-bold shrink-0 transition-colors"
                >
                  Upgrade
                </button>
              </div>
            )}

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

              {/* TOTP Two-Factor Authentication Setup */}
              <TwoFactorManager
                currentUser={currentUser}
                onUserUpdated={(updated) => onSuccess(updated)}
                compact
              />
            </form>
          </div>
        </div>
      </div>
    </div>
  );
}
