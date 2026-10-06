<script setup lang="ts">
import { mdiCartPlus } from '@mdi/js'
import { useI18n } from 'vue-i18n'
import type { Line } from '@/api/types'
import { useLinks } from '@/composables/useLinks'
import { formatRange } from '@/lib/format'
import { refPath } from '@/lib/refs'

defineProps<{ lines: Line[] }>()
const emit = defineEmits<{ add: [line: Line] }>()
const { t } = useI18n()
const links = useLinks()
</script>

<template>
  <div>
    <v-alert v-if="lines.length === 0" type="info" variant="tonal" density="compact">{{
      t('section.select')
    }}</v-alert>
    <v-card v-for="l in lines" :key="l.id" class="mb-2" variant="outlined">
      <v-card-title class="text-subtitle-1">
        {{ l.item || '·' }}/{{ l.variant }} — {{ l.description }}
      </v-card-title>
      <v-card-text>
        <v-table density="compact">
          <tbody>
            <tr>
              <th>{{ t('part.reference') }}</th>
              <td>
                <RouterLink :to="links.to(refPath(l.partNo), {}, false)">{{ l.partNo }}</RouterLink>
              </td>
            </tr>
            <tr v-if="l.alternative">
              <th>{{ t('part.alternative') }}</th>
              <td>
                <RouterLink :to="links.to(refPath(l.alternative), {}, false)">{{
                  l.alternative
                }}</RouterLink>
              </td>
            </tr>
            <tr v-if="l.latest">
              <th>{{ t('part.latest') }}</th>
              <td>
                <RouterLink :to="links.to(refPath(l.latest.partNo), {}, false)">{{
                  l.latest.partNo
                }}</RouterLink>
              </td>
            </tr>
            <tr>
              <th>{{ t('part.qty') }}</th>
              <td>
                {{ l.qty }}<span v-if="l.cap"> ({{ l.cap }})</span>
              </td>
            </tr>
            <tr v-if="l.from || l.to">
              <th>{{ t('part.period') }}</th>
              <td :class="{ 'text-warning font-weight-bold': l.inPeriod === false }">
                {{ formatRange(l.from, l.to)
                }}<span v-if="l.inPeriod === false"> — {{ t('part.outOfPeriod') }}</span>
              </td>
            </tr>
            <tr v-if="l.ica">
              <th>{{ t('part.ica') }}</th>
              <td>{{ l.ica }}</td>
            </tr>
            <tr v-if="l.app">
              <th>{{ t('part.app') }}</th>
              <td>{{ l.app }}</td>
            </tr>
            <tr v-if="l.spec">
              <th>{{ t('part.spec') }}</th>
              <td>{{ l.spec }}</td>
            </tr>
            <tr v-if="l.pnc">
              <th>{{ t('part.pnc') }}</th>
              <td>{{ l.pnc }}</td>
            </tr>
            <tr v-if="l.kd">
              <th>{{ t('part.kd') }}</th>
              <td>{{ l.kd }}</td>
            </tr>
            <tr v-if="l.mark">
              <th>{{ t('part.mark') }}</th>
              <td>{{ l.mark }}</td>
            </tr>
          </tbody>
        </v-table>
      </v-card-text>
      <v-card-actions class="no-print">
        <v-btn
          :prepend-icon="mdiCartPlus"
          color="primary"
          variant="tonal"
          @click="emit('add', l)"
          >{{ t('cart.add') }}</v-btn
        >
      </v-card-actions>
    </v-card>
  </div>
</template>
