<script setup lang="ts">
import AppButton from './AppButton.vue'

withDefaults(
  defineProps<{
    open: boolean
    title?: string
    description?: string
    confirmLabel?: string
    cancelLabel?: string
    danger?: boolean
  }>(),
  {
    title: '确认操作',
    description: '',
    confirmLabel: '确认',
    cancelLabel: '取消',
    danger: false,
  },
)

defineEmits<{
  cancel: []
  confirm: []
}>()
</script>

<template>
  <Teleport to="body">
    <div
      v-if="open"
      class="fixed inset-0 z-50 grid place-items-center bg-background/60 p-4 backdrop-blur-sm"
      role="dialog"
      aria-modal="true"
    >
      <div class="archive-surface w-full max-w-md rounded-archive p-5">
        <h2 class="m-0 text-lg font-bold text-foreground">{{ title }}</h2>
        <p v-if="description" class="mt-2 text-sm text-muted-foreground">{{ description }}</p>
        <slot />
        <div class="mt-5 flex justify-end gap-2">
          <AppButton variant="ghost" @click="$emit('cancel')">{{ cancelLabel }}</AppButton>
          <AppButton :variant="danger ? 'danger' : 'primary'" @click="$emit('confirm')">
            {{ confirmLabel }}
          </AppButton>
        </div>
      </div>
    </div>
  </Teleport>
</template>
