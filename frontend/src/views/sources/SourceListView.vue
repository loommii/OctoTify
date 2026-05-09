<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  listSources,
  disableSource,
  enableSource,
  deleteSource,
  getSourceToken,
} from '@/api/channels'
import PasswordConfirmDialog from '@/components/PasswordConfirmDialog.vue'
import { usePasswordConfirm } from '@/composables/usePasswordConfirm'
import { useToast } from '@/composables/useToast'
import { getStatusText, getStatusClass } from '@/lib/constants'
import type { SourceDTO } from '@/types/api'

const router = useRouter()
const { success: showSuccess, error: showError } = useToast()
const sources = ref<SourceDTO[]>([])
const loading = ref(true)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const tokenResult = ref<string | null>(null)

const {
  showDialog,
  dialogOptions,
  actionLoading,
  requestPassword,
  handleConfirm,
} = usePasswordConfirm()

function formatTime(ts: number): string {
  return new Date(ts).toLocaleString('zh-CN')
}

async function loadSources() {
  loading.value = true
  try {
    const res = await listSources(page.value, pageSize.value)
    if (res.data) {
      sources.value = res.data.list ?? []
      total.value = res.data.total ?? 0
    }
  } catch (err) {
    console.error('加载来源列表失败', err)
  } finally {
    loading.value = false
  }
}

function changePage(newPage: number) {
  page.value = newPage
  loadSources()
}

function goToCreate() {
  router.push({ name: 'SourceCreate' })
}

function goToDetail(id: number) {
  router.push({ name: 'SourceDetail', params: { id } })
}

function goToEdit(id: number) {
  router.push({ name: 'SourceEdit', params: { id } })
}

function handleViewToken(id: number) {
  requestPassword(
    {
      title: '查看来源令牌',
      description: '请输入登录密码以验证身份，令牌仅在验证成功后显示。',
      confirmText: '验证并查看',
    },
    async (pwd: string) => {
      const res = await getSourceToken(id, pwd)
      if (res.data?.token) {
        tokenResult.value = res.data.token
      }
    }
  )
}

function handleDisable(id: number) {
  requestPassword(
    {
      title: '停用消息来源',
      description: '停用后使用该 Token 推送的消息会被拒绝。您可以随时重新启用。',
      confirmText: '停用',
    },
    async (pwd: string) => {
      await disableSource(id, pwd)
      await loadSources()
    }
  )
}

function handleEnable(id: number) {
  requestPassword(
    {
      title: '启用消息来源',
      description: '启用后该来源的 Token 立即生效。',
      confirmText: '启用',
    },
    async (pwd: string) => {
      await enableSource(id, pwd)
      await loadSources()
    }
  )
}

function handleDelete(id: number) {
  requestPassword(
    {
      title: '删除消息来源',
      description: '删除后该来源的 Token 立即失效，关联的渠道关系也会被删除。此操作不可恢复。',
      confirmText: '删除',
    },
    async (pwd: string) => {
      await deleteSource(id!, pwd)
      await loadSources()
    }
  )
}

async function copyToken() {
  if (tokenResult.value) {
    try {
      await navigator.clipboard.writeText(tokenResult.value)
      showSuccess('令牌已复制到剪贴板')
    } catch {
      showError('复制失败，请手动复制')
    }
  }
}

onMounted(() => {
  loadSources()
})
</script>

<template>
  <div class="source-list-page">
    <div class="page-header">
      <h2>消息来源管理</h2>
      <button class="btn-primary" @click="goToCreate">创建来源</button>
    </div>

    <div v-if="loading" class="loading">加载中...</div>

    <div v-else-if="sources.length === 0" class="empty-state">
      <p>暂无消息来源</p>
      <button class="btn-primary" @click="goToCreate">创建第一个来源</button>
    </div>

    <div v-else class="source-table">
      <table>
        <thead>
          <tr>
            <th>名称</th>
            <th>描述</th>
            <th>状态</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="source in sources" :key="source.id">
            <td>{{ source.name }}</td>
            <td class="desc-cell">{{ source.description || '-' }}</td>
            <td>
              <span :class="['status-badge', getStatusClass(source.status ?? 0)]">
                {{ getStatusText(source.status ?? 0) }}
              </span>
            </td>
            <td>{{ formatTime(source.created_at ?? 0) }}</td>
            <td class="actions">
              <button class="btn-link" @click="goToDetail(source.id!)">详情</button>
              <button class="btn-link" @click="goToEdit(source.id!)">编辑</button>
              <button class="btn-link" @click="handleViewToken(source.id!)">查看令牌</button>
              <button
                v-if="source.status === 1"
                class="btn-link btn-warning"
                @click="handleDisable(source.id!)"
              >
                停用
              </button>
              <button
                v-else-if="source.status === 2"
                class="btn-link btn-success"
                @click="handleEnable(source.id!)"
              >
                启用
              </button>
              <button class="btn-link btn-danger" @click="handleDelete(source.id!)">删除</button>
            </td>
          </tr>
        </tbody>
      </table>

      <div class="pagination" v-if="total > pageSize">
        <button :disabled="page <= 1" @click="changePage(page - 1)">上一页</button>
        <span>第 {{ page }} 页 / 共 {{ Math.ceil(total / pageSize) }} 页</span>
        <button :disabled="page * pageSize >= total" @click="changePage(page + 1)">下一页</button>
      </div>
    </div>

    <PasswordConfirmDialog
      v-model:visible="showDialog"
      :title="dialogOptions.title"
      :description="dialogOptions.description"
      :confirm-text="dialogOptions.confirmText"
      :loading="actionLoading"
      @confirm="handleConfirm"
    />

    <div v-if="tokenResult" class="modal-overlay" @click.self="tokenResult = null">
      <div class="modal">
        <h3>来源令牌</h3>
        <div class="token-display">
          <code>{{ tokenResult }}</code>
        </div>
        <p class="token-hint">请使用此令牌调用推送接口</p>
        <div class="modal-actions">
          <button class="btn-primary" @click="copyToken">复制令牌</button>
          <button class="btn-secondary" @click="tokenResult = null">关闭</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.source-list-page {
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

.btn-primary:hover {
  opacity: 0.8;
}

.loading {
  text-align: center;
  padding: var(--space-12);
  color: var(--mid-gray);
}

.empty-state {
  text-align: center;
  padding: var(--space-12) var(--space-8);
  color: var(--mid-gray);
}

.source-table {
  background: var(--dark);
  border: 1px solid var(--border-dark);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th, td {
  padding: var(--space-4) var(--space-6);
  text-align: left;
  border-bottom: 1px solid var(--border-dark);
}

th {
  background: var(--near-black);
  font-weight: 500;
  color: var(--off-white);
}

td {
  color: var(--off-white);
}

.desc-cell {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.status-badge {
  padding: var(--space-1) var(--space-2);
  border-radius: var(--radius-sm);
  font-size: 0.75rem;
}

.status-active {
  background: rgba(62, 207, 142, 0.15);
  color: var(--success);
}

.status-disabled {
  background: rgba(245, 158, 11, 0.15);
  color: var(--warning);
}

.status-deleted {
  background: rgba(240, 95, 95, 0.15);
  color: var(--error);
}

.actions {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.btn-link {
  background: none;
  border: none;
  color: var(--green-link);
  cursor: pointer;
  font-size: 0.8125rem;
  padding: var(--space-1) var(--space-2);
  transition: opacity var(--transition-fast);
}

.btn-link:hover {
  opacity: 0.8;
}

.btn-warning {
  color: var(--warning);
}

.btn-success {
  color: var(--success);
}

.btn-danger {
  color: var(--error);
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: var(--space-6);
  padding: var(--space-6);
  color: var(--off-white);
}

.pagination button {
  padding: var(--space-2) var(--space-4);
  border: 1px solid var(--border-dark);
  background: var(--near-black);
  color: var(--off-white);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: border-color var(--transition-fast);
}

.pagination button:hover:not(:disabled) {
  border-color: var(--green-link);
}

.pagination button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

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

.modal h3 {
  margin: 0 0 var(--space-6);
  font-size: 1.125rem;
  font-weight: 500;
  color: var(--off-white);
}

.token-display {
  background: var(--near-black);
  border: 1px solid var(--border-dark);
  padding: var(--space-4);
  border-radius: var(--radius-sm);
  margin-bottom: var(--space-4);
  word-break: break-all;
}

.token-display code {
  font-size: 0.8125rem;
  color: var(--off-white);
  font-family: var(--font-mono);
}

.token-hint {
  font-size: 0.8125rem;
  color: var(--mid-gray);
  margin: 0 0 var(--space-6);
}

.modal-actions {
  display: flex;
  gap: var(--space-4);
  justify-content: flex-end;
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
