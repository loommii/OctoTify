<template>
  <div class="source-create-page">
    <div class="page-header">
      <h2>创建消息来源</h2>
      <button class="btn-back" @click="goBack">返回</button>
    </div>

    <div v-if="!createdSource" class="form-container">
      <div class="form-group">
        <label>来源名称</label>
        <input v-model="form.name" type="text" placeholder="例如：监控告警系统" />
      </div>

      <div class="form-group">
        <label>描述</label>
        <textarea v-model="form.description" placeholder="请输入描述（可选）"></textarea>
      </div>

      <div class="form-group">
        <label>关联渠道</label>
        <div v-if="channelsLoading" class="loading-sm">加载渠道列表中...</div>
        <div v-else-if="channels.length === 0" class="empty-hint">
          暂无可用渠道，请先 <router-link :to="{ name: 'ChannelCreate' }">创建渠道</router-link>
        </div>
        <div v-else class="channel-grid">
          <div
            v-for="ch in channels"
            :key="ch.id"
            :class="['channel-card', { selected: form.channel_ids.includes(ch.id) }]"
            @click="toggleChannel(ch.id)"
          >
            <div class="channel-icon">{{ getChannelIcon(ch.type) }}</div>
            <div class="channel-info">
              <div class="channel-name">{{ ch.name }}</div>
              <div class="channel-type">{{ getTypeName(ch.type) }}</div>
            </div>
            <div :class="['check-indicator', { checked: form.channel_ids.includes(ch.id) }]">
              <svg v-if="form.channel_ids.includes(ch.id)" viewBox="0 0 16 16" fill="currentColor">
                <path d="M13.854 3.646a.5.5 0 0 1 0 .708l-7 7a.5.5 0 0 1-.708 0l-3.5-3.5a.5.5 0 1 1 .708-.708L6.5 10.293l6.646-6.647a.5.5 0 0 1 .708 0z"/>
              </svg>
            </div>
          </div>
        </div>
        <div v-if="channels.length > 0 && form.channel_ids.length > 0" class="selected-count">
          已选择 {{ form.channel_ids.length }} 个渠道
        </div>
      </div>

      <div class="form-actions">
        <button class="btn-primary" @click="handleSubmit" :disabled="submitting">
          {{ submitting ? '创建中...' : '创建' }}
        </button>
        <button class="btn-secondary" @click="goBack">取消</button>
      </div>
    </div>

    <div v-else class="token-result">
      <div class="result-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
          <polyline points="22 4 12 14.01 9 11.01"/>
        </svg>
      </div>
      <h3 class="result-title">来源创建成功</h3>
      <p class="result-desc">请妥善保存以下 Token，它仅在创建时显示一次</p>

      <div class="token-box">
        <code class="token-value">{{ createdSource.token }}</code>
        <button class="btn-copy" @click="copyToken">
          {{ copied ? '已复制' : '复制' }}
        </button>
      </div>

      <div class="result-actions">
        <router-link class="btn-primary" :to="{ name: 'SourceList' }">完成</router-link>
        <button class="btn-secondary" @click="createAnother">继续创建</button>
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
        {{ toastMessage }}
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { createSource, listChannels } from '@/api/channels'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import { useConfirm } from '@/composables/useConfirm'
import type { ChannelDTO, SourceDTO } from '@/types/api'

const router = useRouter()
const submitting = ref(false)
const channelsLoading = ref(true)
const channels = ref<ChannelDTO[]>([])
const createdSource = ref<SourceDTO | null>(null)
const copied = ref(false)
const form = ref({
  name: '',
  description: '',
  channel_ids: [] as number[],
})
const toastMessage = ref('')
const toastType = ref<'success' | 'error'>('success')
const showToast = ref(false)

const {
  showDialog,
  dialogOptions,
  actionLoading,
  requestConfirm,
  handleConfirm,
} = useConfirm()

function showResult(message: string, type: 'success' | 'error') {
  toastMessage.value = message
  toastType.value = type
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

const channelIcons: Record<string, string> = {
  wechat: '💬',
  telegram: '✈️',
  dingtalk: '🔔',
  email: '📧',
  webhook: '🔗',
  feishu: '🕊️',
}

function getTypeName(type: string): string {
  return typeNames[type] || type
}

function getChannelIcon(type: string): string {
  return channelIcons[type] || '📌'
}

function toggleChannel(id: number) {
  const idx = form.value.channel_ids.indexOf(id)
  if (idx === -1) {
    form.value.channel_ids.push(id)
  } else {
    form.value.channel_ids.splice(idx, 1)
  }
}

async function loadChannels() {
  channelsLoading.value = true
  try {
    const res = await listChannels(1, 100)
    if (res.data) {
      channels.value = res.data.list.filter((ch) => ch.status === 1)
    }
  } catch (err) {
    console.error('加载渠道列表失败', err)
  } finally {
    channelsLoading.value = false
  }
}

function handleSubmit() {
  if (!form.value.name) {
    showResult('请输入来源名称', 'error')
    return
  }
  if (form.value.channel_ids.length === 0) {
    showResult('请至少选择一个关联渠道', 'error')
    return
  }

  requestConfirm(
    {
      title: '创建消息来源',
      description: `确定要创建来源 "${form.value.name}" 吗？`,
      confirmText: '创建',
      confirmType: 'primary',
    },
    async () => {
      submitting.value = true
      try {
        const res = await createSource({
          name: form.value.name,
          description: form.value.description,
          channel_ids: form.value.channel_ids,
        })
        if (res.data) {
          createdSource.value = res.data
        }
      } catch (err) {
        console.error('创建来源失败', err)
        showResult('创建来源失败', 'error')
      } finally {
        submitting.value = false
      }
    }
  )
}

async function copyToken() {
  if (!createdSource.value) return
  try {
    await navigator.clipboard.writeText(createdSource.value.token)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 2000)
  } catch (err) {
    console.error('复制失败', err)
  }
}

function createAnother() {
  createdSource.value = null
  form.value = {
    name: '',
    description: '',
    channel_ids: [],
  }
}

function goBack() {
  router.back()
}

onMounted(() => {
  loadChannels()
})
</script>

<style scoped>
.source-create-page {
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

.form-container {
  max-width: 800px;
  margin: 0 auto;
  background: var(--dark);
  border: 1px solid var(--border-dark);
  border-radius: var(--radius-lg);
  padding: var(--space-8);
}

.form-group {
  margin-bottom: var(--space-6);
}

.form-group label {
  display: block;
  margin-bottom: var(--space-2);
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--off-white);
}

.form-group input[type="text"],
.form-group textarea {
  width: 100%;
  padding: 12px 16px;
  background: var(--near-black);
  border: 1px solid var(--border-dark);
  border-radius: var(--radius-sm);
  font-size: 1rem;
  color: var(--off-white);
  box-sizing: border-box;
  transition: border-color var(--transition-fast);
}

.form-group input[type="text"]:focus,
.form-group textarea:focus {
  outline: none;
  border-color: var(--green-link);
}

.form-group input[type="text"]::placeholder,
.form-group textarea::placeholder {
  color: var(--mid-gray);
}

.form-group textarea {
  min-height: 80px;
  resize: vertical;
}

.loading-sm {
  padding: var(--space-4);
  color: var(--mid-gray);
  font-size: 0.875rem;
}

.empty-hint {
  padding: var(--space-4);
  color: var(--mid-gray);
  font-size: 0.875rem;
}

.empty-hint a {
  color: var(--green-link);
}

.channel-grid {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.channel-card {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  background: var(--near-black);
  border: 1px solid var(--border-dark);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.channel-card:hover {
  border-color: var(--mid-border);
  background: var(--charcoal);
}

.channel-card.selected {
  border-color: var(--green-link);
  background: rgba(0, 197, 115, 0.08);
}

.channel-icon {
  font-size: 1.25rem;
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--dark);
  border-radius: var(--radius-sm);
  flex-shrink: 0;
}

.channel-card.selected .channel-icon {
  background: rgba(0, 197, 115, 0.15);
}

.channel-info {
  flex: 1;
  min-width: 0;
}

.channel-name {
  font-size: 0.9375rem;
  font-weight: 500;
  color: var(--off-white);
  margin-bottom: 2px;
}

.channel-type {
  font-size: 0.75rem;
  color: var(--mid-gray);
}

.check-indicator {
  width: 20px;
  height: 20px;
  border: 2px solid var(--border-dark);
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: all var(--transition-fast);
}

.check-indicator.checked {
  background: var(--green-link);
  border-color: var(--green-link);
}

.check-indicator svg {
  width: 12px;
  height: 12px;
  color: var(--dark);
}

.selected-count {
  margin-top: var(--space-2);
  font-size: 0.8125rem;
  color: var(--green-link);
}

.form-actions {
  display: flex;
  gap: var(--space-4);
  margin-top: var(--space-8);
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
  text-decoration: none;
  text-align: center;
  display: inline-block;
  box-sizing: border-box;
}

.btn-primary:hover:not(:disabled) {
  opacity: 0.8;
}

.btn-primary:disabled {
  opacity: 0.3;
  cursor: not-allowed;
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

.token-result {
  max-width: 600px;
  margin: 0 auto;
  background: var(--dark);
  border: 1px solid var(--border-dark);
  border-radius: var(--radius-lg);
  padding: var(--space-10);
  text-align: center;
}

.result-icon {
  width: 64px;
  height: 64px;
  margin: 0 auto var(--space-6);
  background: rgba(0, 197, 115, 0.1);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--green-link);
}

.result-icon svg {
  width: 32px;
  height: 32px;
}

.result-title {
  font-size: 1.5rem;
  font-weight: 400;
  color: var(--off-white);
  margin: 0 0 var(--space-2);
}

.result-desc {
  font-size: 0.875rem;
  color: var(--mid-gray);
  margin: 0 0 var(--space-8);
}

.token-box {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-4);
  background: var(--near-black);
  border: 1px solid var(--border-dark);
  border-radius: var(--radius-md);
  margin-bottom: var(--space-8);
}

.token-value {
  flex: 1;
  font-family: 'SF Mono', 'Fira Code', monospace;
  font-size: 0.875rem;
  color: var(--green-link);
  word-break: break-all;
  text-align: left;
}

.btn-copy {
  padding: 8px 20px;
  background: var(--green-link);
  color: var(--dark);
  border: none;
  border-radius: var(--radius-pill);
  cursor: pointer;
  font-size: 0.8125rem;
  font-weight: 500;
  white-space: nowrap;
  transition: all var(--transition-fast);
}

.btn-copy:hover {
  opacity: 0.8;
}

.result-actions {
  display: flex;
  gap: var(--space-4);
  justify-content: center;
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

.toast-error {
  background: var(--error);
  color: var(--off-white);
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
