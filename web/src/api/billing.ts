import request from '@/utils/request'
import type { BillingTransactionQuery, CreditTransaction, PaginationResult, PricingResponse } from '@/types/billing'

export const billingAPI = {
  getPricing() {
    return request.get<PricingResponse>('/billing/pricing')
  },
  listTransactions(params?: BillingTransactionQuery) {
    return request.get<PaginationResult<CreditTransaction>>('/billing/transactions', { params })
  }
}
