<script setup lang="ts">
import { ref, computed } from 'vue'
import { useAuth } from '@/composables/useAuth'
import type { RegisterReq } from '@/types/api'

const form = ref<RegisterReq>({
  username: '',
  password: '',
})

const confirmPassword = ref('')

const { isLoading, errorMessage, handleRegister } = useAuth()

const usernameErrors = computed(() => {
  const name = form.value.username
  const errors: string[] = []
  if (name.length > 0 && name.length < 3) errors.push('至少 3 个字符')
  if (name.length > 64) errors.push('最多 64 个字符')
  if (name.length > 0 && !/^[a-zA-Z0-9_]+$/.test(name)) errors.push('仅允许字母、数字和下划线')
  return errors
})

const isPasswordValid = computed(() => {
  const pwd = form.value.password
  return (
    pwd.length >= 8 &&
    pwd.length <= 128 &&
    /[a-z]/.test(pwd) &&
    /[A-Z]/.test(pwd) &&
    /[0-9]/.test(pwd)
  )
})

const passwordErrors = computed(() => {
  const pwd = form.value.password
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
    form.value.username.length >= 3 &&
    isPasswordValid.value &&
    form.value.password === confirmPassword.value
  )
})

const submitForm = () => {
  if (!isFormValid.value) {
    return
  }
  handleRegister(form.value)
}
</script>

<template>
  <div class="auth-page">
    <div class="auth-container">
      <div class="auth-header">
        <h1 class="auth-title">注册 OctoTify</h1>
        <p class="auth-subtitle">创建你的消息总线平台账户</p>
      </div>

      <form class="auth-form" @submit.prevent="submitForm">
        <div class="form-group">
          <label class="form-label" for="username">用户名</label>
          <input
            id="username"
            v-model="form.username"
            type="text"
            class="form-input"
            placeholder="3-64 个字符，仅允许字母、数字和下划线"
            autocomplete="username"
            required
            minlength="3"
            maxlength="64"
            pattern="^[a-zA-Z0-9_]+$"
          />
          <span class="form-hint">3-64 个字符，仅允许字母、数字和下划线</span>
          <ul v-if="usernameErrors.length > 0" class="username-requirements">
            <li v-for="err in usernameErrors" :key="err" class="requirement-error">{{ err }}</li>
          </ul>
        </div>

        <div class="form-group">
          <label class="form-label" for="password">密码</label>
          <input
            id="password"
            v-model="form.password"
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
          <label class="form-label" for="confirmPassword">确认密码</label>
          <input
            id="confirmPassword"
            v-model="confirmPassword"
            type="password"
            class="form-input"
            placeholder="再次输入密码"
            autocomplete="new-password"
            required
          />
        </div>

        <div v-if="errorMessage" class="form-error">
          {{ errorMessage }}
        </div>

        <button type="submit" class="btn btn-primary" :disabled="isLoading || !isFormValid">
          <span v-if="isLoading" class="spinner"></span>
          {{ isLoading ? '注册中...' : '注册' }}
        </button>
      </form>

      <div class="auth-footer">
        <p>
          已有账户？
          <router-link :to="{ name: 'Login' }">登录</router-link>
        </p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: var(--dark);
  padding: var(--space-6);
}

.auth-container {
  width: 100%;
  max-width: 400px;
}

.auth-header {
  margin-bottom: var(--space-8);
  text-align: center;
}

.auth-title {
  font-size: 2.25rem;
  font-weight: 400;
  line-height: 1.25;
  color: var(--off-white);
  margin-bottom: var(--space-2);
}

.auth-subtitle {
  font-size: 1rem;
  color: var(--mid-gray);
}

.auth-form {
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

.username-requirements {
  list-style: none;
  padding: 0;
  margin: var(--space-1) 0 0;
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

.form-error {
  padding: 12px 16px;
  background-color: rgba(240, 95, 95, 0.1);
  border: 1px solid var(--error);
  border-radius: var(--radius-sm);
  color: var(--error);
  font-size: 0.875rem;
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

.auth-footer {
  margin-top: var(--space-6);
  text-align: center;
  font-size: 0.875rem;
  color: var(--mid-gray);
}

.auth-footer a {
  color: var(--green-link);
}

.auth-footer a:hover {
  text-decoration: underline;
}
</style>
