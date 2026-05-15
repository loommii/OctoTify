<template>
  <div class="p-4">
    <ElCard v-loading="loading" shadow="never">
      <template #header>
        <span class="font-semibold">{{ $t('page.settings.profile') }}</span>
      </template>

      <ElDescriptions v-if="userProfile" :column="1" border>
        <ElDescriptionsItem :label="$t('page.settings.username')">
          {{ userProfile.username }}
        </ElDescriptionsItem>
        <ElDescriptionsItem :label="$t('page.settings.userId')">
          {{ userProfile.userId }}
        </ElDescriptionsItem>
        <ElDescriptionsItem :label="$t('page.settings.registerTime')">
          {{ formatTimestamp(userProfile.createdAtTs) }}
        </ElDescriptionsItem>
      </ElDescriptions>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { ElCard, ElDescriptions, ElDescriptionsItem } from 'element-plus';
import { $t } from '#/locales';
import { getUserInfoApi } from '#/api/core/user';
import type { UserInfo } from '#/api/core/user';
import { formatTimestamp } from '#/utils/time';

const loading = ref(false);
const userProfile = ref<UserInfo | null>(null);

async function loadProfile() {
  loading.value = true;
  try {
    userProfile.value = await getUserInfoApi();
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  loadProfile();
});
</script>
