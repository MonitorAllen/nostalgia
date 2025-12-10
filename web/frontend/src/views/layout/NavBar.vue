<template>
  <div class="flex flex-row min-w-min h-4rem relative shadow-1 justify-content-between bg-white z-5">

    <div class="flex flex-row nav-r-box gap-1 h-full align-items-center">
      <router-link to="/" class="flex flex-row align-items-center ml-2 sm:ml-6 cursor-pointer select-none">
        <img src="/logo.svg" alt="Nostalgia Logo" class="h-2rem">
      </router-link>

      <a class="flex align-items-center lg:hidden ml-2 cursor-pointer p-2 hover:bg-gray-100 border-round transition-colors transition-duration-200"
         @click.stop="toggleLeftNav">
        <i class="pi pi-bars text-xl"></i>
      </a>

      <div
          :class="[isLeftNavOpen ? 'flex absolute w-full bg-white nav-open h-auto shadow-1 border-top-1 border-gray-100' : 'hidden']"
          class="flex-column align-content-between pl-1 lg:shadow-none lg:border-none lg:flex lg:flex-row lg:static menu-box lg:h-full"
          @click.stop
      >
        <ul class="flex flex-column lg:flex-row menu-box m-0 pl-0 w-full list-none h-full">
          <li v-for="(item, index) in navItems" :key="index" class="h-full">

            <a
                v-if="item.url"
                :href="item.url"
                :target="item.target"
                :class="[
                  'flex h-full py-3 lg:py-0 px-4 cursor-pointer align-items-center gap-2 menu_a column-a',
                  isLeftNavOpen ? 'down' : 'lg:border-left-none lg:border-bottom-2',
                  { active: $route.path === item.route }
                ]"
            >
              <span :class="item.icon"></span>
              <span class="relative">
                {{ item.label }}
                <i v-if="item.target === '_blank'"
                   class="pi pi-external-link absolute text-700"
                   style="font-size: 0.6rem; top: -4px; right: -10px;"></i>
              </span>
            </a>

            <router-link
                v-else
                :to="{ path: item.route}"
                :class="[
                  'flex h-full py-3 lg:py-0 px-4 cursor-pointer align-items-center gap-2 menu_a column-a',
                  isLeftNavOpen ? 'down' : 'lg:border-left-none lg:border-bottom-2',
                  { active: $route.path === item.route }
                ]"
            >
              <span :class="item.icon"></span>
              <span>{{ item.label }}</span>
            </router-link>
          </li>
        </ul>
      </div>
    </div>

    <a class="flex align-items-center p-2 sm:hidden relative cursor-pointer mr-2"
       @click.stop="toggleRightPanel">
      <i class="pi pi-ellipsis-v text-xl"></i>
    </a>

    <div
        v-if="isRightPanelOpen"
        class="flex flex-column md:flex-row absolute sm:relative z-5 bg-white shadow-1 sm:shadow-none
               w-full sm:w-auto h-auto align-items-start sm:align-items-center
               top-100 sm:top-0 left-0 sm:left-auto border-top-1 sm:border-none border-gray-100 sm:h-full"
        @click.stop
    >
      <ul class="flex flex-column w-full h-full m-0 pl-1 pb-1 row-gap-2 list-none align-items-start sm:flex-row sm:gap-3 sm:align-items-center sm:p-0">

        <li class="w-full md:w-18rem p-2 sm:p-0">
          <InputText
              v-model="searchValue"
              placeholder="搜索..."
              type="text"
              class="w-full"
              v-tooltip.focus.bottom="{
                  value: '使用空格隔开关键词',
                  pt: {
                      root: {
                        style: {
                          height: '8px'
                        }
                      },
                      text: {
                        style: {
                          fontSize: '12px',
                          letterSpacing: '0.1em',
                          padding: '5px'
                        }
                      }
                  }
              }"
              @keydown.enter="handleSearch"
          />
        </li>

        <li class="relative h-full w-full sm:w-auto" v-if="userStore.userInfo">
          <a
              class="flex w-full h-full p-2 sm:py-0 sm:px-0 sm:px-4 column-a sm:border-left-none sm:border-bottom-2 cursor-pointer align-items-center gap-2 menu_a"
              @click="toggleUserMenu"
              aria-haspopup="true"
              aria-controls="user_menu"
          >
            <img src="/images/go.png" class="h-2rem mr-0" alt="Avatar">
            <span>{{ userStore.userInfo.username }}</span>
            <i class="pi pi-angle-down"></i>
          </a>

          <Menu ref="userMenu" id="user_menu" :model="userMenuItems" :popup="true"/>
        </li>

        <li class="flex flex-column sm:flex-row w-full h-full row-gap-2 sm:max-w-max sm:gap-0" v-else>
          <router-link
              class="flex h-full py-2 sm:py-0 px-3 sm:border-left-none sm:border-bottom-2 cursor-pointer align-items-center gap-2 menu_a column-a"
              to="/login"
          >
            <span>登录</span>
          </router-link>
          <router-link
              class="flex h-full py-2 sm:py-0 px-3 sm:border-left-none sm:border-bottom-2 cursor-pointer align-items-center gap-2 menu_a column-a"
              to="/register"
          >
            <span>注册</span>
          </router-link>
        </li>
      </ul>
    </div>

  </div>
</template>

<script setup>
import {computed, onBeforeUnmount, onMounted, ref} from 'vue'
import {useRouter} from 'vue-router'
import {useUserStore} from '@/store/module/user.ts'
import InputText from 'primevue/inputtext'
import Menu from 'primevue/menu'

const router = useRouter()
const userStore = useUserStore()

// --- 状态控制 ---
const isLeftNavOpen = ref(false)
const isRightPanelOpen = ref(true)

const searchValue = ref('')
const userMenu = ref(null)

// --- 导航数据 ---
const navItems = ref([
  {label: '主页', icon: 'pi pi-home', route: '/', target: ''},
  {label: '工具', icon: 'pi pi-wrench', url: 'https://toolx.de5.net', target: '_blank'}
])

const userMenuItems = computed(() => [
  {
    label: '退出登录',
    icon: 'pi pi-sign-out',
    command: () => {
      userStore.logout()
    }
  }
])

// --- 方法 ---

const handleSearch = () => {
  const query = searchValue.value.trim()
  if (!query) return
  searchValue.value = ''
  router.push({path: '/search', query: {q: query}})
  // 搜索后自动收起
  if (window.innerWidth < 576) isRightPanelOpen.value = false
}

const toggleLeftNav = () => {
  isLeftNavOpen.value = !isLeftNavOpen.value
  if (isLeftNavOpen.value && window.innerWidth < 576) {
    isRightPanelOpen.value = false
  }
}

const toggleRightPanel = () => {
  isRightPanelOpen.value = !isRightPanelOpen.value
  if (isRightPanelOpen.value) {
    isLeftNavOpen.value = false
  }
}

const toggleUserMenu = (event) => {
  userMenu.value.toggle(event)
}

// 点击空白关闭菜单
const handleClickOutside = () => {
  const isMobile = window.innerWidth < 992
  const isSmallMobile = window.innerWidth < 576

  if (!isMobile) return

  // 这里的逻辑很简单：因为面板内部都有 @click.stop，
  // 所以只要代码跑到了这里，说明点击的一定是“面板外部”
  if (isLeftNavOpen.value) isLeftNavOpen.value = false
  if (isSmallMobile && isRightPanelOpen.value) isRightPanelOpen.value = false
}

const handleResize = () => {
  const width = window.innerWidth
  if (width >= 992) isLeftNavOpen.value = false
  if (width >= 576) {
    isRightPanelOpen.value = true
  } else {
    isRightPanelOpen.value = false
  }
}

onMounted(() => {
  handleResize()
  window.addEventListener('resize', handleResize)
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.menu_a {
  color: #334155 !important;
  text-decoration: none;
  border-bottom: 0 solid transparent;
  transition: border-color 0.3s ease, background-color 0.2s ease;
}

.menu_a:hover {
  border-color: color-mix(in srgb, #64748b calc(100% * 1), transparent);
}

.active {
  border-bottom: 0 solid color-mix(in srgb, #020617 calc(100% * 1), transparent);
}

.down {
  animation: slide-down 0.3s ease forwards;
}

@keyframes slide-down {
  0% {
    opacity: 0;
    transform: translateY(-10px);
  }
  100% {
    opacity: 1;
    transform: translateY(0);
  }
}

.column-a {
  border-left: 2px solid transparent;
  transition: border-color 0.3s ease;
}

.nav-open {
  top: 100%;
  left: 0;
  z-index: 1004;
}
</style>