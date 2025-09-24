<template>
  <div class="w-full px-3 mx-auto main">
    <div class="flex flex-wrap mt-3 bg-white">
      <router-view />
      <div class="block w-full relative px-2 mt-2 left-side">
        <div class="flex flex-column border-round-md shadow-1 mb-3">
          <div class="flex border-round-top-md justify-content-start align-items-center gap-2 left-side-menu-title">
            <span class="pi pi-github ml-1"></span>
            <span>近期活动</span>
          </div>
          <div class="p-2">
            <GithubContributions></GithubContributions>
          </div>
        </div>
        <div class="flex flex-column border-round-md shadow-1 mb-3">
          <div class="flex border-round-top-md justify-content-start align-items-center gap-2 left-side-menu-title">
            <span class="pi pi-tag ml-1"></span>
            <span>分类列表</span>
          </div>
          <div class="p-2">
            <CategoryList></CategoryList>
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
.left-side-menu-title {
  height: 2rem !important;
  background-color: var(--green-200);
}

@media (min-width: 576px) {
  .main {
    max-width: 540px;
  }
}

@media (min-width: 768px) {
  .main {
    max-width: 720px;
  }
}

@media (min-width: 992px) {
  .main {
    max-width: 960px;
  }

  .center {
    flex: 0 0 66.666667%;
    max-width: 66.666667%;
  }

  .left-side {
    flex: 0 0 33.333333%;
    max-width: 33.333333%;
  }
}

@media (min-width: 1200px) {
  .main {
    max-width: 1140px;
  }
}
</style>
