export function refPath(partNo: string): string {
  return `/part/${encodeURIComponent(partNo.trim().replace(/^-/, ''))}`
}

export function lineId(etd: string, pospie: number): string {
  return `${etd}${String(pospie)}`
}
