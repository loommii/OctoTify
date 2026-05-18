import { type Page, type Locator } from '@playwright/test';
import { BasePage } from './BasePage';

export class PasswordPage extends BasePage {
  readonly oldPasswordInput: Locator;
  readonly newPasswordInput: Locator;
  readonly confirmPasswordInput: Locator;
  readonly submitButton: Locator;

  constructor(page: Page) {
    super(page);
    this.oldPasswordInput = page.getByPlaceholder('请输入旧密码');
    this.newPasswordInput = page.getByPlaceholder('请输入新密码');
    this.confirmPasswordInput = page.getByPlaceholder('请再次输入新密码');
    this.submitButton = page.getByRole('button', { name: '确认' });
  }

  async goto() {
    await super.goto('/settings/password');
  }

  async expectLoaded() {
    await this.page.waitForURL(/\/settings\/password/);
    await this.page.waitForSelector('text=修改密码');
  }

  async changePassword(oldPassword: string, newPassword: string) {
    await this.fill(this.oldPasswordInput, oldPassword);
    await this.fill(this.newPasswordInput, newPassword);
    await this.fill(this.confirmPasswordInput, newPassword);
    await this.click(this.submitButton);
  }
}
