import { type Page, type Locator } from '@playwright/test';
import { BasePage } from './BasePage';

/**
 * 渠道列表页面对象 - /channel/list
 *
 * 对应前端页面：推送渠道列表页
 * 包含：新增、行操作（查看详情/修改/测试连接/停用/启用/删除）
 */
export class ChannelListPage extends BasePage {
  readonly createButton: Locator;
  readonly channelRows: Locator;
  readonly deleteButton: Locator;
  readonly enableButton: Locator;
  readonly disableButton: Locator;
  readonly testButton: Locator;
  readonly errorMessage: Locator;

  constructor(page: Page) {
    super(page);
    // 新增按钮
    this.createButton = page.getByRole('button', { name: '新增' });
    // 表格行
    this.channelRows = page.locator('.el-table__body tr');
    // 操作按钮
    this.deleteButton = page.getByRole('button', { name: '删除' });
    this.enableButton = page.getByRole('button', { name: '启用' });
    this.disableButton = page.getByRole('button', { name: '停用' });
    this.testButton = page.getByRole('button', { name: '测试连接' });
    // 错误提示
    this.errorMessage = page.locator('.el-message--error .el-message__content');
  }

  /** 导航到渠道列表页 */
  async goto() {
    await super.goto('/channel/list');
    await this.page.waitForLoadState('networkidle');
  }

  /** 点击新增按钮，等待跳转到创建页 */
  async clickCreate() {
    await this.createButton.click();
    await this.page.waitForURL(/\/channel\/create/);
  }

  /** 根据渠道名称获取表格行 */
  async getChannelRow(name: string): Promise<Locator> {
    return this.page.locator('tr').filter({ hasText: name });
  }

  /** 点击渠道名称进入详情页 */
  async clickDetail(name: string) {
    const row = await this.getChannelRow(name);
    await row.getByText(name).click();
  }

  /** 点击行内的"修改"按钮 */
  async clickEdit(name: string) {
    const row = await this.getChannelRow(name);
    await row.getByRole('button', { name: '修改' }).click();
  }

  /** 点击行内的"删除"按钮 */
  async clickDelete(name: string) {
    const row = await this.getChannelRow(name);
    await row.getByRole('button', { name: '删除' }).click();
  }

  /** 点击行内的"启用"按钮 */
  async clickEnable(name: string) {
    const row = await this.getChannelRow(name);
    await row.getByRole('button', { name: '启用' }).click();
  }

  /** 点击行内的"停用"按钮 */
  async clickDisable(name: string) {
    const row = await this.getChannelRow(name);
    await row.getByRole('button', { name: '停用' }).click();
  }

  /** 点击行内的"测试连接"按钮 */
  async clickTest(name: string) {
    const row = await this.getChannelRow(name);
    await row.getByRole('button', { name: '测试连接' }).click();
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
