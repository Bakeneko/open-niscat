<script setup lang="ts">
import { useId } from 'vue'
import type { Lang } from '@/lib/lang'

defineProps<{ lang: Lang }>()
// Clip-path ids must be unique per instance (the flag appears in the header and in the menu).
const id = useId()
</script>

<template>
  <svg
    v-if="lang === 'fr'"
    class="flag"
    viewBox="0 0 3 2"
    preserveAspectRatio="xMidYMid slice"
    aria-hidden="true"
  >
    <rect width="1" height="2" fill="#002654" />
    <rect x="1" width="1" height="2" fill="#fff" />
    <rect x="2" width="1" height="2" fill="#ce1126" />
  </svg>
  <!-- Spain without its coat of arms, unreadable at this size. -->
  <svg
    v-else-if="lang === 'es'"
    class="flag"
    viewBox="0 0 3 2"
    preserveAspectRatio="xMidYMid slice"
    aria-hidden="true"
  >
    <rect width="3" height="2" fill="#aa151b" />
    <rect y="0.5" width="3" height="1" fill="#f1bf00" />
  </svg>
  <svg
    v-else-if="lang === 'de'"
    class="flag"
    viewBox="0 0 5 3"
    preserveAspectRatio="xMidYMid slice"
    aria-hidden="true"
  >
    <rect width="5" height="1" fill="#000" />
    <rect y="1" width="5" height="1" fill="#dd0000" />
    <rect y="2" width="5" height="1" fill="#ffce00" />
  </svg>
  <svg
    v-else
    class="flag"
    viewBox="0 0 60 30"
    preserveAspectRatio="xMidYMid slice"
    aria-hidden="true"
  >
    <clipPath :id="`${id}-s`"><path d="M0,0 v30 h60 v-30 z" /></clipPath>
    <clipPath :id="`${id}-t`">
      <path d="M30,15 h30 v15 z v15 h-30 z h-30 v-15 z v-15 h30 z" />
    </clipPath>
    <g :clip-path="`url(#${id}-s)`">
      <path d="M0,0 v30 h60 v-30 z" fill="#012169" />
      <path d="M0,0 L60,30 M60,0 L0,30" stroke="#fff" stroke-width="6" />
      <path
        d="M0,0 L60,30 M60,0 L0,30"
        :clip-path="`url(#${id}-t)`"
        stroke="#c8102e"
        stroke-width="4"
      />
      <path d="M30,0 v30 M0,15 h60" stroke="#fff" stroke-width="10" />
      <path d="M30,0 v30 M0,15 h60" stroke="#c8102e" stroke-width="6" />
    </g>
  </svg>
</template>

<style scoped>
.flag {
  width: 1.4em;
  height: 0.95em;
  border-radius: 2px;
  box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.15);
  vertical-align: middle;
}
</style>
