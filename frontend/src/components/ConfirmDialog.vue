<script setup lang="ts">
defineProps<{
  visible: boolean
  title: string
  description: string
  confirmText?: string
  confirmType?: 'danger' | 'warning' | 'primary'
  loading?: boolean
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  confirm: []
}>()

function close() {
  emit('update:visible', false)
}

function handleConfirm() {
  emit('confirm')
}
</script>

<template>
  <div v-if="visible" class="modal-overlay" @click.self="close">
    <div class="modal">
      <h3 class="modal-title">{{ title }}</h3>
      <p class="modal-description">{{ description }}</p>

      <div class="modal-actions">
        <button
          class="btn btn-secondary"
          @click="close"
          :disabled="loading"
        >
          取消
        </button>
        <button
          class="btn"
          :class="['btn-' + (confirmType || 'primary'), 'btn-confirm']"
          @click="handleConfirm"
          :disabled="loading"
        >
          <span v-if="loading" class="spinner"></span>
          {{ loading ? '处理中...' : confirmText || '确认' }}
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

.btn-warning {
  background: var(--warning);
  color: var(--dark);
  border: none;
  font-weight: 600;
}

.btn-warning:hover:not(:disabled) {
  opacity: 0.8;
}

.btn-danger {
  background: var(--error);
  color: var(--off-white);
  border: none;
  font-weight: 600;
}

.btn-danger:hover:not(:disabled) {
  opacity: 0.8;
}

.btn:disabled {
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
