export type Link = {
  id: string
  short: string
  url: string
  created_at: string
  updated_at: string
  created_by: string
  access_level: string
  allowed_users: string[]
  click_count: number
  expires_at?: string
  is_expired: boolean
}
