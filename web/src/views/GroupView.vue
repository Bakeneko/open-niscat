<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { useDisplay } from 'vuetify'
import { apiPath } from '@/api/client'
import type { GroupDetail } from '@/api/types'
import DrawingViewer from '@/components/DrawingViewer.vue'
import ErrorAlert from '@/components/ErrorAlert.vue'
import { useFetch } from '@/composables/useFetch'
import { useLang } from '@/composables/useLang'
import { useLinks } from '@/composables/useLinks'
import { useScope } from '@/composables/useScope'
import { formatRange } from '@/lib/format'
import { groupHotspots, sectionMatchesCaption } from '@/lib/hotspots'
import { scopeToQuery } from '@/lib/scope'

const { t } = useI18n()
const route = useRoute()
const lang = useLang()
const links = useLinks()
const { mdAndUp } = useDisplay()
const { current } = useScope()
const cat = computed(() => String(route.params.cat))
const etd = computed(() => cat.value.slice(0, 2))
const { data, error, loading, reload } = useFetch<GroupDetail>(() =>
  apiPath(`/api/catalogs/${cat.value}/groups/${String(route.params.group)}`, {
    lang: lang.value,
    ...scopeToQuery(current.value),
  }),
)

// The viewer selects by hotspot key (caption without leading zeros); sections match on the caption itself.
const selected = ref<string | null>(null)
const caption = computed(
  () => data.value?.hotspots.find((h) => h.key === selected.value)?.caption ?? null,
)
watch(
  () => route.params.group,
  () => {
    selected.value = null
  },
)
const showAll = ref(false)
const spots = computed(() =>
  groupHotspots(data.value?.hotspots ?? [], data.value?.sections ?? [], showAll.value),
)
const sections = computed(() =>
  (data.value?.sections ?? []).filter(
    (s) =>
      (showAll.value || s.applicable !== false) &&
      (caption.value === null || sectionMatchesCaption(s.sec, caption.value)),
  ),
)
const hasHidden = computed(() => (data.value?.sections ?? []).some((s) => s.applicable === false))
const titles = computed(() => {
  const out: Record<string, string> = {}
  for (const h of data.value?.hotspots ?? []) {
    out[h.key] = (data.value?.sections ?? [])
      .filter((s) => sectionMatchesCaption(s.sec, h.caption))
      .map((s) => `${s.sec} ${s.name}`)
      .join('\n')
  }
  return out
})
</script>

<template>
  <v-container fluid :class="{ 'page-fill': mdAndUp && data?.group.image }">
    <v-progress-linear v-if="loading" indeterminate />
    <ErrorAlert :error="error" @retry="reload" />
    <template v-if="data">
      <h1 class="text-h6 mb-2">{{ data.group.code }} — {{ data.group.label }}</h1>
      <div :class="{ split: mdAndUp && data.group.image }">
        <div v-if="data.group.image">
          <DrawingViewer
            :src="data.group.image"
            :hotspots="spots.shown"
            :muted="spots.muted"
            labels
            :selected="selected"
            :titles="titles"
            :style="{ height: mdAndUp ? '100%' : '60vh' }"
            @select="(k) => (selected = selected === k ? null : k)"
          />
        </div>
        <div class="split-side">
          <div class="d-flex align-center flex-wrap ga-2 mb-2">
            <v-chip v-if="caption" closable @click:close="selected = null">{{
              t('group.filtered', { caption })
            }}</v-chip>
            <v-switch
              v-if="hasHidden"
              v-model="showAll"
              :label="t('group.showAll')"
              density="compact"
              hide-details
            />
          </div>
          <v-list density="compact" lines="two">
            <v-list-item
              v-for="s in sections"
              :key="s.sec"
              :to="links.to(`/section/${etd}/${s.sec}`)"
              :subtitle="[s.notes, formatRange(s.from, s.to)].filter((x) => x !== '').join(' · ')"
              :class="{ 'text-disabled': s.applicable === false }"
            >
              <template #title>
                <span :title="s.nameEn">{{ s.sec }} — {{ s.name }}</span>
              </template>
              <template v-if="s.applicable === false" #append>
                <v-chip size="x-small" color="error" variant="tonal">{{
                  t('group.notApplicable')
                }}</v-chip>
              </template>
            </v-list-item>
          </v-list>
        </div>
      </div>
    </template>
  </v-container>
</template>
