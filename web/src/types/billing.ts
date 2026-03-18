import type { AIServiceType } from './ai'

export interface ServicePricing {
  service_type: AIServiceType
  config_id: number
  model: string
  credit_cost: number
}

export interface AIServiceConfigPricingView {
  id: number
  user_id: number
  service_type: AIServiceType
  provider: string
  name: string
  base_url: string
  api_key_set: boolean
  model: string[]
  credit_cost: number
  endpoint: string
  query_endpoint: string
  priority: number
  is_default: boolean
  is_active: boolean
  settings?: string
  created_at?: string
  updated_at?: string
}

export interface PricingResponse {
  defaults: ServicePricing[]
  user_configs: AIServiceConfigPricingView[]
  platform_configs: AIServiceConfigPricingView[]
}

export interface CreditTransaction {
  id: number
  user_id: number
  amount: number
  type: string
  reference_id?: string
  service_type?: string
  model?: string
  description?: string
  prompt_tokens?: number
  completion_tokens?: number
  total_tokens?: number
  created_at?: string
}

export interface BillingTransactionQuery {
  page?: number
  page_size?: number
}

export interface PaginationResult<T> {
  items: T[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_pages: number
  }
}
