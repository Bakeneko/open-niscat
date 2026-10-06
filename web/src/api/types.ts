import type { Lang } from '@/lib/lang'

/** "YYYY-MM"; null = open/unknown bound. */
export type YearMonth = string

export interface Meta { schema: number; version: string; edition: string; built: string; defaultLang: Lang }
export interface Catalog {
  cat: string; etd: string; grupo: string; model: string; cmodel: string; drive: string
  from: YearMonth | null; to: YearMonth | null; serie: string; description: string; langs: string[]
}
export interface Attribute { kind: 'T' | 'I'; table: string; name: string; code: string; value: string }
export interface Vehicle {
  vin?: string; model?: string; prodDate?: YearMonth; vinCount?: number
  catalog: Catalog; attributes: Attribute[]; documents: string[]
}
export interface VinMatch { vin: string; model: string; cat: string; prodDate: YearMonth | null }
export interface VinResult { vehicle?: Vehicle; candidates?: VinMatch[] }
export interface ModelInfo { model: string; vinCount: number; attributes: Attribute[] }
export interface Group { code: string; label: string; image?: string }
export interface Hotspot { caption: string; key: string; x: number; y: number; w: number; h: number }
export interface SectionSummary { sec: string; name: string; notes: string; from: YearMonth | null; to: YearMonth | null; applicable?: boolean }
export interface GroupDetail { group: Group; hotspots: Hotspot[]; sections: SectionSummary[] }
export interface RefLink { key: string; partNo: string }
export interface Line {
  id: string; pospie: number; mark: string; item: string; itemKey: string; variant: string; level: number
  partNo: string; partKey: string; description: string; spec: string; qty: string; cap: string; ica: string; app: string
  from: YearMonth | null; to: YearMonth | null; inPeriod?: boolean
  alternative: string; alternativeKey: string; kd: string; pnc: string; latest?: RefLink
}
export interface LineRef extends Line { etd: string; sec: string }
export interface Section {
  etd: string; sec: string; plate: string; group: Group; name: string; notes: string
  from: YearMonth | null; to: YearMonth | null; applicable?: boolean; image: string
  hotspots: Hotspot[]; lines: Line[]; prev?: string; next?: string
}
export interface SectionHit { etd: string; sec: string; group: Group; name: string; notes: string }
export interface SearchResult { total: number; truncated: boolean; parts: LineRef[]; sections: SectionHit[] }
export interface PartInfo { key: string; partNo: string; description: string; occurrences: LineRef[]; previous: RefLink[]; next: RefLink[] }
export interface LinesResult { lines: LineRef[]; missing: string[] }
