<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ClassicEditor } from 'ckeditor5'
import { Ckeditor } from '@ckeditor/ckeditor5-vue'
import 'ckeditor5/ckeditor5.css'

// PrimeVue Components
import Textarea from 'primevue/textarea'
import InputText from 'primevue/inputtext'
import ToggleButton from 'primevue/togglebutton'
import Button from 'primevue/button'
import Select from 'primevue/select'
import Card from 'primevue/card'
import Skeleton from 'primevue/skeleton'
import { useToast } from 'primevue/usetoast'
import FileUpload from "primevue/fileupload";

// Utils & Config
import { editorConfig } from '@/config/editorConfig'
import MyUploadAdapter from '@/util/uploadAdapter'
import { createArticle, fetchArticleById, updateArticle, type UpdateArticleRequest } from '@/api/articles'
import { listAllCategories } from "@/api/category"
import type { Article } from '@/types/article' // 假设你有这个类型定义
import type { Category } from "@/types/category"
import axiosInstance from "@/config/axios.ts";

const props = defineProps({
  id: { type: String, default: '' }
})

const router = useRouter()
const toast = useToast()

// Data
const loading = ref(true) // 全局加载状态
const isLayoutReady = ref(false)
const editorData = ref('')
const article = ref<Article>({
  id: '',
  title: '',
  summary: '',
  content: '',
  likes: 0,
  views: 0,
  is_publish: false,
  created_at: '',
  updated_at: '',
  owner: '',
  category_id: 0, // 注意类型匹配
  category_name: '',
  cover: '',
  slug: '',
  check_outdated: false
})
// 绑定 is_publish (boolean)
const isPublish = ref(false)
const checkOutdated = ref(false)
const categories = ref<Category[]>([])

// 初始化数据
const initData = async () => {
  loading.value = true
  try {
    // 1. 并行获取分类
    const catResp = await listAllCategories()
    categories.value = catResp.data.categories

    // 2. 处理文章逻辑
    if (props.id === '') {
      // 新建模式：先创建一个草稿
      const createResp = await createArticle({ title: '无标题文章', summary: '', is_publish: false })
      article.value = createResp.data.article
      // 替换路由 ID，防止刷新丢失
      await router.replace(`/article/edit/${article.value.id}`)
    } else {
      // 编辑模式
      const fetchResp = await fetchArticleById({ id: props.id, needContent: true })
      article.value = fetchResp.data.article
      editorData.value = article.value.content || ''
      isPublish.value = article.value.is_publish
      checkOutdated.value = article.value.check_outdated
    }

    isLayoutReady.value = true
  } catch (error: any) {
    toast.add({ severity: 'error', summary: '初始化失败', detail: error.response?.data?.error || '网络错误', life: 3000 })
  } finally {
    loading.value = false
  }
}

// 保存逻辑
const isSaving = ref(false)
const save = async () => {
  if (!article.value.id) return

  isSaving.value = true
  try {
    const req: UpdateArticleRequest = {
      id: article.value.id,
      title: article.value.title,
      summary: article.value.summary,
      content: editorData.value,
      is_publish: isPublish.value,
      category_id: Number(article.value.category_id),
      cover: article.value.cover,
      slug: article.value.slug,
      check_outdated: article.value.check_outdated,
    }

    await updateArticle(req)
    toast.add({ severity: 'success', summary: '成功', detail: '文章更新成功', life: 3000 })
  } catch (error: any) {
    toast.add({ severity: 'error', summary: '失败', detail: error.response?.data?.error, life: 3000 })
  } finally {
    isSaving.value = false
  }
}

const customBase64Uploader = async (event: any) => {
  const file = event.files[0];

  // 复用 Adapter 的逻辑或者手动调用
  // 这里手动演示
  const reader = new FileReader();
  reader.readAsDataURL(file);
  reader.onload = async () => {
    const base64 = (reader.result as string).split(',')[1];

    try {
      const res = await axiosInstance.post('/util/upload_file', {
        article_id: article.value.id,
        content: base64,
        type: 'cover' // 👈 标记为封面
      });

      // 更新封面预览 (加时间戳防缓存)
      article.value.cover = `${res.data.url}?t=${Date.now()}`;
      toast.add({ severity: 'success', summary: '成功', detail: '封面已更新', life: 3000 });
    } catch (e) {
      toast.add({ severity: 'error', summary: '失败', detail: '上传出错', life: 3000 });
    }
  };
};

const removeCover = () => {
  article.value.cover = '';
}

// CKEditor 配置
const finalConfig = computed(() => ({
  ...editorConfig,
  extraPlugins: [CustomUploadAdapterPlugin]
}))

// 自定义上传插件
function CustomUploadAdapterPlugin(editor: any) {
  editor.plugins.get('FileRepository').createUploadAdapter = (loader: any) => {
    // 确保 article.value.id 存在，因为我们需要 ID 来上传图片
    return new MyUploadAdapter(loader, article.value.id, 'content')
  }
}

const onEditorReady = (editorInstance: ClassicEditor) => {
  // 可以在这里做一些编辑器初始化后的操作
}

onMounted(() => {
  initData()
})
</script>

<template>
  <div class="w-full px-3 md:px-5 py-5 max-w-8xl mx-auto">
    <div v-if="loading" class="grid formgrid p-fluid gap-4">
      <div class="col-12 lg:col-9">
        <Skeleton height="600px" class="border-round"></Skeleton>
      </div>
      <div class="col-12 lg:col-3 flex flex-column gap-3">
        <Skeleton height="300px" class="border-round"></Skeleton>
        <Skeleton height="50px" class="border-round"></Skeleton>
      </div>
    </div>

    <div v-else class="grid formgrid p-fluid relative">

      <div class="col-12 lg:col-9 pb-4">
        <div class="surface-ground border-round-md p-4 h-full flex flex-column align-items-center shadow-1">
          <div class="paper-container shadow-2 w-full">
            <ckeditor
              v-if="isLayoutReady"
              v-model="editorData"
              :editor="ClassicEditor"
              :config="finalConfig"
              @ready="onEditorReady"
            />
          </div>
        </div>
      </div>

      <div class="col-12 lg:col-3">
        <div class="sticky top-0" style="top: 1rem; z-index: 10;">
          <Card class="shadow-1 border-none">
            <template #title>
              <div class="text-xl font-bold mb-2">文章设置</div>
            </template>
            <template #content>
              <div class="flex flex-column gap-4">
                <div class="field mb-0">
                  <label for="title" class="font-medium text-900">标题</label>
                  <InputText id="title" v-model="article.title" class="w-full" placeholder="输入文章标题" />
                </div>
                <div class="field mb-0">
                  <label for="summary" class="font-medium text-900">摘要</label>
                  <Textarea
                    id="summary"
                    v-model="article.summary"
                    rows="4"
                    autoResize
                    class="w-full"
                    placeholder="简短的介绍..."
                  />
                </div>
                <div class="field mb-0">
                  <label for="title" class="font-medium text-900">短标识</label>
                  <InputText id="title" v-model="article.slug" class="w-full" placeholder="输入标识" />
                </div>
                <div class="field mb-0">
                  <label for="category" class="font-medium text-900">分类</label>
                  <Select
                    id="category"
                    v-model="article.category_id"
                    :options="categories"
                    optionLabel="name"
                    optionValue="id"
                    placeholder="选择分类"
                    class="w-full"
                  />
                </div>
                <div class="field mb-0">
                  <label for="isPublish" class="font-medium text-900">发布状态</label>
                  <ToggleButton
                    id="isPublish"
                    v-model="isPublish"
                    onLabel="发布"
                    offLabel="草稿"
                    onIcon="pi pi-check-circle"
                    offIcon="pi pi-pencil"
                    class="w-full"
                  />
                </div>
                <div class="field mb-0">
                  <label for="checkOutdated" class="font-medium text-900">检查时效</label>
                  <ToggleButton
                    id="checkOutdated"
                    v-model="checkOutdated"
                    onLabel="检查"
                    offLabel="不检查"
                    onIcon="pi pi-check-circle"
                    offIcon="pi pi-times-circle"
                    class="w-full"
                  />
                </div>
                <div class="field mb-0">
                  <label class="font-medium text-900 block mb-2">文章封面</label>
                  <div v-if="article.cover" class="relative mb-2 w-full border-round overflow-hidden" style="height: 150px;">
                    <img :src="article.cover" class="w-full h-full object-cover"  alt=""/>
                    <Button icon="pi pi-times" rounded severity="danger" class="absolute top-0 right-0 m-1 h-2rem w-2rem" @click="removeCover" />
                  </div>
                  <FileUpload
                    mode="basic"
                    name="file"
                    accept="image/*"
                    :maxFileSize="2000000"
                    :auto="true"
                    customUpload
                    @uploader="customBase64Uploader"
                    chooseLabel="上传封面"
                    class="w-full p-button-outlined"
                  />
                </div>
                <div class="pt-2">
                  <Button
                    label="保存文章"
                    icon="pi pi-save"
                    severity="success"
                    class="w-full"
                    :loading="isSaving"
                    @click="save"
                  />
                </div>
              </div>
            </template>
          </Card>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 1. 模拟 A4 纸或阅读器宽度 */
.paper-container {
  max-width: 850px; /* 限制最大宽度，与前台阅读宽度保持一致 */
  background: white;
  min-height: 800px;
}

/* 2. 让 CKEditor 的编辑区域本身没有边框，融入 paper-container */
:deep(.ck-editor__editable) {
  min-height: 800px;
  padding: 2rem 3rem !important; /* 增加内边距，模拟文档页边距 */
  border: none !important;
  box-shadow: none !important;
}

/* 3. 去掉 Toolbar 的圆角和边框，让它看起来像吸附在顶部的工具栏 */
:deep(.ck.ck-toolbar) {
  border: none !important;
  border-bottom: 1px solid #e5e7eb !important;
  background: #f8f9fa; /* 稍微灰一点的工具栏背景 */
}

/* 4. 移动端适配：取消固定宽度，占满全屏 */
@media screen and (max-width: 768px) {
  .paper-container {
    max-width: 100%;
    box-shadow: none !important;
  }

  :deep(.ck-editor__editable) {
    padding: 1rem !important;
  }
}
</style>
