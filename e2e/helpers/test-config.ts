import { readFileSync } from 'fs';
import { join } from 'path';
import { parse } from 'smol-toml';

export interface TelegramBotConfig {
  botToken: string;
  chatId: string;
  proxy?: string;
}

export function loadTestConfig(): { telegramBots: TelegramBotConfig[] } {
  const configPath = join(__dirname, '..', 'config', 'test.config.toml');
  const content = readFileSync(configPath, 'utf-8');
  return parse(content) as unknown as { telegramBots: TelegramBotConfig[] };
}