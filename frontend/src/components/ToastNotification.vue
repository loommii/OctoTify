/**
 * Toast 通知组件
 *
 * 职责：
 * - 显示成功/错误通知
 * - 自动动画过渡
 *
 * 设计原则：
 * - 无状态：完全由 props 驱动
 * - 单一职责：仅负责 UI 渲染
 * - 可复用：可在任何页面使用
 */
<template>
  <Transition name="toast">
    <div v-if="visible" :class="['toast', `toast-${type}`]">
      <!-- 成功图标 -->
      <svg v-if="type === 'success'" class="toast-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
        <polyline points="22 4 12 14.01 9 11.01" />
      </svg>

      <!-- 错误图标 -->
      <svg v-else class="toast-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10" />
        <line x1="15" y1="9" x2="9" y2="15" />
        <line x1="9" y1="9" x2="15" y2="15" />
      </svg>

      <span>{{ message }}</span>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import type { ToastType } from '@/composables/useToast'

// ============================================================
// Props 定义
// ============================================================

interface Props {
  /** Toast 是否可见 */
  visible: boolean
  /** Toast 消息内容 */
  message: string
  /** Toast 类型 */
  type: ToastType
}

defineProps<Props>()
</script>

<style scoped>
/* Toast 容器 */
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

/* 成功样式 */
.toast-success {
  background: rgba(0, 197, 115, 0.95);
  color: var(--dark);
  border: 1px solid rgba(0, 197, 115, 0.3);
}

/* 错误样式 */
.toast-error {
  background: rgba(239, 68, 68, 0.95);
  color: white;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

/* Transition 动画 */
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
