/**
 * 微信 ClawBot 绑定区域组件
 *
 * 职责：
 * - 展示绑定流程的各个状态 UI
 * - 处理用户交互（开始绑定、取消、二维码错误）
 *
 * 设计原则：
 * - 单一职责：仅负责 UI 展示和用户交互
 * - 状态驱动：根据 bindStatus 渲染不同状态
 * - 事件驱动：通过 emit 通知父组件处理业务逻辑
 */
<template>
  <div class="bind-section">
    <div class="bind-header">
      <h3>扫码绑定微信 ClawBot</h3>
    </div>

    <Transition name="bind-fade" mode="out-in">
      <!-- 初始状态：等待用户点击开始绑定 -->
      <div v-if="!bindQRCodeURL" class="bind-start" key="start">
        <p class="bind-desc">点击"开始绑定"按钮，使用微信扫描二维码完成绑定</p>
        <button class="btn-bind" @click="handleStartBind">开始绑定</button>
      </div>

      <!-- 等待扫码/确认状态 -->
      <div
        v-else-if="isPendingStatus"
        class="bind-pending"
        key="pending"
      >
        <div class="qrcode-wrapper">
          <qrcode-vue
            :value="bindQRCodeURL"
            :size="250"
            level="M"
            renderAs="svg"
            @error="handleQRCodeError"
            class="bind-qrcode"
          />
          <div class="polling-indicator">
            <span class="polling-dot"></span>
          </div>
        </div>

        <p v-if="qrcodeLoadError" class="bind-qrcode-error">
          二维码加载失败，请重新开始
        </p>

        <p v-if="bindStatus === 'wait' || bindStatus === 'pending'" class="bind-status">
          请使用微信扫描二维码完成绑定
        </p>
        <p v-else-if="bindStatus === 'scanned'" class="bind-status bind-status-scanned">
          已扫描，请在手机上确认绑定
        </p>
        <p v-else class="bind-status">状态加载中...</p>

        <p class="bind-hint-text">打开微信 → 点击右上角"+" → 选择"扫一扫"</p>
        <button class="btn-bind-cancel" @click="handleCancelBind">取消</button>
      </div>

      <!-- 绑定成功状态 -->
      <div v-else-if="bindStatus === 'confirmed'" class="bind-success" key="success">
        <div class="bind-success-icon">
          <svg class="success-check" viewBox="0 0 52 52">
            <circle class="success-circle" cx="26" cy="26" r="25" fill="none" />
            <path class="success-path" fill="none" d="M14.1 27.2l7.1 7.2 16.7-16.8" />
          </svg>
        </div>
        <p class="bind-success-text">绑定成功！</p>
        <p class="bind-hint">请在 24 小时内向 Bot 发送任意消息（如"你好"）以激活推送</p>
      </div>

      <!-- 二维码过期状态 -->
      <div v-else-if="bindStatus === 'expired'" class="bind-expired" key="expired">
        <div class="bind-expired-icon">
          <svg class="expired-icon" viewBox="0 0 52 52">
            <circle class="expired-circle" cx="26" cy="26" r="25" fill="none" />
            <path class="expired-path" fill="none" d="M16 16l20 20M36 16l-20 20" />
          </svg>
        </div>
        <p class="bind-expired-text">二维码已过期</p>
        <button class="btn-bind" @click="handleStartBind">重新开始</button>
      </div>

      <!-- 异常状态（兜底） -->
      <div v-else class="bind-error" key="error">
        <div class="bind-error-icon">
          <svg class="error-icon" viewBox="0 0 52 52">
            <circle class="error-circle" cx="26" cy="26" r="25" fill="none" />
            <path class="error-path" fill="none" d="M26 16v16M26 36h.01" />
          </svg>
        </div>
        <p class="bind-error-text">微信 ClawBot 异常</p>
        <button class="btn-bind" @click="handleStartBind">重新开始</button>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import QrcodeVue from 'qrcode.vue'
import type { BindStatus } from '@/types/api'

// ============================================================
// 常量
// ============================================================

/** 等待状态列表（这些状态下显示二维码和轮询指示器） */
const PENDING_STATUSES: readonly BindStatus[] = ['wait', 'pending', 'scanned'] as const

// ============================================================
// Props 定义
// ============================================================

interface Props {
  /** 二维码 URL */
  bindQRCodeURL: string
  /** 当前绑定状态 */
  bindStatus: BindStatus
  /** 二维码是否加载失败 */
  qrcodeLoadError: boolean
}

const props = defineProps<Props>()

// ============================================================
// Emits 定义
// ============================================================

interface Emits {
  /** 开始绑定 */
  (e: 'start-bind'): void
  /** 取消绑定 */
  (e: 'cancel-bind'): void
  /** 二维码加载失败 */
  (e: 'qrcode-error'): void
}

const emit = defineEmits<Emits>()

// ============================================================
// 计算属性
// ============================================================

/** 判断是否为等待状态（wait/pending/scanned） */
const isPendingStatus = computed(() => {
  return PENDING_STATUSES.includes(props.bindStatus)
})

// ============================================================
// 事件处理
// ============================================================

/** 处理开始绑定按钮点击 */
function handleStartBind(): void {
  emit('start-bind')
}

/** 处理取消绑定按钮点击 */
function handleCancelBind(): void {
  emit('cancel-bind')
}

/** 处理二维码加载失败 */
function handleQRCodeError(): void {
  emit('qrcode-error')
}
</script>

<style scoped>
/* 绑定区域容器 */
.bind-section {
  margin-bottom: var(--space-8);
  padding: var(--space-6);
  background: var(--near-black);
  border: 1px solid var(--border-dark);
  border-radius: var(--radius-md);
}

.bind-header h3 {
  margin: 0 0 var(--space-4);
  font-size: 1.125rem;
  font-weight: 500;
  color: var(--off-white);
}

/* 初始状态 */
.bind-start {
  text-align: center;
  padding: var(--space-6) 0;
}

.bind-desc {
  margin: 0 0 var(--space-4);
  color: var(--mid-gray);
  font-size: 0.875rem;
}

/* 按钮样式 */
.btn-bind {
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

.btn-bind:hover {
  opacity: 0.8;
}

/* 等待状态 */
.bind-pending {
  text-align: center;
  padding: var(--space-6) 0;
}

.bind-qrcode {
  width: 250px;
  height: 250px;
  border: 1px solid var(--border-dark);
  border-radius: var(--radius-md);
}

.qrcode-wrapper {
  position: relative;
  display: inline-block;
  margin-bottom: var(--space-4);
}

/* 轮询指示器 */
.polling-indicator {
  position: absolute;
  bottom: -12px;
  left: 50%;
  transform: translateX(-50%);
  width: 8px;
  height: 8px;
}

.polling-dot {
  display: block;
  width: 100%;
  height: 100%;
  background: var(--green-link);
  border-radius: 50%;
  animation: pollingPulse 2s ease-in-out infinite;
}

@keyframes pollingPulse {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.4;
    transform: scale(1.3);
  }
}

/* 状态文本 */
.bind-status {
  margin: 0 0 var(--space-3);
  color: var(--off-white);
  font-size: 0.9375rem;
  font-weight: 500;
}

.bind-status-scanned {
  color: var(--green-link);
}

.bind-hint-text {
  margin: 0 0 var(--space-4);
  color: var(--mid-gray);
  font-size: 0.8125rem;
}

.bind-qrcode-error {
  margin: 0 0 var(--space-4);
  color: var(--error);
  font-size: 0.8125rem;
}

.btn-bind-cancel {
  padding: 8px 24px;
  background: transparent;
  color: var(--off-white);
  border: 1px solid var(--border-dark);
  border-radius: var(--radius-pill);
  cursor: pointer;
  font-size: 0.8125rem;
  transition: all var(--transition-fast);
}

.btn-bind-cancel:hover {
  border-color: var(--mid-border);
}

/* 成功状态 */
.bind-success {
  text-align: center;
  padding: var(--space-6) 0;
}

.bind-success-icon {
  width: 52px;
  height: 52px;
  margin: 0 auto var(--space-4);
}

.success-check {
  width: 100%;
  height: 100%;
}

.success-circle {
  stroke: var(--green-link);
  stroke-width: 2;
  stroke-dasharray: 166;
  stroke-dashoffset: 166;
  animation: successCircle 0.6s cubic-bezier(0.65, 0, 0.45, 1) forwards;
}

.success-path {
  stroke: var(--green-link);
  stroke-width: 3;
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-dasharray: 48;
  stroke-dashoffset: 48;
  animation: successPath 0.3s cubic-bezier(0.65, 0, 0.45, 1) 0.5s forwards;
}

@keyframes successCircle {
  0% {
    stroke-dashoffset: 166;
  }
  100% {
    stroke-dashoffset: 0;
  }
}

@keyframes successPath {
  0% {
    stroke-dashoffset: 48;
  }
  100% {
    stroke-dashoffset: 0;
  }
}

.bind-success-text {
  margin: 0 0 var(--space-4);
  color: var(--green-link);
  font-size: 1.125rem;
  font-weight: 500;
  animation: fadeInUp 0.4s ease-out 0.3s both;
}

.bind-hint {
  margin: 0;
  padding: var(--space-3) var(--space-4);
  background: rgba(255, 193, 7, 0.1);
  border: 1px solid rgba(255, 193, 7, 0.3);
  border-radius: var(--radius-sm);
  color: #ffc107;
  font-size: 0.8125rem;
  animation: fadeInUp 0.4s ease-out 0.5s both;
}

/* 过期状态 */
.bind-expired {
  text-align: center;
  padding: var(--space-6) 0;
}

.bind-expired-icon {
  width: 52px;
  height: 52px;
  margin: 0 auto var(--space-4);
}

.expired-icon {
  width: 100%;
  height: 100%;
}

.expired-circle {
  stroke: var(--error);
  stroke-width: 2;
  stroke-dasharray: 166;
  stroke-dashoffset: 166;
  animation: expiredCircle 0.6s cubic-bezier(0.65, 0, 0.45, 1) forwards;
}

.expired-path {
  stroke: var(--error);
  stroke-width: 3;
  stroke-linecap: round;
  stroke-dasharray: 48;
  stroke-dashoffset: 48;
  animation: expiredPath 0.3s cubic-bezier(0.65, 0, 0.45, 1) 0.5s forwards;
}

@keyframes expiredCircle {
  0% {
    stroke-dashoffset: 166;
  }
  100% {
    stroke-dashoffset: 0;
  }
}

@keyframes expiredPath {
  0% {
    stroke-dashoffset: 48;
  }
  100% {
    stroke-dashoffset: 0;
  }
}

.bind-expired-text {
  margin: 0 0 var(--space-4);
  color: var(--error);
  font-size: 1rem;
  animation: fadeInUp 0.4s ease-out 0.3s both;
}

/* 错误状态 */
.bind-error {
  text-align: center;
  padding: var(--space-6) 0;
}

.bind-error-icon {
  width: 52px;
  height: 52px;
  margin: 0 auto var(--space-4);
}

.error-icon {
  width: 100%;
  height: 100%;
}

.error-circle {
  stroke: var(--error);
  stroke-width: 2;
  stroke-dasharray: 166;
  stroke-dashoffset: 166;
  animation: errorCircle 0.6s cubic-bezier(0.65, 0, 0.45, 1) forwards;
}

.error-path {
  stroke: var(--error);
  stroke-width: 3;
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-dasharray: 48;
  stroke-dashoffset: 48;
  animation: errorPath 0.3s cubic-bezier(0.65, 0, 0.45, 1) 0.5s forwards;
}

@keyframes errorCircle {
  0% {
    stroke-dashoffset: 166;
  }
  100% {
    stroke-dashoffset: 0;
  }
}

@keyframes errorPath {
  0% {
    stroke-dashoffset: 48;
  }
  100% {
    stroke-dashoffset: 0;
  }
}

.bind-error-text {
  margin: 0 0 var(--space-4);
  color: var(--error);
  font-size: 1rem;
  animation: fadeInUp 0.4s ease-out 0.3s both;
}

/* 通用动画 */
@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Transition 动画 */
.bind-fade-enter-active,
.bind-fade-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.bind-fade-enter-from {
  opacity: 0;
  transform: translateY(10px);
}

.bind-fade-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
