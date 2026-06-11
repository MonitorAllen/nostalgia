<script setup lang="ts">
import { computed, inject, ref } from 'vue'
import { CornerDownRight, Trash2 } from '@lucide/vue'
import { useUserStore } from '@/store/module/user'
import date from '@/util/date'
import type { ArticleComments } from '@/types/article'
import AppBadge from '@/components/ui/AppBadge.vue'
import { sanitizeHtml } from '@/util/sanitizeHtml'

const props = defineProps<{
  comment: ArticleComments
  articleOwnerId: string
  replyCommentId: number | null
  isChild?: boolean
}>()

const emit = defineEmits<{
  reply: [id: number, toUserId: string, toUserName: string, parentId: number]
}>()

const userStore = useUserStore()
const showAllChildren = ref(false)

const displayedChildren = computed(() => {
  if (!props.comment.child) return []
  return showAllChildren.value ? props.comment.child : props.comment.child.slice(0, 2)
})
const sanitizedContent = computed(() => sanitizeHtml(props.comment.content || '', { profile: 'comment' }))

const deleteComment = inject<(id: number) => void>('deleteComment')

const handleDelete = () => {
  deleteComment?.(props.comment.id)
}

const handleReply = () => {
  emit(
    'reply',
    props.comment.id,
    props.comment.from_user_id,
    props.comment.from_user_name,
    props.comment.parent_id
  )
}
</script>

<template>
  <article
    class="comment-item rounded-archive px-3 py-3 transition hover:bg-muted/45"
    :class="isChild ? 'ml-4 border-l border-border pl-4' : 'border-b border-border/70'"
  >
    <div class="flex flex-wrap items-center gap-x-3 gap-y-2">
      <div class="flex flex-wrap items-center gap-2">
        <h3 class="m-0 text-sm font-black text-foreground">{{ comment.from_user_name }}</h3>
        <AppBadge v-if="comment.from_user_id === articleOwnerId" tone="accent">作者</AppBadge>
        <template
          v-if="isChild && comment.to_user_id && comment.from_user_id !== comment.to_user_id"
        >
          <CornerDownRight class="h-3.5 w-3.5 text-muted-foreground" />
          <h3 class="m-0 text-sm font-black text-foreground">{{ comment.to_user_name }}</h3>
          <AppBadge v-if="comment.to_user_id === articleOwnerId" tone="accent">作者</AppBadge>
        </template>
      </div>

      <div class="ml-auto flex items-center gap-3 text-xs font-semibold text-muted-foreground">
        <span>{{ date.format(comment.created_at, 'YYYY-MM-DD') }}</span>
        <button type="button" class="hover:text-accent" @click="handleReply">
          {{ replyCommentId === comment.id ? '取消回复' : '回复' }}
        </button>
        <button
          v-if="userStore.userInfo && comment.from_user_id === userStore.userInfo.id"
          type="button"
          class="text-muted-foreground hover:text-danger"
          aria-label="删除评论"
          @click="handleDelete"
        >
          <Trash2 class="h-3.5 w-3.5" />
        </button>
      </div>
    </div>

    <div
      class="reading-prose reading-prose--compact ck-content mt-2 text-sm"
      v-html="sanitizedContent"
    />

    <div v-if="comment.child?.length > 0" class="mt-3 space-y-2">
      <CommentItem
        v-for="child in displayedChildren"
        :key="child.id"
        :comment="child"
        :article-owner-id="articleOwnerId"
        :reply-comment-id="replyCommentId"
        is-child
        @reply="(...args) => emit('reply', ...args)"
      />

      <button
        v-if="comment.child.length > 2"
        type="button"
        class="ml-4 text-sm font-semibold text-muted-foreground hover:text-accent"
        @click="showAllChildren = !showAllChildren"
      >
        {{ showAllChildren ? '收起回复' : `展开全部 ${comment.child.length} 条回复` }}
      </button>
    </div>
  </article>
</template>
