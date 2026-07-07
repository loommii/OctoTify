import { type Page, type Locator } from '@playwright/test';
import { BasePage } from './BasePage';

/**
 * 来源创建页面对象 - /source/create
 *
 * 对应前端页面：创建消息来源表单页
 * 包含：名称输入、描述输入、渠道绑定选择、提交/取消
 */
export class SourceCreatePage extends BasePage {
  readonly nameInput: Locator;
  readonly descriptionInput: Locator;
  readonly channelSelector: Locator;
  readonly submitButton: Locator;
  readonly cancelButton: Locator;
  readonly errorMessage: Locator;

  constructor(page: Page) {
    super(page);
    // 来源名称输入框
    this.nameInput = page.getByPlaceholder('请输入来源名称');
    // 描述输入框（textarea）
    this.descriptionInput = page.locator('textarea');
    // 渠道绑定下拉选择器
    this.channelSelector = page.locator('.el-select').first();
    // 提交按钮
    this.submitButton = page.getByRole('button', { name: '新增' });
    // 取消按钮
    this.cancelButton = page.getByRole('button', { name: '取消' });
    // 错误提示
    this.errorMessage = page.locator('.el-message--error .el-message__content');
  }

  /** 导航到来源创建页 */
  async goto() {
    await super.goto('/source/create');
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * 填写创建来源表单
   * @param name 来源名称
   * @param description 来源描述（可选）
   * @param channels 要绑定的渠道名称列表（可选）
   */
  async fillForm(name: string, description?: string, channels?: string[]) {
    await this.nameInput.waitFor({ state: 'visible', timeout: 10000 });
    await this.nameInput.click();
    await this.nameInput.fill(name);
    if (description) {
      await this.descriptionInput.fill(description);
    }
    if (channels && channels.length > 0) {
      await this.channelSelector.click();
      for (const channel of channels) {
        await this.page.locator('.el-select-dropdown__item').filter({ hasText: channel }).click();
      }
      await this.page.keyboard.press('Escape');
    }
  }

  /** 提交表单 */
  async submit() {
    await this.submitButton.click();
  }

  /** 取消创建 */
  async cancel() {
    await this.cancelButton.click();
  }

  /** 获取错误提示文案 */
  async getErrorMessage(): Promise<string> {
    // 优先检查 toast 消息
    const toast = this.page.locator('.el-message--error .el-message__content');
    const toastVisible = await toast.isVisible({ timeout: 2000 }).catch(() => false);
    if (toastVisible) {
      return (await toast.textContent()) ?? '';
    }
    // 回退：检查 Element Plus 表单内联校验错误
    const inlineError = this.page.locator('.el-form-item__error').first();
    const inlineVisible = await inlineError.isVisible({ timeout: 1000 }).catch(() => false);
    if (inlineVisible) {
      return (await inlineError.textContent()) ?? '';
    }
    return '';
  }
}
