export interface User {
  id: string;
  username: string;
  email?: string;
  display_name: string;
  bio?: string;
  avatar_url: string;
  banner_url?: string;
  location?: string;
  website?: string;
  is_verified: boolean;
  verified_type: 'none' | 'blue' | 'business' | 'government';
  followers_count: number;
  following_count: number;
  posts_count?: number;
  role?: string;
  roles?: Role[];
  permissions?: string[];
  is_owner?: boolean;
  is_following?: boolean;
  is_followed_by?: boolean;
  deleted_at?: string | null;
  created_at: string;
}

export interface MediaObject {
  id: string;
  url: string;
  media_type: 'image' | 'video' | 'gif';
  width?: number;
  height?: number;
  duration_seconds?: number;
}

export interface Post {
  id: string;
  author_id: string;
  author: User;
  content: string;
  reply_to_id?: string | null;
  parent_post?: Post | null;
  quote_post_id?: string | null;
  quote_post?: Post | null;
  replies_count: number;
  reposts_count: number;
  likes_count: number;
  views_count: number;
  media?: MediaObject[];
  moderation_status?: 'normal' | 'flagged' | 'hidden' | 'removed';
  visibility?: string;
  is_liked: boolean;
  is_reposted: boolean;
  is_bookmarked: boolean;
  created_at: string;
  updated_at?: string;
  deleted_at?: string | null;
}

export interface Notification {
  id: string;
  recipient_id: string;
  actor_id: string;
  actor: User;
  type: 'like' | 'reply' | 'repost' | 'mention' | 'follow' | 'follow_request' | 'follow_accept' | 'dm';
  post_id?: string;
  post?: Post;
  entity_type?: string;
  entity_id?: string;
  is_read: boolean;
  created_at: string;
}

export interface SocialRelationship {
  target_user_id: string;
  is_following: boolean;
  is_followed_by: boolean;
  is_mutual: boolean;
  is_blocked: boolean;
  is_blocked_by: boolean;
  is_muted: boolean;
  has_pending_follow_request: boolean;
  is_private: boolean;
}

export interface FollowRequest {
  id: string;
  requester_id: string;
  requester?: User;
  target_id: string;
  target?: User;
  status: 'pending' | 'accepted' | 'rejected' | 'cancelled';
  created_at: string;
}

export interface Conversation {
  id: string;
  is_group: boolean;
  title?: string;
  other_member?: User;
  last_message?: Message;
  unread_count?: number;
  created_at: string;
  updated_at: string;
}

export interface Message {
  id: string;
  conversation_id: string;
  sender_id: string;
  sender?: User;
  client_message_id?: string;
  content: string;
  message_type: 'text' | 'media';
  media?: MediaObject;
  reply_to_id?: string;
  created_at: string;
}

export interface AuthResponse {
  user: User;
  session_token: string;
}

export interface SubscriptionPlan {
  id: string;
  name: string;
  description: string;
  monthly_price_cents: number;
  annual_price_cents: number;
  currency: string;
  max_post_length: number;
  can_get_verified_badge: boolean;
  can_edit_posts: boolean;
  can_undo_posts: boolean;
  max_video_minutes: number;
  max_media_count: number;
  reduced_ads: boolean;
  is_active: boolean;
}

export interface Subscription {
  id: string;
  user_id: string;
  plan_id: string;
  plan?: SubscriptionPlan;
  billing_cycle: 'monthly' | 'annual';
  status: 'pending' | 'active' | 'past_due' | 'canceled' | 'expired';
  current_period_start: string;
  current_period_end: string;
  cancel_at_period_end: boolean;
}

export interface Entitlements {
  plan_id: string;
  max_post_length: number;
  has_verified_badge: boolean;
  badge_type: string;
  can_edit_posts: boolean;
  can_undo_posts: boolean;
  max_video_minutes: number;
  max_media_count: number;
  reduced_ads: boolean;
}

export interface VerificationBadge {
  id: string;
  user_id: string;
  badge_type: 'blue' | 'gold' | 'grey' | 'staff' | string;
  display_label: string;
  description: string;
  color_token: string;
  status: string;
}

export interface Role {
  id: string;
  key: string;
  name: string;
  description: string;
  is_system: boolean;
  is_active: boolean;
  user_count?: number;
  created_at: string;
  updated_at: string;
}

export interface Permission {
  id: string;
  key: string;
  name: string;
  description: string;
  created_at: string;
}

export interface MembershipPlan {
  id: string;
  key: string;
  name: string;
  description: string;
  price_amount_cents: number;
  currency: string;
  billing_interval: string;
  is_active: boolean;
  entitlements?: MembershipEntitlement[];
  created_at: string;
  updated_at: string;
}

export interface MembershipEntitlement {
  id: string;
  plan_id: string;
  entitlement_key: string;
  limit_value: number;
  unit: string;
  created_at: string;
  updated_at: string;
}

export interface UserMembership {
  id: string;
  user_id: string;
  plan_id: string;
  plan?: MembershipPlan;
  status: string;
  starts_at: string;
  expires_at?: string;
  created_at: string;
  updated_at: string;
}

export interface AuditLog {
  id: string;
  actor_user_id?: string;
  actor_username?: string;
  action: string;
  target_type: string;
  target_id: string;
  reason: string;
  metadata_redacted: Record<string, unknown>;
  created_at: string;
}

export interface SystemHealth {
  database: {
    status: string;
    ping_ms: number;
    open_connections: number;
    in_use_connections: number;
    idle_connections: number;
  };
  redis: {
    status: string;
    ping_ms: number;
    used_memory_human: string;
  };
  rabbitmq: {
    status: string;
    message?: string;
  };
  runtime: {
    goroutines: number;
    alloc_bytes: number;
    total_alloc_bytes: number;
    sys_bytes: number;
    num_gc: number;
    uptime_seconds: number;
  };
}

export interface AdminOverview {
  total_users: number;
  active_users_24h: number;
  new_users_today: number;
  total_posts: number;
  posts_today: number;
  total_media_count: number;
  estimated_storage_bytes: number;
  flagged_posts_count: number;
  system_health: SystemHealth;
  recent_audit_logs: AuditLog[];
  recent_security_events: AuditLog[];
}


