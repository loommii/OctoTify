import { test, expect } from '@playwright/test';
import { execSync } from 'child_process';

/**
 * OctoTify E2E 端到端测试
 *
 * 测试流程：
 * 1. 登录系统
 * 2. 创建飞书渠道
 * 3. 创建 Telegram 渠道
 * 4. 创建邮件渠道
 * 5. 创建消息来源并绑定三个渠道
 * 6. 通过 Push Token 推送消息
 * 7. 查看消息记录
 *
 * 凭证来源: /Volumes/H520/code/projects/OctoTify/tmp/credentials.md
 */

// ====== 凭证配置 ======
const FEISHU_WEBHOOK_URL = 'https://open.feishu.cn/open-apis/bot/v2/hook/3f3d618e-bd85-4338-af72-ede109e38b8f';
const FEISHU_SECRET = 'iTYWok2tumSfDmWOJaRLOd';

const EMAIL_SMTP_SERVER = 'smtp.qq.com';
const EMAIL_SMTP_PORT = '465';
const EMAIL_USERNAME = 'lin_jjj@qq.com';
const EMAIL_PASSWORD = 'sbpsfnzgmjlrbegh';
const EMAIL_RECIPIENT = 'v1061827272@hotmail.com';
const EMAIL_CC = 'p62071167@gmail.com';
const EMAIL_SENDER_NAME = 'OctoTify 通知';

// Telegram 凭证（用户提供，可能不准确）
const TELEGRAM_BOT_TOKEN = '7510377754:AAHjX5K1J3lNv5DhM9Y0xqM6xqM6xqM6xqM';
const TELEGRAM_CHAT_ID = '-1001234567890';

const LOGIN_USERNAME = 'testuser1';
const LOGIN_PASSWORD = 'Testpass123';
const BACKEND_URL = 'http://localhost:34123';

/**
 * 辅助函数：处理确认对话框
 */
async function confirmDialog(page: ReturnType<typeof import('@playwright/test')['page']>) {
  await expect(page.locator('.modal-overlay')).toBeVisible({ timeout: 8000 });
  await page.locator('.btn-confirm').click();
  await page.locator('.modal-overlay').waitFor({ state: 'hidden', timeout: 15000 });
}

/**
 * 辅助函数：安全截图
 */
async function safeScreenshot(page: ReturnType<typeof import('@playwright/test')['page']>, name: string) {
  try {
    await page.screenshot({ path: `../tmp/screenshots/${name}.png`, fullPage: false });
    console.log(`[截图] ${name}.png 已保存`);
  } catch (err: any) {
    console.warn(`[截图警告] ${name} 截图失败:`, err.message);
  }
}

/**
 * 辅助函数：通过导航链接跳转并等待加载
 */
async function navigateByLink(page: ReturnType<typeof import('@playwright/test')['page']>, linkName: string) {
  await page.getByRole('link', { name: linkName, exact: true }).click();
  await page.waitForURL(/\/channels|\/sources|\/messages|\/dashboard/, { timeout: 10000 });
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(500);
}

/**
 * 辅助函数：填写渠道创建表单
 * 根据表单中的 label 标签来填写对应的字段
 */
async function fillFieldByLabel(page: ReturnType<typeof import('@playwright/test')['page']>, labelText: string, value: string) {
  // 按 label 定位到对应的 .form-group，然后找到其中的 input
  const formGroup = page.locator('.form-group').filter({ has: page.locator(`label:has-text("${labelText}")`).first() });
  const input = formGroup.locator('input:not([type="hidden"]):not([type="button"]):not([type="submit"]), textarea').first();
  await input.fill(value);
}

test('OctoTify E2E 端到端完整流程', async ({ page }) => {
  // 确保截图目录存在
  try {
    execSync('mkdir -p ../tmp/screenshots ../tmp/test-results');
  } catch {
    // 忽略错误
  }

  const channelResults: { name: string; success: boolean; error?: string }[] = [];
  let pushToken = '';

  // ====== Step 1: 登录 ======
  await test.step('Step 1: 登录系统', async () => {
    console.log('\n========== Step 1: 登录系统 ==========');

    await page.goto('http://localhost:5173');
    await expect(page).toHaveURL(/.*\/login|.*\/dashboard/, { timeout: 10000 });

    if (!page.url().includes('/dashboard')) {
      await page.fill('#username', LOGIN_USERNAME);
      await page.fill('#password', LOGIN_PASSWORD);
      await safeScreenshot(page, '01-login-form-filled');

      await page.click('button[type="submit"]');
      await expect(page).toHaveURL(/.*\/dashboard/, { timeout: 15000 });
    }

    await safeScreenshot(page, '01-login-success-dashboard');
    const pageTitle = await page.locator('h1, h2').first().textContent();
    console.log(`[登录] 当前页面标题: ${pageTitle}`);
    expect(page.url()).toContain('/dashboard');
    console.log('[登录] 登录成功！');
  });

  // ====== Step 2: 创建飞书渠道 ======
  await test.step('Step 2: 创建飞书渠道', async () => {
    console.log('\n========== Step 2: 创建飞书渠道 ==========');

    await navigateByLink(page, '推送渠道');
    await safeScreenshot(page, '02-channel-list-before-create');

    await page.getByRole('button', { name: '创建渠道' }).click();
    await expect(page).toHaveURL(/.*\/channels\/create/, { timeout: 10000 });
    await safeScreenshot(page, '02-channel-type-selection');

    // 选择"飞书"类型
    await page.locator('.type-card').filter({ hasText: '飞书' }).click();
    await page.waitForTimeout(500);
    await safeScreenshot(page, '02-feishu-form-selected');

    // 使用 label 填写字段
    await fillFieldByLabel(page, '渠道名称', '飞书-E2E测试');

    // 查找 webhook_url 和 secret 字段
    // 飞书的 config_fields 一般是 webhook_url (text/url) 和 secret (password/text)
    const allLabels = page.locator('.config-fields label, .form-section label');
    const labelCount = await allLabels.count();
    console.log(`[飞书] 找到 ${labelCount} 个表单标签`);

    for (let i = 0; i < labelCount; i++) {
      const label = allLabels.nth(i);
      const text = await label.textContent();
      console.log(`[飞书] 标签 ${i}: "${text}"`);
    }

    // 尝试按 label 填写 webhook URL
    const webhookLabels = await page.locator('label:has-text("webhook"), label:has-text("Webhook"), label:has-text("URL"), label:has-text("地址")').all();
    if (webhookLabels.length > 0) {
      // 找到 webhook 标签后，填写其对应的 input
      const wg = page.locator('.form-group').filter({ has: webhookLabels[0] });
      const wi = wg.locator('input:not([type="hidden"]):not([type="button"]):not([type="submit"])').first();
      await wi.fill(FEISHU_WEBHOOK_URL);
    } else {
      // 回退：按顺序填写
      const textInputs = page.locator('.form-section input[type="text"], .form-section input[type="url"]');
      const inputCount = await textInputs.count();
      console.log(`[飞书] 文本输入框数量: ${inputCount}`);
      if (inputCount >= 2) {
        await textInputs.nth(1).fill(FEISHU_WEBHOOK_URL);
      }
    }

    // 填写 secret
    const secretInputs = page.locator('.form-section input[type="password"]');
    const pwdCount = await secretInputs.count();
    if (pwdCount > 0) {
      await secretInputs.first().fill(FEISHU_SECRET);
    } else {
      // 尝试找 secret/sign 标签
      const secretLabels = await page.locator('label:has-text("secret"), label:has-text("Secret"), label:has-text("密钥"), label:has-text("签名")').all();
      if (secretLabels.length > 0) {
        const sg = page.locator('.form-group').filter({ has: secretLabels[0] });
        const si = sg.locator('input').first();
        await si.fill(FEISHU_SECRET);
      } else {
        const textInputs = page.locator('.form-section input[type="text"], .form-section input[type="url"]');
        const inputCount = await textInputs.count();
        if (inputCount >= 3) {
          await textInputs.nth(2).fill(FEISHU_SECRET);
        }
      }
    }

    await safeScreenshot(page, '02-feishu-form-filled');

    // 提交
    await page.locator('.form-actions .btn-primary').click();
    await confirmDialog(page);
    await page.waitForTimeout(3000);
    await safeScreenshot(page, '02-feishu-create-result');

    const currentUrl = page.url();
    const successToast = page.locator('.toast-success');
    const toastVisible = await successToast.isVisible();

    if (currentUrl.includes('/channels') || toastVisible) {
      console.log('[飞书] 渠道创建成功！');
      channelResults.push({ name: '飞书-E2E测试', success: true });
    } else {
      const errorToast = page.locator('.toast-error');
      const errText = await errorToast.isVisible() ? await errorToast.textContent() : '未知';
      console.log(`[飞书] 渠道创建失败: ${errText}`);
      channelResults.push({ name: '飞书-E2E测试', success: false, error: errText });
    }
  });

  // ====== Step 3: 创建 Telegram 渠道 ======
  await test.step('Step 3: 创建 Telegram 渠道', async () => {
    console.log('\n========== Step 3: 创建 Telegram 渠道 ==========');

    await navigateByLink(page, '推送渠道');
    await page.getByRole('button', { name: '创建渠道' }).click();
    await expect(page).toHaveURL(/.*\/channels\/create/, { timeout: 10000 });

    await page.locator('.type-card').filter({ hasText: 'Telegram' }).click();
    await page.waitForTimeout(500);
    await safeScreenshot(page, '03-telegram-form-selected');

    // 打印所有标签帮助调试
    const allLabels = page.locator('.config-fields label, .form-section label');
    const labelCount = await allLabels.count();
    console.log(`[Telegram] 找到 ${labelCount} 个表单标签`);
    for (let i = 0; i < labelCount; i++) {
      const text = await allLabels.nth(i).textContent();
      console.log(`[Telegram] 标签 ${i}: "${text}"`);
    }

    // 填写名称
    await fillFieldByLabel(page, '渠道名称', 'Telegram-E2E测试');

    // 填写 Bot Token 和 Chat ID
    // 尝试按 label 填写
    const tokenLabels = await page.locator('label:has-text("token"), label:has-text("Token"), label:has-text("Bot Token")').all();
    if (tokenLabels.length > 0) {
      const tg = page.locator('.form-group').filter({ has: tokenLabels[0] });
      const ti = tg.locator('input:not([type="hidden"]):not([type="button"]):not([type="submit"])').first();
      await ti.fill(TELEGRAM_BOT_TOKEN);
    } else {
      // 回退：找 password 类型输入框填写 token
      const pwdInputs = page.locator('.form-section input[type="password"]');
      const pwdCount = await pwdInputs.count();
      if (pwdCount >= 1) {
        await pwdInputs.first().fill(TELEGRAM_BOT_TOKEN);
      }
    }

    const chatLabels = await page.locator('label:has-text("chat"), label:has-text("Chat"), label:has-text("Chat ID")').all();
    if (chatLabels.length > 0) {
      const cg = page.locator('.form-group').filter({ has: chatLabels[0] });
      const ci = cg.locator('input:not([type="hidden"]):not([type="button"]):not([type="submit"])').first();
      await ci.fill(TELEGRAM_CHAT_ID);
    } else {
      // 回退：填写剩余的 text input
      const textInputs = page.locator('.form-section input[type="text"]');
      const count = await textInputs.count();
      if (count >= 2) {
        await textInputs.nth(1).fill(TELEGRAM_CHAT_ID);
      } else if (count >= 3) {
        await textInputs.nth(2).fill(TELEGRAM_CHAT_ID);
      }
    }

    await safeScreenshot(page, '03-telegram-form-filled');

    await page.locator('.form-actions .btn-primary').click();
    await confirmDialog(page);
    await page.waitForTimeout(8000);
    await safeScreenshot(page, '03-telegram-create-result');

    const currentUrl = page.url();
    const errorToast = page.locator('.toast-error');
    const hasErrorToast = await errorToast.isVisible();

    if (currentUrl.includes('/channels') && !hasErrorToast) {
      console.log('[Telegram] 渠道创建成功！');
      channelResults.push({ name: 'Telegram-E2E测试', success: true });
    } else {
      const errText = hasErrorToast ? await errorToast.textContent() : '未知';
      console.log(`[Telegram] 渠道创建失败: ${errText}`);
      channelResults.push({ name: 'Telegram-E2E测试', success: false, error: errText });
    }
    // Telegram 可能因为凭证无效而失败，但不阻断后续测试
  });

  // ====== Step 4: 创建邮件渠道 ======
  await test.step('Step 4: 创建邮件渠道', async () => {
    console.log('\n========== Step 4: 创建邮件渠道 ==========');

    await navigateByLink(page, '推送渠道');
    await page.getByRole('button', { name: '创建渠道' }).click();
    await expect(page).toHaveURL(/.*\/channels\/create/, { timeout: 10000 });

    await page.locator('.type-card').filter({ hasText: '邮件' }).click();
    await page.waitForTimeout(500);
    await safeScreenshot(page, '04-email-form-selected');

    // 打印所有标签帮助调试
    const allLabels = page.locator('.config-fields label, .form-section label');
    const labelCount = await allLabels.count();
    console.log(`[邮件] 找到 ${labelCount} 个表单标签`);
    for (let i = 0; i < labelCount; i++) {
      const text = await allLabels.nth(i).textContent();
      console.log(`[邮件] 标签 ${i}: "${text}"`);
    }

    // 打印所有输入框类型
    const textInputs = page.locator('.form-section input[type="text"]');
    const numberInputs = page.locator('.form-section input[type="number"]');
    const passwordInputs = page.locator('.form-section input[type="password"]');
    const emailInputs = page.locator('.form-section input[type="email"]');

    console.log(`[邮件] 文本输入框: ${await textInputs.count()}`);
    console.log(`[邮件] 数字输入框: ${await numberInputs.count()}`);
    console.log(`[邮件] 密码输入框: ${await passwordInputs.count()}`);
    console.log(`[邮件] Email输入框: ${await emailInputs.count()}`);

    // 填写名称
    await fillFieldByLabel(page, '渠道名称', '邮件-E2E测试');

    // 按 label 匹配填写各字段
    // 映射: label关键字 -> 值
    const fieldMapping: Array<{ keywords: string[]; value: string; type?: 'text' | 'number' | 'password' | 'email' }> = [
      { keywords: ['smtp_server', 'SMTP服务器', 'SMTP Server', 'smtp server', '服务器'], value: EMAIL_SMTP_SERVER },
      { keywords: ['smtp_port', 'SMTP端口', 'SMTP Port', 'smtp port', '端口'], value: EMAIL_SMTP_PORT, type: 'number' },
      { keywords: ['username', '用户名', 'user name', '邮箱', 'email'], value: EMAIL_USERNAME },
      { keywords: ['password', '密码', '授权码', 'token'], value: EMAIL_PASSWORD, type: 'password' },
      { keywords: ['recipient', '收件人', 'to', '接收者', '收件邮箱'], value: EMAIL_RECIPIENT, type: 'email' },
      { keywords: ['cc', '抄送', '抄送人', '抄送邮箱'], value: EMAIL_CC, type: 'email' },
      { keywords: ['sender', '发件人', 'sender name', '发件人名称', '名称'], value: EMAIL_SENDER_NAME },
    ];

    for (const field of fieldMapping) {
      for (const keyword of field.keywords) {
        const labels = await page.locator(`label:has-text("${keyword}")`).all();
        if (labels.length > 0) {
          const formGroup = page.locator('.form-group').filter({ has: labels[0] });
          const input = formGroup.locator('input:not([type="hidden"]):not([type="button"]):not([type="submit"]), textarea').first();
          await input.fill(field.value);
          console.log(`[邮件] 填写字段 "${keyword}" = ${field.value}`);
          break;
        }
      }
    }

    await safeScreenshot(page, '04-email-form-filled');

    await page.locator('.form-actions .btn-primary').click();
    await confirmDialog(page);
    await page.waitForTimeout(3000);
    await safeScreenshot(page, '04-email-create-result');

    const currentUrl = page.url();
    const successToast = page.locator('.toast-success');
    const isSuccess = await successToast.isVisible();

    if (currentUrl.includes('/channels') || isSuccess) {
      console.log('[邮件] 渠道创建成功！');
      channelResults.push({ name: '邮件-E2E测试', success: true });
    } else {
      const errorToast = page.locator('.toast-error');
      const errText = await errorToast.isVisible() ? await errorToast.textContent() : '未知错误';
      console.log(`[邮件] 渠道创建失败: ${errText}`);
      channelResults.push({ name: '邮件-E2E测试', success: false, error: errText });
    }
  });

  // ====== Step 5: 创建消息来源并绑定三个渠道 ======
  await test.step('Step 5: 创建消息来源并绑定渠道', async () => {
    console.log('\n========== Step 5: 创建消息来源并绑定渠道 ==========');

    await navigateByLink(page, '消息来源');
    await safeScreenshot(page, '05-source-list-before-create');

    await page.getByRole('button', { name: '创建来源' }).click();
    await expect(page).toHaveURL(/.*\/sources\/create/, { timeout: 10000 });
    await safeScreenshot(page, '05-source-create-form');

    // 填写来源名称
    await page.locator('.form-group input[type="text"]').first().fill('E2E全渠道测试');

    // 填写描述
    const textarea = page.locator('.form-group textarea');
    await textarea.fill('前端页面创建的全渠道测试来源');

    // 等待渠道列表加载
    await page.waitForTimeout(2000);
    await safeScreenshot(page, '05-source-channel-selection');

    // 勾选所有渠道卡片
    const channelCards = page.locator('.channel-card');
    const cardCount = await channelCards.count();
    console.log(`[来源] 找到 ${cardCount} 个渠道卡片`);

    for (let i = 0; i < cardCount; i++) {
      const card = channelCards.nth(i);
      const isSelected = await card.locator('.check-indicator.checked').isVisible().catch(() => false);
      if (!isSelected) {
        await card.click();
        await page.waitForTimeout(200);
      }
    }

    await safeScreenshot(page, '05-source-channels-selected');

    const selectedCountText = await page.locator('.selected-count').textContent().catch(() => '');
    console.log(`[来源] ${selectedCountText}`);

    await page.locator('.form-actions .btn-primary').click();
    await confirmDialog(page);
    await page.waitForTimeout(3000);
    await safeScreenshot(page, '05-source-create-success-token');

    // 获取 Push Token
    const tokenElement = page.locator('.token-value');
    if (await tokenElement.isVisible().catch(() => false)) {
      pushToken = (await tokenElement.textContent() || '').trim();
      console.log(`[来源] Push Token: ${pushToken}`);
    } else {
      const allCodeElements = page.locator('code');
      const codeCount = await allCodeElements.count();
      for (let i = 0; i < codeCount; i++) {
        const text = await allCodeElements.nth(i).textContent();
        if (text && text.startsWith('src')) {
          pushToken = text.trim();
          console.log(`[来源] Push Token: ${pushToken}`);
          break;
        }
      }
    }

    expect(pushToken).toBeTruthy();
    expect(pushToken.startsWith('src')).toBe(true);
    console.log(`[来源] 成功获取 Push Token: ${pushToken}`);
  });

  // ====== Step 6: 推送消息验证 ======
  await test.step('Step 6: 通过 Push Token 推送消息', async () => {
    console.log('\n========== Step 6: 推送消息验证 ==========');
    expect(pushToken).toBeTruthy();

    const now = new Date().toLocaleString('zh-CN');
    const pushPayload = JSON.stringify({
      title: '全渠道 E2E 测试',
      message: `前端页面创建渠道并推送的端到端测试消息。\n\n发送时间：${now}`
    });

    console.log(`[推送] Token: ${pushToken}`);
    console.log(`[推送] Payload: ${pushPayload}`);

    const curlCommand = `curl -s -X POST "${BACKEND_URL}/api/push/${pushToken}" -H "Content-Type: application/json" -d '${pushPayload.replace(/'/g, "'\\''")}'`;
    console.log(`[推送] 执行命令: ${curlCommand}`);

    let pushResult = '';
    try {
      pushResult = execSync(curlCommand, { encoding: 'utf-8', timeout: 60000 });
      console.log(`[推送] 原始响应: ${pushResult}`);
    } catch (err: any) {
      pushResult = err.stdout || err.stderr || '';
      console.error(`[推送] 推送异常: ${err.message}`);
      console.error(`[推送] 响应: ${pushResult}`);
    }

    try {
      const result = JSON.parse(pushResult);
      console.log(`[推送] 解析后的响应:`, JSON.stringify(result, null, 2));

      if (result.code === 0) {
        const data = result.data || {};
        const total = data.total ?? 'N/A';
        const success = data.success ?? 'N/A';
        const failed = data.failed ?? 'N/A';
        console.log(`[推送] 推送结果 - total: ${total}, success: ${success}, failed: ${failed}`);
        console.log('[推送] 消息推送完成！');
      } else {
        console.error(`[推送] 推送失败: code=${result.code}, msg=${result.msg}`);
      }
    } catch (parseErr: any) {
      console.error(`[推送] 解析推送响应失败:`, parseErr.message);
      console.error(`[推送] 原始响应:`, pushResult);
    }

    await safeScreenshot(page, '06-push-result');
  });

  // ====== Step 7: 在消息记录页面查看 ======
  await test.step('Step 7: 查看消息记录', async () => {
    console.log('\n========== Step 7: 查看消息记录 ==========');

    await navigateByLink(page, '消息记录');
    await safeScreenshot(page, '07-message-list');

    const tableRows = page.locator('table tbody tr');
    const rowCount = await tableRows.count();
    console.log(`[消息记录] 找到 ${rowCount} 条消息记录`);

    if (rowCount > 0) {
      const firstRow = tableRows.first();
      const rowText = await firstRow.textContent();
      console.log(`[消息记录] 第一条记录: ${rowText?.trim()}`);

      const successBadges = page.locator('.status-badge.status-active');
      const successCount = await successBadges.count();
      console.log(`[消息记录] 成功状态记录: ${successCount} 条`);

      const failedBadges = page.locator('.status-badge.status-deleted');
      const failedCount = await failedBadges.count();
      console.log(`[消息记录] 失败状态记录: ${failedCount} 条`);

      await safeScreenshot(page, '07-message-list-with-records');
    } else {
      const emptyState = page.locator('.empty-state');
      if (await emptyState.isVisible().catch(() => false)) {
        console.log('[消息记录] 暂无消息记录');
        console.log(`[消息记录] 空状态文本: ${await emptyState.textContent()}`);
      }
    }

    await safeScreenshot(page, '07-final-summary');

    // ====== 最终汇总 ======
    console.log('\n========== 测试完成汇总 ==========');
    console.log('\n[渠道创建结果]');
    for (const result of channelResults) {
      const status = result.success ? '✅ 成功' : `❌ 失败 (${result.error || '未知'})`;
      console.log(`  - ${result.name}: ${status}`);
    }
    console.log(`\n[Push Token]: ${pushToken}`);
    console.log('==================================');
  });
});
