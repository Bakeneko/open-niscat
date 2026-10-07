<script setup lang="ts">
import { computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { mdiCarArrowRight, mdiCheck, mdiClose } from '@mdi/js'
import { useRoute, useRouter } from 'vue-router'
import { apiPath } from '@/api/client'
import type { Group, Vehicle } from '@/api/types'
import ErrorAlert from '@/components/ErrorAlert.vue'
import VehicleCard from '@/components/VehicleCard.vue'
import { useFetch } from '@/composables/useFetch'
import { useHistory } from '@/composables/useHistory'
import { useLang } from '@/composables/useLang'
import { useLinks } from '@/composables/useLinks'
import { useScope } from '@/composables/useScope'
import { usePageTitle } from '@/composables/usePageTitle'
import { scopeLabel, scopeToQuery } from '@/lib/scope'

const { t } = useI18n()
const route = useRoute()
const lang = useLang()
const links = useLinks()
const history = useHistory()
const router = useRouter()
const { current, isSaved, adopt, clear } = useScope()

// A vehicle opened from someone else's link must not wipe the user's own saved vehicle.
function remove() {
  if (isSaved.value) clear()
  void router.push(links.to('/', {}, false))
}

const vehicle = useFetch<Vehicle>(() =>
  current.value === null
    ? null
    : apiPath('/api/vehicle', { lang: lang.value, ...scopeToQuery(current.value) }),
)
const cat = computed(() => vehicle.data.value?.catalog.cat ?? null)
const groups = useFetch<Group[]>(() =>
  cat.value === null ? null : apiPath(`/api/catalogs/${cat.value}/groups`, { lang: lang.value }),
)

watch(vehicle.data, (v) => {
  if (v !== null && current.value !== null)
    history.push(
      'vehicle',
      `${scopeLabel(current.value)} · ${v.catalog.model} ${v.catalog.cmodel}`,
      route.fullPath,
    )
})

usePageTitle(() => {
  if (vehicle.loading.value) return undefined // data may still be the previous vehicle's
  const v = vehicle.data.value
  if (v === null || current.value === null) return null
  return `${scopeLabel(current.value)} · ${v.catalog.model} ${v.catalog.cmodel}`
})
</script>

<template>
  <v-container style="max-width: 1400px">
    <v-alert v-if="current === null" type="info" variant="tonal">
      <RouterLink :to="links.to('/', {}, false)">{{ t('scope.choose') }}</RouterLink>
    </v-alert>
    <v-progress-linear v-if="vehicle.loading.value || groups.loading.value" indeterminate />
    <ErrorAlert :error="vehicle.error.value" @retry="vehicle.reload" />
    <ErrorAlert :error="groups.error.value" @retry="groups.reload" />
    <div v-if="vehicle.data.value" class="d-flex flex-wrap ga-2 mb-3 no-print">
      <v-btn
        v-if="current && !isSaved"
        :prepend-icon="mdiCheck"
        color="primary"
        :title="t('scope.fromLink')"
        @click="adopt()"
        >{{ t('scope.use') }}</v-btn
      >
      <v-btn :prepend-icon="mdiCarArrowRight" variant="tonal" :to="links.to('/', {}, false)">{{
        t('scope.change')
      }}</v-btn>
      <v-btn :prepend-icon="mdiClose" variant="text" @click="remove">{{ t('scope.clear') }}</v-btn>
    </div>
    <VehicleCard v-if="vehicle.data.value" :vehicle="vehicle.data.value" class="mb-4" />
    <h2 v-if="groups.data.value" class="text-subtitle-1 mb-2">{{ t('vehicle.groups') }}</h2>
    <v-row v-if="groups.data.value && cat" dense>
      <v-col v-for="g in groups.data.value" :key="g.code" cols="12" sm="6" md="4">
        <v-card
          :to="links.to(`/index/${cat}/${g.code}`)"
          :title="g.label"
          :subtitle="g.code"
          variant="tonal"
        />
      </v-col>
    </v-row>
  </v-container>
</template>
