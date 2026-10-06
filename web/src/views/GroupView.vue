<script setup lang="ts">
import { computed, ref } from 'vue'
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
import { sectionMatchesCaption } from '@/lib/hotspots'
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

const caption = ref<string | null>(null)
const showAll = ref(false)
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
  <v-container fluid>
    <v-progress-linear v-if="loading" indeterminate />
    <ErrorAlert :error="error" @retry="reload" />
    <template v-if="data">
      <h1 class="text-h6 mb-2">{{ data.group.code }} — {{ data.group.label }}</h1>
      <v-row>
        <v-col v-if="data.group.image" cols="12" md="7">
          <DrawingViewer
            :src="data.group.image"
            :hotspots="data.hotspots"
            :selected="caption"
            :titles="titles"
            :style="{ height: mdAndUp ? 'calc(100vh - 160px)' : '60vh' }"
            @select="(k) => (caption = caption === k ? null : k)"
          />
        </v-col>
        <v-col cols="12" :md="data.group.image ? 5 : 12">
          <div class="d-flex align-center flex-wrap ga-2 mb-2">
            <v-chip v-if="caption" closable @click:close="caption = null">{{
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
              :title="`${s.sec} — ${s.name}`"
              :subtitle="[s.notes, formatRange(s.from, s.to)].filter((x) => x !== '').join(' · ')"
              :class="{ 'text-disabled': s.applicable === false }"
            >
              <template v-if="s.applicable === false" #append>
                <v-chip size="x-small" color="error" variant="tonal">{{
                  t('group.notApplicable')
                }}</v-chip>
              </template>
            </v-list-item>
          </v-list>
        </v-col>
      </v-row>
    </template>
  </v-container>
</template>
