<script setup lang="ts">
import { mdiCarInfo } from '@mdi/js'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useLinks } from '@/composables/useLinks'
import { useScope } from '@/composables/useScope'
import { scopeLabel } from '@/lib/scope'

const { t } = useI18n()
const links = useLinks()
const { current, saved, isSaved } = useScope()

const shown = computed(() => current.value ?? saved.value)
const label = computed(() => (shown.value === null ? t('scope.choose') : scopeLabel(shown.value)))
// With a vehicle: its sheet (where it can be kept, changed or removed); without: the vehicle search.
const target = computed(() =>
  shown.value === null ? links.to('/', {}, false) : links.to('/vehicle'),
)
</script>

<template>
  <v-chip
    :to="target"
    :prepend-icon="mdiCarInfo"
    :color="current && !isSaved ? 'warning' : 'primary'"
    :title="current && !isSaved ? t('scope.fromLink') : t('scope.sheet')"
    variant="tonal"
    class="mx-2 scope-chip"
  >
    {{ label }}
  </v-chip>
</template>
