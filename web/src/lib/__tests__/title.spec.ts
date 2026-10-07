import { describe, expect, it } from 'vitest'
import { pageTitle, sectionTitle } from '../title'

describe('page titles', () => {
  it('suffixes the page with the app name', () => {
    expect(pageTitle('Catalogs')).toBe('Catalogs · Open Niscat')
    expect(pageTitle(null)).toBe('Open Niscat')
    expect(pageTitle('')).toBe('Open Niscat')
  })
  it('names a plate, with the selected item when there is one', () => {
    expect(sectionTitle({ etd: 'AA', sec: '230A', name: 'Embrayage' }, null)).toBe(
      'AA 230A — Embrayage',
    )
    expect(sectionTitle({ etd: 'AA', sec: '230A', name: 'Embrayage' }, '02')).toBe(
      'AA 230A 02 — Embrayage',
    )
  })
})
