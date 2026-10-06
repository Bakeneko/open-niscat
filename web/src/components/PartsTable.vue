<script setup lang="ts">
import { mdiCartPlus } from '@mdi/js'
import { useI18n } from 'vue-i18n'
import type { Line } from '@/api/types'
import { useLinks } from '@/composables/useLinks'
import { formatRange, readable } from '@/lib/format'
import { refPath } from '@/lib/refs'

defineProps<{ lines: Line[]; selected: string | null }>()
const emit = defineEmits<{ select: [key: string]; add: [line: Line] }>()
const { t } = useI18n()
const links = useLinks()
const indent = (l: Line) => ({ paddingLeft: `${String(Math.max(l.level - 1, 0) * 12 + 6)}px` })
</script>

<template>
  <v-table density="compact" hover class="parts-table">
    <thead>
      <tr>
        <th>{{ t('part.item') }}</th>
        <th>{{ t('part.part') }}</th>
        <th>{{ t('part.period') }}</th>
        <th class="text-right">{{ t('part.qty') }}</th>
        <th class="no-print" />
      </tr>
    </thead>
    <tbody>
      <tr
        v-for="l in lines"
        :key="l.id"
        :data-item="l.itemKey"
        :class="{ selected: selected !== null && l.itemKey === selected }"
        @click="emit('select', l.itemKey)"
      >
        <td class="text-no-wrap">
          <span class="text-medium-emphasis">{{ l.mark }}</span> {{ l.item
          }}<span v-if="l.variant" class="text-caption">/{{ l.variant }}</span>
        </td>
        <!-- One "Part" cell (number, then description) keeps the table within narrow panels. -->
        <td :style="indent(l)" class="part-cell">
          <RouterLink
            :to="links.to(refPath(l.partNo), {}, false)"
            class="font-weight-bold text-no-wrap"
            @click.stop
            >{{ l.partNo }}</RouterLink
          >
          <div>{{ readable(l.description) }}</div>
        </td>
        <td>
          <v-chip
            v-if="l.from || l.to"
            size="x-small"
            class="text-no-wrap"
            :color="l.inPeriod === false ? 'warning' : undefined"
            :variant="l.inPeriod === false ? 'flat' : 'tonal'"
            :title="l.inPeriod === false ? t('part.outOfPeriod') : undefined"
          >
            {{ formatRange(l.from, l.to) }}
          </v-chip>
        </td>
        <td class="text-right text-no-wrap">
          {{ l.qty }}<span v-if="l.cap" class="text-caption"> ({{ l.cap }})</span>
        </td>
        <td class="no-print">
          <v-btn
            :icon="mdiCartPlus"
            size="small"
            variant="text"
            :title="t('cart.add')"
            @click.stop="emit('add', l)"
          />
        </td>
      </tr>
    </tbody>
  </v-table>
</template>

<style scoped>
/* Tight cells: the table lives in a narrow side panel or a phone screen. */
.parts-table :deep(th),
.parts-table :deep(td) {
  padding: 0 6px;
}
.parts-table :deep(td) {
  overflow-wrap: anywhere;
}
.part-cell {
  padding-top: 4px !important;
  padding-bottom: 4px !important;
}
.parts-table tr.selected td {
  background: rgba(var(--v-theme-primary), 0.12);
}
.parts-table tbody tr {
  cursor: pointer;
}
</style>
