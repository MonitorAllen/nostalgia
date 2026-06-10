<script setup lang="ts">
import { computed, ref, type Component } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { BookOpen, ExternalLink, FolderTree, LogOut, Menu, X } from '@lucide/vue'
import { useAdminAuthStore } from '@/admin/stores/adminAuth'
import AppButton from '@/components/ui/AppButton.vue'
import ThemeSwitcher from '@/components/ui/ThemeSwitcher.vue'
import { cn } from '@/lib/utils'

const route = useRoute()
const router = useRouter()
const adminAuth = useAdminAuthStore()
const mobileMenuOpen = ref(false)

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
]

const adminName = computed(() => adminAuth.admin?.username || 'Owner')

const isActive = (activeRoutes: string[]) => {
  return typeof route.name === 'string' && activeRoutes.includes(route.name)
}

const navLinkClass = (activeRoutes: string[]) =>
  cn(
    'flex h-10 items-center gap-3 rounded-full px-3 text-sm font-semibold transition-colors',
    'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent',
    isActive(activeRoutes)
      ? 'bg-accent/10 text-accent'
      : 'text-muted-foreground hover:bg-muted hover:text-foreground',
  )

const closeMobileMenu = () => {
  mobileMenuOpen.value = false
}

const handleLogout = async () => {
  closeMobileMenu()
  adminAuth.clear()
  await router.replace({ name: 'adminLogin' })
}
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
          <img src="/logo.svg" alt="Nostalgia" class="size-9 shrink-0" />
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

    <div class="mx-auto grid min-h-dvh w-full max-w-7xl grid-cols-1 lg:grid-cols-[17rem_1fr]">
      <aside class="hidden border-r border-border/70 px-4 py-5 lg:block">
        <div class="sticky top-5 flex h-[calc(100dvh-2.5rem)] flex-col">
          <RouterLink
            :to="{ name: 'adminArticles' }"
            class="flex items-center gap-3 rounded-archive px-2 py-2 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
          >
            <img src="/logo.svg" alt="Nostalgia" class="size-10 shrink-0" />
            <div class="min-w-0">
              <p class="m-0 truncate text-sm font-black leading-5">Nostalgia Admin</p>
              <p class="m-0 truncate text-xs font-semibold text-muted-foreground">{{ adminName }}</p>
            </div>
          </RouterLink>

          <nav class="mt-7" aria-label="后台导航">
            <ul class="m-0 list-none space-y-1 p-0">
              <li v-for="item in navItems" :key="item.label">
                <RouterLink :to="item.to" :class="navLinkClass(item.activeRoutes)">
                  <component :is="item.icon" class="size-4" aria-hidden="true" />
                  <span>{{ item.label }}</span>
                </RouterLink>
              </li>
            </ul>
          </nav>

          <div class="mt-auto space-y-3 border-t border-border/70 pt-4">
            <ThemeSwitcher />
            <RouterLink
              :to="{ name: 'home' }"
              class="flex h-10 items-center gap-3 rounded-full px-3 text-sm font-semibold text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-accent"
            >
              <ExternalLink class="size-4" aria-hidden="true" />
              <span>查看站点</span>
            </RouterLink>
            <AppButton variant="ghost" class="h-10 w-full justify-start px-3 text-danger hover:text-danger" @click="handleLogout">
              <LogOut class="size-4" aria-hidden="true" />
              <span>退出登录</span>
            </AppButton>
          </div>
        </div>
      </aside>

      <section class="min-w-0 px-4 py-5 sm:px-6 lg:px-8" aria-label="后台内容">
        <RouterView />
      </section>
    </div>
  </div>
</template>
