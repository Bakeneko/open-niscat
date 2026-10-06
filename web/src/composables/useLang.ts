import { computed, type ComputedRef } from 'vue'
import { useRoute } from 'vue-router'
import { langFromParam, type Lang } from '@/lib/lang'

export function useLang(): ComputedRef<Lang> {
  const route = useRoute()
  return computed(() => langFromParam(route.params.lang))
}
