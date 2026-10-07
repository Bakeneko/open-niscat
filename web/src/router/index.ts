import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useMeta } from '@/composables/useMeta'
import { useSavedScope } from '@/composables/useScope'
import { isLang } from '@/lib/lang'
import { scopeFromQuery, scopeToQuery } from '@/lib/scope'
import { load } from '@/lib/storage'

declare module 'vue-router' {
  interface RouteMeta {
    scoped?: boolean
  }
}

const L = '/:lang(fr|es|de)?'

const routes: RouteRecordRaw[] = [
  { path: L, component: () => import('@/views/HomeView.vue') },
  { path: `${L}/vin/:vin`, component: () => import('@/views/VinView.vue') },
  { path: `${L}/catalogs`, component: () => import('@/views/CatalogsView.vue') },
  { path: `${L}/catalogs/:cat`, component: () => import('@/views/ModelsView.vue') },
  {
    path: `${L}/vehicle`,
    component: () => import('@/views/VehicleView.vue'),
    meta: { scoped: true },
  },
  {
    path: `${L}/index/:cat/:group`,
    component: () => import('@/views/GroupView.vue'),
    meta: { scoped: true },
  },
  {
    path: `${L}/section/:etd/:sec`,
    component: () => import('@/views/SectionView.vue'),
    meta: { scoped: true },
  },
  {
    path: `${L}/search`,
    component: () => import('@/views/SearchView.vue'),
    meta: { scoped: true },
  },
  { path: `${L}/part/:ref`, component: () => import('@/views/PartView.vue') },
  { path: `${L}/list`, component: () => import('@/views/PartsListView.vue') },
  { path: '/:pathMatch(.*)*', component: () => import('@/views/NotFoundView.vue') },
]

const router = createRouter({ history: createWebHistory(), routes })

let firstNavigation = true

router.beforeEach(async (to) => {
  const first = firstNavigation
  firstNavigation = false

  // Bare "/" on first load: the stored language, else the server default; English stays unprefixed.
  if (first && to.path === '/' && Object.keys(to.query).length === 0) {
    const stored = load('open-niscat.lang', isLang)
    const pref =
      stored ??
      (await useMeta()
        .then((m) => m.defaultLang)
        .catch(() => 'en' as const))
    if (pref !== 'en') return { path: `/${pref}`, replace: true }
  }

  // Vehicle scope guard: a scoped page without a vehicle in the URL gets the remembered one, so the URL
  // stays the source of truth (shareable).
  if (to.meta.scoped === true && scopeFromQuery(to.query) === null) {
    const saved = useSavedScope().value
    const query = { ...to.query, ...scopeToQuery(saved) }
    if (scopeFromQuery(query) !== null)
      return { path: to.path, query, hash: to.hash, replace: true }
  }
  return true
})

// After a binary upgrade an open tab still references old hashed chunks, which now answer 404: load the
// target page in full once (the reload fetches the new index.html).
router.onError((err: unknown, to) => {
  if (!/dynamically imported module|Importing a module script failed/i.test(String(err))) return
  try {
    if (sessionStorage.getItem('open-niscat.reload') === to.fullPath) return
    sessionStorage.setItem('open-niscat.reload', to.fullPath)
  } catch {
    return
  }
  window.location.assign(to.fullPath)
})

export default router
