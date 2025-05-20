<template>
  <DefaultLayout>
    <div class="p-6">
      <h1 class="text-2xl font-bold mb-6">新建文章</h1>
      <div class="bg-white rounded-lg shadow p-6">
        <form @submit.prevent="submitForm" class="space-y-6">
          <div>
            <label class="block text-sm font-medium text-gray-700">标题</label>
            <input type="text" v-model="form.title" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500">
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">分类</label>
            <select v-model="form.category" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500">
              <option v-for="category in categories" :value="category">{{ category.name }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">状态</label>
            <select v-model="form.status" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500">
              <option v-for="status in statusOptions" :value="status.value">{{ status.label }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700">内容</label>
            <textarea v-model="form.content" rows="10" class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500"></textarea>
          </div>
          <div class="flex justify-end">
            <button type="button" class="bg-gray-200 text-gray-700 px-4 py-2 rounded-lg mr-4 hover:bg-gray-300" @click="goBack">取消</button>
            <button type="submit" class="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700">保存</button>
          </div>
        </form>
      </div>
    </div>
  </DefaultLayout>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import DefaultLayout from '@/layouts/DefaultLayout.vue'

const router = useRouter()

const form = ref({
  title: '',
  category: null,
  status: 'draft',
  content: ''
})

const categories = ref([
  { name: '前端开发', code: 'frontend' },
  { name: '后端开发', code: 'backend' },
  { name: '编程语言', code: 'language' },
  { name: '数据库', code: 'database' }
])

const statusOptions = ref([
  { label: '草稿', value: 'draft' },
  { label: '已发布', value: 'published' }
])

const submitForm = () => {
  // TODO: 实现发布文章的逻辑
  console.log('发布文章:', form.value)
  router.push('/articles')
}

const goBack = () => {
  router.push('/articles')
}
</script>

<style scoped>
.new-article {
  padding: 1rem;
}

.field {
  margin-bottom: 1.5rem;
}
</style> 