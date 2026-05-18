import { test, expect } from '../fixtures/test-fixtures';
import { LoginPage } from '../pages/LoginPage';
import { Header } from '../pages/components/Header';

/**
 * A-E2E: 完整认证流程
 *
 * 测试覆盖完整的用户认证生命周期：
 * 注册 → 登录 → 查看资料 → 修改密码 → 新密码登录 → 登录持久化 → 退出登录 → 验证未登录状态
 *
 * 测试数据：使用随机用户名避免并发冲突
 */
test.describe('A-E2E: 完整认证流程', () => {
  // 生成随机用户名，确保每次测试独立
  const username = `user_${Date.now()}`;
  // 初始密码
  const password = 'Test1234';
  // 修改后的新密码
  const newPassword = 'Test5678';

  test('完整认证流程', async ({ page, browser, loginPage, registerPage, dashboardPage, profilePage, passwordPage, header }) => {
    // ═══════════════════════════════════════════════════════════════
    // A-E2E-01: 注册新用户
    // 验证点：注册成功后自动跳转到登录页
    // ═══════════════════════════════════════════════════════════════
    await test.step('A-E2E-01: 注册新用户', async () => {
      // 访问注册页面
      await registerPage.goto();
      // 填写用户名、密码、确认密码并提交
      await registerPage.register(username, password);
      // 验证跳转到登录页
      await loginPage.expectLoaded();
    });

    // ═══════════════════════════════════════════════════════════════
    // A-E2E-02: 登录
    // 验证点：登录成功后进入 dashboard 页面
    // ═══════════════════════════════════════════════════════════════
    await test.step('A-E2E-02: 登录', async () => {
      // 使用刚注册的账号登录
      await loginPage.login(username, password);
      // 验证 URL 包含 /dashboard
      await expect(page).toHaveURL(/\/dashboard/);
    });

    // ═══════════════════════════════════════════════════════════════
    // A-E2E-03: 查看个人资料
    // 验证点：个人资料页显示正确的用户名、用户ID、注册时间
    // ═══════════════════════════════════════════════════════════════
    await test.step('A-E2E-03: 查看个人资料', async () => {
      // 通过 Header 跳转到个人资料页
      await header.gotoProfile();
      await profilePage.expectLoaded();

      // 获取页面显示的用户信息
      const displayedUsername = await profilePage.getUsername();
      const userId = await profilePage.getUserId();
      const createdAt = await profilePage.getCreatedAt();

      // 验证用户名与注册时一致
      expect(displayedUsername).toBe(username);
      // 验证用户ID存在
      expect(userId).toBeTruthy();
      // 验证注册时间存在
      expect(createdAt).toBeTruthy();
    });

    // ═══════════════════════════════════════════════════════════════
    // A-E2E-04: 修改密码
    // 验证点：密码修改成功，提示成功并跳转到登录页
    // ═══════════════════════════════════════════════════════════════
    await test.step('A-E2E-04: 修改密码', async () => {
      // 通过 Header 跳转到修改密码页
      await header.gotoPassword();
      await passwordPage.expectLoaded();
      // 填写旧密码和新密码并提交
      await passwordPage.changePassword(password, newPassword);
      // 验证跳转到登录页（修改密码后需要重新登录）
      await loginPage.expectLoaded();
    });

    // ═══════════════════════════════════════════════════════════════
    // A-E2E-05: 新密码登录
    // 验证点：使用新密码可以成功登录
    // ═══════════════════════════════════════════════════════════════
    await test.step('A-E2E-05: 新密码登录', async () => {
      // 使用新密码登录
      await loginPage.login(username, newPassword);
      // 验证成功进入 dashboard
      await expect(page).toHaveURL(/\/dashboard/);
    });

    // ═══════════════════════════════════════════════════════════════
    // A-E2E-06: 登录持久化
    // 验证点：关闭浏览器后重新访问，登录状态仍然有效
    // ═══════════════════════════════════════════════════════════════
    await test.step('A-E2E-06: 登录持久化', async () => {
      // 保存当前登录状态（cookies + localStorage）
      const storageState = await page.context().storageState();
      // 关闭当前浏览器上下文
      await page.context().close();

      // 创建新的浏览器上下文，加载保存的登录状态
      const newContext = await browser.newContext({ storageState });
      const newPage = await newContext.newPage();
      // 直接访问 dashboard（不需要重新登录）
      await newPage.goto('/dashboard/index');
      await newPage.waitForURL(/\/dashboard/, { timeout: 10000 });
      // 验证仍然在 dashboard 页面
      expect(newPage.url()).toContain('/dashboard');
      await newPage.close();
    });

    // ═══════════════════════════════════════════════════════════════
    // A-E2E-07: 退出登录
    // 验证点：退出成功后跳转到登录页
    // ═══════════════════════════════════════════════════════════════
    await test.step('A-E2E-07: 退出登录', async () => {
      // 创建新的浏览器上下文
      const newContext = await browser.newContext();
      const newPage = await newContext.newPage();
      // 先访问登录页
      await newPage.goto('/auth/login');
      await newPage.waitForURL(/\/auth\/login/, { timeout: 10000 });

      // 登录
      const newLoginPage = new LoginPage(newPage);
      await newLoginPage.login(username, newPassword);

      // 使用 Header 组件执行退出登录
      const newHeader = new Header(newPage);
      await newHeader.logout();

      // 验证跳转到登录页
      await newPage.waitForURL(/\/auth\/login/, { timeout: 10000 });
      await newPage.close();
    });

    // ═══════════════════════════════════════════════════════════════
    // A-E2E-08: 未登录访问受保护页面
    // 验证点：未登录状态下访问 dashboard 会被重定向到登录页
    // ═══════════════════════════════════════════════════════════════
    await test.step('A-E2E-08: 未登录访问受保护页面', async () => {
      // 创建新的浏览器上下文（无登录状态）
      const newContext = await browser.newContext();
      const newPage = await newContext.newPage();
      // 尝试直接访问 dashboard
      await newPage.goto('/dashboard/index');
      // 验证被重定向到登录页
      await newPage.waitForURL(/\/auth\/login/, { timeout: 10000 });
      expect(newPage.url()).toContain('/auth/login');
      await newPage.close();
    });
  });
});
