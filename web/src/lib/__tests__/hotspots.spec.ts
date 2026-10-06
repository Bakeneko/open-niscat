import { describe, expect, it } from 'vitest'
import { groupHotspots, sectionMatchesCaption } from '../hotspots'

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

const spot = (caption: string) => ({ caption, key: caption, x: 0, y: 0, w: 1, h: 1 })

describe('group hotspot visibility', () => {
  const sections = [
    { sec: '230A', applicable: true },
    { sec: '221', applicable: false },
    { sec: '233' }, // no vehicle scope: applicability unknown
  ]
  const spots = [spot('230'), spot('221'), spot('233'), spot('999')]
  it('hides hotspots without sections, and non-applicable ones unless asked', () => {
    expect(groupHotspots(spots, sections, false)).toEqual({
      shown: [spot('230'), spot('233')],
      muted: [],
    })
    expect(groupHotspots(spots, sections, true)).toEqual({
      shown: [spot('230'), spot('221'), spot('233')],
      muted: ['221'],
    })
  })
})
