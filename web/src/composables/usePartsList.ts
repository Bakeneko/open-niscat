import { computed } from 'vue'
import {
  addItem,
  isPartsList,
  mergeItems,
  removeItem,
  setQty,
  type ListItem,
} from '@/lib/partsList'
import type { Scope } from '@/lib/scope'
import { useStored } from './useStored'

export function usePartsList() {
  const items = useStored<ListItem[]>('open-niscat.list', isPartsList, () => [])
  return {
    items,
    count: computed(() => items.value.length),
    add: (id: string, qty = 1, scope: Scope | null = null) => {
      items.value = addItem(items.value, id, qty, scope ?? undefined)
    },
    setQty: (id: string, qty: number) => {
      items.value = setQty(items.value, id, qty)
    },
    remove: (id: string) => {
      items.value = removeItem(items.value, id)
    },
    replace: (next: ListItem[]) => {
      items.value = [...next]
    },
    merge: (more: ListItem[]) => {
      items.value = mergeItems(items.value, more)
    },
    clear: () => {
      items.value = []
    },
  }
}
