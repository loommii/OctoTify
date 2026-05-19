import { type Page, type Locator } from '@playwright/test';
import { BasePage } from './BasePage';

/**
 * 来源详情页面对象 - /source/detail/:id
 *
 * 对应前端页面：消息来源详情页
 * 包含：基本信息展示、Token 查看/重置、绑定渠道列表、启停用/删除操作
 * 部分操作需要二次密码验证（step-up auth）
 */
export class SourceDetailPage extends BasePage {
  readonly nameDisplay: Locator;
  readonly tokenDisplay: Locator;
  readonly boundChannelsList: Locator;
  readonly statusTag: Locator;
  readonly viewTokenButton: Locator;
  readonly resetTokenButton: Locator;
  readonly enableButton: Locator;
  readonly disableButton: Locator;
  readonly deleteButton: Locator;
  readonly editButton: Locator;
  readonly stepUpPasswordInput: Locator;
  readonly stepUpConfirmButton: Locator;

  constructor(page: Page) {
    super(page);
    // 来源名称（从描述表格的"来源名称"行读取值）
    this.nameDisplay = page.getByRole('row', { name: '来源名称' }).getByRole('cell').nth(1);
    // Token 显示区域
    this.tokenDisplay = page.locator('.token-input input');
    // 已绑定渠道列表
    this.boundChannelsList = page.locator('.el-table__body');
    // 状态标签
    this.statusTag = page.locator('.el-tag').first();
    // 操作按钮
    this.viewTokenButton = page.getByRole('button', { name: '查看Token' });
    this.resetTokenButton = page.getByRole('button', { name: '重置Token' });
    this.enableButton = page.getByRole('button', { name: '启用' });
    this.disableButton = page.getByRole('button', { name: '停用' });
    this.deleteButton = page.getByRole('button', { name: '删除' });
    this.editButton = page.getByRole('button', { name: '编辑' });
    // 二次密码验证弹窗
    this.stepUpPasswordInput = page.locator('.stepup-auth-dialog-content input[type="password"]');
    this.stepUpConfirmButton = page.locator('.stepup-auth-dialog-content input[type="password"]')
      .locator('..').locator('..').locator('..')
      .locator('.dialog-footer button[type="primary"]');
  }

  /** 导航到来源详情页 */
  async goto(id: string) {
    await super.goto(`/source/detail/${id}`);
    await this.page.waitForLoadState('networkidle');
  }

  /** 获取来源名称 */
  async getName(): Promise<string> {
    return (await this.nameDisplay.textContent()) ?? '';
  }

  /** 获取 Token 值 */
  async getToken(): Promise<string> {
    return (await this.tokenDisplay.inputValue()) ?? '';
  }

  /** 获取已绑定渠道列表文本 */
  async getBoundChannels(): Promise<string> {
    return (await this.boundChannelsList.textContent()) ?? '';
  }

  /** 获取状态标签文本 */
  async getStatusTag(): Promise<string> {
    return (await this.statusTag.textContent()) ?? '';
  }

  /** 点击"查看Token"按钮 */
  async clickViewToken() {
    await this.viewTokenButton.click();
  }

  /** 点击"重置Token"按钮 */
  async clickResetToken() {
    await this.resetTokenButton.click();
  }

  /** 点击"启用"按钮 */
  async clickEnable() {
    await this.enableButton.click();
  }

  /** 点击"停用"按钮 */
  async clickDisable() {
    await this.disableButton.click();
  }

  /** 点击"删除"按钮 */
  async clickDelete() {
    await this.deleteButton.click();
  }

  /** 点击"编辑"按钮 */
  async clickEdit() {
    await this.editButton.click();
  }

  /**
   * 填写二次密码验证弹窗并确认
   * @param password 当前用户的密码
   */
  async fillStepUpPassword(password: string) {
    await this.stepUpPasswordInput.fill(password);
    await this.stepUpConfirmButton.click();
  }
}
