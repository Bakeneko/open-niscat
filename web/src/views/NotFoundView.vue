<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useLinks } from '@/composables/useLinks'

const { t } = useI18n()
const links = useLinks()
</script>

<template>
  <v-container>
    <v-alert type="warning" variant="tonal">
      {{ t('app.notFound') }} —
      <RouterLink :to="links.to('/', {}, false)">{{ t('nav.home') }}</RouterLink>
    </v-alert>
  </v-container>
</template>
