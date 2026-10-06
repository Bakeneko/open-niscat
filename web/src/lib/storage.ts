export function load<T>(key: string, guard: (v: unknown) => v is T): T | undefined {
  try {
    const raw = localStorage.getItem(key)
    if (raw === null) return undefined
    const v: unknown = JSON.parse(raw)
    return guard(v) ? v : undefined
  } catch {
    return undefined
  }
}

/** Writes value unless storage already holds the same JSON (avoids storage-event loops between tabs). */
export function saveIfChanged(key: string, value: unknown): boolean {
  try {
    if (value === undefined) {
      if (localStorage.getItem(key) === null) return false
      localStorage.removeItem(key)
      return true
    }
    const raw = JSON.stringify(value)
    if (localStorage.getItem(key) === raw) return false
    localStorage.setItem(key, raw)
    return true
  } catch {
    return false
  }
}
