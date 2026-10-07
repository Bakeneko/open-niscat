import { createI18n } from 'vue-i18n'
import { de as vuetifyDe, en as vuetifyEn, es as vuetifyEs, fr as vuetifyFr } from 'vuetify/locale'
import { de } from './de'
import { en } from './en'
import { es } from './es'
import { fr } from './fr'

export const i18n = createI18n({
  legacy: false,
  locale: 'en',
  fallbackLocale: 'en',
  messages: {
    en: { ...en, $vuetify: vuetifyEn },
    fr: { ...fr, $vuetify: vuetifyFr },
    es: { ...es, $vuetify: vuetifyEs },
    de: { ...de, $vuetify: vuetifyDe },
  },
})
