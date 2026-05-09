<template>
  <div class="message-detail-page">
    <div class="page-header">
      <h2>消息详情</h2>
      <button class="btn-back" @click="goBack">返回</button>
    </div>

    <div v-if="loading" class="loading">加载中...</div>

    <div v-else-if="message" class="detail-container">
      <div class="detail-card">
        <h3>基本信息</h3>
        <div class="detail-row">
          <span class="label">消息ID</span>
          <span class="value">{{ message.id }}</span>
        </div>
        <div class="detail-row">
          <span class="label">来源</span>
          <span class="value">{{ message.source_name }}</span>
        </div>
        <div class="detail-row">
          <span class="label">标题</span>
          <span class="value">{{ message.title }}</span>
        </div>
        <div class="detail-row">
          <span class="label">内容</span>
          <span class="value content-text">{{ message.content }}</span>
        </div>
        <div class="detail-row">
          <span class="label">状态</span>
          <span :class="['value', 'status-badge', getMessageStatusClass(message.status ?? 0)]">
            {{ getMessageStatusText(message.status ?? 0) }}
          </span>
        </div>
        <div class="detail-row">
          <span class="label">推送时间</span>
          <span class="value">{{ formatTime(message.created_at ?? 0) }}</span>
        </div>
      </div>

      <div class="detail-card">
        <h3>推送结果</h3>
        <div v-if="message.push_results && message.push_results.length > 0" class="push-results">
          <div v-for="result in message.push_results" :key="result.channel_id" class="push-result-item">
            <div class="result-header">
              <span class="channel-name">{{ result.channel_name || '未知渠道' }}</span>
              <span :class="['status-badge', getPushStatusClass(result)]">
                {{ getPushStatusText(result) }}
              </span>
            </div>
            <div class="result-detail">
              <span>渠道 ID: {{ result.channel_id }}</span>
              <span v-if="result.error" class="error-msg">错误: {{ result.error }}</span>
              <span>消息 ID: {{ result.message_id }}</span>
            </div>
          </div>
        </div>
        <p v-else class="empty-hint">暂无推送结果</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { getMessageDetail } from '@/api/channels'
import { getMessageStatusText, getMessageStatusClass } from '@/lib/constants'
import type { MessageDetailDTO, PushResult } from '@/types/api'

const router = useRouter()
const route = useRoute()
const message = ref<MessageDetailDTO | null>(null)
const loading = ref(true)

function formatTime(ts: number): string {
  return new Date(ts).toLocaleString('zh-CN')
}

function getPushStatusText(result: PushResult): string {
  if (result.success) return '成功'
  return '失败'
}

function getPushStatusClass(result: PushResult): string {
  return result.success ? 'status-success' : 'status-failed'
}

async function loadMessage() {
  loading.value = true
  try {
    const id = Number(route.params.id)
    const res = await getMessageDetail(id)
    if (res.data) {
      message.value = res.data
    }
  } catch (err) {
    console.error('加载消息详情失败', err)
  } finally {
    loading.value = false
  }
}

function goBack() {
  router.back()
}

onMounted(() => {
  loadMessage()
})
</script>

<style scoped>
.message-detail-page {
  padding: var(--space-8);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-8);
}

.page-header h2 {
  margin: 0;
  font-size: 2.25rem;
  font-weight: 400;
  color: var(--off-white);
}

.btn-back {
  padding: 12px 32px;
  background: transparent;
  color: var(--off-white);
  border: 1px solid var(--border-dark);
  border-radius: var(--radius-pill);
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all var(--transition-fast);
}

.btn-back:hover {
  border-color: var(--mid-border);
}

.loading {
  text-align: center;
  padding: var(--space-12);
  color: var(--mid-gray);
}

.detail-container {
  display: flex;
  flex-direction: column;
  gap: var(--space-8);
}

.detail-card {
  background: var(--dark);
  border: 1px solid var(--border-dark);
  border-radius: var(--radius-lg);
  padding: var(--space-8);
}

.detail-card h3 {
  margin: 0 0 var(--space-6);
  font-size: 1.125rem;
  font-weight: 500;
  color: var(--off-white);
  border-bottom: 1px solid var(--border-dark);
  padding-bottom: var(--space-4);
}

.detail-row {
  display: flex;
  padding: var(--space-4) 0;
  border-bottom: 1px solid var(--border-dark);
}

.detail-row:last-child {
  border-bottom: none;
}

.label {
  width: 140px;
  font-weight: 500;
  color: var(--mid-gray);
}

.value {
  flex: 1;
  color: var(--off-white);
}

.content-text {
  white-space: pre-wrap;
  word-break: break-word;
}

.status-badge {
  padding: var(--space-1) var(--space-2);
  border-radius: var(--radius-sm);
  font-size: 0.75rem;
}

.status-success {
  background: rgba(62, 207, 142, 0.15);
  color: var(--success);
}

.status-failed {
  background: rgba(240, 95, 95, 0.15);
  color: var(--error);
}

.status-partial {
  background: rgba(245, 158, 11, 0.15);
  color: var(--warning);
}

.push-results {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.push-result-item {
  border: 1px solid var(--border-dark);
  border-radius: var(--radius-sm);
  overflow: hidden;
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-3) var(--space-4);
  background: var(--near-black);
}

.channel-name {
  font-weight: 500;
  color: var(--off-white);
}

.result-detail {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  padding: var(--space-3) var(--space-4);
  font-size: 0.8125rem;
  color: var(--mid-gray);
}

.error-msg {
  color: var(--error);
}

.empty-hint {
  color: var(--mid-gray);
  font-size: 0.875rem;
}
</style>
