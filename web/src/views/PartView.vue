<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { apiPath } from '@/api/client'
import type { PartInfo } from '@/api/types'
import ErrorAlert from '@/components/ErrorAlert.vue'
import { useFetch } from '@/composables/useFetch'
import { useLang } from '@/composables/useLang'
import { useLinks } from '@/composables/useLinks'
import { formatRange } from '@/lib/format'
import { refPath } from '@/lib/refs'

const { t } = useI18n()
const route = useRoute()
const lang = useLang()
const links = useLinks()
const partRef = computed(() => String(route.params.ref))
const { data, error, loading, reload } = useFetch<PartInfo>(() =>
  apiPath(`/api/parts/${encodeURIComponent(partRef.value)}`, { lang: lang.value }),
)
</script>

<template>
  <v-container>
    <v-progress-linear v-if="loading" indeterminate />
    <ErrorAlert :error="error" @retry="reload" />
    <template v-if="data">
      <h1 class="text-h6">{{ data.partNo }} — {{ data.description }}</h1>
      <div v-if="data.next.length > 0" class="my-2">
        <span class="text-subtitle-2 mr-2">{{ t('part.replacedBy') }}</span>
        <v-chip
          v-for="r in data.next"
          :key="r.key"
          :to="links.to(refPath(r.partNo), {}, false)"
          size="small"
          color="primary"
          class="ma-1"
          >{{ r.partNo }}</v-chip
        >
      </div>
      <div v-if="data.previous.length > 0" class="my-2">
        <span class="text-subtitle-2 mr-2">{{ t('part.replaces') }}</span>
        <v-chip
          v-for="r in data.previous"
          :key="r.key"
          :to="links.to(refPath(r.partNo), {}, false)"
          size="small"
          class="ma-1"
          >{{ r.partNo }}</v-chip
        >
      </div>
      <h2 class="text-subtitle-1 mt-4">{{ t('part.occurrences') }}</h2>
      <v-table density="compact">
        <thead>
          <tr>
            <th>{{ t('part.series') }}</th>
            <th>{{ t('part.section') }}</th>
            <th>{{ t('part.item') }}</th>
            <th>{{ t('part.description') }}</th>
            <th>{{ t('part.qty') }}</th>
            <th>{{ t('part.period') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="o in data.occurrences" :key="o.id">
            <td>{{ o.etd }}</td>
            <td>
              <RouterLink :to="links.to(`/section/${o.etd}/${o.sec}`, { item: o.itemKey })">{{
                o.sec
              }}</RouterLink>
            </td>
            <td>{{ o.item || o.itemKey }}/{{ o.variant }}</td>
            <td>{{ o.description }}</td>
            <td>{{ o.qty }}</td>
            <td>{{ formatRange(o.from, o.to) }}</td>
          </tr>
        </tbody>
      </v-table>
    </template>
  </v-container>
</template>
