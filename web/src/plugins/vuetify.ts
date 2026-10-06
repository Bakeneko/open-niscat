import { createVuetify } from 'vuetify'
import { aliases, mdi } from 'vuetify/iconsets/mdi-svg'
import { createVueI18nAdapter } from 'vuetify/locale/adapters/vue-i18n'
import { useI18n } from 'vue-i18n'
import 'vuetify/styles'

// The adapter's own I18n type: typed message schemas narrow the locale union, which it does not accept.
type AdapterI18n = Parameters<typeof createVueI18nAdapter>[0]['i18n']

export function createAppVuetify(i18n: AdapterI18n) {
  return createVuetify({
    icons: { defaultSet: 'mdi', aliases, sets: { mdi } },
    locale: { adapter: createVueI18nAdapter({ i18n, useI18n }) },
    theme: { defaultTheme: 'system' },
  })
}
