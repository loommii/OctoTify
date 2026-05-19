import { type Page, type Locator } from '@playwright/test';

/**
 * 基础页面对象 - 所有页面对象的基类
 *
 * 提供通用的页面操作方法，各子类页面对象继承此类
 */
export class BasePage {
  protected page: Page;

  constructor(page: Page) {
    this.page = page;
  }

  /** 导航到指定路径 */
  async goto(path: string) {
    await this.page.goto(path);
  }

  /** 等待 URL 匹配指定模式 */
  async waitForUrl(pattern: string | RegExp) {
    await this.page.waitForURL(pattern);
  }

  /** 获取元素文本内容 */
  async getText(locator: Locator): Promise<string> {
    return (await locator.textContent()) ?? '';
  }

  /** 填写输入框 */
  async fill(locator: Locator, value: string) {
    await locator.fill(value);
  }

  /** 点击元素 */
  async click(locator: Locator) {
    await locator.click();
  }

  /** 检查元素是否可见 */
  async isVisible(locator: Locator): Promise<boolean> {
    return locator.isVisible();
  }

  /** 等待页面网络空闲 */
  async waitForNavigation() {
    await this.page.waitForLoadState('networkidle');
  }
}
