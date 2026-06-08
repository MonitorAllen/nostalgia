<script setup lang="ts">
import { Compass } from '@lucide/vue'
import { useRouter } from 'vue-router'
import AppButton from '@/components/ui/AppButton.vue'

const router = useRouter()
</script>

<template>
  <main class="grid min-h-screen place-items-center px-4">
    <section class="archive-surface w-full max-w-lg rounded-[1.1rem] p-8 text-center">
      <span class="archive-glass mx-auto grid h-20 w-20 place-items-center rounded-full">
        <Compass class="h-9 w-9 text-accent" />
      </span>
      <p class="mt-5 text-sm font-black uppercase text-muted-foreground">404</p>
      <h1 class="mt-2 text-3xl font-black">页面未找到</h1>
      <p class="mx-auto mt-3 max-w-md text-muted-foreground">
        这个页面可能已被移动、重命名，或者暂时不可用。
      </p>
      <div class="mt-6 flex justify-center gap-3">
        <AppButton variant="secondary" @click="router.go(-1)">返回上一页</AppButton>
        <AppButton @click="router.push('/')">回到首页</AppButton>
      </div>
    </section>
  </main>
</template>
