<script setup lang="ts">
import { computed, onMounted, onUnmounted, onUpdated, provide, ref, type Ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  AlertTriangle,
  Calendar,
  Clock,
  Eye,
  Heart,
  HeartOff,
  Link as LinkIcon,
  ShieldCheck
} from '@lucide/vue'
import { Ckeditor } from '@ckeditor/ckeditor5-vue'
import { ClassicEditor, Code, CodeBlock, type EditorConfig, Essentials, Paragraph } from 'ckeditor5'
import translations from 'ckeditor5/translations/zh-cn.js'
import 'ckeditor5/ckeditor5.css'
import 'ckeditor5/ckeditor5-content.css'

import Prism from 'prismjs'
import 'prismjs/components/prism-bash.min.js'
import 'prismjs/components/prism-c.min.js'
import 'prismjs/components/prism-cpp.min.js'
import 'prismjs/components/prism-css.min.js'
import 'prismjs/components/prism-go.min.js'
import 'prismjs/components/prism-java.min.js'
import 'prismjs/components/prism-javascript.min.js'
import 'prismjs/components/prism-json.min.js'
import 'prismjs/components/prism-python.min.js'
import 'prismjs/components/prism-sql.min.js'
import 'prismjs/components/prism-typescript.min.js'
import 'prismjs/themes/prism-solarizedlight.css'

import date from '@/util/date'
import type { Article, ArticleComments } from '@/types/article'
import { useUserStore } from '@/store/module/user'
import { useCommentStore } from '@/store/module/comment'
import {
  getArticle,
  getArticleBySlug,
  incrementArticleLikes,
  incrementArticleViews
} from '@/api/article'
import { listComments } from '@/api/comment'
import CommentItem from '@/components/article/CommentItem.vue'
import { isUUID } from '@/util/validate'
import { useToast } from '@/composables/useToast'
import AppButton from '@/components/ui/AppButton.vue'
import AppBadge from '@/components/ui/AppBadge.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import { sanitizeHtml } from '@/util/sanitizeHtml'

const router = useRouter()
const userStore = useUserStore()
const commentStore = useCommentStore()
const toast = useToast()

const props = defineProps<{
  id: string
}>()

const articlePath = ref(window.location.href)
const articleId = ref(props.id)
const editor = ref<ClassicEditor>()
const editorData = ref('')
const isSubmittingComment = ref(false)
const isLayoutReady = ref(false)
const article = ref<Article | null>(null)
const comments = ref<ArticleComments[]>([])

const replyCommentId = ref(0)
const replyUserName = ref('')
const replyUserId = ref('')
const replyCommentParentId = ref(0)

const deleteDialogOpen = ref(false)
const pendingDeleteId = ref<number | null>(null)
const sanitizedArticleContent = computed(() => sanitizeHtml(article.value?.content || ''))

const scrollProgress = ref(0)
const viewed = ref(false)
const liked = ref(false)
const isOutdated = ref(false)
let timer: ReturnType<typeof setTimeout>

const config: Ref<EditorConfig> = ref({
  toolbar: {
    items: ['undo', 'redo', '|', 'code', 'codeBlock'],
    shouldNotGroupWhenFull: true
  },
  plugins: [Code, CodeBlock, Essentials, Paragraph],
  placeholder: '写下评论，Ctrl/⌘ + Enter 提交',
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
      { language: 'xml', label: 'XML' }
    ]
  },
  language: 'zh-cn',
  translations: [translations]
})

const replyComment = (
  commentId: number,
  toUserId: string,
  toUserName: string,
  parentId: number
) => {
  if (replyCommentId.value === commentId) {
    replyCommentId.value = 0
    replyUserName.value = ''
    replyUserId.value = ''
    replyCommentParentId.value = 0
    return
  }

  document.getElementById('editor')?.scrollIntoView({ behavior: 'smooth', block: 'center' })
  replyCommentId.value = commentId
  replyUserName.value = `@${toUserName}`
  replyUserId.value = toUserId
  replyCommentParentId.value = parentId === 0 ? commentId : parentId
}

const removeCommentFromTree = (list: ArticleComments[], targetId: number): boolean => {
  const index = list.findIndex((comment) => comment.id === targetId)
  if (index > -1) {
    list.splice(index, 1)
    return true
  }

  for (const comment of list) {
    if (comment.child?.length && removeCommentFromTree(comment.child, targetId)) return true
  }
  return false
}

const requestDeleteComment = (commentId: number) => {
  pendingDeleteId.value = commentId
  deleteDialogOpen.value = true
}

const confirmDeleteComment = () => {
  if (!pendingDeleteId.value) return
  const targetId = pendingDeleteId.value
  commentStore.deleteComment(targetId).then(() => {
    if (removeCommentFromTree(comments.value, targetId)) {
      toast.add({
        severity: 'success',
        summary: '评论已删除',
        detail: '评论列表已更新',
        life: 2500
      })
    }
    pendingDeleteId.value = null
    deleteDialogOpen.value = false
  })
}

provide('deleteComment', requestDeleteComment)

const getCommentText = (html: string) => {
  const container = document.createElement('div')
  container.innerHTML = html
  return (container.textContent || '').replace(/\u00a0/g, ' ').trim()
}

const hasMeaningfulComment = (html: string) => getCommentText(html).length > 0

const createComment = async (parentId: number, toUserId: string) => {
  if (!article.value) return
  if (!userStore.userInfo) {
    toast.add({
      severity: 'info',
      summary: '需要登录',
      detail: '登录后才能使用评论功能',
      life: 2500
    })
    return
  }

  if (!hasMeaningfulComment(editorData.value)) {
    toast.add({
      severity: 'warning',
      summary: '评论为空',
      detail: '请写下一点内容再提交',
      life: 2500
    })
    return
  }
  if (isSubmittingComment.value) return

  const resolvedToUserId = replyUserId.value || toUserId
  const resolvedParentId = replyCommentParentId.value || parentId

  isSubmittingComment.value = true
  try {
    const res: any = await commentStore.createComment({
      article_id: articleId.value,
      content: editorData.value,
      parent_id: resolvedParentId,
      from_user_id: userStore.userInfo.id,
      to_user_id: resolvedToUserId
    })

    if (resolvedParentId === 0) {
      comments.value.push(res.data.comment)
    } else {
      comments.value.forEach((comment, index) => {
        if (comment.id === resolvedParentId) comments.value[index].child.push(res.data.comment)
      })
    }

    toast.add({ severity: 'success', summary: '评论成功', detail: '新的评论已发布', life: 2500 })
    editorData.value = ''
    replyCommentId.value = 0
    replyUserName.value = ''
    replyUserId.value = ''
    replyCommentParentId.value = 0
    Prism.highlightAll()
  } catch (error: any) {
    toast.add({
      severity: 'error',
      summary: '评论失败',
      detail: error.response?.data?.error || '请稍后再试',
      life: 3000
    })
  } finally {
    isSubmittingComment.value = false
  }
}

const checkValidView = async () => {
  if (viewed.value || !article.value) return
  try {
    await incrementArticleViews({ id: articleId.value })
    article.value.views++
    viewed.value = true
  } catch (error: any) {
    if (error.response?.status === 409) viewed.value = true
  }
}

const checkValidLike = async () => {
  if (!article.value) return
  try {
    await incrementArticleLikes({ id: articleId.value })
    article.value.likes++
    liked.value = true
    toast.add({
      severity: 'success',
      summary: '感谢点赞',
      detail: '这篇文章对你有帮助就太好了',
      life: 3000
    })
  } catch (error: any) {
    if (error.response?.status === 409) {
      liked.value = true
      toast.add({
        severity: 'info',
        summary: '已经点过赞',
        detail: '你最近已经支持过这篇文章了',
        life: 3000
      })
    } else {
      toast.add({
        severity: 'error',
        summary: '点赞失败',
        detail: error.response?.data?.error || '请稍后再试',
        life: 3000
      })
    }
  }
}

const daysDiff = (lastUpdated: string) => {
  const now = new Date()
  const last = new Date(lastUpdated)
  const diffTime = Math.abs(now.getTime() - last.getTime())
  return Math.ceil(diffTime / (1000 * 60 * 60 * 24))
}

const checkOutdated = (shouldCheck: boolean, lastUpdated: string) => {
  if (shouldCheck && daysDiff(lastUpdated) > 60) {
    isOutdated.value = true
  }
}

const updateScrollProgress = () => {
  const winScroll = document.documentElement.scrollTop
  const height = document.documentElement.scrollHeight - document.documentElement.clientHeight
  scrollProgress.value = height > 0 ? (winScroll / height) * 100 : 0
}

const onEditorReady = (editorInstance: ClassicEditor) => {
  editor.value = editorInstance
  editorInstance.editing.view.document.on('keydown', (event: any, data: any) => {
    const domEvent = data.domEvent as KeyboardEvent
    if ((domEvent.ctrlKey || domEvent.metaKey) && domEvent.key === 'Enter') {
      data.preventDefault()
      event.stop()
      createComment(0, article.value?.owner || '')
    }
  })
}

onUpdated(() => {
  Prism.highlightAll()
})

onMounted(async () => {
  if (articleId.value) {
    try {
      let articleRes
      if (isUUID(articleId.value)) {
        articleRes = await getArticle({ id: articleId.value })
      } else {
        articleRes = await getArticleBySlug({ slug: articleId.value })
        articleId.value = articleRes.data.article.id
      }

      article.value = articleRes.data.article
      checkOutdated(article.value.check_outdated, article.value.last_updated)

      const commentRes = await listComments({ articleId: articleId.value })
      comments.value = commentRes.data.comments === null ? [] : commentRes.data.comments
    } catch (error: any) {
      toast.add({
        severity: 'error',
        summary: '获取文章失败',
        detail: error.response?.data?.error || '获取文章信息失败',
        life: 3000
      })
    }
  }

  isLayoutReady.value = true
  window.addEventListener('scroll', updateScrollProgress)
  timer = setTimeout(() => {
    checkValidView()
  }, 1000 * 10)
})

onUnmounted(() => {
  clearTimeout(timer)
  window.removeEventListener('scroll', updateScrollProgress)
  if (editor.value) editor.value.destroy()
})
</script>

<template>
  <div
    class="fixed left-0 top-0 z-50 h-1 bg-accent transition-[width]"
    :style="{ width: scrollProgress + '%' }"
  />

  <main class="mx-auto flex w-full max-w-[820px] flex-col gap-5 px-4 py-6 md:py-10">
    <div
      v-if="isOutdated"
      class="archive-glass flex items-start gap-3 rounded-archive p-4 text-warning"
    >
      <AlertTriangle class="mt-0.5 h-5 w-5 shrink-0" />
      <div>
        <p class="m-0 text-sm font-bold">这篇文章可能已经过时</p>
        <p class="m-0 mt-1 text-sm text-muted-foreground">
          请结合最新版本、文档或实际环境复核相关内容。
        </p>
      </div>
    </div>

    <article v-if="article" class="archive-surface rounded-[1.1rem] p-5 md:p-8">
      <header class="space-y-5 border-b border-border pb-6">
        <div class="flex flex-wrap gap-2">
          <AppBadge tone="accent">{{ article.category_name }}</AppBadge>
          <AppBadge v-if="article.read_time">{{ article.read_time }}</AppBadge>
        </div>
        <h1 class="m-0 text-3xl font-black leading-tight text-foreground md:text-5xl">
          {{ article.title }}
        </h1>
        <div
          class="flex flex-wrap items-center gap-x-4 gap-y-2 text-sm font-semibold text-muted-foreground"
        >
          <span class="inline-flex items-center gap-1"
            ><Calendar class="h-4 w-4" />{{ date.format(article.created_at, 'YYYY-MM-DD') }}</span
          >
          <span class="inline-flex items-center gap-1"
            ><Heart class="h-4 w-4" />{{ article.likes }}</span
          >
          <span class="inline-flex items-center gap-1"
            ><Eye class="h-4 w-4" />{{ article.views }}</span
          >
          <span v-if="article.read_time" class="inline-flex items-center gap-1"
            ><Clock class="h-4 w-4" />{{ article.read_time }}</span
          >
        </div>
      </header>

      <section class="my-6 rounded-archive border border-border bg-surface-raised p-4">
        <p class="m-0 text-xs font-black uppercase text-muted-foreground">摘要</p>
        <p class="m-0 mt-2 text-base leading-8 text-foreground/85">{{ article.summary }}</p>
      </section>

      <div class="reading-prose ck-content" v-html="sanitizedArticleContent" />
    </article>

    <div v-else class="archive-surface rounded-archive p-8 text-center">
      <p class="m-0 text-lg font-bold">正在加载这篇文章</p>
      <p class="m-0 mt-2 text-sm text-muted-foreground">如果长时间没有显示，请稍后重试。</p>
    </div>

    <section v-if="article" class="archive-glass rounded-archive p-4">
      <div class="flex items-center gap-2 text-sm font-bold text-foreground">
        <ShieldCheck class="h-4 w-4 text-accent" />
        版权声明
      </div>
      <div class="mt-3 space-y-2 text-sm leading-7 text-muted-foreground">
        <p class="m-0 inline-flex max-w-full items-center gap-2 break-all">
          <LinkIcon class="h-4 w-4 shrink-0" />
          {{ articlePath }}
        </p>
        <p class="m-0">
          本文采用
          <a
            class="font-semibold text-accent"
            href="https://creativecommons.org/licenses/by-nc-sa/4.0/"
            target="_blank"
          >
            CC BY-NC-SA 4.0
          </a>
          许可协议，转载请注明出处。
        </p>
      </div>
    </section>

    <section v-if="article" class="archive-glass flex justify-center rounded-archive p-4">
      <AppButton :variant="liked ? 'secondary' : 'primary'" @click="checkValidLike">
        <Heart v-if="!liked" class="h-4 w-4" />
        <HeartOff v-else class="h-4 w-4" />
        {{ liked ? '已记录点赞' : '这篇有帮助' }}
      </AppButton>
    </section>

    <section v-if="article" class="archive-surface rounded-archive p-4 md:p-5">
      <h2 class="m-0 text-xl font-black">评论</h2>

      <div v-if="userStore.userInfo" class="mt-4">
        <div id="editor" class="overflow-hidden rounded-archive border border-border">
          <ckeditor
            v-if="isLayoutReady"
            v-model="editorData"
            :editor="ClassicEditor"
            :config="config"
            @ready="onEditorReady"
          />
        </div>
        <div class="mt-3 flex items-center justify-between gap-3">
          <span class="text-sm font-semibold text-muted-foreground">{{ replyUserName }}</span>
          <div class="flex gap-2">
            <AppButton
              v-if="replyCommentId !== 0"
              variant="ghost"
              size="sm"
              @click="replyComment(replyCommentId, '', '', 0)"
            >
              取消回复
            </AppButton>
            <AppButton
              size="sm"
              :disabled="isSubmittingComment"
              @click="createComment(0, article.owner)"
            >
              {{ isSubmittingComment ? '提交中' : replyCommentId === 0 ? '发表评论' : '回复评论' }}
            </AppButton>
          </div>
        </div>
      </div>

      <div
        v-else
        class="mt-4 rounded-archive border border-border bg-surface-raised p-5 text-center"
      >
        <p class="m-0 text-sm font-semibold text-muted-foreground">登录后才能使用评论功能</p>
        <div class="mt-4 flex justify-center gap-2">
          <AppButton size="sm" @click="router.push('/login')">登录</AppButton>
          <AppButton size="sm" variant="secondary" @click="router.push('/register')"
            >注册</AppButton
          >
        </div>
      </div>

      <div class="mt-6 border-t border-border pt-4">
        <div v-if="comments.length > 0" id="comment-list" class="space-y-1">
          <CommentItem
            v-for="comment in comments"
            :key="comment.id"
            :comment="comment"
            :article-owner-id="article.owner"
            :reply-comment-id="replyCommentId"
            @reply="replyComment"
          />
        </div>
        <p v-else class="m-0 text-sm text-muted-foreground">暂无评论，第一条评论可以由你写下。</p>
      </div>
    </section>
  </main>

  <ConfirmDialog
    :open="deleteDialogOpen"
    title="删除评论"
    description="确认删除这条评论吗？这会从当前评论列表中移除。"
    confirm-label="删除评论"
    danger
    @cancel="deleteDialogOpen = false"
    @confirm="confirmDeleteComment"
  />
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
