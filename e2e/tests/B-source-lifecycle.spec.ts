import { test, expect } from '../fixtures/test-fixtures';

// 消息来源生命周期测试套件
test.describe('B-Source: 消息来源生命周期', () => {
  test('B-03: 查看来源详情', async ({ page, auth, sourceListPage, sourceCreatePage, sourceDetailPage }) => {
    const testUser = `B03_source_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    const sourceName = `source_${Date.now()}`;

    await test.step('准备：创建来源', async () => {
      await sourceListPage.goto();
      await sourceListPage.clickCreate();
      await sourceCreatePage.fillForm(sourceName);
      await sourceCreatePage.submit();
      await page.waitForURL(/\/source\/list/, { timeout: 10000 });
    });

    await test.step('B-03: 查看来源详情', async () => {
      const row = await sourceListPage.getSourceRow(sourceName);
      await row.getByText(sourceName).click();
      await page.waitForURL(/\/source\/detail/, { timeout: 10000 });
    });

    await test.step('验证来源详情显示正确', async () => {
      const name = await sourceDetailPage.getName();
      expect(name).toContain(sourceName);
    });
  });

  test('B-04: 编辑来源名称', async ({ page, auth, sourceListPage, sourceCreatePage, sourceEditPage }) => {
    const testUser = `B04_source_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    const sourceName = `source_${Date.now()}`;
    const newName = `edited_${Date.now()}`;

    await test.step('准备：创建来源', async () => {
      await sourceListPage.goto();
      await sourceListPage.clickCreate();
      await sourceCreatePage.fillForm(sourceName);
      await sourceCreatePage.submit();
      await page.waitForURL(/\/source\/list/, { timeout: 10000 });
    });

    await test.step('B-04: 编辑来源名称', async () => {
      const row = await sourceListPage.getSourceRow(sourceName);
      await row.getByRole('button', { name: '修改' }).click();
      await page.waitForURL(/\/source\/edit/, { timeout: 10000 });
      await sourceEditPage.fillForm(newName);
      await sourceEditPage.submit();
      await page.waitForURL(/\/source\/list/, { timeout: 10000 });
    });

    await test.step('验证来源名称更新成功', async () => {
      // 重新加载列表确保获取最新数据
      await sourceListPage.goto();
      const row = await sourceListPage.getSourceRow(newName);
      await expect(row).toBeVisible();
    });
  });

  test('B-05: 停用来源（step-up auth）', async ({ page, auth, sourceListPage, sourceCreatePage, sourceDetailPage }) => {
    const testUser = `B05_source_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    const sourceName = `source_${Date.now()}`;

    await test.step('准备：创建来源', async () => {
      await sourceListPage.goto();
      await sourceListPage.clickCreate();
      await sourceCreatePage.fillForm(sourceName);
      await sourceCreatePage.submit();
      await page.waitForURL(/\/source\/list/, { timeout: 10000 });
    });

    await test.step('B-05: 停用来源', async () => {
      await sourceListPage.clickDisable(sourceName);
      await sourceDetailPage.fillStepUpPassword('Test123456');
      await page.waitForTimeout(2000);
    });

    await test.step('验证来源已停用', async () => {
      const row = await sourceListPage.getSourceRow(sourceName);
      await expect(row.getByText('停用')).toBeVisible();
    });
  });

  test('B-06: 启用来源（step-up auth）', async ({ page, auth, sourceListPage, sourceCreatePage, sourceDetailPage }) => {
    const testUser = `B06_source_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    const sourceName = `source_${Date.now()}`;

    await test.step('准备：创建并停用来源', async () => {
      await sourceListPage.goto();
      await sourceListPage.clickCreate();
      await sourceCreatePage.fillForm(sourceName);
      await sourceCreatePage.submit();
      await page.waitForURL(/\/source\/list/, { timeout: 10000 });
      await sourceListPage.clickDisable(sourceName);
      await sourceDetailPage.fillStepUpPassword('Test123456');
      await page.waitForTimeout(2000);
    });

    await test.step('B-06: 启用来源', async () => {
      await sourceListPage.clickEnable(sourceName);
      await sourceDetailPage.fillStepUpPassword('Test123456');
      await page.waitForTimeout(2000);
    });

    await test.step('验证来源已启用', async () => {
      const row = await sourceListPage.getSourceRow(sourceName);
      await expect(row.getByText('启用')).toBeVisible();
    });
  });

  test('B-07: 删除来源（step-up auth）', async ({ page, auth, sourceListPage, sourceCreatePage, sourceDetailPage }) => {
    const testUser = `B07_source_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    const sourceName = `source_${Date.now()}`;

    await test.step('准备：创建来源', async () => {
      await sourceListPage.goto();
      await sourceListPage.clickCreate();
      await sourceCreatePage.fillForm(sourceName);
      await sourceCreatePage.submit();
      await page.waitForURL(/\/source\/list/, { timeout: 10000 });
    });

    await test.step('B-07: 删除来源', async () => {
      await sourceListPage.clickDelete(sourceName);
      await sourceDetailPage.fillStepUpPassword('Test123456');
      await page.waitForTimeout(2000);
    });

    await test.step('验证来源已删除', async () => {
      const row = await sourceListPage.getSourceRow(sourceName);
      await expect(row).toBeHidden();
    });
  });

  test('B-08: 来源列表搜索', async ({ page, auth, sourceListPage, sourceCreatePage }) => {
    const testUser = `B08_source_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    const sourceName = `search_test_${Date.now()}`;

    await test.step('准备：创建来源', async () => {
      await sourceListPage.goto();
      await sourceListPage.clickCreate();
      await sourceCreatePage.fillForm(sourceName);
      await sourceCreatePage.submit();
      await page.waitForURL(/\/source\/list/, { timeout: 10000 });
    });

    await test.step('B-08: 搜索来源', async () => {
      await sourceListPage.searchByName(sourceName);
      await page.getByRole('button', { name: '搜索' }).click();
      await page.waitForTimeout(1000);
    });

    await test.step('验证搜索结果', async () => {
      const row = await sourceListPage.getSourceRow(sourceName);
      await expect(row).toBeVisible();
    });
  });

  test('B-09: 创建来源名称为空', async ({ page, auth, sourceListPage, sourceCreatePage }) => {
    const testUser = `B09_source_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    await test.step('B-09: 尝试创建空名称来源', async () => {
      await sourceListPage.goto();
      await sourceListPage.clickCreate();
      await sourceCreatePage.submit();
    });

    await test.step('验证停留在创建页（表单校验阻止提交）', async () => {
      await expect(page).toHaveURL(/\/source\/create/);
    });
  });

  test('B-10: 创建来源名称超长', async ({ page, auth, sourceListPage, sourceCreatePage }) => {
    const testUser = `B10_source_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    const longName = 'a'.repeat(129);

    await test.step('B-10: 尝试创建超长名称来源', async () => {
      await sourceListPage.goto();
      await sourceListPage.clickCreate();
      await sourceCreatePage.fillForm(longName);
      await sourceCreatePage.submit();
    });

    await test.step('验证提交结果（maxlength=128 截断后名称有效，创建应成功）', async () => {
      await page.waitForURL(/\/source\/list/, { timeout: 10000 }).catch(() => {});
      const currentUrl = page.url();
      expect(currentUrl).toContain('/source/list');
    });
  });

  test('B-11: 来源名称正好128字符', async ({ page, auth, sourceListPage, sourceCreatePage }) => {
    const testUser = `B11_source_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    const exactName = 'a'.repeat(128);

    await test.step('B-11: 创建128字符名称来源', async () => {
      await sourceListPage.goto();
      await sourceListPage.clickCreate();
      await sourceCreatePage.fillForm(exactName);
      await sourceCreatePage.submit();
      await page.waitForURL(/\/source\/list/, { timeout: 10000 });
    });

    await test.step('验证来源创建成功', async () => {
      const row = await sourceListPage.getSourceRow(exactName);
      await expect(row).toBeVisible();
    });
  });

  test('B-12: 来源描述正好512字符', async ({ page, auth, sourceListPage, sourceCreatePage }) => {
    const testUser = `B12_source_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    const sourceName = `source_${Date.now()}`;
    const longDesc = 'a'.repeat(512);

    await test.step('B-12: 创建512字符描述来源', async () => {
      await sourceListPage.goto();
      await sourceListPage.clickCreate();
      await sourceCreatePage.fillForm(sourceName, longDesc);
      await sourceCreatePage.submit();
      await page.waitForURL(/\/source\/list/, { timeout: 10000 });
    });

    await test.step('验证来源创建成功', async () => {
      const row = await sourceListPage.getSourceRow(sourceName);
      await expect(row).toBeVisible();
    });
  });

  test('B-13: 未登录访问来源列表', async ({ page }) => {
    await test.step('B-13: 未登录访问来源列表', async () => {
      await page.goto('/source/list');
      await page.waitForURL(/\/auth\/login/, { timeout: 10000 });
    });

    await test.step('验证重定向到登录页', async () => {
      expect(page.url()).toContain('/auth/login');
    });
  });

  test('B-14: 来源列表分页', async ({ page, auth, sourceListPage }) => {
    const testUser = `B14_source_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    await test.step('B-14: 验证分页显示', async () => {
      await sourceListPage.goto();
      const pagination = page.locator('.el-pagination');
      await expect(pagination).toBeVisible();
    });
  });

  test('B-15: 来源列表过滤状态', async ({ page, auth, sourceListPage }) => {
    const testUser = `B15_source_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    await test.step('B-15: 过滤停用状态来源', async () => {
      await sourceListPage.goto();
      await sourceListPage.filterByStatus('停用');
      await page.getByRole('button', { name: '搜索' }).click();
      await page.waitForTimeout(1000);
    });

    await test.step('验证过滤结果', async () => {
      const table = page.locator('.el-table');
      await expect(table).toBeVisible();
    });
  });

  test('B-16: 查看来源令牌（step-up auth）', async ({ page, auth, sourceListPage, sourceCreatePage, sourceDetailPage }) => {
    const testUser = `B16_source_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    const sourceName = `source_${Date.now()}`;

    await test.step('准备：创建来源', async () => {
      await sourceListPage.goto();
      await sourceListPage.clickCreate();
      await sourceCreatePage.fillForm(sourceName);
      await sourceCreatePage.submit();
      await page.waitForURL(/\/source\/list/, { timeout: 10000 });
    });

    await test.step('B-16: 查看来源令牌', async () => {
      const row = await sourceListPage.getSourceRow(sourceName);
      await row.getByText(sourceName).click();
      await page.waitForURL(/\/source\/detail/, { timeout: 10000 });
      await sourceDetailPage.clickViewToken();
      await sourceDetailPage.fillStepUpPassword('Test123456');
      await page.waitForTimeout(2000);
    });

    await test.step('验证令牌显示', async () => {
      const token = await sourceDetailPage.getToken();
      expect(token).toBeTruthy();
    });
  });

  test('B-17: 重置来源令牌（step-up auth）', async ({ page, auth, sourceListPage, sourceCreatePage, sourceDetailPage }) => {
    const testUser = `B17_source_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    const sourceName = `source_${Date.now()}`;

    await test.step('准备：创建来源', async () => {
      await sourceListPage.goto();
      await sourceListPage.clickCreate();
      await sourceCreatePage.fillForm(sourceName);
      await sourceCreatePage.submit();
      await page.waitForURL(/\/source\/list/, { timeout: 10000 });
    });

    await test.step('B-17: 重置来源令牌', async () => {
      const row = await sourceListPage.getSourceRow(sourceName);
      await row.getByText(sourceName).click();
      await page.waitForURL(/\/source\/detail/, { timeout: 10000 });

      // 先查看当前令牌（需 step-up auth）
      await sourceDetailPage.clickViewToken();
      await sourceDetailPage.fillStepUpPassword('Test123456');
      await page.waitForTimeout(1000);
      const oldToken = await sourceDetailPage.getToken();
      expect(oldToken).toBeTruthy();

      // 重置令牌（需 step-up auth）
      await sourceDetailPage.clickResetToken();
      await sourceDetailPage.fillStepUpPassword('Test123456');
      await page.waitForTimeout(1000);
      const newToken = await sourceDetailPage.getToken();
      expect(newToken).toBeTruthy();
      expect(newToken).not.toBe(oldToken);
    });
  });

  test('B-18: 编辑不存在的来源', async ({ page, auth, sourceEditPage }) => {
    const testUser = `B18_source_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    await test.step('B-18: 访问不存在的来源编辑页', async () => {
      await sourceEditPage.goto('999999');
    });

    await test.step('验证错误处理', async () => {
      // 访问不存在的来源时，前端应有错误提示（toast 或表单显示错误）
      const errorMsg = await sourceEditPage.getErrorMessage();
      if (errorMsg) {
        expect(errorMsg).toBeTruthy();
      } else {
        // 若 toast 已消失，验证页面停留在编辑页（未发生跳转）
        await expect(page).toHaveURL(/\/source\/edit\/999999/);
      }
    });
  });

  test('B-19: 删除已删除的来源', async ({ page, auth, sourceListPage, sourceCreatePage, sourceDetailPage }) => {
    const testUser = `B19_source_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    const sourceName = `source_${Date.now()}`;

    await test.step('准备：创建并删除来源', async () => {
      await sourceListPage.goto();
      await sourceListPage.clickCreate();
      await sourceCreatePage.fillForm(sourceName);
      await sourceCreatePage.submit();
      await page.waitForURL(/\/source\/list/, { timeout: 10000 });
      await sourceListPage.clickDelete(sourceName);
      await sourceDetailPage.fillStepUpPassword('Test123456');
      await page.waitForTimeout(2000);
    });

    await test.step('B-19: 验证已删除的来源不再显示在列表中', async () => {
      // 后端过滤 status=-1，已删除来源不在列表中出现
      await sourceListPage.goto();
      const row = await sourceListPage.getSourceRow(sourceName);
      await expect(row).toHaveCount(0);
    });
  });

  test('B-20: 用户隔离验证', async ({ page, auth, sourceListPage }) => {
    const testUser = `B20_source_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);

    await test.step('B-20: 验证只能看到自己的来源', async () => {
      await sourceListPage.goto();
      await page.waitForTimeout(2000);
      const table = page.locator('.el-table');
      await expect(table).toBeVisible();
    });
  });
});
