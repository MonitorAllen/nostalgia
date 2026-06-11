<template>
  <div class="mx-auto w-full max-w-6xl px-4 py-8">
    <div class="flex flex-col gap-5">
      <div class="main-container w-full overflow-x-auto">
        <div class="editor-container editor-container_classic-editor" ref="editorContainerElement">
          <div class="editor-container__editor">
            <div ref="editorElement">
              <ckeditor
                v-if="isLayoutReady"
                v-model="editorData"
                :editor="ClassicEditor"
                :config="config"
                @ready="onEditorReady"
                ref="editorComponent"
              />
            </div>
          </div>
        </div>
      </div>
      <div class="archive-surface flex flex-col gap-5 rounded-archive p-5">
        <label class="block space-y-2">
          <span class="text-sm font-bold">标题</span>
          <AppInput id="title" v-model="post.title" />
        </label>
        <label class="block space-y-2">
          <span class="text-sm font-bold">简介</span>
          <textarea
            id="summary"
            v-model="post.summary"
            rows="3"
            class="w-full rounded-archive border border-border bg-surface px-4 py-3 text-sm text-foreground focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/20"
          />
        </label>
        <div class="flex items-center gap-3">
          <input id="isPublish" v-model="isPublish" type="checkbox" class="h-4 w-4 accent-[rgb(var(--color-accent))]" />
          <label class="text-sm font-bold" for="isPublish">发布文章</label>
        </div>
        <AppButton class="w-max" @click="save">保存</AppButton>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  ClassicEditor, AccessibilityHelp, Alignment, Autoformat, AutoImage, Autosave, BlockQuote, Bold,
  Code, CodeBlock, Essentials, FontBackgroundColor, FontColor, FontFamily, FontSize, FullPage,
  GeneralHtmlSupport, Heading, Highlight, HorizontalLine, HtmlComment, HtmlEmbed,
  ImageBlock, ImageCaption, ImageInline, ImageInsert, ImageInsertViaUrl, ImageResize, ImageStyle,
  ImageTextAlternative, ImageToolbar, ImageUpload, Indent, IndentBlock, Italic, Link, LinkImage,
  List, ListProperties, Markdown, MediaEmbed, Paragraph, PasteFromMarkdownExperimental,
  PasteFromOffice, RemoveFormat, SelectAll, ShowBlocks, SimpleUploadAdapter, SourceEditing,
  SpecialCharacters, SpecialCharactersArrows, SpecialCharactersCurrency, SpecialCharactersEssentials,
  SpecialCharactersLatin, SpecialCharactersMathematical, SpecialCharactersText, Strikethrough,
  Subscript, Superscript, Table, TableCaption, TableCellProperties, TableColumnResize,
  TableProperties, TableToolbar, TextTransformation, TodoList, Underline, Undo, type EditorConfig
} from 'ckeditor5'

import translations from 'ckeditor5/translations/zh-cn.js'

import { Ckeditor } from '@ckeditor/ckeditor5-vue'

import 'ckeditor5/ckeditor5.css'
import 'ckeditor5-premium-features/ckeditor5-premium-features.css'

import { ref, onMounted, type Ref } from 'vue'
import { useArticleStore } from '@/store/module/article'
import { useToast } from '@/composables/useToast';
import router from '@/router'
import { useUserStore } from '@/store/module/user'
import AppButton from '@/components/ui/AppButton.vue'
import AppInput from '@/components/ui/AppInput.vue'

const editor = ref<ClassicEditor|null>(null)
const config: Ref<EditorConfig>= ref({})
const editorData = ref('')
const isLayoutReady = ref(false)

let { id } = defineProps<{
  id?: string
}>()

let post = ref<Post>({
  id: '',
  title: '',
  summary: '',
  content: '',
  views: 0,
  likes: 0,
  is_publish: false,
  owner: '',
  create_at: '',
  update_at: '',
  delete_at: ''
})
let isPublish = ref(false)

const articleStore = useArticleStore()
const userStore = useUserStore()
const toast = useToast();

// 当编辑器准备好时的回调
const onEditorReady = (editorInstance: ClassicEditor) => {
  editor.value = editorInstance
  if (id !== '') {
    articleStore.getArticle(id as string).then((res: any) => {
      post.value = res.data
      editorData.value = post.value.content
      isPublish.value = post.value.is_publish
    })
  }
}

const editorComponent = ref(null)

const save = () => {
  if (id != "") {
    post.value.content = editorData.value
    post.value.is_publish = isPublish.value
    articleStore.updateArticle(post.value).then((res) => {
      toast.add({ severity: 'success', summary: 'Success', detail: "保存成功", life: 2500 })
    }).catch((err) => {
      toast.add({ severity: 'error', summary: 'Error', detail: err.response.data.error, life: 2500})
    })
  } else {
    post.value.content = editorData.value
    post.value.is_publish = isPublish.value
    articleStore.createArticle(post.value).then((res) => {
      toast.add({ severity: 'success', summary: 'Success', detail: "保存成功" })
    }).catch((err) => {
      toast.add({ severity: 'error', summary: 'Error', detail: err.response.data.error, life: 2500})
    })
  }
}


onMounted(async () => {
  if (id === '') {
    post.value.title = '新建文章'
    post.value.is_publish = false
    await articleStore.createArticle(post.value).then((res: any) => {
      id = res.data.id
      router.replace(`/article/edit/${id}`)
      // toast.add({ severity: 'error', summary: 'Error', detail: res.data.error, life: 2500})
    }).catch((err) => {
      toast.add({ severity: 'error', summary: 'Error', detail: err.response.data.error, life: 2500})
    })
  }
  
  const baseUrl = import.meta.env.VITE_APP_BASE_URL

  config.value = {
    toolbar: {
      items: [
        'undo', 'redo', '|', 'sourceEditing', 'showBlocks', '|', 'heading', '|',
        'fontSize', 'fontFamily', 'fontColor', 'fontBackgroundColor', '|',
        'bold', 'italic', 'underline', 'strikethrough', 'subscript', 'superscript', 'code', 'removeFormat',
        '|', 'specialCharacters', 'horizontalLine', 'link', 'insertImage', 'mediaEmbed',
        'insertTable', 'highlight', 'blockQuote', 'codeBlock', 'htmlEmbed', '|', 'alignment',
        '|', 'bulletedList', 'numberedList', 'todoList', 'outdent', 'indent'
      ],
      shouldNotGroupWhenFull: true
    },
    plugins: [
      AccessibilityHelp,
      Alignment,
      Autoformat,
      AutoImage,
      Autosave,
      BlockQuote,
      Bold,
      Code,
      CodeBlock,
      Essentials,
      FontBackgroundColor,
      FontColor,
      FontFamily,
      FontSize,
      FullPage,
      GeneralHtmlSupport,
      Heading,
      Highlight,
      HorizontalLine,
      HtmlComment,
      HtmlEmbed,
      ImageBlock,
      ImageCaption,
      ImageInline,
      ImageInsert,
      ImageInsertViaUrl,
      ImageResize,
      ImageStyle,
      ImageTextAlternative,
      ImageToolbar,
      ImageUpload,
      Indent,
      IndentBlock,
      Italic,
      Link,
      LinkImage,
      List,
      ListProperties,
      Markdown,
      MediaEmbed,
      Paragraph,
      PasteFromMarkdownExperimental,
      PasteFromOffice,
      RemoveFormat,
      SelectAll,
      ShowBlocks,
      SimpleUploadAdapter,
      SourceEditing,
      SpecialCharacters,
      SpecialCharactersArrows,
      SpecialCharactersCurrency,
      SpecialCharactersEssentials,
      SpecialCharactersLatin,
      SpecialCharactersMathematical,
      SpecialCharactersText,
      Strikethrough,
      Subscript,
      Superscript,
      Table,
      TableCaption,
      TableCellProperties,
      TableColumnResize,
      TableProperties,
      TableToolbar,
      TextTransformation,
      TodoList,
      Underline,
      Undo
    ],
    codeBlock: {
      languages: [
        { language: 'plaintext', label: 'Plain text' },
        { language: 'go', label: 'Golang' },
        { language: 'c', label: 'C' },
        { language: 'cs', label: 'C#' },
        { language: 'cpp', label: 'C++' },
        { language: 'css', label: 'CSS' },
        { language: 'diff', label: 'Diff' },
        { language: 'html', label: 'HTML' },
        { language: 'java', label: 'Java' },
        { language: 'javascript', label: 'JavaScript' },
        { language: 'php', label: 'PHP' },
        { language: 'python', label: 'Python' },
        { language: 'ruby', label: 'Ruby' },
        { language: 'typescript', label: 'TypeScript' },
        { language: 'xml', label: 'XML' } ]
    },
    fontFamily: {
      supportAllValues: true
    },
    fontSize: {
      options: [10, 12, 14, 'default', 18, 20, 22],
      supportAllValues: true
    },
    heading: {
      options: [
        {
          model: 'paragraph',
          title: 'Paragraph',
          class: 'ck-heading_paragraph'
        },
        {
          model: 'heading1',
          view: 'h1',
          title: 'Heading 1',
          class: 'ck-heading_heading1'
        },
        {
          model: 'heading2',
          view: 'h2',
          title: 'Heading 2',
          class: 'ck-heading_heading2'
        },
        {
          model: 'heading3',
          view: 'h3',
          title: 'Heading 3',
          class: 'ck-heading_heading3'
        },
        {
          model: 'heading4',
          view: 'h4',
          title: 'Heading 4',
          class: 'ck-heading_heading4'
        },
        {
          model: 'heading5',
          view: 'h5',
          title: 'Heading 5',
          class: 'ck-heading_heading5'
        },
        {
          model: 'heading6',
          view: 'h6',
          title: 'Heading 6',
          class: 'ck-heading_heading6'
        }
      ]
    },
    htmlSupport: {
      allow: [
        {
          name: /^.*$/,
          styles: true,
          attributes: true,
          classes: true
        }
      ]
    },
    image: {
      toolbar: [
        'toggleImageCaption',
        'imageTextAlternative',
        '|',
        'imageStyle:inline',
        'imageStyle:wrapText',
        'imageStyle:breakText',
        '|',
        'resizeImage'
      ]
    },
    simpleUpload: {
      // The URL that the images are uploaded to.
      uploadUrl: `${baseUrl}/upload/${id}`,

      // Enable the XMLHttpRequest.withCredentials property.
      withCredentials: true,

      // Headers sent along with the XMLHttpRequest to the upload server.
      headers: {
        Authorization: `Bearer ${userStore.token}`
      }
    },
    language: 'zh-cn',
    link: {
      addTargetToExternalLinks: true,
      defaultProtocol: 'https://',
      decorators: {
        toggleDownloadable: {
          mode: 'manual',
          label: 'Downloadable',
          attributes: {
            download: 'file'
          }
        }
      }
    },
    list: {
      properties: {
        styles: true,
        startIndex: true,
        reversed: true
      }
    },
    placeholder: '这一刻的想法……',
    table: {
      contentToolbar: ['tableColumn', 'tableRow', 'mergeTableCells', 'tableProperties', 'tableCellProperties']
    },
    translations: [translations]
  }

  isLayoutReady.value = true
})



interface Post {
  id: string
  title: string,
  summary: string,
  content: string,
  views: number,
  likes: number,
  is_publish: boolean,
  owner: string,
  create_at: string,
  update_at: string,
  delete_at: string,
}

</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Lato:ital,wght@0,400;0,700;1,400;1,700&display=swap');

.editor-container_classic-editor .editor-container__editor .ck-editor__editable_inline{
  min-height: 450px !important;
  min-width: min(1000px, calc(100vw - 2rem));
  max-width: none !important;
}

.ck.ck-toolbar {
  width: auto !important;
  max-width: none !important;
}
</style>
