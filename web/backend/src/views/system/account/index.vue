<template>
  <div class="account-container card">
    <el-tabs v-model="activeName">
      <el-tab-pane label="基本信息" name="base">
        <el-form :model="adminInfo" label-width="100px" style="max-width: 500px; padding: 20px">
          <el-form-item label="用户名">
            <el-input v-model="adminInfo.username" disabled />
          </el-form-item>
          <el-form-item label="全名">
            <el-input v-model="adminInfo.full_name" />
          </el-form-item>
          <el-form-item label="邮箱">
            <el-input v-model="adminInfo.email" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="updateInfo">保存修改</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>
      <el-tab-pane label="修改密码" name="password">
        <el-form :model="passwordForm" label-width="100px" style="max-width: 500px; padding: 20px">
          <el-form-item label="原密码" required>
            <el-input v-model="passwordForm.old_password" type="password" show-password />
          </el-form-item>
          <el-form-item label="新密码" required>
            <el-input v-model="passwordForm.new_password" type="password" show-password />
          </el-form-item>
          <el-form-item label="确认新密码" required>
            <el-input v-model="passwordForm.confirm_password" type="password" show-password />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="updatePassword">确认修改</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts" name="account">
import { ref, onMounted } from 'vue'
import http from '@/api'
import { ElMessage } from 'element-plus'

const activeName = ref('base')
const adminInfo = ref({
  username: '',
  full_name: '',
  email: '',
})

const passwordForm = ref({
  old_password: '',
  new_password: '',
  confirm_password: '',
})

const fetchAdminInfo = async () => {
  const res: any = await http.get('/admin/info')
  adminInfo.value = res.data.admin
}

const updateInfo = async () => {
  await http.patch('/admin', {
    full_name: adminInfo.value.full_name,
    email: adminInfo.value.email,
  })
  ElMessage.success('基本信息更新成功')
}

const updatePassword = async () => {
  if (passwordForm.value.new_password !== passwordForm.value.confirm_password) {
    return ElMessage.warning('两次输入的密码不一致')
  }
  await http.patch('/admin', {
    old_password: passwordForm.value.old_password,
    new_password: passwordForm.value.new_password,
  })
  ElMessage.success('密码修改成功')
  passwordForm.value = { old_password: '', new_password: '', confirm_password: '' }
}

onMounted(() => {
  fetchAdminInfo()
})
</script>
