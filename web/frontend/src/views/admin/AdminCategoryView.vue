<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { CalendarDays, FolderTree, Pencil, Plus, Save, Trash2, X } from '@lucide/vue'
import type { AdminCategory } from '@/admin/types'
import {
  createAdminCategory,
  deleteAdminCategory,
  listAllAdminCategories,
  updateAdminCategory
} from '@/admin/api/adminCategoryApi'
import AppBadge from '@/components/ui/AppBadge.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import { useToast } from '@/composables/useToast'

const toast = useToast()

const categories = ref<AdminCategory[]>([])
const loading = ref(false)
const editingId = ref<string | null>(null)
const draftName = ref('')
const newCategoryName = ref('')
const selectedCategory = ref<AdminCategory | null>(null)
const creating = ref(false)
const saving = ref(false)
const deleting = ref(false)

const fetchCategories = async () => {
  loading.value = true

  try {
    const response = await listAllAdminCategories()
    categories.value = response.data.categories ?? []
  } catch {
    categories.value = []
  } finally {
    loading.value = false
  }
}

const createCategory = async () => {
  const name = newCategoryName.value.trim()

  if (!name) {
    toast.add({
      severity: 'warning',
      summary: '分类名称不能为空',
      detail: '请输入一个清楚的分类名称',
      life: 2400
    })
    return
  }

  creating.value = true

  try {
    await createAdminCategory({ name })
    newCategoryName.value = ''
    await fetchCategories()
    toast.add({
      severity: 'success',
      summary: '分类已创建',
      detail: name,
      life: 2400
    })
  } catch {
    // Admin HTTP client already shows a toast for request failures.
  } finally {
    creating.value = false
  }
}

const startEdit = (category: AdminCategory) => {
  editingId.value = String(category.id)
  draftName.value = category.name
}

const cancelEdit = () => {
  editingId.value = null
  draftName.value = ''
}

const saveEdit = async (category: AdminCategory) => {
  const name = draftName.value.trim()

  if (!name) {
    toast.add({
      severity: 'warning',
      summary: '分类名称不能为空',
      detail: '请输入一个清楚的分类名称',
      life: 2400
    })
    return
  }

  saving.value = true

  try {
    await updateAdminCategory({ id: category.id, name })
    cancelEdit()
    await fetchCategories()
    toast.add({
      severity: 'success',
      summary: '分类已更新',
      detail: name,
      life: 2400
    })
  } catch {
    // Admin HTTP client already shows a toast for request failures.
  } finally {
    saving.value = false
  }
}

const askDelete = (category: AdminCategory) => {
  selectedCategory.value = category
}

const cancelDelete = () => {
  if (deleting.value) return
  selectedCategory.value = null
}

const confirmDelete = async () => {
  if (!selectedCategory.value || deleting.value) return

  const category = selectedCategory.value
  deleting.value = true

  try {
    await deleteAdminCategory(category.id)
    selectedCategory.value = null
    await fetchCategories()
    toast.add({
      severity: 'success',
      summary: '分类已删除',
      detail: category.name,
      life: 2400
    })
  } catch {
    // Admin HTTP client already shows a toast for request failures.
  } finally {
    deleting.value = false
  }
}

const formatDate = (value?: string) => {
  if (!value) return '未记录'

  const date = new Date(value)
  if (!Number.isFinite(date.getTime())) return '未记录'

  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const numberLabel = (value?: string | number) => {
  return new Intl.NumberFormat('zh-CN').format(Number(value || 0))
}

onMounted(() => {
  void fetchCategories()
})
</script>

<template>
  <main class="space-y-5">
    <header class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
      <div class="min-w-0 space-y-2">
        <div class="flex flex-wrap items-center gap-2">
          <h1 class="m-0 text-2xl font-black leading-tight text-foreground text-balance">
            分类管理
          </h1>
          <AppBadge tone="neutral" class="tabular-nums">
            共 {{ numberLabel(categories.length) }} 个
          </AppBadge>
        </div>
        <p class="m-0 max-w-2xl text-sm leading-6 text-muted-foreground text-pretty">
          维护公开文章的归档入口，保持分类名称简洁、稳定、容易浏览。
        </p>
      </div>
    </header>

    <form
      class="archive-surface flex flex-col gap-3 rounded-archive p-4 sm:flex-row"
      @submit.prevent="createCategory"
    >
      <label class="min-w-0 flex-1">
        <span class="sr-only">新分类名称</span>
        <AppInput v-model="newCategoryName" placeholder="新分类名称" :disabled="creating" />
      </label>
      <AppButton type="submit" class="sm:shrink-0" :disabled="creating">
        <Plus class="size-4" aria-hidden="true" />
        {{ creating ? '创建中...' : '创建分类' }}
      </AppButton>
    </form>

    <section
      v-if="loading"
      class="archive-surface rounded-archive p-8 text-center text-sm font-semibold text-muted-foreground"
      aria-live="polite"
    >
      正在读取分类
    </section>

    <section v-else-if="categories.length > 0" class="space-y-3" aria-label="分类列表">
      <article
        v-for="category in categories"
        :key="category.id"
        class="archive-surface rounded-archive p-4 transition duration-200 hover:border-accent/35 hover:bg-surface-raised/70"
      >
        <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div class="min-w-0 flex-1 space-y-3">
            <form
              v-if="editingId === String(category.id)"
              class="flex flex-col gap-2 sm:flex-row"
              @submit.prevent="saveEdit(category)"
            >
              <label class="min-w-0 flex-1">
                <span class="sr-only">分类名称</span>
                <AppInput v-model="draftName" placeholder="分类名称" :disabled="saving" />
              </label>
              <div class="flex gap-2">
                <AppButton type="submit" size="sm" :disabled="saving">
                  <Save class="size-4" aria-hidden="true" />
                  保存
                </AppButton>
                <AppButton variant="ghost" size="sm" :disabled="saving" @click="cancelEdit">
                  <X class="size-4" aria-hidden="true" />
                  取消
                </AppButton>
              </div>
            </form>

            <div v-else class="flex min-w-0 items-center gap-3">
              <span
                class="grid size-10 shrink-0 place-items-center rounded-full bg-accent/10 text-accent"
              >
                <FolderTree class="size-4" aria-hidden="true" />
              </span>
              <div class="min-w-0">
                <h2 class="m-0 truncate text-lg font-black leading-snug text-foreground">
                  {{ category.name }}
                </h2>
                <dl
                  class="m-0 mt-1 flex flex-wrap gap-x-4 gap-y-1 text-xs font-semibold text-muted-foreground"
                >
                  <div class="flex items-center gap-1.5">
                    <dt class="sr-only">文章数量</dt>
                    <dd class="m-0 tabular-nums">
                      {{ numberLabel(category.article_count) }} 篇文章
                    </dd>
                  </div>
                  <div class="flex items-center gap-1.5">
                    <dt class="sr-only">创建时间</dt>
                    <CalendarDays class="size-3.5" aria-hidden="true" />
                    <dd class="m-0 tabular-nums">{{ formatDate(category.created_at) }}</dd>
                  </div>
                </dl>
              </div>
            </div>
          </div>

          <div class="flex flex-wrap items-center gap-2 lg:justify-end">
            <AppButton
              variant="secondary"
              size="sm"
              :disabled="editingId === String(category.id)"
              @click="startEdit(category)"
            >
              <Pencil class="size-4" aria-hidden="true" />
              重命名
            </AppButton>
            <AppButton
              variant="ghost"
              size="sm"
              class="text-danger hover:text-danger"
              @click="askDelete(category)"
            >
              <Trash2 class="size-4" aria-hidden="true" />
              删除
            </AppButton>
          </div>
        </div>
      </article>
    </section>

    <section v-else class="archive-surface rounded-archive p-8 text-center">
      <p class="m-0 text-lg font-black text-foreground">还没有分类</p>
      <p class="m-0 mt-2 text-sm leading-6 text-muted-foreground">
        创建第一个分类后，就可以在文章编辑器里归档内容。
      </p>
    </section>

    <ConfirmDialog
      :open="Boolean(selectedCategory)"
      title="删除分类"
      :description="`确认删除「${selectedCategory?.name || '未命名分类'}」吗？删除后相关文章会回到默认分类。`"
      :confirm-label="deleting ? '删除中...' : '删除分类'"
      cancel-label="取消"
      danger
      @cancel="cancelDelete"
      @confirm="confirmDelete"
    />
  </main>
</template>
