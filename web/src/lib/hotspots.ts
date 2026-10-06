/** A group-index caption ("230") designates that section and its lettered variants (230A, 230B...). */
export function sectionMatchesCaption(sec: string, caption: string): boolean {
  const c = caption.trim().toUpperCase()
  const s = sec.trim().toUpperCase()
  return c !== '' && (s === c || (s.startsWith(c) && /^[A-Z]/.test(s.slice(c.length))))
}
