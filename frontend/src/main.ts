import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import './ui/tokens.css'
import './ui/primitives.css'
import './styles.css'
import './ui/service-pages.css'
import '@wavefnd/editor/source/style.css'

createApp(App).use(createPinia()).use(router).mount('#app')
