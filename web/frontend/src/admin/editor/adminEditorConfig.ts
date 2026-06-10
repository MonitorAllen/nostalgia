import {
  AccessibilityHelp,
  Alignment,
  Autoformat,
  AutoImage,
  BlockQuote,
  Bold,
  Code,
  CodeBlock,
  Essentials,
  Heading,
  HorizontalLine,
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
  Italic,
  Link,
  List,
  ListProperties,
  Paragraph,
  PasteFromOffice,
  RemoveFormat,
  SelectAll,
  Strikethrough,
  Table,
  TableCaption,
  TableCellProperties,
  TableColumnResize,
  TableProperties,
  TableToolbar,
  TextTransformation,
  TodoList,
  Underline,
  Undo,
  WordCount,
  type EditorConfig
} from 'ckeditor5'
import translations from 'ckeditor5/translations/zh-cn.js'
import { CODE_BLOCK_LANGUAGES } from '@/editor/contentLanguages'
import { ADMIN_EDITOR_TOOLBAR_ITEMS } from './adminEditorPolicy'

export const adminEditorConfig: EditorConfig = {
  toolbar: {
    items: [...ADMIN_EDITOR_TOOLBAR_ITEMS],
    shouldNotGroupWhenFull: false
  },
  plugins: [
    AccessibilityHelp,
    Alignment,
    Autoformat,
    AutoImage,
    BlockQuote,
    Bold,
    Code,
    CodeBlock,
    Essentials,
    Heading,
    HorizontalLine,
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
    Italic,
    Link,
    List,
    ListProperties,
    Paragraph,
    PasteFromOffice,
    RemoveFormat,
    SelectAll,
    Strikethrough,
    Table,
    TableCaption,
    TableCellProperties,
    TableColumnResize,
    TableProperties,
    TableToolbar,
    TextTransformation,
    TodoList,
    Underline,
    Undo,
    WordCount
  ],
  language: 'zh-cn',
  translations: [translations],
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
  table: {
    contentToolbar: [
      'tableColumn',
      'tableRow',
      'mergeTableCells',
      'tableProperties',
      'tableCellProperties'
    ]
  },
  codeBlock: {
    languages: [...CODE_BLOCK_LANGUAGES]
  },
  heading: {
    options: [
      { model: 'paragraph', title: '正文', class: 'ck-heading_paragraph' },
      { model: 'heading1', view: 'h1', title: '标题 1', class: 'ck-heading_heading1' },
      { model: 'heading2', view: 'h2', title: '标题 2', class: 'ck-heading_heading2' },
      { model: 'heading3', view: 'h3', title: '标题 3', class: 'ck-heading_heading3' },
      { model: 'heading4', view: 'h4', title: '标题 4', class: 'ck-heading_heading4' },
      { model: 'heading5', view: 'h5', title: '标题 5', class: 'ck-heading_heading5' },
      { model: 'heading6', view: 'h6', title: '标题 6', class: 'ck-heading_heading6' }
    ]
  },
  placeholder: '这一刻的想法……'
}
