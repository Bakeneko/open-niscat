import { describe, expect, it } from 'vitest'
import { distance, fitView, panBy, zoomAt } from '../viewport'

describe('viewport', () => {
  it('fits and centers the image', () => {
    expect(fitView(1000, 500, 2000, 500)).toEqual({ scale: 0.5, x: 0, y: 125 })
    expect(fitView(0, 500, 2000, 500)).toEqual({ scale: 1, x: 0, y: 0 })
  })
  it('zooms around a point, keeping it fixed, within bounds', () => {
    const v = zoomAt({ scale: 1, x: 0, y: 0 }, 2, 100, 50, 0.1, 4)
    expect(v).toEqual({ scale: 2, x: -100, y: -50 })
    // The image point under the cursor (100, 50) stays under it.
    expect((100 - v.x) / v.scale).toBe(100)
    expect(zoomAt({ scale: 3, x: 0, y: 0 }, 2, 0, 0, 0.1, 4).scale).toBe(4)
    expect(zoomAt({ scale: 0.2, x: 0, y: 0 }, 0.1, 0, 0, 0.1, 4).scale).toBe(0.1)
  })
  it('pans and measures', () => {
    expect(panBy({ scale: 1, x: 1, y: 2 }, 3, 4)).toEqual({ scale: 1, x: 4, y: 6 })
    expect(distance({ x: 0, y: 0 }, { x: 3, y: 4 })).toBe(5)
  })
})
