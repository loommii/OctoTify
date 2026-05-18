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

  constructor(page: Page) {
    super(page);
    this.usernameInput = page.locator('input[name="username"]');
    this.passwordInput = page.locator('input[name="password"]');
    this.confirmPasswordInput = page.locator('input[name="confirmPassword"]');
    this.registerButton = page.getByRole('button', { name: 'register' });
    this.loginLink = page.getByText('已经有账号了? 去登录');
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
    // 等待页面加载完成
    await this.page.waitForTimeout(1000);
    
    // 使用 locator.fill() 填写表单
    await this.usernameInput.waitFor({ state: 'visible', timeout: 10000 });
    await this.usernameInput.fill(username);
    await this.passwordInput.fill(password);
    await this.confirmPasswordInput.fill(password);
    
    // 点击注册按钮
    await this.registerButton.click();
    
    // 等待跳转到登录页
    await this.page.waitForURL(/\/auth\/login/, { timeout: 15000 });
  }

  /**
   * 检查是否在注册页
   */
  async expectLoaded() {
    await this.page.waitForURL(/\/auth\/register/);
  }
}
