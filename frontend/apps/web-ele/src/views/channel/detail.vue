<template>
  <div class="p-4">
    <ElCard v-loading="loading" shadow="never">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="font-semibold">{{ $t('page.channel.detail') }}</span>
          <ElButton type="primary" @click="handleBack">
            <IconifyIcon icon="mdi:arrow-left" class="mr-1" />
            {{ $t('common.back') }}
          </ElButton>
        </div>
      </template>

      <template v-if="channelDetail">
        <!-- 基本信息 -->
        <ElDescriptions :column="2" border class="mb-6">
          <ElDescriptionsItem :label="$t('page.channel.name')">
            {{ channelDetail.name }}
          </ElDescriptionsItem>
          <ElDescriptionsItem :label="$t('page.channel.type')">
            <ElTag size="small">{{ getChannelTypeLabel(channelDetail.type) }}</ElTag>
          </ElDescriptionsItem>
          <ElDescriptionsItem :label="$t('page.channel.status')">
            <ElTag :type="channelDetail.status === 1 ? 'success' : 'danger'" size="small">
              {{ getStatusLabel(channelDetail.status) }}
            </ElTag>
          </ElDescriptionsItem>
          <ElDescriptionsItem :label="$t('page.channel.createTime')">
            {{ formatTimestamp(channelDetail.created_at_ts) }}
          </ElDescriptionsItem>
          <ElDescriptionsItem :label="$t('page.channel.updateTime')">
            {{ formatTimestamp(channelDetail.updated_at_ts) }}
          </ElDescriptionsItem>
          <ElDescriptionsItem :label="$t('page.channel.lastUsedTime')">
            {{ channelDetail.last_used_at_ts ? formatTimestamp(channelDetail.last_used_at_ts) : $t('page.channel.neverUsed') }}
          </ElDescriptionsItem>
        </ElDescriptions>

        <!-- 渠道配置 -->
        <ElCard :title="$t('page.channel.config')" shadow="never">
          <ElDescriptions :column="1" border>
            <ElDescriptionsItem
              v-for="(value, key) in channelDetail.config"
              :key="key"
              :label="getConfigLabel(key)"
            >
              <span v-if="isSensitiveField(key)">****</span>
              <span v-else>{{ value }}</span>
            </ElDescriptionsItem>
          </ElDescriptions>
        </ElCard>

        <!-- 操作按钮 -->
        <div class="mt-6 flex gap-2">
          <ElButton type="primary" @click="handleTest">
            <IconifyIcon icon="mdi:connection" class="mr-1" />
            {{ $t('page.channel.testConnection') }}
          </ElButton>
          <ElButton @click="handleEdit">
            <IconifyIcon icon="mdi:pencil" class="mr-1" />
            {{ $t('common.edit') }}
          </ElButton>
        </div>
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
  getChannelDetailApi,
  testChannelApi,
  getChannelTypesApi,
  type ChannelApi,
} from '#/api/modules/channel';
import { formatTimestamp } from '#/utils/time';

const route = useRoute();
const router = useRouter();

const loading = ref(false);
const channelDetail = ref<ChannelApi.ChannelDTO | null>(null);
const channelTypeMeta = ref<Record<string, string>>({});

// 加载详情
async function loadDetail() {
  const id = Number(route.params.id);
  if (!id) {
    ElMessage.error($t('page.channel.invalidId'));
    return;
  }

  loading.value = true;
  try {
    const res = await getChannelDetailApi(id);
    channelDetail.value = res;
  } catch {
    ElMessage.error($t('page.channel.loadDetailFailed'));
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

function getChannelTypeLabel(type: string): string {
  return channelTypeMeta.value[type] || type;
}

// 配置字段标签映射
const configLabelMap: Record<string, string> = {
  webhook_url: 'Webhook 地址',
  secret: '签名密钥',
  bot_token: 'Bot Token',
  bot_token_ciphertext: 'Bot Token（加密）',
  bot_token_nonce: 'Token Nonce',
  ilink_bot_id: 'Bot ID',
  ilink_user_id: '用户 ID',
  chat_id: 'Chat ID',
  proxy: 'HTTP 代理',
  smtp_host: 'SMTP 服务器',
  smtp_port: 'SMTP 端口',
  username: '用户名',
  password: '密码/授权码',
  to: '收件人',
  cc: '抄送人',
  from_name: '发件人名称',
};

function getConfigLabel(key: string): string {
  return configLabelMap[key] || key;
}

// 敏感字段（不显示明文）
const sensitiveFields = ['secret', 'password', 'bot_token', 'bot_token_ciphertext', 'bot_token_nonce'];

function isSensitiveField(key: string): boolean {
  return sensitiveFields.includes(key);
}

// 状态映射
function getStatusLabel(status: number): string {
  return status === 1 ? $t('page.channel.enable') : $t('page.channel.disable');
}

// 测试连接
async function handleTest() {
  const id = Number(route.params.id);
  try {
    await testChannelApi(id);
    ElMessage.success($t('page.channel.testSuccess'));
  } catch {
    ElMessage.error($t('page.channel.testFailed'));
  }
}

// 编辑
function handleEdit() {
  const id = Number(route.params.id);
  router.push(`/channel/edit/${id}`);
}

function handleBack() {
  router.push('/channel/list');
}

onMounted(() => {
  Promise.all([loadDetail(), loadTypeMeta()]);
});
</script>
