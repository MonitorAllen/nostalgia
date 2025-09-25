<template>
    <div class="flex flex-row justify-content-center">
      <div class="flex flex-column gap-5">
        <div class="main-container mx-auto">
          <div class="editor-container editor-container_classic-editor " ref="editorContainerElement">
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
        <div class="flex flex-column border-solid border-1 border-round-sm px-4 py-5 row-gap-5"
            style="border-color: #e2e8f0">
          <FloatLabel>
            <InputText id="title" v-model="article!.title" class="w-full"></InputText>
            <label for="title">标题</label>
          </FloatLabel>
          <FloatLabel>
            <Textarea id="summary" class="w-full" rows="3" v-model="article!.summary" autoResize />
            <label for="summary">简介</label>
          </FloatLabel>
          <FloatLabel>
            <Select id="category" name="category_id" v-model="article.category_id" :options="categories" optionLabel="name" optionValue="id"
            class="w-full"/>
            <label for="category" class="font-semibold w-24">分类</label>
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
import MyUploadAdapter from '@/util/uploadAdapter'

import { ref, onMounted, type Ref } from 'vue'
import {
  fetchArticleById,
  createArticle,
  updateArticle,
  type UpdateArticleRequest
} from '@/api/articles'
import { useToast } from 'primevue/usetoast';
import router from '@/router'
import type { Article } from '@/stores/article'
import {listAllCategories} from "@/api/category.ts";
import type {Category} from "@/types/category.ts";
import Select from "primevue/select";

const editor = ref<ClassicEditor|null>(null)
const config: Ref<EditorConfig>= ref({})
const editorData = ref('')
const isLayoutReady = ref(false)

const props = defineProps({
  id: {
    type: String,
    default: ''
  }
})

const article = ref<Article>({
  id: '',
  title: '',
  summary: '',
  content: '',
  likes: 0,
  views: 0,
  is_publish: false,
  created_at: '',
  updated_at: '',
  owner: '',
  category_id: '',
  category_name: '',
})
const isPublish = ref(false)

const categories = ref<Category[]>([])

const toast = useToast();

// 当编辑器准备好时的回调
const onEditorReady = async (editorInstance: ClassicEditor) => {
  editor.value = editorInstance
}

const editorComponent = ref(null)

const save = async () => {
  if (article.value && props.id != '') {
    article.value.content = editorData.value
    article.value.is_publish = isPublish.value

    const updateArticleRequest: UpdateArticleRequest = {
      id: article.value.id,
      title: article.value.title,
      summary: article.value.summary,
      content: article.value.content,
      is_publish: article.value.is_publish,
      category_id: parseInt(article.value.category_id)
    }

    try {
      const resp = await updateArticle(updateArticleRequest)
      article.value = resp.data.article
      toast.add({ severity: 'success', summary: 'Success', detail: "保存成功", life: 2500 })
    } catch(error: any) {
      toast.add({ severity: 'error', summary: 'Error', detail: '保存失败：' + error.response.data.error, life: 2500})
    }
  } else {
    toast.add({ severity: 'error', summary: 'Error', detail: "文章ID不存在", life: 2500})
  }
}

function CustomUploadAdapterPlugin(editor: any) {
    editor.plugins.get('FileRepository').createUploadAdapter = (loader: any) => {
    return new MyUploadAdapter(loader, props.id!)
  }
}

onMounted(async () => {
  if (props.id === '') {
    try {
      const resp = await createArticle({title: '新建文章', "summary": '', 'is_publish': false})
      article.value = resp.data.article
      await router.replace(`/article/edit/${resp.data.article.id}`)
    } catch(error: any)  {
      toast.add({ severity: 'error', summary: 'Error', detail: '创建文章失败：' + error.response.data.error, life: 3000})
    }
  } else {
    try {
      const resp = await fetchArticleById({id: props.id, needContent: true})
      article.value = resp.data.article
      editorData.value = article.value.content as string
      isPublish.value = article.value.is_publish
    } catch (error: any) {
      toast.add({ severity: 'error', summary: 'Error', detail: '获取文章信息失败: ' + error.response?.data?.message, life: 3000})
    }
  }

  try {
    const resp = await listAllCategories()
    categories.value = resp.data.categories
  } catch (error: any) {
    toast.add({ severity: 'error', summary: 'Error', detail: '获取分类失败：' + error.response.data.error, life: 3000})
  }

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
    extraPlugins: [
      CustomUploadAdapterPlugin],
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
    // simpleUpload: {
    //   // The URL that the images are uploaded to.
    //   uploadUrl: `${baseUrl}/v1/upload_file/`,

    //   // Enable the XMLHttpRequest.withCredentials property.
    //   withCredentials: true,

    //   // Headers sent along with the XMLHttpRequest to the upload server.
    //   headers: {
    //     Authorization: `Bearer ${authStore.access_token}`
    //   }
    // },
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
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Lato:ital,wght@0,400;0,700;1,400;1,700&display=swap');

.main-container {
  width: 1000px !important;
}



.editor-container_classic-editor .editor-container__editor .ck-editor__editable_inline{
  min-height: 450px !important;
  min-width: 1000px;
  max-width: none !important;
}

</style>
