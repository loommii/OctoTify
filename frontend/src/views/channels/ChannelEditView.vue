<template>
  <div class="channel-edit-page">
    <div class="page-header">
      <h2>编辑推送渠道</h2>
      <button class="btn-back" @click="goBack">返回</button>
    </div>

    <div v-if="loading" class="loading">加载中...</div>

    <div v-else-if="channel" class="form-container">
      <div class="form-group">
        <label>渠道类型</label>
        <input :value="getTypeName(channel.type)" disabled />
      </div>

      <div class="form-group">
        <label>
          渠道名称
          <span class="required">*</span>
        </label>
        <input v-model="form.name" type="text" placeholder="请输入渠道名称" />
      </div>

      <div class="config-fields">
        <h3>配置信息</h3>
        <div v-for="field in configFields" :key="field.name" class="form-group">
          <label>
            {{ field.label }}
            <span v-if="field.required" class="required">*</span>
          </label>
          <input
            v-if="field.type === 'text' || field.type === 'password' || field.type === 'url' || field.type === 'string'"
            :type="field.type === 'password' ? 'password' : 'text'"
            v-model="form.config[field.name]"
            :placeholder="field.placeholder"
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
          ></textarea>
        </div>
      </div>

      <div class="form-actions">
        <button class="btn-primary" @click="handleSubmit" :disabled="submitting">
          {{ submitting ? '保存中...' : '保存' }}
        </button>
        <button class="btn-secondary" @click="goBack">取消</button>
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
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { getChannelDetail, updateChannel, getChannelTypes } from '@/api/channels'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import { useConfirm } from '@/composables/useConfirm'
import type { ChannelDTO, ChannelTypeMeta } from '@/types/api'

const router = useRouter()
const route = useRoute()
const channel = ref<ChannelDTO | null>(null)
const loading = ref(true)
const submitting = ref(false)
const channelTypes = ref<ChannelTypeMeta[]>([])
const form = ref({
  name: '',
  config: {} as Record<string, string>,
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

function getTypeName(type: string): string {
  return typeNames[type] || type
}

// 根据渠道类型动态获取配置字段元数据
const configFields = computed(() => {
  if (!channel.value) return []
  const typeMeta = channelTypes.value.find((t) => t.type === channel.value?.type)
  return typeMeta?.config_fields ?? []
})

async function loadChannel() {
  loading.value = true
  try {
    const id = Number(route.params.id)
    const res = await getChannelDetail(id)
    if (res.data) {
      channel.value = res.data
      form.value.name = res.data.name
      form.value.config = { ...res.data.config } as Record<string, string>
    }
  } catch (err) {
    console.error('加载渠道详情失败', err)
  } finally {
    loading.value = false
  }
}

// 加载渠道类型元数据，用于获取字段必填规则
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

// 校验必填字段
function validateForm(): boolean {
  if (!form.value.name) {
    showResult('请输入渠道名称', 'error')
    return false
  }

  for (const field of configFields.value) {
    if (field.required && !form.value.config[field.name]) {
      showResult(`请填写 ${field.label}`, 'error')
      return false
    }
  }

  return true
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
  if (!channel.value) return
  if (!validateForm()) return

  const normalizedConfig = normalizeConfig(form.value.config, configFields.value)

  requestConfirm(
    {
      title: '保存渠道修改',
      description: `确定要保存渠道 "${form.value.name}" 的修改吗？`,
      confirmText: '保存',
      confirmType: 'primary',
    },
    async () => {
      submitting.value = true
      try {
        await updateChannel(channel.value!.id, {
          name: form.value.name,
          config: normalizedConfig,
        })
        router.push({ name: 'ChannelDetail', params: { id: channel.value!.id } })
      } catch (err) {
        console.error('更新渠道失败', err)
        showResult('更新失败，请重试', 'error')
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
  loadChannel()
  loadChannelTypes()
})
</script>

<style scoped>
.channel-edit-page {
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

.required {
  color: #ff4d4f;
}

.form-group input {
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

.form-group input:focus {
  outline: none;
  border-color: var(--green-link);
}

.form-group input:disabled {
  background: var(--charcoal);
  color: var(--mid-gray);
  cursor: not-allowed;
}

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
  min-height: 80px;
  resize: vertical;
  font-family: inherit;
}

.form-group textarea:focus {
  outline: none;
  border-color: var(--green-link);
}

.config-fields {
  margin-top: var(--space-8);
  padding-top: var(--space-6);
  border-top: 1px solid var(--border-dark);
}

.config-fields h3 {
  margin: 0 0 var(--space-6);
  font-size: 1.125rem;
  font-weight: 500;
  color: var(--off-white);
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
