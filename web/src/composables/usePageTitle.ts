import { watchEffect } from 'vue'
import { pageTitle } from '@/lib/title'

/**
 * Keeps document.title in step with the view. The getter returns undefined while the page data is loading
 * (the previous title stays until then), null for the app name alone.
 */
export function usePageTitle(page: () => string | null | undefined): void {
  watchEffect(() => {
    const p = page()
    if (p !== undefined) document.title = pageTitle(p)
  })
}
