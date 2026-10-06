import { beforeEach, describe, expect, it } from 'vitest'
import { load, saveIfChanged } from '../storage'

const isNumber = (v: unknown): v is number => typeof v === 'number'

describe('storage', () => {
  beforeEach(() => {
    localStorage.clear()
  })
  it('round-trips and validates', () => {
    expect(saveIfChanged('k', 42)).toBe(true)
    expect(load('k', isNumber)).toBe(42)
    localStorage.setItem('k', '"text"')
    expect(load('k', isNumber)).toBeUndefined()
    localStorage.setItem('k', '{broken')
    expect(load('k', isNumber)).toBeUndefined()
  })
  it('does not rewrite an unchanged value (no storage-event ping-pong between tabs)', () => {
    expect(saveIfChanged('k', [1, 2])).toBe(true)
    expect(saveIfChanged('k', [1, 2])).toBe(false)
    expect(saveIfChanged('k', [1, 3])).toBe(true)
  })
  it('removes on undefined', () => {
    saveIfChanged('k', 1)
    saveIfChanged('k', undefined)
    expect(localStorage.getItem('k')).toBeNull()
  })
})
