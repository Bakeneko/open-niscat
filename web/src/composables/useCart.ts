import { computed } from 'vue'
import { addItem, isCart, mergeItems, removeItem, setQty, type CartItem } from '@/lib/cart'
import type { Scope } from '@/lib/scope'
import { useStored } from './useStored'

export function useCart() {
  const items = useStored<CartItem[]>('open-niscat.cart', isCart, () => [])
  return {
    items,
    count: computed(() => items.value.length),
    add(id: string, qty = 1, scope: Scope | null = null) {
      items.value = addItem(items.value, id, qty, scope ?? undefined)
    },
    setQty(id: string, qty: number) {
      items.value = setQty(items.value, id, qty)
    },
    remove(id: string) {
      items.value = removeItem(items.value, id)
    },
    replace(next: CartItem[]) {
      items.value = [...next]
    },
    merge(more: CartItem[]) {
      items.value = mergeItems(items.value, more)
    },
    clear() {
      items.value = []
    },
  }
}
