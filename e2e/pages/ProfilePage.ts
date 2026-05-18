import { type Page, type Locator } from '@playwright/test';
import { BasePage } from './BasePage';

export class ProfilePage extends BasePage {
  readonly usernameCell: Locator;
  readonly userIdCell: Locator;
  readonly createdAtCell: Locator;

  constructor(page: Page) {
    super(page);
    this.usernameCell = page.locator('tr').filter({ hasText: '用户名' }).locator('td').last();
    this.userIdCell = page.locator('tr').filter({ hasText: '用户 ID' }).locator('td').last();
    this.createdAtCell = page.locator('tr').filter({ hasText: '注册时间' }).locator('td').last();
  }

  async goto() {
    await super.goto('/settings/profile');
  }

  async expectLoaded() {
    await this.page.waitForURL(/\/settings\/profile/);
    await this.page.waitForSelector('text=个人资料');
  }

  async getUsername(): Promise<string> {
    return (await this.usernameCell.textContent()) ?? '';
  }

  async getUserId(): Promise<string> {
    return (await this.userIdCell.textContent()) ?? '';
  }

  async getCreatedAt(): Promise<string> {
    return (await this.createdAtCell.textContent()) ?? '';
  }
}
