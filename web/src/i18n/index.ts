import { createI18n } from 'vue-i18n'
import { en as vuetifyEn, fr as vuetifyFr } from 'vuetify/locale'
import { en } from './en'
import { fr } from './fr'

export const i18n = createI18n({
  legacy: false,
  locale: 'en',
  fallbackLocale: 'en',
  messages: { en: { ...en, $vuetify: vuetifyEn }, fr: { ...fr, $vuetify: vuetifyFr } },
})
