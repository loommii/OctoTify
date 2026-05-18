<template>
  <div class="p-4">
    <ElCard shadow="never">
      <template #header>
        <div class="flex items-center justify-between">
          <span class="font-semibold">{{ $t('page.channel.create') }}</span>
          <ElButton @click="handleBack">
            <IconifyIcon icon="mdi:arrow-left" class="mr-1" />
            {{ $t('common.back') }}
          </ElButton>
        </div>
      </template>

      <!-- 步骤 1：选择渠道类型 -->
      <div v-if="!selectedMeta" class="mb-4">
        <h3 class="text-base font-medium mb-4">{{ $t('page.channel.selectType') }}</h3>
        <ChannelTypeSelector
          :channel-types="channelTypes"
          :selected-type="''"
          @select="handleTypeSelect"
        />
      </div>

      <!-- 步骤 2：填写渠道信息 -->
      <div v-else>
        <!-- 已选类型提示 -->
        <ElAlert
          :title="$t('page.channel.selectedType') + ': ' + selectedMeta.name"
          type="info"
          :closable="false"
          show-icon
          class="mb-4"
        >
          <template #default>
            <ElButton link type="primary" @click="handleChangeType">
              {{ $t('page.channel.changeType') }}
            </ElButton>
          </template>
        </ElAlert>

        <!-- 渠道名称 -->
        <ElForm
          ref="formRef"
          :model="formData"
          :rules="rules"
          label-width="140px"
          class="max-w-2xl mb-4"
        >
          <ElFormItem :label="$t('page.channel.name')" prop="name">
            <ElInput
              v-model="formData.name"
              :placeholder="$t('page.channel.namePlaceholder')"
              maxlength="128"
              show-word-limit
            />
          </ElFormItem>
        </ElForm>

        <!-- 动态渲染配置表单 -->
        <ElForm
          ref="configFormRef"
          :model="configData"
          :rules="configRules"
          label-width="140px"
          class="max-w-2xl mb-4"
        >
          <ElFormItem
            v-for="field in selectedMeta.config_fields"
            :key="field.name"
            :label="field.label"
            :prop="field.name"
          >
            <!-- URL 类型 -->
            <ElInput
              v-if="field.type === 'url'"
              v-model="configData[field.name]"
              :placeholder="field.placeholder"
              :disabled="selectedMeta.type === 'wechat_clawbot'"
            />
            <!-- 密码类型 -->
            <ElInput
              v-else-if="field.type === 'password'"
              v-model="configData[field.name]"
              type="password"
              :placeholder="field.placeholder"
              show-password
              :disabled="selectedMeta.type === 'wechat_clawbot'"
            />
            <!-- 数字类型 -->
            <ElInputNumber
              v-else-if="field.type === 'number'"
              v-model="configData[field.name]"
              :placeholder="field.placeholder"
              style="width: 100%"
              :disabled="selectedMeta.type === 'wechat_clawbot'"
            />
            <!-- 默认文本类型 -->
            <ElInput
              v-else
              v-model="configData[field.name]"
              :placeholder="field.placeholder"
              :disabled="selectedMeta.type === 'wechat_clawbot'"
            />
          </ElFormItem>
        </ElForm>

        <!-- 微信 ClawBot 特殊处理：嵌入微信绑定 -->
        <WechatBindSection
          v-if="selectedMeta.type === 'wechat_clawbot'"
          @bound="handleWechatBound"
        />

        <!-- 提交按钮 -->
        <ElFormItem class="mt-4">
          <ElButton type="primary" :loading="submitting" @click="handleSubmit">
            {{ $t('common.create') }}
          </ElButton>
          <ElButton @click="handleBack">
            {{ $t('common.cancel') }}
          </ElButton>
        </ElFormItem>
      </div>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import {
  ElCard,
  ElForm,
  ElFormItem,
  ElInput,
  ElInputNumber,
  ElButton,
  ElAlert,
  ElMessage,
  type FormInstance,
  type FormRules,
} from 'element-plus';
import { IconifyIcon } from '@vben/icons';
import { $t } from '#/locales';
import {
  getChannelTypesApi,
  createChannelApi,
  type ChannelApi,
} from '#/api/modules/channel';
import ChannelTypeSelector from '#/components/ChannelTypeSelector.vue';
import WechatBindSection from '#/components/WechatBindSection.vue';

const router = useRouter();

// 渠道类型列表
const channelTypes = ref<ChannelApi.ChannelTypeMeta[]>([]);

// 已选类型
const selectedMeta = ref<ChannelApi.ChannelTypeMeta | null>(null);

// 表单数据
const formData = reactive({ name: '' });
const configData = reactive<Record<string, any>>({});

// 动态校验规则
const configRules = computed<FormRules>(() => {
  if (!selectedMeta.value) return {};
  const rules: FormRules = {};
  selectedMeta.value.config_fields.forEach((field) => {
    if (field.required) {
      rules[field.name] = [
        { required: true, message: `${field.label}不能为空`, trigger: 'blur' },
      ];
    }
  });
  return rules;
});

const formRef = ref<FormInstance>();
const configFormRef = ref<FormInstance>();
const submitting = ref(false);

// 微信绑定后的凭证
const wechatCredential = ref<any>(null);

const rules: FormRules = {
  name: [
    { required: true, message: $t('page.channel.nameRequired'), trigger: 'blur' },
    { max: 128, message: $t('page.channel.nameMaxLength'), trigger: 'blur' },
  ],
};

// 加载渠道类型
async function loadChannelTypes() {
  try {
    channelTypes.value = await getChannelTypesApi();
  } catch {
    ElMessage.error($t('page.channel.loadTypesFailed'));
  }
}

// 选择渠道类型
function handleTypeSelect(meta: ChannelApi.ChannelTypeMeta) {
  // 清空旧的配置数据
  Object.keys(configData).forEach((key) => {
    delete configData[key];
  });
  selectedMeta.value = meta;
  // 初始化配置字段
  meta.config_fields.forEach((field) => {
    if (field.type === 'number') {
      configData[field.name] = field.name === 'smtp_port' ? 587 : undefined;
    } else {
      configData[field.name] = '';
    }
  });
  // 重置表单
  formData.name = '';
  wechatCredential.value = null;
}

// 更换类型
function handleChangeType() {
  selectedMeta.value = null;
}

// 微信绑定成功回调
function handleWechatBound(credential: any) {
  wechatCredential.value = credential;
  // 自动填充配置字段
  if (credential) {
    configData.bot_token_ciphertext = credential.bot_token_ciphertext || '';
    configData.bot_token_nonce = credential.bot_token_nonce || '';
    configData.ilink_bot_id = credential.ilink_bot_id || '';
    configData.ilink_user_id = credential.ilink_user_id || '';
  }
}

// 提交表单
async function handleSubmit() {
  if (!selectedMeta.value) return;

  // 校验名称
  const nameValid = await formRef.value?.validate().catch(() => false);
  if (!nameValid) return;

  // 校验配置
  const configValid = await configFormRef.value?.validate().catch(() => false);
  if (!configValid) return;

  // 微信 ClawBot 类型必须绑定成功
  if (selectedMeta.value.type === 'wechat_clawbot' && !wechatCredential.value) {
    ElMessage.warning($t('page.channel.requireBind'));
    return;
  }

  submitting.value = true;
  try {
    const config: Record<string, any> = {};
    selectedMeta.value.config_fields.forEach((field) => {
      config[field.name] = configData[field.name];
    });

    await createChannelApi({
      type: selectedMeta.value.type,
      name: formData.name,
      config,
    });

    ElMessage.success($t('page.channel.createSuccess'));
    router.push('/channel/list');
  } catch {
    ElMessage.error($t('page.channel.createFailed'));
  } finally {
    submitting.value = false;
  }
}

function handleBack() {
  router.push('/channel/list');
}

onMounted(() => {
  loadChannelTypes();
});
</script>
