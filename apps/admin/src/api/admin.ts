import http from './http'

export interface PaginationParams {
  page?: number
  page_size?: number
  q?: string
  [key: string]: any
}

export interface PaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}

// Auth
export const adminLogin = (password: string) =>
  http.post<{ token: string; expires_in: number }>('/admin/login', { password })

// Overview
export const adminOverview = () =>
  http.get('/admin/overview')

// Users
export const adminListUsers = (params: PaginationParams) =>
  http.get('/admin/users', { params })

export const adminGetUser = (id: number) =>
  http.get(`/admin/users/${id}`)

export const adminDeleteUser = (id: number) =>
  http.delete(`/admin/users/${id}`)

// Babies
export const adminListBabies = (params: PaginationParams) =>
  http.get('/admin/babies', { params })

// Mothers
export const adminListMothers = (params: PaginationParams) =>
  http.get('/admin/mothers', { params })

// Records
export const adminListRecords = (params: PaginationParams) =>
  http.get('/admin/records', { params })

export const adminGetRecord = (id: number) =>
  http.get(`/admin/records/${id}`)

export const adminDeleteRecord = (id: number) =>
  http.delete(`/admin/records/${id}`)

export const adminListRecordAttachments = (recordId: number) =>
  http.get(`/admin/records/${recordId}/attachments`)

// Mother Records
export const adminListMotherRecords = (params: PaginationParams) =>
  http.get('/admin/mother-records', { params })

export const adminGetMotherRecord = (id: number) =>
  http.get(`/admin/mother-records/${id}`)

export const adminDeleteMotherRecord = (id: number) =>
  http.delete(`/admin/mother-records/${id}`)

export const adminListMotherRecordAttachments = (recordId: number) =>
  http.get(`/admin/mother-records/${recordId}/attachments`)

// Reminders
export const adminListReminders = (params: PaginationParams) =>
  http.get('/admin/reminders', { params })

export const adminReminderStats = () =>
  http.get('/admin/reminders/stats')

// Analytics
export const adminAnalyticsRecords = (days?: number) =>
  http.get('/admin/analytics/records', { params: { days } })

export const adminAnalyticsActivity = (days?: number) =>
  http.get('/admin/analytics/activity', { params: { days } })

export const adminAnalyticsGrowth = (babyId: number) =>
  http.get('/admin/analytics/growth', { params: { baby_id: babyId } })
