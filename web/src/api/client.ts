export class ApiError extends Error {
  constructor(
    readonly status: number,
    readonly code: string,
    message: string,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

export function apiPath(
  path: string,
  params: Record<string, string | number | undefined> = {},
): string {
  const q = new URLSearchParams()
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== '') q.set(k, String(v))
  }
  const s = q.toString()
  return s === '' ? path : `${path}?${s}`
}

interface ErrorBody {
  error?: string
  message?: string
}

export async function getJSON<T>(url: string, signal?: AbortSignal): Promise<T> {
  const res = await fetch(url, { signal: signal ?? null, headers: { Accept: 'application/json' } })
  if (!res.ok) {
    let body: ErrorBody = {}
    try {
      body = (await res.json()) as ErrorBody
    } catch {
      // not a JSON error body
    }
    throw new ApiError(res.status, body.error ?? 'http_error', body.message ?? res.statusText)
  }
  return (await res.json()) as T
}
