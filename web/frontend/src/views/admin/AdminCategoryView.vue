<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { CalendarDays, FolderTree, Pencil, Plus, Save, Trash2, X } from '@lucide/vue'
import type { AdminCategory } from '@/admin/types'
import {
  createAdminCategory,
  deleteAdminCategory,
  listAdminCategories,
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
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const jumpPage = ref('1')
const editingId = ref<string | null>(null)
const draftName = ref('')
const newCategoryName = ref('')
const selectedCategory = ref<AdminCategory | null>(null)
const selectedCategoryIds = ref<string[]>([])
const bulkDeleteCandidates = ref<AdminCategory[] | null>(null)
const creating = ref(false)
const saving = ref(false)
const deleting = ref(false)
const bulkDeleting = ref(false)

const pageSizeOptions = [10, 20, 50]

const normalizeCategory = (category: AdminCategory): AdminCategory => ({
  ...category,
  is_system: Boolean(category.is_system ?? category.isSystem)
})

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
const showingFrom = computed(() => (total.value === 0 ? 0 : (page.value - 1) * pageSize.value + 1))
const showingTo = computed(() => Math.min(page.value * pageSize.value, total.value))
const selectableCategories = computed(() => categories.value.filter((category) => !category.is_system))
const selectedCategories = computed(() =>
  categories.value.filter(
    (category) => !category.is_system && selectedCategoryIds.value.includes(String(category.id))
  )
)
const selectedCategoryCount = computed(() => selectedCategories.value.length)
const allSelectableSelected = computed(
  () =>
    selectableCategories.value.length > 0 &&
    selectableCategories.value.every((category) =>
      selectedCategoryIds.value.includes(String(category.id))
    )
)
const bulkDeleteDescription = computed(() => {
  const candidates = bulkDeleteCandidates.value ?? []
  const names = candidates.map((category) => category.name).join('、')
  return `确认删除 ${candidates.length} 个分类吗？删除后相关文章会回到默认分类。${names ? `本次包括：${names}` : ''}`
})

const numberLabel = (value?: string | number) => {
  return new Intl.NumberFormat('zh-CN').format(Number(value || 0))
}

const fetchCategories = async () => {
  if (loading.value) return
  loading.value = true
  let shouldFetchPreviousPage = false

  try {
    const response = await listAdminCategories({
      page: page.value,
      limit: pageSize.value
    })
    categories.value = (response.data.categories ?? []).map(normalizeCategory)
    total.value = Number(response.data.count || 0)
    selectedCategoryIds.value = []

    if (categories.value.length === 0 && page.value > 1) {
      page.value -= 1
      jumpPage.value = String(page.value)
      shouldFetchPreviousPage = true
    }
  } catch {
    categories.value = []
    total.value = 0
    selectedCategoryIds.value = []
  } finally {
    loading.value = false
  }

  if (shouldFetchPreviousPage) {
    await fetchCategories()
  }
}

const resetToFirstPage = async () => {
  page.value = 1
  jumpPage.value = '1'
  await fetchCategories()
}

const changePageSize = (event: Event) => {
  pageSize.value = Number((event.target as HTMLSelectElement).value)
  void resetToFirstPage()
}

const goPage = (next: number) => {
  if (next < 1 || next > totalPages.value || next === page.value) return
  page.value = next
  jumpPage.value = String(next)
  void fetchCategories()
}

const jumpToPage = () => {
  const next = Number(jumpPage.value)
  if (!Number.isFinite(next)) return
  goPage(Math.min(totalPages.value, Math.max(1, Math.floor(next))))
}

const isCategorySelected = (category: AdminCategory) => {
  return selectedCategoryIds.value.includes(String(category.id))
}

const toggleCategorySelection = (category: AdminCategory) => {
  if (category.is_system) return

  const id = String(category.id)
  if (selectedCategoryIds.value.includes(id)) {
    selectedCategoryIds.value = selectedCategoryIds.value.filter((value) => value !== id)
    return
  }

  selectedCategoryIds.value = [...selectedCategoryIds.value, id]
}

const togglePageSelection = () => {
  const ids = selectableCategories.value.map((category) => String(category.id))
  if (allSelectableSelected.value) {
    selectedCategoryIds.value = selectedCategoryIds.value.filter((id) => !ids.includes(id))
    return
  }

  selectedCategoryIds.value = Array.from(new Set([...selectedCategoryIds.value, ...ids]))
}

const openBulkDelete = () => {
  if (selectedCategories.value.length === 0) return
  bulkDeleteCandidates.value = [...selectedCategories.value]
}

const cancelBulkDelete = () => {
  if (bulkDeleting.value) return
  bulkDeleteCandidates.value = null
}

const confirmBulkDelete = async () => {
  const candidates = bulkDeleteCandidates.value ?? []
  if (candidates.length === 0 || bulkDeleting.value) return

  bulkDeleting.value = true
  let successCount = 0
  let failedCount = 0

  for (const category of candidates) {
    try {
      await deleteAdminCategory(category.id)
      successCount += 1
    } catch {
      failedCount += 1
    }
  }

  try {
    bulkDeleteCandidates.value = null
    selectedCategoryIds.value = []
    await fetchCategories()

    if (successCount > 0) {
      toast.add({
        severity: 'success',
        summary: '分类已删除',
        detail: `已删除 ${numberLabel(successCount)} 个分类`,
        life: 2400
      })
    }

    if (failedCount > 0) {
      toast.add({
        severity: 'warning',
        summary: '部分分类删除失败',
        detail: `${numberLabel(failedCount)} 个分类未能删除`,
        life: 3000
      })
    }
  } finally {
    bulkDeleting.value = false
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
    await resetToFirstPage()
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
  if (category.is_system) return
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
  if (category.is_system) return
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
    selectedCategoryIds.value = selectedCategoryIds.value.filter((id) => id !== String(category.id))
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

watch(page, (value) => {
  jumpPage.value = String(value)
})

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
            共 {{ numberLabel(total) }} 个
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

    <section class="archive-surface rounded-archive p-4" aria-label="分类操作">
      <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <div class="flex flex-wrap items-center gap-3">
          <label class="inline-flex items-center gap-2 text-sm font-semibold text-foreground">
            <input
              type="checkbox"
              class="size-4 rounded border-border accent-accent"
              :checked="allSelectableSelected"
              :disabled="selectableCategories.length === 0"
              @change="togglePageSelection"
            />
            全选本页
          </label>
          <span class="text-sm text-muted-foreground tabular-nums">
            已选择 {{ numberLabel(selectedCategoryCount) }} 个
          </span>
          <AppButton
            size="sm"
            variant="danger"
            :disabled="selectedCategoryCount === 0 || bulkDeleting"
            @click="openBulkDelete"
          >
            <Trash2 class="size-4" aria-hidden="true" />
            批量删除
          </AppButton>
        </div>

        <label class="inline-flex items-center gap-2 text-sm font-semibold text-muted-foreground">
          每页
          <select
            :value="pageSize"
            class="h-10 rounded-full border border-border bg-surface px-3 text-sm font-semibold text-foreground focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/18"
            @change="changePageSize"
          >
            <option v-for="size in pageSizeOptions" :key="size" :value="size">
              {{ size }} 条
            </option>
          </select>
        </label>
      </div>
    </section>

    <section
      v-if="loading"
      class="archive-surface rounded-archive p-8 text-center text-sm font-semibold text-muted-foreground"
      aria-live="polite"
    >
      正在读取分类
    </section>

    <section v-else class="space-y-3" aria-label="分类列表">
      <template v-if="categories.length > 0">
        <article
          v-for="category in categories"
          :key="category.id"
          class="archive-surface rounded-archive p-4 transition-colors duration-200 hover:border-accent/35 hover:bg-surface-raised/70"
          :class="isCategorySelected(category) ? 'border-accent/50 bg-accent/5' : ''"
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
                <label class="grid size-10 shrink-0 place-items-center">
                  <span class="sr-only">选择 {{ category.name }}</span>
                  <input
                    type="checkbox"
                    class="size-4 rounded border-border accent-accent"
                    :checked="isCategorySelected(category)"
                    :disabled="category.is_system"
                    @change="toggleCategorySelection(category)"
                  />
                </label>
                <span
                  class="grid size-10 shrink-0 place-items-center rounded-full bg-accent/10 text-accent"
                >
                  <FolderTree class="size-4" aria-hidden="true" />
                </span>
                <div class="min-w-0">
                  <div class="flex min-w-0 flex-wrap items-center gap-2">
                    <h2 class="m-0 truncate text-lg font-black leading-snug text-foreground">
                      {{ category.name }}
                    </h2>
                    <AppBadge v-if="category.is_system" tone="neutral">系统</AppBadge>
                  </div>
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
                :disabled="category.is_system || editingId === String(category.id)"
                @click="startEdit(category)"
              >
                <Pencil class="size-4" aria-hidden="true" />
                重命名
              </AppButton>
              <AppButton
                variant="ghost"
                size="sm"
                class="text-danger hover:text-danger"
                :disabled="category.is_system || deleting || bulkDeleting"
                @click="askDelete(category)"
              >
                <Trash2 class="size-4" aria-hidden="true" />
                删除
              </AppButton>
            </div>
          </div>
        </article>
      </template>

      <section v-else class="archive-surface rounded-archive p-8 text-center">
        <p class="m-0 text-lg font-black text-foreground">还没有分类</p>
        <p class="m-0 mt-2 text-sm leading-6 text-muted-foreground">
          创建第一个分类后，就可以在文章编辑器里归档内容。
        </p>
      </section>
    </section>

    <footer
      class="archive-surface flex flex-col gap-3 rounded-archive p-4 text-sm text-muted-foreground lg:flex-row lg:items-center lg:justify-between"
    >
      <span class="tabular-nums">
        显示 {{ numberLabel(showingFrom) }} - {{ numberLabel(showingTo) }}，共
        {{ numberLabel(total) }} 条
      </span>
      <div class="flex flex-wrap items-center gap-2">
        <AppButton size="sm" variant="ghost" :disabled="page <= 1" @click="goPage(page - 1)">
          上一页
        </AppButton>
        <span class="font-semibold text-foreground tabular-nums">
          {{ numberLabel(page) }} / {{ numberLabel(totalPages) }}
        </span>
        <AppButton
          size="sm"
          variant="ghost"
          :disabled="page >= totalPages"
          @click="goPage(page + 1)"
        >
          下一页
        </AppButton>
        <label>
          <span class="sr-only">跳转页码</span>
          <AppInput
            v-model="jumpPage"
            class="h-9 w-20 px-3 text-center tabular-nums"
            inputmode="numeric"
            @keyup.enter="jumpToPage"
          />
        </label>
        <AppButton size="sm" variant="secondary" @click="jumpToPage">跳转</AppButton>
      </div>
    </footer>

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

    <ConfirmDialog
      :open="Boolean(bulkDeleteCandidates)"
      title="批量删除分类"
      :description="bulkDeleteDescription"
      :confirm-label="bulkDeleting ? '删除中...' : '批量删除'"
      cancel-label="取消"
      danger
      @cancel="cancelBulkDelete"
      @confirm="confirmBulkDelete"
    />
  </main>
</template>
