import { test, expect } from '../fixtures/test-fixtures';
import { loadTestConfig } from '../helpers/test-config';

/**
 * B-E2E: 完整业务流程
 *
 * 测试覆盖核心业务生命周期：
 * 注册 → 登录 → 创建渠道 → 创建来源并绑定渠道 → 仪表盘验证 → 退出登录
 *
 * 测试数据：使用随机用户名避免并发冲突，Telegram Bot 配置从 test.config.toml 加载
 */
test.describe('B-E2E: 完整业务流程', () => {
  const config = loadTestConfig();
  const bot = config.telegramBots[0];

  test('完整业务流程', async ({ page, auth, channelListPage, channelCreatePage, sourceListPage, sourceCreatePage, dashboardPage, header }) => {
    const username = `e2e_${Date.now()}`;
    const password = 'Test123456';
    const channelName = `ch_e2e_${Date.now()}`;
    const sourceName = `src_e2e_${Date.now()}`;

    // ═══════════════════════════════════════════════════════════════
    // B-E2E-01: 注册并登录
    // 验证点：注册成功后自动跳转到 dashboard
    // ═══════════════════════════════════════════════════════════════
    await test.step('B-E2E-01: 注册并登录', async () => {
      await auth.registerAndLogin(username, password);
      await expect(page).toHaveURL(/\/dashboard/);
    });

    // ═══════════════════════════════════════════════════════════════
    // B-E2E-02: 创建 Telegram 渠道
    // 验证点：创建成功后跳转到渠道列表，新渠道可见
    // ═══════════════════════════════════════════════════════════════
    await test.step('B-E2E-02: 创建 Telegram 渠道', async () => {
      await channelListPage.goto();
      await channelListPage.clickCreate();
      await channelCreatePage.selectType('Telegram');
      await channelCreatePage.fillForm(channelName, {
        'Bot Token': bot.botToken,
        'Chat ID': bot.chatId,
        'HTTP 代理（可选）': bot.proxy || '',
      });
      await channelCreatePage.submit();
      await page.waitForURL(/\/channel\/list/, { timeout: 15000 });

      const row = await channelListPage.getChannelRow(channelName);
      await expect(row).toBeVisible();
    });

    // ═══════════════════════════════════════════════════════════════
    // B-E2E-03: 创建来源并绑定渠道
    // 验证点：创建成功后跳转到来源列表，新来源可见
    // ═══════════════════════════════════════════════════════════════
    await test.step('B-E2E-03: 创建来源并绑定渠道', async () => {
      await sourceListPage.goto();
      await sourceListPage.clickCreate();
      await sourceCreatePage.fillForm(sourceName, 'E2E 测试来源', [channelName]);
      await sourceCreatePage.submit();
      await page.waitForURL(/\/source\/list/, { timeout: 15000 });

      const row = await sourceListPage.getSourceRow(sourceName);
      await expect(row).toBeVisible();
    });

    // ═══════════════════════════════════════════════════════════════
    // B-E2E-04: 仪表盘验证
    // 验证点：仪表盘加载成功，统计数据显示来源和渠道数量 >= 1
    // ═══════════════════════════════════════════════════════════════
    await test.step('B-E2E-04: 仪表盘验证', async () => {
      await dashboardPage.goto();
      const loaded = await dashboardPage.isLoaded();
      expect(loaded).toBe(true);

      await expect.poll(
        async () => Number(await dashboardPage.getMessageSourceCount()),
        { message: '等待消息来源总数更新', timeout: 15000 },
      ).toBeGreaterThanOrEqual(1);

      await expect.poll(
        async () => Number(await dashboardPage.getChannelCount()),
        { message: '等待推送渠道总数更新', timeout: 15000 },
      ).toBeGreaterThanOrEqual(1);
    });

    // ═══════════════════════════════════════════════════════════════
    // B-E2E-05: 退出登录
    // 验证点：退出成功后跳转到登录页
    // ═══════════════════════════════════════════════════════════════
    await test.step('B-E2E-05: 退出登录', async () => {
      await header.logout();
      await page.waitForURL(/\/auth\/login/, { timeout: 10000 });
      expect(page.url()).toContain('/auth/login');
    });
  });
});
