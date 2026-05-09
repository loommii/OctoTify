<template>
  <div class="message-list-page">
    <div class="page-header">
      <h2>消息记录</h2>
      <button class="btn-filter" @click="showFilter = !showFilter">
        {{ showFilter ? '隐藏筛选' : '显示筛选' }}
      </button>
    </div>

    <div v-if="showFilter" class="filter-panel">
      <div class="filter-row">
        <div class="filter-group">
          <label>来源</label>
          <select v-model="filter.source_id">
            <option value="">全部</option>
            <option v-for="s in sources" :key="s.id" :value="s.id">{{ s.name }}</option>
          </select>
        </div>
        <div class="filter-group">
          <label>状态</label>
          <select v-model="filter.status">
            <option value="">全部</option>
            <option value="200">成功</option>
            <option value="300">失败</option>
            <option value="100">待推送</option>
          </select>
        </div>
        <div class="filter-group">
          <label>开始时间</label>
          <input v-model="filter.start_date" type="datetime-local" />
        </div>
        <div class="filter-group">
          <label>结束时间</label>
          <input v-model="filter.end_date" type="datetime-local" />
        </div>
        <div class="filter-actions">
          <button class="btn-primary" @click="applyFilter">应用筛选</button>
          <button class="btn-secondary" @click="resetFilter">重置</button>
        </div>
      </div>
    </div>

    <div v-if="loading" class="loading">加载中...</div>

    <div v-else-if="messages.length === 0" class="empty-state">
      <p>暂无消息记录</p>
    </div>

    <div v-else class="message-table">
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>来源</th>
            <th>标题</th>
            <th>状态</th>
            <th>推送时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="msg in messages" :key="msg.id">
            <td>{{ msg.id }}</td>
            <td>{{ msg.source_name || '-' }}</td>
            <td class="title-cell">{{ msg.title || '' }}</td>
            <td>
              <span :class="['status-badge', getMessageStatusClass(msg.status ?? 0)]">
                {{ getMessageStatusText(msg.status ?? 0) }}
              </span>
            </td>
            <td>{{ formatTime(msg.created_at ?? 0) }}</td>
            <td class="actions">
              <button class="btn-link" @click="goToDetail(msg.id!)">详情</button>
              <button class="btn-link btn-danger" @click="handleDelete(msg.id!)">删除</button>
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
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { listMessages, filterMessages, deleteMessage, listSources } from '@/api/channels'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import { useConfirm } from '@/composables/useConfirm'
import { getMessageStatusText, getMessageStatusClass } from '@/lib/constants'
import type { MessageDTO, SourceDTO } from '@/types/api'

const router = useRouter()
const messages = ref<MessageDTO[]>([])
const sources = ref<SourceDTO[]>([])
const loading = ref(true)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const showFilter = ref(false)
const filter = ref({
  source_id: '',
  status: '',
  start_date: '',
  end_date: '',
})

const {
  showDialog,
  dialogOptions,
  actionLoading,
  requestConfirm,
  handleConfirm,
} = useConfirm()

function formatTime(ts: number): string {
  return new Date(ts).toLocaleString('zh-CN')
}

async function loadMessages() {
  loading.value = true
  try {
    const res = await listMessages(page.value, pageSize.value)
    if (res.data) {
      messages.value = res.data.list ?? []
      total.value = res.data.total ?? 0
    }
  } catch (err) {
    console.error('加载消息列表失败', err)
  } finally {
    loading.value = false
  }
}

async function loadSources() {
  try {
    const res = await listSources(1, 100)
    if (res.data) {
      sources.value = res.data.list ?? []
    }
  } catch (err) {
    console.error('加载来源列表失败', err)
  }
}

function changePage(newPage: number) {
  page.value = newPage
  loadMessages()
}

async function applyFilter() {
  loading.value = true
  try {
    const res = await filterMessages({
      source_id: filter.value.source_id ? Number(filter.value.source_id) : undefined,
      status: filter.value.status ? Number(filter.value.status) : undefined,
      start_date: filter.value.start_date ? new Date(filter.value.start_date).getTime() : undefined,
      end_date: filter.value.end_date ? new Date(filter.value.end_date).getTime() : undefined,
    })
    if (res.data) {
      messages.value = res.data.list ?? []
      total.value = res.data.total ?? 0
      page.value = 1
    }
  } catch (err) {
    console.error('筛选消息失败', err)
  } finally {
    loading.value = false
  }
}

function resetFilter() {
  filter.value = {
    source_id: '',
    status: '',
    start_date: '',
    end_date: '',
  }
  page.value = 1
  loadMessages()
}

function goToDetail(id: number) {
  router.push({ name: 'MessageDetail', params: { id } })
}

function handleDelete(id: number) {
  requestConfirm(
    {
      title: '删除消息记录',
      description: '确定要删除该消息记录吗？此操作不可恢复。',
      confirmText: '删除',
      confirmType: 'danger',
    },
    async () => {
      try {
        await deleteMessage(id)
        await loadMessages()
      } catch (err) {
        console.error('删除消息失败', err)
      }
    }
  )
}

onMounted(() => {
  loadMessages()
  loadSources()
})
</script>

<style scoped>
.message-list-page {
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

.btn-filter {
  padding: var(--space-2) var(--space-4);
  background: var(--near-black);
  border: 1px solid var(--border-dark);
  border-radius: var(--radius-sm);
  color: var(--off-white);
  cursor: pointer;
  font-size: 0.875rem;
  transition: border-color var(--transition-fast);
}

.btn-filter:hover {
  border-color: var(--green-link);
}

.filter-panel {
  background: var(--dark);
  border: 1px solid var(--border-dark);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  margin-bottom: var(--space-6);
}

.filter-row {
  display: flex;
  gap: var(--space-6);
  align-items: flex-end;
  flex-wrap: wrap;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.filter-group label {
  font-size: 0.8125rem;
  color: var(--mid-gray);
}

.filter-group select,
.filter-group input {
  padding: var(--space-2) var(--space-4);
  background: var(--near-black);
  border: 1px solid var(--border-dark);
  border-radius: var(--radius-sm);
  font-size: 0.875rem;
  color: var(--off-white);
  transition: border-color var(--transition-fast);
}

.filter-group select:focus,
.filter-group input:focus {
  outline: none;
  border-color: var(--green-link);
}

.filter-actions {
  display: flex;
  gap: var(--space-2);
}

.btn-primary {
  padding: var(--space-2) var(--space-4);
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

.btn-secondary {
  padding: var(--space-2) var(--space-4);
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

.message-table {
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

.title-cell {
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.status-badge {
  padding: var(--space-1) var(--space-2);
  border-radius: var(--radius-sm);
  font-size: 0.75rem;
}

.status-success {
  background: rgba(62, 207, 142, 0.15);
  color: var(--success);
}

.status-failed {
  background: rgba(240, 95, 95, 0.15);
  color: var(--error);
}

.status-partial {
  background: rgba(245, 158, 11, 0.15);
  color: var(--warning);
}

.actions {
  display: flex;
  gap: var(--space-2);
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
</style>
