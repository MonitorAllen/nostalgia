import { defineStore } from 'pinia'
import { ref } from 'vue'
import http from "@/util/http.ts";

interface SysMenu {
    id: string
    name: string
    icon: string
    path?: string
    parentId?: string
    children?: SysMenu[]
}

export const useMenuStore = defineStore('menu', () => {
    const sysMenu = ref<SysMenu[]>([])

    const initMenu = async () => {
        const res = await http.get('/menu/init')
        sysMenu.value = res.data.init_sys_menu
    }

    return {
        sysMenu,
        initMenu
    }
})
