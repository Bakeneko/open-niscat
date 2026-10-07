import { describe, expect, it } from 'vitest'
import { MAX_HISTORY, historyPath, isHistory, pushHistory, type HistoryEntry } from '../history'

const e = (path: string, at = 1): HistoryEntry => ({ kind: 'section', label: path, path, at })

describe('history', () => {
  it('keeps the newest first, without duplicates, bounded', () => {
    let list = pushHistory([], e('/a'))
    list = pushHistory(list, e('/b'))
    list = pushHistory(list, e('/a', 2))
    expect(list.map((x) => x.path)).toEqual(['/a', '/b'])
    for (let i = 0; i < 30; i++) list = pushHistory(list, e(`/x${String(i)}`))
    expect(list).toHaveLength(MAX_HISTORY)
  })
  it('stores a path without the language prefix nor the selected item', () => {
    expect(historyPath('/fr/section/AA/230A?vin=X&item=2')).toBe('/section/AA/230A?vin=X')
    expect(historyPath('/de/vin/VSKJ1234?item=3#top')).toBe('/vin/VSKJ1234#top')
    expect(historyPath('/section/AA/230A?item=2')).toBe('/section/AA/230A')
    expect(historyPath('/es')).toBe('/')
    expect(historyPath('/vin/X?model=A%20B&item=1')).toBe('/vin/X?model=A%20B')
    expect(historyPath('/section/AA/1?item=3&tab=info')).toBe('/section/AA/1')
    expect(historyPath('/vin/X?q=a?b&item=1#x#y')).toBe('/vin/X?q=a?b#x#y')
    expect(historyPath('/vin/X?item&items=2')).toBe('/vin/X?items=2')
  })
  it('treats the same page in another language or with a selected item as one entry', () => {
    let list = pushHistory([], e('/section/AA/230A?vin=X'))
    list = pushHistory(list, e('/fr/section/AA/230A?vin=X&item=2', 2))
    expect(list.map((x) => [x.path, x.at])).toEqual([['/section/AA/230A?vin=X', 2]])
  })
  it('merges duplicates left by older stored entries', () => {
    const old = [e('/fr/section/AA/230A'), e('/section/AA/230A?item=1')]
    expect(pushHistory(old, e('/section/BB/100')).map((x) => x.path)).toEqual([
      '/section/BB/100',
      '/section/AA/230A',
    ])
  })
  it('validates stored history', () => {
    expect(isHistory([e('/a')])).toBe(true)
    expect(isHistory([{ kind: 'other', label: '', path: '/', at: 1 }])).toBe(false)
  })
})
