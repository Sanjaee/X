'use client';

import React, { useState, useEffect } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import {
  Shield,
  ShieldAlert,
  Users,
  CreditCard,
  FileText,
  Key,
  Check,
  X,
  Edit2,
  Trash2,
  Plus,
  Search,
  RefreshCw,
  Lock,
  Unlock,
  AlertTriangle,
  Sliders,
  Sparkles,
  Zap,
  Star,
  UserRound,
  Crown,
  Activity,
  Server,
  Database,
  Cpu,
  Eye,
  EyeOff,
  Flag,
  HardDrive,
  Clock,
  Radio,
  CheckCircle2,
  XCircle,
  HelpCircle,
  ChevronRight,
  ChevronLeft,
  Filter,
} from 'lucide-react';
import { TierBadge, TIER_PRESENTATION } from '@/components/common/TierPresentation';
import {
  User,
  Role,
  Permission,
  MembershipPlan,
  AuditLog,
  AdminOverview,
  SystemHealth,
  Post,
} from '@/types';
import { apiRequest } from '@/lib/api';
import { cn, formatNumber } from '@/lib/utils';
import { ConfirmModal } from '@/components/common/ConfirmModal';

type AdminTab =
  | 'overview'
  | 'users'
  | 'posts'
  | 'roles'
  | 'memberships'
  | 'security'
  | 'monitoring'
  | 'audit';

export default function AdminDashboardPage() {
  const router = useRouter();
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [authStatus, setAuthStatus] = useState<'checking' | 'authorized' | 'denied'>('checking');
  const [denialReason, setDenialReason] = useState<string>('');
  const [activeTab, setActiveTab] = useState<AdminTab>('overview');

  // Confirmation Modal state
  const [confirmDialog, setConfirmDialog] = useState<{
    isOpen: boolean;
    title: string;
    description: string;
    confirmText?: string;
    variant?: 'destructive' | 'primary';
    isLoading?: boolean;
    onConfirm: () => Promise<void> | void;
  }>({
    isOpen: false,
    title: '',
    description: '',
    onConfirm: () => {},
  });

  // Overview & Monitoring state
  const [overview, setOverview] = useState<AdminOverview | null>(null);
  const [refreshingOverview, setRefreshingOverview] = useState(false);

  // Roles state
  const [roles, setRoles] = useState<Role[]>([]);
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [selectedRole, setSelectedRole] = useState<Role | null>(null);
  const [rolePermissions, setRolePermissions] = useState<Permission[]>([]);
  const [editPermissionsModal, setEditPermissionsModal] = useState(false);
  const [selectedPermIds, setSelectedPermIds] = useState<string[]>([]);
  const [roleModalOpen, setRoleModalOpen] = useState(false);
  const [isCreatingRole, setIsCreatingRole] = useState(false);
  const [roleFormData, setRoleFormData] = useState({
    key: '',
    name: '',
    description: '',
    is_active: true,
  });

  // Membership & Entitlements state
  const [membershipPlans, setMembershipPlans] = useState<MembershipPlan[]>([]);
  const [selectedPlan, setSelectedPlan] = useState<MembershipPlan | null>(null);
  const [planEntitlementsModal, setPlanEntitlementsModal] = useState(false);
  const [entitlementsFormData, setEntitlementsFormData] = useState<Record<string, number>>({});

  // Users state
  const [adminUsers, setAdminUsers] = useState<User[]>([]);
  const [userSearch, setUserSearch] = useState('');
  const [userStatusFilter, setUserStatusFilter] = useState<'all' | 'active' | 'suspended'>('all');
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [userRoleModal, setUserRoleModal] = useState(false);
  const [userSelectedRoleIds, setUserSelectedRoleIds] = useState<string[]>([]);
  const [roleChangeReason, setRoleChangeReason] = useState('');

  // Posts state
  const [adminPosts, setAdminPosts] = useState<Post[]>([]);
  const [postSearch, setPostSearch] = useState('');
  const [postModFilter, setPostModFilter] = useState<'all' | 'normal' | 'flagged' | 'hidden' | 'removed'>('all');
  const [selectedPost, setSelectedPost] = useState<Post | null>(null);
  const [postModerateModal, setPostModerateModal] = useState(false);
  const [moderateStatus, setModerateStatus] = useState<'normal' | 'flagged' | 'hidden'>('hidden');
  const [moderateReason, setModerateReason] = useState('');

  // Security Events state
  const [securityEvents, setSecurityEvents] = useState<AuditLog[]>([]);

  // Audit Logs state
  const [auditLogs, setAuditLogs] = useState<AuditLog[]>([]);
  const [auditSearch, setAuditSearch] = useState('');

  // Action status message
  const [statusMsg, setStatusMsg] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  useEffect(() => {
    loadUserAndInit();
  }, []);

  const showStatus = (type: 'success' | 'error', text: string) => {
    setStatusMsg({ type, text });
    setTimeout(() => setStatusMsg(null), 4000);
  };

  const loadUserAndInit = async () => {
    setLoading(true);
    setAuthStatus('checking');
    try {
      const user = await apiRequest<User>('/me');
      setCurrentUser(user);

      const isPlatformOwner =
        user.is_owner ||
        user.role === 'owner' ||
        user.email === 'afrizaahmad18@gmail.com';

      if (!isPlatformOwner) {
        setAuthStatus('denied');
        setDenialReason('Access Denied: Only verified Platform Owners are authorized to access this panel.');
        setLoading(false);
        setTimeout(() => {
          router.replace('/');
        }, 1500);
        return;
      }

      setAuthStatus('authorized');

      await Promise.all([
        fetchOverview(),
        fetchRoles(),
        fetchPermissions(),
        fetchPlans(),
        fetchUsers(),
        fetchPosts(),
        fetchSecurityEvents(),
        fetchAuditLogs(),
      ]);
    } catch (err: any) {
      setAuthStatus('denied');
      setDenialReason('Authentication Required: Please sign in with an authorized Platform Owner account.');
      setTimeout(() => {
        router.replace('/');
      }, 1500);
    } finally {
      setLoading(false);
    }
  };

  const fetchOverview = async () => {
    try {
      setRefreshingOverview(true);
      const data = await apiRequest<AdminOverview>('/admin/overview');
      setOverview(data);
    } catch (err: any) {
      console.error('Failed to load admin overview', err);
    } finally {
      setRefreshingOverview(false);
    }
  };

  const fetchRoles = async () => {
    try {
      const res = await apiRequest<{ roles: Role[] }>('/admin/roles');
      setRoles(res.roles || []);
    } catch (err: any) {
      console.error('Failed to load roles', err);
    }
  };

  const fetchPermissions = async () => {
    try {
      const res = await apiRequest<{ permissions: Permission[] }>('/admin/permissions');
      setPermissions(res.permissions || []);
    } catch (err: any) {
      console.error('Failed to load permissions', err);
    }
  };

  const fetchPlans = async () => {
    try {
      const res = await apiRequest<{ plans: MembershipPlan[] }>('/admin/memberships');
      setMembershipPlans(res.plans || []);
    } catch (err: any) {
      console.error('Failed to load plans', err);
    }
  };

  const fetchUsers = async () => {
    try {
      const q = userSearch ? `&search=${encodeURIComponent(userSearch)}` : '';
      const res = await apiRequest<{ users: User[] }>(`/admin/users?limit=50${q}`);
      setAdminUsers(res.users || []);
    } catch (err: any) {
      console.error('Failed to load users', err);
    }
  };

  const fetchPosts = async () => {
    try {
      const params = new URLSearchParams();
      if (postSearch) params.set('query', postSearch);
      if (postModFilter !== 'all') params.set('moderation_status', postModFilter);
      params.set('limit', '50');

      const res = await apiRequest<{ posts: Post[]; total: number }>(`/admin/posts?${params.toString()}`);
      setAdminPosts(res.posts || []);
    } catch (err: any) {
      console.error('Failed to load admin posts', err);
    }
  };

  const fetchSecurityEvents = async () => {
    try {
      const res = await apiRequest<{ security_events: AuditLog[] }>('/admin/security/events?limit=50');
      setSecurityEvents(res.security_events || []);
    } catch (err: any) {
      console.error('Failed to load security events', err);
    }
  };

  const fetchAuditLogs = async () => {
    try {
      const res = await apiRequest<{ audit_logs: AuditLog[] }>('/admin/audit-logs?limit=50');
      setAuditLogs(res.audit_logs || []);
    } catch (err: any) {
      console.error('Failed to load audit logs', err);
    }
  };

  // Handlers for Post Moderation
  const handleOpenModeration = (post: Post) => {
    setSelectedPost(post);
    setModerateStatus((post.moderation_status as any) || 'hidden');
    setModerateReason('');
    setPostModerateModal(true);
  };

  const handleSaveModeration = async () => {
    if (!selectedPost) return;
    try {
      await apiRequest(`/admin/posts/${selectedPost.id}/moderation`, {
        method: 'PATCH',
        body: JSON.stringify({
          moderation_status: moderateStatus,
          reason: moderateReason || 'Manual administrative review',
        }),
      });
      showStatus('success', `Post moderation status set to ${moderateStatus}`);
      setPostModerateModal(false);
      fetchPosts();
      fetchOverview();
    } catch (err: any) {
      showStatus('error', err.message || 'Failed to update post moderation');
    }
  };

  const handleDeletePostAdmin = (post: Post) => {
    setConfirmDialog({
      isOpen: true,
      title: 'Delete post?',
      description: 'Are you sure you want to delete this post administratively? This action cannot be undone.',
      confirmText: 'Delete',
      variant: 'destructive',
      onConfirm: async () => {
        try {
          await apiRequest(`/admin/posts/${post.id}`, {
            method: 'DELETE',
            body: JSON.stringify({
              reason: 'Severe platform terms violation',
            }),
          });
          showStatus('success', 'Post successfully removed by administrator');
          fetchPosts();
          fetchOverview();
        } catch (err: any) {
          showStatus('error', err.message || 'Failed to delete post');
        } finally {
          setConfirmDialog((prev) => ({ ...prev, isOpen: false }));
        }
      },
    });
  };

  // Handlers for User Roles & Status
  const handleOpenUserRoles = async (user: User) => {
    setSelectedUser(user);
    setRoleChangeReason('');
    try {
      const res = await apiRequest<{ roles: Role[] }>(`/users/${user.username}/roles`);
      setUserSelectedRoleIds(res.roles.map((r) => r.id));
    } catch {
      setUserSelectedRoleIds([]);
    }
    setUserRoleModal(true);
  };

  const handleSaveUserRoles = async () => {
    if (!selectedUser) return;
    try {
      await apiRequest(`/admin/users/${selectedUser.id}/roles`, {
        method: 'PATCH',
        body: JSON.stringify({
          role_ids: userSelectedRoleIds,
          reason: roleChangeReason || 'Administrative role modification',
        }),
      });
      showStatus('success', `Updated roles for @${selectedUser.username}`);
      setUserRoleModal(false);
      fetchUsers();
      fetchOverview();
    } catch (err: any) {
      showStatus('error', err.message || 'Failed to update user roles');
    }
  };

  const handleToggleUserStatus = (user: User) => {
    const isSuspended = !!user.deleted_at;
    const newStatus = isSuspended ? 'active' : 'suspended';
    setConfirmDialog({
      isOpen: true,
      title: isSuspended ? `Reactivate @${user.username}?` : `Suspend @${user.username}?`,
      description: isSuspended
        ? `This account will be reactivated and permitted to log in and interact on X.`
        : `This account will be suspended and blocked from logging in or interacting on X.`,
      confirmText: isSuspended ? 'Reactivate' : 'Suspend',
      variant: isSuspended ? 'primary' : 'destructive',
      onConfirm: async () => {
        try {
          await apiRequest(`/admin/users/${user.id}/status`, {
            method: 'PATCH',
            body: JSON.stringify({ status: newStatus }),
          });
          showStatus('success', `Account @${user.username} is now ${newStatus}`);
          fetchUsers();
          fetchOverview();
        } catch (err: any) {
          showStatus('error', err.message || 'Failed to update user status');
        } finally {
          setConfirmDialog((prev) => ({ ...prev, isOpen: false }));
        }
      },
    });
  };

  // Handlers for Roles & Permissions
  const handleOpenPermissions = async (role: Role) => {
    setSelectedRole(role);
    try {
      const res = await apiRequest<{ permissions: Permission[] }>(`/admin/roles/${role.id}/permissions`);
      setRolePermissions(res.permissions || []);
      setSelectedPermIds(res.permissions.map((p) => p.id));
      setEditPermissionsModal(true);
    } catch (err: any) {
      showStatus('error', 'Failed to fetch role permissions');
    }
  };

  const handleSaveRolePermissions = async () => {
    if (!selectedRole) return;
    try {
      await apiRequest(`/admin/roles/${selectedRole.id}/permissions`, {
        method: 'PUT',
        body: JSON.stringify({ permission_ids: selectedPermIds }),
      });
      showStatus('success', `Updated permissions for ${selectedRole.name}`);
      setEditPermissionsModal(false);
      fetchRoles();
    } catch (err: any) {
      showStatus('error', err.message || 'Failed to update permissions');
    }
  };

  const handleOpenCreateRole = () => {
    setIsCreatingRole(true);
    setSelectedRole(null);
    setRoleFormData({ key: '', name: '', description: '', is_active: true });
    setRoleModalOpen(true);
  };

  const handleSaveRole = async () => {
    try {
      if (isCreatingRole) {
        await apiRequest('/admin/roles', {
          method: 'POST',
          body: JSON.stringify(roleFormData),
        });
        showStatus('success', `Role ${roleFormData.name} created successfully`);
      } else if (selectedRole) {
        await apiRequest(`/admin/roles/${selectedRole.id}`, {
          method: 'PATCH',
          body: JSON.stringify(roleFormData),
        });
        showStatus('success', `Role ${roleFormData.name} updated successfully`);
      }
      setRoleModalOpen(false);
      fetchRoles();
    } catch (err: any) {
      showStatus('error', err.message || 'Failed to save role');
    }
  };

  // Handlers for Memberships & Entitlements
  const handleOpenPlanEntitlements = (plan: MembershipPlan) => {
    setSelectedPlan(plan);
    const formVals: Record<string, number> = {};
    if (plan.entitlements) {
      plan.entitlements.forEach((e) => {
        formVals[e.entitlement_key] = e.limit_value;
      });
    }
    setEntitlementsFormData(formVals);
    setPlanEntitlementsModal(true);
  };

  const handleSavePlanEntitlements = async () => {
    if (!selectedPlan) return;
    try {
      const entitlements = Object.entries(entitlementsFormData).map(([key, value]) => ({
        entitlement_key: key,
        limit_value: Number(value),
        unit: 'count',
      }));

      await apiRequest(`/admin/plans/${selectedPlan.id}/entitlements`, {
        method: 'PUT',
        body: JSON.stringify({ entitlements }),
      });
      showStatus('success', `Updated entitlements for ${selectedPlan.name}`);
      setPlanEntitlementsModal(false);
      fetchPlans();
    } catch (err: any) {
      showStatus('error', err.message || 'Failed to update plan entitlements');
    }
  };

  const isOwner = currentUser?.is_owner || currentUser?.role === 'owner' || currentUser?.email === 'afrizaahmad18@gmail.com';

  if (authStatus === 'checking') {
    return (
      <div className="min-h-screen bg-background text-foreground flex flex-col items-center justify-center p-4">
        <div className="flex flex-col items-center gap-4 text-center max-w-sm animate-in fade-in">
          <div className="w-16 h-16 rounded-2xl bg-brand/10 border border-brand/20 flex items-center justify-center animate-pulse">
            <Shield className="w-8 h-8 text-brand" />
          </div>
          <div>
            <h2 className="text-base font-bold">Verifying Owner Access</h2>
            <p className="text-xs text-muted mt-1">
              Checking cryptographic permissions and role entitlements...
            </p>
          </div>
        </div>
      </div>
    );
  }

  if (authStatus === 'denied' || !isOwner) {
    return (
      <div className="min-h-screen bg-background text-foreground flex flex-col items-center justify-center p-4">
        <div className="w-full max-w-md bg-subtle/40 border border-rose-500/30 rounded-2xl p-6 sm:p-8 text-center space-y-4 shadow-2xl animate-in fade-in">
          <div className="w-14 h-14 mx-auto rounded-2xl bg-rose-500/10 border border-rose-500/30 flex items-center justify-center text-rose-500">
            <Lock className="w-7 h-7" />
          </div>
          <div>
            <span className="px-2.5 py-1 rounded-full bg-rose-500/15 text-rose-400 text-[11px] font-bold uppercase tracking-wider border border-rose-500/30 inline-block mb-2">
              403 Forbidden
            </span>
            <h1 className="text-xl font-black">Restricted Area</h1>
            <p className="text-xs text-muted mt-2">
              {denialReason || 'Access to this panel is strictly restricted to verified Platform Owners.'}
            </p>
          </div>

          <div className="pt-2">
            <Link
              href="/"
              className="inline-flex items-center justify-center px-5 py-2.5 rounded-full bg-foreground text-background font-bold text-xs hover:opacity-90 transition-opacity"
            >
              Return to Home
            </Link>
          </div>

          <p className="text-[11px] text-muted animate-pulse">
            Redirecting to home page automatically...
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background text-foreground flex flex-col">
      {/* Top Admin Sticky Navigation Bar (Full Page Layout) */}
      <header className="sticky top-0 z-40 w-full border-b border-border bg-background/95 backdrop-blur-md">
        <div className="w-full max-w-[1600px] mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Link
              href="/"
              className="flex items-center gap-2 p-1.5 -ml-1.5 rounded-full hover:bg-subtle transition-colors group"
              title="Return to Main App"
            >
              <div className="w-9 h-9 rounded-full bg-subtle border border-border flex items-center justify-center font-black group-hover:scale-105 group-hover:border-foreground/30 transition-all">
                <svg viewBox="0 0 24 24" aria-hidden="true" className="w-4 h-4 fill-current">
                  <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z" />
                </svg>
              </div>
            </Link>

            <div className="h-6 w-px bg-border hidden sm:block" />

            <div className="flex items-center gap-2.5">
              <div className="p-2 bg-brand/10 text-brand rounded-xl">
                <Shield className="w-5 h-5" />
              </div>
              <div>
                <div className="flex items-center gap-2">
                  <h1 className="text-base font-black tracking-tight leading-none">Platform Owner Admin Panel</h1>
                  <span className="px-2 py-0.5 rounded-full text-[10px] font-extrabold uppercase bg-brand/15 text-brand border border-brand/30">
                    Console
                  </span>
                </div>
                <p className="text-[11px] text-muted leading-tight mt-0.5 hidden sm:block">
                  RBAC Governance, Dynamic Limits, Posts Moderation & Security Monitoring
                </p>
              </div>
            </div>
          </div>

          <div className="flex items-center gap-2 sm:gap-3">
            <Link
              href="/"
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-full border border-border hover:bg-subtle text-xs font-semibold text-muted hover:text-foreground transition-colors"
            >
              <ChevronLeft className="w-3.5 h-3.5" />
              <span>Back to App</span>
            </Link>

            {isOwner ? (
              <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-amber-500/10 border border-amber-500/30 text-amber-500 text-xs font-bold">
                <Crown className="w-3.5 h-3.5 fill-current" />
                <span className="hidden md:inline">Verified Owner:</span>
                <span className="font-mono text-[11px]">{currentUser?.email || 'afrizaahmad18@gmail.com'}</span>
              </div>
            ) : (
              <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-subtle border border-border text-muted text-xs font-semibold">
                <ShieldAlert className="w-3.5 h-3.5" />
                <span>Scoped Admin Access</span>
              </div>
            )}

            <button
              onClick={() => {
                fetchOverview();
                fetchSecurityEvents();
                fetchAuditLogs();
                showStatus('success', 'Refreshed live data');
              }}
              disabled={refreshingOverview}
              className="p-2 hover:bg-subtle rounded-full text-muted hover:text-foreground transition-colors disabled:opacity-50 border border-border"
              title="Refresh metrics"
            >
              <RefreshCw className={cn('w-4 h-4', refreshingOverview && 'animate-spin')} />
            </button>
          </div>
        </div>
      </header>

      {/* Main Full-Width Admin Panel */}
      <main className="flex-1 w-full max-w-[1600px] mx-auto px-4 sm:px-6 lg:px-8 py-6">

          {/* Toast Status Notification */}
          {statusMsg && (
            <div
              className={cn(
                'my-4 p-3 rounded-xl border text-sm flex items-center justify-between transition-all animate-in fade-in',
                statusMsg.type === 'success'
                  ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-400'
                  : 'bg-rose-500/10 border-rose-500/30 text-rose-400'
              )}
            >
              <div className="flex items-center gap-2 font-medium">
                {statusMsg.type === 'success' ? (
                  <CheckCircle2 className="w-4 h-4" />
                ) : (
                  <AlertTriangle className="w-4 h-4" />
                )}
                <span>{statusMsg.text}</span>
              </div>
              <button onClick={() => setStatusMsg(null)} className="p-1 hover:opacity-70">
                <X className="w-4 h-4" />
              </button>
            </div>
          )}

          {/* Tab Navigation Navigation Bar */}
          <div className="flex items-center gap-1 mt-6 border-b border-border overflow-x-auto no-scrollbar pb-px text-sm font-semibold">
            {[
              { id: 'overview', label: 'Overview', icon: Activity },
              { id: 'users', label: 'Users', icon: Users },
              { id: 'posts', label: 'Post Moderation', icon: Flag },
              { id: 'roles', label: 'Roles & RBAC', icon: Key },
              { id: 'memberships', label: 'Memberships & Quotas', icon: CreditCard },
              { id: 'security', label: 'Security Events', icon: ShieldAlert },
              { id: 'monitoring', label: 'Infrastructure', icon: Server },
              { id: 'audit', label: 'Audit Trail', icon: FileText },
            ].map((tab) => {
              const Icon = tab.icon;
              const isActive = activeTab === tab.id;
              return (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id as AdminTab)}
                  className={cn(
                    'flex items-center gap-2 px-4 py-3 border-b-2 transition-all whitespace-nowrap',
                    isActive
                      ? 'border-brand text-foreground font-bold'
                      : 'border-transparent text-muted hover:text-foreground hover:bg-subtle/50'
                  )}
                >
                  <Icon className="w-4 h-4" />
                  <span>{tab.label}</span>
                </button>
              );
            })}
          </div>

          {/* ============================================================ */}
          {/* TAB 1: OVERVIEW DASHBOARD */}
          {/* ============================================================ */}
          {activeTab === 'overview' && (
            <div className="mt-6 space-y-6 animate-in fade-in">
              {/* KPI Top Cards */}
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
                <div className="p-4 rounded-2xl bg-card border border-border flex flex-col justify-between">
                  <div className="flex items-center justify-between text-muted text-xs font-bold uppercase tracking-wider">
                    <span>Total Users</span>
                    <Users className="w-4 h-4 text-brand" />
                  </div>
                  <div className="mt-3">
                    <span className="text-2xl font-black">
                      {overview ? formatNumber(overview.total_users) : '—'}
                    </span>
                    <span className="text-xs text-muted block mt-0.5">
                      +{overview?.new_users_today || 0} registered today
                    </span>
                  </div>
                </div>

                <div className="p-4 rounded-2xl bg-card border border-border flex flex-col justify-between">
                  <div className="flex items-center justify-between text-muted text-xs font-bold uppercase tracking-wider">
                    <span>24h Active Users</span>
                    <Activity className="w-4 h-4 text-emerald-400" />
                  </div>
                  <div className="mt-3">
                    <span className="text-2xl font-black">
                      {overview ? formatNumber(overview.active_users_24h) : '—'}
                    </span>
                    <span className="text-xs text-emerald-400 block mt-0.5">Live period active</span>
                  </div>
                </div>

                <div className="p-4 rounded-2xl bg-card border border-border flex flex-col justify-between">
                  <div className="flex items-center justify-between text-muted text-xs font-bold uppercase tracking-wider">
                    <span>Total Posts</span>
                    <FileText className="w-4 h-4 text-purple-400" />
                  </div>
                  <div className="mt-3">
                    <span className="text-2xl font-black">
                      {overview ? formatNumber(overview.total_posts) : '—'}
                    </span>
                    <span className="text-xs text-muted block mt-0.5">
                      +{overview?.posts_today || 0} posted today
                    </span>
                  </div>
                </div>

                <div className="p-4 rounded-2xl bg-card border border-border flex flex-col justify-between">
                  <div className="flex items-center justify-between text-muted text-xs font-bold uppercase tracking-wider">
                    <span>Flagged Content</span>
                    <Flag className="w-4 h-4 text-rose-400" />
                  </div>
                  <div className="mt-3">
                    <span className="text-2xl font-black text-rose-400">
                      {overview?.flagged_posts_count ?? 0}
                    </span>
                    <span className="text-xs text-muted block mt-0.5">Pending moderation</span>
                  </div>
                </div>
              </div>

              {/* System Dependency Health Grid */}
              <div>
                <h3 className="text-sm font-bold uppercase tracking-wider text-muted mb-3">
                  Live Infrastructure Status
                </h3>
                <div className="grid grid-cols-1 sm:grid-cols-4 gap-4">
                  {/* Database Card */}
                  <div className="p-4 rounded-2xl bg-card border border-border">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <Database className="w-4 h-4 text-brand" />
                        <span className="font-bold text-sm">PostgreSQL</span>
                      </div>
                      <span
                        className={cn(
                          'px-2 py-0.5 rounded-full text-[11px] font-bold uppercase',
                          overview?.system_health.database.status === 'healthy'
                            ? 'bg-emerald-500/10 text-emerald-400'
                            : 'bg-rose-500/10 text-rose-400'
                        )}
                      >
                        {overview?.system_health.database.status || 'Active'}
                      </span>
                    </div>
                    <div className="mt-3 text-xs text-muted space-y-1">
                      <div className="flex justify-between">
                        <span>Query Latency:</span>
                        <span className="font-mono text-foreground font-semibold">
                          {overview?.system_health.database.ping_ms.toFixed(2)} ms
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span>Connection Pool:</span>
                        <span className="font-mono text-foreground">
                          {overview?.system_health.database.in_use_connections}/
                          {overview?.system_health.database.open_connections} in-use
                        </span>
                      </div>
                    </div>
                  </div>

                  {/* Redis Card */}
                  <div className="p-4 rounded-2xl bg-card border border-border">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <HardDrive className="w-4 h-4 text-rose-400" />
                        <span className="font-bold text-sm">Redis 7</span>
                      </div>
                      <span className="px-2 py-0.5 rounded-full text-[11px] font-bold uppercase bg-emerald-500/10 text-emerald-400">
                        {overview?.system_health.redis.status || 'Active'}
                      </span>
                    </div>
                    <div className="mt-3 text-xs text-muted space-y-1">
                      <div className="flex justify-between">
                        <span>Ping Latency:</span>
                        <span className="font-mono text-foreground font-semibold">
                          {overview?.system_health.redis.ping_ms.toFixed(2)} ms
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span>Memory Used:</span>
                        <span className="font-mono text-foreground">
                          {overview?.system_health.redis.used_memory_human || 'Active'}
                        </span>
                      </div>
                    </div>
                  </div>

                  {/* RabbitMQ Card */}
                  <div className="p-4 rounded-2xl bg-card border border-border">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <Radio className="w-4 h-4 text-amber-400" />
                        <span className="font-bold text-sm">RabbitMQ</span>
                      </div>
                      <span className="px-2 py-0.5 rounded-full text-[11px] font-bold uppercase bg-emerald-500/10 text-emerald-400">
                        Healthy
                      </span>
                    </div>
                    <div className="mt-3 text-xs text-muted space-y-1">
                      <div className="flex justify-between">
                        <span>Event Bus:</span>
                        <span className="font-mono text-foreground font-semibold">Active</span>
                      </div>
                      <div className="flex justify-between">
                        <span>Queue Status:</span>
                        <span className="font-mono text-foreground">Operational</span>
                      </div>
                    </div>
                  </div>

                  {/* Go Runtime Card */}
                  <div className="p-4 rounded-2xl bg-card border border-border">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <Cpu className="w-4 h-4 text-cyan-400" />
                        <span className="font-bold text-sm">Go Runtime</span>
                      </div>
                      <span className="px-2 py-0.5 rounded-full text-[11px] font-bold uppercase bg-emerald-500/10 text-emerald-400">
                        v1.24
                      </span>
                    </div>
                    <div className="mt-3 text-xs text-muted space-y-1">
                      <div className="flex justify-between">
                        <span>Goroutines:</span>
                        <span className="font-mono text-foreground font-semibold">
                          {overview?.system_health.runtime.goroutines || 0}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span>Heap Alloc:</span>
                        <span className="font-mono text-foreground">
                          {overview
                            ? (overview.system_health.runtime.alloc_bytes / 1024 / 1024).toFixed(1) + ' MB'
                            : '—'}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              {/* Recent Security Alerts Banner */}
              <div className="p-5 rounded-2xl bg-card border border-border">
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center gap-2">
                    <ShieldAlert className="w-5 h-5 text-amber-400" />
                    <h3 className="font-bold text-base">Recent Critical Operations</h3>
                  </div>
                  <button
                    onClick={() => setActiveTab('audit')}
                    className="text-xs text-brand hover:underline font-semibold flex items-center gap-1"
                  >
                    View All Audit Logs <ChevronRight className="w-3.5 h-3.5" />
                  </button>
                </div>

                <div className="divide-y divide-border">
                  {overview?.recent_audit_logs && overview.recent_audit_logs.length > 0 ? (
                    overview.recent_audit_logs.slice(0, 5).map((log) => (
                      <div key={log.id} className="py-2.5 flex items-center justify-between text-xs">
                        <div className="flex items-center gap-3">
                          <span className="px-2 py-0.5 rounded bg-subtle font-mono text-[11px] font-bold text-foreground">
                            {log.action}
                          </span>
                          <span className="text-muted line-clamp-1">{log.reason}</span>
                        </div>
                        <span className="text-muted shrink-0 font-mono">
                          {new Date(log.created_at).toLocaleTimeString()}
                        </span>
                      </div>
                    ))
                  ) : (
                    <div className="py-6 text-center text-muted text-xs">No recent audit activity</div>
                  )}
                </div>
              </div>
            </div>
          )}

          {/* ============================================================ */}
          {/* TAB 2: USER DIRECTORY */}
          {/* ============================================================ */}
          {activeTab === 'users' && (
            <div className="mt-6 space-y-4 animate-in fade-in">
              <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
                <div className="relative flex-1 max-w-md">
                  <Search className="w-4 h-4 absolute left-3 top-3 text-muted" />
                  <input
                    type="text"
                    placeholder="Search by username, email, or user ID..."
                    value={userSearch}
                    onChange={(e) => setUserSearch(e.target.value)}
                    onKeyDown={(e) => e.key === 'Enter' && fetchUsers()}
                    className="w-full pl-9 pr-4 py-2 bg-subtle border border-border rounded-xl text-sm text-foreground focus:outline-none focus:border-brand"
                  />
                </div>

                <div className="flex items-center gap-2">
                  <select
                    value={userStatusFilter}
                    onChange={(e) => setUserStatusFilter(e.target.value as any)}
                    className="px-3 py-2 bg-subtle border border-border rounded-xl text-sm text-foreground"
                  >
                    <option value="all">All Statuses</option>
                    <option value="active">Active Only</option>
                    <option value="suspended">Suspended Only</option>
                  </select>

                  <button
                    onClick={fetchUsers}
                    className="px-4 py-2 bg-brand text-white font-bold rounded-xl text-sm hover:bg-brand/90 transition-colors"
                  >
                    Search
                  </button>
                </div>
              </div>

              <div className="border border-border rounded-2xl overflow-hidden bg-card">
                <div className="overflow-x-auto">
                  <table className="w-full text-left text-sm">
                    <thead className="bg-subtle/60 border-b border-border text-xs uppercase tracking-wider text-muted font-bold">
                      <tr>
                        <th className="py-3 px-4">User</th>
                        <th className="py-3 px-4">Email</th>
                        <th className="py-3 px-4">Role</th>
                        <th className="py-3 px-4">Status</th>
                        <th className="py-3 px-4 text-right">Actions</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-border">
                      {adminUsers
                        .filter((u) => {
                          if (userStatusFilter === 'active') return !u.deleted_at;
                          if (userStatusFilter === 'suspended') return !!u.deleted_at;
                          return true;
                        })
                        .map((u) => {
                          const isSuspended = !!u.deleted_at;
                          const isTargetOwner =
                            u.is_owner || u.role === 'owner' || u.email === 'afrizaahmad18@gmail.com';
                          return (
                            <tr key={u.id} className="hover:bg-subtle/40 transition-colors">
                              <td className="py-3 px-4 flex items-center gap-3">
                                <img
                                  src={
                                    u.avatar_url ||
                                    'https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png'
                                  }
                                  alt={u.username}
                                  className="w-8 h-8 rounded-full object-cover border border-border"
                                />
                                <div>
                                  <div className="font-bold text-foreground flex items-center gap-1.5">
                                    <span>{u.display_name || u.username}</span>
                                    {isTargetOwner && <Crown className="w-3.5 h-3.5 text-amber-400" />}
                                  </div>
                                  <span className="text-xs text-muted">@{u.username}</span>
                                </div>
                              </td>

                              <td className="py-3 px-4 font-mono text-xs text-muted">{u.email || '—'}</td>

                              <td className="py-3 px-4">
                                <span className="px-2.5 py-0.5 rounded-full text-xs font-semibold bg-brand/10 text-brand">
                                  {u.role || 'member'}
                                </span>
                              </td>

                              <td className="py-3 px-4">
                                <span
                                  className={cn(
                                    'px-2.5 py-0.5 rounded-full text-xs font-semibold',
                                    isSuspended
                                      ? 'bg-rose-500/10 text-rose-400'
                                      : 'bg-emerald-500/10 text-emerald-400'
                                  )}
                                >
                                  {isSuspended ? 'Suspended' : 'Active'}
                                </span>
                              </td>

                              <td className="py-3 px-4 text-right">
                                <div className="flex items-center justify-end gap-2">
                                  <button
                                    onClick={() => handleOpenUserRoles(u)}
                                    className="p-1.5 hover:bg-subtle rounded-lg text-muted hover:text-foreground text-xs font-semibold flex items-center gap-1 border border-border"
                                  >
                                    <Key className="w-3.5 h-3.5" />
                                    <span>Roles</span>
                                  </button>

                                  {!isTargetOwner && (
                                    <button
                                      onClick={() => handleToggleUserStatus(u)}
                                      className={cn(
                                        'p-1.5 hover:bg-subtle rounded-lg text-xs font-semibold flex items-center gap-1 border border-border',
                                        isSuspended ? 'text-emerald-400' : 'text-rose-400'
                                      )}
                                    >
                                      {isSuspended ? (
                                        <Unlock className="w-3.5 h-3.5" />
                                      ) : (
                                        <Lock className="w-3.5 h-3.5" />
                                      )}
                                      <span>{isSuspended ? 'Restore' : 'Suspend'}</span>
                                    </button>
                                  )}
                                </div>
                              </td>
                            </tr>
                          );
                        })}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          )}

          {/* ============================================================ */}
          {/* TAB 3: POST MODERATION */}
          {/* ============================================================ */}
          {activeTab === 'posts' && (
            <div className="mt-6 space-y-4 animate-in fade-in">
              <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
                <div className="relative flex-1 max-w-md">
                  <Search className="w-4 h-4 absolute left-3 top-3 text-muted" />
                  <input
                    type="text"
                    placeholder="Search posts content or author..."
                    value={postSearch}
                    onChange={(e) => setPostSearch(e.target.value)}
                    onKeyDown={(e) => e.key === 'Enter' && fetchPosts()}
                    className="w-full pl-9 pr-4 py-2 bg-subtle border border-border rounded-xl text-sm text-foreground focus:outline-none focus:border-brand"
                  />
                </div>

                <div className="flex items-center gap-2">
                  <select
                    value={postModFilter}
                    onChange={(e) => setPostModFilter(e.target.value as any)}
                    className="px-3 py-2 bg-subtle border border-border rounded-xl text-sm text-foreground"
                  >
                    <option value="all">All Moderation Statuses</option>
                    <option value="normal">Normal</option>
                    <option value="flagged">Flagged</option>
                    <option value="hidden">Hidden</option>
                    <option value="removed">Removed</option>
                  </select>

                  <button
                    onClick={fetchPosts}
                    className="px-4 py-2 bg-brand text-white font-bold rounded-xl text-sm hover:bg-brand/90 transition-colors"
                  >
                    Apply Filter
                  </button>
                </div>
              </div>

              <div className="border border-border rounded-2xl overflow-hidden bg-card">
                <div className="overflow-x-auto">
                  <table className="w-full text-left text-sm">
                    <thead className="bg-subtle/60 border-b border-border text-xs uppercase tracking-wider text-muted font-bold">
                      <tr>
                        <th className="py-3 px-4">Author</th>
                        <th className="py-3 px-4">Post Content</th>
                        <th className="py-3 px-4">Moderation</th>
                        <th className="py-3 px-4">Created</th>
                        <th className="py-3 px-4 text-right">Actions</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-border">
                      {adminPosts.map((p) => {
                        return (
                          <tr key={p.id} className="hover:bg-subtle/40 transition-colors">
                            <td className="py-3 px-4">
                              <div className="flex items-center gap-2">
                                <img
                                  src={
                                    p.author?.avatar_url ||
                                    'https://abs.twimg.com/sticky/default_profile_images/default_profile_400x400.png'
                                  }
                                  alt="avatar"
                                  className="w-7 h-7 rounded-full object-cover border border-border"
                                />
                                <div>
                                  <span className="font-bold text-foreground text-xs block">
                                    @{p.author?.username || 'user'}
                                  </span>
                                  <span className="text-[11px] text-muted">{p.author?.display_name}</span>
                                </div>
                              </div>
                            </td>

                            <td className="py-3 px-4 max-w-xs sm:max-w-md">
                              <p className="line-clamp-2 text-xs text-foreground font-normal whitespace-pre-wrap">
                                {p.content}
                              </p>
                            </td>

                            <td className="py-3 px-4">
                              <span
                                className={cn(
                                  'px-2 py-0.5 rounded-full text-xs font-semibold uppercase',
                                  p.moderation_status === 'hidden'
                                    ? 'bg-amber-500/10 text-amber-400'
                                    : p.moderation_status === 'removed'
                                    ? 'bg-rose-500/10 text-rose-400'
                                    : p.moderation_status === 'flagged'
                                    ? 'bg-purple-500/10 text-purple-400'
                                    : 'bg-emerald-500/10 text-emerald-400'
                                )}
                              >
                                {p.moderation_status || 'normal'}
                              </span>
                            </td>

                            <td className="py-3 px-4 text-xs text-muted font-mono whitespace-nowrap">
                              {new Date(p.created_at).toLocaleDateString()}
                            </td>

                            <td className="py-3 px-4 text-right">
                              <div className="flex items-center justify-end gap-1.5">
                                <button
                                  onClick={() => handleOpenModeration(p)}
                                  className="p-1.5 hover:bg-subtle rounded-lg text-xs font-semibold flex items-center gap-1 border border-border"
                                  title="Moderate Post"
                                >
                                  <Flag className="w-3.5 h-3.5 text-amber-400" />
                                  <span>Moderate</span>
                                </button>
                                <button
                                  onClick={() => handleDeletePostAdmin(p)}
                                  className="p-1.5 hover:bg-subtle rounded-lg text-xs font-semibold flex items-center gap-1 border border-border text-rose-400"
                                  title="Delete Post"
                                >
                                  <Trash2 className="w-3.5 h-3.5" />
                                </button>
                              </div>
                            </td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          )}

          {/* ============================================================ */}
          {/* TAB 4: ROLES & PERMISSIONS */}
          {/* ============================================================ */}
          {activeTab === 'roles' && (
            <div className="mt-6 space-y-6 animate-in fade-in">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-lg font-black">Role Registry</h2>
                  <p className="text-xs text-muted">Platform Authorization & Administrative Hierarchy</p>
                </div>
                <button
                  onClick={handleOpenCreateRole}
                  className="px-4 py-2 bg-brand text-white text-xs font-bold rounded-xl hover:bg-brand/90 transition-colors flex items-center gap-1.5 shadow-sm"
                >
                  <Plus className="w-4 h-4" />
                  <span>Create Custom Role</span>
                </button>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {roles.map((r) => {
                  const isProtected = r.is_system || r.key === 'owner';
                  return (
                    <div
                      key={r.id}
                      className="p-5 rounded-2xl bg-card border border-border flex flex-col justify-between"
                    >
                      <div>
                        <div className="flex items-center justify-between">
                          <span className="font-mono text-xs font-bold px-2 py-0.5 rounded bg-subtle text-brand">
                            {r.key}
                          </span>
                          {isProtected ? (
                            <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-amber-500/10 text-amber-400 border border-amber-500/30 uppercase">
                              System Role
                            </span>
                          ) : (
                            <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 uppercase">
                              Custom Role
                            </span>
                          )}
                        </div>

                        <h3 className="font-bold text-base mt-2">{r.name}</h3>
                        <p className="text-xs text-muted mt-1 line-clamp-2">{r.description || 'No description'}</p>

                        <div className="mt-4 flex items-center justify-between text-xs text-muted font-mono">
                          <span>Assigned Users:</span>
                          <span className="font-bold text-foreground">{r.user_count || 0}</span>
                        </div>
                      </div>

                      <div className="mt-5 pt-3 border-t border-border flex items-center justify-between gap-2">
                        <button
                          onClick={() => handleOpenPermissions(r)}
                          className="flex-1 py-1.5 px-3 rounded-lg border border-border hover:bg-subtle text-xs font-semibold flex items-center justify-center gap-1.5 transition-colors"
                        >
                          <Key className="w-3.5 h-3.5 text-brand" />
                          <span>Permissions</span>
                        </button>

                        {!isProtected && (
                          <button
                            onClick={() => {
                              setConfirmDialog({
                                isOpen: true,
                                title: `Delete role ${r.name}?`,
                                description: 'Are you sure you want to delete this custom role? Members with this role will lose its permissions.',
                                confirmText: 'Delete',
                                variant: 'destructive',
                                onConfirm: async () => {
                                  try {
                                    await apiRequest(`/admin/roles/${r.id}`, { method: 'DELETE' });
                                    showStatus('success', 'Role deleted');
                                    fetchRoles();
                                  } catch (e: any) {
                                    showStatus('error', e.message);
                                  } finally {
                                    setConfirmDialog((prev) => ({ ...prev, isOpen: false }));
                                  }
                                },
                              });
                            }}
                            className="p-1.5 rounded-lg border border-border hover:bg-subtle text-rose-400"
                            title="Delete custom role"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {/* ============================================================ */}
          {/* TAB 5: MEMBERSHIPS & QUOTAS */}
          {/* ============================================================ */}
          {activeTab === 'memberships' && (
            <div className="mt-6 space-y-6 animate-in fade-in">
              <div>
                <h2 className="text-lg font-black">Membership Tiers & Quota Limits</h2>
                <p className="text-xs text-muted">
                  Configurable daily post allowances, media attachment limits, and billing amounts
                </p>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
                {membershipPlans.map((plan) => {
                  return (
                    <div
                      key={plan.id}
                      className="p-6 rounded-2xl bg-card border border-border flex flex-col justify-between"
                    >
                      <div>
                        <div className="flex items-center justify-between">
                          <span className="font-mono text-xs font-bold uppercase tracking-wider text-muted">
                            {plan.key} Tier
                          </span>
                          <TierBadge tier={plan.key} />
                        </div>

                        <h3 className="text-xl font-black mt-2">{plan.name}</h3>
                        <p className="text-xs text-muted mt-1">{plan.description}</p>

                        <div className="mt-4 p-3 bg-subtle rounded-xl">
                          <span className="text-2xl font-black">
                            ${(plan.price_amount_cents / 100).toFixed(2)}
                          </span>
                          <span className="text-xs text-muted font-semibold ml-1">
                            / {plan.billing_interval}
                          </span>
                        </div>

                        <div className="mt-4 space-y-2 text-xs">
                          {plan.entitlements?.map((ent) => (
                            <div key={ent.id} className="flex items-center justify-between text-muted">
                              <span>{ent.entitlement_key}:</span>
                              <span className="font-mono font-bold text-foreground">
                                {ent.limit_value} {ent.unit}
                              </span>
                            </div>
                          ))}
                        </div>
                      </div>

                      <div className="mt-6 pt-4 border-t border-border">
                        <button
                          onClick={() => handleOpenPlanEntitlements(plan)}
                          className="w-full py-2 px-4 rounded-xl border border-border hover:bg-subtle text-xs font-bold flex items-center justify-center gap-1.5 transition-colors"
                        >
                          <Sliders className="w-3.5 h-3.5 text-brand" />
                          <span>Configure Quotas & Limits</span>
                        </button>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {/* ============================================================ */}
          {/* TAB 6: SECURITY MONITORING */}
          {/* ============================================================ */}
          {activeTab === 'security' && (
            <div className="mt-6 space-y-6 animate-in fade-in">
              <div>
                <h2 className="text-lg font-black">Security Threat & Anomaly Monitoring</h2>
                <p className="text-xs text-muted">
                  Live detection of unauthorized elevation attempts, authentication anomalies, and sensitive operations
                </p>
              </div>

              <div className="border border-border rounded-2xl overflow-hidden bg-card">
                <div className="overflow-x-auto">
                  <table className="w-full text-left text-sm">
                    <thead className="bg-subtle/60 border-b border-border text-xs uppercase tracking-wider text-muted font-bold">
                      <tr>
                        <th className="py-3 px-4">Event / Action</th>
                        <th className="py-3 px-4">Actor</th>
                        <th className="py-3 px-4">Target Entity</th>
                        <th className="py-3 px-4">Justification / Reason</th>
                        <th className="py-3 px-4">Timestamp</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-border font-mono text-xs">
                      {securityEvents.map((evt) => (
                        <tr key={evt.id} className="hover:bg-subtle/40 transition-colors">
                          <td className="py-3 px-4">
                            <span className="px-2 py-0.5 rounded bg-rose-500/10 text-rose-400 font-bold">
                              {evt.action}
                            </span>
                          </td>
                          <td className="py-3 px-4 text-foreground font-semibold">
                            {evt.actor_username || 'system'}
                          </td>
                          <td className="py-3 px-4 text-muted">
                            {evt.target_type}:{evt.target_id.slice(0, 8)}
                          </td>
                          <td className="py-3 px-4 text-muted line-clamp-1 font-sans">{evt.reason}</td>
                          <td className="py-3 px-4 text-muted whitespace-nowrap">
                            {new Date(evt.created_at).toLocaleString()}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          )}

          {/* ============================================================ */}
          {/* TAB 7: INFRASTRUCTURE MONITORING */}
          {/* ============================================================ */}
          {activeTab === 'monitoring' && (
            <div className="mt-6 space-y-6 animate-in fade-in">
              <div>
                <h2 className="text-lg font-black">Technical System Telemetry</h2>
                <p className="text-xs text-muted">PostgreSQL connection metrics, Redis memory, and Go runtime stats</p>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="p-5 rounded-2xl bg-card border border-border">
                  <h3 className="font-bold text-sm uppercase text-muted tracking-wider mb-4 flex items-center gap-2">
                    <Database className="w-4 h-4 text-brand" /> Database Pool Metrics
                  </h3>
                  <div className="space-y-3 text-sm">
                    <div className="flex justify-between py-1 border-b border-border/50">
                      <span className="text-muted">Connection Health:</span>
                      <span className="font-bold text-emerald-400">
                        {overview?.system_health.database.status || 'healthy'}
                      </span>
                    </div>
                    <div className="flex justify-between py-1 border-b border-border/50">
                      <span className="text-muted">Roundtrip Latency:</span>
                      <span className="font-mono font-bold">
                        {overview?.system_health.database.ping_ms.toFixed(2)} ms
                      </span>
                    </div>
                    <div className="flex justify-between py-1 border-b border-border/50">
                      <span className="text-muted">Active Open Connections:</span>
                      <span className="font-mono font-bold">
                        {overview?.system_health.database.open_connections}
                      </span>
                    </div>
                    <div className="flex justify-between py-1">
                      <span className="text-muted">Connections In-Use:</span>
                      <span className="font-mono font-bold text-brand">
                        {overview?.system_health.database.in_use_connections}
                      </span>
                    </div>
                  </div>
                </div>

                <div className="p-5 rounded-2xl bg-card border border-border">
                  <h3 className="font-bold text-sm uppercase text-muted tracking-wider mb-4 flex items-center gap-2">
                    <Cpu className="w-4 h-4 text-cyan-400" /> Go Runtime Heap & Engine
                  </h3>
                  <div className="space-y-3 text-sm">
                    <div className="flex justify-between py-1 border-b border-border/50">
                      <span className="text-muted">Active Goroutines:</span>
                      <span className="font-mono font-bold">
                        {overview?.system_health.runtime.goroutines || 0}
                      </span>
                    </div>
                    <div className="flex justify-between py-1 border-b border-border/50">
                      <span className="text-muted">Memory Allocated:</span>
                      <span className="font-mono font-bold">
                        {overview
                          ? (overview.system_health.runtime.alloc_bytes / 1024 / 1024).toFixed(2) + ' MB'
                          : '—'}
                      </span>
                    </div>
                    <div className="flex justify-between py-1 border-b border-border/50">
                      <span className="text-muted">Total Cumulative Alloc:</span>
                      <span className="font-mono font-bold">
                        {overview
                          ? (overview.system_health.runtime.total_alloc_bytes / 1024 / 1024).toFixed(2) + ' MB'
                          : '—'}
                      </span>
                    </div>
                    <div className="flex justify-between py-1">
                      <span className="text-muted">GC Cycles Executed:</span>
                      <span className="font-mono font-bold text-emerald-400">
                        {overview?.system_health.runtime.num_gc || 0}
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* ============================================================ */}
          {/* TAB 8: AUDIT TRAIL */}
          {/* ============================================================ */}
          {activeTab === 'audit' && (
            <div className="mt-6 space-y-4 animate-in fade-in">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-lg font-black">Audit Trail</h2>
                  <p className="text-xs text-muted">Immutable log of all administrative and moderation actions</p>
                </div>
                <button
                  onClick={fetchAuditLogs}
                  className="px-3 py-1.5 rounded-xl border border-border text-xs font-semibold hover:bg-subtle flex items-center gap-1.5"
                >
                  <RefreshCw className="w-3.5 h-3.5" />
                  <span>Refresh</span>
                </button>
              </div>

              <div className="border border-border rounded-2xl overflow-hidden bg-card">
                <div className="overflow-x-auto">
                  <table className="w-full text-left text-sm font-mono text-xs">
                    <thead className="bg-subtle/60 border-b border-border text-[11px] uppercase tracking-wider text-muted font-bold">
                      <tr>
                        <th className="py-3 px-4">Action</th>
                        <th className="py-3 px-4">Actor</th>
                        <th className="py-3 px-4">Target</th>
                        <th className="py-3 px-4">Reason</th>
                        <th className="py-3 px-4">Timestamp</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-border">
                      {auditLogs.map((log) => (
                        <tr key={log.id} className="hover:bg-subtle/40 transition-colors">
                          <td className="py-3 px-4">
                            <span className="px-2 py-0.5 rounded bg-subtle font-bold text-foreground">
                              {log.action}
                            </span>
                          </td>
                          <td className="py-3 px-4 text-foreground font-semibold">
                            {log.actor_username || 'system'}
                          </td>
                          <td className="py-3 px-4 text-muted">
                            {log.target_type}:{log.target_id.slice(0, 8)}
                          </td>
                          <td className="py-3 px-4 text-muted font-sans line-clamp-1">{log.reason}</td>
                          <td className="py-3 px-4 text-muted whitespace-nowrap">
                            {new Date(log.created_at).toLocaleString()}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          )}
        </main>

      {/* ============================================================ */}
      {/* MODAL: POST MODERATION */}
      {/* ============================================================ */}
      {postModerateModal && selectedPost && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in">
          <div className="w-full max-w-md bg-background border border-border rounded-2xl shadow-2xl p-6">
            <div className="flex items-center justify-between pb-3 border-b border-border">
              <h3 className="font-bold text-base flex items-center gap-2">
                <Flag className="w-4 h-4 text-amber-400" />
                <span>Moderate Post</span>
              </h3>
              <button onClick={() => setPostModerateModal(false)} className="p-1 hover:opacity-70">
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="mt-4 space-y-4 text-sm">
              <div className="p-3 bg-subtle rounded-xl text-xs text-muted">
                <span className="font-bold text-foreground block mb-1">
                  Post by @{selectedPost.author?.username}:
                </span>
                <p className="line-clamp-3">{selectedPost.content}</p>
              </div>

              <div>
                <label className="font-bold text-xs block mb-1">Moderation Status</label>
                <select
                  value={moderateStatus}
                  onChange={(e) => setModerateStatus(e.target.value as any)}
                  className="w-full px-3 py-2 bg-subtle border border-border rounded-xl text-sm text-foreground focus:outline-none focus:border-brand"
                >
                  <option value="normal">Normal (Publicly Visible)</option>
                  <option value="flagged">Flagged (Marked for review)</option>
                  <option value="hidden">Hidden (Invisible in feed)</option>
                </select>
              </div>

              <div>
                <label className="font-bold text-xs block mb-1">Justification Reason</label>
                <textarea
                  value={moderateReason}
                  onChange={(e) => setModerateReason(e.target.value)}
                  placeholder="Explain reason for audit log..."
                  rows={3}
                  className="w-full px-3 py-2 bg-subtle border border-border rounded-xl text-sm text-foreground focus:outline-none focus:border-brand resize-none"
                />
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  onClick={() => setPostModerateModal(false)}
                  className="px-4 py-2 rounded-xl border border-border text-xs font-semibold hover:bg-subtle"
                >
                  Cancel
                </button>
                <button
                  onClick={handleSaveModeration}
                  className="px-4 py-2 rounded-xl bg-brand text-white text-xs font-bold hover:bg-brand/90"
                >
                  Save Status
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ============================================================ */}
      {/* MODAL: ASSIGN USER ROLES */}
      {/* ============================================================ */}
      {userRoleModal && selectedUser && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in">
          <div className="w-full max-w-md bg-background border border-border rounded-2xl shadow-2xl p-6">
            <div className="flex items-center justify-between pb-3 border-b border-border">
              <h3 className="font-bold text-base flex items-center gap-2">
                <Key className="w-4 h-4 text-brand" />
                <span>Assign Roles to @{selectedUser.username}</span>
              </h3>
              <button onClick={() => setUserRoleModal(false)} className="p-1 hover:opacity-70">
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="mt-4 space-y-4 text-sm">
              <div className="space-y-2 max-h-60 overflow-y-auto pr-1">
                {roles.map((r) => {
                  const isChecked = userSelectedRoleIds.includes(r.id);
                  return (
                    <label
                      key={r.id}
                      className="flex items-center gap-3 p-3 rounded-xl border border-border hover:bg-subtle/50 cursor-pointer transition-colors"
                    >
                      <input
                        type="checkbox"
                        checked={isChecked}
                        onChange={(e) => {
                          if (e.target.checked) {
                            setUserSelectedRoleIds([...userSelectedRoleIds, r.id]);
                          } else {
                            setUserSelectedRoleIds(userSelectedRoleIds.filter((id) => id !== r.id));
                          }
                        }}
                        className="rounded border-border text-brand focus:ring-brand"
                      />
                      <div className="flex-1">
                        <span className="font-bold text-xs block text-foreground">{r.name}</span>
                        <span className="text-[11px] text-muted font-mono">{r.key}</span>
                      </div>
                    </label>
                  );
                })}
              </div>

              <div>
                <label className="font-bold text-xs block mb-1">Audit Log Reason</label>
                <input
                  type="text"
                  value={roleChangeReason}
                  onChange={(e) => setRoleChangeReason(e.target.value)}
                  placeholder="Document the administrative reason..."
                  className="w-full px-3 py-2 bg-subtle border border-border rounded-xl text-xs text-foreground focus:outline-none focus:border-brand"
                />
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  onClick={() => setUserRoleModal(false)}
                  className="px-4 py-2 rounded-xl border border-border text-xs font-semibold hover:bg-subtle"
                >
                  Cancel
                </button>
                <button
                  onClick={handleSaveUserRoles}
                  className="px-4 py-2 rounded-xl bg-brand text-white text-xs font-bold hover:bg-brand/90"
                >
                  Save Roles
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ============================================================ */}
      {/* MODAL: EDIT ROLE PERMISSIONS */}
      {/* ============================================================ */}
      {editPermissionsModal && selectedRole && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in">
          <div className="w-full max-w-2xl bg-background border border-border rounded-2xl shadow-2xl p-6 flex flex-col max-h-[85vh]">
            <div className="flex items-center justify-between pb-3 border-b border-border">
              <h3 className="font-bold text-base flex items-center gap-2">
                <Shield className="w-4 h-4 text-brand" />
                <span>Configure Permissions for {selectedRole.name}</span>
              </h3>
              <button onClick={() => setEditPermissionsModal(false)} className="p-1 hover:opacity-70">
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="mt-4 flex-1 overflow-y-auto pr-1 grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs">
              {permissions.map((p) => {
                const isSelected = selectedPermIds.includes(p.id);
                return (
                  <label
                    key={p.id}
                    className="flex items-start gap-2.5 p-2.5 rounded-xl border border-border hover:bg-subtle/50 cursor-pointer"
                  >
                    <input
                      type="checkbox"
                      checked={isSelected}
                      onChange={(e) => {
                        if (e.target.checked) {
                          setSelectedPermIds([...selectedPermIds, p.id]);
                        } else {
                          setSelectedPermIds(selectedPermIds.filter((id) => id !== p.id));
                        }
                      }}
                      className="mt-0.5 rounded border-border text-brand focus:ring-brand"
                    />
                    <div>
                      <span className="font-mono font-bold text-foreground block">{p.key}</span>
                      <span className="text-muted text-[11px]">{p.description}</span>
                    </div>
                  </label>
                );
              })}
            </div>

            <div className="flex justify-end gap-2 pt-4 border-t border-border mt-4">
              <button
                onClick={() => setEditPermissionsModal(false)}
                className="px-4 py-2 rounded-xl border border-border text-xs font-semibold hover:bg-subtle"
              >
                Cancel
              </button>
              <button
                onClick={handleSaveRolePermissions}
                className="px-4 py-2 rounded-xl bg-brand text-white text-xs font-bold hover:bg-brand/90"
              >
                Save Permissions
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ============================================================ */}
      {/* MODAL: CONFIGURE PLAN ENTITLEMENTS */}
      {/* ============================================================ */}
      {planEntitlementsModal && selectedPlan && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in">
          <div className="w-full max-w-md bg-background border border-border rounded-2xl shadow-2xl p-6">
            <div className="flex items-center justify-between pb-3 border-b border-border">
              <h3 className="font-bold text-base flex items-center gap-2">
                <Sliders className="w-4 h-4 text-brand" />
                <span>Quotas for {selectedPlan.name}</span>
              </h3>
              <button onClick={() => setPlanEntitlementsModal(false)} className="p-1 hover:opacity-70">
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="mt-4 space-y-4 text-xs">
              <div>
                <label className="font-bold block mb-1">Daily Posts Limit</label>
                <input
                  type="number"
                  value={entitlementsFormData['daily_posts_limit'] ?? 2}
                  onChange={(e) =>
                    setEntitlementsFormData({
                      ...entitlementsFormData,
                      daily_posts_limit: Number(e.target.value),
                    })
                  }
                  className="w-full px-3 py-2 bg-subtle border border-border rounded-xl font-mono text-sm"
                />
              </div>

              <div>
                <label className="font-bold block mb-1">Max Images per Post</label>
                <input
                  type="number"
                  value={entitlementsFormData['max_images_per_post'] ?? 2}
                  onChange={(e) =>
                    setEntitlementsFormData({
                      ...entitlementsFormData,
                      max_images_per_post: Number(e.target.value),
                    })
                  }
                  className="w-full px-3 py-2 bg-subtle border border-border rounded-xl font-mono text-sm"
                />
              </div>

              <div>
                <label className="font-bold block mb-1">Max Videos per Post</label>
                <input
                  type="number"
                  value={entitlementsFormData['max_videos_per_post'] ?? 0}
                  onChange={(e) =>
                    setEntitlementsFormData({
                      ...entitlementsFormData,
                      max_videos_per_post: Number(e.target.value),
                    })
                  }
                  className="w-full px-3 py-2 bg-subtle border border-border rounded-xl font-mono text-sm"
                />
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  onClick={() => setPlanEntitlementsModal(false)}
                  className="px-4 py-2 rounded-xl border border-border text-xs font-semibold hover:bg-subtle"
                >
                  Cancel
                </button>
                <button
                  onClick={handleSavePlanEntitlements}
                  className="px-4 py-2 rounded-xl bg-brand text-white text-xs font-bold hover:bg-brand/90"
                >
                  Save Entitlements
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ============================================================ */}
      {/* MODAL: CREATE ROLE */}
      {/* ============================================================ */}
      {roleModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in">
          <div className="w-full max-w-md bg-background border border-border rounded-2xl shadow-2xl p-6">
            <div className="flex items-center justify-between pb-3 border-b border-border">
              <h3 className="font-bold text-base flex items-center gap-2">
                <Plus className="w-4 h-4 text-brand" />
                <span>Create Custom Role</span>
              </h3>
              <button onClick={() => setRoleModalOpen(false)} className="p-1 hover:opacity-70">
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="mt-4 space-y-3 text-xs">
              <div>
                <label className="font-bold block mb-1">Role Key (Identifier)</label>
                <input
                  type="text"
                  placeholder="e.g. content_lead"
                  value={roleFormData.key}
                  onChange={(e) => setRoleFormData({ ...roleFormData, key: e.target.value })}
                  className="w-full px-3 py-2 bg-subtle border border-border rounded-xl font-mono text-sm"
                />
              </div>

              <div>
                <label className="font-bold block mb-1">Display Name</label>
                <input
                  type="text"
                  placeholder="e.g. Content Lead"
                  value={roleFormData.name}
                  onChange={(e) => setRoleFormData({ ...roleFormData, name: e.target.value })}
                  className="w-full px-3 py-2 bg-subtle border border-border rounded-xl text-sm"
                />
              </div>

              <div>
                <label className="font-bold block mb-1">Description</label>
                <textarea
                  placeholder="Role responsibilities..."
                  value={roleFormData.description}
                  onChange={(e) => setRoleFormData({ ...roleFormData, description: e.target.value })}
                  rows={2}
                  className="w-full px-3 py-2 bg-subtle border border-border rounded-xl text-sm resize-none"
                />
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  onClick={() => setRoleModalOpen(false)}
                  className="px-4 py-2 rounded-xl border border-border text-xs font-semibold hover:bg-subtle"
                >
                  Cancel
                </button>
                <button
                  onClick={handleSaveRole}
                  className="px-4 py-2 rounded-xl bg-brand text-white text-xs font-bold hover:bg-brand/90"
                >
                  Create Role
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Confirmation Modal */}
      <ConfirmModal
        isOpen={confirmDialog.isOpen}
        onClose={() => setConfirmDialog((prev) => ({ ...prev, isOpen: false }))}
        onConfirm={confirmDialog.onConfirm}
        title={confirmDialog.title}
        description={confirmDialog.description}
        confirmText={confirmDialog.confirmText}
        variant={confirmDialog.variant}
        isLoading={confirmDialog.isLoading}
      />
    </div>
  );
}
