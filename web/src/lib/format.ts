export function formatYearMonth(ym: string | null | undefined, long = false): string {
  if (ym === null || ym === undefined || ym === '') return ''
  const m = /^(\d{4})-(\d{2})$/.exec(ym)
  if (m?.[1] === undefined || m[2] === undefined) return ym
  return long ? `${m[2]}/${m[1]}` : `${m[2]}/${m[1].slice(2)}`
}

export function formatRange(
  from: string | null | undefined,
  to: string | null | undefined,
): string {
  const f = formatYearMonth(from)
  const t = formatYearMonth(to)
  if (f === '' && t === '') return ''
  return `${f} – ${t}`.trim()
}
