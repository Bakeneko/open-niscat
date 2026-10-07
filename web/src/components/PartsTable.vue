<script setup lang="ts">
import { mdiPlaylistPlus } from '@mdi/js'
import { useI18n } from 'vue-i18n'
import type { Line } from '@/api/types'
import { useLinks } from '@/composables/useLinks'
import { formatRange, readable } from '@/lib/format'
import { refPath } from '@/lib/refs'

const props = defineProps<{ lines: Line[]; selected: string | null }>()
const emit = defineEmits<{ select: [key: string]; add: [line: Line] }>()
const { t } = useI18n()
const links = useLinks()
const extraFields = ['ica', 'app', 'spec', 'pnc', 'kd'] as const
const isSelected = (l: Line) => props.selected !== null && l.itemKey === props.selected
const hasDetails = (l: Line) =>
  l.alternative !== '' ||
  l.latest !== undefined ||
  l.inPeriod === false ||
  extraFields.some((f) => l[f] !== '')
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
      <template v-for="l in lines" :key="l.id">
        <tr
          :data-item="l.itemKey"
          :class="{ selected: isSelected(l) }"
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
              size="small"
              class="text-no-wrap px-2"
              :color="l.inPeriod === false ? 'warning' : undefined"
              variant="tonal"
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
              :icon="mdiPlaylistPlus"
              size="small"
              variant="text"
              :title="t('cart.add')"
              @click.stop="emit('add', l)"
            />
          </td>
        </tr>
        <!-- Details of the selected item open in place, under each of its lines. -->
        <tr
          v-if="isSelected(l) && hasDetails(l)"
          class="detail-row"
          @click="emit('select', l.itemKey)"
        >
          <td />
          <td colspan="3">
            <dl class="details">
              <template v-if="l.alternative">
                <dt>{{ t('part.alternative') }}</dt>
                <dd>
                  <RouterLink :to="links.to(refPath(l.alternative), {}, false)" @click.stop>{{
                    l.alternative
                  }}</RouterLink>
                </dd>
              </template>
              <template v-if="l.latest">
                <dt>{{ t('part.latest') }}</dt>
                <dd>
                  <RouterLink :to="links.to(refPath(l.latest.partNo), {}, false)" @click.stop>{{
                    l.latest.partNo
                  }}</RouterLink>
                </dd>
              </template>
              <template v-if="l.inPeriod === false">
                <dt>{{ t('part.period') }}</dt>
                <dd class="text-warning font-weight-bold">{{ t('part.outOfPeriod') }}</dd>
              </template>
              <template v-for="f in extraFields" :key="f">
                <template v-if="l[f]">
                  <dt>{{ t(`part.${f}`) }}</dt>
                  <dd>{{ l[f] }}</dd>
                </template>
              </template>
            </dl>
          </td>
          <td class="no-print" />
        </tr>
      </template>
    </tbody>
  </v-table>
</template>

<style scoped>
/* Scrolling a selected item to the top of the panel leaves a little context above it. */
.parts-table tbody tr {
  scroll-margin-top: 48px;
}
.detail-row td {
  background: rgba(var(--v-theme-primary), 0.05);
}
.details {
  display: grid;
  grid-template-columns: max-content 1fr;
  gap: 2px 12px;
  padding: 4px 0 6px;
  font-size: 0.85rem;
}
.details dt {
  color: rgba(var(--v-theme-on-surface), var(--v-medium-emphasis-opacity));
}
.details dd {
  margin: 0;
}
/* Tight cells: the table lives in a narrow side panel or a phone screen. */
.parts-table :deep(th),
.parts-table :deep(td) {
  padding: 0 5px;
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
