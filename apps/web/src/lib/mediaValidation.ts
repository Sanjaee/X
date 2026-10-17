/**
 * Standard X-compliant media formats
 * Supported extensions matching screenshot:
 * .pjp, .jfif, .jpe, .pjpeg, .jpeg, .jpg, .png, .webp, .gif, .m4v, .mp4, .mov
 */

export const X_ALLOWED_MEDIA_EXTENSIONS = [
  ".pjp",
  ".jfif",
  ".jpe",
  ".pjpeg",
  ".jpeg",
  ".jpg",
  ".png",
  ".webp",
  ".gif",
  ".m4v",
  ".mp4",
  ".mov",
] as const;

export const X_ALLOWED_MEDIA_ACCEPT = [
  ".pjp",
  ".jfif",
  ".jpe",
  ".pjpeg",
  ".jpeg",
  ".jpg",
  ".png",
  ".webp",
  ".gif",
  ".m4v",
  ".mp4",
  ".mov",
  "image/jpeg",
  "image/png",
  "image/webp",
  "image/gif",
  "image/pjpeg",
  "video/mp4",
  "video/quicktime",
  "video/x-m4v",
].join(",");

export function isValidMediaFile(file: File): boolean {
  const name = file.name.toLowerCase();
  const hasValidExt = X_ALLOWED_MEDIA_EXTENSIONS.some((ext) => name.endsWith(ext));
  if (hasValidExt) return true;

  // Fallback to MIME check
  const type = file.type.toLowerCase();
  return (
    type.startsWith("image/") ||
    type === "video/mp4" ||
    type === "video/quicktime" ||
    type === "video/x-m4v"
  );
}

export function filterValidMediaFiles(files: File[] | FileList): {
  valid: File[];
  rejected: File[];
  validFiles: File[];
  rejectedCount: number;
} {
  const fileArray = Array.from(files);
  const valid: File[] = [];
  const rejected: File[] = [];

  for (const f of fileArray) {
    if (isValidMediaFile(f)) {
      valid.push(f);
    } else {
      rejected.push(f);
    }
  }

  return {
    valid,
    rejected,
    validFiles: valid,
    rejectedCount: rejected.length,
  };
}
