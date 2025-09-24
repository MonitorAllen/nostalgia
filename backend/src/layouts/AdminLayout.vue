<template>
  <div class="min-h-screen flex flex-column">
    <!-- 只在 showNav 为 true 时显示导航栏 -->
    <template v-if="props.showNav">
      <header>
        <!-- 顶部导航栏 -->
        <nav class="layout-topbar">
          <div class="flex align-items-center px-3">
            <a
              href="#"
              class="flex align-items-center justify-content-center p-2 hover:surface-200 border-round mr-2 cursor-pointer"
              @click.prevent="toggleSidebar"
            >
              <i class="pi pi-bars"></i>
            </a>
            <span class="text-xl font-bold">Nostalgia Admin</span>
          </div>
          <ul class="flex list-none m-0 p-0 gap-2 align-items-center px-3">
            <li>
              <a
                href="#"
                class="flex align-items-center gap-2 px-3 py-2 hover:surface-200 border-round cursor-pointer"
                @click.prevent="toggleUserMenu"
                aria-haspopup="true"
                aria-controls="user-menu"
              >
                <Avatar
                  :label="authStore.admin?.username?.[0]?.toUpperCase() || '?'"
                  shape="circle"
                  class="mr-1"
                />
                <span class="text-lg">{{ authStore.admin?.username || '用户' }}</span>
                <i class="pi pi-angle-down"></i>
              </a>
              <ul
                v-show="isUserMenuVisible"
                id="user-menu"
                class="user-menu list-none m-0 p-0 absolute w-12rem surface-overlay border-none shadow-2"
              >
                <li>
                  <router-link
                    :to="{name: 'update-admin'}"
                    class="flex align-items-center gap-2 px-3 py-2 hover:surface-200 cursor-pointer text-color"
                  >
                    <i class="pi pi-user-edit"></i>
                    <span>修改密码</span>
                  </router-link>
                </li>
                <li class="border-top-1 surface-border"></li>
                <li>
                  <a
                    href="#"
                    class="flex align-items-center gap-2 px-3 py-2 hover:surface-200 cursor-pointer text-color"
                    @click.prevent="handleLogout"
                  >
                    <i class="pi pi-sign-out"></i>
                    <span>退出登录</span>
                  </a>
                </li>
              </ul>
            </li>
          </ul>
        </nav>
      </header>

      <div class="flex flex-grow-1">
        <!-- 侧边栏 -->
        <aside class="layout-sidebar" :class="{ 'layout-sidebar-collapsed': isSidebarCollapsed }">
          <nav class="h-full overflow-y-auto">
            <ul class="list-none pl-0 my-0">
              <li v-for="(item, index) in sysMenu" :key="index" class="mb-2">
                <template v-if="item.children?.length">
                  <a
                    href="#"
                    class="flex align-items-center justify-content-between px-2 py-3 hover:surface-200 cursor-pointer text-color"
                    @click.prevent="toggleSubmenu(index)"
                  >
                    <div class="flex align-items-center gap-2">
                      <i :class="[item.icon || 'pi pi-folder', 'text-xl']"></i>
                      <span v-show="!isSidebarCollapsed">{{ item.name }}</span>
                    </div>
                    <i
                      :class="[
                        'pi pi-angle-down transition-transform transition-duration-200',
                        { 'rotate-180': openSubmenus.includes(index) }
                      ]"
                      v-show="!isSidebarCollapsed"
                    ></i>
                  </a>
                  <ul
                    class="list-none pl-0 m-0 overflow-hidden transition-all transition-duration-200"
                    :class="{
                      'max-h-0': !openSubmenus.includes(index),
                      'max-h-20rem': openSubmenus.includes(index)
                    }"
                  >
                    <li v-for="(subItem, subIndex) in item.children" :key="subIndex">
                      <router-link
                        :to="subItem.path || ''"
                        class="flex align-items-center gap-2 p-3 pl-5 hover:surface-200 cursor-pointer text-color no-underline"
                      >
                        <i :class="[subItem.icon || 'pi pi-circle-fill', 'text-lg']"></i>
                        <span v-show="!isSidebarCollapsed">{{ subItem.name }}</span>
                      </router-link>
                    </li>
                  </ul>
                </template>
                <router-link
                  v-else
                  :to="item.path || ''"
                  class="flex align-items-center gap-2 p-3 hover:surface-200 border-round cursor-pointer text-color no-underline"
                >
                  <i :class="[item.icon || 'pi pi-circle', 'text-xl']"></i>
                  <span v-show="!isSidebarCollapsed">{{ item.name }}</span>
                </router-link>
              </li>
            </ul>
          </nav>
        </aside>

        <!-- 主内容区 -->
        <main class="layout-main p-2" :class="{ 'layout-main-collapsed': isSidebarCollapsed }">
          <router-view></router-view>
        </main>
      </div>
    </template>

    <!-- 当 showNav 为 false 时，直接显示内容 -->
    <template v-else>
      <main class="layout-main-full">
        <router-view></router-view>
      </main>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import Avatar from 'primevue/avatar'
import { useMenuStore } from '../stores/menu'
// 添加 showNav 属性，默认为 true
const props = withDefaults(defineProps<{
  showNav?: boolean
}>(), {
  showNav: true
})

const router = useRouter()
const authStore = useAuthStore()
const isSidebarCollapsed = ref(false)
const isUserMenuVisible = ref(false)
const openSubmenus = ref<number[]>([])

const menuStore = useMenuStore()
menuStore.initMenu()
const sysMenu = computed(() => menuStore.sysMenu)

const toggleSubmenu = (index: number) => {
  const position = openSubmenus.value.indexOf(index)
  if (position > -1) {
    openSubmenus.value.splice(position, 1)
  } else {
    openSubmenus.value.push(index)
  }
}

const navigateTo = (path?: string) => {
  if (path) {
    router.push(path)
  }
}

const toggleSidebar = () => {
  isSidebarCollapsed.value = !isSidebarCollapsed.value
}

const toggleUserMenu = (event: Event) => {
  event.stopPropagation()
  isUserMenuVisible.value = !isUserMenuVisible.value
}

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
  isUserMenuVisible.value = false
}

// 点击外部关闭用户菜单
const handleClickOutside = (event: MouseEvent) => {
  const userMenu = document.getElementById('user-menu')
  const userMenuTrigger = event.target as HTMLElement
  if (
    isUserMenuVisible.value &&
    userMenu &&
    !userMenu.contains(userMenuTrigger) &&
    !userMenuTrigger.closest('.user-menu-trigger')
  ) {
    isUserMenuVisible.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.layout-topbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 1000;
  background-color: var(--surface-card);
  height: 4rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  border-bottom: 1px solid var(--surface-border);
}

.layout-sidebar {
  position: fixed;
  top: 4rem;
  left: 0;
  width: 250px;
  height: calc(100vh - 4rem);
  background-color: var(--surface-card);
  box-shadow: 2px 0 4px rgba(0,0,0,0.1);
  transition: all 0.3s;
  z-index: 999;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  overflow-x: hidden;
  border-right: 1px solid var(--surface-border);
}

.layout-sidebar-collapsed {
  width: 4rem;
}

:deep(.router-link-active) {
  background-color: #f1f5f9;
  color: var(--primary-color);

  .pi {
    color: var(--primary-color);
  }
}

.layout-main {
  position: fixed;
  top: 4rem;
  left: 250px;
  right: 0;
  bottom: 0;
  background-color: var(--surface-ground);
  transition: all 0.3s;
  overflow-y: auto;
}

.layout-main-collapsed {
  left: 4rem;
}

.layout-main-full {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  padding: 1rem;
  overflow-y: auto;
}

.user-menu {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 0.5rem;
  border-radius: var(--border-radius);
  z-index: 1000;
  background-color: var(--surface-overlay);
  border: 1px solid var(--surface-border);
  box-shadow: var(--card-shadow);
}
</style>
