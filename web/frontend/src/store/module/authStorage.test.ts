import { describe, expect, test } from 'bun:test'
import {
  AUTH_STORAGE_KEYS,
  LEGACY_AUTH_STORAGE_KEYS,
  cleanupLegacyAuthStorage,
} from './authStorage'

class MemoryStorage implements Pick<Storage, 'clear' | 'getItem' | 'key' | 'length' | 'removeItem' | 'setItem'> {
  private values = new Map<string, string>()

  get length() {
    return this.values.size
  }

  clear() {
    this.values.clear()
  }

  getItem(key: string) {
    return this.values.get(key) ?? null
  }

  key(index: number) {
    return Array.from(this.values.keys())[index] ?? null
  }

  removeItem(key: string) {
    this.values.delete(key)
  }

  setItem(key: string, value: string) {
    this.values.set(key, value)
  }
}

describe('cleanupLegacyAuthStorage', () => {
  test('removes legacy auth keys without removing current auth or unrelated keys', () => {
    const storage = new MemoryStorage()

    Object.values(AUTH_STORAGE_KEYS).forEach((key) => storage.setItem(key, `current:${key}`))
    LEGACY_AUTH_STORAGE_KEYS.forEach((key) => storage.setItem(key, `legacy:${key}`))
    storage.setItem('nostalgia-theme-mode', 'system')

    cleanupLegacyAuthStorage(storage)

    Object.values(AUTH_STORAGE_KEYS).forEach((key) => {
      expect(storage.getItem(key)).toBe(`current:${key}`)
    })
    LEGACY_AUTH_STORAGE_KEYS.forEach((key) => {
      expect(storage.getItem(key)).toBeNull()
    })
    expect(storage.getItem('nostalgia-theme-mode')).toBe('system')
  })
})
