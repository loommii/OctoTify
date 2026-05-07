<template>
  <div class="page">
    <div class="page-header">
      <h1 class="page-title">欢迎使用 OctoTify</h1>
      <p class="page-description">消息总线平台 - 管理你的消息来源和推送渠道</p>
    </div>

    <div class="stats-grid">
      <router-link class="stat-card" to="/sources">
        <div class="stat-icon">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/>
          </svg>
        </div>
        <div class="stat-content">
          <span class="stat-label">消息来源</span>
          <span class="stat-value">{{ sourceCount }}</span>
        </div>
      </router-link>

      <router-link class="stat-card" to="/channels">
        <div class="stat-icon">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M22 2L11 13M22 2l-7 20-4-9-9-4 20-7z"/>
          </svg>
        </div>
        <div class="stat-content">
          <span class="stat-label">推送渠道</span>
          <span class="stat-value">{{ channelCount }}</span>
        </div>
      </router-link>

      <router-link class="stat-card" to="/messages">
        <div class="stat-icon">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/>
            <polyline points="22,6 12,13 2,6"/>
          </svg>
        </div>
        <div class="stat-content">
          <span class="stat-label">消息总数</span>
          <span class="stat-value">{{ messageCount }}</span>
        </div>
      </router-link>
    </div>

    <div class="quick-actions">
      <h2 class="section-title">快速开始</h2>
      <div class="action-grid">
        <router-link class="action-card" to="/sources/create">
          <h3 class="action-title">创建消息来源</h3>
          <p class="action-desc">创建一个新的消息来源，获取推送 Token</p>
        </router-link>

        <router-link class="action-card" to="/channels/create">
          <h3 class="action-title">配置推送渠道</h3>
          <p class="action-desc">添加微信、Telegram、钉钉等推送渠道</p>
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { listSources, listChannels, listMessages } from '@/api/channels'

const sourceCount = ref(0)
const channelCount = ref(0)
const messageCount = ref(0)

// 并发请求获取三项统计数据
async function loadStats() {
  try {
    const [sourcesRes, channelsRes, messagesRes] = await Promise.all([
      listSources(1, 1),
      listChannels(1, 1),
      listMessages(1, 1),
    ])
    sourceCount.value = sourcesRes.data?.total ?? 0
    channelCount.value = channelsRes.data?.total ?? 0
    messageCount.value = messagesRes.data?.total ?? 0
  } catch (err) {
    console.error('加载统计数据失败', err)
  }
}

onMounted(() => {
  loadStats()
})
</script>

<style scoped>
.page {
  max-width: 1000px;
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

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--space-4);
  margin-bottom: var(--space-10);
}

.stat-card {
  background-color: var(--dark);
  border: 1px solid var(--border-dark);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  display: flex;
  align-items: center;
  gap: var(--space-4);
}

.stat-icon {
  width: 48px;
  height: 48px;
  background-color: var(--near-black);
  border: 1px solid var(--border-dark);
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--green-link);
}

.stat-content {
  display: flex;
  flex-direction: column;
}

.stat-label {
  font-size: 0.875rem;
  color: var(--mid-gray);
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 400;
  color: var(--off-white);
}

.quick-actions {
  margin-top: var(--space-8);
}

.section-title {
  font-size: 1.125rem;
  font-weight: 500;
  color: var(--off-white);
  margin-bottom: var(--space-6);
}

.action-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: var(--space-4);
}

.action-card {
  background-color: var(--dark);
  border: 1px solid var(--border-dark);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  text-decoration: none;
  transition: border-color var(--transition-fast);
  cursor: pointer;
}

.action-card:hover {
  border-color: var(--green-border);
}

.action-title {
  font-size: 1rem;
  font-weight: 500;
  color: var(--off-white);
  margin-bottom: var(--space-2);
}

.action-desc {
  font-size: 0.875rem;
  color: var(--mid-gray);
}
</style>
