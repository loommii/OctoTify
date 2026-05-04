<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const user = authStore.user
const isDropdownOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)

const toggleDropdown = () => {
  isDropdownOpen.value = !isDropdownOpen.value
}

const closeDropdown = () => {
  isDropdownOpen.value = false
}

const handleClickOutside = (event: MouseEvent) => {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target as Node)) {
    closeDropdown()
  }
}

const handleLogout = async () => {
  await authStore.logout()
  router.push({ name: 'Login' })
}

onMounted(async () => {
  document.addEventListener('click', handleClickOutside)
  if (!authStore.user) {
    try {
      await authStore.fetchProfile()
    } catch {
      router.push({ name: 'Login' })
    }
  }
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<template>
  <div class="main-layout">
    <header class="header">
      <div class="header-content">
        <div class="header-left">
          <router-link :to="{ name: 'Dashboard' }" class="logo" v-once>
            OctoTify
          </router-link>
          <nav class="nav">
            <router-link :to="{ name: 'Dashboard' }" class="nav-link" v-once>
              首页
            </router-link>
          </nav>
        </div>
        <div class="header-right">
          <div class="user-menu">
            <span class="username">{{ user?.username }}</span>
            <div ref="dropdownRef" class="dropdown">
              <button class="dropdown-trigger" @click="toggleDropdown">
                <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
                  <path d="M4 6l4 4 4-4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
              </button>
              <div v-if="isDropdownOpen" class="dropdown-menu">
                <router-link :to="{ name: 'Profile' }" class="dropdown-item" @click="closeDropdown">
                  个人资料
                </router-link>
                <router-link :to="{ name: 'ChangePassword' }" class="dropdown-item" @click="closeDropdown">
                  修改密码
                </router-link>
                <div class="dropdown-divider"></div>
                <button class="dropdown-item logout-btn" @click="handleLogout">
                  退出登录
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </header>

    <main class="main-content">
      <router-view />
    </main>
  </div>
</template>

<style scoped>
.main-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background-color: var(--dark);
}

.header {
  background-color: var(--dark);
  border-bottom: 1px solid var(--border-dark);
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 var(--space-6);
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--space-8);
}

.logo {
  font-size: 1.25rem;
  font-weight: 500;
  color: var(--off-white);
  text-decoration: none;
}

.logo:hover {
  color: var(--green-link);
  opacity: 1;
}

.nav {
  display: flex;
  gap: var(--space-6);
}

.nav-link {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--mid-gray);
  text-decoration: none;
  transition: color var(--transition-fast);
}

.nav-link:hover,
.nav-link.router-link-active {
  color: var(--off-white);
}

.header-right {
  display: flex;
  align-items: center;
}

.user-menu {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  position: relative;
}

.username {
  font-size: 0.875rem;
  color: var(--off-white);
}

.dropdown {
  position: relative;
}

.dropdown-trigger {
  padding: var(--space-2);
  color: var(--mid-gray);
  transition: color var(--transition-fast);
}

.dropdown-trigger:hover {
  color: var(--off-white);
}

.dropdown-menu {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: var(--space-2);
  background-color: var(--near-black);
  border: 1px solid var(--border-dark);
  border-radius: var(--radius-md);
  min-width: 180px;
  padding: var(--space-2) 0;
  box-shadow: var(--transition-fast);
}

.dropdown-item {
  display: block;
  width: 100%;
  padding: var(--space-2) var(--space-4);
  font-size: 0.875rem;
  color: var(--off-white);
  text-align: left;
  text-decoration: none;
  transition: background-color var(--transition-fast);
}

.dropdown-item:hover {
  background-color: var(--border-dark);
  opacity: 1;
}

.dropdown-divider {
  height: 1px;
  background-color: var(--border-dark);
  margin: var(--space-2) 0;
}

.logout-btn {
  color: var(--error);
}

.logout-btn:hover {
  background-color: rgba(240, 95, 95, 0.1);
}

.main-content {
  flex: 1;
  max-width: 1200px;
  width: 100%;
  margin: 0 auto;
  padding: var(--space-8) var(--space-6);
}
</style>
