<template>
  <div class="channel-create-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <h2>创建推送渠道</h2>
      <button class="btn-back" @click="goBack">返回</button>
    </div>

    <!-- 表单容器 -->
    <div class="form-container">
      <!-- 渠道类型选择 -->
      <div class="section-label">选择渠道类型</div>

      <div class="type-grid">
        <div
          v-for="channelType in channelTypes"
          :key="channelType.type"
          :class="['type-card', { selected: form.type === channelType.type }]"
          @click="handleSelectType(channelType.type ?? '')"
        >
          <div class="type-icon">{{ getIcon(channelType.type ?? '') }}</div>
          <div class="type-info">
            <div class="type-name">{{ channelType.name }}</div>
            <div class="type-desc">{{ channelType.description }}</div>
          </div>
        </div>
      </div>

      <!-- 表单区域（选择类型后显示） -->
      <template v-if="selectedType">
        <div class="form-divider"></div>

        <div class="form-section">
          <!-- 微信 ClawBot 绑定区域 -->
          <WechatBindSection
            v-if="isBindMode"
            :bindQRCodeURL="bindQRCodeURL"
            :bindStatus="bindStatus"
            :qrcodeLoadError="qrcodeLoadError"
            @start-bind="startBind"
            @cancel-bind="cancelBind"
            @qrcode-error="handleQRCodeError"
          />

          <!-- 渠道名称 -->
          <div class="form-group">
            <label>渠道名称</label>
            <input v-model="form.name" type="text" placeholder="例如：微信-通知" />
          </div>

          <!-- 配置字段 -->
          <div class="config-fields">
            <div v-for="field in visibleConfigFields" :key="field.name" class="form-group">
              <label>
                {{ field.label }}
                <span v-if="field.required" class="required">*</span>
              </label>

              <!-- 凭证字段（绑定成功后脱敏展示，不可编辑） -->
              <template v-if="isCredentialField(field.name ?? '')">
                <input
                  type="text"
                  :value="maskValue(form.config[field.name ?? ''] ?? '')"
                  readonly
                  disabled
                  class="credential-masked"
                />
                <span class="credential-hint">已加密保护，用户无法查看或修改</span>
              </template>

              <!-- 文本/密码/URL 字段 -->
              <input
                v-else-if="isTextFieldType(field.type)"
                :type="field.type === 'password' ? 'password' : 'text'"
                v-model="form.config[field.name ?? '']"
                :placeholder="field.placeholder"
                :required="field.required"
                :readonly="isBindMode && bindStatus === 'confirmed'"
              />

              <!-- 数字字段 -->
              <input
                v-else-if="field.type === 'number'"
                type="number"
                v-model="form.config[field.name ?? '']"
                :placeholder="field.placeholder"
                :required="field.required"
                :readonly="isBindMode && bindStatus === 'confirmed'"
              />

              <!-- 多行文本字段 -->
              <textarea
                v-else-if="field.type === 'textarea'"
                v-model="form.config[field.name ?? '']"
                :placeholder="field.placeholder"
                :required="field.required"
                :readonly="isBindMode && bindStatus === 'confirmed'"
              ></textarea>
            </div>
          </div>

          <!-- 表单操作按钮 -->
          <div class="form-actions">
            <button class="btn-primary" @click="handleSubmitForm" :disabled="submitting">
              {{ submitting ? '创建中...' : '创建' }}
            </button>
            <button class="btn-secondary" @click="goBack">取消</button>
          </div>
        </div>
      </template>
    </div>

    <!-- 确认对话框 -->
    <ConfirmDialog
      v-model:visible="showDialog"
      :title="dialogOptions.title"
      :description="dialogOptions.description"
      :confirm-text="dialogOptions.confirmText"
      :confirm-type="dialogOptions.confirmType"
      :loading="actionLoading"
      @confirm="handleConfirm"
    />
  </div>
</template>

<script setup lang="ts">
/**
 * 渠道创建页面
 *
 * 职责：
 * - 整合微信绑定和渠道表单两个 Composable
 * - 管理 UI 层面的绑定模式状态
 * - 协调凭证数据的传递
 */
import { ref, onMounted, watch } from 'vue'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import WechatBindSection from '@/components/WechatBindSection.vue'
import { useWechatBind } from '@/composables/useWechatBind'
import { useChannelForm } from '@/composables/useChannelForm'

// ==================== 微信绑定状态 ====================

const {
  bindQRCodeURL,
  bindStatus,
  qrcodeLoadError,
  startBind,
  cancelBind,
  handleQRCodeError,
  getCredentials,
} = useWechatBind()

// ==================== UI 状态 ====================

/** 是否处于绑定模式（UI 层面，由渠道类型选择控制） */
const isBindMode = ref(false)

// ==================== 凭证自动填充 ====================

/**
 * 监听绑定状态变化
 *
 * 绑定成功时将凭证写入 form.config，触发配置字段自动填充
 * form.config 中保存原始值（用于提交），模板层通过 maskValue 脱敏展示
 */
watch(bindStatus, (newStatus) => {
  if (newStatus === 'confirmed') {
    const credentials = getCredentials()
    if (credentials) {
      Object.assign(form.value.config, credentials)
    }
  }
})

// ==================== 渠道表单 ====================

const {
  channelTypes,
  form,
  submitting,
  selectedType,
  visibleConfigFields,
  showDialog,
  dialogOptions,
  actionLoading,
  handleConfirm,
  isCredentialField,
  maskValue,
  selectType,
  handleSubmit,
  loadChannelTypes,
  goBack,
} = useChannelForm({
  onEnterBindMode: () => {
    isBindMode.value = true
  },
  onExitBindMode: () => {
    isBindMode.value = false
  },
  getBindStatus: () => bindStatus.value,
})

// ==================== 辅助方法 ====================

/** 渠道类型图标映射 */
const CHANNEL_ICONS: Record<string, string> = {
  wechat: '💬',
  wechat_clawbot: '🤖',
  telegram: '✈️',
  dingtalk: '🔔',
  email: '📧',
  webhook: '🔗',
  feishu: '🕊️',
}

/** 获取渠道类型对应的图标 */
function getIcon(type: string): string {
  return CHANNEL_ICONS[type] || '📌'
}

/** 处理渠道类型选择 */
function handleSelectType(type: string): void {
  selectType(type)
}

/** 判断是否为文本类型字段 */
function isTextFieldType(type: string | undefined): boolean {
  const textTypes = ['text', 'password', 'url', 'string']
  return type !== undefined && textTypes.includes(type)
}

/** 处理表单提交 */
function handleSubmitForm(): void {
  const credentials = bindStatus.value === 'confirmed' ? getCredentials() : null
  handleSubmit(credentials)
}

// ==================== 生命周期 ====================

/** 组件挂载时加载渠道类型 */
onMounted(() => {
  loadChannelTypes()
})
</script>

<style scoped>
/* 页面布局 */
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

/* 表单容器 */
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

/* 类型选择网格 */
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

/* 分隔线 */
.form-divider {
  margin-top: var(--space-8);
  border-top: 1px solid var(--border-dark);
}

.form-section {
  margin-top: var(--space-8);
}

/* 表单组 */
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

/* 凭证字段样式 */
.credential-masked {
  background: var(--near-black) !important;
  color: var(--mid-gray) !important;
  cursor: not-allowed !important;
  letter-spacing: 0.15em;
}

.credential-hint {
  display: inline-block;
  margin-top: 6px;
  font-size: 0.75rem;
  color: var(--mid-gray);
}

.config-fields {
  margin-top: var(--space-8);
  padding-top: var(--space-6);
  border-top: 1px solid var(--border-dark);
}

/* 表单操作按钮 */
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
</style>
