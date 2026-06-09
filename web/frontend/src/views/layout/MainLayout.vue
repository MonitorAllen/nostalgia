<script setup lang="ts">
import { Activity, Tags } from '@lucide/vue'
import GithubContributions from '@/components/GithubContributions.vue'
import CategoryList from '@/components/category/CategoryList.vue'
import ArchivePanel from '@/components/ui/ArchivePanel.vue'
</script>

<template>
  <div class="mx-auto w-full max-w-7xl px-4 py-6 md:py-8">
    <div class="grid gap-5 lg:grid-cols-[minmax(0,1fr)_22rem]">
      <RouterView />

      <aside class="space-y-5">
        <ArchivePanel title="近期活动" :icon="Activity">
          <GithubContributions />
        </ArchivePanel>
        <ArchivePanel title="分类索引" :icon="Tags">
          <CategoryList />
        </ArchivePanel>
      </aside>
    </div>
  </div>
</template>
