<template>
  <div class="p-4">
    <!-- 搜索栏 -->
    <ElCard shadow="never" class="mb-4">
      <ElForm :model="queryParams" inline>
        <ElFormItem :label="$t('page.source.name')">
          <ElInput
            v-model="queryParams.name"
            :placeholder="$t('page.source.namePlaceholder')"
            clearable
            @keyup.enter="handleSearch"
          />
        </ElFormItem>
        <ElFormItem :label="$t('page.source.status')">
          <ElSelect v-model="queryParams.status" clearable :placeholder="$t('page.source.allStatus')" style="width: 140px">
            <ElOption :label="$t('page.source.enable')" :value="1" />
            <ElOption :label="$t('page.source.disable')" :value="2" />
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
          <span class="font-semibold">{{ $t('page.source.list') }}</span>
          <ElButton type="primary" @click="handleCreate">
            <IconifyIcon icon="mdi:plus" class="mr-1" />
            {{ $t('common.create') }}
          </ElButton>
        </div>
      </template>

      <ElTable v-loading="loading" :data="sourceList" stripe style="width: 100%">
        <ElTableColumn prop="id" label="ID" width="80" align="center" />
        <ElTableColumn :label="$t('page.source.name')" prop="name" min-width="150">
          <template #default="{ row }">
            <a
              class="text-blue-500 hover:text-blue-700 cursor-pointer"
              @click="handleDetail(row.id)"
            >
              {{ row.name }}
            </a>
          </template>
        </ElTableColumn>
        <ElTableColumn :label="$t('page.source.description')" prop="description" min-width="200" show-overflow-tooltip />
        <ElTableColumn :label="$t('page.source.status')" prop="status" width="100" align="center">
          <template #default="{ row }">
            <ElTag :type="getEntityStatusTagType(row.status)" size="small">
              {{ row.status === 1 ? $t('page.source.enable') : $t('page.source.disable') }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn :label="$t('page.source.createTime')" prop="created_at_ts" width="180">
          <template #default="{ row }">
            {{ formatTimestamp(row.created_at_ts) }}
          </template>
        </ElTableColumn>
        <ElTableColumn :label="$t('page.source.actions')" width="320" align="center" fixed="right">
          <template #default="{ row }">
            <ElButton link type="primary" size="small" @click="handleDetail(row.id)">
              {{ $t('page.source.viewDetail') }}
            </ElButton>
            <ElButton link type="primary" size="small" @click="handleEdit(row.id)">
              {{ $t('common.edit') }}
            </ElButton>
            <ElButton
              v-if="row.status === 1"
              link
              type="warning"
              size="small"
              @click="handleDisable(row)"
            >
              {{ $t('page.source.disable') }}
            </ElButton>
            <ElButton
              v-else-if="row.status === 2"
              link
              type="success"
              size="small"
              @click="handleEnable(row)"
            >
              {{ $t('page.source.enable') }}
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

    <!-- 密码二次验证对话框 -->
    <StepUpAuthDialog
      v-model:visible="stepUpVisible"
      :title="stepUpTitle"
      :description="stepUpDescription"
      @confirm="handleStepUpConfirm"
    />
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
} from 'element-plus';
import { IconifyIcon } from '@vben/icons';
import { $t } from '#/locales';
import {
  getSourceListApi,
  deleteSourceApi,
  enableSourceApi,
  disableSourceApi,
  type SourceApi,
} from '#/api/modules/source';
import StepUpAuthDialog from '#/components/StepUpAuthDialog.vue';
import { formatTimestamp } from '#/utils/time';
import { getEntityStatusTagType } from '#/utils/status';

const router = useRouter();

// 查询参数
const queryParams = reactive({
  name: '',
  status: undefined as number | undefined,
  page: 1,
  page_size: 20,
});

// 列表数据
const sourceList = ref<SourceApi.SourceDTO[]>([]);
const total = ref(0);
const loading = ref(false);

// 二次验证状态
const stepUpVisible = ref(false);
const stepUpTitle = ref('');
const stepUpDescription = ref('');
let stepUpResolve: ((password: string) => void) | null = null;

// 加载数据
async function loadData() {
  loading.value = true;
  try {
    const res = await getSourceListApi({
      page: queryParams.page,
      page_size: queryParams.page_size,
    });
    let list = res?.list ?? [];

    // 前端过滤
    if (queryParams.name) {
      list = list.filter((item) => item.name.toLowerCase().includes(queryParams.name.toLowerCase()));
    }
    if (queryParams.status !== undefined) {
      list = list.filter((item) => item.status === queryParams.status);
    }

    sourceList.value = list;
    total.value = res?.total ?? 0;
  } catch {
    ElMessage.error('加载来源列表失败');
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
  queryParams.status = undefined;
  queryParams.page = 1;
  loadData();
}

function handleSizeChange() {
  queryParams.page = 1;
  loadData();
}

// 路由跳转
function handleCreate() {
  router.push('/source/create');
}

function handleDetail(id: number) {
  router.push(`/source/detail/${id}`);
}

function handleEdit(id: number) {
  router.push(`/source/edit/${id}`);
}

// 显示二次验证对话框
async function showStepUp(title: string, description: string): Promise<string | null> {
  return new Promise((resolve) => {
    stepUpTitle.value = title;
    stepUpDescription.value = description;
    stepUpResolve = resolve;
    stepUpVisible.value = true;
  });
}

function handleStepUpConfirm(password: string) {
  stepUpVisible.value = false;
  stepUpResolve?.(password);
  stepUpResolve = null;
}

// 启用来源
async function handleEnable(row: SourceApi.SourceDTO) {
  const password = await showStepUp(
    $t('page.source.confirmEnable'),
    $t('page.source.confirmEnableDesc')
  );
  if (!password) return;

  try {
    await enableSourceApi(row.id, { password });
    ElMessage.success($t('page.source.enableSuccess'));
    loadData();
  } catch {
    ElMessage.error('启用来源失败');
  }
}

// 停用来源
async function handleDisable(row: SourceApi.SourceDTO) {
  const password = await showStepUp(
    $t('page.source.confirmDisable'),
    $t('page.source.confirmDisableDesc')
  );
  if (!password) return;

  try {
    await disableSourceApi(row.id, { password });
    ElMessage.success($t('page.source.disableSuccess'));
    loadData();
  } catch {
    ElMessage.error('停用来源失败');
  }
}

// 删除来源
async function handleDelete(row: SourceApi.SourceDTO) {
  const password = await showStepUp(
    $t('page.source.confirmDelete'),
    $t('page.source.confirmDeleteDesc')
  );
  if (!password) return;

  try {
    await deleteSourceApi(row.id, { password });
    ElMessage.success($t('page.source.deleteSuccess'));
    loadData();
  } catch {
    ElMessage.error('删除来源失败');
  }
}

onMounted(() => {
  loadData();
});
</script>
