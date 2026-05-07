<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { getSourceDetail } from '@/api/channels'
import PasswordConfirmDialog from '@/components/PasswordConfirmDialog.vue'
import { usePasswordConfirm } from '@/composables/usePasswordConfirm'
import type { SourceDetailDTO, ChannelDTO } from '@/types/api'

const router = useRouter()
const route = useRoute()
const source = ref<SourceDetailDTO | null>(null)
const channels = ref<ChannelDTO[]>([])
const loading = ref(true)
const tokenResult = ref<string | null>(null)
const toastMessage = ref('')
const showToast = ref(false)

const {
  showDialog,
  dialogOptions,
  actionLoading,
  requestPassword,
  handleConfirm,
} = usePasswordConfirm()

function showResult(message: string) {
  toastMessage.value = message
  showToast.value = true
  setTimeout(() => {
    showToast.value = false
  }, 3000)
}

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

async function loadSource() {
  loading.value = true
  try {
    const id = Number(route.params.id)
    const res = await getSourceDetail(id)
    if (res.data) {
      source.value = res.data
      channels.value = res.data.channels || []
    }
  } catch (err) {
    console.error('加载来源详情失败', err)
  } finally {
    loading.value = false
  }
}

function handleViewToken() {
  if (!source.value) return
  requestPassword(
    {
      title: '查看来源令牌',
      description: '请输入登录密码以验证身份，令牌仅在验证成功后显示。',
      confirmText: '验证并查看',
    },
    async (pwd: string) => {
      const { getSourceToken } = await import('@/api/channels')
      const res = await getSourceToken(source.value!.id, pwd)
      if (res.data) {
        tokenResult.value = res.data.token
      }
    }
  )
}

function copyToken() {
  if (tokenResult.value) {
    navigator.clipboard.writeText(tokenResult.value)
    showResult('令牌已复制到剪贴板')
  }
}

function goToEdit() {
  if (!source.value) return
  router.push({ name: 'SourceEdit', params: { id: source.value.id } })
}

function goBack() {
  router.back()
}

onMounted(() => {
  loadSource()
})
</script>

<template>
  <div class="source-detail-page">
    <div class="page-header">
      <h2>来源详情</h2>
      <div class="header-actions">
        <button class="btn-secondary" @click="goToEdit">编辑</button>
        <button class="btn-back" @click="goBack">返回</button>
      </div>
    </div>

    <div v-if="loading" class="loading">加载中...</div>

    <div v-else-if="source" class="detail-container">
      <div class="detail-card">
        <h3>基本信息</h3>
        <div class="detail-row">
          <span class="label">来源名称</span>
          <span class="value">{{ source.name }}</span>
        </div>
        <div class="detail-row">
          <span class="label">描述</span>
          <span class="value">{{ source.description || '-' }}</span>
        </div>
        <div class="detail-row">
          <span class="label">状态</span>
          <span :class="['value', 'status-badge', getStatusClass(source.status)]">
            {{ getStatusText(source.status) }}
          </span>
        </div>
        <div class="detail-row">
          <span class="label">创建时间</span>
          <span class="value">{{ formatTime(source.created_at) }}</span>
        </div>
        <div class="detail-row">
          <span class="label">更新时间</span>
          <span class="value">{{ formatTime(source.updated_at) }}</span>
        </div>
        <div class="detail-row" v-if="source.last_used_at">
          <span class="label">最后使用时间</span>
          <span class="value">{{ formatTime(source.last_used_at) }}</span>
        </div>
      </div>

      <div class="detail-card">
        <h3>关联渠道</h3>
        <div v-if="channels.length > 0" class="channel-list">
          <div v-for="ch in channels" :key="ch.id" class="channel-item">
            <span>{{ ch.name }}</span>
            <span class="channel-type">{{ getTypeName(ch.type) }}</span>
            <span :class="['status-badge', getStatusClass(ch.status)]">
              {{ getStatusText(ch.status) }}
            </span>
          </div>
        </div>
        <p v-else class="empty-hint">暂无关联渠道</p>
      </div>

      <div class="detail-card">
        <h3>推送令牌</h3>
        <div class="token-section">
          <p class="token-hint">需要密码验证才能查看令牌</p>
          <button class="btn-primary" @click="handleViewToken">查看令牌</button>
        </div>
      </div>
    </div>

    <PasswordConfirmDialog
      v-model:visible="showDialog"
      :title="dialogOptions.title"
      :description="dialogOptions.description"
      :confirm-text="dialogOptions.confirmText"
      :loading="actionLoading"
      @confirm="handleConfirm"
    />

    <Transition name="toast">
      <div v-if="showToast" class="toast toast-success">
        {{ toastMessage }}
      </div>
    </Transition>

    <div v-if="tokenResult" class="modal-overlay" @click.self="tokenResult = null">
      <div class="modal">
        <h3>来源令牌</h3>
        <div class="token-display">
          <code>{{ tokenResult }}</code>
        </div>
        <p class="token-hint">请使用此令牌调用推送接口</p>
        <div class="modal-actions">
          <button class="btn-primary" @click="copyToken">复制令牌</button>
          <button class="btn-secondary" @click="tokenResult = null">关闭</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.source-detail-page {
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

.channel-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.channel-item {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-3) var(--space-4);
  background: var(--near-black);
  border-radius: var(--radius-sm);
}

.channel-type {
  color: var(--mid-gray);
  font-size: 0.8125rem;
}

.empty-hint {
  color: var(--mid-gray);
  font-size: 0.875rem;
}

.token-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.token-hint {
  font-size: 0.8125rem;
  color: var(--mid-gray);
  margin: 0;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: var(--dark);
  border: 1px solid var(--border-dark);
  border-radius: var(--radius-lg);
  padding: var(--space-8);
  width: 420px;
  max-width: 90%;
}

.modal h3 {
  margin: 0 0 var(--space-6);
  font-size: 1.125rem;
  font-weight: 500;
  color: var(--off-white);
}

.token-display {
  background: var(--near-black);
  border: 1px solid var(--border-dark);
  padding: var(--space-4);
  border-radius: var(--radius-sm);
  margin-bottom: var(--space-4);
  word-break: break-all;
}

.token-display code {
  font-size: 0.8125rem;
  color: var(--off-white);
  font-family: var(--font-mono);
}

.modal-actions {
  display: flex;
  gap: var(--space-4);
  justify-content: flex-end;
}

.toast {
  position: fixed;
  top: 24px;
  left: 50%;
  transform: translateX(-50%);
  padding: 12px 24px;
  border-radius: var(--radius-md);
  font-size: 0.875rem;
  font-weight: 500;
  z-index: 2000;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.toast-success {
  background: var(--success);
  color: var(--dark);
}

.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
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
