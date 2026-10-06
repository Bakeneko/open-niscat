<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { apiPath } from '@/api/client'
import type { ModelInfo } from '@/api/types'
import ErrorAlert from '@/components/ErrorAlert.vue'
import { useFetch } from '@/composables/useFetch'
import { useLang } from '@/composables/useLang'
import { useLinks } from '@/composables/useLinks'
import { useScope } from '@/composables/useScope'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const lang = useLang()
const links = useLinks()
const { adopt } = useScope()
const cat = computed(() => String(route.params.cat))
const { data, error, loading, reload } = useFetch<ModelInfo[]>(() =>
  apiPath(`/api/catalogs/${cat.value}/models`, { lang: lang.value }),
)

/** One filter per attribute table; value '' = any. */
const filters = ref<Record<string, string>>({})
const tables = computed(() =>
  (data.value?.[0]?.attributes ?? []).map((a) => ({ table: a.table, name: a.name })),
)
function choices(table: string): string[] {
  const values = new Set<string>()
  for (const m of data.value ?? []) {
    const a = m.attributes.find((x) => x.table === table)
    if (a !== undefined) values.add(a.value)
  }
  return [...values].toSorted()
}
const shown = computed(() =>
  (data.value ?? []).filter((m) =>
    Object.entries(filters.value).every(
      ([table, v]) => v === '' || m.attributes.find((a) => a.table === table)?.value === v,
    ),
  ),
)

function use(model: string) {
  adopt({ cat: cat.value, model })
  void router.push(links.to('/vehicle', { cat: cat.value, model }, false))
}
</script>

<template>
  <v-container fluid>
    <h1 class="text-h6 mb-2">{{ t('models.title') }} · {{ cat }}</h1>
    <v-progress-linear v-if="loading" indeterminate />
    <ErrorAlert :error="error" @retry="reload" />
    <v-row v-if="data" dense>
      <v-col v-for="tb in tables" :key="tb.table" cols="6" md="3" lg="2">
        <v-select
          :model-value="filters[tb.table] ?? ''"
          :items="['', ...choices(tb.table)]"
          :label="tb.name"
          density="compact"
          hide-details
          @update:model-value="(v: string) => (filters = { ...filters, [tb.table]: v })"
        />
      </v-col>
    </v-row>
    <div v-if="data" class="text-caption my-2">{{ t('models.count', { n: shown.length }) }}</div>
    <v-list v-if="data" lines="two" density="compact">
      <v-list-item
        v-for="m in shown"
        :key="m.model"
        :title="m.model"
        :subtitle="m.attributes.map((a) => a.value).join(' · ')"
      >
        <template #append>
          <span class="text-caption mr-2">{{ m.vinCount }} {{ t('models.vins') }}</span>
          <v-btn size="small" color="primary" variant="tonal" @click="use(m.model)">{{
            t('models.use')
          }}</v-btn>
        </template>
      </v-list-item>
    </v-list>
  </v-container>
</template>
