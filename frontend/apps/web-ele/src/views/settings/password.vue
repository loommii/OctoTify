<template>
  <div class="p-4">
    <ElCard shadow="never">
      <template #header>
        <span class="font-semibold">{{ $t('page.settings.password') }}</span>
      </template>

      <ElForm
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-width="140px"
        style="max-width: 480px"
        @submit.prevent="handleSubmit"
      >
        <ElFormItem :label="$t('page.settings.oldPassword')" prop="old_password">
          <ElInput
            v-model="formData.old_password"
            type="password"
            show-password
            :placeholder="$t('page.settings.oldPasswordPlaceholder')"
          />
        </ElFormItem>

        <ElFormItem :label="$t('page.settings.newPassword')" prop="new_password">
          <ElInput
            v-model="formData.new_password"
            type="password"
            show-password
            :placeholder="$t('page.settings.newPasswordPlaceholder')"
          />
        </ElFormItem>

        <ElFormItem :label="$t('page.settings.confirmPassword')" prop="confirm_password">
          <ElInput
            v-model="formData.confirm_password"
            type="password"
            show-password
            :placeholder="$t('page.settings.confirmPasswordPlaceholder')"
          />
        </ElFormItem>

        <ElFormItem>
          <div class="text-gray-500 text-sm mb-2">
            {{ $t('page.settings.passwordStrengthRule') }}
          </div>
          <ElButton type="primary" :loading="submitting" @click="handleSubmit">
            {{ $t('common.confirm') }}
          </ElButton>
        </ElFormItem>
      </ElForm>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import type { FormInstance, FormRules } from 'element-plus';
import {
  ElCard,
  ElForm,
  ElFormItem,
  ElInput,
  ElButton,
  ElMessage,
} from 'element-plus';
import { $t } from '#/locales';
import { changePasswordApi } from '#/api/core/user';
import { useAuthStore } from '#/store';

const authStore = useAuthStore();
const formRef = ref<FormInstance>();
const submitting = ref(false);

const formData = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
});

// 密码强度校验：8-64 字符，至少包含小写字母、大写字母、数字各一个
const passwordPattern = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).{8,64}$/;

const rules: FormRules = {
  old_password: [
    { required: true, message: $t('page.settings.oldPasswordRequired'), trigger: 'blur' },
  ],
  new_password: [
    { required: true, message: $t('page.settings.newPasswordRequired'), trigger: 'blur' },
    {
      validator: (_rule, value, callback) => {
        if (value === formData.old_password) {
          callback(new Error($t('page.settings.newPasswordSameAsOld')));
        } else {
          callback();
        }
      },
      trigger: 'blur',
    },
    {
      pattern: passwordPattern,
      message: $t('page.settings.passwordStrengthRule'),
      trigger: 'blur',
    },
  ],
  confirm_password: [
    { required: true, message: $t('page.settings.confirmPasswordRequired'), trigger: 'blur' },
    {
      validator: (_rule, value, callback) => {
        if (value !== formData.new_password) {
          callback(new Error($t('page.settings.passwordMismatch')));
        } else {
          callback();
        }
      },
      trigger: 'blur',
    },
  ],
};

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false);
  if (!valid) return;

  submitting.value = true;
  try {
    await changePasswordApi({
      old_password: formData.old_password,
      new_password: formData.new_password,
    });
    ElMessage.success($t('page.settings.passwordChangeSuccess'));
    // 跳转登录页
    await authStore.logout(false);
  } catch {
    // 错误消息已由框架的 errorMessageResponseInterceptor 统一处理
  } finally {
    submitting.value = false;
  }
}
</script>
