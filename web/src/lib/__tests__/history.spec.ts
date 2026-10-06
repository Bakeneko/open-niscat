import { describe, expect, it } from 'vitest'
import { MAX_HISTORY, isHistory, pushHistory, type HistoryEntry } from '../history'

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
  it('validates stored history', () => {
    expect(isHistory([e('/a')])).toBe(true)
    expect(isHistory([{ kind: 'other', label: '', path: '/', at: 1 }])).toBe(false)
  })
})
