import { ref } from 'vue'

export type ToastSeverity = 'success' | 'info' | 'warning' | 'error'

export interface ToastOptions {
  severity?: ToastSeverity
  summary: string
  detail?: string
  life?: number
}

export interface ToastItem extends Required<Omit<ToastOptions, 'detail'>> {
  id: number
  detail: string
}

const toasts = ref<ToastItem[]>([])
let nextId = 1

export function useToast() {
  const remove = (id: number) => {
    toasts.value = toasts.value.filter((toast) => toast.id !== id)
  }

  const add = (options: ToastOptions) => {
    const toast: ToastItem = {
      id: nextId++,
      severity: options.severity ?? 'info',
      summary: options.summary,
      detail: options.detail ?? '',
      life: options.life ?? 3000,
    }
    toasts.value = [...toasts.value, toast]
    if (toast.life > 0) {
      window.setTimeout(() => remove(toast.id), toast.life)
    }
  }

  return {
    toasts,
    add,
    remove,
  }
}
