import { apiClient } from '../client'
import type { UserRiskDetail } from './users'

export interface UserRiskSettings {
  ip_intel_enabled: boolean
  ip_intel_provider: 'ipinfo_lite' | string
  ip_intel_token_configured: boolean
  ip_intel_docs_url: string
  review_threshold: number
  throttle_threshold: number
  freeze_threshold: number
  auto_enabled: boolean
  auto_throttle: boolean
  auto_freeze: boolean
  auto_throttle_concurrency_cap: number
}

export interface UpdateUserRiskSettingsRequest {
  ip_intel_enabled?: boolean
  ip_intel_provider?: 'ipinfo_lite' | string
  ip_intel_token?: string
  clear_ip_intel_token?: boolean
  review_threshold?: number
  throttle_threshold?: number
  freeze_threshold?: number
  auto_enabled?: boolean
  auto_throttle?: boolean
  auto_freeze?: boolean
  auto_throttle_concurrency_cap?: number
}

export interface UserRiskSummaryItem {
  user_id: number
  summary: {
    risk_score: number
    decision: 'normal' | 'observe' | 'review' | 'throttle' | 'freeze_review'
    decision_label: string
    computed_at: string
  }
  metrics: {
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
  rule_hit_count: number
  top_user_agent?: string
}

export async function getSettings(): Promise<UserRiskSettings> {
  const { data } = await apiClient.get<UserRiskSettings>('/admin/risk/settings')
  return data
}

export async function updateSettings(payload: UpdateUserRiskSettingsRequest): Promise<UserRiskSettings> {
  const { data } = await apiClient.put<UserRiskSettings>('/admin/risk/settings', payload)
  return data
}

export async function getSummaries(userIds: number[], timezone?: string): Promise<UserRiskSummaryItem[]> {
  const { data } = await apiClient.post<{ items: UserRiskSummaryItem[] }>('/admin/risk/summaries', {
    user_ids: userIds,
    timezone: timezone || Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC'
  })
  return data.items || []
}

export async function getUserDetail(id: number, timezone?: string): Promise<UserRiskDetail> {
  const { data } = await apiClient.get<UserRiskDetail>(`/admin/risk/users/${id}`, {
    params: timezone ? { timezone } : undefined
  })
  return data
}

export const riskAPI = {
  getSettings,
  updateSettings,
  getSummaries,
  getUserDetail
}

export default riskAPI
