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

const L = '/:lang(fr)?'

const routes: RouteRecordRaw[] = [
  { path: L, component: () => import('@/views/HomeView.vue') },
  { path: `${L}/vin/:vin`, component: () => import('@/views/VinView.vue') },
  { path: `${L}/catalogs`, component: () => import('@/views/CatalogsView.vue') },
  { path: `${L}/catalogs/:cat`, component: () => import('@/views/ModelsView.vue') },
  { path: `${L}/vehicle`, component: () => import('@/views/VehicleView.vue'), meta: { scoped: true } },
  { path: `${L}/index/:cat/:group`, component: () => import('@/views/GroupView.vue'), meta: { scoped: true } },
  { path: `${L}/section/:etd/:sec`, component: () => import('@/views/SectionView.vue'), meta: { scoped: true } },
  { path: `${L}/search`, component: () => import('@/views/SearchView.vue'), meta: { scoped: true } },
  { path: `${L}/part/:ref`, component: () => import('@/views/PartView.vue') },
  { path: `${L}/cart`, component: () => import('@/views/CartView.vue') },
  { path: '/:pathMatch(.*)*', component: () => import('@/views/NotFoundView.vue') },
]

const router = createRouter({ history: createWebHistory(), routes })

let firstNavigation = true

router.beforeEach(async (to) => {
  const first = firstNavigation
  firstNavigation = false

  // Language on "/": stored preference, else the server default (spec §4).
  if (first && to.path === '/' && Object.keys(to.query).length === 0) {
    const stored = load('open-niscat.lang', isLang)
    const pref = stored ?? (await useMeta().then((m) => m.defaultLang).catch(() => 'en' as const))
    if (pref === 'fr') return { path: '/fr', replace: true }
  }

  // Pages that need a vehicle get the remembered one when the URL has none (spec §5).
  if (to.meta.scoped === true && scopeFromQuery(to.query) === null) {
    const saved = useSavedScope().value
    if (saved !== null) return { path: to.path, query: { ...to.query, ...scopeToQuery(saved) }, hash: to.hash, replace: true }
  }
  return true
})

export default router
