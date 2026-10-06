<script setup lang="ts">
import { mdiMagnify } from '@mdi/js'
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { apiPath } from '@/api/client'
import type { SearchResult } from '@/api/types'
import ErrorAlert from '@/components/ErrorAlert.vue'
import { useFetch } from '@/composables/useFetch'
import { useLang } from '@/composables/useLang'
import { useLinks } from '@/composables/useLinks'
import { useScope } from '@/composables/useScope'
import { formatRange } from '@/lib/format'
import { refPath } from '@/lib/refs'
import { scopeToQuery } from '@/lib/scope'

const PAGE = 50
const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const lang = useLang()
const links = useLinks()
const { current } = useScope()

const q = computed(() => (typeof route.query.q === 'string' ? route.query.q : ''))
const type = computed(() => (route.query.type === 'sections' ? 'sections' : 'parts'))
const page = computed(() => Math.max(1, Number(route.query.page) || 1))
const { data, error, loading, reload } = useFetch<SearchResult>(() =>
  q.value.trim() === ''
    ? null
    : apiPath('/api/search', {
        q: q.value,
        type: type.value,
        offset: (page.value - 1) * PAGE,
        limit: PAGE,
        lang: lang.value,
        ...scopeToQuery(current.value),
      }),
)
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
function submit() {
  const v = text.value.trim()
  if (v !== '') setQuery({ q: v, page: '1' })
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
    <v-tabs
      :model-value="type"
      class="mb-2"
      @update:model-value="(v) => setQuery({ type: String(v), page: '1' })"
    >
      <v-tab value="parts">{{ t('search.parts') }}</v-tab>
      <v-tab value="sections">{{ t('search.sections') }}</v-tab>
    </v-tabs>
    <div v-if="current" class="text-caption mb-2">{{ t('search.scoped') }}</div>
    <v-progress-linear v-if="loading" indeterminate />
    <ErrorAlert :error="error" @retry="reload" />
    <template v-if="data">
      <div class="text-caption mb-2">{{ t('search.results', { n: data.total }) }}</div>
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
      <v-list v-if="type === 'parts'" density="compact" lines="two">
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
          :title="`${s.etd} ${s.sec} — ${s.name}`"
          :subtitle="`${s.group.code} ${s.group.label}${s.notes ? ' · ' + s.notes : ''}`"
        />
      </v-list>
      <v-pagination
        v-if="pages > 1"
        :model-value="page"
        :length="pages"
        :total-visible="7"
        @update:model-value="(n) => setQuery({ page: String(n) })"
      />
    </template>
  </v-container>
</template>
