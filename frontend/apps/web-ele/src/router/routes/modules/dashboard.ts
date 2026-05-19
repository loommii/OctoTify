import type { RouteRecordRaw } from 'vue-router';

import { $t } from '#/locales';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:layout-dashboard',
      order: 0,
      title: $t('page.dashboard.title'),
    },
    name: 'Dashboard',
    path: '/dashboard',
    redirect: '/dashboard/index',
    children: [
      {
        name: 'DashboardIndex',
        path: '/dashboard/index',
        component: () => import('#/views/dashboard/index.vue'),
        meta: {
          icon: 'lucide:layout-dashboard',
          title: $t('page.dashboard.index'),
        },
      },
    ],
  },
];

export default routes;
