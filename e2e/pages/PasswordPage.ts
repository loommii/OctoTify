import { type Page, type Locator } from '@playwright/test';
import { BasePage } from './BasePage';

/**
 * 修改密码页面对象 - /settings/password
 *
 * 对应前端页面：系统设置 - 修改密码
 * 包含：旧密码输入、新密码输入、确认新密码输入、提交按钮
 */
export class PasswordPage extends BasePage {
  readonly oldPasswordInput: Locator;
  readonly newPasswordInput: Locator;
  readonly confirmPasswordInput: Locator;
  readonly submitButton: Locator;
  readonly errorToast: Locator;
  readonly formErrors: Locator;

  constructor(page: Page) {
    super(page);
    // 旧密码输入框
    this.oldPasswordInput = page.getByPlaceholder('请输入旧密码');
    // 新密码输入框
    this.newPasswordInput = page.getByPlaceholder('请输入新密码');
    // 确认新密码输入框
    this.confirmPasswordInput = page.getByPlaceholder('请再次输入新密码');
    // 确认按钮
    this.submitButton = page.getByRole('button', { name: '确认' });
    // 错误提示（Element UI toast）
    this.errorToast = page.locator('.el-message--error .el-message__content');
    // 表单校验错误
    this.formErrors = page.locator('.el-form-item__error');
  }

  /** 导航到修改密码页 */
  async goto() {
    await super.goto('/settings/password');
  }

  /** 等待修改密码页加载完成 */
  async expectLoaded() {
    await this.page.waitForURL(/\/settings\/password/);
    await this.page.waitForSelector('text=修改密码');
  }

  /**
   * 修改密码（填写表单并提交）
   * @param oldPassword 旧密码
   * @param newPassword 新密码
   */
  async changePassword(oldPassword: string, newPassword: string) {
    await this.fill(this.oldPasswordInput, oldPassword);
    await this.fill(this.newPasswordInput, newPassword);
    await this.fill(this.confirmPasswordInput, newPassword);
    await this.click(this.submitButton);
  }

  /**
   * 填写密码表单（不提交）
   * 用于测试各种错误场景
   */
  async fillPasswordForm(oldPwd: string, newPwd: string, confirmPwd: string) {
    await this.oldPasswordInput.fill(oldPwd);
    await this.newPasswordInput.fill(newPwd);
    await this.confirmPasswordInput.fill(confirmPwd);
  }

  /** 获取错误提示文案（优先 toast，其次表单校验） */
  async getErrorMessage(): Promise<string> {
    const toastAppeared = await this.errorToast.waitFor({ state: 'visible', timeout: 2000 }).then(() => true).catch(() => false);
    if (toastAppeared) {
      return (await this.errorToast.textContent()) ?? '';
    }
    await this.formErrors.first().waitFor({ state: 'visible', timeout: 5000 });
    return (await this.formErrors.first().textContent()) ?? '';
  }
}
