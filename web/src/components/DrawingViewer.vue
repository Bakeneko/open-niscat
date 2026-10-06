<script setup lang="ts">
import { mdiFitToScreen, mdiMagnifyMinus, mdiMagnifyPlus } from '@mdi/js'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Hotspot } from '@/api/types'
import { distance, fitView, panBy, zoomAt, type View } from '@/lib/viewport'

const props = defineProps<{
  src: string
  hotspots: Hotspot[]
  selected: string | null
  titles?: Record<string, string>
}>()
const emit = defineEmits<{ select: [key: string] }>()
const { t } = useI18n()

const box = ref<HTMLDivElement | null>(null)
const view = ref<View>({ scale: 1, x: 0, y: 0 })
const natural = ref({ w: 0, h: 0 })
const fitted = ref(1)
const DRAG_THRESHOLD = 6

const pointers = new Map<number, { x: number; y: number }>()
let travelled = 0
let pinch = 0

const stageStyle = computed(() => ({
  transform: `translate(${String(view.value.x)}px, ${String(view.value.y)}px) scale(${String(view.value.scale)})`,
  width: `${String(natural.value.w)}px`,
  height: `${String(natural.value.h)}px`,
}))
const spotStyle = (h: Hotspot) => ({
  left: `${String(h.x)}px`,
  top: `${String(h.y)}px`,
  width: `${String(h.w)}px`,
  height: `${String(h.h)}px`,
})

function fit() {
  const el = box.value
  if (el === null) return
  view.value = fitView(el.clientWidth, el.clientHeight, natural.value.w, natural.value.h)
  fitted.value = view.value.scale
}
function onLoad(e: Event) {
  const img = e.target as HTMLImageElement
  natural.value = { w: img.naturalWidth, h: img.naturalHeight }
  fit()
}
function local(e: { clientX: number; clientY: number }) {
  const r = box.value?.getBoundingClientRect()
  return { x: e.clientX - (r?.left ?? 0), y: e.clientY - (r?.top ?? 0) }
}
function zoom(factor: number, at?: { x: number; y: number }) {
  const el = box.value
  if (el === null) return
  const p = at ?? { x: el.clientWidth / 2, y: el.clientHeight / 2 }
  view.value = zoomAt(
    view.value,
    factor,
    p.x,
    p.y,
    fitted.value * 0.5,
    Math.max(4, fitted.value * 8),
  )
}
function onWheel(e: WheelEvent) {
  e.preventDefault()
  zoom(e.deltaY < 0 ? 1.2 : 1 / 1.2, local(e))
}
function onDown(e: PointerEvent) {
  box.value?.setPointerCapture(e.pointerId)
  pointers.set(e.pointerId, local(e))
  if (pointers.size === 1) travelled = 0
  if (pointers.size === 2) {
    const [a, b] = [...pointers.values()]
    if (a !== undefined && b !== undefined) pinch = distance(a, b)
  }
}
function onMove(e: PointerEvent) {
  const prev = pointers.get(e.pointerId)
  if (prev === undefined) return
  const p = local(e)
  pointers.set(e.pointerId, p)
  if (pointers.size === 1) {
    travelled += distance(prev, p)
    view.value = panBy(view.value, p.x - prev.x, p.y - prev.y)
  } else if (pointers.size === 2) {
    travelled = DRAG_THRESHOLD + 1
    const [a, b] = [...pointers.values()]
    if (a === undefined || b === undefined || pinch === 0) return
    const d = distance(a, b)
    zoom(d / pinch, { x: (a.x + b.x) / 2, y: (a.y + b.y) / 2 })
    pinch = d
  }
}
function onUp(e: PointerEvent) {
  pointers.delete(e.pointerId)
  if (pointers.size < 2) pinch = 0
}
function onSpot(key: string) {
  if (travelled <= DRAG_THRESHOLD) emit('select', key)
}

let resize: ResizeObserver | null = null
onMounted(() => {
  if (box.value !== null) {
    resize = new ResizeObserver(() => {
      if (view.value.scale === fitted.value) fit()
    })
    resize.observe(box.value)
  }
})
onBeforeUnmount(() => resize?.disconnect())
watch(
  () => props.src,
  () => {
    natural.value = { w: 0, h: 0 }
  },
)
</script>

<template>
  <div
    ref="box"
    class="viewer"
    @wheel="onWheel"
    @pointerdown="onDown"
    @pointermove="onMove"
    @pointerup="onUp"
    @pointercancel="onUp"
  >
    <div class="stage" :style="stageStyle">
      <img :src="src" alt="" draggable="false" @load="onLoad" />
      <button
        v-for="(h, i) in hotspots"
        :key="i"
        type="button"
        class="hotspot"
        :class="{ selected: h.key === selected }"
        :style="spotStyle(h)"
        :title="titles?.[h.key] ?? h.caption"
        :aria-label="h.caption"
        @click="onSpot(h.key)"
      />
    </div>
    <div class="tools no-print">
      <v-btn :icon="mdiMagnifyPlus" size="small" :title="t('section.zoomIn')" @click="zoom(1.4)" />
      <v-btn
        :icon="mdiMagnifyMinus"
        size="small"
        :title="t('section.zoomOut')"
        @click="zoom(1 / 1.4)"
      />
      <v-btn :icon="mdiFitToScreen" size="small" :title="t('section.fit')" @click="fit" />
    </div>
  </div>
</template>

<style scoped>
.viewer {
  position: relative;
  overflow: hidden;
  touch-action: none;
  user-select: none;
  background: #fff;
  height: 100%;
  min-height: 320px;
  cursor: grab;
}
.stage {
  position: absolute;
  left: 0;
  top: 0;
  transform-origin: 0 0;
}
.stage img {
  display: block;
  width: 100%;
  height: 100%;
}
.hotspot {
  position: absolute;
  border: 2px solid transparent;
  border-radius: 4px;
  background: transparent;
  cursor: pointer;
}
.hotspot:hover {
  border-color: rgba(var(--v-theme-primary), 0.5);
}
.hotspot.selected {
  border-color: rgb(var(--v-theme-primary));
  background: rgba(var(--v-theme-primary), 0.18);
}
.tools {
  position: absolute;
  right: 8px;
  bottom: 8px;
  display: flex;
  gap: 4px;
}
@media print {
  .viewer {
    height: auto;
    min-height: 0;
  }
  .stage {
    position: static;
    transform: none !important;
    width: 100% !important;
    height: auto !important;
  }
  .stage img {
    height: auto;
  }
  .hotspot {
    display: none;
  }
}
</style>
