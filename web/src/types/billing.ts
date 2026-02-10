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
