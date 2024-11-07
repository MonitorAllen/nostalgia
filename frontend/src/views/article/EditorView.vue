<template>
  <div class="flex flex-row justify-content-center">
    <div class="flex flex-column gap-5" style="width: 1000px;">

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
          <label for="title">Title</label>
        </FloatLabel>
        <FloatLabel>
          <Textarea id="summary" class="w-full" rows="3" v-model="post.summary" autoResize />
          <label for="summary">Summary</label>
        </FloatLabel>
        <div class="flex flex-column gap-3">
          <label class="ml-2.5" style="line-height: 1px; font-size: 12px; color: #64748b; margin-top: -1rem;"
                 for="isPublish">IsPublish</label>
          <ToggleButton id="isPublish" v-model="isPublish" onLabel="Publish" offLabel="Unpublish"
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

import 'ckeditor5/ckeditor5.css'

import { Ckeditor } from '@ckeditor/ckeditor5-vue'

import 'ckeditor5/ckeditor5.css'
import 'ckeditor5-premium-features/ckeditor5-premium-features.css'

import { ref, onMounted, type Ref } from 'vue'
import { useArticleStore } from '@/store/module/article'
import { useToast } from 'primevue/usetoast';

const editor = ref<ClassicEditor|null>(null)
const config: Ref<EditorConfig>= ref({})
const editorData = ref('')
const isLayoutReady = ref(false)

const { id } = defineProps<{
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
      toast.add({ severity: 'success', summary: 'Success Message', detail: "保存成功" })
    }).catch((err) => {
      toast.add({ severity: 'error', summary: 'Error Message', detail: err.response.data.error, life: 1.5})
    })
  } else {
    post.value.content = editorData.value
    post.value.is_publish = isPublish.value
    articleStore.createArticle(post.value).then((res) => {
      toast.add({ severity: 'success', summary: 'Success Message', detail: "保存成功" })
    }).catch((err) => {
      toast.add({ severity: 'error', summary: 'Error Message', detail: err.response.data.error, life: 1.5})
    })
  }
}


onMounted(() => {
  // 设置 CKEditor 编辑区域的最小高度
  /*const editableElement = editorComponent.value.$el.querySelector('.ck-editor__editable_inline')
  if (editableElement) {
    editableElement.style.minHeight = '450px' // 设置最小高度为 300px
  }*/
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
      uploadUrl: `${baseUrl}/upload`,

      // Enable the XMLHttpRequest.withCredentials property.
      withCredentials: true,

      // Headers sent along with the XMLHttpRequest to the upload server.
      headers: {
        // 'X-CSRF-TOKEN': 'CSRF-Token',
        Authorization: 'Bearer <JSON Web Token>'
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
}
</style>