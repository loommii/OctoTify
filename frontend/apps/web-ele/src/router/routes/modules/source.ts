import type { RouteRecordRaw } from 'vue-router';
import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'mdi:source-branch',
      order: 10,
      title: $t('page.source.title'),
    },
    name: 'Source',
    path: '/source',
    redirect: '/source/list',
    children: [
      {
        name: 'SourceList',
        path: '/source/list',
        component: () => import('#/views/source/list.vue'),
        meta: {
          icon: 'mdi:list-box-outline',
          title: $t('page.source.list'),
        },
      },
      {
        name: 'SourceCreate',
        path: '/source/create',
        component: () => import('#/views/source/create.vue'),
        meta: {
          icon: 'mdi:plus',
          title: $t('page.source.create'),
        },
      },
      {
        name: 'SourceDetail',
        path: '/source/detail/:id',
        component: () => import('#/views/source/detail.vue'),
        meta: {
          hideInMenu: true,
          title: $t('page.source.detail'),
        },
      },
      {
        name: 'SourceEdit',
        path: '/source/edit/:id',
        component: () => import('#/views/source/edit.vue'),
        meta: {
          hideInMenu: true,
          title: $t('page.source.edit'),
        },
      },
    ],
  },
];

export default routes;
