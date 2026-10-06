/** Image-to-screen transform: screen = image * scale + (x, y). */
export interface View {
  scale: number
  x: number
  y: number
}

export function fitView(cw: number, ch: number, iw: number, ih: number): View {
  if (cw <= 0 || ch <= 0 || iw <= 0 || ih <= 0) return { scale: 1, x: 0, y: 0 }
  const scale = Math.min(cw / iw, ch / ih)
  return { scale, x: (cw - iw * scale) / 2, y: (ch - ih * scale) / 2 }
}

export function zoomAt(v: View, factor: number, px: number, py: number, min: number, max: number): View {
  const scale = Math.min(max, Math.max(min, v.scale * factor))
  const k = scale / v.scale
  return { scale, x: px - (px - v.x) * k, y: py - (py - v.y) * k }
}

export function panBy(v: View, dx: number, dy: number): View {
  return { scale: v.scale, x: v.x + dx, y: v.y + dy }
}

export function distance(a: { x: number; y: number }, b: { x: number; y: number }): number {
  return Math.hypot(a.x - b.x, a.y - b.y)
}
