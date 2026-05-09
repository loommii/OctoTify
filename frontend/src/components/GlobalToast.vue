<script setup lang="ts">
import { useToast } from '@/composables/useToast'

const { visible, message, type } = useToast()
</script>

<template>
  <Transition name="toast">
    <div v-if="visible" :class="['global-toast', `toast-${type}`]">
      <span>{{ message }}</span>
    </div>
  </Transition>
</template>

<style scoped>
.global-toast {
  position: fixed;
  top: 24px;
  left: 50%;
  transform: translateX(-50%);
  padding: 12px 24px;
  border-radius: var(--radius-md);
  font-size: 0.875rem;
  font-weight: 500;
  z-index: 9999;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  white-space: nowrap;
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
