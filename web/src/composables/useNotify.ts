import { ref } from 'vue'

const message = ref('')
const visible = ref(false)

/** App-wide snackbar (rendered once in App.vue). */
export function useNotify() {
  return {
    message,
    visible,
    notify(text: string) {
      message.value = text
      visible.value = true
    },
  }
}
