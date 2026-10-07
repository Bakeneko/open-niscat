<script setup lang="ts">
import { watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppHeader from '@/components/AppHeader.vue'
import { useLang } from '@/composables/useLang'
import { useNotify } from '@/composables/useNotify'

const { locale } = useI18n()
const lang = useLang()
const { message, visible } = useNotify()

watch(
  lang,
  (l) => {
    locale.value = l
    document.documentElement.lang = l
  },
  { immediate: true },
)
</script>

<template>
  <v-app>
    <AppHeader />
    <v-main>
      <RouterView />
    </v-main>
    <v-snackbar v-model="visible" :timeout="2500" class="no-print">{{ message }}</v-snackbar>
  </v-app>
</template>
