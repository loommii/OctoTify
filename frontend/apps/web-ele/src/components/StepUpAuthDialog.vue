<template>
  <ElDialog
    :model-value="visible"
    :title="title"
    width="440px"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    @update:model-value="$emit('update:visible', $event)"
  >
    <div class="stepup-auth-dialog-content">
      <p class="description">{{ description }}</p>
      <ElForm
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-position="top"
        @submit.prevent="handleConfirm"
      >
        <ElFormItem prop="password" :label="$t('page.source.authPassword')">
          <ElInput
            v-model="formData.password"
            type="password"
            show-password
            :placeholder="$t('page.source.authPasswordPlaceholder')"
            @keyup.enter="handleConfirm"
          />
        </ElFormItem>
      </ElForm>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <ElButton @click="handleCancel">
          {{ $t('common.cancel') }}
        </ElButton>
        <ElButton
          type="primary"
          :loading="loading"
          :disabled="!formData.password"
          @click="handleConfirm"
        >
          {{ $t('common.confirm') }}
        </ElButton>
      </div>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue';
import { ElDialog, ElForm, ElFormItem, ElInput, ElButton, type FormInstance, type FormRules } from 'element-plus';
import { $t } from '#/locales';

interface Props {
  visible: boolean;
  title: string;
  description: string;
}

interface Emits {
  (e: 'update:visible', value: boolean): void;
  (e: 'confirm', password: string): void;
  (e: 'cancel'): void;
}

defineProps<Props>();
const emit = defineEmits<Emits>();

const formRef = ref<FormInstance>();
const loading = ref(false);

const formData = reactive({
  password: '',
});

const rules: FormRules = {
  password: [
    { required: true, message: $t('page.source.authPasswordRequired'), trigger: 'blur' },
  ],
};

async function handleConfirm() {
  if (!formRef.value) return;

  const valid = await formRef.value.validate().catch(() => false);
  if (!valid) return;

  loading.value = true;
  try {
    emit('confirm', formData.password);
  } finally {
    loading.value = false;
  }
}

function handleCancel() {
  formData.password = '';
  formRef.value?.clearValidate();
  emit('cancel');
  emit('update:visible', false);
}
</script>

<style scoped>
.stepup-auth-dialog-content {
  padding: 8px 0;
}

.description {
  color: var(--el-text-color-regular);
  font-size: 14px;
  margin-bottom: 20px;
  line-height: 1.5;
}
</style>
