import { switchLang } from './lang'

export interface HistoryEntry {
  kind: 'vehicle' | 'section'
  label: string
  path: string
  at: number
}

export const MAX_HISTORY = 20

/** Query parameters that only reflect the view state of a page: the selected item and the phone tab. */
const VIEW_PARAMS = new Set(['item', 'tab'])

/**
 * The page an entry stands for: no language prefix (it opens in the current language) and no view state,
 * so the same plate seen in another language or with another callout selected is one entry.
 */
export function historyPath(fullPath: string): string {
  const full = switchLang(fullPath, 'en')
  const h = full.indexOf('#')
  const hash = h === -1 ? '' : full.slice(h)
  const rest = h === -1 ? full : full.slice(0, h)
  const q = rest.indexOf('?')
  const path = q === -1 ? rest : rest.slice(0, q)
  // Filtered by hand: URLSearchParams would re-encode the other parameters (a space becomes "+").
  const kept = (q === -1 ? '' : rest.slice(q + 1))
    .split('&')
    .filter((p) => p !== '' && !VIEW_PARAMS.has(p.split('=', 1)[0] ?? ''))
  return path + (kept.length > 0 ? `?${kept.join('&')}` : '') + hash
}

/** Normalised paths, newest first, without duplicates (also merges entries stored before normalisation). */
export function dedupeHistory(list: readonly HistoryEntry[]): HistoryEntry[] {
  const out: HistoryEntry[] = []
  for (const x of list) {
    const path = historyPath(x.path)
    if (!out.some((o) => o.path === path)) out.push({ ...x, path })
  }
  return out
}

export function pushHistory(list: readonly HistoryEntry[], e: HistoryEntry): HistoryEntry[] {
  return dedupeHistory([e, ...list]).slice(0, MAX_HISTORY)
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
