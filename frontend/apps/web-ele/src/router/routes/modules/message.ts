import type { RouteRecordRaw } from 'vue-router';
import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'mdi:message-text',
      order: 30,
      title: $t('page.message.title'),
    },
    name: 'Message',
    path: '/message',
    redirect: '/message/list',
    children: [
      {
        name: 'MessageList',
        path: '/message/list',
        component: () => import('#/views/message/list.vue'),
        meta: {
          icon: 'mdi:list-box-outline',
          title: $t('page.message.list'),
        },
      },
      {
        name: 'MessageDetail',
        path: '/message/detail/:id',
        component: () => import('#/views/message/detail.vue'),
        meta: {
          hideInMenu: true,
          title: $t('page.message.detail'),
        },
      },
    ],
  },
];

export default routes;
