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

  constructor(page: Page) {
    super(page);
    this.usernameInput = page.locator('input[name="username"]');
    this.passwordInput = page.locator('input[name="password"]');
    this.loginButton = page.getByRole('button', { name: 'login' });
    this.registerLink = page.getByText('还没有账号? 创建账号');
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
    // 等待页面加载完成
    await this.page.waitForTimeout(1000);
    
    // 使用 locator.fill() 填写表单
    await this.usernameInput.waitFor({ state: 'visible', timeout: 10000 });
    await this.usernameInput.fill(username);
    await this.passwordInput.fill(password);
    
    // 点击登录按钮
    await this.loginButton.click();
    
    // 等待跳转到 dashboard
    await this.page.waitForURL(/\/dashboard/, { timeout: 15000 });
  }

  /**
   * 检查是否在登录页
   */
  async expectLoaded() {
    await this.page.waitForURL(/\/auth\/login/);
  }
}
