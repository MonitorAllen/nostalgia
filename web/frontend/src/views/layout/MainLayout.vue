<template>
  <div class="w-full px-3 mx-auto main pb-3">
    <div class="grid mt-3">

      <div class="col-12 lg:col-8 p-0">
        <router-view />
      </div>

      <div class="col-12 lg:col-4 px-0 pt-0 lg:px-2 mt-3 lg:mt-0">
        <div class="flex flex-column gap-3">
          <div class="flex flex-column border-round-md shadow-1 surface-card overflow-hidden">
            <div class="flex justify-content-start align-items-center gap-2 p-2 border-bottom-1 surface-border theme-header-bg">
              <span class="pi pi-github"></span>
              <span class="font-bold">近期活动</span>
            </div>
            <div class="p-3">
              <GithubContributions></GithubContributions>
            </div>
          </div>

          <div class="flex flex-column border-round-md shadow-1 surface-card overflow-hidden">
            <div class="flex justify-content-start align-items-center gap-2 p-2 border-bottom-1 surface-border theme-header-bg">
              <span class="pi pi-tag"></span>
              <span class="font-bold">分类列表</span>
            </div>
            <div class="p-3">
              <CategoryList></CategoryList>
            </div>
          </div>

        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import GithubContributions from "@/components/GithubContributions.vue";
import CategoryList from "@/components/category/CategoryList.vue";
</script>

<style scoped>
/* 响应式宽度限制 */
@media (min-width: 576px) { .main { max-width: 540px; } }
@media (min-width: 768px) { .main { max-width: 720px; } }
@media (min-width: 992px) { .main { max-width: 960px; } }
@media (min-width: 1200px) { .main { max-width: 1140px; } }

.theme-header-bg {
  background-color: var(--green-200);
}
</style>