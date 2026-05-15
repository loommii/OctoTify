import type { RouteRecordRaw } from 'vue-router';

import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'mdi:cog',
      order: 40,
      title: $t('page.settings.title'),
    },
    name: 'Settings',
    path: '/settings',
    redirect: '/settings/profile',
    children: [
      {
        name: 'SettingsProfile',
        path: '/settings/profile',
        component: () => import('#/views/settings/profile.vue'),
        meta: {
          icon: 'mdi:account',
          title: $t('page.settings.profile'),
        },
      },
      {
        name: 'SettingsPassword',
        path: '/settings/password',
        component: () => import('#/views/settings/password.vue'),
        meta: {
          icon: 'mdi:lock',
          title: $t('page.settings.password'),
        },
      },
    ],
  },
];

export default routes;
