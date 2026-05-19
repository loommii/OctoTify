import { test, expect } from '../fixtures/test-fixtures';

test.describe('B-Navigation: 导航与路由', () => {
  test.beforeEach(async ({ page, auth }) => {
    const testUser = `B_navigation_${Date.now()}`;
    const testPass = 'Test123456';
    await auth.registerAndLogin(testUser, testPass);
  });

  test('B-56: 正向 - 侧边栏显示所有模块菜单', async ({ page }) => {
    await test.step('B-56-02: 验证侧边栏菜单项存在', async () => {
      await expect(page.getByRole('menuitem', { name: '仪表盘' })).toBeVisible();
      await expect(page.getByRole('menuitem', { name: '消息来源' })).toBeVisible();
      await expect(page.getByRole('menuitem', { name: '推送渠道' })).toBeVisible();
      await expect(page.getByRole('menuitem', { name: '消息记录' })).toBeVisible();
      await expect(page.getByRole('menuitem', { name: '系统设置' })).toBeVisible();
    });
  });

  test('B-57: 正向 - 点击"仪表盘"菜单导航正确', async ({ page }) => {
    await page.goto('/source/list');
    await page.waitForURL(/\/source\/list/, { timeout: 10000 });

    await test.step('B-57-02: 点击"仪表盘"菜单', async () => {
      await page.getByRole('menuitem', { name: '仪表盘' }).click();
    });

    await test.step('B-57-03: 验证URL为仪表盘', async () => {
      await expect(page).toHaveURL(/\/dashboard/);
    });
  });

  test('B-58: 正向 - 点击"消息来源"菜单导航正确', async ({ page }) => {
    await test.step('B-58-02: 点击"消息来源"菜单', async () => {
      await page.getByRole('menuitem', { name: '消息来源' }).click();
    });

    await test.step('B-58-03: 验证URL为消息来源列表', async () => {
      await expect(page).toHaveURL(/\/source\/list/);
    });

    await test.step('B-58-04: 验证消息来源菜单高亮', async () => {
      await expect(page.locator('.el-menu-item.is-active')).toContainText('消息来源');
    });
  });

  test('B-59: 正向 - 点击"推送渠道"菜单导航正确', async ({ page }) => {
    await test.step('B-59-02: 点击"推送渠道"菜单', async () => {
      await page.getByRole('menuitem', { name: '推送渠道' }).click();
    });

    await test.step('B-59-03: 验证URL为推送渠道列表', async () => {
      await expect(page).toHaveURL(/\/channel\/list/);
    });

    await test.step('B-59-04: 验证推送渠道菜单高亮', async () => {
      await expect(page.locator('.el-menu-item.is-active')).toContainText('推送渠道');
    });
  });

  test('B-60: 正向 - 点击"消息记录"菜单导航正确', async ({ page }) => {
    await test.step('B-60-02: 点击"消息记录"菜单', async () => {
      await page.getByRole('menuitem', { name: '消息记录' }).click();
    });

    await test.step('B-60-03: 验证URL为消息记录列表', async () => {
      await expect(page).toHaveURL(/\/message\/list/);
    });

    await test.step('B-60-04: 验证消息记录菜单高亮', async () => {
      await expect(page.locator('.el-menu-item.is-active')).toContainText('消息记录');
    });
  });

  test('B-61: 正向 - 详情页面包屑显示', async ({ page, sourceListPage }) => {
    await test.step('B-61-02: 进入消息来源列表', async () => {
      await page.goto('/source/list');
      await page.waitForURL(/\/source\/list/, { timeout: 10000 });
    });

    await test.step('B-61-03: 创建测试数据并进入详情页', async () => {
      const createBtn = page.getByRole('button', { name: '新增' });
      if (await createBtn.isVisible().catch(() => false)) {
        await createBtn.click();
        await page.waitForURL(/\/source\/create/, { timeout: 10000 });
        await page.locator('.el-input').first().fill('Test Source');
        await page.getByRole('button', { name: '新增' }).click();
        await page.waitForURL(/\/source\/list/, { timeout: 10000 });
      }

      const detailBtn = page.locator('.el-table__body tr').first().getByRole('button', { name: '查看详情' });
      if (await detailBtn.isVisible().catch(() => false)) {
        await detailBtn.click();
        await page.waitForURL(/\/source\/detail\/\d+/, { timeout: 10000 });
      }
    });

    await test.step('B-61-04: 验证面包屑显示', async () => {
      const breadcrumb = page.locator('.el-breadcrumb');
      await expect(breadcrumb).toBeVisible();
      await expect(breadcrumb).toContainText('首页');
      await expect(breadcrumb).toContainText('消息来源');
    });
  });

  test('B-62: 正向 - 编辑页面包屑显示', async ({ page, sourceListPage }) => {
    await page.goto('/source/list');
    await page.waitForURL(/\/source\/list/, { timeout: 10000 });

    await test.step('B-62-02: 进入编辑页面', async () => {
      const editBtn = page.locator('.el-table__body tr').first().getByRole('button', { name: '修改' });
      if (await editBtn.isVisible().catch(() => false)) {
        await editBtn.click();
        await page.waitForURL(/\/source\/edit\/\d+/, { timeout: 10000 });
      }
    });

    await test.step('B-62-03: 验证面包屑显示', async () => {
      const breadcrumb = page.locator('.el-breadcrumb');
      await expect(breadcrumb).toBeVisible();
      await expect(breadcrumb).toContainText('首页');
      await expect(breadcrumb).toContainText('消息来源');
      await expect(breadcrumb).toContainText('编辑');
    });
  });

  test('B-63: 反向 - 未登录访问受保护路由重定向到登录页', async ({ browser }) => {
    await test.step('B-08-01: 创建无登录状态的浏览器上下文', async () => {
      const context = await browser.newContext();
      const newPage = await context.newPage();

      await test.step('B-08-02: 访问受保护路由仪表盘', async () => {
        await newPage.goto('/dashboard/index');
        await newPage.waitForURL(/\/auth\/login/, { timeout: 10000 });
      });

      await test.step('B-08-03: 验证重定向到登录页', async () => {
        await expect(newPage).toHaveURL(/\/auth\/login/);
      });

      await newPage.close();
      await context.close();
    });

    await test.step('B-08-04: 测试消息来源页面受保护', async () => {
      const context = await browser.newContext();
      const newPage = await context.newPage();

      await newPage.goto('/source/list');
      await newPage.waitForURL(/\/auth\/login/, { timeout: 10000 });
      await expect(newPage).toHaveURL(/\/auth\/login/);

      await newPage.close();
      await context.close();
    });
  });

  test('B-64: 正向 - 返回按钮工作正常', async ({ page, sourceListPage }) => {
    await page.goto('/source/list');
    await page.waitForURL(/\/source\/list/, { timeout: 10000 });

    await test.step('B-09-02: 进入编辑页面', async () => {
      const editBtn = page.locator('.el-table__body tr').first().getByRole('button', { name: '修改' });
      if (await editBtn.isVisible().catch(() => false)) {
        await editBtn.click();
        await page.waitForURL(/\/source\/edit\/\d+/, { timeout: 10000 });
      }
    });

    await test.step('B-09-03: 点击返回按钮', async () => {
      const backBtn = page.getByRole('button', { name: '返回' });
      if (await backBtn.isVisible().catch(() => false)) {
        await backBtn.click();
      }
    });

    await test.step('B-09-04: 验证返回到列表页', async () => {
      await page.waitForURL(/\/source\/list/, { timeout: 10000 });
      await expect(page).toHaveURL(/\/source\/list/);
    });
  });

  test('B-65: 正向 - 侧边栏活跃菜单高亮', async ({ page }) => {
    await page.goto('/source/list');
    await page.waitForURL(/\/source\/list/, { timeout: 10000 });

    await test.step('B-10-02: 验证消息来源菜单高亮', async () => {
      const activeMenu = page.locator('.el-menu-item.is-active');
      await expect(activeMenu).toContainText('消息来源');
    });

    await test.step('B-10-03: 导航到推送渠道列表', async () => {
      await page.getByRole('menuitem', { name: '推送渠道' }).click();
      await page.waitForURL(/\/channel\/list/, { timeout: 10000 });
    });

    await test.step('B-10-04: 验证推送渠道菜单高亮', async () => {
      const activeMenu = page.locator('.el-menu-item.is-active');
      await expect(activeMenu).toContainText('推送渠道');
    });

    await test.step('B-10-05: 导航到消息记录列表', async () => {
      await page.getByRole('menuitem', { name: '消息记录' }).click();
      await page.waitForURL(/\/message\/list/, { timeout: 10000 });
    });

    await test.step('B-10-06: 验证消息记录菜单高亮', async () => {
      const activeMenu = page.locator('.el-menu-item.is-active');
      await expect(activeMenu).toContainText('消息记录');
    });
  });
});