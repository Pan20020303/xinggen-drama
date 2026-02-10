import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import type { AIServiceType } from '@/types/ai'
import { billingAPI } from '@/api/billing'
import type { PricingResponse, ServicePricing } from '@/types/billing'

export const usePricingStore = defineStore('pricing', () => {
  const loaded = ref(false)
  const loading = ref(false)
  const pricing = ref<PricingResponse | null>(null)

  const defaultsMap = computed(() => {
    const map = new Map<AIServiceType, ServicePricing>()
    for (const d of pricing.value?.defaults ?? []) {
      map.set(d.service_type, d)
    }
    return map
  })

  const getDefaultCost = (serviceType: AIServiceType): number => {
    return defaultsMap.value.get(serviceType)?.credit_cost ?? 0
  }

  const getDefaultModel = (serviceType: AIServiceType): string => {
    return defaultsMap.value.get(serviceType)?.model ?? ''
  }

  const loadPricing = async (force = false) => {
    if (loading.value) return
    if (loaded.value && !force) return
    loading.value = true
    try {
      pricing.value = await billingAPI.getPricing()
      loaded.value = true
    } finally {
      loading.value = false
    }
  }

  return {
    loaded,
    loading,
    pricing,
    getDefaultCost,
    getDefaultModel,
    loadPricing
  }
})

