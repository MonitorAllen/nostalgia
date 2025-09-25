<script setup lang="ts">
import Toast from 'primevue/toast'
</script>

<template>
  <Toast />
  <router-view />
</template>

<style>
html, body {
  margin: 0;
  padding: 0;
  height: 100%;
}

#app {
  height: 100%;
}
</style>
