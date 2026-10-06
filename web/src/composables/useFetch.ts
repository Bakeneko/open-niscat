import { onScopeDispose, ref, shallowRef, watch, type Ref, type ShallowRef } from 'vue'
import { getJSON } from '@/api/client'

export interface Fetched<T> {
  data: ShallowRef<T | null>
  error: ShallowRef<Error | null>
  loading: Ref<boolean>
  reload(): void
}

/** Fetches JSON whenever url() changes; aborts superseded requests; null url = idle. */
export function useFetch<T>(url: () => string | null): Fetched<T> {
  const data = shallowRef<T | null>(null)
  const error = shallowRef<Error | null>(null)
  const loading = ref(false)
  let ctrl: AbortController | null = null

  async function load(u: string): Promise<void> {
    ctrl?.abort()
    const c = new AbortController()
    ctrl = c
    loading.value = true
    error.value = null
    try {
      const v = await getJSON<T>(u, c.signal)
      if (!c.signal.aborted) data.value = v
    } catch (e) {
      if (!c.signal.aborted) {
        data.value = null
        error.value = e instanceof Error ? e : new Error(String(e))
      }
    } finally {
      if (ctrl === c) loading.value = false
    }
  }

  function run(u: string | null) {
    if (u === null) {
      ctrl?.abort()
      data.value = null
      error.value = null
      loading.value = false
      return
    }
    void load(u)
  }

  watch(url, run, { immediate: true })
  onScopeDispose(() => ctrl?.abort())
  return {
    data,
    error,
    loading,
    reload() {
      run(url())
    },
  }
}
