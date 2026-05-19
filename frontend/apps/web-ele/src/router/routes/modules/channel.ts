import type { RouteRecordRaw } from 'vue-router';
import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    meta: { icon: 'mdi:pipe', order: 20, title: $t('page.channel.title') },
    name: 'Channel',
    path: '/channel',
    redirect: '/channel/list',
    children: [
      {
        name: 'ChannelList',
        path: '/channel/list',
        component: () => import('#/views/channel/list.vue'),
        meta: { icon: 'mdi:list-box-outline', title: $t('page.channel.list') },
      },
      {
        name: 'ChannelCreate',
        path: '/channel/create',
        component: () => import('#/views/channel/create.vue'),
        meta: { icon: 'mdi:plus', title: $t('page.channel.create') },
      },
      {
        name: 'ChannelDetail',
        path: '/channel/detail/:id',
        component: () => import('#/views/channel/detail.vue'),
        meta: { hideInMenu: true, title: $t('page.channel.detail') },
      },
      {
        name: 'ChannelEdit',
        path: '/channel/edit/:id',
        component: () => import('#/views/channel/edit.vue'),
        meta: { hideInMenu: true, title: $t('page.channel.edit') },
      },
    ],
  },
];

export default routes;
