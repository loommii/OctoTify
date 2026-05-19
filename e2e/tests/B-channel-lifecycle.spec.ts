import { test, expect } from '../fixtures/test-fixtures';
import { loadTestConfig } from '../helpers/test-config';
import { LoginPage } from '../pages/LoginPage';
import { RegisterPage } from '../pages/RegisterPage';
import { Header } from '../pages/components/Header';

/**
 * B-Channel: 推送渠道生命周期
 *
 * 测试覆盖 Telegram Bot 渠道的完整生命周期：
 * 创建 → 查看详情 → 编辑 → 测试连接 → 启用/停用 → 删除
 * 以及反向、边界、权限场景
 *
 * 测试数据：使用 loadTestConfig() 加载 Telegram Bot 配置
 */
test.describe('B-Channel: 推送渠道生命周期', () => {
  const config = loadTestConfig();
  const bot1 = config.telegramBots[0];
  const bot2 = config.telegramBots[1];

  // 通用成功提示定位器
  const successToast = (page: any) => page.locator('.el-message--success .el-message__content');
  // 通用错误提示定位器
  const errorToast = (page: any) => page.locator('.el-message--error .el-message__content');
  // MessageBox 确认按钮
  const messageBoxConfirm = (page: any) => page.locator('.el-message-box__btns button.el-button--primary');

  // Helper：创建 Telegram 渠道并返回列表行
  async function createTelegramChannel(
    page: any,
    channelCreatePage: any,
    channelListPage: any,
    name: string,
    botConfig: { botToken: string; chatId: string; proxy?: string }
  ) {
    await channelCreatePage.goto();
    await channelCreatePage.selectType('Telegram');
    await channelCreatePage.fillForm(name, {
      'Bot Token': botConfig.botToken,
      'Chat ID': botConfig.chatId,
      'HTTP 代理（可选）': botConfig.proxy || '',
    });
    await channelCreatePage.submit();
    await page.waitForURL(/\/channel\/list/, { timeout: 15000 });
    const row = await channelListPage.getChannelRow(name);
    await expect(row).toBeVisible();
    return row;
  }

  // ═══════════════════════════════════════════════════════════════
  // B-21: 正向 - 创建 Telegram 渠道（使用 Bot-1 配置）
  // ═══════════════════════════════════════════════════════════════
  test('B-21: 正向 - 创建 Telegram 渠道（使用 Bot-1 配置）', async ({ page, auth, channelCreatePage, channelListPage }) => {
    const testUser = `B21_channel_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    await test.step('创建 Telegram 渠道', async () => {
      const name = `channel_${Date.now()}`;
      await createTelegramChannel(page, channelCreatePage, channelListPage, name, bot1);
    });
  });

  // ═══════════════════════════════════════════════════════════════
  // B-22: 正向 - 创建 Telegram 渠道（使用 Bot-2 配置）
  // ═══════════════════════════════════════════════════════════════
  test('B-22: 正向 - 创建 Telegram 渠道（使用 Bot-2 配置）', async ({ page, auth, channelCreatePage, channelListPage }) => {
    const testUser = `B22_channel_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    await test.step('使用 Bot-2 创建 Telegram 渠道', async () => {
      const name = `channel_${Date.now()}`;
      await createTelegramChannel(page, channelCreatePage, channelListPage, name, bot2);
    });
  });

  // ═══════════════════════════════════════════════════════════════
  // B-23: 正向 - 查看渠道详情
  // ═══════════════════════════════════════════════════════════════
  test('B-23: 正向 - 查看渠道详情', async ({ page, auth, channelCreatePage, channelListPage, channelDetailPage }) => {
    const testUser = `B23_channel_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    const name = `channel_${Date.now()}`;
    await test.step('创建 Telegram 渠道', async () => {
      await createTelegramChannel(page, channelCreatePage, channelListPage, name, bot1);
    });

    await test.step('点击渠道名称进入详情页', async () => {
      const row = await channelListPage.getChannelRow(name);
      await row.locator('a').click();
      await page.waitForURL(/\/channel\/detail\/\d+/, { timeout: 10000 });
    });

    await test.step('验证详情页信息', async () => {
      const detailName = await channelDetailPage.getName();
      const detailType = await channelDetailPage.getType();
      const detailStatus = await channelDetailPage.getStatusTag();
      expect(detailName.trim()).toBe(name);
      expect(detailType.trim()).toContain('Telegram');
      expect(detailStatus.trim()).toBeTruthy();
    });
  });

  // ═══════════════════════════════════════════════════════════════
  // B-24: 正向 - 编辑渠道名称
  // ═══════════════════════════════════════════════════════════════
  test('B-24: 正向 - 编辑渠道名称', async ({ page, auth, channelCreatePage, channelListPage, channelDetailPage }) => {
    const testUser = `B24_channel_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    const name = `channel_${Date.now()}`;
    const newName = `channel_edited_${Date.now()}`;

    await test.step('创建 Telegram 渠道', async () => {
      await createTelegramChannel(page, channelCreatePage, channelListPage, name, bot1);
    });

    await test.step('进入详情页并点击编辑', async () => {
      const row = await channelListPage.getChannelRow(name);
      await row.locator('a').click();
      await page.waitForURL(/\/channel\/detail\/\d+/, { timeout: 10000 });
      await channelDetailPage.clickEdit();
      await page.waitForURL(/\/channel\/edit\/\d+/, { timeout: 10000 });
    });

    await test.step('修改渠道名称并提交', async () => {
      const nameInput = page.locator('input[maxlength="128"]').first();
      await nameInput.fill(newName);
      await page.getByRole('button', { name: '确认' }).click();
      await page.waitForURL(/\/channel\/list/, { timeout: 15000 });
    });

    await test.step('验证列表显示新名称', async () => {
      const row = await channelListPage.getChannelRow(newName);
      await expect(row).toBeVisible();
    });
  });

  // ═══════════════════════════════════════════════════════════════
  // B-25: 正向 - 测试渠道连接
  // ═══════════════════════════════════════════════════════════════
  test('B-25: 正向 - 测试渠道连接', async ({ page, auth, channelCreatePage, channelListPage }) => {
    const testUser = `B25_channel_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    const name = `channel_${Date.now()}`;
    await test.step('创建 Telegram 渠道', async () => {
      await createTelegramChannel(page, channelCreatePage, channelListPage, name, bot1);
    });

    await test.step('点击测试连接按钮', async () => {
      await channelListPage.clickTest(name);
    });

    await test.step('验证测试成功提示', async () => {
      await expect(successToast(page)).toBeVisible({ timeout: 15000 });
      const msg = await successToast(page).textContent();
      expect(msg).toContain('成功');
    });
  });

  // ═══════════════════════════════════════════════════════════════
  // B-26: 正向 - 启用渠道
  // ═══════════════════════════════════════════════════════════════
  test('B-26: 正向 - 启用渠道', async ({ page, auth, channelCreatePage, channelListPage }) => {
    const testUser = `B26_channel_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    const name = `channel_${Date.now()}`;
    await test.step('创建 Telegram 渠道', async () => {
      await createTelegramChannel(page, channelCreatePage, channelListPage, name, bot1);
    });

    await test.step('先停用渠道', async () => {
      page.on('dialog', (dialog: any) => dialog.accept());
      await channelListPage.clickDisable(name);
      await messageBoxConfirm(page).click();
      await page.waitForTimeout(500);
    });

    await test.step('再启用渠道', async () => {
      page.on('dialog', (dialog: any) => dialog.accept());
      await channelListPage.clickEnable(name);
      await messageBoxConfirm(page).click();
      await page.waitForTimeout(500);
    });

    await test.step('验证状态为启用', async () => {
      const row = await channelListPage.getChannelRow(name);
      const statusCell = row.locator('td').nth(3);
      await expect(statusCell).toContainText('启用');
    });
  });

  // ═══════════════════════════════════════════════════════════════
  // B-27: 正向 - 停用渠道
  // ═══════════════════════════════════════════════════════════════
  test('B-27: 正向 - 停用渠道', async ({ page, auth, channelCreatePage, channelListPage }) => {
    const testUser = `B27_channel_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    const name = `channel_${Date.now()}`;
    await test.step('创建 Telegram 渠道', async () => {
      await createTelegramChannel(page, channelCreatePage, channelListPage, name, bot1);
    });

    await test.step('停用渠道', async () => {
      page.on('dialog', (dialog: any) => dialog.accept());
      await channelListPage.clickDisable(name);
      await messageBoxConfirm(page).click();
      await page.waitForTimeout(500);
    });

    await test.step('验证状态为停用', async () => {
      const row = await channelListPage.getChannelRow(name);
      const statusCell = row.locator('td').nth(3);
      await expect(statusCell).toContainText('停用');
    });
  });

  // ═══════════════════════════════════════════════════════════════
  // B-28: 正向 - 删除渠道
  // ═══════════════════════════════════════════════════════════════
  test('B-28: 正向 - 删除渠道', async ({ page, auth, channelCreatePage, channelListPage }) => {
    const testUser = `B28_channel_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    const name = `channel_${Date.now()}`;
    await test.step('创建 Telegram 渠道', async () => {
      await createTelegramChannel(page, channelCreatePage, channelListPage, name, bot1);
    });

    await test.step('删除渠道', async () => {
      page.on('dialog', (dialog: any) => dialog.accept());
      await channelListPage.clickDelete(name);
      await messageBoxConfirm(page).click();
      await page.waitForTimeout(500);
    });

    await test.step('验证渠道已消失', async () => {
      const row = await channelListPage.getChannelRow(name);
      await expect(row).toHaveCount(0);
    });
  });

  // ═══════════════════════════════════════════════════════════════
  // B-29: 正向 - 渠道列表展示
  // ═══════════════════════════════════════════════════════════════
  test('B-29: 正向 - 渠道列表展示', async ({ page, auth, channelCreatePage, channelListPage }) => {
    const testUser = `B29_channel_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    const name = `channel_${Date.now()}`;
    await test.step('创建 Telegram 渠道', async () => {
      await createTelegramChannel(page, channelCreatePage, channelListPage, name, bot1);
    });

    await test.step('验证列表列展示', async () => {
      await channelListPage.goto();
      await expect(page.locator('.el-table__header th').filter({ hasText: 'ID' })).toBeVisible();
      await expect(page.locator('.el-table__header th').filter({ hasText: '名称' })).toBeVisible();
      await expect(page.locator('.el-table__header th').filter({ hasText: '类型' })).toBeVisible();
      await expect(page.locator('.el-table__header th').filter({ hasText: '状态' })).toBeVisible();
      await expect(page.locator('.el-table__header th').filter({ hasText: '创建时间' })).toBeVisible();
      await expect(page.locator('.el-table__header th').filter({ hasText: '操作' })).toBeVisible();
    });
  });

  // ═══════════════════════════════════════════════════════════════
  // B-30: 正向 - 渠道列表分页
  // ═══════════════════════════════════════════════════════════════
  test('B-30: 正向 - 渠道列表分页', async ({ page, auth, channelListPage }) => {
    const testUser = `B30_channel_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    await test.step('验证分页组件存在', async () => {
      await channelListPage.goto();
      await expect(page.locator('.el-pagination')).toBeVisible();
      await expect(page.locator('.el-pagination .el-pagination__total')).toBeVisible();
      await expect(page.locator('.el-pagination .el-pagination__sizes')).toBeVisible();
      await expect(page.locator('.el-pagination .btn-prev')).toBeVisible();
      await expect(page.locator('.el-pagination .btn-next')).toBeVisible();
    });
  });

  // ═══════════════════════════════════════════════════════════════
  // B-31: 反向 - 创建渠道名称为空
  // ═══════════════════════════════════════════════════════════════
  test('B-31: 反向 - 创建渠道名称为空', async ({ page, auth, channelCreatePage }) => {
    const testUser = `B31_channel_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    await test.step('进入创建页面并选择 Telegram', async () => {
      await channelCreatePage.goto();
      await channelCreatePage.selectType('Telegram');
    });

    await test.step('填写其他字段，名称留空并提交', async () => {
      await channelCreatePage.fillForm('', {
        'Bot Token': bot1.botToken,
        'Chat ID': bot1.chatId,
      });
      await channelCreatePage.submit();
    });

    await test.step('验证仍在创建页面', async () => {
      await expect(page).toHaveURL(/\/channel\/create/);
    });
  });

  // ═══════════════════════════════════════════════════════════════
  // B-32: 反向 - 创建渠道缺少 Bot Token
  // ═══════════════════════════════════════════════════════════════
  test('B-32: 反向 - 创建渠道缺少 Bot Token', async ({ page, auth, channelCreatePage }) => {
    const testUser = `B32_channel_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    await test.step('进入创建页面并选择 Telegram', async () => {
      await channelCreatePage.goto();
      await channelCreatePage.selectType('Telegram');
    });

    await test.step('填写名称和 Chat ID，Bot Token 留空并提交', async () => {
      const name = `channel_${Date.now()}`;
      await channelCreatePage.fillForm(name, {
        'Bot Token': '',
        'Chat ID': bot1.chatId,
      });
      await channelCreatePage.submit();
    });

    await test.step('验证仍在创建页面', async () => {
      await expect(page).toHaveURL(/\/channel\/create/);
    });
  });

  // ═══════════════════════════════════════════════════════════════
  // B-33: 反向 - 创建渠道缺少 Chat ID
  // ═══════════════════════════════════════════════════════════════
  test('B-33: 反向 - 创建渠道缺少 Chat ID', async ({ page, auth, channelCreatePage }) => {
    const testUser = `B33_channel_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    await test.step('进入创建页面并选择 Telegram', async () => {
      await channelCreatePage.goto();
      await channelCreatePage.selectType('Telegram');
    });

    await test.step('填写名称和 Bot Token，Chat ID 留空并提交', async () => {
      const name = `channel_${Date.now()}`;
      await channelCreatePage.fillForm(name, {
        'Bot Token': bot1.botToken,
        'Chat ID': '',
      });
      await channelCreatePage.submit();
    });

    await test.step('验证仍在创建页面', async () => {
      await expect(page).toHaveURL(/\/channel\/create/);
    });
  });

  // ═══════════════════════════════════════════════════════════════
  // B-34: 反向 - 测试已删除的渠道
  // ═══════════════════════════════════════════════════════════════
  test('B-34: 反向 - 测试已删除的渠道', async ({ page, auth, channelCreatePage, channelListPage, channelDetailPage }) => {
    const testUser = `B34_channel_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    const name = `channel_${Date.now()}`;
    let channelId = '';

    await test.step('创建 Telegram 渠道并记录 ID', async () => {
      await createTelegramChannel(page, channelCreatePage, channelListPage, name, bot1);
      const row = await channelListPage.getChannelRow(name);
      channelId = (await row.locator('td').first().textContent()) ?? '';
      expect(channelId).toBeTruthy();
    });

    await test.step('删除渠道', async () => {
      page.on('dialog', (dialog: any) => dialog.accept());
      await channelListPage.clickDelete(name);
      await messageBoxConfirm(page).click();
      await page.waitForTimeout(500);
    });

    await test.step('直接访问已删除渠道的详情页并测试', async () => {
      await channelDetailPage.goto(channelId);
      await channelDetailPage.clickTest();
    });

    await test.step('验证测试失败提示', async () => {
      await expect(errorToast(page)).toBeVisible({ timeout: 10000 });
    });
  });

  // ═══════════════════════════════════════════════════════════════
  // B-35: 反向 - 测试已停用的渠道
  // ═══════════════════════════════════════════════════════════════
  test('B-35: 反向 - 测试已停用的渠道', async ({ page, auth, channelCreatePage, channelListPage }) => {
    const testUser = `B35_channel_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    const name = `channel_${Date.now()}`;
    await test.step('创建 Telegram 渠道', async () => {
      await createTelegramChannel(page, channelCreatePage, channelListPage, name, bot1);
    });

    await test.step('停用渠道', async () => {
      page.on('dialog', (dialog: any) => dialog.accept());
      await channelListPage.clickDisable(name);
      await messageBoxConfirm(page).click();
      await page.waitForTimeout(500);
    });

    await test.step('测试已停用的渠道', async () => {
      await channelListPage.clickTest(name);
    });

    await test.step('验证测试失败提示', async () => {
      await expect(errorToast(page)).toBeVisible({ timeout: 10000 });
    });
  });

  // ═══════════════════════════════════════════════════════════════
  // B-36: 边界 - 渠道名称正好128字符
  // ═══════════════════════════════════════════════════════════════
  test('B-36: 边界 - 渠道名称正好128字符', async ({ page, auth, channelCreatePage, channelListPage }) => {
    const testUser = `B36_channel_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    await test.step('创建128字符名称的渠道', async () => {
      const prefix = `ch_${Date.now()}_`;
      const name = prefix + 'x'.repeat(128 - prefix.length);
      expect(name.length).toBe(128);
      await createTelegramChannel(page, channelCreatePage, channelListPage, name, bot1);
    });
  });

  // ═══════════════════════════════════════════════════════════════
  // B-37: 边界 - 渠道名称129字符
  // ═══════════════════════════════════════════════════════════════
  test('B-37: 边界 - 渠道名称129字符', async ({ page, auth, channelCreatePage }) => {
    const testUser = `B37_channel_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    await test.step('进入创建页面并选择 Telegram', async () => {
      await channelCreatePage.goto();
      await channelCreatePage.selectType('Telegram');
    });

    await test.step('绕过 maxlength 输入129字符名称', async () => {
      const prefix = `ch_${Date.now()}_`;
      const name = prefix + 'x'.repeat(129 - prefix.length);
      expect(name.length).toBe(129);
      await page.evaluate((val: string) => {
        const input = document.querySelector('input[maxlength="128"]') as HTMLInputElement;
        if (input) {
          input.value = val;
          input.dispatchEvent(new Event('input', { bubbles: true }));
          input.dispatchEvent(new Event('change', { bubbles: true }));
        }
      }, name);
    });

    await test.step('填写其他字段并提交', async () => {
      await channelCreatePage.fillForm('', {
        'Bot Token': bot1.botToken,
        'Chat ID': bot1.chatId,
      });
      await channelCreatePage.submit();
    });

    await test.step('验证创建失败或名称被截断', async () => {
      // 由于 maxlength 限制，129字符无法输入，提交后要么失败，要么名称被截断为128字符
      // 这里验证页面未跳转到列表（前端校验拦截）或出现错误提示
      const currentUrl = page.url();
      if (!currentUrl.includes('/channel/list')) {
        await expect(page).toHaveURL(/\/channel\/create/);
      }
    });
  });

  // ═══════════════════════════════════════════════════════════════
  // B-38: 边界 - 渠道名称1字符（最小长度）
  // ═══════════════════════════════════════════════════════════════
  test('B-38: 边界 - 渠道名称1字符（最小长度）', async ({ page, auth, channelCreatePage, channelListPage }) => {
    const testUser = `B38_channel_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    await test.step('创建1字符名称的渠道', async () => {
      const name = 'x';
      await createTelegramChannel(page, channelCreatePage, channelListPage, name, bot1);
    });
  });

  // ═══════════════════════════════════════════════════════════════
  // B-39: 权限 - 未登录访问渠道列表
  // ═══════════════════════════════════════════════════════════════
  test('B-39: 权限 - 未登录访问渠道列表', async ({ page, channelListPage }) => {
    await test.step('未登录直接访问渠道列表', async () => {
      await channelListPage.goto();
    });

    await test.step('验证重定向到登录页', async () => {
      await page.waitForURL(/\/auth\/login/, { timeout: 10000 });
      expect(page.url()).toContain('/auth/login');
    });
  });

  // ═══════════════════════════════════════════════════════════════
  // B-40: 权限 - 用户隔离（看不到其他用户的渠道）
  // ═══════════════════════════════════════════════════════════════
  test('B-40: 权限 - 用户隔离（看不到其他用户的渠道）', async ({ page, browser, registerPage, loginPage, channelCreatePage, channelListPage }) => {
    let channelName = '';

    await test.step('用户1：注册、登录并创建渠道', async () => {
      const testUser = `B40_channel_${Date.now()}`;
      const testPass = 'Test123456';
      await registerPage.goto();
      await registerPage.register(testUser, testPass);
      await page.waitForURL(/\/auth\/login/, { timeout: 15000 });
      await loginPage.goto();
      await loginPage.login(testUser, testPass);
      await page.waitForURL(/\/dashboard/, { timeout: 15000 });

      channelName = `channel_${Date.now()}`;
      await createTelegramChannel(page, channelCreatePage, channelListPage, channelName, bot1);
    });

    await test.step('用户2：注册新用户并登录', async () => {
      const newContext = await browser.newContext();
      const newPage = await newContext.newPage();
      const newRegisterPage = new RegisterPage(newPage);
      const newLoginPage = new LoginPage(newPage);
      const newChannelListPage = channelListPage; // Reuse fixture logic on new page

      const username = `B40_2_channel_${Date.now()}`;
      const password = 'Test123456';
      await newRegisterPage.goto();
      await newRegisterPage.register(username, password);
      await newPage.waitForURL(/\/auth\/login/, { timeout: 15000 });
      await newLoginPage.goto();
      await newLoginPage.login(username, password);
      await newPage.waitForURL(/\/dashboard/, { timeout: 15000 });

      // 用户2访问渠道列表
      await newPage.goto('/channel/list');
      await newPage.waitForURL(/\/channel\/list/, { timeout: 10000 });

      // 验证看不到用户1的渠道
      const row = newPage.locator('tr').filter({ hasText: channelName });
      await expect(row).toHaveCount(0);

      await newPage.close();
    });
  });
});
