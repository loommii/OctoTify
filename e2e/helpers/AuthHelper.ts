import { type Page } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { RegisterPage } from '../pages/RegisterPage';

/**
 * 认证辅助类
 * 用于封装注册+登录的重复流程
 */
export class AuthHelper {
  private readonly page: Page;
  private readonly loginPage: LoginPage;
  private readonly registerPage: RegisterPage;

  constructor(page: Page) {
    this.page = page;
    this.loginPage = new LoginPage(page);
    this.registerPage = new RegisterPage(page);
  }

  /**
   * 注册新用户并登录
   * @param username 用户名（由调用方生成）
   * @param password 密码（由调用方指定）
   */
  async registerAndLogin(username: string, password: string) {
    await this.registerPage.goto();
    await this.registerPage.register(username, password);
    await this.page.waitForURL(/\/auth\/login/, { timeout: 15000 });

    await this.loginPage.goto();
    await this.loginPage.login(username, password);
    await this.page.waitForURL(/\/dashboard/, { timeout: 15000 });
  }
}
