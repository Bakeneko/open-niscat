import type { RouteLocationRaw } from 'vue-router'
import { useRoute } from 'vue-router'
import { localizedPath } from '@/lib/lang'
import { scopeFromQuery, scopeToQuery } from '@/lib/scope'
import { useLang } from './useLang'

/** Builds localized links that carry the current vehicle scope (unless keepScope is false). */
export function useLinks() {
  const lang = useLang()
  const route = useRoute()
  return {
    to(path: string, query: Record<string, string> = {}, keepScope = true): RouteLocationRaw {
      const scope = keepScope ? scopeToQuery(scopeFromQuery(route.query)) : {}
      return { path: localizedPath(lang.value, path), query: { ...scope, ...query } }
    },
  }
}
