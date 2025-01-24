<template>
    <div class="flex w-full h-100 align-content-center justify-content-center align-content-center">
        <div class="w-10 h-10rem surface-ground">
            
        </div>
    </div>
</template> 

<script lang="ts" setup>
import { ref } from 'vue'

const { token } = defineProps<{
    token: string
}>()

console.log('token: ' + token)

const isAuth = ref(false)





</script>