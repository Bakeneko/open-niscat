import { getJSON } from '@/api/client'
import type { Meta } from '@/api/types'

// Shared by the whole app; a failed request is forgotten so the next call retries.
let meta: Promise<Meta> | null = null

export function useMeta(): Promise<Meta> {
  meta ??= getJSON<Meta>('/api/meta').catch((e: unknown) => {
    meta = null
    throw e
  })
  return meta
}
