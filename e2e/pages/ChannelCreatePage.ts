import { type Page, type Locator } from '@playwright/test';
import { BasePage } from './BasePage';

/**
 * 渠道创建页面对象 - /channel/create
 *
 * 对应前端页面：创建推送渠道页（两步向导）
 * 步骤1：选择渠道类型（卡片选择器）
 * 步骤2：填写渠道名称 + 动态配置字段
 */
export class ChannelCreatePage extends BasePage {
  readonly typeCards: Locator;
  readonly nameInput: Locator;
  readonly configFields: Locator;
  readonly submitButton: Locator;
  readonly errorMessage: Locator;

  constructor(page: Page) {
    super(page);
    // 渠道类型选择卡片
    this.typeCards = page.locator('.el-card.is-hover-shadow');
    // 渠道名称输入框
    this.nameInput = page.locator('input[placeholder*="channel name" i], input[maxlength="128"]').first();
    // 动态配置字段
    this.configFields = page.locator('.el-form-item');
    // 创建按钮
    this.submitButton = page.getByRole('button', { name: '新增' });
    // 错误提示
    this.errorMessage = page.locator('.el-message--error .el-message__content');
  }

  /** 导航到渠道创建页 */
  async goto() {
    await super.goto('/channel/create');
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * 选择渠道类型
   * @param typeName 渠道类型显示名称，如 'Telegram'、'飞书'、'钉钉'
   */
  async selectType(typeName: string) {
    await this.typeCards.filter({ hasText: typeName }).click();
  }

  /**
   * 填写渠道表单
   * @param name 渠道名称
   * @param config 渠道配置键值对，key 为表单 label 文案，value 为填充值
   */
  async fillForm(name: string, config: Record<string, string>) {
    await this.nameInput.waitFor({ state: 'visible', timeout: 10000 });
    await this.nameInput.click();
    await this.nameInput.fill(name);
    for (const [key, value] of Object.entries(config)) {
      await this.page.getByLabel(key).fill(value);
    }
  }

  /** 提交表单 */
  async submit() {
    await this.submitButton.click();
  }

  /** 获取错误提示文案 */
  async getErrorMessage(): Promise<string> {
    const toastVisible = await this.errorMessage.isVisible({ timeout: 2000 }).catch(() => false);
    if (toastVisible) {
      return (await this.errorMessage.textContent()) ?? '';
    }
    return '';
  }
}
