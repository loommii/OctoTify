<template>
  <div class="p-4">
    <!-- 搜索栏 -->
    <ElCard shadow="never" class="mb-4">
      <ElForm :model="queryParams" inline>
        <ElFormItem :label="$t('page.channel.name')">
          <ElInput
            v-model="queryParams.name"
            :placeholder="$t('page.channel.namePlaceholder')"
            clearable
            @keyup.enter="handleSearch"
          />
        </ElFormItem>
        <ElFormItem :label="$t('page.channel.type')">
          <ElSelect v-model="queryParams.type" clearable :placeholder="$t('page.channel.allTypes')" style="width: 180px">
            <ElOption
              v-for="t in channelTypeOptions"
              :key="t.type"
              :label="t.name"
              :value="t.type"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem :label="$t('page.channel.status')">
          <ElSelect v-model="queryParams.status" clearable :placeholder="$t('page.channel.allStatus')" style="width: 140px">
            <ElOption :label="$t('page.channel.enable')" :value="1" />
            <ElOption :label="$t('page.channel.disable')" :value="2" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem>
          <ElButton type="primary" @click="handleSearch">
            <IconifyIcon icon="mdi:magnify" class="mr-1" />
            {{ $t('common.search') }}
          </ElButton>
          <ElButton @click="handleReset">
            {{ $t('common.reset') }}
          </ElButton>
        </ElFormItem>
      </ElForm>
    </ElCard>

    <!-- 操作按钮 + 表格 -->
    <ElCard shadow="never">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="font-semibold">{{ $t('page.channel.list') }}</span>
          <ElButton type="primary" @click="handleCreate">
            <IconifyIcon icon="mdi:plus" class="mr-1" />
            {{ $t('common.create') }}
          </ElButton>
        </div>
      </template>

      <ElTable v-loading="loading" :data="channelList" stripe style="width: 100%">
        <ElTableColumn prop="id" label="ID" width="80" align="center" />
        <ElTableColumn :label="$t('page.channel.name')" prop="name" min-width="150">
          <template #default="{ row }">
            <a
              class="text-blue-500 hover:text-blue-700 cursor-pointer"
              @click="handleDetail(row.id)"
            >
              {{ row.name }}
            </a>
          </template>
        </ElTableColumn>
        <ElTableColumn :label="$t('page.channel.type')" prop="type" width="160" align="center">
          <template #default="{ row }">
            <ElTag size="small">{{ getChannelTypeLabel(row.type) }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn :label="$t('page.channel.status')" prop="status" width="100" align="center">
          <template #default="{ row }">
            <ElTag :type="getStatusTagType(row.status)" size="small">
              {{ getStatusLabel(row.status) }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn :label="$t('page.channel.createTime')" prop="created_at_ts" width="180">
          <template #default="{ row }">
            {{ formatTimestamp(row.created_at_ts) }}
          </template>
        </ElTableColumn>
        <ElTableColumn :label="$t('page.channel.actions')" width="360" align="center" fixed="right">
          <template #default="{ row }">
            <ElButton link type="primary" size="small" @click="handleDetail(row.id)">
              {{ $t('page.channel.viewDetail') }}
            </ElButton>
            <ElButton link type="primary" size="small" @click="handleEdit(row.id)">
              {{ $t('common.edit') }}
            </ElButton>
            <ElButton link type="info" size="small" @click="handleTest(row)">
              {{ $t('page.channel.testConnection') }}
            </ElButton>
            <ElButton
              v-if="row.status === 1"
              link
              type="warning"
              size="small"
              @click="handleDisable(row)"
            >
              {{ $t('page.channel.disable') }}
            </ElButton>
            <ElButton
              v-else-if="row.status === 2"
              link
              type="success"
              size="small"
              @click="handleEnable(row)"
            >
              {{ $t('page.channel.enable') }}
            </ElButton>
            <ElButton link type="danger" size="small" @click="handleDelete(row)">
              {{ $t('common.delete') }}
            </ElButton>
          </template>
        </ElTableColumn>
      </ElTable>

      <!-- 分页 -->
      <div class="flex justify-end mt-4">
        <ElPagination
          v-model:current-page="queryParams.page"
          v-model:page-size="queryParams.page_size"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          background
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="loadData"
          @size-change="handleSizeChange"
        />
      </div>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import {
  ElCard,
  ElForm,
  ElFormItem,
  ElInput,
  ElSelect,
  ElOption,
  ElButton,
  ElTable,
  ElTableColumn,
  ElTag,
  ElPagination,
  ElMessage,
  ElMessageBox,
} from 'element-plus';
import { IconifyIcon } from '@vben/icons';
import { $t } from '#/locales';
import {
  getChannelListApi,
  getChannelTypesApi,
  deleteChannelApi,
  enableChannelApi,
  disableChannelApi,
  testChannelApi,
  type ChannelApi,
} from '#/api/modules/channel';
import { formatTimestamp } from '#/utils/time';

const router = useRouter();

// 渠道类型选项
const channelTypeOptions = ref<ChannelApi.ChannelTypeMeta[]>([]);

// 查询参数
const queryParams = reactive({
  name: '',
  type: undefined as string | undefined,
  status: undefined as number | undefined,
  page: 1,
  page_size: 20,
});

// 列表数据
const channelList = ref<ChannelApi.ChannelDTO[]>([]);
const total = ref(0);
const loading = ref(false);

// 加载数据
async function loadData() {
  loading.value = true;
  try {
    const res = await getChannelListApi({
      page: queryParams.page,
      page_size: queryParams.page_size,
    });
    let list = res?.list ?? [];

    // 前端过滤
    if (queryParams.name) {
      list = list.filter((item) => item.name.toLowerCase().includes(queryParams.name.toLowerCase()));
    }
    if (queryParams.type) {
      list = list.filter((item) => item.type === queryParams.type);
    }
    if (queryParams.status !== undefined) {
      list = list.filter((item) => item.status === queryParams.status);
    }

    channelList.value = list;
    total.value = res?.total ?? 0;
  } catch {
    ElMessage.error($t('page.channel.loadFailed'));
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  queryParams.page = 1;
  loadData();
}

function handleReset() {
  queryParams.name = '';
  queryParams.type = undefined;
  queryParams.status = undefined;
  queryParams.page = 1;
  loadData();
}

function handleSizeChange() {
  queryParams.page = 1;
  loadData();
}

// 状态映射
const statusMap: Record<number, { label: string; type: 'success' | 'danger' | 'info' }> = {
  1: { label: $t('page.channel.enable'), type: 'success' },
  2: { label: $t('page.channel.disable'), type: 'danger' },
};

function getStatusLabel(status: number): string {
  return statusMap[status]?.label ?? String(status);
}

function getStatusTagType(status: number): 'success' | 'danger' | 'info' {
  return statusMap[status]?.type ?? 'info';
}

// 渠道类型标签映射
const channelTypeLabelMap: Record<string, string> = {};

function getChannelTypeLabel(type: string): string {
  return channelTypeLabelMap[type] || type;
}

// 路由跳转
function handleCreate() {
  router.push('/channel/create');
}

function handleDetail(id: number) {
  router.push(`/channel/detail/${id}`);
}

function handleEdit(id: number) {
  router.push(`/channel/edit/${id}`);
}

// 测试连接
async function handleTest(row: ChannelApi.ChannelDTO) {
  try {
    await testChannelApi(row.id);
    ElMessage.success($t('page.channel.testSuccess'));
  } catch {
    ElMessage.error($t('page.channel.testFailed'));
  }
}

// 启用渠道
async function handleEnable(row: ChannelApi.ChannelDTO) {
  try {
    await ElMessageBox.confirm(
      $t('page.channel.confirmEnableDesc'),
      $t('page.channel.confirmEnable'),
      { type: 'warning' }
    );
    await enableChannelApi(row.id);
    ElMessage.success($t('page.channel.enableSuccess'));
    loadData();
  } catch {
    // 用户取消或请求失败
  }
}

// 停用渠道
async function handleDisable(row: ChannelApi.ChannelDTO) {
  try {
    await ElMessageBox.confirm(
      $t('page.channel.confirmDisableDesc'),
      $t('page.channel.confirmDisable'),
      { type: 'warning' }
    );
    await disableChannelApi(row.id);
    ElMessage.success($t('page.channel.disableSuccess'));
    loadData();
  } catch {
    // 用户取消或请求失败
  }
}

// 删除渠道
async function handleDelete(row: ChannelApi.ChannelDTO) {
  try {
    await ElMessageBox.confirm(
      $t('page.channel.confirmDeleteDesc'),
      $t('page.channel.confirmDelete'),
      { type: 'error' }
    );
    await deleteChannelApi(row.id);
    ElMessage.success($t('page.channel.deleteSuccess'));
    loadData();
  } catch {
    // 用户取消或请求失败
  }
}

onMounted(async () => {
  // 加载渠道类型元数据
  try {
    const types = await getChannelTypesApi();
    channelTypeOptions.value = types;
    // 构建类型标签映射
    types.forEach((t) => {
      channelTypeLabelMap[t.type] = t.name;
    });
  } catch {
    // 渠道类型加载失败不影响列表展示
  }
  loadData();
});
</script>
