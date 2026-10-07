const APP = 'Open Niscat'

/** Browser tab title: "<page> · Open Niscat", or the app name alone (home page). */
export function pageTitle(page: string | null): string {
  return page === null || page === '' ? APP : `${page} · ${APP}`
}

/** A plate: series and section code, the selected item number if any, then the section name. */
export function sectionTitle(
  s: { etd: string; sec: string; name: string },
  item: string | null,
): string {
  return `${s.etd} ${s.sec}${item === null ? '' : ` ${item}`} — ${s.name}`
}
