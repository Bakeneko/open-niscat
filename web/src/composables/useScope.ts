import { computed, type Ref } from 'vue'
import { useRoute } from 'vue-router'
import { isScopeOrNull, sameScope, scopeFromQuery, type Scope } from '@/lib/scope'
import { useStored } from './useStored'

export const SCOPE_KEY = 'open-niscat.scope'

export function useSavedScope(): Ref<Scope | null> {
  return useStored<Scope | null>(SCOPE_KEY, isScopeOrNull, () => null)
}

/** The page's vehicle (from the URL) and the remembered one (localStorage). */
export function useScope() {
  const route = useRoute()
  const saved = useSavedScope()
  const current = computed(() => scopeFromQuery(route.query))
  const isSaved = computed(() => current.value !== null && sameScope(current.value, saved.value))
  return {
    current,
    saved,
    isSaved,
    adopt(s: Scope | null = current.value) {
      saved.value = s
    },
    clear() {
      saved.value = null
    },
  }
}
