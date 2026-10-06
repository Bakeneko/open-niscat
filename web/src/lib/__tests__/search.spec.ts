import { describe, expect, it } from 'vitest'
import { defaultTab, isSearchTab, SEARCH_TABS } from '../search'

describe('search tabs', () => {
  it('lists VINs, sections then parts', () => {
    expect(SEARCH_TABS).toEqual(['vins', 'sections', 'parts'])
    expect(isSearchTab('vins')).toBe(true)
    expect(isSearchTab('bogus')).toBe(false)
  })
  it('opens the first tab with results, VINs first', () => {
    expect(defaultTab({ vins: 1, sections: 3, parts: 9 })).toBe('vins')
    expect(defaultTab({ vins: 0, sections: 3, parts: 9 })).toBe('sections')
    expect(defaultTab({ vins: 0, sections: 0, parts: 9 })).toBe('parts')
  })
  it('falls back to sections when nothing matches or counts are missing', () => {
    expect(defaultTab({ vins: 0, sections: 0, parts: 0 })).toBe('sections')
    expect(defaultTab({})).toBe('sections')
  })
})
