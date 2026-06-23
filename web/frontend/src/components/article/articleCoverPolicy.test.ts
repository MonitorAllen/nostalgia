import { describe, expect, test } from 'bun:test'
import {
  ARTICLE_COVER_ASPECT_RATIO,
  ARTICLE_COVER_HIGH_RES_HEIGHT,
  ARTICLE_COVER_HIGH_RES_WIDTH,
  ARTICLE_COVER_MIN_HEIGHT,
  ARTICLE_COVER_MIN_WIDTH,
  ARTICLE_COVER_RATIO_TOLERANCE,
  ARTICLE_COVER_RECOMMENDED_HEIGHT,
  ARTICLE_COVER_RECOMMENDED_WIDTH,
  inspectArticleCoverDimensions
} from './articleCoverPolicy'

describe('article cover policy', () => {
  test('defines the canonical 16:9 cover standard', () => {
    expect(ARTICLE_COVER_ASPECT_RATIO).toBe(16 / 9)
    expect(ARTICLE_COVER_RATIO_TOLERANCE).toBe(0.12)
    expect(ARTICLE_COVER_RECOMMENDED_WIDTH).toBe(1600)
    expect(ARTICLE_COVER_RECOMMENDED_HEIGHT).toBe(900)
    expect(ARTICLE_COVER_HIGH_RES_WIDTH).toBe(1920)
    expect(ARTICLE_COVER_HIGH_RES_HEIGHT).toBe(1080)
    expect(ARTICLE_COVER_MIN_WIDTH).toBe(1200)
    expect(ARTICLE_COVER_MIN_HEIGHT).toBe(675)
  })

  test('accepts recommended 16:9 cover dimensions without warnings', () => {
    expect(inspectArticleCoverDimensions({ width: 1600, height: 900 })).toEqual({
      width: 1600,
      height: 900,
      ratio: 16 / 9,
      status: 'ok',
      warnings: []
    })
  })

  test('warns when the image is below minimum recommended dimensions', () => {
    const result = inspectArticleCoverDimensions({ width: 900, height: 506 })

    expect(result.status).toBe('warning')
    expect(result.warnings).toContain('建议封面至少为 1200x675，当前图片可能在高清屏上显得模糊。')
  })

  test('warns when the image ratio is far from 16:9', () => {
    const result = inspectArticleCoverDimensions({ width: 1200, height: 1200 })

    expect(result.status).toBe('warning')
    expect(result.warnings).toContain('当前图片比例偏离 16:9，详情页、列表或分享预览中可能出现明显裁切。')
  })

  test('warns when dimensions cannot be read', () => {
    expect(inspectArticleCoverDimensions(null)).toEqual({
      width: 0,
      height: 0,
      ratio: 0,
      status: 'warning',
      warnings: ['无法读取图片尺寸，仍可继续保存。']
    })
  })
})
