import { test as base } from '@playwright/test';
import { LoginPage } from '../pages/LoginPage';
import { RegisterPage } from '../pages/RegisterPage';
import { DashboardPage } from '../pages/DashboardPage';
import { ProfilePage } from '../pages/ProfilePage';
import { PasswordPage } from '../pages/PasswordPage';
import { Header } from '../pages/components/Header';
import { SourceListPage } from '../pages/SourceListPage';
import { SourceCreatePage } from '../pages/SourceCreatePage';
import { SourceDetailPage } from '../pages/SourceDetailPage';
import { SourceEditPage } from '../pages/SourceEditPage';
import { ChannelListPage } from '../pages/ChannelListPage';
import { ChannelCreatePage } from '../pages/ChannelCreatePage';
import { ChannelDetailPage } from '../pages/ChannelDetailPage';
import { AuthHelper } from '../helpers/AuthHelper';

type TestFixtures = {
  loginPage: LoginPage;
  registerPage: RegisterPage;
  dashboardPage: DashboardPage;
  profilePage: ProfilePage;
  passwordPage: PasswordPage;
  header: Header;
  sourceListPage: SourceListPage;
  sourceCreatePage: SourceCreatePage;
  sourceDetailPage: SourceDetailPage;
  sourceEditPage: SourceEditPage;
  channelListPage: ChannelListPage;
  channelCreatePage: ChannelCreatePage;
  channelDetailPage: ChannelDetailPage;
  auth: AuthHelper;
};

export const test = base.extend<TestFixtures>({
  loginPage: async ({ page }, use) => {
    await use(new LoginPage(page));
  },
  registerPage: async ({ page }, use) => {
    await use(new RegisterPage(page));
  },
  dashboardPage: async ({ page }, use) => {
    await use(new DashboardPage(page));
  },
  profilePage: async ({ page }, use) => {
    await use(new ProfilePage(page));
  },
  passwordPage: async ({ page }, use) => {
    await use(new PasswordPage(page));
  },
  header: async ({ page }, use) => {
    await use(new Header(page));
  },
  sourceListPage: async ({ page }, use) => {
    await use(new SourceListPage(page));
  },
  sourceCreatePage: async ({ page }, use) => {
    await use(new SourceCreatePage(page));
  },
  sourceDetailPage: async ({ page }, use) => {
    await use(new SourceDetailPage(page));
  },
  sourceEditPage: async ({ page }, use) => {
    await use(new SourceEditPage(page));
  },
  channelListPage: async ({ page }, use) => {
    await use(new ChannelListPage(page));
  },
  channelCreatePage: async ({ page }, use) => {
    await use(new ChannelCreatePage(page));
  },
  channelDetailPage: async ({ page }, use) => {
    await use(new ChannelDetailPage(page));
  },
  auth: async ({ page }, use) => {
    await use(new AuthHelper(page));
  },
});

export { expect } from '@playwright/test';
