import { type Page, type Locator } from '@playwright/test';
import { BasePage } from './BasePage';

/**
 * 登录页面对象 - /auth/login
 *
 * 对应前端页面：用户登录页
 * 包含：用户名输入、密码输入、登录按钮、注册链接
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
    // 用户名输入框
    this.usernameInput = page.locator('input[name="username"]');
    // 密码输入框
    this.passwordInput = page.locator('input[name="password"]');
    // 登录按钮
    this.loginButton = page.getByRole('button', { name: 'login' });
    // 注册链接
    this.registerLink = page.getByText('还没有账号? 创建账号');
    // 错误提示（Element UI toast）
    this.errorToast = page.locator('.el-message--error .el-message__content');
    // 表单校验错误
    this.formAlert = page.locator('[role="alert"]');
  }

  /** 导航到登录页 */
  async goto() {
    await super.goto('/auth/login');
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * 登录并等待跳转到仪表盘
   * @param username 用户名
   * @param password 密码
   */
  async login(username: string, password: string) {
    await this.fillLoginForm(username, password);
    await this.page.waitForURL(/\/dashboard/, { timeout: 30000 });
  }

  /** 等待登录页加载完成 */
  async expectLoaded() {
    await this.page.waitForURL(/\/auth\/login/);
  }

  /**
   * 填写登录表单并点击登录（不等待跳转）
   * 使用 pressSequentially 逐字符输入，确保 Element Plus v-model 正确绑定
   */
  async fillLoginForm(username: string, password: string) {
    await this.usernameInput.waitFor({ state: 'visible', timeout: 10000 });
    await this.usernameInput.click();
    await this.usernameInput.pressSequentially(username);
    await this.passwordInput.click();
    await this.passwordInput.pressSequentially(password);
    await this.loginButton.click();
  }

  /** 获取错误提示文案（优先 toast，其次表单校验） */
  async getErrorMessage(): Promise<string> {
    const toastVisible = await this.errorToast.isVisible({ timeout: 2000 }).catch(() => false);
    if (toastVisible) {
      return (await this.errorToast.textContent()) ?? '';
    }
    await this.formAlert.first().waitFor({ state: 'visible', timeout: 5000 });
    return (await this.formAlert.first().textContent()) ?? '';
  }
}
