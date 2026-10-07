import { effectScope, ref, watch, type Ref } from 'vue'
import { load, saveIfChanged } from '@/lib/storage'

const instances = new Map<string, Ref<unknown>>()

/** A ref persisted in localStorage, shared by every caller and kept in sync with other tabs. */
export function useStored<T>(
  key: string,
  guard: (v: unknown) => v is T,
  fallback: () => T,
): Ref<T> {
  const existing = instances.get(key)
  if (existing !== undefined) return existing as Ref<T>
  const state = ref(load(key, guard) ?? fallback()) as Ref<T>
  instances.set(key, state)
  // Another tab changed the key (e.key === null: it cleared localStorage).
  window.addEventListener('storage', (e) => {
    if (e.key === key || e.key === null) state.value = load(key, guard) ?? fallback()
  })
  // Detached scope: the watcher must outlive the component that happened to call first.
  effectScope(true).run(() => {
    watch(
      state,
      (v) => {
        saveIfChanged(key, v)
      },
      { deep: true },
    )
  })
  return state
}
