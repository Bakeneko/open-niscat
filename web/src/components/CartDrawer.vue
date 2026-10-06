<script setup lang="ts">
import { mdiDelete } from '@mdi/js'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { apiPath } from '@/api/client'
import type { LinesResult } from '@/api/types'
import { useCart } from '@/composables/useCart'
import { useFetch } from '@/composables/useFetch'
import { useLang } from '@/composables/useLang'
import { useLinks } from '@/composables/useLinks'

const open = defineModel<boolean>({ required: true })
const { t } = useI18n()
const lang = useLang()
const links = useLinks()
const cart = useCart()

const url = computed(() =>
  open.value && cart.items.value.length > 0
    ? apiPath('/api/lines', { ids: cart.items.value.map((i) => i.id).join(','), lang: lang.value })
    : null,
)
const { data } = useFetch<LinesResult>(() => url.value)
const byId = computed(() => new Map((data.value?.lines ?? []).map((l) => [l.id, l])))
</script>

<template>
  <v-navigation-drawer v-model="open" location="right" temporary width="360" class="no-print">
    <v-list density="compact">
      <v-list-subheader>{{ t('cart.title') }}</v-list-subheader>
      <v-list-item v-if="cart.items.value.length === 0" :title="t('cart.empty')" />
      <v-list-item
        v-for="item in cart.items.value"
        :key="item.id"
        :title="byId.get(item.id)?.description ?? item.id"
        :subtitle="`${byId.get(item.id)?.partNo ?? ''} × ${String(item.qty)}`"
      >
        <template #append>
          <v-btn
            :icon="mdiDelete"
            size="small"
            variant="text"
            :title="t('cart.remove')"
            @click="cart.remove(item.id)"
          />
        </template>
      </v-list-item>
    </v-list>
    <template #append>
      <div class="pa-2">
        <v-btn block color="primary" :to="links.to('/cart', {}, false)" @click="open = false">{{
          t('cart.open')
        }}</v-btn>
      </div>
    </template>
  </v-navigation-drawer>
</template>
