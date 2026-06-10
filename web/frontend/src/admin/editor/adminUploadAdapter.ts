import { uploadAdminFile } from '../api/adminUploadApi'
import { getAdminUploadErrorMessage, validateAdminImageFile } from './uploadPolicy'

export default class AdminUploadAdapter {
  private abortController = new AbortController()

  constructor(
    private loader: any,
    private articleId: string,
    private type: 'content' | 'cover' = 'content'
  ) {}

  async upload() {
    const file = (await this.loader.file) as File | null
    const validationMessage = validateAdminImageFile(file)

    if (validationMessage || !file) throw new Error(validationMessage || '请选择要上传的图片')

    this.loader.uploadTotal = file.size

    try {
      const content = await this.fileToBase64(file)
      const response = await uploadAdminFile(
        { article_id: this.articleId, content, type: this.type },
        this.abortController.signal
      )

      return { default: response.data.url, url: response.data.url }
    } catch (error) {
      throw new Error(getAdminUploadErrorMessage(error, '图片上传失败，请稍后再试'))
    }
  }

  abort() {
    this.abortController.abort()
  }

  private fileToBase64(file: File): Promise<string> {
    return new Promise((resolve, reject) => {
      const reader = new FileReader()
      reader.readAsDataURL(file)
      reader.onload = () => {
        this.loader.uploaded = file.size
        resolve(String(reader.result).split(',')[1] || '')
      }
      reader.onprogress = (event) => {
        if (event.lengthComputable) this.loader.uploaded = event.loaded
      }
      reader.onerror = () => reject(new Error('读取图片失败'))
    })
  }
}
