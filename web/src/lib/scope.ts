import type { LocationQuery, LocationQueryValue } from 'vue-router'

export interface Scope {
  vin?: string
  cat?: string
  model?: string
}

function one(v: LocationQueryValue | LocationQueryValue[] | undefined): string | undefined {
  const s = Array.isArray(v) ? v[0] : v
  return typeof s === 'string' && s.trim() !== '' ? s.trim() : undefined
}

export function scopeFromQuery(q: LocationQuery): Scope | null {
  const vin = one(q.vin)
  if (vin !== undefined) return { vin }
  const cat = one(q.cat)
  if (cat === undefined) return null
  const model = one(q.model)
  return model === undefined ? { cat } : { cat, model }
}

export function scopeToQuery(s: Scope | null): Record<string, string> {
  if (s === null) return {}
  if (s.vin !== undefined) return { vin: s.vin }
  const out: Record<string, string> = {}
  if (s.cat !== undefined) out.cat = s.cat
  if (s.model !== undefined) out.model = s.model
  return out
}

export function sameScope(a: Scope | null, b: Scope | null): boolean {
  return JSON.stringify(scopeToQuery(a)) === JSON.stringify(scopeToQuery(b))
}

export function isScope(v: unknown): v is Scope {
  if (typeof v !== 'object' || v === null) return false
  const o = v as Record<string, unknown>
  const ok = (k: string) => o[k] === undefined || typeof o[k] === 'string'
  return (
    ok('vin') &&
    ok('cat') &&
    ok('model') &&
    (typeof o.vin === 'string' || typeof o.cat === 'string')
  )
}

export function isScopeOrNull(v: unknown): v is Scope | null {
  return v === null || isScope(v)
}

export function scopeLabel(s: Scope): string {
  if (s.vin !== undefined) return s.vin
  return s.model === undefined ? (s.cat ?? '') : `${s.cat ?? ''} · ${s.model}`
}
