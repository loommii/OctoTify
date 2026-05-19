import { type Page, type Locator } from '@playwright/test';
import { BasePage } from './BasePage';

/**
 * 个人资料页面对象 - /settings/profile
 *
 * 对应前端页面：系统设置 - 个人资料
 * 包含：用户名、用户ID、注册时间等信息展示
 */
export class ProfilePage extends BasePage {
  readonly usernameCell: Locator;
  readonly userIdCell: Locator;
  readonly createdAtCell: Locator;

  constructor(page: Page) {
    super(page);
    // 用户名（从描述表格的"用户名"行读取值）
    this.usernameCell = page.locator('tr').filter({ hasText: '用户名' }).locator('td').last();
    // 用户ID
    this.userIdCell = page.locator('tr').filter({ hasText: '用户 ID' }).locator('td').last();
    // 注册时间
    this.createdAtCell = page.locator('tr').filter({ hasText: '注册时间' }).locator('td').last();
  }

  /** 导航到个人资料页 */
  async goto() {
    await super.goto('/settings/profile');
  }

  /** 等待个人资料页加载完成 */
  async expectLoaded() {
    await this.page.waitForURL(/\/settings\/profile/);
    await this.page.waitForSelector('text=个人资料');
  }

  /** 获取用户名 */
  async getUsername(): Promise<string> {
    return (await this.usernameCell.textContent()) ?? '';
  }

  /** 获取用户ID */
  async getUserId(): Promise<string> {
    return (await this.userIdCell.textContent()) ?? '';
  }

  /** 获取注册时间 */
  async getCreatedAt(): Promise<string> {
    return (await this.createdAtCell.textContent()) ?? '';
  }
}
