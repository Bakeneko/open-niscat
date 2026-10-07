<script setup lang="ts">
import { mdiMagnify } from '@mdi/js'
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useHistory } from '@/composables/useHistory'
import { useLang } from '@/composables/useLang'
import { useLinks } from '@/composables/useLinks'
import type { Meta } from '@/api/types'
import { useMeta } from '@/composables/useMeta'
import { usePageTitle } from '@/composables/usePageTitle'
import { formatYearMonth } from '@/lib/format'
import { dedupeHistory } from '@/lib/history'
import { switchLang } from '@/lib/lang'

const { t } = useI18n()
const router = useRouter()
const links = useLinks()
const lang = useLang()
const history = useHistory()
// The home page shows the latest few; the rest of the stored history is one click away.
const HISTORY_SHOWN = 8
const allHistory = ref(false)
const historyList = computed(() => dedupeHistory(history.entries.value)) // older entries too
const shownHistory = computed(() =>
  allHistory.value ? historyList.value : historyList.value.slice(0, HISTORY_SHOWN),
)
const vin = ref('')
const q = ref('')
const meta = ref<Meta | null>(null)

onMounted(() => {
  useMeta()
    .then((m) => {
      meta.value = m
    })
    .catch(() => undefined)
})

function identify() {
  const v = vin.value.replace(/\s+/g, '')
  if (v !== '') void router.push(links.to(`/vin/${encodeURIComponent(v)}`, {}, false))
}
function search() {
  const text = q.value.trim()
  if (text !== '') void router.push(links.to('/search', { q: text }))
}

usePageTitle(() => null)
</script>

<template>
  <v-container class="home py-6" style="max-width: 760px">
    <v-card class="mb-4">
      <v-card-text>
        <v-text-field
          v-model="vin"
          :label="t('home.vinLabel')"
          :hint="t('home.vinHint')"
          persistent-hint
          autofocus
          @keyup.enter="identify"
        >
          <template #append>
            <v-btn color="primary" @click="identify">{{ t('home.find') }}</v-btn>
          </template>
        </v-text-field>
      </v-card-text>
    </v-card>
    <v-card class="mb-4">
      <v-card-text>
        <v-text-field
          v-model="q"
          :label="t('home.searchLabel')"
          :prepend-inner-icon="mdiMagnify"
          hide-details
          @keyup.enter="search"
        />
      </v-card-text>
      <v-card-actions>
        <v-btn :to="links.to('/catalogs', {}, false)">{{ t('home.browse') }}</v-btn>
      </v-card-actions>
    </v-card>
    <v-card v-if="historyList.length > 0">
      <v-card-title class="d-flex align-center text-subtitle-1">
        {{ t('home.history') }}
        <v-spacer />
        <v-btn size="small" variant="text" @click="history.clear()">{{
          t('home.historyClear')
        }}</v-btn>
      </v-card-title>
      <v-list density="compact">
        <!-- Stored without a language prefix: entries open in the current language. -->
        <v-list-item
          v-for="h in shownHistory"
          :key="h.path"
          :to="switchLang(h.path, lang)"
          :title="h.label"
        />
      </v-list>
      <v-card-actions v-if="historyList.length > HISTORY_SHOWN">
        <v-btn size="small" @click="allHistory = !allHistory">{{
          allHistory ? t('home.historyLess') : t('home.historyAll', { n: historyList.length })
        }}</v-btn>
      </v-card-actions>
    </v-card>
    <footer v-if="meta" class="text-caption text-medium-emphasis text-center pt-6 mt-auto">
      {{
        t('app.version', {
          app: meta.appVersion,
          edition: formatYearMonth(meta.source.edition, true),
          data: meta.version,
        })
      }}
    </footer>
  </v-container>
</template>

<style scoped>
/* Fills the viewport under the app bar so the version footer sits at the bottom of short pages. */
.home {
  display: flex;
  flex-direction: column;
  min-height: calc(100dvh - var(--v-layout-top, 0px));
}
</style>
