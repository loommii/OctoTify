<script lang="ts" setup>
import type { VbenFormSchema } from '@vben/common-ui';

import { computed } from 'vue';

import { AuthenticationLogin, z } from '@vben/common-ui';
import { $t } from '@vben/locales';

defineOptions({ name: 'ForgetPassword' });

const formSchema = computed((): VbenFormSchema[] => {
  return [
    {
      component: 'VbenInput',
      componentProps: {
        placeholder: $t('authentication.email'),
      },
      fieldName: 'email',
      label: $t('authentication.email'),
      rules: z.string().min(1, { message: $t('authentication.emailTip') }),
    },
  ];
});
</script>

<template>
  <AuthenticationLogin
    :form-schema="formSchema"
    :loading="false"
    @submit="() => {}"
  />
</template>
