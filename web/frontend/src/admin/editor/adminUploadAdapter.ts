import { uploadAdminFile } from '../api/adminUploadApi'

export default class AdminUploadAdapter {
  private abortController = new AbortController()

  constructor(
    private loader: any,
    private articleId: string,
    private type: 'content' | 'cover' = 'content'
  ) {}

  async upload() {
    const file = (await this.loader.file) as File | null

    if (!file) throw new Error('请选择要上传的图片')
    if (!file.type.startsWith('image/')) throw new Error('只能上传图片文件')

    this.loader.uploadTotal = file.size

    const content = await this.fileToBase64(file)
    const response = await uploadAdminFile(
      { article_id: this.articleId, content, type: this.type },
      this.abortController.signal
    )

    return { default: response.data.url, url: response.data.url }
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
