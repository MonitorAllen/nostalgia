import { describe, expect, test } from 'bun:test'
import { CODE_BLOCK_LANGUAGES } from './contentLanguages'

describe('CODE_BLOCK_LANGUAGES', () => {
  test('keeps article and comment editors aligned with expected languages', () => {
    const languages = CODE_BLOCK_LANGUAGES.map((item) => item.language)

    expect(languages).toEqual([
      'plaintext',
      'go',
      'python',
      'javascript',
      'typescript',
      'java',
      'c',
      'cpp',
      'sql',
      'json',
      'bash',
      'html',
      'css',
    ])
    expect(new Set(languages).size).toBe(languages.length)
  })
})
