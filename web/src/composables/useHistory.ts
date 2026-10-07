import { isHistory, pushHistory, type HistoryEntry } from '@/lib/history'
import { useStored } from './useStored'

export function useHistory() {
  const entries = useStored<HistoryEntry[]>('open-niscat.history', isHistory, () => [])
  return {
    entries,
    push: (kind: HistoryEntry['kind'], label: string, path: string) => {
      entries.value = pushHistory(entries.value, { kind, label, path, at: Date.now() })
    },
    clear: () => {
      entries.value = []
    },
  }
}
