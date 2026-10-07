import { describe, expect, it } from 'vitest'
import { effectScope, nextTick, ref } from 'vue'
import { usePageTitle } from '../usePageTitle'

describe('usePageTitle', () => {
  it('follows the page, keeps the previous title while loading, falls back to the app name', async () => {
    document.title = 'before'
    const page = ref<string | null | undefined>(undefined)
    const scope = effectScope()
    scope.run(() => {
      usePageTitle(() => page.value)
    })
    expect(document.title).toBe('before')
    page.value = 'Catalogs'
    await nextTick()
    expect(document.title).toBe('Catalogs · Open Niscat')
    page.value = undefined
    await nextTick()
    expect(document.title).toBe('Catalogs · Open Niscat')
    page.value = null
    await nextTick()
    expect(document.title).toBe('Open Niscat')
    scope.stop()
  })
})
