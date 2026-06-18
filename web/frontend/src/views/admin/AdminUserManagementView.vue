<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { Check, Pencil, RotateCcw, Search, ShieldOff, X } from '@lucide/vue'
import {
  disableAdminUser,
  enableAdminUser,
  listAdminUsers,
  updateAdminUser
} from '@/admin/api/adminUserApi'
import type { AdminUserStatusFilter, ManagedAdminUser } from '@/admin/types'
import AppBadge from '@/components/ui/AppBadge.vue'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import { useToast } from '@/composables/useToast'

const toast = useToast()

const users = ref<ManagedAdminUser[]>([])
const loading = ref(false)
const saving = ref(false)
const activeAction = ref('')
const total = ref(0)
const searchText = ref('')
const selectedStatus = ref<AdminUserStatusFilter>('all')
const page = ref(1)
const pageSize = ref(20)
const jumpPage = ref('1')
const editingUser = ref<ManagedAdminUser | null>(null)
const disablingUser = ref<ManagedAdminUser | null>(null)
const enablingUser = ref<ManagedAdminUser | null>(null)
const disableReason = ref('')

const editForm = reactive({
  full_name: '',
  email: '',
  is_email_verified: false
})

const pageSizeOptions = [10, 20, 50]
const statusOptions: Array<{ label: string; value: AdminUserStatusFilter }> = [
  { label: '全部', value: 'all' },
  { label: '启用', value: 'enabled' },
  { label: '禁用', value: 'disabled' }
]

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
const showingFrom = computed(() => (total.value === 0 ? 0 : (page.value - 1) * pageSize.value + 1))
const showingTo = computed(() => Math.min(page.value * pageSize.value, total.value))

const fetchUsers = async () => {
  if (loading.value) return
  loading.value = true

  try {
    const response = await listAdminUsers({
      q: searchText.value.trim() || undefined,
      status: selectedStatus.value,
      page: page.value,
      limit: pageSize.value
    })
    users.value = response.data.users ?? []
    total.value = Number(response.data.count || 0)

    if (users.value.length === 0 && page.value > 1) {
      page.value -= 1
      await fetchUsers()
    }
  } catch {
    users.value = []
  } finally {
    loading.value = false
  }
}

const resetToFirstPage = () => {
  page.value = 1
  jumpPage.value = '1'
  void fetchUsers()
}

const changePageSize = (event: Event) => {
  pageSize.value = Number((event.target as HTMLSelectElement).value)
  resetToFirstPage()
}

const changeStatus = (status: AdminUserStatusFilter) => {
  selectedStatus.value = status
  resetToFirstPage()
}

const goPage = (next: number) => {
  if (next < 1 || next > totalPages.value || next === page.value) return
  page.value = next
  jumpPage.value = String(next)
  void fetchUsers()
}

const jumpToPage = () => {
  const next = Number(jumpPage.value)
  if (!Number.isFinite(next)) return
  goPage(Math.min(totalPages.value, Math.max(1, Math.floor(next))))
}

const openEdit = (user: ManagedAdminUser) => {
  editingUser.value = user
  editForm.full_name = user.full_name
  editForm.email = user.email
  editForm.is_email_verified = user.is_email_verified
}

const closeEdit = () => {
  if (saving.value) return
  editingUser.value = null
}

const saveEdit = async () => {
  if (!editingUser.value || saving.value) return
  const fullName = editForm.full_name.trim()
  const email = editForm.email.trim()

  if (!fullName || !email.includes('@')) {
    toast.add({
      severity: 'warning',
      summary: '用户信息不完整',
      detail: '请输入姓名和有效邮箱',
      life: 2400
    })
    return
  }

  saving.value = true
  try {
    await updateAdminUser({
      id: editingUser.value.id,
      full_name: fullName,
      email,
      is_email_verified: editForm.is_email_verified
    })
    editingUser.value = null
    await fetchUsers()
    toast.add({ severity: 'success', summary: '用户已更新', detail: email, life: 2400 })
  } catch {
    // Admin HTTP client already shows request errors.
  } finally {
    saving.value = false
  }
}

const openDisable = (user: ManagedAdminUser) => {
  disablingUser.value = user
  disableReason.value = user.disabled_reason || ''
}

const closeDisable = () => {
  if (activeAction.value) return
  disablingUser.value = null
}

const confirmDisable = async () => {
  if (!disablingUser.value) return
  activeAction.value = `disable:${disablingUser.value.id}`
  try {
    await disableAdminUser(disablingUser.value.id, {
      reason: disableReason.value.trim() || undefined
    })
    disablingUser.value = null
    await fetchUsers()
    toast.add({
      severity: 'success',
      summary: '用户已禁用',
      detail: '该用户现有会话已阻断',
      life: 2600
    })
  } catch {
    // Admin HTTP client already shows request errors.
  } finally {
    activeAction.value = ''
  }
}

const openEnable = (user: ManagedAdminUser) => {
  enablingUser.value = user
}

const closeEnable = () => {
  if (activeAction.value) return
  enablingUser.value = null
}

const confirmEnable = async () => {
  if (!enablingUser.value) return
  activeAction.value = `enable:${enablingUser.value.id}`
  try {
    await enableAdminUser(enablingUser.value.id)
    enablingUser.value = null
    await fetchUsers()
    toast.add({ severity: 'success', summary: '用户已恢复', detail: '用户可重新登录', life: 2600 })
  } catch {
    // Admin HTTP client already shows request errors.
  } finally {
    activeAction.value = ''
  }
}

const isDisabled = (user: ManagedAdminUser) => Boolean(user.disabled_at)
const isBusy = (key: string) => activeAction.value === key

const formatDate = (value?: string) => {
  if (!value) return '未记录'
  const date = new Date(value)
  if (!Number.isFinite(date.getTime()) || date.getFullYear() <= 1) return '未记录'
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(
    date.getDate()
  ).padStart(2, '0')}`
}

const numberLabel = (value: number) => new Intl.NumberFormat('zh-CN').format(value || 0)

watch(page, (value) => {
  jumpPage.value = String(value)
})

onMounted(() => {
  void fetchUsers()
})
</script>

<template>
  <main class="space-y-5">
    <header class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
      <div class="min-w-0 space-y-2">
        <div class="flex flex-wrap items-center gap-2">
          <h1 class="m-0 text-2xl font-black leading-tight text-foreground text-balance">
            用户管理
          </h1>
          <AppBadge tone="neutral" class="tabular-nums">共 {{ numberLabel(total) }} 位</AppBadge>
        </div>
        <p class="m-0 max-w-2xl text-sm leading-6 text-muted-foreground text-pretty">
          管理前台注册的访客账号，支持资料维护、账号禁用和恢复。
        </p>
      </div>
    </header>

    <section class="archive-surface rounded-archive p-4" aria-label="用户筛选">
      <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <div class="flex flex-1 flex-col gap-2 sm:flex-row sm:items-center">
          <label class="relative min-w-0 flex-1">
            <span class="sr-only">搜索用户</span>
            <Search
              class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"
              aria-hidden="true"
            />
            <AppInput
              v-model="searchText"
              class="pl-10"
              placeholder="搜索用户名、姓名或邮箱"
              @keyup.enter="resetToFirstPage"
            />
          </label>
          <AppButton variant="secondary" class="sm:shrink-0" @click="resetToFirstPage">搜索</AppButton>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <div class="inline-flex rounded-full border border-border bg-surface p-1" role="group" aria-label="账号状态筛选">
            <button
              v-for="item in statusOptions"
              :key="item.value"
              type="button"
              class="h-9 rounded-full px-3 text-sm font-semibold transition-colors duration-200 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
              :class="
                selectedStatus === item.value
                  ? 'bg-accent text-accent-foreground'
                  : 'text-muted-foreground hover:bg-muted hover:text-foreground'
              "
              @click="changeStatus(item.value)"
            >
              {{ item.label }}
            </button>
          </div>

          <label>
            <span class="sr-only">每页数量</span>
            <select
              :value="pageSize"
              class="h-10 rounded-full border border-border bg-surface px-3 text-sm font-semibold text-foreground focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/18"
              @change="changePageSize"
            >
              <option v-for="size in pageSizeOptions" :key="size" :value="size">
                {{ size }} / 页
              </option>
            </select>
          </label>
        </div>
      </div>

      <div class="mt-4 overflow-x-auto">
        <table class="w-full min-w-[72rem] table-fixed border-separate border-spacing-0 text-left text-sm">
          <colgroup>
            <col class="w-[23%]" />
            <col class="w-[21%]" />
            <col class="w-[24%]" />
            <col class="w-[6.5rem]" />
            <col class="w-[6.5rem]" />
            <col class="w-[8rem]" />
            <col class="w-[11rem]" />
          </colgroup>
          <thead class="text-xs uppercase text-muted-foreground">
            <tr>
              <th class="border-b border-border px-3 py-3 font-semibold" scope="col">用户名</th>
              <th class="border-b border-border px-3 py-3 font-semibold" scope="col">姓名</th>
              <th class="border-b border-border px-3 py-3 font-semibold" scope="col">邮箱</th>
              <th class="border-b border-border px-3 py-3 font-semibold whitespace-nowrap" scope="col">
                邮箱状态
              </th>
              <th class="border-b border-border px-3 py-3 font-semibold whitespace-nowrap" scope="col">
                账号状态
              </th>
              <th class="border-b border-border px-3 py-3 font-semibold whitespace-nowrap" scope="col">
                注册时间
              </th>
              <th class="border-b border-border px-3 py-3 text-right font-semibold" scope="col">
                操作
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td class="px-3 py-8 text-center text-muted-foreground" colspan="7" aria-live="polite">
                加载中...
              </td>
            </tr>
            <tr v-else-if="users.length === 0">
              <td class="px-3 py-8 text-center text-muted-foreground" colspan="7">
                暂无用户
              </td>
            </tr>
            <tr v-for="user in users" v-else :key="user.id">
              <td class="max-w-0 border-b border-border/70 px-3 py-3 font-semibold text-foreground">
                <span class="block truncate break-normal" :title="user.username">
                  {{ user.username }}
                </span>
              </td>
              <td class="max-w-0 border-b border-border/70 px-3 py-3 text-foreground">
                <span class="block truncate break-normal" :title="user.full_name">
                  {{ user.full_name }}
                </span>
              </td>
              <td class="max-w-0 border-b border-border/70 px-3 py-3 text-muted-foreground">
                <span class="block truncate break-normal" :title="user.email">
                  {{ user.email }}
                </span>
              </td>
              <td class="border-b border-border/70 px-3 py-3 whitespace-nowrap">
                <AppBadge :tone="user.is_email_verified ? 'accent' : 'warning'">
                  <Check v-if="user.is_email_verified" class="size-3" aria-hidden="true" />
                  <X v-else class="size-3" aria-hidden="true" />
                  {{ user.is_email_verified ? '已验证' : '未验证' }}
                </AppBadge>
              </td>
              <td class="border-b border-border/70 px-3 py-3 whitespace-nowrap">
                <AppBadge :tone="isDisabled(user) ? 'danger' : 'accent'">
                  {{ isDisabled(user) ? '已禁用' : '启用中' }}
                </AppBadge>
              </td>
              <td class="border-b border-border/70 px-3 py-3 text-muted-foreground tabular-nums whitespace-nowrap">
                {{ formatDate(user.created_at) }}
              </td>
              <td class="border-b border-border/70 px-3 py-3 whitespace-nowrap">
                <div class="flex justify-end gap-2">
                  <AppButton size="sm" variant="ghost" @click="openEdit(user)">
                    <Pencil class="size-[18px]" aria-hidden="true" />
                    编辑
                  </AppButton>
                  <AppButton
                    v-if="isDisabled(user)"
                    size="sm"
                    variant="secondary"
                    :disabled="isBusy(`enable:${user.id}`)"
                    @click="openEnable(user)"
                  >
                    <RotateCcw class="size-[18px]" aria-hidden="true" />
                    恢复
                  </AppButton>
                  <AppButton
                    v-else
                    size="sm"
                    variant="danger"
                    :disabled="isBusy(`disable:${user.id}`)"
                    @click="openDisable(user)"
                  >
                    <ShieldOff class="size-[18px]" aria-hidden="true" />
                    禁用
                  </AppButton>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <footer
        class="mt-4 flex flex-col gap-3 text-sm text-muted-foreground lg:flex-row lg:items-center lg:justify-between"
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
    </section>

    <Teleport to="body">
      <div
        v-if="editingUser"
        class="fixed inset-0 z-50 grid place-items-center bg-background/60 p-4 backdrop-blur-sm"
        role="dialog"
        aria-modal="true"
        aria-labelledby="edit-user-title"
      >
        <form class="archive-surface w-full max-w-lg rounded-archive p-5" @submit.prevent="saveEdit">
          <h2 id="edit-user-title" class="m-0 text-lg font-bold text-foreground">编辑用户</h2>
          <div class="mt-4 space-y-3">
            <label class="block text-sm font-semibold text-foreground">
              姓名
              <AppInput v-model="editForm.full_name" class="mt-2" />
            </label>
            <label class="block text-sm font-semibold text-foreground">
              邮箱
              <AppInput v-model="editForm.email" class="mt-2" type="email" />
            </label>
            <label class="flex items-center gap-2 text-sm font-semibold text-foreground">
              <input
                v-model="editForm.is_email_verified"
                type="checkbox"
                class="size-4 rounded border-border accent-accent"
              />
              邮箱已验证
            </label>
          </div>
          <div class="mt-5 flex justify-end gap-2">
            <AppButton variant="ghost" :disabled="saving" @click="closeEdit">取消</AppButton>
            <AppButton type="submit" :disabled="saving">
              {{ saving ? '保存中...' : '保存' }}
            </AppButton>
          </div>
        </form>
      </div>
    </Teleport>

    <Teleport to="body">
      <div
        v-if="disablingUser"
        class="fixed inset-0 z-50 grid place-items-center bg-background/60 p-4 backdrop-blur-sm"
        role="dialog"
        aria-modal="true"
        aria-labelledby="disable-user-title"
      >
        <div class="archive-surface w-full max-w-md rounded-archive p-5">
          <h2 id="disable-user-title" class="m-0 text-lg font-bold text-foreground">禁用用户</h2>
          <p class="mt-2 text-sm text-muted-foreground">
            禁用后该用户无法重新登录，现有刷新会话会被阻断。
          </p>
          <label class="mt-4 block text-sm font-semibold text-foreground">
            原因
            <textarea
              v-model="disableReason"
              class="mt-2 min-h-24 w-full rounded-2xl border border-border bg-surface px-4 py-3 text-sm text-foreground focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/18"
              placeholder="可选"
            />
          </label>
          <div class="mt-5 flex justify-end gap-2">
            <AppButton variant="ghost" :disabled="Boolean(activeAction)" @click="closeDisable">
              取消
            </AppButton>
            <AppButton variant="danger" :disabled="Boolean(activeAction)" @click="confirmDisable">
              {{ activeAction ? '禁用中...' : '确认禁用' }}
            </AppButton>
          </div>
        </div>
      </div>
    </Teleport>

    <ConfirmDialog
      :open="Boolean(enablingUser)"
      title="恢复用户"
      description="恢复后用户可以重新登录，但旧会话不会自动恢复。"
      :confirm-label="activeAction ? '恢复中...' : '确认恢复'"
      :danger="false"
      @cancel="closeEnable"
      @confirm="confirmEnable"
    />
  </main>
</template>
