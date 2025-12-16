// src/utils/validate.ts

/**
 * 检查字符串是否为合法的 UUID
 * @param str 待检查的字符串
 * @returns boolean
 */
export const isUUID = (str: string): boolean => {
    const uuidRegex = /^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$/;
    return uuidRegex.test(str);
}