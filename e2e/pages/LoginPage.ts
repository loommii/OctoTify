import { type Page, type Locator } from '@playwright/test';
import { BasePage } from './BasePage';

/**
 * 登录页面对象
 */
export class LoginPage extends BasePage {
  readonly usernameInput: Locator;
  readonly passwordInput: Locator;
  readonly loginButton: Locator;
  readonly registerLink: Locator;
  readonly errorToast: Locator;
  readonly formAlert: Locator;

  constructor(page: Page) {
    super(page);
    this.usernameInput = page.locator('input[name="username"]');
    this.passwordInput = page.locator('input[name="password"]');
    this.loginButton = page.getByRole('button', { name: 'login' });
    this.registerLink = page.getByText('还没有账号? 创建账号');
    this.errorToast = page.locator('.el-message--error .el-message__content');
    this.formAlert = page.locator('[role="alert"]');
  }

  async goto() {
    await super.goto('/auth/login');
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * 登录
   * @param username 用户名
   * @param password 密码
   */
  async login(username: string, password: string) {
    await this.fillLoginForm(username, password);
    await this.page.waitForURL(/\/dashboard/, { timeout: 15000 });
  }

  async expectLoaded() {
    await this.page.waitForURL(/\/auth\/login/);
  }

  async fillLoginForm(username: string, password: string) {
    await this.usernameInput.waitFor({ state: 'visible', timeout: 10000 });
    await this.usernameInput.click();
    await this.usernameInput.type(username);
    await this.passwordInput.click();
    await this.passwordInput.type(password);
    await this.loginButton.click();
  }

  async getErrorMessage(): Promise<string> {
    const toastVisible = await this.errorToast.isVisible({ timeout: 2000 }).catch(() => false);
    if (toastVisible) {
      return (await this.errorToast.textContent()) ?? '';
    }
    await this.formAlert.first().waitFor({ state: 'visible', timeout: 5000 });
    return (await this.formAlert.first().textContent()) ?? '';
  }
}
