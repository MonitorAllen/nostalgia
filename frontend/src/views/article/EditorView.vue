<template>
  <div class="flex flex-row justify-content-center">
    <div class="flex flex-column gap-5">
      <div class="main-container">
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
      <div class="flex flex-column border-solid border-1 border-round-sm pl-4 pr-4 pt-5 pb-5 row-gap-5"
           style="border-color: #e2e8f0">
        <FloatLabel>
          <InputText id="title" v-model="post.title" class="w-full"></InputText>
          <label for="title">标题</label>
        </FloatLabel>
        <FloatLabel>
          <Textarea id="summary" class="w-full" rows="3" v-model="post.summary" autoResize />
          <label for="summary">简介</label>
        </FloatLabel>
        <div class="flex flex-column gap-3">
          <label style="position: relative; left: 0.75rem; line-height: 1px; font-size: 12px; color: #64748b; margin-top: -1rem;"
                 for="isPublish">是否发布</label>
          <ToggleButton id="isPublish" v-model="isPublish" onLabel="发布" offLabel="取消发布"
                        onIcon="pi pi-lock-open"
                        offIcon="pi pi-lock" class="" aria-label="Do you confirm" />
        </div>
        <Button label="保存" severity="success" @click="save"/>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import FloatLabel from 'primevue/floatlabel'
import Textarea from 'primevue/textarea'
import InputText from 'primevue/inputtext'
import ToggleButton from 'primevue/togglebutton'
import Button from 'primevue/button'

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
import { useToast } from 'primevue/usetoast';
import router from '@/router'
import { useUserStore } from '@/store/module/user'

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
  console.log("editorInstance", editorInstance)
  editor.value = editorInstance
  if (id !== '') {
    articleStore.getArticle(id as string).then((res: any) => {
      post.value = res.data
      editorData.value = post.value.content
      isPublish.value = post.value.is_publish
      console.log(res)
    })
  }
}

const editorComponent = ref(null)

/*editor.value.plugins.get('FileRepository').on('uploadComplete', (event, data) => {
  console.log('Image upload completed:', data);
});*/

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

.main-container {
  width: 1000px !important;
  font-family: 'Lato';
  width: fit-content;
  margin-left: auto;
  margin-right: auto;
}

.ck-content {
  font-family: 'Lato';
  line-height: 1.6;
  word-break: break-word;
}

.editor-container_classic-editor .editor-container__editor .ck-editor__editable_inline{
  min-height: 450px !important;
  min-width: 1000px;
  max-width: none !important;
}

.ck.ck-toolbar {
  width: auto !important;
  max-width: none !important;
}
</style>