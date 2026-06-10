<script setup lang="ts">
import { computed, onMounted, ref, type Ref } from 'vue'
import { Ckeditor } from '@ckeditor/ckeditor5-vue'
import { ClassicEditor, Code, CodeBlock, type EditorConfig, Essentials, Paragraph } from 'ckeditor5'
import translations from 'ckeditor5/translations/zh-cn.js'
import 'ckeditor5/ckeditor5.css'
import 'ckeditor5/ckeditor5-content.css'

import { CODE_BLOCK_LANGUAGES } from '@/editor/contentLanguages'

const props = withDefaults(
  defineProps<{
    modelValue: string
    disabled?: boolean
  }>(),
  {
    disabled: false
  }
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
  submit: []
}>()

const editorData = computed({
  get: () => props.modelValue,
  set: (value: string) => emit('update:modelValue', value)
})

const isLayoutReady = ref(false)

const config: Ref<EditorConfig> = ref({
  toolbar: {
    items: ['undo', 'redo', '|', 'code', 'codeBlock'],
    shouldNotGroupWhenFull: true
  },
  plugins: [Code, CodeBlock, Essentials, Paragraph],
  placeholder: '写下评论，Ctrl/⌘ + Enter 提交',
  codeBlock: {
    languages: [...CODE_BLOCK_LANGUAGES]
  },
  language: 'zh-cn',
  translations: [translations]
})

const onEditorReady = (editorInstance: ClassicEditor) => {
  editorInstance.ui.view.editable.element?.classList.add(
    'reading-prose',
    'reading-prose--compact',
    'comment-editor-content'
  )
  editorInstance.editing.view.document.on('keydown', (event: unknown, data: unknown) => {
    const keyEvent = event as { stop: () => void }
    const keyData = data as { domEvent: KeyboardEvent; preventDefault: () => void }

    if (
      (keyData.domEvent.ctrlKey || keyData.domEvent.metaKey) &&
      keyData.domEvent.key === 'Enter'
    ) {
      keyData.preventDefault()
      keyEvent.stop()
      emit('submit')
    }
  })
}

onMounted(() => {
  isLayoutReady.value = true
})
</script>

<template>
  <div id="comment-editor" class="overflow-hidden rounded-archive border border-border">
    <Ckeditor
      v-if="isLayoutReady"
      v-model="editorData"
      :editor="ClassicEditor"
      :config="config"
      :disabled="disabled"
      @ready="onEditorReady"
    />
  </div>
</template>

<style scoped>
:deep(.ck-editor__editable_inline) {
  min-height: 8rem;
  background: rgb(var(--color-surface));
  color: rgb(var(--color-foreground));
}

:deep(.ck-toolbar) {
  background: rgb(var(--color-surface-raised)) !important;
  border-color: rgb(var(--color-border)) !important;
}

:deep(.ck-content) {
  background: rgb(var(--color-surface)) !important;
}
</style>
