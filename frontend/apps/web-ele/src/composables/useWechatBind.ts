import { ref } from 'vue';
import { ElMessage } from 'element-plus';
import { wechatBindBindApi, wechatBindStatusApi, type ChannelApi } from '#/api/modules/channel';

// 微信绑定状态：idle（未开始）| binding（绑定中）| confirmed（已确认）| expired（已过期）
export type WechatBindState = 'idle' | 'binding' | 'confirmed' | 'expired';

export function useWechatBind() {
  // 当前绑定状态
  const state = ref<WechatBindState>('idle');
  // 二维码图片 URL（已废弃，保留兼容）
  const qrCodeUrl = ref('');
  // 二维码原始值（用于轮询请求）
  const qrcode = ref('');
  // 绑定成功后的凭证信息
  const credential = ref<ChannelApi.WechatBindStatus['credential'] | null>(null);

  // 轮询定时器
  let pollingTimer: ReturnType<typeof setTimeout> | null = null;
  // 轮询计数器
  let pollCount = 0;
  // 最大轮询次数
  const MAX_POLL_COUNT = 10;

  /**
   * 发起扫码绑定
   * 调用后端接口获取二维码，启动轮询查询绑定状态
   */
  async function startBind() {
    try {
      const res = await wechatBindBindApi();
      qrcode.value = res.qrcode; // 保存二维码原始值用于轮询
      qrCodeUrl.value = res.qrcode_url; // 保存二维码 URL（兼容旧代码）
      state.value = 'binding';
      startPolling();
    } catch {
      ElMessage.error('发起扫码绑定失败');
    }
  }

  /**
   * 开始轮询绑定状态
   * 立即执行第一次查询，后端返回后立即发起下一次，最多轮询 10 次
   */
  function startPolling() {
    stopPolling();
    pollCount = 0;
    pollOnce();
  }

  /**
   * 单次轮询执行
   */
  async function pollOnce() {
    if (!qrcode.value) return;

    if (pollCount >= MAX_POLL_COUNT) {
      state.value = 'expired';
      stopPolling();
      ElMessage.warning('轮询次数已达上限，请重新扫码');
      return;
    }

    pollCount++;

    try {
      const res = await wechatBindStatusApi(qrcode.value);
      if (res.status === 'confirmed') {
        credential.value = res.credential ?? null;
        state.value = 'confirmed';
        stopPolling();
        ElMessage.success('微信绑定成功');
      } else if (res.status === 'expired') {
        state.value = 'expired';
        stopPolling();
        ElMessage.warning('二维码已过期，请重新扫码');
      } else {
        pollingTimer = setTimeout(pollOnce, 1000);
      }
    } catch {
      pollingTimer = setTimeout(pollOnce, 1000);
    }
  }

  /**
   * 停止轮询
   * 清除定时器，避免内存泄漏
   */
  function stopPolling() {
    if (pollingTimer) {
      clearTimeout(pollingTimer);
      pollingTimer = null;
    }
  }

  /**
   * 重置状态
   * 停止轮询并清空所有状态数据
   */
  function reset() {
    stopPolling();
    state.value = 'idle';
    qrCodeUrl.value = '';
    qrcode.value = '';
    credential.value = null;
  }

  return {
    state,
    qrCodeUrl,
    qrcode,
    credential,
    startBind,
    startPolling,
    stopPolling,
    reset,
  };
}
