import { describe, expect, it } from 'vitest'
import {
  isScope,
  isScopeOrNull,
  sameScope,
  scopeFromQuery,
  scopeLabel,
  scopeToQuery,
} from '../scope'

describe('scope', () => {
  it('reads the query, VIN first', () => {
    expect(scopeFromQuery({ vin: 'V1', cat: 'AA-G01' })).toEqual({ vin: 'V1' })
    expect(scopeFromQuery({ cat: 'AA-G01', model: 'M' })).toEqual({ cat: 'AA-G01', model: 'M' })
    expect(scopeFromQuery({ cat: 'AA-G01' })).toEqual({ cat: 'AA-G01' })
    expect(scopeFromQuery({ model: 'M' })).toBeNull()
    expect(scopeFromQuery({ vin: ['V2', 'V3'] })).toEqual({ vin: 'V2' })
    expect(scopeFromQuery({ vin: '  ' })).toBeNull()
    expect(scopeFromQuery({})).toBeNull()
  })
  it('writes the query', () => {
    expect(scopeToQuery({ vin: 'V1' })).toEqual({ vin: 'V1' })
    expect(scopeToQuery({ cat: 'AA-G01', model: 'M' })).toEqual({ cat: 'AA-G01', model: 'M' })
    expect(scopeToQuery(null)).toEqual({})
  })
  it('compares and validates', () => {
    expect(sameScope({ vin: 'V1' }, { vin: 'V1' })).toBe(true)
    expect(sameScope({ vin: 'V1' }, null)).toBe(false)
    expect(sameScope(null, null)).toBe(true)
    expect(isScope({ vin: 'V1' })).toBe(true)
    expect(isScope({ model: 'M' })).toBe(false)
    expect(isScope({ vin: 3 })).toBe(false)
    expect(isScopeOrNull(null)).toBe(true)
  })
  it('labels', () => {
    expect(scopeLabel({ vin: 'V1' })).toBe('V1')
    expect(scopeLabel({ cat: 'AA-G01', model: 'M' })).toBe('AA-G01 · M')
    expect(scopeLabel({ cat: 'AA-G01' })).toBe('AA-G01')
  })
})
