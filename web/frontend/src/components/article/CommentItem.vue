<script setup lang="ts">
import { ref, computed } from "vue"
import {useUserStore} from "@/store/module/user";
import date  from "@/util/date"
import Badge from "primevue/badge"
import Divider from "primevue/divider";
import type {ArticleComments} from "@/types/article";

const props = defineProps<{
  comment: ArticleComments
  articleOwnerId: string
  replyCommentId: number | null
  isChild?: boolean
}>()

const emit = defineEmits(["delete", "reply"])

const userStore = useUserStore()
const hover = ref(false)

// 折叠控制
const showAllChildren = ref(false)

// 计算显示的子评论（默认 2 条，按时间升序）
const displayedChildren = computed(() => {
  if (!props.comment.child) return []
  return showAllChildren.value ? props.comment.child : props.comment.child.slice(0, 2)
})

const handleDelete = () => {
  emit("delete", props.comment.id)
}

const handleReply = () => {
  emit("reply", props.comment.id, props.comment.from_user_id, props.comment.from_user_name, props.comment.parent_id)
}
</script>

<template>
  <div
      class="comment-item"
      :class="{ child: isChild, 'bg-gray-50': hover }"
      @mouseenter="hover = true"
      @mouseleave="hover = false"
  >
    <!-- 头部 -->
    <div class="flex flex-row gap-2">
      <div class="flex align-items-center gap-2">
        <h6 class="m-0 text-base font-semibold">{{ comment.from_user_name }}</h6>
        <Badge
            v-if="comment.from_user_id === articleOwnerId"
            severity="success"
            size="small"
            value="作者"
        />
        <template v-if="isChild && comment.to_user_id && comment.from_user_id !== comment.to_user_id">
          <i class="pi pi-sort-up-fill rotate-90 text-xs"></i>
          <h6 class="m-0 text-base font-semibold">{{ comment.to_user_name }}</h6>
          <Badge
              v-if="comment.to_user_id === articleOwnerId"
              severity="success"
              size="small"
              value="作者"
          />
        </template>
      </div>

      <div class="flex flex-row align-items-center gap-2 text-sm text-gray">
        <small class="text-sm">{{ date.format(comment.created_at, "YYYY-MM-DD") }}</small>
        <span
            class="text-sm text-gray cursor-pointer select-none hover:text-blue-700 min-w-max"
            @click="handleReply"
        >
        {{ replyCommentId === comment.id ? "取消回复" : "回复" }}
      </span>
        <i
            v-if="hover && userStore.userInfo && comment.from_user_id === userStore.userInfo.id"
            class="pi pi-trash cursor-pointer hover:text-red-500"
            style="font-size: 0.75rem"
            @click="handleDelete"
        />
      </div>
    </div>

    <!-- 内容 -->
    <div class="flex flex-1 gap-2">
      <div class="text-base" v-html="comment.content"></div>
    </div>

    <!-- 子评论 -->
    <div v-if="comment.child?.length > 0" class="mt-2 space-y-2">
      <!-- 子评论列表（默认两条，点击展开全部） -->
      <CommentItem
          v-for="child in displayedChildren"
          :key="child.id"
          :comment="child"
          :article-owner-id="articleOwnerId"
          :reply-comment-id="replyCommentId"
          is-child
          @delete="emit('delete', $event)"
          @reply="(...args) => emit('reply', ...args)"
      />

      <!-- 展开/收起按钮 -->
      <div
          v-if="comment.child.length > 2"
          class="text-sm text-gray-500 cursor-pointer hover:text-blue-600 select-none ml-4"
          @click="showAllChildren = !showAllChildren"
      >
        {{ showAllChildren ? "收起回复" : `展开全部 ${comment.child.length} 条回复` }}
      </div>
    </div>

    <Divider v-if="!isChild" />
  </div>
</template>

<style scoped>
.comment-item {
  padding: 0.5rem 0;
  transition: background-color 0.2s;
}
.comment-item.child {
  margin-left: 1.5rem;
  padding-left: 1rem;
  border-left: 2px solid #e5e7eb;
}
</style>
