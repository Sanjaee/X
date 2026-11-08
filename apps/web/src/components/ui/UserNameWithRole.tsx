"use client";

import React from "react";
import { getRoleBadge, getRoleNameClass } from "@/utils/roleStyles";
import { CustomNameStyle } from "@/types";

export interface UserNameWithRoleProps {
  /** Display name (username or full_name) */
  displayName: string;
  /** Role or user_type from backend */
  role?: string | null;
  /** Custom JSON string or object for user-defined name appearance */
  nameStyle?: string | CustomNameStyle | null;
  /** Optional class for wrapper (e.g. font-semibold text-sm) */
  className?: string;
}

export function parseCustomNameStyle(
  raw?: string | CustomNameStyle | null
): CustomNameStyle | null {
  if (!raw) return null;
  if (typeof raw === "object") {
    if (raw.enabled === false) return null;
    return raw;
  }
  const trimmed = raw.trim();
  if (!trimmed) return null;
  try {
    const parsed = JSON.parse(trimmed) as CustomNameStyle;
    if (!parsed || typeof parsed !== "object" || parsed.enabled === false) {
      return null;
    }
    return parsed;
  } catch {
    return null;
  }
}

export function buildCustomNameInlineStyle(
  styleObj: CustomNameStyle | null
): React.CSSProperties | undefined {
  if (!styleObj) return undefined;

  const hasCustomColor = Boolean(styleObj.color && styleObj.color.trim());
  const hasShadow = Boolean(
    (styleObj.text_shadow && styleObj.text_shadow !== "none") ||
      (styleObj.shadow_color && styleObj.shadow_color !== "none")
  );
  const hasBgImage = Boolean(
    styleObj.bg_image && styleObj.bg_image !== "none" && styleObj.bg_image.trim() !== ""
  );
  const hasBgColor = Boolean(
    styleObj.bg_color &&
      styleObj.bg_color !== "transparent" &&
      styleObj.bg_color !== "none" &&
      styleObj.bg_color.trim() !== ""
  );
  const hasFontWeight = Boolean(styleObj.font_weight);
  const hasFontFamily = Boolean(styleObj.font_family);

  if (
    !hasCustomColor &&
    !hasShadow &&
    !hasBgImage &&
    !hasBgColor &&
    !hasFontWeight &&
    !hasFontFamily
  ) {
    return undefined;
  }

  let textShadow = "none";
  if (styleObj.text_shadow && styleObj.text_shadow.trim() !== "") {
    textShadow = styleObj.text_shadow;
  } else if (styleObj.shadow_color && styleObj.shadow_color !== "none") {
    const x = styleObj.shadow_x ?? 2;
    const y = styleObj.shadow_y ?? 2;
    const blur = styleObj.shadow_blur ?? 6;
    textShadow = `${x}px ${y}px ${blur}px ${styleObj.shadow_color}`;
  }

  let backgroundImage = "none";
  if (hasBgImage && styleObj.bg_image) {
    backgroundImage = styleObj.bg_image.startsWith("url(")
      ? styleObj.bg_image
      : `url("${styleObj.bg_image}")`;
  }

  return {
    color: hasCustomColor ? styleObj.color : undefined,
    textShadow,
    backgroundImage,
    backgroundColor: hasBgColor ? styleObj.bg_color : "transparent",
    fontWeight: styleObj.font_weight || "800",
    fontFamily: styleObj.font_family || "Roboto, Arial, sans-serif",
    lineHeight: 1.25,
    padding: hasBgColor || hasBgImage ? "0 3px" : undefined,
    borderRadius: hasBgColor || hasBgImage ? "3px" : undefined,
    display: "inline-block",
  };
}

/**
 * Renders username with optional role badge and custom user-selected name styling
 * (color, text-shadow, background GIF, background-color, font-weight, font-family).
 */
export function UserNameWithRole({
  displayName,
  role,
  nameStyle,
  className = "",
}: UserNameWithRoleProps) {
  const r = role ?? undefined;
  const hasBadge = Boolean(r && r !== "member" && r !== "user");

  const parsedStyle = parseCustomNameStyle(nameStyle);
  const inlineStyle = buildCustomNameInlineStyle(parsedStyle);

  if (r || inlineStyle) {
    return (
      <div className={`inline-flex items-baseline gap-0.5 ${className} min-w-0 leading-snug`}>
        {hasBadge && (
          <span
            className={`${getRoleBadge(r)} shrink-0 whitespace-nowrap inline-block leading-snug`}
          />
        )}
        <span
          className={`font-semibold truncate ${inlineStyle ? "" : getRoleNameClass(r)}`}
          style={inlineStyle}
        >
          {displayName || "—"}
        </span>
      </div>
    );
  }

  return (
    <div className={`inline-flex items-baseline gap-1 ${className} min-w-0 leading-snug`}>
      <span className="font-semibold truncate">{displayName || "—"}</span>
    </div>
  );
}
