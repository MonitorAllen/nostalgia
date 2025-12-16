import {
  AccessibilityHelp, Alignment, Autoformat, AutoImage, Autosave, BlockQuote, Bold,
  Code, CodeBlock, Essentials, FontBackgroundColor, FontColor, FontFamily, FontSize,
  FullPage, GeneralHtmlSupport, Heading, Highlight, HorizontalLine, HtmlComment,
  HtmlEmbed, ImageBlock, ImageCaption, ImageInline, ImageInsert, ImageInsertViaUrl,
  ImageResize, ImageStyle, ImageTextAlternative, ImageToolbar, ImageUpload, Indent,
  IndentBlock, Italic, Link, LinkImage, List, ListProperties, Markdown, MediaEmbed,
  Paragraph, PasteFromMarkdownExperimental, PasteFromOffice, RemoveFormat, SelectAll,
  ShowBlocks, SimpleUploadAdapter, SourceEditing, SpecialCharacters,
  SpecialCharactersArrows, SpecialCharactersCurrency, SpecialCharactersEssentials,
  SpecialCharactersLatin, SpecialCharactersMathematical, SpecialCharactersText,
  Strikethrough, Subscript, Superscript, Table, TableCaption, TableCellProperties,
  TableColumnResize, TableProperties, TableToolbar, TextTransformation, TodoList,
  Underline, Undo
} from 'ckeditor5';
import translations from 'ckeditor5/translations/zh-cn.js';

export const editorConfig: any = {
  toolbar: {
    items: [
      'undo', 'redo', '|', 'heading', '|',
      'fontSize', 'fontColor', 'fontBackgroundColor', '|',
      'bold', 'italic', 'underline', 'strikethrough', 'code', '|',
      'bulletedList', 'numberedList', 'todoList', '|',
      'link', 'insertImage', 'insertTable', 'blockQuote', 'codeBlock', '|',
      'alignment', 'outdent', 'indent', '|', 'sourceEditing', 'showBlocks'
    ],
    shouldNotGroupWhenFull: true
  },
  plugins: [
    AccessibilityHelp, Alignment, Autoformat, AutoImage, Autosave, BlockQuote, Bold,
    Code, CodeBlock, Essentials, FontBackgroundColor, FontColor, FontFamily, FontSize,
    FullPage, GeneralHtmlSupport, Heading, Highlight, HorizontalLine, HtmlComment,
    HtmlEmbed, ImageBlock, ImageCaption, ImageInline, ImageInsert, ImageInsertViaUrl,
    ImageResize, ImageStyle, ImageTextAlternative, ImageToolbar, ImageUpload, Indent,
    IndentBlock, Italic, Link, LinkImage, List, ListProperties, Markdown, MediaEmbed,
    Paragraph, PasteFromMarkdownExperimental, PasteFromOffice, RemoveFormat, SelectAll,
    ShowBlocks, SimpleUploadAdapter, SourceEditing, SpecialCharacters,
    SpecialCharactersArrows, SpecialCharactersCurrency, SpecialCharactersEssentials,
    SpecialCharactersLatin, SpecialCharactersMathematical, SpecialCharactersText,
    Strikethrough, Subscript, Superscript, Table, TableCaption, TableCellProperties,
    TableColumnResize, TableProperties, TableToolbar, TextTransformation, TodoList,
    Underline, Undo
  ],
  language: 'zh-cn',
  translations: [translations],
  image: {
    toolbar: [
      'toggleImageCaption', 'imageTextAlternative', '|',
      'imageStyle:inline', 'imageStyle:wrapText', 'imageStyle:breakText', '|',
      'resizeImage'
    ]
  },
  table: {
    contentToolbar: ['tableColumn', 'tableRow', 'mergeTableCells', 'tableProperties', 'tableCellProperties']
  },
  codeBlock: {
    languages: [
      { language: 'plaintext', label: 'Plain text' },
      { language: 'go', label: 'Golang' },
      { language: 'python', label: 'Python' },
      { language: 'javascript', label: 'JavaScript' },
      { language: 'typescript', label: 'TypeScript' },
      { language: 'java', label: 'Java' },
      { language: 'c', label: 'C' },
      { language: 'cpp', label: 'C++' },
      { language: 'sql', label: 'SQL' },
      { language: 'json', label: 'JSON' },
      { language: 'bash', label: 'Bash' },
      { language: 'html', label: 'HTML' },
      { language: 'css', label: 'CSS' }
    ]
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
  fontFamily: {
    supportAllValues: true
  },
  fontSize: {
    options: [10, 12, 14, 'default', 18, 20, 22],
    supportAllValues: true
  },
  placeholder: '这一刻的想法……',
};
