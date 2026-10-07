<script setup lang="ts">
import { mdiMagnify } from '@mdi/js'
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useHistory } from '@/composables/useHistory'
import { useLinks } from '@/composables/useLinks'
import type { Meta } from '@/api/types'
import { useMeta } from '@/composables/useMeta'
import { formatYearMonth } from '@/lib/format'

const { t } = useI18n()
const router = useRouter()
const links = useLinks()
const history = useHistory()
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
    <v-card v-if="history.entries.value.length > 0">
      <v-card-title class="text-subtitle-1">{{ t('home.history') }}</v-card-title>
      <v-list density="compact">
        <v-list-item
          v-for="h in history.entries.value"
          :key="h.path"
          :to="h.path"
          :title="h.label"
        />
      </v-list>
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
