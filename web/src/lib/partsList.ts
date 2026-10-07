import { isScope, type Scope } from './scope'

export interface ListItem {
  id: string
  qty: number
  scope?: Scope
}

export interface ListRow {
  reference: string
  description: string
  qty: number
  section: string
}

export const MAX_ITEMS = 500
const MAX_QTY = 9999
const ID = /^([A-Z]{2})0*([1-9]\d*)$/

/** Canonical line id ("aa02605" -> "AA2605", as the API returns it), or null when malformed. */
export function normalizeId(id: string): string | null {
  const m = ID.exec(id.trim().toUpperCase())
  return m === null ? null : `${m[1] ?? ''}${m[2] ?? ''}`
}

export function addItem(
  items: readonly ListItem[],
  id: string,
  qty = 1,
  scope?: Scope,
): ListItem[] {
  const nid = normalizeId(id)
  if (nid === null || !Number.isInteger(qty) || qty <= 0) return [...items]
  if (items.some((i) => i.id === nid)) {
    return items.map((i) => (i.id === nid ? { ...i, qty: Math.min(i.qty + qty, MAX_QTY) } : i))
  }
  if (items.length >= MAX_ITEMS) return [...items]
  return [...items, scope === undefined ? { id: nid, qty } : { id: nid, qty, scope }]
}

export function setQty(items: readonly ListItem[], id: string, qty: number): ListItem[] {
  if (!Number.isFinite(qty) || qty < 1) return removeItem(items, id)
  return items.map((i) => (i.id === id ? { ...i, qty: Math.min(Math.floor(qty), MAX_QTY) } : i))
}

export function removeItem(items: readonly ListItem[], id: string): ListItem[] {
  return items.filter((i) => i.id !== id)
}

export function mergeItems(a: readonly ListItem[], b: readonly ListItem[]): ListItem[] {
  let out = [...a]
  for (const i of b) out = addItem(out, i.id, i.qty, i.scope)
  return out
}

export function encodeList(items: readonly ListItem[]): string {
  return items.map((i) => `${i.id}x${String(i.qty)}`).join(',')
}

export function decodeList(s: string): { items: ListItem[]; invalid: string[] } {
  let items: ListItem[] = []
  const invalid: string[] = []
  for (const raw of s.split(',')) {
    const part = raw.trim()
    if (part === '') continue
    const m = /^([A-Za-z]{2}\d+)x(\d{1,4})$/.exec(part)
    const qty = m?.[2] === undefined ? 0 : Number(m[2])
    if (m?.[1] === undefined || qty <= 0) {
      invalid.push(part)
      continue
    }
    items = addItem(items, m[1], qty)
  }
  return { items, invalid }
}

function isItem(v: unknown): v is ListItem {
  if (typeof v !== 'object' || v === null) return false
  const o = v as Record<string, unknown>
  return (
    typeof o.id === 'string' &&
    normalizeId(o.id) === o.id &&
    typeof o.qty === 'number' &&
    Number.isInteger(o.qty) &&
    o.qty > 0 &&
    (o.scope === undefined || isScope(o.scope))
  )
}

export function isPartsList(v: unknown): v is ListItem[] {
  return Array.isArray(v) && v.every(isItem)
}

// Spreadsheets evaluate cells starting with = + - @ ("-23319-D9700" would become a formula): wrap them as ="...".
const asText = (s: string) => (/^[=+\-@]/.test(s) ? `="${s.replace(/"/g, '""')}"` : s)
const cells = (r: ListRow) => [r.reference, r.description, String(r.qty), r.section].map(asText)

const clean = (s: string) => s.replace(/[\t\r\n]+/g, ' ')
const quote = (s: string) => (/[";\r\n]/.test(s) ? `"${s.replace(/"/g, '""')}"` : s)

export function toTSV(headers: readonly string[], rows: readonly ListRow[]): string {
  return [headers, ...rows.map(cells)].map((r) => r.map(clean).join('\t')).join('\r\n')
}

export function toCSV(headers: readonly string[], rows: readonly ListRow[]): string {
  return [headers, ...rows.map(cells)].map((r) => r.map(quote).join(';')).join('\r\n') + '\r\n'
}
