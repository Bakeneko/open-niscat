import { describe, expect, it } from 'vitest'
import { sectionMatchesCaption } from '../hotspots'

describe('group hotspot captions', () => {
  it('match a section and its lettered variants only', () => {
    expect(sectionMatchesCaption('230', '230')).toBe(true)
    expect(sectionMatchesCaption('230A', '230')).toBe(true)
    expect(sectionMatchesCaption('186AB', '186')).toBe(true)
    expect(sectionMatchesCaption('2301', '230')).toBe(false)
    expect(sectionMatchesCaption('23', '230')).toBe(false)
    expect(sectionMatchesCaption('040', '040')).toBe(true)
    expect(sectionMatchesCaption('230a', ' 230 ')).toBe(true)
  })
})
