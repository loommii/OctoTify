import { type Page, type Locator, expect } from '@playwright/test';
import { BasePage } from '../BasePage';

/**
 * Header 组件 - 页面顶部导航栏
 *
 * 包含：
 * - 用户头像菜单
 * - 退出登录
 * - 关闭通知弹窗
 */
export class Header extends BasePage {
  readonly userDropdown: Locator;
  readonly logoutButton: Locator;
  readonly profileButton: Locator;
  readonly passwordButton: Locator;
  readonly notificationCloseBtn: Locator;
  readonly confirmButton: Locator;

  constructor(page: Page) {
    super(page);
    this.userDropdown = page.locator('header button').last();
    this.logoutButton = page.getByText('退出登录');
    this.profileButton = page.getByText('个人中心');
    this.passwordButton = page.getByText('修改密码');
    this.notificationCloseBtn = page.locator('.el-notification__closeBtn');
    this.confirmButton = page.getByRole('button', { name: '确认' });
  }

  /**
   * 关闭通知弹窗（如果存在）
   *
   * 使用 web-first 断言检查可见性，Playwright 自动等待动画完成
   */
  async closeNotification() {
    try {
      await expect(this.notificationCloseBtn).toBeVisible({ timeout: 2000 });
      await this.notificationCloseBtn.click();
    } catch {
      // 通知不存在或超时，忽略
    }
  }

  /**
   * 打开用户菜单
   *
   * Playwright click() 会自动等待元素稳定（动画完成）
   */
  async openUserMenu() {
    await this.userDropdown.click();
  }

  /**
   * 退出登录
   *
   * 流程：关闭通知 → 点头像 → 点退出登录 → 确认
   */
  async logout() {
    await this.closeNotification();
    await this.openUserMenu();
    await this.logoutButton.click({ force: true });
    await expect(this.confirmButton).toBeVisible({ timeout: 5000 });
    await this.confirmButton.click();
    await this.page.waitForURL(/\/auth\/login/, { timeout: 10000 });
  }

  /**
   * 跳转到个人资料页（直接导航）
   */
  async gotoProfile() {
    await this.closeNotification();
    await this.page.goto('/settings/profile');
    await this.page.waitForURL(/\/settings\/profile/, { timeout: 10000 });
  }

  /**
   * 跳转到修改密码页（直接导航）
   */
  async gotoPassword() {
    await this.closeNotification();
    await this.page.goto('/settings/password');
    await this.page.waitForURL(/\/settings\/password/, { timeout: 10000 });
  }
}
