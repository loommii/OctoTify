import { type Page, type Locator } from '@playwright/test';
import { BasePage } from './BasePage';

/**
 * 渠道详情页面对象 - /channel/detail/:id
 *
 * 对应前端页面：推送渠道详情页
 * 包含：基本信息展示、渠道配置展示、测试连接/修改操作
 */
export class ChannelDetailPage extends BasePage {
  readonly nameDisplay: Locator;
  readonly typeDisplay: Locator;
  readonly configDisplay: Locator;
  readonly statusTag: Locator;
  readonly testButton: Locator;
  readonly editButton: Locator;

  constructor(page: Page) {
    super(page);
    // 渠道名称（从描述表格的"渠道名称"行读取值）
    this.nameDisplay = page.getByRole('row', { name: '渠道名称' }).getByRole('cell').nth(1);
    // 渠道类型（从描述表格的"渠道类型"行读取标签）
    this.typeDisplay = page.getByRole('row', { name: '渠道类型' }).getByRole('cell').nth(1).locator('.el-tag');
    // 渠道配置区域
    this.configDisplay = page.locator('.el-descriptions--border');
    // 状态标签
    this.statusTag = page.getByRole('row', { name: '状态' }).getByRole('cell').nth(1).locator('.el-tag');
    // 操作按钮
    this.testButton = page.getByRole('button', { name: '测试连接' });
    this.editButton = page.getByRole('button', { name: '修改' });
  }

  /** 导航到渠道详情页 */
  async goto(id: string) {
    await super.goto(`/channel/detail/${id}`);
    await this.page.waitForLoadState('networkidle');
  }

  /** 获取渠道名称 */
  async getName(): Promise<string> {
    return (await this.nameDisplay.textContent()) ?? '';
  }

  /** 获取渠道类型 */
  async getType(): Promise<string> {
    return (await this.typeDisplay.textContent()) ?? '';
  }

  /** 获取渠道配置文本 */
  async getConfig(): Promise<string> {
    return (await this.configDisplay.textContent()) ?? '';
  }

  /** 获取状态标签文本 */
  async getStatusTag(): Promise<string> {
    return (await this.statusTag.textContent()) ?? '';
  }

  /** 点击"测试连接"按钮 */
  async clickTest() {
    await this.testButton.click();
  }

  /** 点击"修改"按钮 */
  async clickEdit() {
    await this.editButton.click();
  }
}
