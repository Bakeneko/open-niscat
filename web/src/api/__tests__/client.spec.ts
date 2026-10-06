import { afterEach, describe, expect, it, vi } from 'vitest'
import { ApiError, apiPath, getJSON } from '../client'

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('apiPath', () => {
  it('drops empty params and encodes the rest', () => {
    expect(apiPath('/api/search', { q: 'joint & co', type: 'parts', vin: undefined, offset: 0, lang: '' })).toBe(
      '/api/search?q=joint+%26+co&type=parts&offset=0',
    )
    expect(apiPath('/api/catalogs')).toBe('/api/catalogs')
  })
})

describe('getJSON', () => {
  it('returns the body', async () => {
    vi.stubGlobal('fetch', vi.fn(() => Promise.resolve(new Response('{"a":1}', { status: 200 }))))
    await expect(getJSON<{ a: number }>('/x')).resolves.toEqual({ a: 1 })
  })
  it('throws ApiError with the API error code', async () => {
    vi.stubGlobal('fetch', vi.fn(() => Promise.resolve(new Response('{"error":"not_found","message":"not found: VIN X"}', { status: 404 }))))
    await expect(getJSON('/x')).rejects.toMatchObject({ status: 404, code: 'not_found', message: 'not found: VIN X' })
  })
  it('copes with a non-JSON error body', async () => {
    vi.stubGlobal('fetch', vi.fn(() => Promise.resolve(new Response('oops', { status: 503, statusText: 'Service Unavailable' }))))
    const err = await getJSON('/x').catch((e: unknown) => e)
    expect(err).toBeInstanceOf(ApiError)
    expect(err).toMatchObject({ status: 503, code: 'http_error' })
  })
})
