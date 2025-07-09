<script setup lang="ts">
import { RouterView, useRoute } from 'vue-router'
import Toast from 'primevue/toast'
import NavBar from '@/views/layout/NavBar.vue'
import FooterView from './views/layout/FooterView.vue'

const route = useRoute()
</script>

<template>
  <div class="card flex justify-center">
    <Toast />
  </div>
  <div class="layout">
    <header>
      <div>
      <NavBar v-if="!route.meta.hideNavbar"></NavBar>
      </div>
    </header>
    <main class="content">
      <RouterView />
    </main>
    <footer class="footer">
      <FooterView></FooterView>
    </footer>
  </div>
</template>

<style scoped>
.layout {
  display: flex;
  flex-direction: column;
  min-height: 100vh; /* 视口高度 */
}

.content {
  flex: 1; /* 自动撑开剩余空间 */
}

.footer {
  background-color: #f8f8f8;
  padding: 16px;
  text-align: center;
}
</style>
