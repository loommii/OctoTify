import { type Page } from '@playwright/test';
import { BasePage } from './BasePage';

/**
 * Dashboard 页面对象 - 仪表盘页面
 *
 * 包含仪表盘特有的操作（统计数据、最近推送等）
 * Header 相关操作已移至 Header 组件
 */
export class DashboardPage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  async goto() {
    await super.goto('/dashboard/index');
  }

  /**
   * 检查是否已加载到 Dashboard 页面
   */
  async isLoaded(): Promise<boolean> {
    try {
      await this.page.waitForURL(/\/dashboard/, { timeout: 5000 });
      return true;
    } catch {
      return false;
    }
  }

  /**
   * 获取消息来源总数
   */
  async getMessageSourceCount(): Promise<string> {
    const el = this.page.locator('text=消息来源总数').locator('..').locator('p').last();
    return (await el.textContent()) ?? '0';
  }

  /**
   * 获取推送渠道总数
   */
  async getChannelCount(): Promise<string> {
    const el = this.page.locator('text=推送渠道总数').locator('..').locator('p').last();
    return (await el.textContent()) ?? '0';
  }

  /**
   * 获取今日推送数量
   */
  async getTodayPushCount(): Promise<string> {
    const el = this.page.locator('text=今日推送').locator('..').locator('p').last();
    return (await el.textContent()) ?? '0';
  }
}
