import { type Page } from '@playwright/test';
import { BasePage } from './BasePage';

/**
 * 仪表盘页面对象 - /dashboard/index
 *
 * 对应前端页面：系统概览仪表盘
 * 包含：消息来源总数、推送渠道总数、今日推送等统计卡片
 */
export class DashboardPage extends BasePage {
  constructor(page: Page) {
    super(page);
  }

  /** 导航到仪表盘页 */
  async goto() {
    await super.goto('/dashboard/index');
    await this.page.waitForLoadState('networkidle');
  }

  /** 检查是否已加载到仪表盘页面 */
  async isLoaded(): Promise<boolean> {
    try {
      await this.page.waitForURL(/\/dashboard/, { timeout: 5000 });
      return true;
    } catch {
      return false;
    }
  }

  /** 获取消息来源总数 */
  async getMessageSourceCount(): Promise<string> {
    const el = this.page.locator('text=消息来源总数').locator('..').locator('p').last();
    return (await el.textContent()) ?? '0';
  }

  /** 获取推送渠道总数 */
  async getChannelCount(): Promise<string> {
    const el = this.page.locator('text=推送渠道总数').locator('..').locator('p').last();
    return (await el.textContent()) ?? '0';
  }

  /** 获取今日推送数量 */
  async getTodayPushCount(): Promise<string> {
    const el = this.page.locator('text=今日推送').locator('..').locator('p').last();
    return (await el.textContent()) ?? '0';
  }
}
