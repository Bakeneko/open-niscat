import { afterEach, describe, expect, it, vi } from 'vitest'
import { effectScope, nextTick, ref } from 'vue'
import { useFetch } from '../useFetch'

afterEach(() => {
  vi.unstubAllGlobals()
})

const flush = async () => {
  await new Promise((r) => setTimeout(r, 0))
  await nextTick()
}

describe('useFetch', () => {
  it('loads, reports errors, and ignores stale responses', async () => {
    const calls: string[] = []
    vi.stubGlobal(
      'fetch',
      vi.fn((url: string) => {
        calls.push(url)
        if (url === '/bad') return Promise.resolve(new Response('{"error":"not_found","message":"nope"}', { status: 404 }))
        return Promise.resolve(new Response(JSON.stringify({ url }), { status: 200 }))
      }),
    )
    const url = ref<string | null>('/a')
    const scope = effectScope()
    const r = scope.run(() => useFetch<{ url: string }>(() => url.value))
    if (r === undefined) throw new Error('no scope')
    await flush()
    expect(r.data.value).toEqual({ url: '/a' })
    url.value = '/bad'
    await flush()
    expect(r.data.value).toBeNull()
    expect(r.error.value?.message).toBe('nope')
    url.value = null
    await flush()
    expect(r.error.value).toBeNull()
    expect(calls).toEqual(['/a', '/bad'])
    scope.stop()
  })
})
