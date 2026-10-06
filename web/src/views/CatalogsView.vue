<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { Catalog } from '@/api/types'
import ErrorAlert from '@/components/ErrorAlert.vue'
import { useFetch } from '@/composables/useFetch'
import { useLinks } from '@/composables/useLinks'
import { formatRange } from '@/lib/format'

const { t } = useI18n()
const links = useLinks()
const { data, error, loading, reload } = useFetch<Catalog[]>(() => '/api/catalogs')
</script>

<template>
  <v-container>
    <h1 class="text-h6 mb-2">{{ t('catalogs.title') }}</h1>
    <v-progress-linear v-if="loading" indeterminate />
    <ErrorAlert :error="error" @retry="reload" />
    <v-list v-if="data" lines="two">
      <v-list-item
        v-for="c in data"
        :key="c.cat"
        :to="links.to(`/catalogs/${c.cat}`, {}, false)"
        :title="`${c.model} ${c.cmodel} · ${c.drive}`"
        :subtitle="[c.cat, formatRange(c.from, c.to), c.description].filter(Boolean).join(' · ')"
      />
    </v-list>
  </v-container>
</template>
