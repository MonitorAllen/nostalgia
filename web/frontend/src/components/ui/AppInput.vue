<script setup lang="ts">
import { computed } from 'vue'
import { cn } from '@/lib/utils'

const props = withDefaults(
  defineProps<{
    modelValue: string
    id?: string
    type?: string
    placeholder?: string
    disabled?: boolean
    class?: string
  }>(),
  {
    id: undefined,
    type: 'text',
    placeholder: '',
    disabled: false,
    class: '',
  },
)

defineEmits<{
  'update:modelValue': [value: string]
}>()

const classes = computed(() =>
  cn(
    'h-11 w-full rounded-full border border-border bg-surface px-4 text-sm text-foreground',
    'placeholder:text-muted-foreground/80 shadow-none transition-colors',
    'focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/18',
    'disabled:cursor-not-allowed disabled:opacity-60',
    props.class,
  ),
)
</script>

<template>
  <input
    :id="id"
    :type="type"
    :value="modelValue"
    :placeholder="placeholder"
    :disabled="disabled"
    :class="classes"
    @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
  />
</template>
