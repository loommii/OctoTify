<script lang="ts" setup>
import type { VbenFormSchema } from '@vben/common-ui';

import { computed } from 'vue';

import { AuthenticationLogin, z } from '@vben/common-ui';
import { $t } from '@vben/locales';

import { useAuthStore } from '#/store';

defineOptions({ name: 'CodeLogin' });

const authStore = useAuthStore();

const formSchema = computed((): VbenFormSchema[] => {
  return [
    {
      component: 'VbenInput',
      componentProps: {
        placeholder: $t('authentication.phone'),
      },
      fieldName: 'phone',
      label: $t('authentication.phone'),
      rules: z.string().min(1, { message: $t('authentication.phoneTip') }),
    },
    {
      component: 'VbenInput',
      componentProps: {
        placeholder: $t('authentication.code'),
      },
      fieldName: 'code',
      label: $t('authentication.code'),
      rules: z.string().min(1, { message: $t('authentication.codeTip') }),
    },
  ];
});
</script>

<template>
  <AuthenticationLogin
    :form-schema="formSchema"
    :loading="authStore.loginLoading"
    @submit="authStore.authLogin"
  />
</template>
