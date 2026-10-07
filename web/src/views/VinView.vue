<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { ApiError, apiPath } from '@/api/client'
import type { VinResult } from '@/api/types'
import ErrorAlert from '@/components/ErrorAlert.vue'
import VehicleCard from '@/components/VehicleCard.vue'
import { useFetch } from '@/composables/useFetch'
import { useLang } from '@/composables/useLang'
import { useLinks } from '@/composables/useLinks'
import { useScope } from '@/composables/useScope'
import { usePageTitle } from '@/composables/usePageTitle'
import { formatYearMonth } from '@/lib/format'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const lang = useLang()
const links = useLinks()
const { adopt } = useScope()
const vin = computed(() => String(route.params.vin))
const { data, error, loading, reload } = useFetch<VinResult>(() =>
  apiPath(`/api/vin/${encodeURIComponent(vin.value)}`, { lang: lang.value }),
)

const notFound = computed(() => error.value instanceof ApiError && error.value.code === 'not_found')

function use() {
  const v = data.value?.vehicle?.vin
  if (v === undefined) return
  adopt({ vin: v })
  void router.push(links.to('/vehicle', { vin: v }, false))
}

usePageTitle(() => vin.value)
</script>

<template>
  <v-container style="max-width: 760px">
    <v-progress-linear v-if="loading" indeterminate />
    <v-alert v-if="notFound" type="warning" variant="tonal">
      {{ t('vin.notFound') }} <strong>{{ vin }}</strong>
      <div class="mt-2">
        <RouterLink :to="links.to('/', {}, false)">{{ t('scope.choose') }}</RouterLink>
      </div>
    </v-alert>
    <ErrorAlert v-else :error="error" @retry="reload" />
    <template v-if="data?.vehicle">
      <VehicleCard :vehicle="data.vehicle" />
      <v-btn color="primary" class="mt-3" @click="use">{{ t('scope.use') }}</v-btn>
    </template>
    <v-card v-else-if="data?.candidates">
      <v-card-title>{{ t('vin.candidates') }}</v-card-title>
      <v-list>
        <v-list-item
          v-for="c in data.candidates"
          :key="c.vin"
          :to="links.to(`/vin/${encodeURIComponent(c.vin)}`, {}, false)"
          :title="c.vin"
          :subtitle="`${c.cat} · ${c.model} · ${formatYearMonth(c.prodDate, true)}`"
        />
      </v-list>
    </v-card>
  </v-container>
</template>
