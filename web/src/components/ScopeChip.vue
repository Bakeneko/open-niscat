<script setup lang="ts">
import { mdiCarInfo, mdiCheck, mdiClose } from '@mdi/js'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useLinks } from '@/composables/useLinks'
import { useScope } from '@/composables/useScope'
import { scopeLabel } from '@/lib/scope'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const links = useLinks()
const { current, saved, isSaved, adopt, clear } = useScope()

const shown = computed(() => current.value ?? saved.value)
const label = computed(() => (shown.value === null ? t('scope.none') : scopeLabel(shown.value)))

function remove() {
  // A vehicle opened from someone else's link must not wipe the user's own saved vehicle.
  if (current.value === null || isSaved.value) clear()
  const query = { ...route.query }
  delete query.vin
  delete query.cat
  delete query.model
  void router.replace({ path: route.path, query })
}
</script>

<template>
  <v-menu>
    <template #activator="{ props: menu }">
      <v-chip
        v-bind="menu"
        :prepend-icon="mdiCarInfo"
        :color="current && !isSaved ? 'warning' : 'primary'"
        variant="tonal"
        class="mx-2 scope-chip"
      >
        {{ label }}
      </v-chip>
    </template>
    <v-list density="compact">
      <v-list-item
        v-if="current && !isSaved"
        :prepend-icon="mdiCheck"
        :title="t('scope.use')"
        :subtitle="t('scope.fromLink')"
        @click="adopt()"
      />
      <v-list-item v-if="shown" :to="links.to('/vehicle')" :title="t('scope.sheet')" />
      <v-list-item :to="links.to('/', {}, false)" :title="t('scope.choose')" />
      <v-list-item
        v-if="shown"
        :prepend-icon="mdiClose"
        :title="t('scope.clear')"
        @click="remove"
      />
    </v-list>
  </v-menu>
</template>
