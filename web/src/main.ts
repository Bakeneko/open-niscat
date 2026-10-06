import { createApp } from 'vue'
import { createI18n } from 'vue-i18n'

import App from './App.vue'
import router from './router'
import { createAppVuetify } from './plugins/vuetify'

const i18n = createI18n<false>({ legacy: false, locale: 'en', messages: { en: {}, fr: {} } })
createApp(App).use(router).use(i18n).use(createAppVuetify(i18n)).mount('#app')
