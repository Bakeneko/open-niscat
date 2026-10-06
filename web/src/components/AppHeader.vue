<script setup lang="ts">
import { mdiCart, mdiMagnify } from '@mdi/js'
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useCart } from '@/composables/useCart'
import { useLang } from '@/composables/useLang'
import { useLinks } from '@/composables/useLinks'
import { isLang, switchLang, type Lang } from '@/lib/lang'
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

function setLang(v: unknown) {
  if (!isLang(v)) return
  const next: Lang = v
  saveIfChanged('open-niscat.lang', next)
  void router.push(switchLang(route.fullPath, next))
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
      density="compact"
      variant="solo-filled"
      flat
      hide-details
      single-line
      class="header-search mx-2"
      @keyup.enter="search"
    />
    <v-btn-toggle
      :model-value="lang"
      density="compact"
      mandatory
      variant="outlined"
      class="mx-1"
      @update:model-value="setLang"
    >
      <v-btn value="en">EN</v-btn>
      <v-btn value="fr">FR</v-btn>
    </v-btn-toggle>
    <v-btn icon :title="t('nav.cart')" @click="emit('cart')">
      <v-badge :content="cart.count.value" :model-value="cart.count.value > 0" color="primary">
        <v-icon :icon="mdiCart" />
      </v-badge>
    </v-btn>
  </v-app-bar>
</template>
