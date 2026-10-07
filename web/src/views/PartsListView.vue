<script setup lang="ts">
import { mdiContentCopy, mdiDelete, mdiFileDelimited, mdiLink, mdiPrinter } from '@mdi/js'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { apiPath } from '@/api/client'
import type { LineRef, LinesResult } from '@/api/types'
import ErrorAlert from '@/components/ErrorAlert.vue'
import { usePartsList } from '@/composables/usePartsList'
import { useFetch } from '@/composables/useFetch'
import { useLang } from '@/composables/useLang'
import { useLinks } from '@/composables/useLinks'
import { useNotify } from '@/composables/useNotify'
import { usePageTitle } from '@/composables/usePageTitle'
import { decodeList, encodeList, toCSV, toTSV, type ListItem, type ListRow } from '@/lib/partsList'
import { localizedPath } from '@/lib/lang'
import { readable } from '@/lib/format'
import { refPath } from '@/lib/refs'
import { scopeLabel, scopeToQuery } from '@/lib/scope'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const lang = useLang()
const links = useLinks()
const partsList = usePartsList()
const { notify } = useNotify()

// A ?items= share link shows that list read-only, without touching the stored one until it is adopted.
const shared = computed(() =>
  typeof route.query.items === 'string' ? decodeList(route.query.items) : null,
)
const items = computed<ListItem[]>(() => shared.value?.items ?? partsList.items.value)
const { data, error, loading, reload } = useFetch<LinesResult>(() =>
  items.value.length === 0
    ? null
    : apiPath('/api/lines', { ids: items.value.map((i) => i.id).join(','), lang: lang.value }),
)
const byId = computed(() => new Map((data.value?.lines ?? []).map((l) => [l.id, l])))
const resolved = computed(() =>
  items.value.flatMap((item) => {
    const line = byId.value.get(item.id)
    return line === undefined ? [] : [{ item, line }]
  }),
)

function sectionLink(item: ListItem, l: LineRef) {
  return {
    path: localizedPath(lang.value, `/section/${l.etd}/${l.sec}`),
    query: { ...scopeToQuery(item.scope ?? null), item: l.itemKey },
  }
}

const headers = computed(() => [
  t('part.reference'),
  t('part.description'),
  t('list.qty'),
  t('list.section'),
])
const rows = computed<ListRow[]>(() =>
  items.value.flatMap((i) => {
    const l = byId.value.get(i.id)
    return l === undefined
      ? []
      : [
          {
            reference: l.partNo,
            description: l.description,
            qty: i.qty,
            section: `${l.etd} ${l.sec} / ${l.callout}`,
          },
        ]
  }),
)

async function copy(text: string, done: string) {
  try {
    await navigator.clipboard.writeText(text)
    notify(done)
  } catch {
    // Clipboard unavailable (e.g. plain http on another machine): show the text to copy by hand.
    notify(text)
  }
}
function copyTable() {
  void copy(toTSV(headers.value, rows.value), t('list.copied'))
}
// Share links carry ids and quantities only, not the vehicle of each line.
function shareLink() {
  const url = `${window.location.origin}${localizedPath(lang.value, '/list')}?items=${encodeList(items.value)}`
  void copy(url, t('list.linkCopied'))
}
function downloadCSV() {
  // The leading BOM makes Excel read the file as UTF-8.
  const blob = new Blob(['\uFEFF', toCSV(headers.value, rows.value)], {
    type: 'text/csv;charset=utf-8',
  })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = `open-niscat-parts-list-${new Date().toISOString().slice(0, 10)}.csv`
  a.click()
  URL.revokeObjectURL(a.href)
}
function print() {
  window.print()
}
function adoptShared(mode: 'replace' | 'merge') {
  const s = shared.value
  if (s === null) return
  if (mode === 'replace') partsList.replace(s.items)
  else partsList.merge(s.items)
  void router.replace(links.to('/list', {}, false)) // back to the stored list (drops ?items=)
}
// Only whole quantities >= 1 are applied: an emptied field while typing must not delete the line.
function setQty(id: string, v: unknown) {
  const n = Number(v)
  if (Number.isInteger(n) && n >= 1) partsList.setQty(id, n)
}
const confirming = ref(false)
function clearList() {
  partsList.clear()
  confirming.value = false
}
function removeMissing() {
  for (const id of data.value?.missing ?? []) partsList.remove(id)
}

usePageTitle(() => (shared.value ? t('list.shared') : t('list.title')))
</script>

<template>
  <v-container>
    <div class="d-flex align-center flex-wrap ga-2 mb-2">
      <h1 class="text-h6">{{ shared ? t('list.shared') : t('list.title') }}</h1>
      <v-spacer />
      <template v-if="shared && shared.items.length > 0">
        <v-btn color="primary" class="no-print" @click="adoptShared('replace')">{{
          t('list.replace')
        }}</v-btn>
        <v-btn variant="tonal" class="no-print" @click="adoptShared('merge')">{{
          t('list.merge')
        }}</v-btn>
      </template>
      <template v-if="rows.length > 0">
        <v-btn :prepend-icon="mdiContentCopy" variant="text" class="no-print" @click="copyTable">{{
          t('list.copy')
        }}</v-btn>
        <v-btn
          :prepend-icon="mdiFileDelimited"
          variant="text"
          class="no-print"
          @click="downloadCSV"
          >{{ t('list.csv') }}</v-btn
        >
        <v-btn :prepend-icon="mdiPrinter" variant="text" class="no-print" @click="print">{{
          t('list.print')
        }}</v-btn>
        <v-btn :prepend-icon="mdiLink" variant="text" class="no-print" @click="shareLink">{{
          t('list.share')
        }}</v-btn>
      </template>
      <v-btn
        v-if="!shared && items.length > 0 && !confirming"
        :prepend-icon="mdiDelete"
        variant="text"
        color="error"
        class="no-print"
        @click="confirming = true"
        >{{ t('list.clear') }}</v-btn
      >
    </div>
    <v-alert
      v-if="confirming"
      type="warning"
      variant="tonal"
      density="compact"
      class="mb-2 no-print"
    >
      {{ t('list.confirmClear') }}
      <template #append>
        <v-btn color="error" variant="text" @click="clearList">{{ t('list.confirm') }}</v-btn>
        <v-btn variant="text" @click="confirming = false">{{ t('list.cancel') }}</v-btn>
      </template>
    </v-alert>
    <v-alert
      v-if="shared && shared.invalid.length > 0"
      type="warning"
      variant="tonal"
      density="compact"
      class="mb-2"
    >
      {{ t('list.invalid', { n: shared.invalid.length }) }}
    </v-alert>
    <v-alert
      v-if="data && data.missing.length > 0"
      type="warning"
      variant="tonal"
      density="compact"
      class="mb-2"
    >
      {{ t('list.missing', { n: data.missing.length }) }}: {{ data.missing.join(', ') }}
      <template v-if="!shared" #append>
        <v-btn variant="text" class="no-print" @click="removeMissing">{{
          t('list.removeMissing')
        }}</v-btn>
      </template>
    </v-alert>
    <v-alert v-if="items.length === 0" type="info" variant="tonal">{{ t('list.empty') }}</v-alert>
    <v-progress-linear v-if="loading" indeterminate />
    <ErrorAlert :error="error" @retry="reload" />
    <v-table v-if="data && items.length > 0" density="compact" hover class="list-table">
      <thead>
        <tr>
          <th>{{ t('part.part') }}</th>
          <th>{{ t('list.section') }}</th>
          <th class="text-right">{{ t('list.qty') }}</th>
          <th v-if="!shared" class="no-print" />
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="{ item: i, line: l } in resolved"
          :key="i.id"
          class="list-row"
          @click="router.push(sectionLink(i, l))"
        >
          <!-- Same "Part" cell as the plate table: number, then description. -->
          <td class="part-cell">
            <RouterLink
              :to="links.to(refPath(l.partNo), {}, false)"
              class="font-weight-bold text-no-wrap"
              @click.stop
              >{{ l.partNo }}</RouterLink
            >
            <div>{{ readable(l.description) }}</div>
          </td>
          <!-- The vehicle the line was added for sits under its plate: one column less on phones. -->
          <td>
            <RouterLink :to="sectionLink(i, l)" class="text-no-wrap" @click.stop
              >{{ l.etd }} {{ l.sec }} / {{ l.callout }}</RouterLink
            >
            <div v-if="i.scope" class="text-caption text-medium-emphasis vehicle-label">
              {{ scopeLabel(i.scope) }}
            </div>
          </td>
          <td class="text-right" @click.stop>
            <span v-if="shared">{{ i.qty }}</span>
            <v-text-field
              v-else
              :model-value="i.qty"
              type="number"
              min="1"
              density="compact"
              hide-details
              class="qty-field"
              hide-spin-buttons
              @update:model-value="(v) => setQty(i.id, v)"
            />
          </td>
          <td v-if="!shared" class="no-print" @click.stop>
            <v-btn
              :icon="mdiDelete"
              size="small"
              variant="text"
              :title="t('list.remove')"
              @click="partsList.remove(i.id)"
            />
          </td>
        </tr>
      </tbody>
    </v-table>
  </v-container>
</template>

<style scoped>
.list-row {
  cursor: pointer;
}
.part-cell {
  padding-top: 4px !important;
  padding-bottom: 4px !important;
}
.qty-field {
  width: 48px;
  margin-left: auto;
}
.qty-field :deep(.v-field__input) {
  padding-inline: 4px;
  text-align: center;
}
.vehicle-label {
  overflow-wrap: anywhere;
}
/* Tight cells, as in the plate table: the list must fit a phone screen. */
.list-table :deep(th),
.list-table :deep(td) {
  padding: 0 5px;
}
</style>
