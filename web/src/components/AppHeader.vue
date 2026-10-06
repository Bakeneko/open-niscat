<script setup lang="ts">
import { mdiCart, mdiCheck, mdiMagnify, mdiTranslate } from '@mdi/js'
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useCart } from '@/composables/useCart'
import { useLang } from '@/composables/useLang'
import { useLinks } from '@/composables/useLinks'
import { switchLang, type Lang } from '@/lib/lang'
import { saveIfChanged } from '@/lib/storage'
import ScopeChip from './ScopeChip.vue'

const emit = defineEmits<{ cart: [] }>()
const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const lang = useLang()
const links = useLinks()
const cart = useCart()
const q = ref('')
const langs: readonly Lang[] = ['en', 'fr']

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
    <v-app-bar-title class="flex-grow-0">
      <RouterLink :to="links.to('/')" class="brand">{{ t('app.title') }}</RouterLink>
    </v-app-bar-title>
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
          :prepend-icon="mdiTranslate"
          variant="text"
          min-width="0"
          class="mx-1"
          :title="t('lang.label')"
          :aria-label="t('lang.label')"
        >
          {{ lang.toUpperCase() }}
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
          <template #prepend>
            <span class="lang-code">{{ l.toUpperCase() }}</span>
          </template>
          <v-list-item-title>{{ t(`lang.${l}`) }}</v-list-item-title>
          <template v-if="lang === l" #append>
            <v-icon :icon="mdiCheck" size="small" />
          </template>
        </v-list-item>
      </v-list>
    </v-menu>
    <v-btn icon :title="t('nav.cart')" :aria-label="t('nav.cart')" @click="emit('cart')">
      <v-badge :content="cart.count.value" :model-value="cart.count.value > 0" color="primary">
        <v-icon :icon="mdiCart" />
      </v-badge>
    </v-btn>
  </v-app-bar>
</template>
