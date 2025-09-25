import './assets/main.css'

import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import { createPinia } from 'pinia'

import VueAxios from 'vue-axios'
import axios from 'axios'

import PrimeVue from 'primevue/config'
import Aura from '@primevue/themes/aura'

import ToastService from 'primevue/toastservice'
import ConfirmationService from 'primevue/confirmationservice';
import Tooltip from 'primevue/tooltip';

import { CkeditorPlugin } from '@ckeditor/ckeditor5-vue';

const app = createApp(App)

app.use(router)

const pinia = createPinia()
app.use(pinia)

app.use(VueAxios, axios as any)
axios.defaults.baseURL = import.meta.env.VITE_APP_BASE_URL

app.use(ToastService)
app.use(ConfirmationService)
app.directive('tooltip', Tooltip);
app.use(PrimeVue, {
  theme: {
    preset: Aura
  }
})

app.use(CkeditorPlugin)

app.mount('#app')
