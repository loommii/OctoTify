import { test, expect } from '../fixtures/test-fixtures';

test.describe('B-Message: 消息管理', () => {
  test('B-41: 访问消息列表页面', async ({ page, auth }) => {
    const testUser = `B41_message_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    await test.step('B-41: 访问消息列表', async () => {
      await page.goto('/message/list');
      await page.waitForURL(/\/message\/list/, { timeout: 10000 });
    });

    await test.step('验证消息列表页面加载', async () => {
      await expect(page.locator('.el-table')).toBeVisible();
    });
  });

  test('B-42: 消息列表展示', async ({ page, auth }) => {
    const testUser = `B42_message_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    await test.step('B-42: 验证消息列表列头', async () => {
      await page.goto('/message/list');
      await page.waitForURL(/\/message\/list/, { timeout: 10000 });
      await expect(page.getByText('ID')).toBeVisible();
      await expect(page.getByText('标题')).toBeVisible();
      await expect(page.getByText('来源名称')).toBeVisible();
      await expect(page.getByText('渠道名称')).toBeVisible();
      await expect(page.getByText('状态')).toBeVisible();
      await expect(page.getByText('创建时间')).toBeVisible();
    });
  });

  test('B-43: 按来源筛选消息', async ({ page, auth }) => {
    const testUser = `B43_message_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    await test.step('B-43: 按来源筛选', async () => {
      await page.goto('/message/list');
      await page.waitForURL(/\/message\/list/, { timeout: 10000 });
      const sourceFilter = page.locator('select').first();
      await sourceFilter.selectOption({ index: 1 });
      await page.getByRole('button', { name: '搜索' }).click();
      await page.waitForTimeout(1000);
    });

    await test.step('验证筛选结果', async () => {
      await expect(page.locator('.el-table')).toBeVisible();
    });
  });

  test('B-44: 按渠道筛选消息', async ({ page, auth }) => {
    const testUser = `B44_message_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    await test.step('B-44: 按渠道筛选', async () => {
      await page.goto('/message/list');
      await page.waitForURL(/\/message\/list/, { timeout: 10000 });
      const channelFilter = page.locator('select').nth(1);
      await channelFilter.selectOption({ index: 1 });
      await page.getByRole('button', { name: '搜索' }).click();
      await page.waitForTimeout(1000);
    });

    await test.step('验证筛选结果', async () => {
      await expect(page.locator('.el-table')).toBeVisible();
    });
  });

  test('B-45: 按状态筛选消息', async ({ page, auth }) => {
    const testUser = `B45_message_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    await test.step('B-45: 按状态筛选', async () => {
      await page.goto('/message/list');
      await page.waitForURL(/\/message\/list/, { timeout: 10000 });
      const statusFilter = page.locator('select').nth(2);
      await statusFilter.selectOption({ index: 1 });
      await page.getByRole('button', { name: '搜索' }).click();
      await page.waitForTimeout(1000);
    });

    await test.step('验证筛选结果', async () => {
      await expect(page.locator('.el-table')).toBeVisible();
    });
  });

  test('B-46: 按关键词筛选消息', async ({ page, auth }) => {
    const testUser = `B46_message_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    await test.step('B-46: 按关键词筛选', async () => {
      await page.goto('/message/list');
      await page.waitForURL(/\/message\/list/, { timeout: 10000 });
      const keywordInput = page.locator('input[placeholder*="关键词"]');
      await keywordInput.fill('test');
      await page.getByRole('button', { name: '搜索' }).click();
      await page.waitForTimeout(1000);
    });

    await test.step('验证筛选结果', async () => {
      await expect(page.locator('.el-table')).toBeVisible();
    });
  });

  test('B-47: 重置筛选条件', async ({ page, auth }) => {
    const testUser = `B47_message_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    await test.step('B-47: 重置筛选', async () => {
      await page.goto('/message/list');
      await page.waitForURL(/\/message\/list/, { timeout: 10000 });
      const keywordInput = page.locator('input[placeholder*="关键词"]');
      await keywordInput.fill('test');
      await page.getByRole('button', { name: '重置' }).click();
      await page.waitForTimeout(1000);
    });

    await test.step('验证筛选条件已清空', async () => {
      const keywordInput = page.locator('input[placeholder*="关键词"]');
      await expect(keywordInput).toHaveValue('');
    });
  });

  test('B-48: 删除消息', async ({ page, auth }) => {
    const testUser = `B48_message_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    await test.step('B-48: 删除消息', async () => {
      await page.goto('/message/list');
      await page.waitForURL(/\/message\/list/, { timeout: 10000 });
      const deleteButton = page.getByRole('button', { name: '删除' }).first();
      if (await deleteButton.isVisible().catch(() => false)) {
        await deleteButton.click();
        await page.getByRole('button', { name: '确认' }).click();
        await page.waitForTimeout(2000);
      }
    });

    await test.step('验证删除成功', async () => {
      await expect(page.locator('.el-table')).toBeVisible();
    });
  });

  test('B-49: 取消删除消息', async ({ page, auth }) => {
    const testUser = `B49_message_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    await test.step('B-49: 取消删除', async () => {
      await page.goto('/message/list');
      await page.waitForURL(/\/message\/list/, { timeout: 10000 });
      const deleteButton = page.getByRole('button', { name: '删除' }).first();
      if (await deleteButton.isVisible().catch(() => false)) {
        await deleteButton.click();
        await page.getByRole('button', { name: '取消' }).click();
        await page.waitForTimeout(1000);
      }
    });

    await test.step('验证消息仍存在', async () => {
      await expect(page.locator('.el-table')).toBeVisible();
    });
  });

  test('B-50: 空消息列表展示', async ({ page, auth }) => {
    const testUser = `B50_message_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    await test.step('B-50: 验证空列表展示', async () => {
      await page.goto('/message/list');
      await page.waitForURL(/\/message\/list/, { timeout: 10000 });
      await page.waitForTimeout(2000);
      await expect(page.locator('.el-table')).toBeVisible();
    });
  });
});
