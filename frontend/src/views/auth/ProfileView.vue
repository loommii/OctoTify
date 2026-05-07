<script setup lang="ts">
import { computed } from 'vue'
import { useAuth } from '@/composables/useAuth'

const { user } = useAuth()

const formattedCreatedAt = computed(() => {
  if (!user.value?.created_at) return ''
  const date = new Date(user.value.created_at)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
})
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h1 class="page-title">个人资料</h1>
      <p class="page-description">查看和管理你的账户信息</p>
    </div>

    <div class="card">
      <div class="profile-section">
        <h2 class="section-title">基本信息</h2>

        <div class="info-grid">
          <div class="info-item">
            <span class="info-label">用户 ID</span>
            <span class="info-value">{{ user?.id }}</span>
          </div>

          <div class="info-item">
            <span class="info-label">用户名</span>
            <span class="info-value">{{ user?.username }}</span>
          </div>

          <div class="info-item">
            <span class="info-label">注册时间</span>
            <span class="info-value">{{ formattedCreatedAt }}</span>
          </div>
        </div>
      </div>

      <div class="profile-actions">
        <router-link :to="{ name: 'ChangePassword' }" class="btn btn-secondary">
          修改密码
        </router-link>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page {
  max-width: 600px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: var(--space-8);
}

.page-title {
  font-size: 2.25rem;
  font-weight: 400;
  line-height: 1.25;
  color: var(--off-white);
  margin-bottom: var(--space-2);
}

.page-description {
  font-size: 1rem;
  color: var(--mid-gray);
}

.card {
  background-color: var(--dark);
  border: 1px solid var(--border-dark);
  border-radius: var(--radius-lg);
  padding: var(--space-8);
}

.profile-section {
  margin-bottom: var(--space-8);
}

.section-title {
  font-size: 1.125rem;
  font-weight: 500;
  color: var(--off-white);
  margin-bottom: var(--space-6);
}

.info-grid {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-4) 0;
  border-bottom: 1px solid var(--border-dark);
}

.info-item:last-child {
  border-bottom: none;
}

.info-label {
  font-size: 0.875rem;
  color: var(--mid-gray);
}

.info-value {
  font-size: 0.875rem;
  color: var(--off-white);
  font-family: var(--font-mono);
}

.profile-actions {
  display: flex;
  gap: var(--space-4);
  padding-top: var(--space-6);
  border-top: 1px solid var(--border-dark);
}

.btn {
  padding: 12px 32px;
  border-radius: var(--radius-pill);
  font-size: 0.875rem;
  font-weight: 500;
  transition: all var(--transition-fast);
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.btn-secondary {
  background-color: transparent;
  color: var(--off-white);
  border: 1px solid var(--border-dark);
}

.btn-secondary:hover {
  border-color: var(--mid-border);
  opacity: 1;
}
</style>
