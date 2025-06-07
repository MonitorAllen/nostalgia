import axiosInstance from "@/config/axios";

export default class MyUploadAdapter {
    private article_id: string
    private loader: any

    constructor(loader: any, article_id: string) {
        this.loader = loader
        this.article_id = article_id
    }

    async upload() {
        const file = this.loader.file
        const data = await uploadFile(file, this.article_id)

        return { default: data.url }
    }


}

async function uploadFile(inputFile: Promise<File>, article_id: string = '') {
    const file = await inputFile
    const arrayBuffer = await file.arrayBuffer()
    const bytes = new Uint8Array(arrayBuffer)
    const binaryString = Array.from(bytes).map((b) => String.fromCharCode(b)).join('')
    const base64Content = btoa(binaryString)

    const params = {
        article_id: article_id,
        content: base64Content
    }

    const res = await axiosInstance.post('/util/upload_file', params)
    return res.data
}