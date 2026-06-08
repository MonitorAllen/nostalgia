<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { ExternalLink, Home, LogIn, LogOut, Menu, Search, User, Wrench, X } from '@lucide/vue'
import { useUserStore } from '@/store/module/user'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'
import ThemeSwitcher from '@/components/ui/ThemeSwitcher.vue'

const router = useRouter()
const userStore = useUserStore()

const searchValue = ref('')
const mobileOpen = ref(false)
const userMenuOpen = ref(false)

const hasUser = computed(() => !!userStore.userInfo)
const username = computed(() => userStore.userInfo?.username ?? '访客')

const handleSearch = () => {
  const query = searchValue.value.trim()
  if (!query) return
  searchValue.value = ''
  mobileOpen.value = false
  router.push({ path: '/search', query: { q: query } })
}

const logout = () => {
  userStore.logout()
}
</script>

<template>
  <nav class="border-b border-glass-border/50 bg-background/55 px-4 py-3 backdrop-blur-xl">
    <div class="mx-auto flex max-w-7xl items-center gap-3">
      <RouterLink
        to="/"
        class="group flex items-center rounded-full text-foreground"
        aria-label="回到 Nostalgia 首页"
      >
        <img src="/logo.svg" alt="Nostalgia Logo" class="h-10 w-auto max-w-[11rem]" />
      </RouterLink>

      <div class="hidden items-center gap-1 md:flex">
        <RouterLink
          to="/"
          class="inline-flex h-10 items-center gap-2 rounded-full px-3 text-sm font-semibold text-muted-foreground transition hover:bg-muted hover:text-foreground"
        >
          <Home class="h-4 w-4" />
          主页
        </RouterLink>
        <a
          href="https://toolx.de5.net"
          target="_blank"
          rel="noreferrer"
          class="inline-flex h-10 items-center gap-2 rounded-full px-3 text-sm font-semibold text-muted-foreground transition hover:bg-muted hover:text-foreground"
        >
          <Wrench class="h-4 w-4" />
          工具
          <ExternalLink class="h-3.5 w-3.5" />
        </a>
      </div>

      <div class="ml-auto hidden min-w-[16rem] max-w-sm flex-1 items-center lg:flex">
        <label class="relative block w-full">
          <Search class="pointer-events-none absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <AppInput
            v-model="searchValue"
            class="pl-10"
            placeholder="搜索文章，空格分隔关键词"
            @keydown.enter="handleSearch"
          />
        </label>
      </div>

      <ThemeSwitcher class="hidden sm:inline-flex" />

      <div class="hidden items-center md:flex">
        <div v-if="hasUser" class="relative">
          <AppButton variant="ghost" size="sm" @click="userMenuOpen = !userMenuOpen">
            <User class="h-4 w-4" />
            {{ username }}
          </AppButton>
          <div
            v-if="userMenuOpen"
            class="archive-glass absolute right-0 mt-2 w-44 rounded-archive p-2"
          >
            <button
              type="button"
              class="flex w-full items-center gap-2 rounded-full px-3 py-2 text-sm font-semibold text-muted-foreground hover:bg-muted hover:text-foreground"
              @click="logout"
            >
              <LogOut class="h-4 w-4" />
              退出登录
            </button>
          </div>
        </div>
        <div v-else class="flex items-center gap-1">
          <RouterLink
            to="/login"
            class="inline-flex h-10 items-center gap-2 rounded-full px-3 text-sm font-semibold text-muted-foreground transition hover:bg-muted hover:text-foreground"
          >
            <LogIn class="h-4 w-4" />
            登录
          </RouterLink>
          <RouterLink
            to="/register"
            class="inline-flex h-10 items-center rounded-full bg-accent px-4 text-sm font-bold text-accent-foreground transition hover:bg-accent/90"
          >
            注册
          </RouterLink>
        </div>
      </div>

      <AppButton variant="ghost" size="icon" class="md:hidden" @click="mobileOpen = !mobileOpen">
        <X v-if="mobileOpen" class="h-5 w-5" />
        <Menu v-else class="h-5 w-5" />
      </AppButton>
    </div>

    <div v-if="mobileOpen" class="mx-auto mt-3 max-w-7xl md:hidden">
      <div class="archive-glass rounded-archive p-3">
        <label class="relative block">
          <Search class="pointer-events-none absolute left-4 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <AppInput
            v-model="searchValue"
            class="pl-10"
            placeholder="搜索文章"
            @keydown.enter="handleSearch"
          />
        </label>
        <div class="mt-3 flex flex-col gap-2">
          <RouterLink to="/" class="rounded-full px-3 py-2 text-sm font-semibold hover:bg-muted">主页</RouterLink>
          <a
            href="https://toolx.de5.net"
            target="_blank"
            rel="noreferrer"
            class="rounded-full px-3 py-2 text-sm font-semibold hover:bg-muted"
          >
            工具
          </a>
          <ThemeSwitcher class="w-max" />
          <button
            v-if="hasUser"
            type="button"
            class="rounded-full px-3 py-2 text-left text-sm font-semibold hover:bg-muted"
            @click="logout"
          >
            退出登录
          </button>
          <RouterLink v-else to="/login" class="rounded-full px-3 py-2 text-sm font-semibold hover:bg-muted">
            登录
          </RouterLink>
        </div>
      </div>
    </div>
  </nav>
</template>
