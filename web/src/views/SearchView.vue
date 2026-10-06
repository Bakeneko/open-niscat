<script setup lang="ts">
import { mdiMagnify } from '@mdi/js'
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useDisplay } from 'vuetify'
import { useRoute, useRouter, type LocationQuery } from 'vue-router'
import { apiPath } from '@/api/client'
import type { SearchResult } from '@/api/types'
import ErrorAlert from '@/components/ErrorAlert.vue'
import { useFetch, type Fetched } from '@/composables/useFetch'
import { useLang } from '@/composables/useLang'
import { useLinks } from '@/composables/useLinks'
import { useScope } from '@/composables/useScope'
import { formatRange, formatYearMonth } from '@/lib/format'
import { refPath } from '@/lib/refs'
import { scopeToQuery } from '@/lib/scope'
import { defaultTab, isSearchTab, SEARCH_TABS, type SearchTab } from '@/lib/search'

const PAGE = 50
const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const lang = useLang()
const links = useLinks()
const { current } = useScope()
const { xs } = useDisplay()

const q = computed(() => (typeof route.query.q === 'string' ? route.query.q : ''))
const page = computed(() => Math.max(1, Number(route.query.page) || 1))
// "All vehicles" ignores the active vehicle for this search only (kept in the URL, so shareable).
const all = computed(() => route.query.all === '1')
const scope = computed(() => (all.value ? null : current.value))

// One request per tab, so every tab shows its count; only the open tab is paged.
const explicit = computed(() => (isSearchTab(route.query.type) ? route.query.type : null))
const fetches = Object.fromEntries(
  SEARCH_TABS.map((tab) => [
    tab,
    useFetch<SearchResult>(() =>
      q.value.trim() === ''
        ? null
        : apiPath('/api/search', {
            q: q.value,
            type: tab,
            offset: tab === explicit.value ? (page.value - 1) * PAGE : 0,
            limit: PAGE,
            lang: lang.value,
            ...scopeToQuery(scope.value),
          }),
    ),
  ]),
) as Record<SearchTab, Fetched<SearchResult>>
const loaded = computed(() => SEARCH_TABS.every((tab) => fetches[tab].data.value !== null))
const type = computed<SearchTab | null>(() => {
  if (explicit.value !== null) return explicit.value
  if (!loaded.value) return null
  const counts: Partial<Record<SearchTab, number>> = {}
  for (const tab of SEARCH_TABS) counts[tab] = fetches[tab].data.value?.total ?? 0
  return defaultTab(counts)
})
const active = computed(() => (type.value === null ? null : fetches[type.value]))
const data = computed(() => active.value?.data.value ?? null)
const loading = computed(() => SEARCH_TABS.some((tab) => fetches[tab].loading.value))
const error = computed(
  () => SEARCH_TABS.map((tab) => fetches[tab].error.value).find((e) => e !== null) ?? null,
)
function reload() {
  for (const tab of SEARCH_TABS) fetches[tab].reload()
}
function count(tab: SearchTab): string {
  const d = fetches[tab].data.value
  return d === null ? '' : ` (${String(d.total)}${d.truncated ? '+' : ''})`
}
const pages = computed(() => Math.max(1, Math.ceil((data.value?.total ?? 0) / PAGE)))

function setQuery(patch: Record<string, string>) {
  void router.replace({ query: { ...route.query, ...patch } })
}

// The page has its own field: on phones the header search box is hidden.
const text = ref('')
watch(
  q,
  (v) => {
    text.value = v
  },
  { immediate: true },
)
// A new page starts from its first results.
function setPage(n: number) {
  setQuery({ page: String(n) })
  window.scrollTo({ top: 0 })
}
// Only a user's choice is written to the URL: v-tabs selects its first tab by itself while results load.
function pickTab(v: unknown) {
  if (type.value !== null && isSearchTab(v) && v !== type.value) setQuery({ type: v, page: '1' })
}
// A new search lets the tab be chosen again from the results.
function submit() {
  const v = text.value.trim()
  if (v === '') return
  const query: LocationQuery = { ...route.query, q: v, page: '1' }
  delete query.type
  void router.replace({ query })
}
function setAll(v: boolean | null) {
  const query: LocationQuery = { ...route.query, page: '1' }
  if (v === true) query.all = '1'
  else delete query.all
  void router.replace({ query })
}
</script>

<template>
  <v-container>
    <v-text-field
      v-model="text"
      :prepend-inner-icon="mdiMagnify"
      :label="t('home.searchLabel')"
      density="compact"
      hide-details
      autofocus
      class="mb-2"
      @keyup.enter="submit"
    />
    <v-tabs :model-value="type" class="mb-2" @update:model-value="pickTab">
      <v-tab v-for="tab in SEARCH_TABS" :key="tab" :value="tab"
        >{{ t(`search.${tab}`) }}{{ count(tab) }}</v-tab
      >
    </v-tabs>
    <div v-if="current" class="d-flex align-center flex-wrap ga-4 mb-2">
      <span v-if="!all" class="text-caption">{{ t('search.scoped') }}</span>
      <v-switch
        :model-value="all"
        :label="t('search.allVehicles')"
        density="compact"
        hide-details
        class="ms-2 flex-grow-0"
        @update:model-value="setAll"
      />
    </div>
    <v-progress-linear v-if="loading" indeterminate />
    <ErrorAlert :error="error" @retry="reload" />
    <template v-if="data">
      <!-- One results bar, kept under the header while scrolling: count and pages. -->
      <div class="results-bar d-flex align-center flex-wrap ga-2 mb-2">
        <span class="text-caption">{{ t('search.results', { n: data.total }) }}</span>
        <v-spacer />
        <v-pagination
          v-if="pages > 1"
          :model-value="page"
          :length="pages"
          :total-visible="xs ? 3 : 7"
          density="compact"
          @update:model-value="setPage"
        />
      </div>
      <v-alert
        v-if="data.truncated"
        type="warning"
        variant="tonal"
        density="compact"
        class="mb-2"
        >{{ t('search.truncated', { n: data.total }) }}</v-alert
      >
      <v-alert v-if="data.total === 0" type="info" variant="tonal" density="compact">{{
        t('search.none')
      }}</v-alert>
      <v-list v-if="type === 'vins'" density="compact" lines="two">
        <v-list-item
          v-for="v in data.vins"
          :key="v.vin"
          :to="links.to(`/vin/${encodeURIComponent(v.vin)}`, {}, false)"
          :title="v.vin"
          :subtitle="
            [v.cat, v.model, formatYearMonth(v.prodDate, true)].filter(Boolean).join(' · ')
          "
        />
      </v-list>
      <v-list v-else-if="type === 'parts'" density="compact" lines="two">
        <v-list-item
          v-for="p in data.parts"
          :key="p.id"
          :to="links.to(`/section/${p.etd}/${p.sec}`, { item: p.itemKey })"
          :title="`${p.partNo} — ${p.description}`"
          :subtitle="`${p.etd} ${p.sec} · ${t('part.item')} ${p.item || p.itemKey} · ${formatRange(p.from, p.to)}`"
        >
          <template #append>
            <RouterLink
              :to="links.to(refPath(p.partNo), {}, false)"
              class="text-caption"
              @click.stop
              >{{ t('part.occurrences') }}</RouterLink
            >
          </template>
        </v-list-item>
      </v-list>
      <v-list v-else density="compact" lines="two">
        <v-list-item
          v-for="s in data.sections"
          :key="`${s.etd}${s.sec}`"
          :to="links.to(`/section/${s.etd}/${s.sec}`)"
          :subtitle="`${s.group.code} ${s.group.label}${s.notes ? ' · ' + s.notes : ''}`"
        >
          <template #title>
            <span :title="s.nameEn">{{ s.etd }} {{ s.sec }} — {{ s.name }}</span>
          </template>
        </v-list-item>
      </v-list>
    </template>
  </v-container>
</template>

<style scoped>
.results-bar {
  position: sticky;
  top: var(--v-layout-top, 0px);
  z-index: 2;
  min-height: 40px;
  background: rgb(var(--v-theme-background));
}
</style>
