import articleService from '@/service/articleService'
import { defineStore } from 'pinia'
import type { CreateArticleRequest } from '@/types/request/article'

interface UpdateArticleRequest {
  id: string
  title: string
  summary: string
  content: string
  is_publish: boolean
  owner: string
}

export const useArticleStore = defineStore('article', {
  actions: {
    getArticle(id: string) {
      return new Promise((resolve, reject) => {
        articleService
          .getArticle(id)
          .then((res) => {
            resolve(res)
          })
          .catch((err) => {
            reject(err)
          })
      })
    },
    listArticle(category_id: number, page: number, limit: number) {
      return new Promise((resolve, reject) => {
        articleService
          .listArticle(category_id, page, limit)
          .then((res) => {
            resolve(res)
          })
          .catch((err) => {
            reject(err)
          })
      })
    },
    createArticle({ title, summary, content, is_publish }: CreateArticleRequest) {
      return new Promise((resolve, reject) => {
        articleService
          .createArticle({ title, summary, content, is_publish })
          .then((res) => {
            resolve(res)
          })
          .catch((err) => {
            reject(err)
          })
      })
    },
    updateArticle({ id, title, summary, content, is_publish, owner }: UpdateArticleRequest) {
      return new Promise((resolve, reject) => {
        articleService
          .updateArticle({ id, title, summary, content, is_publish, owner })
          .then((res) => {
            resolve(res)
          })
          .catch((err) => {
            reject(err)
          })
      })
    },
  },
})
