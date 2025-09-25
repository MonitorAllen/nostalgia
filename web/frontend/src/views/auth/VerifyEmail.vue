<template>
  <div class="flex justify-content-center align-items-center" style="height: 100vh;">
    <div class="">
      <div>
        <p v-if="!isAuth"><i class="pi pi-spin pi-spinner"></i> 账号验证中……</p>
        <p v-if="isAuth && isVerified"><i class="pi pi-check" style="color: green"></i> 账号验证成功</p>
        <p v-if="isAuth && !isVerified"><i class="pi pi-times" style="color: red"></i> 账号验证失败</p>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import {onBeforeMount, ref} from 'vue'
import {useUserStore} from "@/store/module/user";
import {useToast} from "primevue/usetoast";
import router from "@/router";

const {email_id, secret_code} = defineProps<{
  email_id: number,
  secret_code: string
}>()

const isAuth = ref(false)
const isVerified = ref(false)
const timer = ref<number|null>(null)

if (email_id !== 0 && secret_code !== '') {
  const userStore = useUserStore()
  const toast = useToast()
  userStore.verifyEmail(email_id, secret_code)
      .then(() => {
        isAuth.value = true
        isVerified.value = true

        timer.value = window.setTimeout(() => {
          router.replace({ name: 'home' })
        }, 2000)
      })
      .catch((err) => {
            toast.add({
              severity: 'error',
              summary: 'Verify failed',
              detail: err.response.data.error,
              life: 3000
            })
            isAuth.value = true
          }
      )
}

onBeforeMount(() => {
  if (timer.value) {
    clearTimeout(timer.value)
    timer.value = null
  }
})
</script>