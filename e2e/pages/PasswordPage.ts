import { type Page, type Locator } from '@playwright/test';
import { BasePage } from './BasePage';

export class PasswordPage extends BasePage {
  readonly oldPasswordInput: Locator;
  readonly newPasswordInput: Locator;
  readonly confirmPasswordInput: Locator;
  readonly submitButton: Locator;
  readonly errorToast: Locator;
  readonly formErrors: Locator;

  constructor(page: Page) {
    super(page);
    this.oldPasswordInput = page.getByPlaceholder('请输入旧密码');
    this.newPasswordInput = page.getByPlaceholder('请输入新密码');
    this.confirmPasswordInput = page.getByPlaceholder('请再次输入新密码');
    this.submitButton = page.getByRole('button', { name: '确认' });
    this.errorToast = page.locator('.el-message--error .el-message__content');
    this.formErrors = page.locator('.el-form-item__error');
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

  async fillPasswordForm(oldPwd: string, newPwd: string, confirmPwd: string) {
    await this.oldPasswordInput.fill(oldPwd);
    await this.newPasswordInput.fill(newPwd);
    await this.confirmPasswordInput.fill(confirmPwd);
  }

  async getErrorMessage(): Promise<string> {
    const toastAppeared = await this.errorToast.waitFor({ state: 'visible', timeout: 2000 }).then(() => true).catch(() => false);
    if (toastAppeared) {
      return (await this.errorToast.textContent()) ?? '';
    }
    await this.formErrors.first().waitFor({ state: 'visible', timeout: 5000 });
    return (await this.formErrors.first().textContent()) ?? '';
  }
}
