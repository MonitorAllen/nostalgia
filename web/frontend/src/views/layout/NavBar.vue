<template>
  <div class="flex flex-row min-w-min h-4rem relative shadow-1 justify-content-between">
    <div class="flex flex-row nav-r-box gap-1">
      <router-link to="/" class="flex flex-row align-items-center ml-2 sm:ml-6 cursor-pointer select-none">
        <img src="../../../public/logo.svg" alt="Nostalgia Logo" class="h-2rem">
      </router-link>
      <a class="flex align-items-center lg:hidden" @click="toggleNav" ref="navIcon">
        <i class="pi pi-bars"></i>
      </a>
      <div
        :class="[isNavOpen ? 'flex absolute w-full bg-white nav-open h-auto shadow-1' : '']"
        class="flex-column align-content-between hidden lg:shadow-none lg:flex lg:flex-row lg:static menu-box"
        ref="leftPanelRef"
      >
        <ul class="flex flex-column hidden lg:flex-row  menu-box m-0 pl-0"
            :class="[isNavOpen ? 'flex w-full' :  '']">
          <li v-for="(item, index) in navItems" :key="index" class="h-full">
            <router-link
              :class="[isNavOpen ? 'py-3 md:border-bottom-none down' : '', $route.path === item.route ? 'border-left-2' : '', { active: $route.path === item.route }]"
              class="flex h-full py-0 px-4 lg:border-left-none lg:border-bottom-2 cursor-pointer align-items-center gap-2 menu_a column-a"
              :to="{ path: item.route }"
            >
              <span :class="item.icon"></span>
              <span>{{ item.label }}</span>
            </router-link>
          </li>
        </ul>
      </div>
    </div>
    <a class="flex align-items-center p-2 sm:hidden relative" @click="toggleProfile" ref="toggleProfileRef">
      <i class="pi pi-ellipsis-v"></i>
    </a>
    <div
      v-if="isRightSidePanelOpen"
      ref="rightSidePanelOpenRef"
      class="flex flex-column md:flex-row absolute sm:relative z-5 bg-white shadow-1 sm:shadow-none
      w-full sm:w-auto h-auto align-items-start sm:align-items-center
      top-100 sm:top-0 left-0 sm:left-auto">
      <ul class="flex w-full flex-column sm:flex-row sm:gap-3 row-gap-2 align-items-start sm:align-items-center sm:p-0 m-0 h-full">
        <li class="w-full sm:w-auto p-2 sm:p-0">
          <InputText placeholder="Search" type="text" class="w-full sm:w-auto" />
        </li>
        <li class="relative h-full w-full" v-if="userStore.userInfo" ref="profileRef" @click="toggleDropdown">
          <a
             class="flex w-full h-full p-2 sm:py-0 sm:px-0 sm:px-4 column-a sm:border-left-none sm:border-bottom-2 cursor-pointer align-items-center gap-2 menu_a"
          >
            <img src="/images/go.png" class="h-2rem mr-0" alt="" >
            <span>{{userStore.userInfo.username}}</span>
            <i class="pi pi-angle-down"></i>
          </a>
          <div
            class="relative sm:absolute w-full top-100 left-0 shadow-none sm:shadow-1"
            v-if="dropdownVisible"
            ref="dropdownRef"
          >
            <ul class="p-0 m-0">
              <li>
                <a class="inline-flex w-full p-3 align-items-center gap-3 column-a sm:bg-white cursor-pointer" @click="userStore.logout()">
                  <i class="pi pi-sign-out"></i>
                  <span>退出登录</span>
                </a>
              </li>
            </ul>
          </div>
        </li>
        <li class="flex flex-column sm:flex-row h-full row-gap-2" v-else>
          <router-link
            class="flex h-full py-0 px-3 sm:border-left-none sm:border-bottom-2 cursor-pointer align-items-center gap-2 menu_a column-a"
            to="/login"
          >
            <span>登录</span>
          </router-link>
          <router-link
            class="flex h-full py-0 px-3 sm:border-left-none sm:border-bottom-2 cursor-pointer align-items-center gap-2 menu_a column-a"
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
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useRoute } from 'vue-router'
import { useUserStore } from '@/store/module/user.ts'

import InputText from 'primevue/inputtext'

const $route = useRoute()
const userStore = useUserStore()

// 控制下拉框的可见性

const dropdownVisible = ref(false)
const profileRef = ref(null)
const dropdownRef = ref(null)

const screenWidth = ref(window.innerWidth)

// 切换下拉框的可见性
const toggleDropdown = (event) => {
  dropdownVisible.value = !dropdownVisible.value
  if (isNavOpen.value) {
    isNavOpen.value = false
  }
  event.stopPropagation()  // 防止事件冒泡导致点击头像时关闭下拉框
}

const toggleProfileRef = ref(null)
const rightSidePanelOpenRef = ref(null)
const leftPanelRef = ref(null)

// 点击其他地方关闭下拉框
const handleClickOutside = (event) => {
  if (
    navIcon.value &&
    leftPanelRef.value &&
    !leftPanelRef.value.contains(event.target) &&
    !navIcon.value.contains(event.target)
  ) {
    isNavOpen.value = false
  }

  if (
    dropdownVisible.value &&
    !profileRef.value.contains(event.target) &&
    !dropdownRef.value.contains(event.target)
  ) {
    dropdownVisible.value = false
  }

  if (isRightSidePanelOpen.value && !toggleProfileRef.value.contains(event.target)
    && !rightSidePanelOpenRef.value.contains(event.target)
    && screenWidth.value < 576
  ) {
    isRightSidePanelOpen.value  = false
  }
}

const navItems = ref([
  {
    label: '主页',
    icon: 'pi pi-home',
    route: '/'
  },
])

const menu = ref()

const clickAvatarBox = (event) => {
  menu.value.toggle(event)
}

const isNavOpen = ref(false) // 控制导航栏显示状态

const toggleNav = () => {
  isNavOpen.value = !isNavOpen.value // 切换导航栏状态

  if (isRightSidePanelOpen.value && screenWidth.value < 576) {
    isRightSidePanelOpen.value = !isRightSidePanelOpen.value
  }
}

const navIcon = ref(null)

const isRightSidePanelOpen = ref(false)

const toggleProfile = () => {
  isRightSidePanelOpen.value = !isRightSidePanelOpen.value

  if (isNavOpen.value) {
    isNavOpen.value = !isNavOpen.value
  }
}

// 处理屏幕宽度变化的函数
const handleResize = () => {
  screenWidth.value = window.innerWidth
  isRightSidePanelOpen.value = window.innerWidth > 576 // 设置移动端的逻辑
  dropdownVisible.value = false
}

onMounted(() => {
  // 创建 MutationObserver 实例
  const observer = new MutationObserver((mutations) => {
    mutations.forEach((mutation) => {
      if (mutation.attributeName === 'style') {
        const currentDisplay = getComputedStyle(mutation.target).display
        if (currentDisplay === 'none') {
          console.log('Element is hidden')
          isNavOpen.value = false
        } else {
          console.log('Element is shown')
          // 元素显示时触发的逻辑
        }
      }
    })
  })

  if (window.innerWidth > 576) {
    isRightSidePanelOpen.value = true
  }

  // 开始监听 box 元素的属性变化
  observer.observe(navIcon.value, {
    attributes: true, // 监听属性变化
    attributeFilter: ['style', ':class'] // 只监听 style 属性
  })

  // 在 mounted 时添加点击事件监听器，检查点击是否在下拉框外部
  document.addEventListener('click', handleClickOutside)
  window.addEventListener('resize', handleResize)

  onBeforeUnmount(() => {
    observer.disconnect() // 组件卸载时停止监听
    document.removeEventListener('click', handleClickOutside)
    window.removeEventListener('resize', handleResize)
  })

})

</script>

<style scoped>
.menu_a:hover {
  border-color: color-mix(in srgb, #64748b calc(100% * 1), transparent);
}

.menu_a {
  color: #334155 !important;
  text-decoration: none;
  border-bottom: 0 solid transparent; /* 默认透明的边框 */
  transition: border-color 0.3s ease; /* 边框颜色渐变效果 */
}

.active {
  border-bottom: 0 solid color-mix(in srgb, #020617 calc(100% * 1), transparent); /* 悬停时显示底部边框颜色 */; /* 当前路由匹配时显示底部边框 */
}

.down {
  animation: slide-down 0.5s ease;
}

/* 导航栏下拉动画 */
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
  border-left: 2px solid transparent; /* 默认透明的边框 */
  transition: border-color 0.3s ease; /* 边框颜色渐变效果 */
}

.column-a:hover {
  border-color: color-mix(in srgb, #64748b calc(100% * 1), transparent);
}

.column-a-hover {
  border-left: 2px solid transparent; /* 默认透明的边框 */
  transition: border-color 0.3s ease; /* 边框颜色渐变效果 */
}

.avatar-box:hover {
  cursor: pointer;
  transition: background 0.2s, color 0.2s;
  border-radius: 4px;
  color: #334155;
}

.nav-open {
  top: 100%;
  left: 0;
  z-index: 1004;
}


.login-bar {
  a {
    cursor: pointer;
    border-radius: 6px
  }

  a:hover {
    background: #f1f5f9;
  }
}

ul li{
  list-style: none;
}
</style>
