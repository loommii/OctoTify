import { type Page, type Locator } from '@playwright/test';
import { BasePage } from './BasePage';

/**
 * 注册页面对象
 */
export class RegisterPage extends BasePage {
  readonly usernameInput: Locator;
  readonly passwordInput: Locator;
  readonly confirmPasswordInput: Locator;
  readonly registerButton: Locator;
  readonly loginLink: Locator;
  readonly errorToast: Locator;
  readonly formAlert: Locator;

  constructor(page: Page) {
    super(page);
    this.usernameInput = page.locator('input[name="username"]');
    this.passwordInput = page.locator('input[name="password"]');
    this.confirmPasswordInput = page.locator('input[name="confirmPassword"]');
    this.registerButton = page.getByRole('button', { name: 'register' });
    this.loginLink = page.getByText('已经有账号了? 去登录');
    this.errorToast = page.locator('.el-message--error .el-message__content');
    this.formAlert = page.locator('[role="alert"]');
  }

  async goto() {
    await super.goto('/auth/register');
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * 注册新用户
   * @param username 用户名
   * @param password 密码
   */
  async register(username: string, password: string) {
    await this.fillRegisterForm(username, password);
    await this.page.waitForURL(/\/auth\/login/, { timeout: 15000 });
  }

  /**
   * 检查是否在注册页
   */
  async expectLoaded() {
    await this.page.waitForURL(/\/auth\/register/);
  }

  /**
   * 填写表单并提交，但不等待页面跳转
   * 用于测试注册失败场景
   */
  async fillRegisterForm(username: string, password: string) {
    await this.usernameInput.waitFor({ state: 'visible', timeout: 10000 });
    await this.usernameInput.click();
    await this.usernameInput.type(username);
    await this.passwordInput.click();
    await this.passwordInput.type(password);
    await this.confirmPasswordInput.click();
    await this.confirmPasswordInput.type(password);
    await this.registerButton.click();
  }

  /**
   * 获取错误提示文本（Element UI toast）
   */
  async getErrorMessage(): Promise<string> {
    const toastVisible = await this.errorToast.isVisible({ timeout: 2000 }).catch(() => false);
    if (toastVisible) {
      return (await this.errorToast.textContent()) ?? '';
    }
    await this.formAlert.first().waitFor({ state: 'visible', timeout: 5000 });
    return (await this.formAlert.first().textContent()) ?? '';
  }
}
