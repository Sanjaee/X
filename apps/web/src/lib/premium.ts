import { User } from "@/types";

export function checkIsPremiumUser(user: User | null | undefined): boolean {
  if (!user) return false;
  if (user.is_owner || user.role === "owner" || user.email === "afrizaahmad18@gmail.com") {
    return true;
  }
  const role = (user.role || "").toLowerCase();
  if (["vip", "mvp", "god", "mod", "admin", "owner"].includes(role)) {
    return true;
  }
  if (user.is_verified) {
    return true;
  }
  return false;
}
