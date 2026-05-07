<template>
  <div class="channel-detail-page">
    <div class="page-header">
      <h2>渠道详情</h2>
      <div class="header-actions">
        <button class="btn-primary" @click="handleTest">测试连接</button>
        <button class="btn-secondary" @click="goToEdit">编辑</button>
        <button class="btn-back" @click="goBack">返回</button>
      </div>
    </div>

    <div v-if="loading" class="loading">加载中...</div>

    <div v-else-if="channel" class="detail-container">
      <div class="detail-card">
        <h3>基本信息</h3>
        <div class="detail-row">
          <span class="label">渠道名称</span>
          <span class="value">{{ channel.name }}</span>
        </div>
        <div class="detail-row">
          <span class="label">渠道类型</span>
          <span class="value">{{ getTypeName(channel.type) }}</span>
        </div>
        <div class="detail-row">
          <span class="label">状态</span>
          <span :class="['value', 'status-badge', getStatusClass(channel.status)]">
            {{ getStatusText(channel.status) }}
          </span>
        </div>
        <div class="detail-row">
          <span class="label">创建时间</span>
          <span class="value">{{ formatTime(channel.created_at) }}</span>
        </div>
        <div class="detail-row">
          <span class="label">更新时间</span>
          <span class="value">{{ formatTime(channel.updated_at) }}</span>
        </div>
        <div class="detail-row" v-if="channel.last_used_at">
          <span class="label">最后使用时间</span>
          <span class="value">{{ formatTime(channel.last_used_at) }}</span>
        </div>
      </div>

      <div class="detail-card">
        <h3>配置信息</h3>
        <div v-for="(value, key) in channel.config" :key="key" class="detail-row">
          <span class="label">{{ key }}</span>
          <span class="value">{{ value }}</span>
        </div>
      </div>
    </div>

    <ConfirmDialog
      v-model:visible="showDialog"
      :title="dialogOptions.title"
      :description="dialogOptions.description"
      :confirm-text="dialogOptions.confirmText"
      :confirm-type="dialogOptions.confirmType"
      :loading="actionLoading"
      @confirm="handleConfirm"
    />

    <Transition name="toast">
      <div v-if="showToast" :class="['toast', 'toast-' + toastType]">
        <svg v-if="toastType === 'success'" class="toast-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
          <polyline points="22 4 12 14.01 9 11.01"/>
        </svg>
        <svg v-else class="toast-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"/>
          <line x1="15" y1="9" x2="9" y2="15"/>
          <line x1="9" y1="9" x2="15" y2="15"/>
        </svg>
        <span>{{ toastMessage }}</span>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { getChannelDetail, testChannel } from '@/api/channels'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import { useConfirm } from '@/composables/useConfirm'
import { useToast } from '@/composables/useToast'
import type { ChannelDTO } from '@/types/api'

const router = useRouter()
const route = useRoute()
const channel = ref<ChannelDTO | null>(null)
const loading = ref(true)

const {
  showDialog,
  dialogOptions,
  actionLoading,
  requestConfirm,
  handleConfirm,
} = useConfirm()

const { visible: showToast, message: toastMessage, type: toastType, success: showSuccess, error: showError } = useToast()

const typeNames: Record<string, string> = {
  wechat: '企业微信',
  telegram: 'Telegram',
  dingtalk: '钉钉',
  email: '邮件',
  webhook: 'Webhook',
  feishu: '飞书',
}

function getTypeName(type: string): string {
  return typeNames[type] || type
}

function getStatusText(status: number): string {
  const map: Record<number, string> = { 1: '正常', 2: '已停用', '-1': '已删除' }
  return map[status] || '未知'
}

function getStatusClass(status: number): string {
  const map: Record<number, string> = { 1: 'status-active', 2: 'status-disabled', '-1': 'status-deleted' }
  return map[status] || ''
}

function formatTime(ts: number): string {
  return new Date(ts).toLocaleString('zh-CN')
}

async function loadChannel() {
  loading.value = true
  try {
    const id = Number(route.params.id)
    const res = await getChannelDetail(id)
    if (res.data) {
      channel.value = res.data
    }
  } catch (err) {
    console.error('加载渠道详情失败', err)
  } finally {
    loading.value = false
  }
}

function handleTest() {
  if (!channel.value) return
  requestConfirm(
    {
      title: '测试渠道连接',
      description: '将发送一条测试消息到该渠道，确认要继续吗？',
      confirmText: '测试',
      confirmType: 'primary',
    },
    async () => {
      try {
        await testChannel(channel.value!.id)
        showSuccess('测试消息已推送')
      } catch (err) {
        console.error('测试渠道连接失败', err)
        showError('测试失败，请检查配置')
      }
    }
  )
}

function goToEdit() {
  if (!channel.value) return
  router.push({ name: 'ChannelEdit', params: { id: channel.value.id } })
}

function goBack() {
  router.back()
}

onMounted(() => {
  loadChannel()
})
</script>

<style scoped>
.channel-detail-page {
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

.header-actions {
  display: flex;
  gap: var(--space-4);
}

.btn-primary {
  padding: 12px 32px;
  background: var(--green-link);
  color: var(--dark);
  border: none;
  border-radius: var(--radius-pill);
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all var(--transition-fast);
}

.btn-primary:hover {
  opacity: 0.8;
}

.btn-secondary {
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

.btn-secondary:hover {
  border-color: var(--mid-border);
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

.status-badge {
  padding: var(--space-1) var(--space-2);
  border-radius: var(--radius-sm);
  font-size: 0.75rem;
}

.status-active {
  background: rgba(62, 207, 142, 0.15);
  color: var(--success);
}

.status-disabled {
  background: rgba(245, 158, 11, 0.15);
  color: var(--warning);
}

.status-deleted {
  background: rgba(240, 95, 95, 0.15);
  color: var(--error);
}

.toast {
  position: fixed;
  top: 32px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 28px;
  border-radius: var(--radius-md);
  font-size: 0.9375rem;
  font-weight: 500;
  z-index: 9999;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(8px);
  min-width: 200px;
  justify-content: center;
}

.toast-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.toast-success {
  background: rgba(0, 197, 115, 0.95);
  color: var(--dark);
  border: 1px solid rgba(0, 197, 115, 0.3);
}

.toast-error {
  background: rgba(239, 68, 68, 0.95);
  color: white;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.toast-enter-active,
.toast-leave-active {
  transition: all 0.4s cubic-bezier(0.68, -0.55, 0.265, 1.55);
}

.toast-enter-from {
  opacity: 0;
  transform: translateX(-50%) translateY(-20px);
}

.toast-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(-20px);
}
</style>
