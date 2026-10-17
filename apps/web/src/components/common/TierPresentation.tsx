import React from "react";
import {
  Star,
  Zap,
  Shield,
  Sparkles,
  UserRound,
  Crown,
} from "lucide-react";

export const TIER_PRESENTATION = {
  free: {
    name: "Member",
    icon: UserRound,
    description: "Standard account for core community interactions",
    badgeColor: "bg-zinc-800 text-zinc-300 border-zinc-700",
  },
  vip: {
    name: "VIP Member",
    icon: Star,
    description: "Enhanced limits, video uploads, and verified badge",
    badgeColor: "bg-blue-500/10 text-blue-400 border-blue-500/30",
  },
  mvp: {
    name: "MVP",
    icon: Zap,
    description: "Maximum posting limits, highest upload sizes, and creator tools",
    badgeColor: "bg-amber-500/10 text-amber-400 border-amber-500/30",
  },
  mod: {
    name: "Moderator",
    icon: Shield,
    description: "Content moderation and community safety",
    badgeColor: "bg-emerald-500/10 text-emerald-400 border-emerald-500/30",
  },
  god: {
    name: "God",
    icon: Sparkles,
    description: "Special administrative governance",
    badgeColor: "bg-purple-500/10 text-purple-400 border-purple-500/30",
  },
  owner: {
    name: "Platform Owner",
    icon: Crown,
    description: "Platform owner with complete governance oversight",
    badgeColor: "bg-rose-500/10 text-rose-400 border-rose-500/30",
  },
} as const;

export type TierKey = keyof typeof TIER_PRESENTATION;

interface TierBadgeProps {
  tier: string;
  showIcon?: boolean;
  className?: string;
}

export function TierBadge({ tier, showIcon = true, className = "" }: TierBadgeProps) {
  const norm = (tier || "free").toLowerCase() as TierKey;
  const config = TIER_PRESENTATION[norm] || TIER_PRESENTATION.free;
  const Icon = config.icon;

  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold border ${config.badgeColor} ${className}`}
    >
      {showIcon && <Icon className="w-3.5 h-3.5" />}
      <span>{config.name}</span>
    </span>
  );
}
