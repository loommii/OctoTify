<template>
  <div class="channel-list-page">
    <div class="page-header">
      <h2>推送渠道管理</h2>
      <button class="btn-primary" @click="goToCreate">创建渠道</button>
    </div>

    <div v-if="loading" class="loading">加载中...</div>

    <div v-else-if="channels.length === 0" class="empty-state">
      <p>暂无推送渠道</p>
      <button class="btn-primary" @click="goToCreate">创建第一个渠道</button>
    </div>

    <div v-else class="channel-table">
      <table>
        <thead>
          <tr>
            <th>名称</th>
            <th>类型</th>
            <th>状态</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="channel in channels" :key="channel.id">
            <td>{{ channel.name }}</td>
            <td>{{ getTypeName(channel.type ?? '') }}</td>
            <td>
              <span :class="['status-badge', getStatusClass(channel.status ?? 0)]">
                {{ getStatusText(channel.status ?? 0) }}
              </span>
            </td>
            <td>{{ formatTime(channel.created_at ?? 0) }}</td>
            <td class="actions">
              <button class="btn-link" @click="goToDetail(channel.id!)">详情</button>
              <button class="btn-link" @click="goToEdit(channel.id!)">编辑</button>
              <button
                v-if="channel.status === 1"
                class="btn-link btn-warning"
                @click="handleDisable(channel.id!)"
              >
                停用
              </button>
              <button
                v-else-if="channel.status === 2"
                class="btn-link btn-success"
                @click="handleEnable(channel.id!)"
              >
                启用
              </button>
              <button class="btn-link btn-danger" @click="handleDelete(channel.id!)">删除</button>
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
import { listChannels, disableChannel, enableChannel, deleteChannel } from '@/api/channels'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import { useConfirm } from '@/composables/useConfirm'
import { getChannelTypeName, getStatusText, getStatusClass } from '@/lib/constants'
import type { ChannelDTO } from '@/types/api'

const router = useRouter()
const channels = ref<ChannelDTO[]>([])
const loading = ref(true)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

const {
  showDialog,
  dialogOptions,
  actionLoading,
  requestConfirm,
  handleConfirm,
} = useConfirm()

function getTypeName(type: string): string {
  return getChannelTypeName(type)
}

function formatTime(ts: number): string {
  return new Date(ts).toLocaleString('zh-CN')
}

async function loadChannels() {
  loading.value = true
  try {
    const res = await listChannels(page.value, pageSize.value)
    if (res.data) {
      channels.value = res.data.list ?? []
      total.value = res.data.total ?? 0
    }
  } catch (err) {
    console.error('加载渠道列表失败', err)
  } finally {
    loading.value = false
  }
}

function changePage(newPage: number) {
  page.value = newPage
  loadChannels()
}

function goToCreate() {
  router.push({ name: 'ChannelCreate' })
}

function goToDetail(id: number) {
  router.push({ name: 'ChannelDetail', params: { id } })
}

function goToEdit(id: number) {
  router.push({ name: 'ChannelEdit', params: { id } })
}

function handleDisable(id: number) {
  requestConfirm(
    {
      title: '停用推送渠道',
      description: '停用后该渠道将无法接收消息。您可以随时重新启用。',
      confirmText: '停用',
      confirmType: 'warning',
    },
    async () => {
      await disableChannel(id)
      await loadChannels()
    }
  )
}

function handleEnable(id: number) {
  requestConfirm(
    {
      title: '启用推送渠道',
      description: '启用后该渠道立即生效。',
      confirmText: '启用',
      confirmType: 'primary',
    },
    async () => {
      await enableChannel(id)
      await loadChannels()
    }
  )
}

function handleDelete(id: number) {
  requestConfirm(
    {
      title: '删除推送渠道',
      description: '删除后该渠道及其配置将永久丢失。此操作不可恢复。',
      confirmText: '删除',
      confirmType: 'danger',
    },
    async () => {
      await deleteChannel(id)
      await loadChannels()
    }
  )
}

onMounted(() => {
  loadChannels()
})
</script>

<style scoped>
.channel-list-page {
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

.channel-table {
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
</style>
