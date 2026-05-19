import { type Page, type Locator } from '@playwright/test';
import { BasePage } from './BasePage';

/**
 * 注册页面对象 - /auth/register
 *
 * 对应前端页面：用户注册页
 * 包含：用户名输入、密码输入、确认密码输入、注册按钮、登录链接
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
    // 用户名输入框
    this.usernameInput = page.locator('input[name="username"]');
    // 密码输入框
    this.passwordInput = page.locator('input[name="password"]');
    // 确认密码输入框
    this.confirmPasswordInput = page.locator('input[name="confirmPassword"]');
    // 注册按钮
    this.registerButton = page.getByRole('button', { name: 'register' });
    // 登录链接
    this.loginLink = page.getByText('已经有账号了? 去登录');
    // 错误提示（Element UI toast）
    this.errorToast = page.locator('.el-message--error .el-message__content');
    // 表单校验错误
    this.formAlert = page.locator('[role="alert"]');
  }

  /** 导航到注册页 */
  async goto() {
    await super.goto('/auth/register');
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * 注册新用户并等待跳转到登录页
   * @param username 用户名
   * @param password 密码
   */
  async register(username: string, password: string) {
    await this.fillRegisterForm(username, password);
    await this.page.waitForURL(/\/auth\/login/, { timeout: 15000 });
  }

  /** 等待注册页加载完成 */
  async expectLoaded() {
    await this.page.waitForURL(/\/auth\/register/);
  }

  /**
   * 填写注册表单并提交（不等待页面跳转）
   * 用于测试注册失败场景
   */
  async fillRegisterForm(username: string, password: string) {
    await this.usernameInput.waitFor({ state: 'visible', timeout: 10000 });
    await this.usernameInput.fill(username);
    await this.passwordInput.fill(password);
    await this.confirmPasswordInput.fill(password);
    await this.registerButton.click();
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
