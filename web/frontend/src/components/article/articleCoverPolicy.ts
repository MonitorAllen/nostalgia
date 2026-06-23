export const ARTICLE_COVER_ASPECT_RATIO = 16 / 9
export const ARTICLE_COVER_RATIO_TOLERANCE = 0.12
export const ARTICLE_COVER_RECOMMENDED_WIDTH = 1600
export const ARTICLE_COVER_RECOMMENDED_HEIGHT = 900
export const ARTICLE_COVER_HIGH_RES_WIDTH = 1920
export const ARTICLE_COVER_HIGH_RES_HEIGHT = 1080
export const ARTICLE_COVER_MIN_WIDTH = 1200
export const ARTICLE_COVER_MIN_HEIGHT = 675

export type ArticleCoverInspectionStatus = 'ok' | 'warning'

export interface ArticleCoverDimensions {
  width: number
  height: number
}

export interface ArticleCoverInspection {
  width: number
  height: number
  ratio: number
  status: ArticleCoverInspectionStatus
  warnings: string[]
}

const LOW_RESOLUTION_WARNING =
  '建议封面至少为 1200x675，当前图片可能在高清屏上显得模糊。'
const OFF_RATIO_WARNING = '当前图片比例偏离 16:9，详情页、列表或分享预览中可能出现明显裁切。'
const UNREADABLE_WARNING = '无法读取图片尺寸，仍可继续保存。'

export function inspectArticleCoverDimensions(
  dimensions: ArticleCoverDimensions | null
): ArticleCoverInspection {
  if (!dimensions || dimensions.width <= 0 || dimensions.height <= 0) {
    return {
      width: 0,
      height: 0,
      ratio: 0,
      status: 'warning',
      warnings: [UNREADABLE_WARNING]
    }
  }

  const ratio = dimensions.width / dimensions.height
  const warnings: string[] = []

  if (dimensions.width < ARTICLE_COVER_MIN_WIDTH || dimensions.height < ARTICLE_COVER_MIN_HEIGHT) {
    warnings.push(LOW_RESOLUTION_WARNING)
  }

  if (Math.abs(ratio - ARTICLE_COVER_ASPECT_RATIO) > ARTICLE_COVER_RATIO_TOLERANCE) {
    warnings.push(OFF_RATIO_WARNING)
  }

  return {
    width: dimensions.width,
    height: dimensions.height,
    ratio,
    status: warnings.length > 0 ? 'warning' : 'ok',
    warnings
  }
}

export function loadArticleCoverDimensions(src: string): Promise<ArticleCoverDimensions> {
  return new Promise((resolve, reject) => {
    if (!src) {
      reject(new Error(UNREADABLE_WARNING))
      return
    }

    const image = new Image()
    image.onload = () => resolve({ width: image.naturalWidth, height: image.naturalHeight })
    image.onerror = () => reject(new Error(UNREADABLE_WARNING))
    image.src = src
  })
}

export function loadArticleCoverFileDimensions(file: File): Promise<ArticleCoverDimensions> {
  return new Promise((resolve, reject) => {
    const objectUrl = URL.createObjectURL(file)

    loadArticleCoverDimensions(objectUrl)
      .then(resolve)
      .catch(reject)
      .finally(() => URL.revokeObjectURL(objectUrl))
  })
}
