import { describe, expect, it } from 'vitest'
import { lineId, refPath } from '../refs'

describe('refs', () => {
  it('builds part paths and line ids', () => {
    expect(refPath('-23319-D9700')).toBe('/part/23319-D9700')
    expect(refPath('D-4100-17C90')).toBe('/part/D-4100-17C90')
    expect(refPath('A B/C')).toBe('/part/A%20B%2FC')
    expect(lineId('AA', 2605)).toBe('AA2605')
  })
})
