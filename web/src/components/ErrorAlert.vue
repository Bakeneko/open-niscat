<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ApiError } from '@/api/client'

const props = defineProps<{ error: Error | null }>()
const emit = defineEmits<{ retry: [] }>()
const { t } = useI18n()

const text = computed(() => {
  const e = props.error
  if (e === null) return ''
  if (e instanceof ApiError) {
    const known = ['not_found', 'invalid', 'internal'] as const
    const code = known.find((k) => k === e.code)
    return code === undefined ? e.message : `${t(`error.${code}`)} ${e.message}`
  }
  return t('error.network')
})
</script>

<template>
  <v-alert v-if="error" type="error" variant="tonal" class="ma-2">
    {{ text }}
    <template #append>
      <v-btn variant="text" @click="emit('retry')">{{ t('app.retry') }}</v-btn>
    </template>
  </v-alert>
</template>
