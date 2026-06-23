<script setup lang="ts">
import { computed, ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    src?: string
    alt?: string
    variant?: 'detail' | 'list' | 'preview'
    fallbackSrc?: string
  }>(),
  {
    src: '',
    alt: '文章封面',
    variant: 'detail',
    fallbackSrc: ''
  }
)

const normalizeSrc = (src: string) => src.trim()
const resolveDisplaySrc = (src: string) => normalizeSrc(src) || props.fallbackSrc
const displaySrc = ref(resolveDisplaySrc(props.src))

watch(
  () => [props.src, props.fallbackSrc] as const,
  ([nextSrc]) => {
    displaySrc.value = resolveDisplaySrc(nextSrc)
  }
)

const containerClass = computed(() => {
  const base = 'relative block aspect-[16/9] w-full overflow-hidden bg-muted'
  if (props.variant === 'list') return `${base} h-full`
  if (props.variant === 'preview') return `${base} rounded-archive`
  return `${base} rounded-archive`
})

const imageClass = computed(() => {
  const base = 'h-full w-full object-cover transition duration-500'
  if (props.variant === 'list') return `${base} group-hover:scale-[1.03]`
  return base
})

const handleImageError = () => {
  if (props.fallbackSrc && displaySrc.value !== props.fallbackSrc) {
    displaySrc.value = props.fallbackSrc
    return
  }

  displaySrc.value = ''
}
</script>

<template>
  <span :class="containerClass">
    <img v-if="displaySrc" :src="displaySrc" :alt="alt" :class="imageClass" @error="handleImageError" />
    <span
      v-else
      class="absolute inset-0 bg-muted/70"
      aria-hidden="true"
    />
  </span>
</template>
