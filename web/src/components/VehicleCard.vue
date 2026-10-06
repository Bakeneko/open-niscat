<script setup lang="ts">
import { mdiFilePdfBox } from '@mdi/js'
import { useI18n } from 'vue-i18n'
import type { Vehicle } from '@/api/types'
import { formatRange, formatYearMonth } from '@/lib/format'

defineProps<{ vehicle: Vehicle }>()
const { t } = useI18n()
const fileName = (url: string) => url.split('/').pop() ?? url
</script>

<template>
  <v-card>
    <v-card-title
      >{{ vehicle.catalog.model }} {{ vehicle.catalog.cmodel }}
      {{ vehicle.catalog.drive }}</v-card-title
    >
    <v-card-subtitle>
      {{ vehicle.catalog.cat }} · {{ formatRange(vehicle.catalog.from, vehicle.catalog.to) }} ·
      {{ vehicle.catalog.description }}
    </v-card-subtitle>
    <v-card-text>
      <v-table density="compact">
        <tbody>
          <tr v-if="vehicle.vin">
            <th>{{ t('vehicle.vin') }}</th>
            <td>{{ vehicle.vin }}</td>
          </tr>
          <tr v-if="vehicle.model">
            <th>{{ t('vehicle.model') }}</th>
            <td>{{ vehicle.model }}</td>
          </tr>
          <tr v-if="vehicle.prodDate">
            <th>{{ t('vehicle.prodDate') }}</th>
            <td>{{ formatYearMonth(vehicle.prodDate, true) }}</td>
          </tr>
          <tr v-for="a in vehicle.attributes" :key="a.table">
            <th>{{ a.name }}</th>
            <td>{{ a.value }}</td>
          </tr>
          <tr v-if="vehicle.vinCount">
            <th>{{ t('vehicle.vinCount') }}</th>
            <td>{{ vehicle.vinCount }}</td>
          </tr>
        </tbody>
      </v-table>
      <div v-if="vehicle.documents.length > 0" class="mt-3">
        <div class="text-subtitle-2">{{ t('vehicle.documents') }}</div>
        <v-chip
          v-for="d in vehicle.documents"
          :key="d"
          :href="d"
          target="_blank"
          :prepend-icon="mdiFilePdfBox"
          size="small"
          class="ma-1"
        >
          {{ fileName(d) }}
        </v-chip>
      </div>
    </v-card-text>
  </v-card>
</template>
