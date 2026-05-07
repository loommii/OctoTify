<script setup lang="ts">
import { ref } from 'vue'
import { useAuth } from '@/composables/useAuth'
import type { LoginReq } from '@/types/api'

const form = ref<LoginReq>({
  username: '',
  password: '',
})

const { isLoading, errorMessage, handleLogin } = useAuth()
</script>

<template>
  <div class="auth-page">
    <div class="auth-container">
      <div class="auth-header">
        <h1 class="auth-title">登录 OctoTify</h1>
        <p class="auth-subtitle">消息总线平台</p>
      </div>

      <form class="auth-form" @submit.prevent="() => handleLogin(form)">
        <div class="form-group">
          <label class="form-label" for="username">用户名</label>
          <input
            id="username"
            v-model="form.username"
            type="text"
            class="form-input"
            placeholder="输入用户名"
            autocomplete="username"
            required
          />
        </div>

        <div class="form-group">
          <label class="form-label" for="password">密码</label>
          <input
            id="password"
            v-model="form.password"
            type="password"
            class="form-input"
            placeholder="输入密码"
            autocomplete="current-password"
            required
          />
        </div>

        <div v-if="errorMessage" class="form-error">
          {{ errorMessage }}
        </div>

        <button type="submit" class="btn btn-primary" :disabled="isLoading">
          <span v-if="isLoading" class="spinner"></span>
          {{ isLoading ? '登录中...' : '登录' }}
        </button>
      </form>

      <div class="auth-footer">
        <p>
          还没有账户？
          <router-link :to="{ name: 'Register' }">注册</router-link>
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
