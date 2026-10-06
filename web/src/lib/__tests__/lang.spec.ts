import { describe, expect, it } from 'vitest'
import { isLang, langFromParam, localizedPath, switchLang } from '../lang'

describe('lang', () => {
  it('parses the route param', () => {
    expect(langFromParam('fr')).toBe('fr')
    expect(langFromParam(undefined)).toBe('en')
    expect(langFromParam(['fr'])).toBe('en')
    expect(isLang('de')).toBe(false)
  })
  it('localizes paths', () => {
    expect(localizedPath('en', '/cart')).toBe('/cart')
    expect(localizedPath('fr', '/cart')).toBe('/fr/cart')
    expect(localizedPath('fr', '/')).toBe('/fr')
    expect(localizedPath('fr', 'cart')).toBe('/fr/cart')
  })
  it('switches language keeping path, query and hash', () => {
    expect(switchLang('/section/AA/233C?vin=X&item=4&tab=info', 'fr')).toBe(
      '/fr/section/AA/233C?vin=X&item=4&tab=info',
    )
    expect(switchLang('/fr/section/AA/233C?vin=X#top', 'en')).toBe('/section/AA/233C?vin=X#top')
    expect(switchLang('/fr', 'en')).toBe('/')
    expect(switchLang('/fr?x=1', 'en')).toBe('/?x=1')
    expect(switchLang('/', 'fr')).toBe('/fr')
    expect(switchLang('/french-fries', 'fr')).toBe('/fr/french-fries')
    expect(switchLang('/fr/cart', 'fr')).toBe('/fr/cart')
  })
})
