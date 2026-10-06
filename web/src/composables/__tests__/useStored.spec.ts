import { describe, expect, it } from 'vitest'
import { nextTick } from 'vue'
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
})
