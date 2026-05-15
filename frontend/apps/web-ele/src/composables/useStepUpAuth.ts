import { ref } from 'vue';

/**
 * 密码二次验证 Composable
 *
 * 用法示例：
 * ```ts
 * const { visible, verifyPassword, handleConfirm, handleCancel } = useStepUpAuth();
 *
 * async function handleDelete(id: number) {
 *   const password = await verifyPassword('确认删除', '删除操作不可恢复');
 *   if (password) {
 *     await deleteSourceApi(id, { password });
 *   }
 * }
 * ```
 */
export function useStepUpAuth() {
  const visible = ref(false);
  const pendingResolve = ref<((value: string | null) => void) | null>(null);

  /**
   * 显示密码验证对话框，返回 Promise<string | null>
   * @param title 对话框标题
   * @param description 对话框描述
   * @returns 用户输入的密码，取消时返回 null
   */
  function verifyPassword(title: string, description: string): Promise<string | null> {
    return new Promise((resolve) => {
      pendingResolve.value = resolve;
      (window as any).__stepUpAuthPending = { title, description, resolve };
      visible.value = true;
    });
  }

  function handleConfirm(password: string) {
    visible.value = false;
    pendingResolve.value?.(password);
    pendingResolve.value = null;
  }

  function handleCancel() {
    visible.value = false;
    pendingResolve.value?.(null);
    pendingResolve.value = null;
  }

  return {
    visible,
    verifyPassword,
    handleConfirm,
    handleCancel,
  };
}
