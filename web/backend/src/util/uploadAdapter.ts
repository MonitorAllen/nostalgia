import axiosInstance from "@/util/http"; // 确保引用正确的 http 工具

export default class MyUploadAdapter {
  private loader: any;
  private readonly article_id: string;
  private type: string

  constructor(loader: any, article_id: string, type:string = 'content') {
    this.loader = loader;
    this.article_id = article_id;
    this.type = type
  }

  // 开始上传
  async upload() {
    // 读取文件
    const file = await this.loader.file;
    // 转 Base64
    const base64Content = await this.fileToBase64(file);

    // 构造 gRPC 风格的 JSON 请求体
    // gRPC Gateway 会把 base64 字符串自动映射到 bytes 字段
    const payload = {
      article_id: this.article_id,
      content: base64Content, // 这里发的是纯 Base64 字符串（不带 data:image/png;base64, 前缀）
      type: this.type
    };

    try {
      const res = await axiosInstance.post('/util/upload_file', payload);
      // gRPC 返回 { url: "...", filename: "..." }
      // CKEditor 需要 { default: url }
      return { default: res.data.url, url: res.data.url };
    } catch (error) {
      throw error;
    }
  }

  private fileToBase64(file: File): Promise<string> {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.readAsDataURL(file);
      reader.onload = () => {
        const result = reader.result as string;
        // 去掉 Data URL 前缀 (例如 "data:image/png;base64,")
        // 因为 gRPC bytes 字段只需要纯数据
        const base64 = result.split(',')[1];
        resolve(base64);
      };
      reader.onerror = error => reject(error);
    });
  }

  // 终止上传
  abort() {
    // 如果需要支持取消上传，可以结合 axios 的 CancelToken
  }
}
