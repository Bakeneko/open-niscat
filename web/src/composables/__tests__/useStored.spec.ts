import { describe, expect, it } from 'vitest'
import { effectScope, nextTick } from 'vue'
import { useStored } from '../useStored'

const isNumber = (v: unknown): v is number => typeof v === 'number'

describe('useStored', () => {
  it('is a singleton per key, persists, and follows other tabs', async () => {
    localStorage.clear()
    const a = useStored('t.k', isNumber, () => 0)
    const b = useStored('t.k', isNumber, () => 0)
    expect(a).toBe(b)
    a.value = 5
    await nextTick()
    expect(localStorage.getItem('t.k')).toBe('5')
    localStorage.setItem('t.k', '9')
    window.dispatchEvent(new StorageEvent('storage', { key: 't.k' }))
    expect(a.value).toBe(9)
  })
  it('keeps persisting after the first caller is gone', async () => {
    localStorage.clear()
    const owner = effectScope()
    const first = owner.run(() => useStored('t.scoped', isNumber, () => 0))
    owner.stop() // e.g. the page that called it first unmounts
    const later = useStored('t.scoped', isNumber, () => 0)
    expect(later).toBe(first)
    later.value = 7
    await nextTick()
    expect(localStorage.getItem('t.scoped')).toBe('7')
  })
})
