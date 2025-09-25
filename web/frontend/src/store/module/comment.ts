import {defineStore} from "pinia";
import commentService from "@/service/commentService";
import type {CreateCommentRequest} from "@/types/request/comment";

export const useCommentStore = defineStore('comment', {
    actions: {
        createComment(req: CreateCommentRequest) {
            return new Promise((resolve, reject) => {
                commentService
                    .createComment(req)
                    .then((res) => {
                        resolve(res)
                    })
                    .catch((err) => {
                        reject(err)
                    })
            })
        },
        listComments(articleId: string) {
            return new Promise((resolve, reject) => {
                commentService
                    .listComments(articleId)
                    .then((res) => {
                        resolve(res)
                    })
                    .catch((err) => {
                        reject(err)
                    })
            })
        },
        deleteComment(id: number) {
            return new Promise((resolve, reject) => {
                commentService.deleteComment(id)
                    .then((res) => {
                        resolve(res)
                    })
                    .catch((err) => {
                        reject(err)
                    })
            })
        }
    },
})