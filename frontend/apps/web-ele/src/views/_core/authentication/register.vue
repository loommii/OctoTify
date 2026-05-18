<script lang="ts" setup>
import type { VbenFormSchema } from '@vben/common-ui';
import type { Recordable } from '@vben/types';

import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';

import { AuthenticationRegister, z } from '@vben/common-ui';
import { LOGIN_PATH } from '@vben/constants';
import { $t } from '@vben/locales';

import { ElMessage } from 'element-plus';

import { registerApi } from '#/api/core/auth';

defineOptions({ name: 'Register' });

const router = useRouter();
const loading = ref(false);

/**
 * 密码强度校验：8-128 字符，需包含大小写字母+数字
 */
const passwordRule = z
  .string()
  .min(8, { message: $t('authentication.passwordMinLength') })
  .max(64, { message: $t('authentication.passwordMaxLength') })
  .regex(/[a-z]/, { message: $t('authentication.passwordRequireLowercase') })
  .regex(/[A-Z]/, { message: $t('authentication.passwordRequireUppercase') })
  .regex(/[0-9]/, { message: $t('authentication.passwordRequireNumber') });

const formSchema = computed((): VbenFormSchema[] => {
  return [
    {
      component: 'VbenInput',
      componentProps: {
        placeholder: $t('authentication.usernameTip'),
      },
      fieldName: 'username',
      label: $t('authentication.username'),
      rules: z
        .string()
        .min(1, { message: $t('authentication.usernameTip') })
        .max(64, { message: $t('authentication.usernameMaxLength') }),
    },
    {
      component: 'VbenInputPassword',
      componentProps: {
        passwordStrength: true,
        placeholder: $t('authentication.password'),
      },
      fieldName: 'password',
      label: $t('authentication.password'),
      renderComponentContent() {
        return {
          strengthText: () => $t('authentication.passwordStrength'),
        };
      },
      rules: passwordRule,
    },
    {
      component: 'VbenInputPassword',
      componentProps: {
        placeholder: $t('authentication.confirmPassword'),
      },
      dependencies: {
        rules(values) {
          const { password } = values;
          return z
            .string({ required_error: $t('authentication.passwordTip') })
            .min(1, { message: $t('authentication.passwordTip') })
            .refine((value) => value === password, {
              message: $t('authentication.confirmPasswordTip'),
            });
        },
        triggerFields: ['password'],
      },
      fieldName: 'confirmPassword',
      label: $t('authentication.confirmPassword'),
    },
  ];
});

async function handleSubmit(values: Recordable<any>) {
  try {
    loading.value = true;
    await registerApi({
      password: values.password,
      username: values.username,
    });
    ElMessage.success($t('authentication.registerSuccess'));
    // 注册成功后跳转登录页
    router.push({
      path: LOGIN_PATH,
    });
  } catch {
    // 错误已由全局拦截器处理
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <AuthenticationRegister
    :form-schema="formSchema"
    :loading="loading"
    @submit="handleSubmit"
  />
</template>
