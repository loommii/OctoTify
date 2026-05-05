<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'

const props = defineProps<{
  visible: boolean
  title: string
  description: string
  confirmText?: string
  loading?: boolean
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  confirm: [password: string]
}>()

const password = ref('')
const error = ref('')
const inputRef = ref<HTMLInputElement>()

watch(() => props.visible, (val) => {
  if (val) {
    password.value = ''
    error.value = ''
    nextTick(() => {
      inputRef.value?.focus()
    })
  }
})

function close() {
  if (!props.loading) {
    emit('update:visible', false)
  }
}

function handleConfirm() {
  if (!password.value) {
    error.value = '请输入密码'
    return
  }
  error.value = ''
  emit('confirm', password.value)
}

function handleKeyup(e: KeyboardEvent) {
  if (e.key === 'Enter') {
    handleConfirm()
  }
}
</script>

<template>
  <div v-if="visible" class="modal-overlay" @click.self="close">
    <div class="modal">
      <h3 class="modal-title">{{ title }}</h3>
      <p class="modal-description">{{ description }}</p>

      <div class="form-group">
        <label class="form-label" for="password-confirm">请输入登录密码</label>
        <input
          id="password-confirm"
          ref="inputRef"
          v-model="password"
          type="password"
          class="form-input"
          placeholder="请输入密码"
          :disabled="loading"
          @keyup="handleKeyup"
        />
        <p v-if="error" class="error-msg">{{ error }}</p>
      </div>

      <div class="modal-actions">
        <button
          class="btn btn-secondary"
          @click="close"
          :disabled="loading"
        >
          取消
        </button>
        <button
          class="btn btn-primary"
          @click="handleConfirm"
          :disabled="loading || !password"
        >
          <span v-if="loading" class="spinner"></span>
          {{ loading ? '验证中...' : confirmText || '确认' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
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

.modal-title {
  margin: 0 0 var(--space-2);
  font-size: 1.25rem;
  font-weight: 500;
  color: var(--off-white);
}

.modal-description {
  margin: 0 0 var(--space-6);
  font-size: 0.875rem;
  color: var(--mid-gray);
}

.form-group {
  margin-bottom: var(--space-6);
}

.form-label {
  display: block;
  margin-bottom: var(--space-2);
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--off-white);
}

.form-input {
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

.form-input:focus {
  outline: none;
  border-color: var(--green-link);
}

.form-input:disabled {
  background: var(--charcoal);
  color: var(--mid-gray);
  cursor: not-allowed;
}

.error-msg {
  margin-top: var(--space-2);
  font-size: 0.8125rem;
  color: var(--error);
}

.modal-actions {
  display: flex;
  gap: var(--space-4);
  justify-content: flex-end;
}

.btn {
  padding: 12px 32px;
  border-radius: var(--radius-pill);
  font-size: 0.875rem;
  font-weight: 500;
  transition: all var(--transition-fast);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
}

.btn-primary {
  background: var(--green-link);
  color: var(--dark);
  border: none;
  font-weight: 600;
}

.btn-primary:hover:not(:disabled) {
  opacity: 0.8;
}

.btn-primary:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.btn-secondary {
  background: transparent;
  color: var(--off-white);
  border: 1px solid var(--border-dark);
}

.btn-secondary:hover:not(:disabled) {
  border-color: var(--mid-border);
}

.btn-secondary:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid var(--dark);
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
