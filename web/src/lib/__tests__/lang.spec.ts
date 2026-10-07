import { describe, expect, it } from 'vitest'
import { isLang, LANGS, langFromParam, localizedPath, switchLang } from '../lang'

describe('lang', () => {
  it('lists the data languages, English first', () => {
    expect(LANGS).toEqual(['en', 'fr', 'es', 'de'])
  })
  it('parses the route param', () => {
    expect(langFromParam('fr')).toBe('fr')
    expect(langFromParam('es')).toBe('es')
    expect(langFromParam('de')).toBe('de')
    expect(langFromParam(undefined)).toBe('en')
    expect(langFromParam(['fr'])).toBe('en')
    expect(isLang('de')).toBe(true)
    expect(isLang('it')).toBe(false)
  })
  it('localizes paths', () => {
    expect(localizedPath('en', '/list')).toBe('/list')
    expect(localizedPath('fr', '/list')).toBe('/fr/list')
    expect(localizedPath('de', '/')).toBe('/de')
    expect(localizedPath('es', 'list')).toBe('/es/list')
  })
  it('switches language keeping path, query and hash', () => {
    expect(switchLang('/section/AA/233C?vin=X&item=4&tab=info', 'fr')).toBe(
      '/fr/section/AA/233C?vin=X&item=4&tab=info',
    )
    expect(switchLang('/fr/section/AA/233C?vin=X#top', 'en')).toBe('/section/AA/233C?vin=X#top')
    expect(switchLang('/fr/section/AA/233C?vin=X', 'de')).toBe('/de/section/AA/233C?vin=X')
    expect(switchLang('/es', 'en')).toBe('/')
    expect(switchLang('/de?x=1', 'es')).toBe('/es?x=1')
    expect(switchLang('/', 'fr')).toBe('/fr')
    expect(switchLang('/french-fries', 'fr')).toBe('/fr/french-fries')
    expect(switchLang('/design', 'en')).toBe('/design')
    expect(switchLang('/fr/list', 'fr')).toBe('/fr/list')
  })
})
