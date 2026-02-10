import request from '@/utils/request'
import type { PricingResponse } from '@/types/billing'

export const billingAPI = {
  getPricing() {
    return request.get<PricingResponse>('/billing/pricing')
  }
}

