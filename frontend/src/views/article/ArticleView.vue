<template>
  <div class="flex flex-column justify-content-center">
    <div class="article-container flex flex-column m-auto p-2 mt-3 mb-3">
      <div class="flex flex-column align-items-center">
        <div>
          <h2 class="article-title">{{ title }}</h2>
        </div>
        <div class="flex flex-row gap-3">
          <div class="flex align-items-center">
            <i class="pi pi-calendar" style="font-size: .75rem"></i>
            <div class="font-medium text-xs ml-1">{{ date.format(created_at, 'YYYY-MM-DD HH:mm') }}</div>
          </div>
          <div class="flex align-items-center">
            <i class="pi pi-thumbs-up" style="font-size: .75rem"></i>
            <div class="font-medium text-xs ml-1">{{ likes }}</div>
          </div>
          <div class="flex align-items-center">
            <i class="pi pi-eye" style="font-size: .75rem; padding-top: 1px"></i>
            <div class="font-medium text-xs ml-1">{{ views }}</div>
          </div>
        </div>
      </div>
      <Divider/>
      <div class="article-summary">
        <h3>摘要</h3>
      </div>
      <div>
        {{ summary }}
      </div>
      <Divider/>
      <div class="flex flex-row">
        <div class="ck-content" v-html="content"></div>
      </div>
    </div>
    <div class="comment-box flex flex-column m-auto p-2 mb-3">
      <div class="flex flex-column" v-if="userStore.userInfo">
        <div class="editor-container editor-container_classic-editor">
          <div id="editor" class="editor-container__editor">
            <ckeditor
                v-if="isLayoutReady"
                v-model="editorData"
                :editor="ClassicEditor"
                :config="config"
                @ready="onEditorReady"
            />
          </div>
          <div class="flex flex-row align-items-center justify-content-between">
            <span class="text-lg text-color-secondary select-none">{{ replyUserName }}</span>
            <div class=" flex flex-row gap-2 mt-1">
              <Button label="取消回复" raised size="small" class="mt-1"
                      v-if="replyCommentId !== 0"
                      @click="replyComment(replyCommentId, '', '', 0)"/>
              <Button :label="replyCommentId === 0 ? '评论' : '回复'" raised size="small" class="mt-1"
                      @click="createComment(0, owner)"/>
            </div>
          </div>
        </div>
      </div>
      <div class="flex flex-column align-items-center" v-else>
        <div class="flex flex-column gap-2">
          <span>登录后才能使用评论功能</span>
          <div class="flex flex-row gap-2 justify-content-center">
            <Button label="登 录" size="small" raised @click="router.push('/login')"/>
            <Button label="注 册" size="small" raised @click="router.push('/register')"/>
          </div>
        </div>
      </div>
      <Divider/>
      <ConfirmDialog>
        <template #message="slotProps">
          <div class="flex flex-col items-center w-full gap-4 border-b">
            <p>{{ slotProps.message.message }}</p>
          </div>
        </template>
      </ConfirmDialog>
      <div class="flex flex-column gap-1"
           v-if="comments !== null && comments.length > 0">
        <div class="comment flex flex-column"
             v-for="(comment, index) in comments"
             :key="index">
          <div class="flex flex-row gap-2 text-color-secondary align-items-baseline select-none">
            <span class="text-lg">{{ comment.from_user_name }}</span>
            <div class="bg-green-100 border-round-sm px-1 text-sm" v-if="comment.from_user_id === owner">作者</div>
            <span class="text-sm">{{ date.format(comment.created_at, 'YYYY-MM-DD') }}</span>
            <i class="pi pi-trash cursor-pointer" style="font-size: 0.75rem"
               v-if="userStore.userInfo && comment.from_user_id === userStore.userInfo.id"
               @click="deleteComment(comment.id, index)"></i>
          </div>
          <div class="flex flex-row gap-2 align-items-end">
            <span class="text-sm" v-html="comment.content"></span>
            <span class="text-sm text-color-secondary cursor-pointer justify-content-between select-none"
                  @click="replyComment(comment.id, comment.from_user_id, comment.from_user_name, comment.parent_id)">
              {{ replyCommentId === comment.id ? "取消回复" : "回复" }}
            </span>
          </div>
          <div v-if="comment.child.length > 0">
            <div class="child-comment ml-2 align-items-baseline"
                 v-for="(childComment, childIndex) in comment.child"
                 :key="childIndex">
              <div class="flex flex-row gap-2 text-color-secondary align-items-baseline">
                <div class="flex flex-row align-items-baseline gap-1">
                  <span class="text-lg">{{ childComment.from_user_name }}</span>
                  <span class="bg-green-100 border-round-sm px-1 text-sm"
                        v-if="childComment.from_user_id === owner">作者</span>
                </div>
                <div class="flex flex-row align-items-baseline gap-1"
                     v-if="childComment.from_user_id !== childComment.to_user_id">
                  <i class="pi pi-angle-right"></i>
                  <span class="text-lg">{{ childComment.to_user_name }}</span>
                  <span class="bg-green-100 border-round-sm px-1 text-sm"
                        v-if="childComment.to_user_id === owner">作者</span>
                </div>
                <span class="text-sm">{{ date.format(childComment.created_at, 'YYYY-MM-DD') }}</span>
                <i class="pi pi-trash cursor-pointer" style="font-size: 0.75rem"
                   v-if="userStore.userInfo && childComment.from_user_id === userStore.userInfo.id"
                   @click="deleteComment(childComment.id, childIndex)"></i>
              </div>
              <div class="flex flex-row gap-2 align-items-end">
                <span class="text-sm" v-html="childComment.content"></span>
                <span class="text-sm text-color-secondary cursor-pointer justify-content-between select-none"
                      @click="replyComment(childComment.id, childComment.from_user_id, childComment.from_user_name, childComment.parent_id)">
                {{ replyCommentId === childComment.id ? "取消回复" : "回复" }}
              </span>
              </div>
            </div>
          </div>
          <Divider/>

        </div>
      </div>
      <div v-else>
        <span class="text-color-secondary">暂无评论</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {onMounted, onUpdated, type Ref, ref} from 'vue'
import router from "@/router";

import Divider from 'primevue/divider';
import Button from 'primevue/button';
import ConfirmDialog from 'primevue/confirmdialog';

import {useToast} from "primevue/usetoast";
import {useConfirm} from "primevue/useconfirm";

import Prism from 'prismjs';
import 'prismjs/components/prism-go.min.js';
import "prismjs/themes/prism-solarizedlight.css"

import date from '@/util/date'

import {ClassicEditor, Code, CodeBlock, type EditorConfig, Essentials, Paragraph} from 'ckeditor5'

import translations from 'ckeditor5/translations/zh-cn.js'

import {Ckeditor} from "@ckeditor/ckeditor5-vue";

import 'ckeditor5/ckeditor5.css'
import 'ckeditor5-premium-features/ckeditor5-premium-features.css'

import type {ArticleComments} from "@/types/article";
import {useArticleStore} from '@/store/module/article'
import {useUserStore} from "@/store/module/user";
import {useCommentStore} from "@/store/module/comment";

const userStore = useUserStore()
const articleStore = useArticleStore()
const commentStore = useCommentStore()

const toast = useToast()
const confirm = useConfirm()

const {id} = defineProps<{
  id: string
}>()

const editor = ref<ClassicEditor>()

const replyCommentId = ref(0)
const replyUserName = ref('')
const replyUserId = ref('')
const replyCommentParentId = ref(0)

const replyComment = (id: number, to_user_id: string, to_user_name: string, parent_id: number) => {
  //取消回复
  if (replyCommentId.value === id) {
    replyCommentId.value = 0
    replyUserName.value = ''
    replyUserId.value = ''
    replyCommentParentId.value = 0
  } else {
    const editor = document.getElementById('editor')
    editor?.scrollIntoView({behavior: 'smooth', block: 'center'})
    replyCommentId.value = id
    replyUserName.value = `@${to_user_name}`
    replyUserId.value = to_user_id
    // 如果当前评论没有父评论，则将 parent_id 设置为当前评论 id
    if (parent_id === 0) {
      replyCommentParentId.value = id
    } else {
      replyCommentParentId.value = parent_id
    }
  }

}

const deleteComment = (id: number, index: number) => {
  confirm.require({
    message: '确认删除这条评论吗？',
    rejectProps: {
      label: '取消',
      outlined: true,
      size: 'small'
    },
    acceptProps: {
      label: '确定',
      size: 'small'
    },
    accept: () => {
      commentStore.deleteComment(id)
          .then((res: any) => {
            comments.value?.splice(index, 1)
            toast.add({severity: 'success', summary: '成功', detail: '该评论已删除', life: 3000});
          })
    },
    reject: () => {
      return
    }
  });
}

const comments = ref<ArticleComments[]>([])

const createComment = (parent_id: number, to_user_id: string) => {
  if (!userStore.userInfo) {
    toast.add({severity: 'info', summary: 'Tips', detail: "请登录后使用评论功能", life: 2500})
    return
  }

  if (editorData.value.length === 0) {
    toast.add({severity: 'warn', summary: 'Warning', detail: "评论内容不能为空", life: 2500})
    return
  }

  // 默认为回复作者，如果用户指定了回复的用户，则将 to_user_id 设置为指定的用户
  if (replyUserId.value !== '') {
    to_user_id = replyUserId.value
  }

  // 如果存在指定的 parent_id
  if (replyCommentParentId.value !== 0) {
    parent_id = replyCommentParentId.value
  }

  commentStore.createComment(
      {
        article_id: id,
        content: editorData.value,
        parent_id: parent_id,
        from_user_id: userStore.userInfo.id,
        to_user_id: to_user_id
      }).then((res: any) => {
    // 如果 parent_id 为 0，则直接将评论添加到末尾
    if (parent_id === 0) {
      comments.value.push(res.data.comment)
    } else {
      comments.value.forEach((comment, index) => {
        if (comment.id === parent_id) {
          comments.value[index].child.push(res.data.comment)
        }
      })
    }
    // 回复成功后
    toast.add({severity: 'success', summary: 'Success', detail: "评论成功", life: 2500})

    // 重置编辑框内容
    editorData.value = ''

    // 重置回复相关数据
    replyCommentId.value = 0
    replyUserName.value = ''
    replyUserId.value = ''
    replyCommentParentId.value = 0

    // 确保回复内容中的代码高亮
    Prism.highlightAll();
  })

}

const config: Ref<EditorConfig> = ref({
  toolbar: {
    items: [
      'code', 'codeBlock'
    ],
    shouldNotGroupWhenFull: true
  },
  plugins: [
    Code,
    CodeBlock,
    Essentials,
    Paragraph,
  ],
  codeBlock: {
    languages: [
      {language: 'plaintext', label: 'Plain text'},
      {language: 'go', label: 'Golang'},
      {language: 'c', label: 'C'},
      {language: 'cs', label: 'C#'},
      {language: 'cpp', label: 'C++'},
      {language: 'css', label: 'CSS'},
      {language: 'diff', label: 'Diff'},
      {language: 'html', label: 'HTML'},
      {language: 'java', label: 'Java'},
      {language: 'javascript', label: 'JavaScript'},
      {language: 'php', label: 'PHP'},
      {language: 'python', label: 'Python'},
      {language: 'ruby', label: 'Ruby'},
      {language: 'typescript', label: 'TypeScript'},
      {language: 'xml', label: 'XML'}]
  },
  language: 'zh-cn',
  translations: [translations]
})
const editorData = ref('')
const isLayoutReady = ref(false)

const title = ref<string>('')
const summary = ref<string>('')
const content = ref<string>('')
const likes = ref<number>(0)
const views = ref<number>(0)
const owner = ref<string>('')
const created_at = ref<string>('')

onUpdated(() => {
  Prism.highlightAll();
})

onMounted(async () => {
  if (id != '') {
    await articleStore.getArticle(id).then((res: any) => {
      title.value = res.data.title
      summary.value = res.data.summary
      content.value = res.data.content
      likes.value = res.data.likes
      views.value = res.data.views
      owner.value = res.data.owner
      created_at.value = res.data.created_at
    }).catch(err => console.log(err))

    commentStore.listComments(id).then((res: any) => {
      comments.value = res.data === null ? [] : res.data
    })
  }

  isLayoutReady.value = true
})

// 当编辑器准备好时的回调
const onEditorReady = (editorInstance: ClassicEditor) => {
  editor.value = editorInstance
}

</script>

<style>
.article-container {
  width: 1000px;
  border: .1rem solid #ebebeb;
  background-color: rgb(239 239 239 / 0.3);

  .article-title {
    margin: 2px 0;
    height: fit-content;
    font-size: 26px;
  }

  .article-summary {
    border-left: 5px solid #20c997;
    background-color: #f7f7f7;
    font-size: 22px;
    margin-bottom: 5px;

    h3 {
      margin: 5px 10px;
    }
  }
}

.comment-box {
  width: 1000px;
  border: .1rem solid #ebebeb;
  background-color: rgb(239 239 239 / 0.3);
}

.editor-container_classic-editor .editor-container__editor .ck-editor__editable_inline {
  min-height: 100px !important;
}

pre {
  margin: 0 !important;
}
</style>