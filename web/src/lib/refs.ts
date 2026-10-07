/** Part page path. NISCAT prefixes many part numbers with "-"; the URL drops it (the API matches letters and digits). */
export function refPath(partNo: string): string {
  return `/part/${encodeURIComponent(partNo.trim().replace(/^-/, ''))}`
}

/** Line id as the API returns it: series + pospie (row number in the NISCAT parts table). */
export function lineId(etd: string, pospie: number): string {
  return `${etd}${String(pospie)}`
}
