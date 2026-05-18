import { test, expect } from '../fixtures/test-fixtures';

test.describe('UC_Register: 注册账户', () => {
  test('A-01: 用户名已存在 - 注册失败，页面不跳转', async ({ page, registerPage }) => {
    const username = `user_${Math.random().toString(36).substring(2, 8)}_${Date.now()}`;
    const password = 'Test123456';

    await test.step('先注册一个用户', async () => {
      await registerPage.goto();
      await registerPage.register(username, password);
      await page.waitForURL(/\/auth\/login/, { timeout: 15000 });
    });

    await test.step('再次用相同用户名注册', async () => {
      await registerPage.goto();
      await registerPage.fillRegisterForm(username, password);
    });

    await test.step('验证页面不跳转', async () => {
      await expect(page).toHaveURL(/\/auth\/register/);
    });

    await test.step('验证错误提示', async () => {
      const errorMessage = await registerPage.getErrorMessage();
      expect(errorMessage).toBe('用户名已存在');
    });
  });

  test('A-02: 密码少于8位 - 注册失败，显示错误提示', async ({ page, registerPage }) => {
    const username = `user_${Math.random().toString(36).substring(2, 8)}_${Date.now()}`;
    const password = '123';

    await test.step('填写注册表单', async () => {
      await registerPage.goto();
      await registerPage.fillRegisterForm(username, password);
    });

    await test.step('验证错误提示', async () => {
      const errorMessage = await registerPage.getErrorMessage();
      expect(errorMessage).toBe('密码长度不能少于 8 个字符');
    });
  });

  test('A-03: 两次密码不一致 - 注册失败，显示错误提示', async ({ page, registerPage }) => {
    const username = `user_${Math.random().toString(36).substring(2, 8)}_${Date.now()}`;

    await test.step('填写注册表单', async () => {
      await registerPage.goto();
      await registerPage.usernameInput.fill(username);
      await registerPage.passwordInput.fill('Test123456');
      await registerPage.confirmPasswordInput.fill('Test12345678');
      await registerPage.registerButton.click();
    });

    await test.step('验证错误提示', async () => {
      const errorMessage = await registerPage.getErrorMessage();
      expect(errorMessage).toBe('两次输入的密码不一致');
    });
  });

  test('A-04: 用户名为空 - 注册失败，显示错误提示', async ({ page, registerPage }) => {
    await test.step('填写注册表单', async () => {
      await registerPage.goto();
      await registerPage.fillRegisterForm('', 'Test123456');
    });

    await test.step('验证错误提示', async () => {
      const errorMessage = await registerPage.getErrorMessage();
      expect(errorMessage).toBe('用户名长度不能少于 3 个字符');
    });
  });

  test('A-05: 密码为空 - 注册失败，显示错误提示', async ({ page, registerPage }) => {
    const username = `user_${Math.random().toString(36).substring(2, 8)}_${Date.now()}`;

    await test.step('填写注册表单', async () => {
      await registerPage.goto();
      await registerPage.usernameInput.fill(username);
      await registerPage.registerButton.click();
    });

    await test.step('验证错误提示', async () => {
      const errorMessage = await registerPage.getErrorMessage();
      expect(errorMessage).toBe('密码长度不能少于 8 个字符');
    });
  });

  test('A-06: 特殊字符用户名 - 注册失败，显示表单错误', async ({ page, registerPage }) => {
    await test.step('填写注册表单', async () => {
      await registerPage.goto();
      await registerPage.usernameInput.fill('user@test#123');
      await registerPage.passwordInput.fill('Test123456');
      await registerPage.confirmPasswordInput.fill('Test123456');
      // 触发 blur 校验
      await registerPage.passwordInput.click();
    });

    await test.step('验证错误提示', async () => {
      const errorMessage = await registerPage.getErrorMessage();
      expect(errorMessage).toBe('用户名只能包含字母、数字和下划线');
    });
  });

  test('A-07: 64位密码(上限) - 注册成功', async ({ page, registerPage }) => {
    const username = `user_${Math.random().toString(36).substring(2, 8)}_${Date.now()}`;
    const password = 'Test123456789012345678901234567890123456789012345678901234567890'; // 64位

    await test.step('验证密码长度', async () => {
      expect(password.length).toBe(64);
    });

    await test.step('填写注册表单', async () => {
      await registerPage.goto();
      await registerPage.register(username, password);
    });

    await test.step('验证页面跳转', async () => {
      await expect(page).toHaveURL(/\/auth\/login/);
    });
  });

  test('A-08: 65位密码(超过上限) - 注册失败，显示表单错误', async ({ page, registerPage }) => {
    const password = 'Test1234567890123456789012345678901234567890123456789012345678901'; // 65位

    await test.step('验证密码长度', async () => {
      expect(password.length).toBe(65);
    });

    await test.step('填写注册表单', async () => {
      await registerPage.goto();
      await registerPage.usernameInput.fill(`user_${Math.random().toString(36).substring(2, 8)}`);
      await registerPage.passwordInput.fill(password);
      await registerPage.confirmPasswordInput.fill(password);
      // 触发 blur 校验
      await registerPage.passwordInput.click();
    });

    await test.step('验证错误提示', async () => {
      const errorMessage = await registerPage.getErrorMessage();
      expect(errorMessage).toBe('密码长度不能超过 64 个字符');
    });
  });

  test('A-09: 64位用户名(上限) - 注册成功', async ({ page, registerPage }) => {
    const randomPart = Math.random().toString(36).substring(2).repeat(10).substring(0, 59);
    const username = `user_${randomPart}`;
    const password = 'Test123456';

    await test.step('验证用户名长度', async () => {
      expect(username.length).toBe(64);
    });

    await test.step('填写注册表单', async () => {
      await registerPage.goto();
      await registerPage.register(username, password);
    });

    await test.step('验证页面跳转', async () => {
      await expect(page).toHaveURL(/\/auth\/login/);
    });
  });

  test('A-10: 65位用户名(超过上限) - 注册失败，显示表单错误', async ({ page, registerPage }) => {
    const randomPart = Math.random().toString(36).substring(2).repeat(10).substring(0, 60);
    const username = `user_${randomPart}`;

    await test.step('验证用户名长度', async () => {
      expect(username.length).toBe(65);
    });

    await test.step('填写注册表单', async () => {
      await registerPage.goto();
      await registerPage.usernameInput.fill(username);
      await registerPage.passwordInput.fill('Test123456');
      await registerPage.confirmPasswordInput.fill('Test123456');
      // 触发 blur 校验
      await registerPage.passwordInput.click();
    });

    await test.step('验证错误提示', async () => {
      const errorMessage = await registerPage.getErrorMessage();
      expect(errorMessage).toBe('用户名长度不能超过 64 个字符');
    });
  });
});

test.describe('UC_Login: 登录', () => {
  test('A-11: 用户名不存在 - 登录失败，显示错误提示', async ({ page, loginPage }) => {
    const username = `nonexistent_${Math.random().toString(36).substring(2, 8)}`;
    const password = 'Test123456';

    await test.step('填写登录表单', async () => {
      await loginPage.goto();
      await loginPage.fillLoginForm(username, password);
    });

    await test.step('验证页面不跳转', async () => {
      await expect(page).toHaveURL(/\/auth\/login/);
    });

    await test.step('验证错误提示', async () => {
      const errorMessage = await loginPage.getErrorMessage();
      expect(errorMessage).toBe('用户名或密码错误');
    });
  });

  test('A-12: 密码错误 - 登录失败，显示错误提示', async ({ page, loginPage, registerPage }) => {
    const username = `user_${Math.random().toString(36).substring(2, 8)}_${Date.now()}`;
    const password = 'Test123456';

    await test.step('先注册一个用户', async () => {
      await registerPage.goto();
      await registerPage.register(username, password);
      await page.waitForURL(/\/auth\/login/, { timeout: 15000 });
    });

    await test.step('用错误密码登录', async () => {
      await loginPage.goto();
      await loginPage.fillLoginForm(username, 'WrongPassword123');
    });

    await test.step('验证页面不跳转', async () => {
      await expect(page).toHaveURL(/\/auth\/login/);
    });

    await test.step('验证错误提示', async () => {
      const errorMessage = await loginPage.getErrorMessage();
      expect(errorMessage).toBe('用户名或密码错误');
    });
  });

  test('A-13: 用户名为空 - 登录失败，显示错误提示', async ({ page, loginPage }) => {
    await test.step('填写登录表单', async () => {
      await loginPage.goto();
      await loginPage.fillLoginForm('', 'Test123456');
    });

    await test.step('验证错误提示', async () => {
      const errorMessage = await loginPage.getErrorMessage();
      expect(errorMessage).toBe('请输入用户名');
    });
  });

  test('A-14: 密码为空 - 登录失败，显示错误提示', async ({ page, loginPage }) => {
    const username = `user_${Math.random().toString(36).substring(2, 8)}_${Date.now()}`;

    await test.step('填写登录表单', async () => {
      await loginPage.goto();
      await loginPage.usernameInput.fill(username);
      await loginPage.loginButton.click();
    });

    await test.step('验证错误提示', async () => {
      const errorMessage = await loginPage.getErrorMessage();
      expect(errorMessage).toBe('请输入密码');
    });
  });

  test('A-15: 特殊字符用户名登录 - 登录失败，显示错误提示', async ({ page, loginPage }) => {
    const password = 'Test123456';

    await test.step('填写登录表单', async () => {
      await loginPage.goto();
      await loginPage.fillLoginForm('user@test#123', password);
    });

    await test.step('验证页面不跳转', async () => {
      await expect(page).toHaveURL(/\/auth\/login/);
    });

    await test.step('验证错误提示', async () => {
      const errorMessage = await loginPage.getErrorMessage();
      expect(errorMessage).toBe('用户名或密码错误');
    });
  });
});

test.describe('UC_ChangePassword: 修改密码', () => {
  test('修改密码错误场景验证', async ({ page, passwordPage, registerPage, loginPage }) => {
    const username = `user_${Date.now()}`;
    const password = 'Test123456';

    await test.step('准备：注册新用户并登录', async () => {
      await registerPage.goto();
      await registerPage.register(username, password);
      await page.waitForURL(/\/auth\/login/, { timeout: 15000 });

      await loginPage.goto();
      await loginPage.login(username, password);
      await page.waitForURL(/\/dashboard/, { timeout: 15000 });

      await passwordPage.goto();
      await passwordPage.expectLoaded();
    });

    await test.step('A-16: 旧密码错误', async () => {
      await passwordPage.fillPasswordForm('WrongOldPwd1', 'NewTest123', 'NewTest123');
      await passwordPage.submitButton.click();
      const errorMessage = await passwordPage.getErrorMessage();
      expect(errorMessage).toBe('旧密码错误');
      // 等待 toast 消失，避免影响后续测试
      await passwordPage.errorToast.waitFor({ state: 'hidden', timeout: 5000 }).catch(() => {});
    });

    await test.step('A-17: 新密码太短', async () => {
      await passwordPage.fillPasswordForm(password, '123', '123');
      await passwordPage.submitButton.click();
      const errorMessage = await passwordPage.getErrorMessage();
      expect(errorMessage).toContain('8-64 字符');
    });

    await test.step('A-18: 两次新密码不一致', async () => {
      await passwordPage.fillPasswordForm(password, 'NewTest123', 'NewTest456');
      await passwordPage.submitButton.click();
      const errorMessage = await passwordPage.getErrorMessage();
      expect(errorMessage).toBe('两次输入的密码不一致');
    });

    await test.step('A-19: 新密码与旧密码相同', async () => {
      await passwordPage.fillPasswordForm(password, password, password);
      await passwordPage.submitButton.click();
      const errorMessage = await passwordPage.getErrorMessage();
      expect(errorMessage).toBe('新密码不能与旧密码相同');
    });
  });
});

test.describe('UC_RefreshToken: 刷新令牌', () => {
  test('A-20: 使用无效 refresh_token - 刷新失败，返回错误提示', async ({ page, registerPage, loginPage, header }) => {
    const username = `user_${Date.now()}`;
    const password = 'Test123456';
    let refreshToken = '';

    await test.step('准备：注册新用户并登录', async () => {
      await registerPage.goto();
      await registerPage.register(username, password);
      await page.waitForURL(/\/auth\/login/, { timeout: 15000 });

      await loginPage.goto();
      await loginPage.login(username, password);
      await page.waitForURL(/\/dashboard/, { timeout: 15000 });
    });

    await test.step('查看个人资料', async () => {
      await header.gotoProfile();
      await page.waitForURL(/\/settings\/profile/, { timeout: 10000 });
      const displayedUsername = await page.getByText(username).textContent();
      expect(displayedUsername).toBe(username);
    });

    await test.step('获取 refresh_token', async () => {
      const response = await page.request.post('/api/auth/login', {
        data: { username, password },
      });
      const respData = await response.json();
      expect(respData.code).toBe(0);
      refreshToken = respData.data.refresh_token;
      expect(refreshToken).toBeTruthy();
    });

    await test.step('退出登录', async () => {
      await header.logout();
      await page.waitForURL(/\/auth\/login/, { timeout: 10000 });
    });

    await test.step('使用已失效的 refresh_token 刷新', async () => {
      const response = await page.request.post('/api/auth/refresh', {
        data: { refresh_token: refreshToken },
      });
      const respData = await response.json();
      expect(respData.code).toBe(110301);
      expect(respData.msg).toBe('刷新令牌已撤销');
    });
  });
});
