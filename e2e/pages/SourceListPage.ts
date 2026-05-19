import { type Page, type Locator } from '@playwright/test';
import { BasePage } from './BasePage';

/**
 * 来源列表页面对象 - /source/list
 *
 * 对应前端页面：消息来源列表页
 * 包含：搜索、筛选、新增、行操作（查看详情/修改/停用/启用/删除）
 */
export class SourceListPage extends BasePage {
  readonly createButton: Locator;
  readonly searchInput: Locator;
  readonly statusFilter: Locator;
  readonly sourceRows: Locator;
  readonly deleteButton: Locator;
  readonly enableButton: Locator;
  readonly disableButton: Locator;
  readonly errorMessage: Locator;

  constructor(page: Page) {
    super(page);
    // 新增按钮
    this.createButton = page.getByRole('button', { name: '新增' });
    // 搜索输入框
    this.searchInput = page.locator('input[placeholder]').first();
    // 状态筛选下拉框
    this.statusFilter = page.locator('.el-select').first();
    // 表格行
    this.sourceRows = page.locator('.el-table__body tr');
    // 操作按钮
    this.deleteButton = page.getByRole('button', { name: '删除' });
    this.enableButton = page.getByRole('button', { name: '启用' });
    this.disableButton = page.getByRole('button', { name: '停用' });
    // 错误提示
    this.errorMessage = page.locator('.el-message--error .el-message__content');
  }

  /** 导航到来源列表页 */
  async goto() {
    await super.goto('/source/list');
    await this.page.waitForLoadState('networkidle');
  }

  /** 点击新增按钮，等待跳转到创建页 */
  async clickCreate() {
    await this.createButton.click();
    await this.page.waitForURL(/\/source\/create/);
  }

  /** 按名称搜索来源 */
  async searchByName(name: string) {
    await this.searchInput.fill(name);
  }

  /** 按状态筛选来源 */
  async filterByStatus(status: string) {
    await this.statusFilter.click();
    await this.page.locator('.el-select-dropdown__item').filter({ hasText: status }).click();
  }

  /** 根据来源名称获取表格行 */
  async getSourceRow(name: string): Promise<Locator> {
    return this.page.locator('tr').filter({ hasText: name }).first();
  }

  /** 点击行内的"查看详情"按钮 */
  async clickDetail(name: string) {
    const row = await this.getSourceRow(name);
    await row.getByRole('button', { name: '查看详情' }).click();
  }

  /** 点击行内的"修改"按钮 */
  async clickEdit(name: string) {
    const row = await this.getSourceRow(name);
    await row.getByRole('button', { name: '修改' }).click();
  }

  /** 点击行内的"删除"按钮 */
  async clickDelete(name: string) {
    const row = await this.getSourceRow(name);
    await row.getByRole('button', { name: '删除' }).click();
  }

  /** 点击行内的"启用"按钮 */
  async clickEnable(name: string) {
    const row = await this.getSourceRow(name);
    await row.getByRole('button', { name: '启用' }).click();
  }

  /** 点击行内的"停用"按钮 */
  async clickDisable(name: string) {
    const row = await this.getSourceRow(name);
    await row.getByRole('button', { name: '停用' }).click();
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
