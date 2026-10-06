/** Search result tabs, in display (and priority) order. */
export const SEARCH_TABS = ['vins', 'sections', 'parts'] as const
export type SearchTab = (typeof SEARCH_TABS)[number]

export function isSearchTab(v: unknown): v is SearchTab {
  return SEARCH_TABS.some((t) => t === v)
}

/** The tab a search opens on: the first one with results, sections when there are none. */
export function defaultTab(counts: Partial<Record<SearchTab, number>>): SearchTab {
  return SEARCH_TABS.find((t) => (counts[t] ?? 0) > 0) ?? 'sections'
}
