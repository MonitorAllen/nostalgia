import { describe, expect, test } from 'bun:test'
import { getAdminUploadErrorMessage, validateAdminImageFile } from './uploadPolicy'

describe('validateAdminImageFile', () => {
  test('requires a file', () => {
    expect(validateAdminImageFile()).toBe('请选择要上传的图片')
  })

  test('rejects unsupported image types', () => {
    expect(validateAdminImageFile({ type: 'image/webp', size: 1024 })).toBe('仅支持 JPG 或 PNG 图片')
  })

  test('rejects oversized images', () => {
    expect(validateAdminImageFile({ type: 'image/png', size: 5 * 1024 * 1024 + 1 })).toBe(
      '图片不能超过 5 MB'
    )
  })

  test('accepts jpg and png images within the size limit', () => {
    expect(validateAdminImageFile({ type: 'image/jpeg', size: 1024 })).toBe('')
    expect(validateAdminImageFile({ type: 'image/png', size: 1024 })).toBe('')
  })
})

describe('getAdminUploadErrorMessage', () => {
  test('uses backend error text when available', () => {
    expect(getAdminUploadErrorMessage({ response: { data: { error: '不支持的文件类型' } } })).toBe(
      '不支持的文件类型'
    )
  })

  test('uses abort message for canceled uploads', () => {
    expect(getAdminUploadErrorMessage({ name: 'AbortError' })).toBe('上传已取消')
    expect(getAdminUploadErrorMessage({ name: 'CanceledError' })).toBe('上传已取消')
  })
})
