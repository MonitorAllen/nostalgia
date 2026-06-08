import { computed, onMounted, onUnmounted, ref, watch } from 'vue'

export type ThemeMode = 'system' | 'light' | 'dark'
export type ResolvedTheme = 'light' | 'dark'

const STORAGE_KEY = 'nostalgia-theme-mode'
const modes: ThemeMode[] = ['system', 'light', 'dark']

const mode = ref<ThemeMode>('system')
const systemTheme = ref<ResolvedTheme>('light')
const initialized = ref(false)
let mediaQuery: MediaQueryList | null = null

const resolvedTheme = computed<ResolvedTheme>(() => {
  return mode.value === 'system' ? systemTheme.value : mode.value
})

function applyTheme(theme: ResolvedTheme) {
  document.documentElement.dataset.theme = theme
  document.documentElement.classList.toggle('dark', theme === 'dark')
}

function readStoredMode(): ThemeMode {
  const stored = window.localStorage.getItem(STORAGE_KEY)
  return modes.includes(stored as ThemeMode) ? (stored as ThemeMode) : 'system'
}

function updateSystemTheme() {
  if (!mediaQuery) return
  systemTheme.value = mediaQuery.matches ? 'dark' : 'light'
}

export function initTheme() {
  if (typeof window === 'undefined' || initialized.value) return

  mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
  updateSystemTheme()
  mode.value = readStoredMode()
  applyTheme(resolvedTheme.value)

  mediaQuery.addEventListener('change', updateSystemTheme)
  initialized.value = true
}

export function useTheme() {
  const setMode = (nextMode: ThemeMode) => {
    mode.value = nextMode
    window.localStorage.setItem(STORAGE_KEY, nextMode)
  }

  onMounted(() => {
    initTheme()
  })

  onUnmounted(() => {
    // The listener intentionally stays alive for the app lifetime once initialized.
  })

  watch(resolvedTheme, (theme) => {
    if (typeof document !== 'undefined') {
      applyTheme(theme)
    }
  })

  return {
    mode,
    modes,
    resolvedTheme,
    setMode,
  }
}
