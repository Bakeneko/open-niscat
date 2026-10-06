export type Lang = 'en' | 'fr'

export function isLang(v: unknown): v is Lang {
  return v === 'en' || v === 'fr'
}

export function langFromParam(p: unknown): Lang {
  return p === 'fr' ? 'fr' : 'en'
}

export function localizedPath(lang: Lang, path: string): string {
  const p = path.startsWith('/') ? path : `/${path}`
  if (lang === 'en') return p
  return p === '/' ? '/fr' : `/fr${p}`
}

/** Toggles the /fr prefix of a full path (path + query + hash). */
export function switchLang(fullPath: string, lang: Lang): string {
  const cut = fullPath.search(/[?#]/)
  const path = cut === -1 ? fullPath : fullPath.slice(0, cut)
  const rest = cut === -1 ? '' : fullPath.slice(cut)
  let bare = path
  if (path === '/fr') bare = '/'
  else if (path.startsWith('/fr/')) bare = path.slice(3)
  return localizedPath(lang, bare) + rest
}
