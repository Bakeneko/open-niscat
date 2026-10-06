import { describe, expect, it } from 'vitest'
import { formatRange, formatYearMonth, readable } from '../format'

describe('format', () => {
  it('formats year-months', () => {
    expect(formatYearMonth('1989-05')).toBe('05/89')
    expect(formatYearMonth('1989-05', true)).toBe('05/1989')
    expect(formatYearMonth(null)).toBe('')
    expect(formatYearMonth('garbage')).toBe('garbage')
  })
  it('formats ranges with open bounds', () => {
    expect(formatRange('1989-01', '1992-09')).toBe('01/89 – 09/92')
    expect(formatRange(null, '1988-02')).toBe('– 02/88')
    expect(formatRange('1994-11', null)).toBe('11/94 –')
    expect(formatRange(null, null)).toBe('')
  })
})

describe('readable', () => {
  it('adds a space after commas so long labels can wrap', () => {
    expect(readable('DEMARREUR,COMPLET')).toBe('DEMARREUR, COMPLET')
    expect(readable('A, B,C')).toBe('A, B, C')
    expect(readable('1,5 L')).toBe('1,5 L')
  })
})
