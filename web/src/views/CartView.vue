<script setup lang="ts">
import { mdiContentCopy, mdiDelete, mdiFileDelimited, mdiLink, mdiPrinter } from '@mdi/js'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { apiPath } from '@/api/client'
import type { LineRef, LinesResult } from '@/api/types'
import ErrorAlert from '@/components/ErrorAlert.vue'
import { useCart } from '@/composables/useCart'
import { useFetch } from '@/composables/useFetch'
import { useLang } from '@/composables/useLang'
import { useLinks } from '@/composables/useLinks'
import { useNotify } from '@/composables/useNotify'
import { decodeCart, encodeCart, toCSV, toTSV, type CartItem, type CartRow } from '@/lib/cart'
import { localizedPath } from '@/lib/lang'
import { refPath } from '@/lib/refs'
import { scopeLabel, scopeToQuery } from '@/lib/scope'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const lang = useLang()
const links = useLinks()
const cart = useCart()
const { notify } = useNotify()

const shared = computed(() =>
  typeof route.query.items === 'string' ? decodeCart(route.query.items) : null,
)
const items = computed<CartItem[]>(() => shared.value?.items ?? cart.items.value)
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

function sectionLink(item: CartItem, l: LineRef) {
  return {
    path: localizedPath(lang.value, `/section/${l.etd}/${l.sec}`),
    query: { ...scopeToQuery(item.scope ?? null), item: l.itemKey },
  }
}

const headers = computed(() => [
  t('part.reference'),
  t('part.description'),
  t('cart.qty'),
  t('cart.section'),
])
const rows = computed<CartRow[]>(() =>
  items.value.flatMap((i) => {
    const l = byId.value.get(i.id)
    return l === undefined
      ? []
      : [
          {
            reference: l.partNo,
            description: l.description,
            qty: i.qty,
            section: `${l.etd} ${l.sec} / ${l.item || l.itemKey}`,
          },
        ]
  }),
)

async function copy(text: string, done: string) {
  try {
    await navigator.clipboard.writeText(text)
    notify(done)
  } catch {
    notify(text)
  }
}
function copyTable() {
  void copy(toTSV(headers.value, rows.value), t('cart.copied'))
}
function shareLink() {
  const url = `${window.location.origin}${localizedPath(lang.value, '/cart')}?items=${encodeCart(items.value)}`
  void copy(url, t('cart.linkCopied'))
}
function downloadCSV() {
  const blob = new Blob(['﻿', toCSV(headers.value, rows.value)], { type: 'text/csv;charset=utf-8' })
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = `open-niscat-cart-${new Date().toISOString().slice(0, 10)}.csv`
  a.click()
  URL.revokeObjectURL(a.href)
}
function print() {
  window.print()
}
function adoptShared(mode: 'replace' | 'merge') {
  const s = shared.value
  if (s === null) return
  if (mode === 'replace') cart.replace(s.items)
  else cart.merge(s.items)
  void router.replace(links.to('/cart', {}, false))
}
// Only whole quantities >= 1 are applied: an emptied field while typing must not delete the line.
function setQty(id: string, v: unknown) {
  const n = Number(v)
  if (Number.isInteger(n) && n >= 1) cart.setQty(id, n)
}
const confirming = ref(false)
function clearCart() {
  cart.clear()
  confirming.value = false
}
function removeMissing() {
  for (const id of data.value?.missing ?? []) cart.remove(id)
}
</script>

<template>
  <v-container>
    <div class="d-flex align-center flex-wrap ga-2 mb-2">
      <h1 class="text-h6">{{ shared ? t('cart.shared') : t('cart.title') }}</h1>
      <v-spacer />
      <template v-if="shared">
        <v-btn color="primary" class="no-print" @click="adoptShared('replace')">{{
          t('cart.replace')
        }}</v-btn>
        <v-btn variant="tonal" class="no-print" @click="adoptShared('merge')">{{
          t('cart.merge')
        }}</v-btn>
      </template>
      <template v-if="rows.length > 0">
        <v-btn :prepend-icon="mdiContentCopy" variant="text" class="no-print" @click="copyTable">{{
          t('cart.copy')
        }}</v-btn>
        <v-btn
          :prepend-icon="mdiFileDelimited"
          variant="text"
          class="no-print"
          @click="downloadCSV"
          >{{ t('cart.csv') }}</v-btn
        >
        <v-btn :prepend-icon="mdiPrinter" variant="text" class="no-print" @click="print">{{
          t('cart.print')
        }}</v-btn>
        <v-btn :prepend-icon="mdiLink" variant="text" class="no-print" @click="shareLink">{{
          t('cart.share')
        }}</v-btn>
      </template>
      <v-btn
        v-if="!shared && items.length > 0 && !confirming"
        :prepend-icon="mdiDelete"
        variant="text"
        color="error"
        class="no-print"
        @click="confirming = true"
        >{{ t('cart.clear') }}</v-btn
      >
    </div>
    <v-alert
      v-if="confirming"
      type="warning"
      variant="tonal"
      density="compact"
      class="mb-2 no-print"
    >
      {{ t('cart.confirmClear') }}
      <template #append>
        <v-btn color="error" variant="text" @click="clearCart">{{ t('cart.confirm') }}</v-btn>
        <v-btn variant="text" @click="confirming = false">{{ t('cart.cancel') }}</v-btn>
      </template>
    </v-alert>
    <v-alert
      v-if="shared && shared.invalid.length > 0"
      type="warning"
      variant="tonal"
      density="compact"
      class="mb-2"
    >
      {{ t('cart.invalid', { n: shared.invalid.length }) }}
    </v-alert>
    <v-alert
      v-if="data && data.missing.length > 0"
      type="warning"
      variant="tonal"
      density="compact"
      class="mb-2"
    >
      {{ t('cart.missing', { n: data.missing.length }) }}: {{ data.missing.join(', ') }}
      <template v-if="!shared" #append>
        <v-btn variant="text" class="no-print" @click="removeMissing">{{
          t('cart.removeMissing')
        }}</v-btn>
      </template>
    </v-alert>
    <v-alert v-if="items.length === 0" type="info" variant="tonal">{{ t('cart.empty') }}</v-alert>
    <v-progress-linear v-if="loading" indeterminate />
    <ErrorAlert :error="error" @retry="reload" />
    <v-table v-if="data && items.length > 0" density="compact" hover>
      <thead>
        <tr>
          <th>{{ t('part.reference') }}</th>
          <th>{{ t('part.description') }}</th>
          <th>{{ t('cart.section') }}</th>
          <th>{{ t('cart.vehicle') }}</th>
          <th class="text-right">{{ t('cart.qty') }}</th>
          <th v-if="!shared" class="no-print" />
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="{ item: i, line: l } in resolved"
          :key="i.id"
          class="cart-row"
          @click="router.push(sectionLink(i, l))"
        >
          <td class="text-no-wrap">
            <RouterLink :to="links.to(refPath(l.partNo), {}, false)" @click.stop>{{
              l.partNo
            }}</RouterLink>
          </td>
          <td>{{ l.description }}</td>
          <td class="text-no-wrap">
            <RouterLink :to="sectionLink(i, l)" @click.stop
              >{{ l.etd }} {{ l.sec }} / {{ l.item || l.itemKey }}</RouterLink
            >
          </td>
          <td class="text-caption">{{ i.scope ? scopeLabel(i.scope) : '' }}</td>
          <td class="text-right" @click.stop>
            <span v-if="shared">{{ i.qty }}</span>
            <v-text-field
              v-else
              :model-value="i.qty"
              type="number"
              min="1"
              density="compact"
              hide-details
              style="max-width: 90px; margin-left: auto"
              @update:model-value="(v) => setQty(i.id, v)"
            />
          </td>
          <td v-if="!shared" class="no-print" @click.stop>
            <v-btn
              :icon="mdiDelete"
              size="small"
              variant="text"
              :title="t('cart.remove')"
              @click="cart.remove(i.id)"
            />
          </td>
        </tr>
      </tbody>
    </v-table>
  </v-container>
</template>

<style scoped>
.cart-row {
  cursor: pointer;
}
</style>
