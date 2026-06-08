import './assets/main.css'

import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import { createPinia } from 'pinia'

import VueAxios from 'vue-axios'
import axios from 'axios'

import { CkeditorPlugin } from '@ckeditor/ckeditor5-vue';

import dayjs from 'dayjs'
import localData from 'dayjs/plugin/localeData'
import { initTheme } from '@/composables/useTheme'

const app = createApp(App)

initTheme()

app.use(router)

const pinia = createPinia()
app.use(pinia)

app.use(VueAxios, axios as any)
axios.defaults.baseURL = import.meta.env.VITE_APP_BASE_URL

app.use(CkeditorPlugin)

dayjs.extend(localData)

app.mount('#app')
