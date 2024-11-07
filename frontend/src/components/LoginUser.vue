<script setup lang="ts">
import InputGroup from 'primevue/inputgroup'
import InputGroupAddon from 'primevue/inputgroupaddon'
import InputText from 'primevue/inputtext'
import FloatLabel from 'primevue/floatlabel'
import Button from 'primevue/button'
import Panel from 'primevue/panel'

import { useRouter } from 'vue-router'
import { computed, ref } from 'vue'
import { useUserStore } from '@/store/module/user'
import { useToast } from 'primevue/usetoast'
import type { AxiosResponse } from 'axios'

const user = ref({
  username: '',
  password: ''
})

const isLoginDisabled = computed(() => !user.value.username || !user.value.password)

const userStore = useUserStore()

const router = useRouter()

const toast = useToast()

const errorMessage = ref<string>('')

const handleLogin = () => {
  userStore.login(user.value)
    .then(() => {
      toast.add({
        severity: 'success',
        summary: `Hello, ${userStore.userInfo.full_name}`,
        detail: 'You have successfully logged in.',
        life: 3000
      })
      // 跳转主页
      router.replace({ name: 'home' })
    })
    .catch((err: any) => {
      if (err.response && err.response.status === 401) {
        errorMessage.value = err.response.data.error
      } else {
        errorMessage.value = 'An error occurred, please try again.'
      }
      toast.add({
        severity: 'error',
        summary: 'Login failed.',
        detail: errorMessage.value,
        life: 3000
      })
    })
}
</script>

<template>
  <div class="flex flex-column login">
    <Panel header="Welcome to Nostalgia!">
      <div class="flex flex-column row-gap-5 login-form">
        <InputGroup>
          <InputGroupAddon>
            <i class="pi pi-user"></i>
          </InputGroupAddon>
          <FloatLabel>
            <InputText id="username" v-model="user.username" />
            <label for="username">Username</label>
          </FloatLabel>

        </InputGroup>

        <InputGroup>
          <InputGroupAddon>
            <i class="pi pi-lock"></i>
          </InputGroupAddon>
          <FloatLabel>
            <InputText id="password" type="password" v-model="user.password" :feedback="false" />
            <label for="password">Password</label>
          </FloatLabel>
        </InputGroup>

        <Button label="Login" :disabled="isLoginDisabled" @click="handleLogin" />
      </div>
    </Panel>
  </div>
</template>

<style scoped>
.login {
  width: 400px;
  margin: 10% auto;
  padding: 8px;
}

.login-form {
  margin-top: 8px;
}
</style>
