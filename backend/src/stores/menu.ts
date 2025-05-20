import axiosInstance from '@/config/axios'
import { defineStore } from 'pinia'
import { ref } from 'vue'

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
        const res = await axiosInstance.get('/menu/init')
        sysMenu.value = res.data.init_sys_menu
    }

    return {
        sysMenu,
        initMenu
    }
})
