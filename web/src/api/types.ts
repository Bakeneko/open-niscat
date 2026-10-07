import type { Lang } from '@/lib/lang'

/** "YYYY-MM"; null = open/unknown bound. */
export type YearMonth = string

export interface Meta {
  schema: number
  /** Data delivery: source edition YYYY.MM, then tool revision ("2015.01-2"). */
  version: string
  source: { name: string; publisher: string; edition: YearMonth }
  build: { tool: string; revision: number; date: string }
  defaultLang: Lang
  /** Build version of the program (git tag or commit). */
  appVersion: string
}
export interface Catalog {
  cat: string
  etd: string
  grupo: string
  model: string
  cmodel: string
  drive: string
  from: YearMonth | null
  to: YearMonth | null
  serie: string
  description: string
  langs: string[]
}
export interface Attribute {
  kind: 'T' | 'I'
  table: string
  name: string
  code: string
  value: string
}
export interface Vehicle {
  vin?: string
  model?: string
  prodDate?: YearMonth
  vinCount?: number
  catalog: Catalog
  attributes: Attribute[]
  documents: string[]
}
export interface VinMatch {
  vin: string
  model: string
  cat: string
  prodDate: YearMonth | null
}
export interface VinResult {
  vehicle?: Vehicle
  candidates?: VinMatch[]
}
export interface ModelInfo {
  model: string
  vinCount: number
  attributes: Attribute[]
}
export interface Group {
  code: string
  label: string
  image?: string
}
export interface Hotspot {
  caption: string
  key: string
  x: number
  y: number
  w: number
  h: number
}
export interface SectionSummary {
  sec: string
  name: string
  nameEn?: string
  notes: string
  from: YearMonth | null
  to: YearMonth | null
  applicable?: boolean
}
export interface GroupDetail {
  group: Group
  hotspots: Hotspot[]
  sections: SectionSummary[]
}
export interface RefLink {
  key: string
  partNo: string
}
export interface Line {
  id: string
  pospie: number
  /** NISCAT mark shown before the callout: "*", "#" or "" (meaning undocumented). */
  mark: string
  /** Drawing number as printed, on an item's first line only. */
  item: string
  /** Drawing number without leading zeros, on every line; matches Hotspot.key. */
  itemKey: string
  variant: string
  /** "01-02": drawing number then line number within the item (set on every line). */
  callout: string
  /** Indentation level (1-5): sub-assemblies are deeper. */
  level: number
  partNo: string
  partKey: string
  description: string
  spec: string
  qty: string
  /** Shown in brackets after the quantity, as NISCAT does (meaning undocumented). */
  cap: string
  /** Probably an interchangeability code ("2-0"); not documented by NISCAT. */
  ica: string
  /** Models this line applies to (free text). */
  app: string
  from: YearMonth | null
  to: YearMonth | null
  inPeriod?: boolean
  alternative: string
  alternativeKey: string
  /** K.D. flag ("*"), probably knock-down kits for local assembly. */
  kd: string
  /** Part Name Code: the 5-digit base shared by the variants of a part. */
  pnc: string
  latest?: RefLink
}
export interface LineRef extends Line {
  etd: string
  sec: string
}
export interface Section {
  etd: string
  sec: string
  plate: string
  group: Group
  name: string
  nameEn?: string
  notes: string
  from: YearMonth | null
  to: YearMonth | null
  applicable?: boolean
  image: string
  hotspots: Hotspot[]
  lines: Line[]
  prev?: string
  next?: string
}
export interface SectionHit {
  etd: string
  sec: string
  group: Group
  name: string
  nameEn?: string
  notes: string
}
export interface SearchResult {
  total: number
  truncated: boolean
  parts: LineRef[]
  sections: SectionHit[]
  vins: VinMatch[]
}
export interface PartInfo {
  key: string
  partNo: string
  description: string
  occurrences: LineRef[]
  previous: RefLink[]
  next: RefLink[]
}
export interface LinesResult {
  lines: LineRef[]
  missing: string[]
}
