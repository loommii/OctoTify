<script setup lang="ts">
import { ref, computed } from 'vue'
import { useAuth } from '@/composables/useAuth'
import type { ChangePasswordReq } from '@/types/api'

const form = ref<ChangePasswordReq>({
  old_password: '',
  new_password: '',
})

const confirmPassword = ref('')

const { isLoading, errorMessage, handleChangePassword } = useAuth()

const isPasswordValid = computed(() => {
  const pwd = form.value.new_password
  return (
    pwd.length >= 8 &&
    pwd.length <= 128 &&
    /[a-z]/.test(pwd) &&
    /[A-Z]/.test(pwd) &&
    /[0-9]/.test(pwd)
  )
})

const passwordErrors = computed(() => {
  const pwd = form.value.new_password
  const errors: string[] = []
  if (pwd.length > 0 && pwd.length < 8) errors.push('至少 8 个字符')
  if (pwd.length > 128) errors.push('最多 128 个字符')
  if (pwd.length > 0 && !/[a-z]/.test(pwd)) errors.push('需包含小写字母')
  if (pwd.length > 0 && !/[A-Z]/.test(pwd)) errors.push('需包含大写字母')
  if (pwd.length > 0 && !/[0-9]/.test(pwd)) errors.push('需包含数字')
  return errors
})

const isFormValid = computed(() => {
  return (
    form.value.old_password.length > 0 &&
    isPasswordValid.value &&
    form.value.new_password === confirmPassword.value
  )
})

const submitForm = () => {
  if (!isFormValid.value) {
    return
  }
  handleChangePassword(form.value)
}
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h1 class="page-title">修改密码</h1>
      <p class="page-description">修改你的登录密码。修改成功后需要重新登录。</p>
    </div>

    <div class="card">
      <form class="form" @submit.prevent="submitForm">
        <div class="form-group">
          <label class="form-label" for="oldPassword">旧密码</label>
          <input
            id="oldPassword"
            v-model="form.old_password"
            type="password"
            class="form-input"
            placeholder="输入当前密码"
            autocomplete="current-password"
            required
          />
        </div>

        <div class="form-group">
          <label class="form-label" for="newPassword">新密码</label>
          <input
            id="newPassword"
            v-model="form.new_password"
            type="password"
            class="form-input"
            placeholder="8-128 个字符，需包含大小写字母和数字"
            autocomplete="new-password"
            required
            minlength="8"
            maxlength="128"
          />
          <span class="form-hint">8-128 个字符，需包含大小写字母和数字</span>
          <ul v-if="passwordErrors.length > 0" class="password-requirements">
            <li v-for="err in passwordErrors" :key="err" class="requirement-error">{{ err }}</li>
          </ul>
        </div>

        <div class="form-group">
          <label class="form-label" for="confirmPassword">确认新密码</label>
          <input
            id="confirmPassword"
            v-model="confirmPassword"
            type="password"
            class="form-input"
            placeholder="再次输入新密码"
            autocomplete="new-password"
            required
          />
        </div>

        <div v-if="errorMessage" class="alert alert-error">
          {{ errorMessage }}
        </div>

        <div class="form-actions">
          <button type="submit" class="btn btn-primary" :disabled="isLoading || !isFormValid">
            <span v-if="isLoading" class="spinner"></span>
            {{ isLoading ? '修改中...' : '修改密码' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
.page {
  max-width: 600px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: var(--space-8);
}

.page-title {
  font-size: 2.25rem;
  font-weight: 400;
  line-height: 1.25;
  color: var(--off-white);
  margin-bottom: var(--space-2);
}

.page-description {
  font-size: 1rem;
  color: var(--mid-gray);
}

.card {
  background-color: var(--dark);
  border: 1px solid var(--border-dark);
  border-radius: var(--radius-lg);
  padding: var(--space-8);
}

.form {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.form-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--off-white);
}

.form-input {
  padding: 12px 16px;
  background-color: var(--near-black);
  border: 1px solid var(--border-dark);
  border-radius: var(--radius-sm);
  color: var(--off-white);
  font-size: 1rem;
  transition: border-color var(--transition-fast);
}

.form-input:focus {
  outline: none;
  border-color: var(--green-link);
}

.form-input::placeholder {
  color: var(--mid-gray);
}

.form-hint {
  font-size: 0.75rem;
  color: var(--mid-gray);
}

.password-requirements {
  list-style: none;
  padding: 0;
  margin: var(--space-1) 0 0;
}

.requirement-error {
  font-size: 0.75rem;
  color: var(--error);
}

.alert {
  padding: 12px 16px;
  border-radius: var(--radius-sm);
  font-size: 0.875rem;
}

.alert-error {
  background-color: rgba(240, 95, 95, 0.1);
  border: 1px solid var(--error);
  color: var(--error);
}

.alert-success {
  background-color: rgba(62, 207, 142, 0.1);
  border: 1px solid var(--green-link);
  color: var(--green-link);
}

.form-actions {
  margin-top: var(--space-4);
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
  background-color: var(--green-link);
  color: var(--near-black);
  border: none;
  font-weight: 600;
}

.btn-primary:hover:not(:disabled) {
  opacity: 0.9;
  transform: translateY(-1px);
}

.btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
  background-color: var(--charcoal);
  color: var(--mid-gray);
  border: 1px solid var(--border-dark);
}

.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid var(--off-white);
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
