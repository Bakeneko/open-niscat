/** Interface languages: those of the NISCAT data. English has no URL prefix, the others /fr, /es, /de. */
export const LANGS = ['en', 'fr', 'es', 'de'] as const
export type Lang = (typeof LANGS)[number]

export function isLang(v: unknown): v is Lang {
  return LANGS.some((l) => l === v)
}

export function langFromParam(p: unknown): Lang {
  return isLang(p) ? p : 'en'
}

export function localizedPath(lang: Lang, path: string): string {
  const p = path.startsWith('/') ? path : `/${path}`
  if (lang === 'en') return p
  return p === '/' ? `/${lang}` : `/${lang}${p}`
}

/** Replaces the language prefix of a full path (path + query + hash). */
export function switchLang(fullPath: string, lang: Lang): string {
  const cut = fullPath.search(/[?#]/)
  const path = cut === -1 ? fullPath : fullPath.slice(0, cut)
  const rest = cut === -1 ? '' : fullPath.slice(cut)
  const m = /^\/(fr|es|de)(\/.*)?$/.exec(path)
  const bare = m === null ? path : (m[2] ?? '/')
  return localizedPath(lang, bare) + rest
}
