<template>
    <div class="w-full flex-row justify-content-center ">
        <div class="flex flex-column text-color-secondary text-sm">
            <p>Copyright © {{ currentYear }} Nostalgia. Design by Primevue.</p>
            <p>粤ICP备2025431323号</p>
        </div>
    </div>
</template>

<script setup>
const date = new Date()
const currentYear = date.getFullYear()
</script>