<template>
    <div class="flex flex-wrap justify-content-center align-content-center h-full">
        <Card class="flex flex-column my-auto">
            <template #title>修改密码</template>
            <template #content>
                <div class="flex flex-column w-20rem">
                     <Form v-slot="$form" :resolver="resolver" :initialValues="initialValues" @submit="onFormSubmit" class="flex flex-col gap-4 w-full sm:w-64">
                        <div class="flex flex-column w-full gap-4">
                            <div>
                                <label for="oldPassword">原密码</label>
                                <Password id="oldPassword" name="oldPassword" placeholder="请输入原密码" :feedback="false" fluid />
                                <template v-if="$form.oldPassword?.invalid">
                                    <Message v-for="(error, index) of $form.oldPassword.errors" :key="index" severity="error" size="small" variant="simple">{{ error.message }}</Message>
                                </template>
                            </div>
                            <div>
                                <label for="newPassword">新密码</label>
                                <Password id="newPassword" name="newPassword" placeholder="请输入新密码" :feedback="false" fluid />
                                <template v-if="$form.newPassword?.invalid">
                                    <Message v-for="(error, index) of $form.newPassword.errors" :key="index" severity="error" size="small" variant="simple">{{ error.message }}</Message>
                                </template>
                            </div>
                            <div>
                            <label for="confirmPassword">新密码</label>
                                <Password id="confirmPassword" name="confirmPassword" placeholder="再次输入新密码" :feedback="false" fluid />
                                <template v-if="$form.confirmPassword?.invalid">
                                    <Message v-for="(error, index) of $form.confirmPassword.errors" :key="index" severity="error" size="small" variant="simple">{{ error.message }}</Message>
                                </template>
                            </div>
                            <Button type="submit" severity="secondary" label="提交" />
                        </div>
                    </Form>
                </div>
            </template>
        </Card>
    </div>
</template>

<script lang="ts" setup>
import Card from 'primevue/card';
import Password from 'primevue/password';
import Button from 'primevue/button';
import Message from 'primevue/message';


import { Form } from '@primevue/forms';
import { ref } from 'vue';
import { zodResolver } from '@primevue/forms/resolvers/zod';
import { useToast } from "primevue/usetoast";
import { z } from 'zod';

const toast = useToast();
const initialValues = ref({
    oldPassword: '',
    newPassword: '',
    confirmPassword: ''
});
const resolver = ref(zodResolver(
    z.object({
        oldPassword: z
            .string()
            .min(5, { message: 'Minimum 5 characters.' })
            .max(16, { message: 'Maximum 16 characters.' }),
        newPassword: z
        .string()
        .min(5, { message: 'Minimum 5 characters.' })
        .max(16, { message: 'Maximum 16 characters.' }),
        confirmPassword: z
            .string()
            .min(5, { message: 'Minimum 5 characters.' })
            .max(16, { message: 'Maximum 16 characters.' })
            
    }).refine(data => data.newPassword !== data.oldPassword, {
                message: '新密码不能与旧密码相同',
                path: ['newPassword']
            })
            .refine(data => data.newPassword === data.confirmPassword, {
                message: '两次密码输入不一致',
                path: ['confirmPassword']
            })
));

const onFormSubmit = (e: any) => {
    if (e.valid) {
        console.log(e.values)
        toast.add({ severity: 'success', summary: 'Form is submitted.', life: 3000 });
    }
};

</script>