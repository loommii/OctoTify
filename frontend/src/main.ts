import { createApp } from 'vue'
import { createPinia } from 'pinia'
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'
import { createHead } from '@unhead/vue/client'

import App from '@/App.vue'
import router from '@/router/index'
import { useAuth } from '@/composables/useAuth'

const app = createApp(App)

const pinia = createPinia()
pinia.use(piniaPluginPersistedstate)
app.use(pinia)

app.use(router)

const head = createHead()
app.use(head)

const { initAuth } = useAuth()
initAuth()

app.mount('#app')
