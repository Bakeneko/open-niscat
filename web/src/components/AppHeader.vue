<script setup lang="ts">
import { mdiBookOpenVariant, mdiHome, mdiMagnify, mdiPlaylistEdit } from '@mdi/js'
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { usePartsList } from '@/composables/usePartsList'
import { useLang } from '@/composables/useLang'
import { useLinks } from '@/composables/useLinks'
import { LANGS, switchLang, type Lang } from '@/lib/lang'
import { saveIfChanged } from '@/lib/storage'
import FlagIcon from './FlagIcon.vue'
import ScopeChip from './ScopeChip.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const lang = useLang()
const links = useLinks()
const partsList = usePartsList()
const q = ref('')
const langs = LANGS

watch(
  () => route.query.q,
  (v) => {
    q.value = typeof v === 'string' ? v : ''
  },
  { immediate: true },
)

function rememberLang(l: Lang) {
  saveIfChanged('open-niscat.lang', l)
}

function search() {
  const text = q.value.trim()
  if (text !== '') void router.push(links.to('/search', { q: text }))
}
</script>

<template>
  <v-app-bar density="comfortable" class="no-print">
    <!-- Labels show from medium screens up; phones keep the icons only. -->
    <v-btn
      :to="links.to('/', {}, false)"
      variant="text"
      min-width="0"
      class="ms-1 px-2"
      :title="t('nav.home')"
      :aria-label="t('nav.home')"
    >
      <v-icon :icon="mdiHome" />
      <span class="d-none d-md-inline ms-2 font-weight-bold">{{ t('app.title') }}</span>
    </v-btn>
    <v-btn
      :to="links.to('/catalogs', {}, false)"
      variant="text"
      min-width="0"
      class="px-2"
      :title="t('nav.catalogs')"
      :aria-label="t('nav.catalogs')"
    >
      <v-icon :icon="mdiBookOpenVariant" />
      <span class="d-none d-md-inline ms-2">{{ t('nav.catalogs') }}</span>
    </v-btn>
    <ScopeChip />
    <v-text-field
      v-model="q"
      :prepend-inner-icon="mdiMagnify"
      :placeholder="t('search.placeholder')"
      :aria-label="t('home.searchLabel')"
      density="compact"
      variant="solo-filled"
      flat
      hide-details
      single-line
      class="header-search mx-2"
      @keyup.enter="search"
    />
    <v-btn
      icon
      class="header-search-btn"
      :to="links.to('/search')"
      :title="t('home.searchLabel')"
      :aria-label="t('home.searchLabel')"
    >
      <v-icon :icon="mdiMagnify" />
    </v-btn>
    <v-menu>
      <template #activator="{ props: menu }">
        <v-btn
          v-bind="menu"
          variant="text"
          min-width="0"
          class="mx-1"
          :title="t('lang.label')"
          :aria-label="t('lang.label')"
        >
          <FlagIcon :lang="lang" class="mr-2" />{{ lang.toUpperCase() }}
        </v-btn>
      </template>
      <v-list density="compact">
        <v-list-item
          v-for="l in langs"
          :key="l"
          :to="switchLang(route.fullPath, l)"
          :active="lang === l"
          @click="rememberLang(l)"
        >
          <v-list-item-title class="d-flex align-center">
            <FlagIcon :lang="l" class="mr-2" />{{ l.toUpperCase() }}
          </v-list-item-title>
        </v-list-item>
      </v-list>
    </v-menu>
    <v-btn
      icon
      :to="links.to('/list', {}, false)"
      :title="t('nav.list')"
      :aria-label="t('nav.list')"
    >
      <v-badge
        :content="partsList.count.value"
        :model-value="partsList.count.value > 0"
        color="primary"
      >
        <v-icon :icon="mdiPlaylistEdit" />
      </v-badge>
    </v-btn>
  </v-app-bar>
</template>
