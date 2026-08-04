import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import 'github-markdown-css/github-markdown-light.css'
import './ui/tokens.css'
import './ui/primitives.css'
import './styles.css'
import './ui/service-pages.css'

createApp(App).use(createPinia()).use(router).mount('#app')
