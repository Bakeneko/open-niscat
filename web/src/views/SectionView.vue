<script setup lang="ts">
import { mdiChevronLeft, mdiChevronRight, mdiPrinter } from '@mdi/js'
import { computed, nextTick, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useDisplay } from 'vuetify'
import { apiPath } from '@/api/client'
import type { Line, Section, Vehicle } from '@/api/types'
import DrawingViewer from '@/components/DrawingViewer.vue'
import ErrorAlert from '@/components/ErrorAlert.vue'
import PartsTable from '@/components/PartsTable.vue'
import { useCart } from '@/composables/useCart'
import { useFetch } from '@/composables/useFetch'
import { useHistory } from '@/composables/useHistory'
import { useLang } from '@/composables/useLang'
import { useLinks } from '@/composables/useLinks'
import { useNotify } from '@/composables/useNotify'
import { useScope } from '@/composables/useScope'
import { formatRange } from '@/lib/format'
import { scopeToQuery } from '@/lib/scope'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const lang = useLang()
const links = useLinks()
const cart = useCart()
const history = useHistory()
const { notify } = useNotify()
const { current } = useScope()
const { smAndDown } = useDisplay()

const etd = computed(() => String(route.params.etd).toUpperCase())
const sec = computed(() => String(route.params.sec).toUpperCase())
const { data, error, loading, reload } = useFetch<Section>(() =>
  apiPath(`/api/sections/${etd.value}/${sec.value}`, {
    lang: lang.value,
    ...scopeToQuery(current.value),
  }),
)

const selected = computed(() => {
  const v = route.query.item
  return typeof v === 'string' && v !== '' ? v.replace(/^0+(?=.)/, '') : null
})
const tab = computed(() => (route.query.tab === 'info' ? 'info' : 'drawing'))
const titles = computed(() => {
  const out: Record<string, string> = {}
  for (const l of data.value?.lines ?? []) out[l.itemKey] ??= `${l.partNo} ${l.description}`
  return out
})
// The group index URL needs the catalog id (series + period); a VIN scope only gives it through the vehicle.
const vehicle = useFetch<Vehicle>(() =>
  current.value === null
    ? null
    : apiPath('/api/vehicle', { lang: lang.value, ...scopeToQuery(current.value) }),
)
const catOfScope = computed(() => {
  const c = vehicle.data.value?.catalog
  return c?.etd === etd.value ? c.cat : undefined
})

function print() {
  window.print()
}

// Selecting the selected item again closes it.
function select(key: string) {
  const query = { ...route.query }
  if (key === selected.value) delete query.item
  else Object.assign(query, { item: key }, smAndDown.value ? { tab: 'info' } : {})
  void router.replace({ query })
}
function setTab(v: unknown) {
  void router.replace({ query: { ...route.query, tab: v === 'info' ? 'info' : 'drawing' } })
}
function add(line: Line) {
  cart.add(line.id, 1, current.value)
  notify(t('cart.added'))
}

watch(data, (s) => {
  if (s === null) return
  history.push('section', `${s.etd} ${s.sec} — ${s.name}`, route.fullPath)
})
// Also runs when the data arrives, so a shared link with ?item= scrolls to its row.
watch([selected, data], async ([key, s]) => {
  if (key === null || s === null) return
  await nextTick()
  document
    .querySelector(`.parts-table tr[data-item="${CSS.escape(key)}"]`)
    ?.scrollIntoView({ block: 'start', behavior: 'smooth' })
})
</script>

<template>
  <v-container fluid class="pa-2" :class="{ 'page-fill': !smAndDown }">
    <v-progress-linear v-if="loading" indeterminate />
    <ErrorAlert :error="error" @retry="reload" />
    <template v-if="data">
      <v-alert
        v-if="data.applicable === false"
        type="error"
        variant="tonal"
        density="compact"
        class="mb-2"
        >{{ t('section.notApplicable') }}</v-alert
      >
      <div class="d-flex align-center flex-wrap ga-2 mb-2">
        <h1 class="text-h6">{{ data.sec }} — {{ data.name }}</h1>
        <span class="text-medium-emphasis">{{
          [data.notes, formatRange(data.from, data.to)].filter((x) => x !== '').join(' · ')
        }}</span>
        <v-spacer />
        <v-btn
          v-if="catOfScope"
          size="small"
          variant="text"
          class="no-print"
          :to="links.to(`/index/${catOfScope}/${data.group.code}`)"
          >{{ t('section.group') }} {{ data.group.code }}</v-btn
        >
        <v-btn
          v-if="data.prev"
          size="small"
          variant="tonal"
          class="no-print"
          :prepend-icon="mdiChevronLeft"
          :to="links.to(`/section/${data.etd}/${data.prev}`)"
        >
          {{ data.prev }}
        </v-btn>
        <v-btn
          v-if="data.next"
          size="small"
          variant="tonal"
          class="no-print"
          :append-icon="mdiChevronRight"
          :to="links.to(`/section/${data.etd}/${data.next}`)"
        >
          {{ data.next }}
        </v-btn>
        <v-btn
          size="small"
          variant="text"
          class="no-print"
          :prepend-icon="mdiPrinter"
          @click="print"
          >{{ t('section.print') }}</v-btn
        >
      </div>

      <template v-if="smAndDown">
        <v-tabs :model-value="tab" grow class="sticky-tabs no-print" @update:model-value="setTab">
          <v-tab value="drawing">{{ t('section.drawing') }}</v-tab>
          <v-tab value="info">{{ t('section.info') }}</v-tab>
        </v-tabs>
        <div v-show="tab === 'drawing'" class="print-show">
          <DrawingViewer
            :src="data.image"
            :hotspots="data.hotspots"
            :selected="selected"
            :titles="titles"
            labels
            class="viewer-mobile"
            @select="select"
          />
        </div>
        <div v-show="tab === 'info'" class="print-show">
          <p v-if="selected === null" class="text-caption text-medium-emphasis mb-1 no-print">
            {{ t('section.select') }}
          </p>
          <PartsTable :lines="data.lines" :selected="selected" @select="select" @add="add" />
        </div>
      </template>
      <div v-else class="split">
        <DrawingViewer
          :src="data.image"
          :hotspots="data.hotspots"
          :selected="selected"
          :titles="titles"
          labels
          @select="select"
        />
        <div class="split-side">
          <p v-if="selected === null" class="text-caption text-medium-emphasis mb-1 no-print">
            {{ t('section.select') }}
          </p>
          <PartsTable :lines="data.lines" :selected="selected" @select="select" @add="add" />
        </div>
      </div>
    </template>
  </v-container>
</template>

<style scoped>
.viewer-mobile {
  height: 70vh;
}
/* Printed plate: drawing at full width, then the table; mobile tabs print both panes. */
@media print {
  .viewer-mobile {
    height: auto;
    overflow: visible;
  }
  .print-show {
    display: block !important;
  }
}
</style>
