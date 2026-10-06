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
  /** Write captions over the drawing, hiding the printed numbers (as NISCAT does). */
  labels?: boolean
  /** Hotspot keys shown greyed out (e.g. sections not applicable to the vehicle). */
  muted?: string[]
}>()
const emit = defineEmits<{ select: [key: string] }>()
const { t } = useI18n()

const box = ref<HTMLDivElement | null>(null)
const view = ref<View>({ scale: 1, x: 0, y: 0 })
const natural = ref({ w: 0, h: 0 })
const fitted = ref(1)
const DRAG_THRESHOLD = 6

const failed = ref(false)
const DOUBLE_TAP_MS = 300

// Pointers are tracked by hand: capture starts only once a drag begins, so a plain click still reaches the
// hotspot under it; a hotspot is activated on release when the gesture stayed a tap (one pointer, no drag).
const pointers = new Map<number, { x: number; y: number }>()
let travelled = 0
let pinch = 0
let multi = false
let downKey: string | null = null
let lastTap = 0

const stageStyle = computed(() => ({
  transform: `translate(${String(view.value.x)}px, ${String(view.value.y)}px) scale(${String(view.value.scale)})`,
  width: `${String(natural.value.w)}px`,
  height: `${String(natural.value.h)}px`,
}))
// Hotspots are placed in percent of the drawing and captions sized in cqw (the stage is a size container):
// identical on screen, and they follow the image when printing resizes it to the page.
const pct = (v: number, of: number) => `${String((v / of) * 100)}%`
const spotStyle = (h: Hotspot) => {
  const { w, h: ih } = natural.value
  return w === 0 || ih === 0
    ? { display: 'none' }
    : { left: pct(h.x, w), top: pct(h.y, ih), width: pct(h.w, w), height: pct(h.h, ih) }
}
const captionStyle = (h: Hotspot) => ({
  fontSize: natural.value.w === 0 ? '0' : `${String(((h.h * 0.6) / natural.value.w) * 100)}cqw`,
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
  if (e.deltaY === 0) return
  e.preventDefault()
  // Smooth for trackpads (small deltas), about x1.2 per mouse notch; deltaMode 1 counts lines.
  const dy = e.deltaMode === 1 ? e.deltaY * 33 : e.deltaY
  zoom(Math.exp(-Math.max(-300, Math.min(300, dy)) * 0.002), local(e))
}
function spotKey(target: EventTarget | null): string | null {
  return target instanceof Element
    ? (target.closest<HTMLElement>('[data-key]')?.dataset.key ?? null)
    : null
}
// Hover hint: the browser's title tooltip waits about a second, too long while scanning a drawing.
const TIP_DELAY = 300
const tip = ref<{ text: string; x: number; y: number } | null>(null)
let tipTimer: ReturnType<typeof setTimeout> | undefined
function tipStyle(at: { x: number; y: number }) {
  const w = box.value?.clientWidth ?? 0
  return at.x > w * 0.6
    ? { right: `${String(w - at.x + 12)}px`, top: `${String(at.y + 16)}px` }
    : { left: `${String(at.x + 12)}px`, top: `${String(at.y + 16)}px` }
}
function onSpotEnter(e: PointerEvent, h: Hotspot) {
  if (e.pointerType !== 'mouse' || pointers.size > 0) return
  clearTimeout(tipTimer)
  const at = local(e)
  tipTimer = setTimeout(() => {
    tip.value = { text: props.titles?.[h.key] ?? h.caption, ...at }
  }, TIP_DELAY)
}
function onSpotMove(e: PointerEvent) {
  if (tip.value !== null) tip.value = { ...tip.value, ...local(e) }
}
function hideTip() {
  clearTimeout(tipTimer)
  tip.value = null
}

function onDown(e: PointerEvent) {
  hideTip()
  if (e.target instanceof Element && e.target.closest('.tools') !== null) return
  if (e.pointerType === 'mouse' && e.button !== 0) return
  if (e.isPrimary) {
    pointers.clear() // drops pointers whose release was lost
    travelled = 0
    multi = false
    downKey = spotKey(e.target)
  }
  pointers.set(e.pointerId, local(e))
  if (pointers.size >= 2) {
    multi = true
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
    if (travelled > DRAG_THRESHOLD && !box.value?.hasPointerCapture(e.pointerId)) {
      box.value?.setPointerCapture(e.pointerId)
    }
    view.value = panBy(view.value, p.x - prev.x, p.y - prev.y)
  } else if (pointers.size === 2) {
    const [a, b] = [...pointers.values()]
    if (a === undefined || b === undefined || pinch === 0) return
    const d = distance(a, b)
    zoom(d / pinch, { x: (a.x + b.x) / 2, y: (a.y + b.y) / 2 })
    pinch = d
  }
}
function onUp(e: PointerEvent) {
  if (!pointers.delete(e.pointerId)) return
  if (pointers.size < 2) pinch = 0
  if (pointers.size > 0 || multi || travelled > DRAG_THRESHOLD) return
  const key = spotKey(e.target)
  if (key !== null && key === downKey) {
    emit('select', key)
    lastTap = 0
  } else if (key === null && e.pointerType !== 'mouse') {
    // Double tap on the background fits the drawing (mouse users have dblclick).
    if (e.timeStamp - lastTap < DOUBLE_TAP_MS) {
      fit()
      lastTap = 0
    } else lastTap = e.timeStamp
  }
}
function onCancel(e: PointerEvent) {
  pointers.delete(e.pointerId)
  if (pointers.size < 2) pinch = 0
  multi = true // a cancelled gesture never activates a hotspot
}
function onKeySpot(e: MouseEvent, key: string) {
  if (e.detail === 0) emit('select', key) // keyboard activation; pointer taps are handled in onUp
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
onBeforeUnmount(() => {
  resize?.disconnect()
  clearTimeout(tipTimer)
})
watch(
  () => props.src,
  () => {
    natural.value = { w: 0, h: 0 }
    failed.value = false
    pointers.clear()
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
    @pointercancel="onCancel"
    @lostpointercapture="onCancel"
    @dblclick="fit"
  >
    <v-alert v-if="failed" type="warning" variant="tonal" class="ma-2">{{
      t('section.noImage')
    }}</v-alert>
    <div v-else class="stage" :style="stageStyle">
      <img :src="src" alt="" draggable="false" @load="onLoad" @error="failed = true" />
      <button
        v-for="(h, i) in hotspots"
        :key="i"
        type="button"
        class="hotspot"
        :class="{ selected: h.key === selected, labelled: labels, muted: muted?.includes(h.key) }"
        :style="spotStyle(h)"
        :data-key="h.key"
        :aria-label="h.caption"
        @pointerenter="onSpotEnter($event, h)"
        @pointermove="onSpotMove"
        @pointerleave="hideTip"
        @click="onKeySpot($event, h.key)"
      >
        <span v-if="labels" class="caption" :style="captionStyle(h)">{{ h.caption }}</span>
      </button>
    </div>
    <div v-if="tip" class="tip" :style="tipStyle(tip)">{{ tip.text }}</div>
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
  container-type: inline-size;
}
.stage img {
  display: block;
  width: 100%;
  height: 100%;
}
.hotspot {
  position: absolute;
  border: 2px solid rgba(211, 47, 47, 0.45);
  border-radius: 4px;
  background: transparent;
  cursor: pointer;
}
.hotspot:hover,
.hotspot:focus-visible {
  border-color: rgb(211, 47, 47);
  outline: none;
}
/* Labelled hotspots (group indexes), like NISCAT: the caption is written in red on an opaque white patch
   hiding the stale printed number; the clickable zone itself stays invisible until hovered. */
.hotspot.labelled {
  display: flex;
  align-items: center;
  justify-content: flex-start; /* printed numbers start at the zone's left edge */
  padding: 0;
  border-color: transparent;
  background: transparent;
}
.hotspot.labelled:hover,
.hotspot.labelled:focus-visible {
  border-color: transparent;
}
.hotspot.labelled:hover .caption,
.hotspot.labelled:focus-visible .caption {
  box-shadow: 0 0 0 2px rgba(211, 47, 47, 0.6);
}
.hotspot .caption {
  padding: 0.08em 0.25em;
  border-radius: 3px;
  background: #fff;
  color: rgb(211, 47, 47);
  font-weight: 700;
  line-height: 1;
  pointer-events: none;
}
.hotspot.labelled.muted {
  border-color: transparent;
}
.hotspot.labelled.muted .caption {
  color: rgb(150, 150, 150);
  outline: 1px dashed rgb(150, 150, 150);
}
.hotspot.labelled.selected {
  border-color: transparent;
  background: transparent;
}
.hotspot.labelled.selected .caption {
  color: rgb(var(--v-theme-primary));
  box-shadow: 0 0 0 3px rgb(var(--v-theme-primary));
}
.hotspot.muted {
  border-style: dashed;
  border-color: rgba(120, 120, 120, 0.6);
}
.hotspot.selected {
  border-color: rgb(var(--v-theme-primary));
  background: rgba(var(--v-theme-primary), 0.18);
}
.tip {
  position: absolute;
  z-index: 3;
  max-width: 360px;
  padding: 4px 8px;
  border-radius: 4px;
  background: rgba(33, 33, 33, 0.92);
  color: #fff;
  font-size: 0.8rem;
  line-height: 1.35;
  white-space: pre-line;
  pointer-events: none;
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
    height: auto !important;
    min-height: 0;
    overflow: visible;
  }
  .stage {
    position: relative;
    transform: none !important;
    width: 100% !important;
    height: auto !important;
  }
  .stage img {
    height: auto;
  }
  /* Captions are printed (they replace stale printed numbers); bare zones and hover/selection marks are not. */
  .hotspot:not(.labelled),
  .tip {
    display: none;
  }
  .hotspot.labelled {
    border-color: transparent !important;
  }
  .hotspot .caption {
    box-shadow: none !important;
    outline: none !important;
    color: rgb(211, 47, 47) !important;
  }
}
</style>
