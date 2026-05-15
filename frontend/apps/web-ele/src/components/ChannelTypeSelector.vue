<template>
  <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
    <ElCard
      v-for="item in channelTypes"
      :key="item.type"
      shadow="hover"
      class="cursor-pointer transition-all hover:shadow-md"
      :class="{
        'border-2 border-primary border-opacity-80': selectedType === item.type,
        'border border-gray-200': selectedType !== item.type,
      }"
      @click="handleSelect(item)"
    >
      <div class="flex items-start">
        <div class="flex-1 min-w-0">
          <div class="flex items-center">
            <IconifyIcon :icon="getChannelIcon(item.type)" class="mr-2 text-xl" />
            <span class="font-medium text-base truncate">{{ item.name }}</span>
          </div>
          <p class="text-gray-500 text-sm mt-1 truncate">{{ item.description }}</p>
        </div>
      </div>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
import { ElCard } from 'element-plus';
import { IconifyIcon } from '@vben/icons';
import type { ChannelApi } from '#/api/modules/channel';

defineProps<{
  channelTypes: ChannelApi.ChannelTypeMeta[];
  selectedType: string;
}>();

const emit = defineEmits<{
  select: [type: ChannelApi.ChannelTypeMeta];
}>();

function handleSelect(item: ChannelApi.ChannelTypeMeta) {
  emit('select', item);
}

// 渠道类型图标映射
const iconMap: Record<string, string> = {
  feishu: 'mdi:rocket-launch',
  dingtalk: 'mdi:message',
  telegram: 'mdi:send',
  email: 'mdi:email',
  wechat_clawbot: 'mdi:wechat',
  wechat: 'mdi:wechat',
};

function getChannelIcon(type: string): string {
  return iconMap[type] || 'mdi:pipe';
}
</script>
