<script lang="ts" setup>
import { $t } from '@vben/locales';

defineOptions({ name: 'QrCodeLogin' });
</script>

<template>
  <div class="flex-center">
    <p>{{ $t('authentication.qrcodeLogin') }}</p>
  </div>
</template>
