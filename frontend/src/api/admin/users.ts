/**
 * Admin Users API endpoints
 * Handles user management for administrators
 */

import { apiClient } from '../client'
import type { AdminUser, UpdateUserRequest, PaginatedResponse, ApiKey } from '@/types'

export interface AdminBindAuthIdentityChannelRequest {
  channel: string
  channel_app_id: string
  channel_subject: string
  metadata?: Record<string, unknown> | null
}

export interface AdminBindAuthIdentityRequest {
  provider_type: string
  provider_key: string
  provider_subject: string
  issuer?: string | null
  metadata?: Record<string, unknown> | null
  channel?: AdminBindAuthIdentityChannelRequest
}

export interface AdminBoundAuthIdentityChannel {
  channel: string
  channel_app_id: string
  channel_subject: string
  metadata: Record<string, unknown> | null
  created_at: string
  updated_at: string
}

export interface AdminBoundAuthIdentity {
  user_id: number
  provider_type: string
  provider_key: string
  provider_subject: string
  verified_at?: string | null
  issuer?: string | null
  metadata: Record<string, unknown> | null
  created_at: string
  updated_at: string
  channel?: AdminBoundAuthIdentityChannel | null
}

/**
 * List all users with pagination
 * @param page - Page number (default: 1)
 * @param pageSize - Items per page (default: 20)
 * @param filters - Optional filters (status, role, search, attributes)
 * @param options - Optional request options (signal)
 * @returns Paginated list of users
 */
export async function list(
  page: number = 1,
  pageSize: number = 20,
  filters?: {
    status?: 'active' | 'disabled'
    role?: 'admin' | 'user'
    search?: string
    group_name?: string         // fuzzy filter by allowed group name
    attributes?: Record<number, string>  // attributeId -> value
    include_subscriptions?: boolean
    sort_by?: string
    sort_order?: 'asc' | 'desc'
  },
  options?: {
    signal?: AbortSignal
  }
): Promise<PaginatedResponse<AdminUser>> {
  // Build params with attribute filters in attr[id]=value format
  const params: Record<string, any> = {
    page,
    page_size: pageSize,
    status: filters?.status,
    role: filters?.role,
    search: filters?.search,
    group_name: filters?.group_name,
    include_subscriptions: filters?.include_subscriptions,
    sort_by: filters?.sort_by,
    sort_order: filters?.sort_order
  }

  // Add attribute filters as attr[id]=value
  if (filters?.attributes) {
    for (const [attrId, value] of Object.entries(filters.attributes)) {
      if (value) {
        params[`attr[${attrId}]`] = value
      }
    }
  }
  const { data } = await apiClient.get<PaginatedResponse<AdminUser>>('/admin/users', {
    params,
    signal: options?.signal
  })
  return data
}

/**
 * Get user by ID
 * @param id - User ID
 * @returns User details
 */
export async function getById(id: number): Promise<AdminUser> {
  const { data } = await apiClient.get<AdminUser>(`/admin/users/${id}`)
  return data
}

/**
 * Create new user
 * @param userData - User data (email, password, etc.)
 * @returns Created user
 */
export async function create(userData: {
  email: string
  password: string
  balance?: number
  concurrency?: number
  allowed_groups?: number[] | null
}): Promise<AdminUser> {
  const { data } = await apiClient.post<AdminUser>('/admin/users', userData)
  return data
}

/**
 * Update user
 * @param id - User ID
 * @param updates - Fields to update
 * @returns Updated user
 */
export async function update(id: number, updates: UpdateUserRequest): Promise<AdminUser> {
  const { data } = await apiClient.put<AdminUser>(`/admin/users/${id}`, updates)
  return data
}

/**
 * Delete user
 * @param id - User ID
 * @returns Success confirmation
 */
export async function deleteUser(id: number): Promise<{ message: string }> {
  const { data } = await apiClient.delete<{ message: string }>(`/admin/users/${id}`)
  return data
}

/**
 * Update user balance
 * @param id - User ID
 * @param balance - New balance
 * @param operation - Operation type ('set', 'add', 'subtract')
 * @param notes - Optional notes for the balance adjustment
 * @returns Updated user
 */
export async function updateBalance(
  id: number,
  balance: number,
  operation: 'set' | 'add' | 'subtract' = 'set',
  notes?: string
): Promise<AdminUser> {
  const { data } = await apiClient.post<AdminUser>(`/admin/users/${id}/balance`, {
    balance,
    operation,
    notes: notes || ''
  })
  return data
}

/**
 * Update user concurrency
 * @param id - User ID
 * @param concurrency - New concurrency limit
 * @returns Updated user
 */
export async function updateConcurrency(id: number, concurrency: number): Promise<AdminUser> {
  return update(id, { concurrency })
}

/**
 * Toggle user status
 * @param id - User ID
 * @param status - New status
 * @returns Updated user
 */
export async function toggleStatus(id: number, status: 'active' | 'disabled'): Promise<AdminUser> {
  return update(id, { status })
}

/**
 * Get user's API keys
 * @param id - User ID
 * @returns List of user's API keys
 */
export async function getUserApiKeys(id: number): Promise<PaginatedResponse<ApiKey>> {
  const { data } = await apiClient.get<PaginatedResponse<ApiKey>>(`/admin/users/${id}/api-keys`)
  return data
}

/**
 * Get user's usage statistics
 * @param id - User ID
 * @param period - Time period
 * @returns User usage statistics
 */
export async function getUserUsageStats(
  id: number,
  period: string = 'month'
): Promise<{
  total_requests: number
  total_cost: number
  total_tokens: number
}> {
  const { data } = await apiClient.get<{
    total_requests: number
    total_cost: number
    total_tokens: number
  }>(`/admin/users/${id}/usage`, {
    params: { period }
  })
  return data
}

/**
 * Balance history item returned from the API
 */
export interface BalanceHistoryItem {
  id: number
  code: string
  type: string
  value: number
  status: string
  used_by: number | null
  used_at: string | null
  created_at: string
  group_id: number | null
  validity_days: number
  notes: string
  user?: { id: number; email: string } | null
  group?: { id: number; name: string } | null
}

// Balance history response extends pagination with total_recharged summary
export interface BalanceHistoryResponse extends PaginatedResponse<BalanceHistoryItem> {
  total_recharged: number
}

export interface UserRiskWindow {
  start_at: string
  end_at: string
  timezone: string
}

export interface UserRiskSummary {
  risk_score: number
  decision: 'normal' | 'observe' | 'review' | 'throttle' | 'freeze_review'
  decision_label: string
  computed_at: string
}

export interface UserRiskMetrics {
  request_count_24h: number
  actual_cost_24h: number
  first_ip?: string
  historical_ip_count: number
  ua_24h_count: number
  active_hours_count: number
  active_hours: number[]
  longest_silence_hours: number
  all_day_active: boolean
  hour_concentration: number
  key_count: number
  concurrent_multi_ip_ua_minutes_24h: number
}

export interface UserRiskRuleHit {
  code: string
  label: string
  description: string
  score: number
}

export interface UserRiskIPDetail {
  ip_address: string
  requests: number
  ip_type: string
  label: string
  asn?: string
  organization?: string
  domain?: string
  country_code?: string
  country?: string
  continent?: string
}

export interface UserRiskUADetail {
  user_agent: string
  requests: number
  category: string
  description: string
  base_score: number
  config_status: string
  hit_rule?: string
  programmatic: boolean
  normal_allowed: boolean
}

export interface UserRiskDetail {
  user: Pick<AdminUser, 'id' | 'email' | 'username' | 'balance' | 'status'>
  window: UserRiskWindow
  summary: UserRiskSummary
  metrics: UserRiskMetrics
  rule_hits: UserRiskRuleHit[]
  ip_details: UserRiskIPDetail[]
  ua_details: UserRiskUADetail[]
}

/**
 * Get user's balance/concurrency change history
 * @param id - User ID
 * @param page - Page number
 * @param pageSize - Items per page
 * @param type - Optional type filter (balance, affiliate_balance, admin_balance, concurrency, admin_concurrency, subscription)
 * @returns Paginated balance history with total_recharged
 */
export async function getUserBalanceHistory(
  id: number,
  page: number = 1,
  pageSize: number = 20,
  type?: string
): Promise<BalanceHistoryResponse> {
  const params: Record<string, any> = { page, page_size: pageSize }
  if (type) params.type = type
  const { data } = await apiClient.get<BalanceHistoryResponse>(
    `/admin/users/${id}/balance-history`,
    { params }
  )
  return data
}

/**
 * Get rolling abuse-risk detail for a user.
 * @param id - User ID
 * @param timezone - Optional IANA timezone for activity-hour bucketing
 */
export async function getUserRiskDetail(
  id: number,
  timezone?: string
): Promise<UserRiskDetail> {
  const { data } = await apiClient.get<UserRiskDetail>(`/admin/users/${id}/risk`, {
    params: timezone ? { timezone } : undefined
  })
  return data
}

/**
 * Replace user's exclusive group
 * @param userId - User ID
 * @param oldGroupId - Current group ID to replace
 * @param newGroupId - New group ID to replace with
 * @returns Number of migrated keys
 */
export async function replaceGroup(
  userId: number,
  oldGroupId: number,
  newGroupId: number
): Promise<{ migrated_keys: number }> {
  const { data } = await apiClient.post<{ migrated_keys: number }>(
    `/admin/users/${userId}/replace-group`,
    { old_group_id: oldGroupId, new_group_id: newGroupId }
  )
  return data
}

export async function bindUserAuthIdentity(
  userId: number,
  input: AdminBindAuthIdentityRequest
): Promise<AdminBoundAuthIdentity> {
  const { data } = await apiClient.post<AdminBoundAuthIdentity>(
    `/admin/users/${userId}/auth-identities`,
    input
  )
  return data
}

export const usersAPI = {
  list,
  getById,
  create,
  update,
  delete: deleteUser,
  updateBalance,
  updateConcurrency,
  toggleStatus,
  getUserApiKeys,
  getUserUsageStats,
  getUserBalanceHistory,
  getUserRiskDetail,
  replaceGroup,
  bindUserAuthIdentity
}

export default usersAPI
