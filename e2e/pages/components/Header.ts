import { type Page, type Locator, expect } from '@playwright/test';
import { BasePage } from '../BasePage';

/**
 * 顶部导航栏组件 - 通用组件（所有已登录页面共享）
 *
 * 包含：
 * - 用户头像菜单（退出登录、个人中心、修改密码）
 * - 通知弹窗关闭
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
    // 用户头像下拉菜单触发按钮
    this.userDropdown = page.locator('header button').last();
    // 退出登录菜单项
    this.logoutButton = page.getByText('退出登录');
    // 个人中心菜单项
    this.profileButton = page.getByText('个人中心');
    // 修改密码菜单项
    this.passwordButton = page.getByText('修改密码');
    // 通知弹窗关闭按钮
    this.notificationCloseBtn = page.locator('.el-notification__closeBtn');
    // 确认按钮（退出登录确认弹窗）
    this.confirmButton = page.getByRole('button', { name: '确认' });
  }

  /** 关闭通知弹窗（如果存在） */
  async closeNotification() {
    try {
      await expect(this.notificationCloseBtn).toBeVisible({ timeout: 2000 });
      await this.notificationCloseBtn.click();
    } catch {
      // 通知不存在或超时，忽略
    }
  }

  /** 打开用户头像菜单 */
  async openUserMenu() {
    await this.userDropdown.click();
  }

  /**
   * 退出登录完整流程
   * 流程：关闭通知 → 点头像 → 点退出登录 → 确认 → 等待跳转到登录页
   */
  async logout() {
    await this.closeNotification();
    await this.openUserMenu();
    await this.logoutButton.click({ force: true });
    await expect(this.confirmButton).toBeVisible({ timeout: 5000 });
    await this.confirmButton.click();
    await this.page.waitForURL(/\/auth\/login/, { timeout: 10000 });
  }

  /** 跳转到个人资料页 */
  async gotoProfile() {
    await this.closeNotification();
    await this.page.goto('/settings/profile');
    await this.page.waitForURL(/\/settings\/profile/, { timeout: 10000 });
  }

  /** 跳转到修改密码页 */
  async gotoPassword() {
    await this.closeNotification();
    await this.page.goto('/settings/password');
    await this.page.waitForURL(/\/settings\/password/, { timeout: 10000 });
  }
}
