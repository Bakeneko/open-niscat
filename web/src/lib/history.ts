export interface HistoryEntry {
  kind: 'vehicle' | 'section'
  label: string
  path: string
  at: number
}

export const MAX_HISTORY = 20

export function pushHistory(list: readonly HistoryEntry[], e: HistoryEntry): HistoryEntry[] {
  return [e, ...list.filter((x) => x.path !== e.path)].slice(0, MAX_HISTORY)
}

function isEntry(v: unknown): v is HistoryEntry {
  if (typeof v !== 'object' || v === null) return false
  const o = v as Record<string, unknown>
  return (
    (o.kind === 'vehicle' || o.kind === 'section') &&
    typeof o.label === 'string' &&
    typeof o.path === 'string' &&
    typeof o.at === 'number'
  )
}

export function isHistory(v: unknown): v is HistoryEntry[] {
  return Array.isArray(v) && v.every(isEntry)
}
