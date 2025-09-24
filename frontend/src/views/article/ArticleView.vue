<script setup lang="ts">
import {onMounted, onUnmounted, onUpdated, type Ref, ref} from 'vue'
import router from "@/router";

import Divider from 'primevue/divider';
import Button from 'primevue/button';
import ConfirmDialog from 'primevue/confirmdialog';

import {useToast} from "primevue/usetoast";
import {useConfirm} from "primevue/useconfirm";

import 'ckeditor5/ckeditor5.css'
import 'ckeditor5/ckeditor5-content.css'
import {Ckeditor} from "@ckeditor/ckeditor5-vue";
import {ClassicEditor, Code, CodeBlock, type EditorConfig, Essentials, Paragraph} from 'ckeditor5'
import translations from 'ckeditor5/translations/zh-cn.js'

import Prism from 'prismjs';
import 'prismjs/components/prism-go.min.js';
import "prismjs/themes/prism-solarizedlight.css"

import date from '@/util/date'
import type {Article, ArticleComments} from "@/types/article";
import {useUserStore} from "@/store/module/user";
import {useCommentStore} from "@/store/module/comment";
import {getArticle, incrementArticleLikes, incrementArticleViews} from "@/api/article";
import {listComments} from "@/api/comment";
import CommentItem from "@/components/article/CommentItem.vue";

const userStore = useUserStore()
const commentStore = useCommentStore()

const toast = useToast()
const confirm = useConfirm()

const {id} = defineProps<{
  id: string
}>()

const articlePath = ref(window.location.href)

const editor = ref<ClassicEditor>()

const replyCommentId = ref(0)
const replyUserName = ref('')
const replyUserId = ref('')
const replyCommentParentId = ref(0)

const replyComment = (id: number, to_user_id: string, to_user_name: string, parent_id: number) => {
  //取消回复
  console.log('id: ' + id, 'to user id: ' + to_user_id, 'to user name: ' + to_user_name, 'parent id: ' + parent_id)
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

const article = ref<Article | null>(null)

let timer: NodeJS.Timeout
const viewed = ref(false)

const checkValidView = async () => {
  if (viewed.value) return
  try {
    await incrementArticleViews({id});
    article.value!.views++
    viewed.value = true
  } catch (error: any) {
    if (error.response.status == 409) {
      viewed.value = true
    }
  }
}

const liked  = ref(false)

const checkValidLike = async () => {
  // if (liked.value) return
  try {
    await incrementArticleLikes({id});
    article.value!.likes++
    liked.value = true
    toast.add({
      severity: 'success',
      summary: '成功',
      detail: '感谢您的点赞！',
      life: 3000
    })
  } catch (error: any) {
    if (error.response.status == 409) {
      liked.value = true
      toast.add({
        severity: 'info',
        summary: '提示',
        detail: '您最近已经点过赞了',
        // life: 3000
      })
    } else {
      toast.add({
        severity: 'error',
        summary: '失败',
        detail: error.response.data.error || '点赞失败',
        life: 3000
      })
    }
  }
}

onUpdated(() => {
  Prism.highlightAll();
})

onMounted(async () => {
  if (id != '') {
    try {
      const [articleRes, commentRes] = await Promise.all([getArticle({id}), listComments({articleId: id})]);
      article.value = articleRes.data.article
      comments.value = commentRes.data.comments
    } catch (error: any) {
      toast.add({
        severity: 'error',
        summary: '错误',
        detail: error.response?.data?.error || '获取文章信息失败',
        life: 3000
      })
    }
  }
  isLayoutReady.value = true

  timer = setTimeout(() => {
    checkValidView()
  }, 1000 * 10)
})

onUnmounted(() => {
  clearTimeout(timer)
  if (editor.value) {
    editor.value.destroy()
  }
})

// 当编辑器准备好时的回调
const onEditorReady = (editorInstance: ClassicEditor) => {
  editor.value = editorInstance
}

</script>

<template>
  <div class="flex flex-column row-gap-3 justify-content-center w-11 md:w-10 lg:w-8 mx-auto">
    <div class="article-container surface-0 w-full flex flex-column m-auto p-2 mt-3">
      <div class="flex flex-column align-items-center">
        <div>
          <h2 class="article-title text-green-600">{{ article?.title }}</h2>
        </div>
        <div class="flex flex-row gap-3">
          <div class="flex align-items-center">
            <i class="pi pi-calendar" style="font-size: .8rem"></i>
            <div class="font-medium text-sm ml-1">{{ date.format(article?.created_at as string, 'YYYY-MM-DD HH:mm') }}</div>
          </div>
          <div class="flex align-items-center">
            <i class="pi pi-heart" style="font-size: .8rem"></i>
            <div class="font-medium text-sm ml-1">{{ article?.likes }}</div>
          </div>
          <div class="flex align-items-center">
            <i class="pi pi-eye" style="font-size: .8rem;"></i>
            <div class="font-medium text-sm ml-1">{{ article?.views }}</div>
          </div>
        </div>
      </div>
      <Divider/>
      <div class="flex flex-row h-3rem align-items-center article-summary">
        <div class="pl-2">摘要</div>
      </div>
      <div>
        {{ article?.summary }}
      </div>
      <Divider/>
      <div class="w-full ck-content" style="overflow-wrap: break-word; word-break: break-word;" v-html="article?.content"></div>
    </div>
    <div class="flex flex-column w-full bg-gray-50 border-1 border-gray-200 m-auto">
      <div class="flex flex-row column-gap-1 h-2rem bg-green-200 p-1 align-items-center">
        <i class="pi pi-exclamation-triangle"></i>
        <div>小提示</div>
      </div>
      <div class="flex flex-column p-1">
        <div>
          版权声明：原创不易，转载请注明出处。
        </div>
        <div>
          原文链接：{{articlePath}}
        </div>
      </div>
    </div>
    <div class="flex flex-row w-full p-3 bg-gray-50 border-1 border-gray-200 m-auto justify-content-center">
      <div class="flex" v-tooltip.left="{
        value: '对你有帮助？点个赞吧！',
        pt: {
          text: {
            style: {
              fontSize: '12px',
            }
          }
        }
      }" @click="checkValidLike">
        <i class="pi cursor-pointer"
           :class="liked ? 'pi-heart-fill text-red-500' : 'pi-heart'" style="font-size: 1.25rem"></i>
      </div>
    </div>
    <div class="comment-box w-full flex flex-column m-auto p-2 mb-3">
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
                      @click="createComment(0, article!.owner)"/>
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
      <div id="comment-list" class="" v-if="comments !== null && comments.length > 0">
        <CommentItem
            v-for="comment in comments"
            :key="comment.id"
            :comment="comment"
            :article-owner-id="article!.owner"
            :reply-comment-id="replyCommentId"
            @delete="deleteComment"
            @reply="replyComment"
        />
      </div>
      <div v-else>
        <span class="text-color-secondary">暂无评论</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.article-container {
  border: .1rem solid #ebebeb;

  .article-title {
    margin: 2px 0;
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
  border: .1rem solid #ebebeb;
  background-color: rgb(239 239 239 / 0.3);
}

:deep(.comment-item) code {
  background-color: #fdf6e3;
}

:deep(.ck-content) code {
  background-color: #fdf6e3;
}

:deep(.ck-content) blockquote {
  border-left: solid 5px #20c997;
  background-color: #f7f7f7;
}

.editor-container_classic-editor .editor-container__editor .ck-editor__editable_inline {
  min-height: 100px !important;
}


</style>