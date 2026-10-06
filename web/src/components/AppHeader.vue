<script setup lang="ts">
import { mdiCart, mdiMagnify } from '@mdi/js'
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
    <div class="mx-1 d-flex">
      <v-btn
        v-for="l in langs"
        :key="l"
        :to="switchLang(route.fullPath, l)"
        :active="lang === l"
        :variant="lang === l ? 'tonal' : 'text'"
        density="compact"
        min-width="0"
        @click="rememberLang(l)"
      >
        {{ l.toUpperCase() }}
      </v-btn>
    </div>
    <v-btn icon :title="t('nav.cart')" :aria-label="t('nav.cart')" @click="emit('cart')">
      <v-badge :content="cart.count.value" :model-value="cart.count.value > 0" color="primary">
        <v-icon :icon="mdiCart" />
      </v-badge>
    </v-btn>
  </v-app-bar>
</template>
