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
import type { RegisterRequest } from '@/types/request/user'

const user = ref<RegisterRequest>({
  username: '',
  email: '',
  full_name: '',
  password: ''
})

const isRegisterDisabled = computed(() => !user.value.email || !user.value.username || !user.value.full_name || !user.value.password)

const userStore = useUserStore()

const router = useRouter()

const toast = useToast()

const errorMessage = ref<string>('')

const handleRegister = () => {
  userStore.register(user.value)
    .then(() => {
      // 跳转主页
      router.replace({ name: 'home' })
    })
    .catch((err) => {
      if (err.response && err.response.status === 403) {
        errorMessage.value = 'Email or username already exists.'
      } else if (err.response && err.response.status === 400) {
        errorMessage.value = 'Parameters error.'
      } else {
        errorMessage.value = 'An error occurred, please try again.'
      }
      toast.add({
        severity: 'error',
        summary: 'Register failed.',
        detail: errorMessage.value,
        life: 3000
      })
    })
}
</script>

<template>
  <div class="flex flex-column register">
    <Panel header="Welcome to Nostalgia!">
      <div class="flex flex-column row-gap-5 register-form">
        <InputGroup>
          <InputGroupAddon>
            <i class="pi pi-user"></i>
          </InputGroupAddon>
          <FloatLabel>
            <InputText id="email" v-model="user.email" />
            <label for="username">Email</label>
          </FloatLabel>
        </InputGroup>
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
            <i class="pi pi-user"></i>
          </InputGroupAddon>
          <FloatLabel>
            <InputText id="full_name" v-model="user.full_name" />
            <label for="username">FullName</label>
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

        <Button label="Register" :disabled="isRegisterDisabled" @click="handleRegister" />
      </div>
    </Panel>
  </div>
</template>

<style scoped>
.register {
  width: 400px;
  margin: 10% auto;
  padding: 8px;
}

.register-form {
  margin-top: 8px;
}
</style>
