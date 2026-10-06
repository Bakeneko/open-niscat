/** A group-index caption ("230") designates that section and its lettered variants (230A, 230B...). */
export function sectionMatchesCaption(sec: string, caption: string): boolean {
  const c = caption.trim().toUpperCase()
  const s = sec.trim().toUpperCase()
  return c !== '' && (s === c || (s.startsWith(c) && /^[A-Z]/.test(s.slice(c.length))))
}

interface SpotLike {
  caption: string
  key: string
}

/**
 * Group-index hotspots as NISCAT shows them: a hotspot without any section is hidden, and so is one whose
 * sections all do not apply to the vehicle — unless non-applicable sections are shown, then it is muted.
 */
export function groupHotspots<H extends SpotLike>(
  hotspots: readonly H[],
  sections: readonly { sec: string; applicable?: boolean }[],
  showAll: boolean,
): { shown: H[]; muted: string[] } {
  const shown: H[] = []
  const muted: string[] = []
  for (const h of hotspots) {
    const matching = sections.filter((s) => sectionMatchesCaption(s.sec, h.caption))
    if (matching.length === 0) continue
    const applicable = matching.some((s) => s.applicable !== false)
    if (!applicable && !showAll) continue
    shown.push(h)
    if (!applicable && !muted.includes(h.key)) muted.push(h.key)
  }
  return { shown, muted }
}
