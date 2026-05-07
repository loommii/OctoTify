import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/auth/LoginView.vue'),
    meta: { requiresGuest: true },
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/auth/RegisterView.vue'),
    meta: { requiresGuest: true },
  },
  {
    path: '/',
    component: () => import('@/components/Layout/AppLayout.vue'),
    meta: { requiresAuth: true },
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/DashboardView.vue'),
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/views/auth/ProfileView.vue'),
      },
      {
        path: 'change-password',
        name: 'ChangePassword',
        component: () => import('@/views/auth/ChangePasswordView.vue'),
      },
      {
        path: 'channels',
        name: 'ChannelList',
        component: () => import('@/views/channels/ChannelListView.vue'),
      },
      {
        path: 'channels/create',
        name: 'ChannelCreate',
        component: () => import('@/views/channels/ChannelCreateView.vue'),
      },
      {
        path: 'channels/:id',
        name: 'ChannelDetail',
        component: () => import('@/views/channels/ChannelDetailView.vue'),
      },
      {
        path: 'channels/:id/edit',
        name: 'ChannelEdit',
        component: () => import('@/views/channels/ChannelEditView.vue'),
      },
      {
        path: 'sources',
        name: 'SourceList',
        component: () => import('@/views/sources/SourceListView.vue'),
      },
      {
        path: 'sources/create',
        name: 'SourceCreate',
        component: () => import('@/views/sources/SourceCreateView.vue'),
      },
      {
        path: 'sources/:id',
        name: 'SourceDetail',
        component: () => import('@/views/sources/SourceDetailView.vue'),
      },
      {
        path: 'sources/:id/edit',
        name: 'SourceEdit',
        component: () => import('@/views/sources/SourceEditView.vue'),
      },
      {
        path: 'messages',
        name: 'MessageList',
        component: () => import('@/views/messages/MessageListView.vue'),
      },
      {
        path: 'messages/:id',
        name: 'MessageDetail',
        component: () => import('@/views/messages/MessageDetailView.vue'),
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to, _from, next) => {
  const accessToken = localStorage.getItem('access_token')
  const refreshToken = localStorage.getItem('refresh_token')
  const isAuthenticated = !!(accessToken && refreshToken)

  if (to.meta.requiresAuth && !isAuthenticated) {
    next({ name: 'Login' })
  } else if (to.meta.requiresGuest && isAuthenticated) {
    next({ name: 'Dashboard' })
  } else {
    next()
  }
})

export default router
