<template>
  <ElCard shadow="never" class="mb-4">
    <template #header>
      <div class="flex items-center">
        <IconifyIcon icon="mdi:wechat" class="mr-2 text-green-500 text-xl" />
        <span class="font-semibold">{{ $t('page.channel.bind') }}</span>
      </div>
    </template>

    <!-- 未开始绑定 -->
    <div v-if="state === 'idle'" class="flex-col-center py-6">
      <ElButton type="primary" size="large" @click="handleStartBind">
        <IconifyIcon icon="mdi:qrcode-scan" class="mr-1" />
        {{ $t('page.channel.bind') }}
      </ElButton>
      <p class="text-gray-400 mt-2 text-sm">{{ $t('page.channel.bindDesc') }}</p>
    </div>

    <!-- 绑定中（等待扫码） -->
    <div v-else-if="state === 'binding'" class="flex-col-center py-6">
      <img v-if="qrCodeUrl" :src="qrCodeImage" alt="QR Code" class="w-52 h-52 border rounded-lg" />
      <p v-else class="w-52 h-52 border rounded-lg flex items-center justify-center text-gray-400">加载中...</p>
      <p class="mt-4 text-gray-600">{{ $t('page.channel.waitingScan') }}</p>
      <p class="text-gray-400 text-sm mt-1">{{ $t('page.channel.scanTip') }}</p>
    </div>

    <!-- 待激活（已扫码，等待发送消息） -->
    <div v-else-if="state === 'pending_activation'" class="flex-col-center py-6">
      <IconifyIcon icon="mdi:message-processing" class="text-blue-500 text-5xl" />
      <p class="mt-3 text-lg font-medium text-gray-700">{{ $t('page.channel.pendingActivation') }}</p>
      <p class="text-gray-400 text-sm mt-1">{{ $t('page.channel.activationTip') }}</p>
    </div>

    <!-- 扫码确认 -->
    <div v-else-if="state === 'confirmed'" class="flex-col-center py-6">
      <IconifyIcon icon="mdi:check-circle" class="text-green-500 text-5xl" />
      <p class="mt-3 text-lg font-medium text-gray-700">{{ $t('page.channel.scanConfirmed') }}</p>
      <p class="text-gray-400 text-sm mt-1">{{ $t('page.channel.bindSuccess') }}</p>
    </div>

    <!-- 二维码过期 -->
    <div v-else-if="state === 'expired'" class="flex-col-center py-6">
      <IconifyIcon icon="mdi:alert-circle" class="text-red-500 text-5xl" />
      <p class="mt-3 text-lg font-medium text-gray-700">{{ $t('page.channel.expired') }}</p>
      <ElButton type="primary" class="mt-4" @click="handleRefresh">
        <IconifyIcon icon="mdi:refresh" class="mr-1" />
        {{ $t('page.channel.refreshing') }}
      </ElButton>
    </div>
  </ElCard>
</template>

<script setup lang="ts">
import { watch, onBeforeUnmount } from 'vue';
import { ElCard, ElButton } from 'element-plus';
import { IconifyIcon } from '@vben/icons';
import { $t } from '#/locales';
import { useWechatBind } from '#/composables/useWechatBind';
import { useQRCode } from '@vueuse/integrations/useQRCode';

// 定义组件事件
const emit = defineEmits<{
  bound: [credential: any];
}>();

// 获取微信绑定状态和方法
const { state, qrCodeUrl, credential, startBind, stopPolling, stopActivationPolling } = useWechatBind();

// 使用 useQRCode 将二维码 URL 转换为二维码图片
const qrCodeImage = useQRCode(qrCodeUrl, {
  errorCorrectionLevel: 'M', // 中等容错级别
  margin: 2, // 二维码边距
});

// 开始绑定
function handleStartBind() {
  startBind();
}

// 刷新二维码
function handleRefresh() {
  startBind();
}

// 监听凭证变化，通知父组件
watch(credential, (newVal) => {
  if (newVal && state.value === 'pending_activation') {
    emit('bound', newVal);
  }
});

// 组件卸载时停止轮询
onBeforeUnmount(() => {
  stopPolling();
  stopActivationPolling();
});
</script>
