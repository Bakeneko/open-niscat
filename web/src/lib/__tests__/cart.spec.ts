import { describe, expect, it } from 'vitest'
import {
  MAX_ITEMS,
  addItem,
  decodeCart,
  encodeCart,
  isCart,
  mergeItems,
  normalizeId,
  removeItem,
  setQty,
  toCSV,
  toTSV,
} from '../cart'

describe('cart', () => {
  it('normalizes ids', () => {
    expect(normalizeId(' aa2605 ')).toBe('AA2605')
    expect(normalizeId('AA02605')).toBe('AA2605') // the API answers with the canonical id
    expect(normalizeId('AA')).toBeNull()
    expect(normalizeId('A12')).toBeNull()
    expect(normalizeId('AA+3')).toBeNull()
  })
  it('adds, merges quantities, edits and removes', () => {
    let items = addItem([], 'AA1', 2, { vin: 'V' })
    items = addItem(items, 'aa1', 3)
    expect(items).toEqual([{ id: 'AA1', qty: 5, scope: { vin: 'V' } }])
    items = setQty(items, 'AA1', 1)
    expect(items[0]?.qty).toBe(1)
    expect(setQty(items, 'AA1', 0)).toEqual([])
    expect(removeItem(items, 'AA1')).toEqual([])
    expect(addItem([], 'bad', 1)).toEqual([])
    expect(addItem([], 'AA1', 0)).toEqual([])
  })
  it('caps the number of lines', () => {
    let items: ReturnType<typeof addItem> = []
    for (let i = 1; i <= MAX_ITEMS + 5; i++) items = addItem(items, `AA${String(i)}`)
    expect(items).toHaveLength(MAX_ITEMS)
  })
  it('encodes and decodes share links, reporting junk', () => {
    const items = [
      { id: 'AA2605', qty: 2 },
      { id: 'AA5658', qty: 8 },
    ]
    expect(encodeCart(items)).toBe('AA2605x2,AA5658x8')
    expect(decodeCart('AA2605x2,aa5658x8')).toEqual({ items, invalid: [] })
    expect(decodeCart('AA2605x2,AA2605x1,ZZ,AA1x0,,AA7x99999')).toEqual({
      items: [{ id: 'AA2605', qty: 3 }],
      invalid: ['ZZ', 'AA1x0', 'AA7x99999'],
    })
    expect(decodeCart('')).toEqual({ items: [], invalid: [] })
  })
  it('merges carts', () => {
    expect(
      mergeItems(
        [{ id: 'AA1', qty: 1 }],
        [
          { id: 'AA1', qty: 2 },
          { id: 'AB2', qty: 1 },
        ],
      ),
    ).toEqual([
      { id: 'AA1', qty: 3 },
      { id: 'AB2', qty: 1 },
    ])
  })
  it('validates stored carts', () => {
    expect(isCart([{ id: 'AA1', qty: 1 }])).toBe(true)
    expect(isCart([{ id: 'AA1', qty: 1.5 }])).toBe(false)
    expect(isCart([{ id: 'nope', qty: 1 }])).toBe(false)
    expect(isCart({})).toBe(false)
  })
  it('exports TSV and CSV', () => {
    const rows = [
      { reference: '-23319-D9700', description: 'PALIER; "X"\tY', qty: 2, section: 'AA 233C' },
    ]
    const headers = ['Reference', 'Description', 'Qty', 'Section']
    expect(toTSV(headers, rows)).toBe(
      'Reference\tDescription\tQty\tSection\r\n="-23319-D9700"\tPALIER; "X" Y\t2\tAA 233C',
    )
    expect(toCSV(headers, rows)).toBe(
      'Reference;Description;Qty;Section\r\n"=""-23319-D9700""";"PALIER; ""X""\tY";2;AA 233C\r\n',
    )
  })
  it('keeps cells that look like formulas as text', () => {
    const rows = [{ reference: '+1', description: '=SUM(A1)', qty: 1, section: '@x' }]
    expect(toTSV([], rows)).toBe('\r\n="+1"\t="=SUM(A1)"\t1\t="@x"')
  })
})
