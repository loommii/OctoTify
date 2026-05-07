<template>
  <div class="channel-create-page">
    <div class="page-header">
      <h2>创建推送渠道</h2>
      <button class="btn-back" @click="goBack">返回</button>
    </div>

    <div class="form-container">
      <div class="section-label">选择渠道类型</div>
      
      <div class="type-grid">
        <div
          v-for="t in channelTypes"
          :key="t.type"
          :class="['type-card', { selected: form.type === t.type }]"
          @click="selectType(t.type)"
        >
          <div class="type-icon">{{ getIcon(t.type) }}</div>
          <div class="type-info">
            <div class="type-name">{{ t.name }}</div>
            <div class="type-desc">{{ t.description }}</div>
          </div>
        </div>
      </div>

      <div v-if="selectedType" class="form-divider"></div>

      <div v-if="selectedType" class="form-section">
        <div class="form-group">
          <label>渠道名称</label>
          <input v-model="form.name" type="text" placeholder="例如：钉钉-运维群" />
        </div>

        <div class="config-fields">
          <div v-for="field in selectedType.config_fields" :key="field.name" class="form-group">
            <label>
              {{ field.label }}
              <span v-if="field.required" class="required">*</span>
            </label>
            <input
              v-if="field.type === 'text' || field.type === 'password' || field.type === 'url' || field.type === 'string'"
              :type="field.type === 'password' ? 'password' : 'text'"
              v-model="form.config[field.name]"
              :placeholder="field.placeholder"
              :required="field.required"
            />
            <input
              v-else-if="field.type === 'number'"
              type="number"
              v-model="form.config[field.name]"
              :placeholder="field.placeholder"
              :required="field.required"
            />
            <textarea
              v-else-if="field.type === 'textarea'"
              v-model="form.config[field.name]"
              :placeholder="field.placeholder"
              :required="field.required"
            ></textarea>
          </div>
        </div>

        <div class="form-actions">
          <button class="btn-primary" @click="handleSubmit" :disabled="submitting">
            {{ submitting ? '创建中...' : '创建' }}
          </button>
          <button class="btn-secondary" @click="goBack">取消</button>
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
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { getChannelTypes, createChannel } from '@/api/channels'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import { useConfirm } from '@/composables/useConfirm'
import { useToast } from '@/composables/useToast'
import type { ChannelTypeMeta } from '@/types/api'

const router = useRouter()
const channelTypes = ref<ChannelTypeMeta[]>([])
const submitting = ref(false)
const form = ref({
  type: '',
  name: '',
  config: {} as Record<string, string>,
})

const {
  showDialog,
  dialogOptions,
  actionLoading,
  requestConfirm,
  handleConfirm,
} = useConfirm()

const { visible: showToast, message: toastMessage, type: toastType, success: showSuccess, error: showError } = useToast()

const selectedType = computed(() => {
  return channelTypes.value.find((t) => t.type === form.value.type)
})

const icons: Record<string, string> = {
  wechat: '💬',
  telegram: '✈️',
  dingtalk: '🔔',
  email: '📧',
  webhook: '🔗',
  feishu: '🕊️',
}

function getIcon(type: string): string {
  return icons[type] || '📌'
}

function selectType(type: string) {
  if (form.value.type === type) return
  form.value.type = type
  form.value.config = {}
}

function validateForm(): boolean {
  if (!form.value.type) {
    showError('请选择渠道类型')
    return false
  }

  if (!form.value.name) {
    showError('请输入渠道名称')
    return false
  }

  if (!selectedType.value) return false

  for (const field of selectedType.value.config_fields) {
    if (field.required && !form.value.config[field.name]) {
      showError(`请填写 ${field.label}`)
      return false
    }
  }

  return true
}

async function loadChannelTypes() {
  try {
    const res = await getChannelTypes()
    if (res.data) {
      channelTypes.value = res.data
    }
  } catch (err) {
    console.error('加载渠道类型失败', err)
  }
}

function normalizeConfig(
  config: Record<string, string>,
  fields: { type: string; name: string }[]
): Record<string, unknown> {
  const normalized: Record<string, unknown> = {}
  for (const field of fields) {
    const value = config[field.name]
    if (field.type === 'number' && value !== '' && value !== undefined) {
      normalized[field.name] = Number(value)
    } else {
      normalized[field.name] = value
    }
  }
  return normalized
}

function handleSubmit() {
  if (!validateForm()) return
  if (!selectedType.value) return

  const normalizedConfig = normalizeConfig(form.value.config, selectedType.value.config_fields)

  requestConfirm(
    {
      title: '创建推送渠道',
      description: `确定要创建渠道 "${form.value.name}" 吗？`,
      confirmText: '创建',
      confirmType: 'primary',
    },
    async () => {
      submitting.value = true
      try {
        await createChannel({
          type: form.value.type,
          name: form.value.name,
          config: normalizedConfig,
        })
        showSuccess('创建成功', 1500)
        setTimeout(() => {
          router.push({ name: 'ChannelList' })
        }, 1500)
      } catch (err) {
        console.error('创建渠道失败', err)
        showError('创建失败，请重试')
      } finally {
        submitting.value = false
      }
    }
  )
}

function goBack() {
  router.back()
}

onMounted(() => {
  loadChannelTypes()
})
</script>

<style scoped>
.channel-create-page {
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

.section-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--mid-gray);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: var(--space-4);
}

.type-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--space-3);
}

.type-card {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-4);
  background: var(--near-black);
  border: 1px solid var(--border-dark);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
  min-width: 0;
}

.type-card:hover {
  border-color: var(--mid-border);
  background: var(--charcoal);
}

.type-card.selected {
  border-color: var(--green-link);
  background: rgba(0, 197, 115, 0.08);
}

.type-icon {
  font-size: 1.5rem;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--dark);
  border-radius: var(--radius-sm);
  flex-shrink: 0;
}

.type-card.selected .type-icon {
  background: rgba(0, 197, 115, 0.15);
}

.type-info {
  flex: 1;
  min-width: 0;
}

.type-name {
  font-size: 0.9375rem;
  font-weight: 500;
  color: var(--off-white);
  margin-bottom: 2px;
}

.type-desc {
  font-size: 0.75rem;
  color: var(--mid-gray);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.form-divider {
  margin-top: var(--space-8);
  border-top: 1px solid var(--border-dark);
}

.form-section {
  margin-top: var(--space-8);
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

.form-group input,
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

.form-group input:focus,
.form-group textarea:focus {
  outline: none;
  border-color: var(--green-link);
}

.form-group input::placeholder,
.form-group textarea::placeholder {
  color: var(--mid-gray);
}

.form-group textarea {
  min-height: 80px;
  resize: vertical;
}

.required {
  color: var(--error);
}

.config-fields {
  margin-top: var(--space-8);
  padding-top: var(--space-6);
  border-top: 1px solid var(--border-dark);
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
