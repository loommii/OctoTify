<template>
  <div class="source-edit-page">
    <div class="page-header">
      <h2>编辑消息来源</h2>
      <button class="btn-back" @click="goBack">返回</button>
    </div>

    <div v-if="loading" class="loading">加载中...</div>

    <div v-else-if="source" class="form-container">
      <div class="form-group">
        <label>来源名称</label>
        <input v-model="form.name" type="text" placeholder="请输入来源名称" />
      </div>

      <div class="form-group">
        <label>描述</label>
        <textarea v-model="form.description" placeholder="请输入描述（可选）"></textarea>
      </div>

      <div class="form-group">
        <label>关联渠道</label>
        <div v-if="channelsLoading" class="loading-sm">加载渠道列表中...</div>
        <div v-else-if="channels.length === 0" class="empty-hint">
          暂无可用渠道，请先 <router-link :to="{ name: 'ChannelCreate' }">创建渠道</router-link>
        </div>
        <div v-else class="channel-checkboxes">
          <label v-for="ch in channels" :key="ch.id" class="checkbox-label">
            <input type="checkbox" :value="ch.id" v-model="form.channel_ids" />
            {{ ch.name }} ({{ getTypeName(ch.type ?? '') }})
          </label>
        </div>
      </div>

      <div class="form-actions">
        <button class="btn-primary" @click="handleSubmit" :disabled="submitting">
          {{ submitting ? '保存中...' : '保存' }}
        </button>
        <button class="btn-secondary" @click="goBack">取消</button>
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
import { useRouter, useRoute } from 'vue-router'
import { getSourceDetail, updateSource, listChannels } from '@/api/channels'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import { useConfirm } from '@/composables/useConfirm'
import { getChannelTypeName } from '@/lib/constants'
import { useToast } from '@/composables/useToast'
import type { SourceDetailDTO, ChannelDTO } from '@/types/api'

const router = useRouter()
const route = useRoute()
const source = ref<SourceDetailDTO | null>(null)
const loading = ref(true)
const submitting = ref(false)
const channelsLoading = ref(true)
const channels = ref<ChannelDTO[]>([])
const form = ref({
  name: '',
  description: '',
  channel_ids: [] as number[],
})

const { error: showError } = useToast()

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

async function loadSource() {
  loading.value = true
  try {
    const id = Number(route.params.id)
    const res = await getSourceDetail(id)
    if (res.data) {
      source.value = res.data
      form.value.name = res.data.name ?? ''
      form.value.description = res.data.description ?? ''
      form.value.channel_ids = res.data.channels?.map((ch) => ch.id ?? 0).filter((id) => id > 0) || []
    }
  } catch (err) {
    console.error('加载来源详情失败', err)
  } finally {
    loading.value = false
  }
}

async function loadChannels() {
  channelsLoading.value = true
  try {
    const res = await listChannels(1, 100)
    if (res.data) {
      channels.value = (res.data.list ?? []).filter((ch) => ch.status === 1)
    }
  } catch (err) {
    console.error('加载渠道列表失败', err)
  } finally {
    channelsLoading.value = false
  }
}

function handleSubmit() {
  if (!form.value.name) {
    showError('请输入来源名称')
    return
  }
  if (form.value.channel_ids.length === 0) {
    showError('请至少选择一个关联渠道')
    return
  }

  requestConfirm(
    {
      title: '保存来源修改',
      description: `确定要保存来源 "${form.value.name}" 的修改吗？`,
      confirmText: '保存',
      confirmType: 'primary',
    },
    async () => {
      submitting.value = true
      try {
        await updateSource(source.value!.id!, {
          name: form.value.name,
          description: form.value.description,
          channel_ids: form.value.channel_ids,
        })
        router.push({ name: 'SourceDetail', params: { id: source.value!.id } })
      } catch (err) {
        console.error('更新来源失败', err)
        showError('更新来源失败，请重试')
      } finally {
        submitting.value = false
      }
    }
  )
}

function goBack() {
  router.back()
}

onMounted(() => {
  loadSource()
  loadChannels()
})
</script>

<style scoped>
.source-edit-page {
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

.btn-back {
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

.btn-back:hover {
  border-color: var(--mid-border);
}

.loading {
  text-align: center;
  padding: var(--space-12);
  color: var(--mid-gray);
}

.form-container {
  max-width: 800px;
  margin: 0 auto;
  background: var(--dark);
  border: 1px solid var(--border-dark);
  border-radius: var(--radius-lg);
  padding: var(--space-8);
}

.form-group {
  margin-bottom: var(--space-6);
}

.form-group label {
  display: block;
  margin-bottom: var(--space-2);
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--off-white);
}

.form-group input[type="text"],
.form-group textarea {
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

.form-group input[type="text"]:focus,
.form-group textarea:focus {
  outline: none;
  border-color: var(--green-link);
}

.form-group input[type="text"]::placeholder,
.form-group textarea::placeholder {
  color: var(--mid-gray);
}

.form-group textarea {
  min-height: 80px;
  resize: vertical;
}

.loading-sm {
  padding: var(--space-4);
  color: var(--mid-gray);
  font-size: 0.875rem;
}

.empty-hint {
  padding: var(--space-4);
  color: var(--mid-gray);
  font-size: 0.875rem;
}

.empty-hint a {
  color: var(--green-link);
}

.channel-checkboxes {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  max-height: 200px;
  overflow-y: auto;
  padding: var(--space-2);
  border: 1px solid var(--border-dark);
  border-radius: var(--radius-sm);
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  cursor: pointer;
  font-weight: normal;
  margin-bottom: 0;
  color: var(--off-white);
}

.form-actions {
  display: flex;
  gap: var(--space-4);
  margin-top: var(--space-8);
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

.btn-primary:hover:not(:disabled) {
  opacity: 0.8;
}

.btn-primary:disabled {
  opacity: 0.3;
  cursor: not-allowed;
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
