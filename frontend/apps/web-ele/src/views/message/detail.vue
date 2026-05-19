<template>
  <div class="p-4">
    <ElCard v-loading="loading" shadow="never">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="font-semibold">{{ $t('page.message.detail') }}</span>
          <ElButton type="primary" @click="handleBack">
            <IconifyIcon icon="mdi:arrow-left" class="mr-1" />
            {{ $t('common.back') }}
          </ElButton>
        </div>
      </template>

      <template v-if="messageDetail">
        <!-- 消息基本信息 -->
        <ElDescriptions :column="2" border class="mb-6">
          <ElDescriptionsItem :label="$t('page.message.messageTitle')" :span="2">
            {{ messageDetail.title }}
          </ElDescriptionsItem>
          <ElDescriptionsItem :label="$t('page.message.pushStatus')">
            <ElTag :type="getStatusTagType(messageDetail.status)" size="small">
              {{ getStatusLabel(messageDetail.status) }}
            </ElTag>
          </ElDescriptionsItem>
          <ElDescriptionsItem :label="$t('page.message.createTime')">
            {{ formatTimestamp(messageDetail.created_at_ts) }}
          </ElDescriptionsItem>
          <ElDescriptionsItem :label="$t('page.message.updateTime')">
            {{ formatTimestamp(messageDetail.updated_at_ts) }}
          </ElDescriptionsItem>
        </ElDescriptions>

        <!-- 消息内容 -->
        <ElCard :title="$t('page.message.messageContent')" shadow="never" class="mb-6">
          <div class="whitespace-pre-wrap text-gray-700 leading-relaxed">
            {{ messageDetail.content || '--' }}
          </div>
        </ElCard>

        <!-- 来源信息 -->
        <ElCard :title="$t('page.message.sourceInfo')" shadow="never" class="mb-6">
          <ElDescriptions :column="2" border>
            <ElDescriptionsItem :label="$t('page.message.source')">
              {{ messageDetail.source_name || '--' }}
            </ElDescriptionsItem>
            <ElDescriptionsItem label="ID">
              {{ messageDetail.source_id }}
            </ElDescriptionsItem>
          </ElDescriptions>
        </ElCard>

        <!-- 渠道信息 -->
        <ElCard :title="$t('page.message.channelInfo')" shadow="never">
          <ElDescriptions :column="2" border>
            <ElDescriptionsItem :label="$t('page.message.channel')">
              {{ messageDetail.channel_name || '--' }}
            </ElDescriptionsItem>
            <ElDescriptionsItem label="ID">
              {{ messageDetail.channel_id }}
            </ElDescriptionsItem>
            <ElDescriptionsItem :label="$t('page.channel.type')" v-if="messageDetail.channel_type">
              <ElTag size="small">{{ getChannelTypeLabel(messageDetail.channel_type) }}</ElTag>
            </ElDescriptionsItem>
          </ElDescriptions>
        </ElCard>
      </template>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import {
  ElCard,
  ElButton,
  ElDescriptions,
  ElDescriptionsItem,
  ElTag,
  ElMessage,
} from 'element-plus';
import { IconifyIcon } from '@vben/icons';
import { $t } from '#/locales';
import {
  getMessageDetailApi,
  type MessageApi,
} from '#/api/modules/message';
import { getChannelTypesApi } from '#/api/modules/channel';
import { formatTimestamp } from '#/utils/time';

const route = useRoute();
const router = useRouter();

const loading = ref(false);
const messageDetail = ref<MessageApi.MessageDTO | null>(null);
const channelTypeMeta = ref<Record<string, string>>({});

// 加载详情
async function loadDetail() {
  const id = Number(route.params.id);
  if (!id) {
    ElMessage.error($t('page.message.invalidId'));
    return;
  }

  loading.value = true;
  try {
    const res = await getMessageDetailApi(id);
    messageDetail.value = res;
  } catch {
    ElMessage.error($t('page.message.loadDetailFailed'));
  } finally {
    loading.value = false;
  }
}

// 加载渠道类型元数据
async function loadTypeMeta() {
  try {
    const types = await getChannelTypesApi();
    types.forEach((t) => {
      channelTypeMeta.value[t.type] = t.name;
    });
  } catch {
    // 不影响主流程
  }
}

// 状态映射
const statusMap: Record<number, { label: string; type: 'info' | 'success' | 'danger' }> = {
  100: { label: $t('page.message.statusPending'), type: 'info' },
  200: { label: $t('page.message.statusSuccess'), type: 'success' },
  300: { label: $t('page.message.statusFailed'), type: 'danger' },
};

function getStatusLabel(status: number): string {
  return statusMap[status]?.label ?? String(status);
}

function getStatusTagType(status: number): 'info' | 'success' | 'danger' {
  return statusMap[status]?.type ?? 'info';
}

// 渠道类型标签
function getChannelTypeLabel(type: string): string {
  return channelTypeMeta.value[type] || type;
}

function handleBack() {
  router.push('/message/list');
}

onMounted(() => {
  Promise.all([loadDetail(), loadTypeMeta()]);
});
</script>
