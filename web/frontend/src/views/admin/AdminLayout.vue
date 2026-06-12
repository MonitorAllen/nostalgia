<script setup lang="ts">
import { computed, ref, watch, type Component } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import {
  Bot,
  BookOpen,
  ExternalLink,
  FolderTree,
  LogOut,
  Menu,
  PanelLeftClose,
  PanelLeftOpen,
  Tags,
  X
} from '@lucide/vue'
import { useAdminAuthStore } from '@/admin/stores/adminAuth'
import AppButton from '@/components/ui/AppButton.vue'
import ThemeSwitcher from '@/components/ui/ThemeSwitcher.vue'
import { cn } from '@/lib/utils'

const route = useRoute()
const router = useRouter()
const adminAuth = useAdminAuthStore()
const mobileMenuOpen = ref(false)
const sidebarCollapsed = ref(false)

const navItems: Array<{
  label: string
  to: { name: string }
  icon: Component
  activeRoutes: string[]
}> = [
  {
    label: '文章',
    to: { name: 'adminArticles' },
    icon: BookOpen,
    activeRoutes: ['adminArticles', 'adminArticleNew', 'adminArticleEdit'],
  },
  {
    label: '分类',
    to: { name: 'adminCategories' },
    icon: FolderTree,
    activeRoutes: ['adminCategories'],
  },
  {
    label: 'AI 设置',
    to: { name: 'adminAiSettings' },
    icon: Bot,
    activeRoutes: ['adminAiSettings'],
  },
]

const adminName = computed(() => adminAuth.admin?.username || 'Owner')
const isEditorRoute = computed(() =>
  ['adminArticleNew', 'adminArticleEdit'].includes(String(route.name || ''))
)
const isSidebarCollapsed = computed(() => sidebarCollapsed.value)
const adminShellStyle = computed(() => ({
  '--admin-sidebar-width': isSidebarCollapsed.value ? '4.75rem' : '16.5rem',
}))

const isActive = (activeRoutes: string[]) => {
  return typeof route.name === 'string' && activeRoutes.includes(route.name)
}

const navLinkClass = (activeRoutes: string[], collapsed = false) =>
  cn(
    'flex h-10 items-center gap-3 rounded-full text-sm font-semibold transition-all duration-200',
    'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent',
    collapsed ? 'justify-center px-0' : 'px-3',
    isActive(activeRoutes)
      ? 'bg-accent/10 text-accent'
      : 'text-muted-foreground hover:bg-muted hover:text-foreground',
  )

const toggleSidebar = () => {
  sidebarCollapsed.value = !sidebarCollapsed.value
}

const closeMobileMenu = () => {
  mobileMenuOpen.value = false
}

const handleLogout = async () => {
  closeMobileMenu()
  adminAuth.clear()
  await router.replace({ name: 'adminLogin' })
}

watch(
  () => route.name,
  () => {
    if (isEditorRoute.value) {
      sidebarCollapsed.value = true
    }
  },
  { immediate: true },
)
</script>

<template>
  <div class="min-h-dvh bg-background text-foreground">
    <header class="sticky top-0 z-40 border-b border-border/70 bg-background/95 px-4 py-3 backdrop-blur lg:hidden">
      <div class="mx-auto flex max-w-6xl items-center justify-between gap-3">
        <RouterLink
          :to="{ name: 'adminArticles' }"
          class="flex min-w-0 items-center gap-3 rounded-full focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
          @click="closeMobileMenu"
        >
          <span class="archive-glass grid size-9 shrink-0 place-items-center rounded-full">
            <Tags class="size-5 text-accent" aria-hidden="true" />
          </span>
          <div class="min-w-0">
            <p class="m-0 truncate text-sm font-black leading-5">Nostalgia Admin</p>
            <p class="m-0 truncate text-xs font-semibold text-muted-foreground">{{ adminName }}</p>
          </div>
        </RouterLink>

        <div class="flex items-center gap-2">
          <ThemeSwitcher />
          <AppButton
            variant="ghost"
            size="icon"
            :aria-label="mobileMenuOpen ? '关闭后台导航' : '打开后台导航'"
            :aria-expanded="mobileMenuOpen"
            aria-controls="admin-mobile-nav"
            @click="mobileMenuOpen = !mobileMenuOpen"
          >
            <X v-if="mobileMenuOpen" class="size-4" aria-hidden="true" />
            <Menu v-else class="size-4" aria-hidden="true" />
          </AppButton>
        </div>
      </div>

      <nav
        v-if="mobileMenuOpen"
        id="admin-mobile-nav"
        class="archive-glass mx-auto mt-3 max-w-6xl rounded-archive p-2"
        aria-label="后台导航"
      >
        <ul class="m-0 list-none space-y-1 p-0">
          <li v-for="item in navItems" :key="item.label">
            <RouterLink :to="item.to" :class="navLinkClass(item.activeRoutes)" @click="closeMobileMenu">
              <component :is="item.icon" class="size-4" aria-hidden="true" />
              <span>{{ item.label }}</span>
            </RouterLink>
          </li>
          <li>
            <RouterLink
              :to="{ name: 'home' }"
              class="flex h-10 items-center gap-3 rounded-full px-3 text-sm font-semibold text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
              @click="closeMobileMenu"
            >
              <ExternalLink class="size-4" aria-hidden="true" />
              <span>查看站点</span>
            </RouterLink>
          </li>
          <li>
            <AppButton variant="ghost" class="h-10 w-full justify-start px-3 text-danger hover:text-danger" @click="handleLogout">
              <LogOut class="size-4" aria-hidden="true" />
              <span>退出登录</span>
            </AppButton>
          </li>
        </ul>
      </nav>
    </header>

    <div
      class="grid min-h-dvh w-full grid-cols-1 transition-[grid-template-columns] duration-200 lg:grid-cols-[var(--admin-sidebar-width)_minmax(0,1fr)]"
      :style="adminShellStyle"
    >
      <aside class="hidden border-r border-border/70 bg-surface/45 px-3 py-4 lg:block">
        <div class="sticky top-4 flex h-[calc(100dvh-2rem)] flex-col">
          <div
            class="flex items-center gap-2"
            :class="isSidebarCollapsed ? 'justify-center' : 'justify-between'"
          >
            <RouterLink
              :to="{ name: 'adminArticles' }"
              class="flex min-w-0 items-center gap-3 rounded-archive px-2 py-2 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
              :aria-label="isSidebarCollapsed ? 'Nostalgia Admin' : undefined"
            >
              <span class="archive-glass grid size-10 shrink-0 place-items-center rounded-full">
                <Tags class="size-5 text-accent" aria-hidden="true" />
              </span>
              <div v-if="!isSidebarCollapsed" class="min-w-0">
                <p class="m-0 truncate text-sm font-black leading-5">Nostalgia Admin</p>
                <p class="m-0 truncate text-xs font-semibold text-muted-foreground">{{ adminName }}</p>
              </div>
            </RouterLink>
            <AppButton
              variant="ghost"
              size="icon"
              :aria-label="isSidebarCollapsed ? '展开后台导航' : '收起后台导航'"
              :title="isSidebarCollapsed ? '展开后台导航' : '收起后台导航'"
              @click="toggleSidebar"
            >
              <PanelLeftOpen v-if="isSidebarCollapsed" class="size-4" aria-hidden="true" />
              <PanelLeftClose v-else class="size-4" aria-hidden="true" />
            </AppButton>
          </div>

          <nav class="mt-7" aria-label="后台导航">
            <ul class="m-0 list-none space-y-1 p-0">
              <li v-for="item in navItems" :key="item.label">
                <RouterLink
                  :to="item.to"
                  :class="navLinkClass(item.activeRoutes, isSidebarCollapsed)"
                  :title="isSidebarCollapsed ? item.label : undefined"
                >
                  <component :is="item.icon" class="size-4 shrink-0" aria-hidden="true" />
                  <span v-if="!isSidebarCollapsed">{{ item.label }}</span>
                </RouterLink>
              </li>
            </ul>
          </nav>

          <div class="mt-auto space-y-3 border-t border-border/70 pt-4">
            <ThemeSwitcher v-if="!isSidebarCollapsed" />
            <RouterLink
              :to="{ name: 'home' }"
              class="flex h-10 items-center gap-3 rounded-full text-sm font-semibold text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
              :class="isSidebarCollapsed ? 'justify-center px-0' : 'px-3'"
              :title="isSidebarCollapsed ? '查看站点' : undefined"
            >
              <ExternalLink class="size-4 shrink-0" aria-hidden="true" />
              <span v-if="!isSidebarCollapsed">查看站点</span>
            </RouterLink>
            <AppButton
              variant="ghost"
              class="h-10 w-full text-danger hover:text-danger"
              :class="isSidebarCollapsed ? 'justify-center px-0' : 'justify-start px-3'"
              :aria-label="isSidebarCollapsed ? '退出登录' : undefined"
              :title="isSidebarCollapsed ? '退出登录' : undefined"
              @click="handleLogout"
            >
              <LogOut class="size-4 shrink-0" aria-hidden="true" />
              <span v-if="!isSidebarCollapsed">退出登录</span>
            </AppButton>
          </div>
        </div>
      </aside>

      <section class="min-w-0 px-4 py-5 sm:px-6 lg:px-8 xl:px-10" aria-label="后台内容">
        <RouterView />
      </section>
    </div>
  </div>
</template>
