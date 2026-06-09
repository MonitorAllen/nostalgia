<script setup lang="ts">
import { Lock } from '@lucide/vue'
import { useRouter } from 'vue-router'
import AppButton from '@/components/ui/AppButton.vue'

const router = useRouter()
</script>

<template>
  <main class="grid min-h-screen place-items-center px-4">
    <section class="archive-surface w-full max-w-lg rounded-[1.1rem] p-8 text-center">
      <span class="archive-glass mx-auto grid h-20 w-20 place-items-center rounded-full">
        <Lock class="h-9 w-9 text-accent" />
      </span>
      <h1 class="mt-5 text-3xl font-black">访问受限 (403)</h1>
      <p class="mx-auto mt-3 max-w-md text-muted-foreground">
        抱歉，您没有权限访问该页面。如果您认为这是一个错误，请尝试重新登录。
      </p>
      <div class="mt-6 flex justify-center gap-3">
        <AppButton variant="secondary" @click="router.push('/')">回到首页</AppButton>
        <AppButton @click="router.push('/login')">去登录</AppButton>
      </div>
    </section>
  </main>
</template>
