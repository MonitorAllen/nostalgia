import moment from 'moment'

/**
 * 通用时间格式化函数（基于 moment.js）
 * @param {string|number} value - 时间值，支持：
 *   - 秒级时间戳（10位数字）
 *   - 毫秒级时间戳（13位数字）
 *   - 符合 ISO 8601 的字符串，如 '2023-01-01T12:00:00Z'
 *   - 其他 moment 可解析的日期字符串
 * @param {string} formatStr - 输出格式，默认 'YYYY-MM-DD HH:mm:ss'
 * @param {string} [invalidReturn=''] - 无效值时返回的内容，默认空字符串
 * @returns {string} 格式化后的时间字符串
 */
export function formatTime(value, formatStr = 'YYYY-MM-DD HH:mm:ss', invalidReturn = '') {
  // 处理空值
  if (value === null || value === undefined || value === '') {
    return invalidReturn
  }

  // 特殊处理 UTC 零值字符串
  if (value === '0001-01-01T00:00:00Z') {
    return '无'
  }

  let momentObj

  // 数字类型：自动判断是秒(10位)还是毫秒(13位)
  if (typeof value === 'number') {
    const len = value.toString().length
    if (len === 10) {
      // 秒级时间戳，转换为毫秒
      momentObj = moment(value * 1000)
    } else if (len === 13) {
      // 毫秒级时间戳
      momentObj = moment(value)
    } else {
      // 其他位数，直接尝试解析
      momentObj = moment(value)
    }
  } else {
    // 字符串：直接交给 moment 解析
    momentObj = moment(value)
  }

  // 校验有效性
  if (!momentObj.isValid()) {
    return invalidReturn
  }

  return momentObj.format(formatStr)
}
