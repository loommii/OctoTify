# OctoTify 前端开发规范

> 基于 Vue 3 + Vite + TypeScript 企业级最佳实践

## 目录

- [技术栈](#技术栈)
- [项目结构](#项目结构)
- [编码规范](#编码规范)
- [组件开发规范](#组件开发规范)
- [状态管理规范](#状态管理规范)
- [路由规范](#路由规范)
- [样式规范](#样式规范)
- [TypeScript 规范](#typescript-规范)
- [工具与插件](#工具与插件)
- [开发流程](#开发流程)
- [测试规范](#测试规范)
- [性能优化](#性能优化)

---

## 技术栈

### 核心技术

| 技术 | 版本 | 用途 |
|------|------|------|
| Vue | 3.4+ | 前端框架 |
| Vite | 4.3+ | 构建工具 |
| TypeScript | 5.0+ | 类型安全 |
| Vue Router | 4.2+ | 路由管理 |
| Pinia | 2.0+ | 状态管理 |
| SCSS | 1.77+ | CSS 预处理器 |

### 推荐工具库

| 库 | 用途 |
|------|------|
| @vueuse/core | Vue 组合式工具集 |
| @unhead/vue | SEO 和头部标签管理 |
| unplugin-auto-import | API 自动导入 |
| unplugin-vue-components | 组件自动注册 |

### 开发工具

| 工具 | 用途 |
|------|------|
| ESLint | 代码质量检查 |
| Prettier | 代码格式化 |
| Vitest | 单元测试 |
| Playwright | E2E 测试 |
| VitePress | 文档生成 |

---

## 项目结构

### 目录结构

```
src/
├── assets/          # 静态资源(图片、字体等)
├── components/      # 可复用组件
│   ├── Base*.vue    # 基础组件(全局注册)
│   └── *.spec.ts    # 组件单元测试
├── composables/     # 组合式函数
│   └── use*.ts      # 以 use 开头
├── design/          # 设计系统
│   ├── index.scss   # 设计变量入口
│   ├── _colors.scss       # 颜色变量
│   ├── _typography.scss   # 排版变量
│   ├── _sizes.scss        # 尺寸变量
│   ├── _fonts.scss        # 字体导入
│   ├── _durations.scss    # 动画时长
│   └── _layers.scss       # z-index 管理
├── layouts/         # 布局组件
│   └── *Layout.vue
├── pages/           # 页面组件(对应路由)
├── router/          # 路由配置
│   ├── index.ts     # 路由实例
│   └── routes.ts    # 路由定义
├── stores/          # Pinia 状态管理
│   └── *Store.ts    # 按领域划分
├── App.vue          # 根组件
├── main.ts          # 应用入口
└── types.ts         # 全局类型定义
```

### 配置文件

| 文件 | 用途 |
|------|------|
| vite.config.ts | Vite 构建配置 |
| tsconfig.json | TypeScript 根配置 |
| tsconfig.app.json | 应用代码配置 |
| tsconfig.node.json | Node.js 工具配置 |
| tsconfig.vitest.json | 测试配置 |
| .eslintrc.cjs | ESLint 规则 |
| .prettierrc.json | Prettier 规则 |
| vitest.config.ts | 单元测试配置 |
| playwright.config.ts | E2E 测试配置 |
| env.d.ts | 环境变量类型声明 |

---

## 编码规范

### 包管理器

**使用 pnpm** (推荐)

```bash
# 安装依赖
pnpm install

# 添加依赖
pnpm add <package>

# 添加开发依赖
pnpm add -D <package>
```

**为什么选择 pnpm:**
- 更快的安装速度(符号链接和内容可寻址存储)
- 磁盘空间高效(全局存储,不重复)
- 严格的依赖管理(防止"幽灵依赖")
- Vue 生态标准(Vue、Vite、Nuxt 都使用 pnpm)

### NPM Scripts

```json
{
  "scripts": {
    "dev": "vite",
    "build": "run-p type-check build-only",
    "preview": "vite preview",
    "test:unit": "vitest",
    "test:e2e": "playwright test",
    "build-only": "vite build",
    "type-check": "vue-tsc --noEmit -p tsconfig.vitest.json --composite false",
    "lint": "eslint . --ext .vue,.js,.jsx,.cjs,.mjs,.ts,.tsx,.cts,.mts --fix --ignore-path .gitignore",
    "format": "prettier --write src/"
  }
}
```

### ESLint 配置

```javascript
// .eslintrc.cjs
module.exports = {
  root: true,
  extends: [
    'plugin:vue/vue3-essential',
    'eslint:recommended',
    '@vue/eslint-config-typescript',
    '@vue/eslint-config-prettier/skip-formatting'
  ],
  parserOptions: {
    ecmaVersion: 'latest'
  }
}
```

**扩展规则说明:**
- `plugin:vue/vue3-essential` - Vue 3 特定规则
- `eslint:recommended` - 核心 ESLint 规则
- `@vue/eslint-config-typescript` - TypeScript 支持
- `@vue/eslint-config-prettier/skip-formatting` - 禁用格式化规则(Prettier 处理)

### Prettier 配置

```json
{
  "semi": false,
  "tabWidth": 2,
  "singleQuote": true,
  "printWidth": 100,
  "trailingComma": "none"
}
```

**规则说明:**
- 不使用分号(依赖 ASI)
- 2 空格缩进
- 单引号
- 每行最大 100 字符
- 不使用尾逗号

### 语言覆盖

| 语言 | 工具 |
|------|------|
| TypeScript/JavaScript | ESLint + Prettier |
| Vue SFC | ESLint (eslint-plugin-vue) + Prettier |
| JSON/HTML/CSS/Markdown | Prettier |

---

## 组件开发规范

### Vue 组件语法

**使用 `<script setup>` + Composition API**

```vue
<script setup lang="ts">
import { ref, computed } from 'vue'

const count = ref(0)
const doubled = computed(() => count.value * 2)

function increment() {
  count.value++
}
</script>

<template>
  <div>
    <p>{{ count }}</p>
    <button @click="increment">增加</button>
  </div>
</template>

<style lang="scss" scoped>
/* 组件样式 */
</style>
```

**为什么使用 Composition API:**
- 更好的 TypeScript 集成
- Composables 比 mixins 更灵活
- 符合 Vue 生态发展方向
- 更好的代码组织方式

### 组件命名规范

#### 基础组件 (Base Components)

以 `Base` 前缀命名,用于全局注册的可复用 UI 组件:

```
BaseButton.vue
BaseInputText.vue
BaseModal.vue
BaseCard.vue
```

**特点:**
- 只负责 UI 展示
- 不包含业务逻辑
- 应用特定的样式
- 自动全局注册

#### 业务组件

使用描述性名称,反映其功能:

```
UserProfile.vue
ProductCard.vue
SearchFilter.vue
```

#### 页面组件

放在 `pages/` 目录,对应路由:

```
pages/index.vue
pages/about.vue
pages/user/profile.vue
```

### 组件结构顺序

```vue
<script setup lang="ts">
// 1. 导入
import { ref, computed } from 'vue'

// 2. Props 定义
const props = defineProps<{
  title: string
  count?: number
}>()

// 3. Emits 定义
const emit = defineEmits<{
  update: [value: string]
  delete: []
}>()

// 4. 响应式状态
const localState = ref('')

// 5. 计算属性
const computedValue = computed(() => props.title)

// 6. 方法
function handleClick() {
  emit('update', localState.value)
}

// 7. 生命周期
onMounted(() => {
  // 初始化逻辑
})
</script>

<template>
  <!-- 模板内容 -->
</template>

<style lang="scss" scoped>
/* 样式 */
</style>
```

### Props 定义

**使用 TypeScript 类型定义:**

```vue
<script setup lang="ts">
// 推荐: 使用 TypeScript 类型
const props = defineProps<{
  type: 'primary' | 'secondary'
  disabled?: boolean
  onClick?: () => void
}>()
</script>
```

**或者使用运行时声明(需要类型验证时):**

```vue
<script setup lang="ts">
defineProps({
  type: {
    type: String,
    default: 'text',
    validator: (value: string) =>
      ['email', 'number', 'password', 'text'].includes(value)
  }
})

const model = defineModel()
</script>
```

### Slots 使用

```vue
<template>
  <button :class="$style.button">
    <slot>默认内容</slot>
  </button>
</template>
```

**具名 Slot:**

```vue
<template>
  <div>
    <slot name="header" />
    <slot />
    <slot name="footer" />
  </div>
</template>
```

### CSS Modules

使用 `module` 替代 `scoped` 获取更好的样式隔离:

```vue
<template>
  <button :class="$style.button">
    <slot>Submit</slot>
  </button>
</template>

<style lang="scss" module>
.button {
  @extend %typography-small;
  padding: $size-button-padding;
  border: none;
  background: $color-button-bg;
  color: $color-button-text;
  cursor: pointer;

  &:disabled {
    cursor: not-allowed;
    background: $color-button-disabled-bg;
  }
}
</style>
```

---

## 状态管理规范

### Pinia Store

**使用 Composition API 风格 (Setup Store):**

```typescript
// src/stores/authStore.ts
import { ref, computed } from 'vue'
import { defineStore } from 'pinia'

export const useAuthStore = defineStore('authStore', () => {
  // 状态
  const user = ref<User | null>(null)
  const token = ref<string | null>(null)

  // 计算属性
  const isAuthenticated = computed(() => !!token.value)

  // 方法
  async function login(credentials: Credentials) {
    const response = await api.login(credentials)
    user.value = response.user
    token.value = response.token
  }

  function logout() {
    user.value = null
    token.value = null
  }

  return { user, token, isAuthenticated, login, logout }
})
```

### Store 命名规范

| 元素 | 规范 | 示例 |
|------|------|------|
| 文件名 | `[domain]Store.ts` | `authStore.ts`, `cartStore.ts` |
| 导出名 | `use[Domain]Store` | `useAuthStore`, `useCartStore` |
| Store ID | `[domain]Store` | `'authStore'`, `'cartStore'` |

**避免使用单个单词** - 明确说明 store 管理的领域:
- ✅ `userProfileStore`
- ❌ `user`
- ✅ `shoppingCartStore`
- ❌ `cart`

### Store 使用

**在组件中使用:**

```vue
<script setup lang="ts">
import { useAuthStore } from '@/stores/authStore'

const authStore = useAuthStore()
</script>

<template>
  <div v-if="authStore.isAuthenticated">
    <p>欢迎, {{ authStore.user?.name }}</p>
    <button @click="authStore.logout">退出</button>
  </div>
</template>
```

**解构时保持响应性:**

```vue
<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useAuthStore } from '@/stores/authStore'

const authStore = useAuthStore()

// 状态和计算属性使用 storeToRefs
const { user, isAuthenticated } = storeToRefs(authStore)

// 方法不需要 storeToRefs
const { login, logout } = authStore
</script>
```

### Store 组织

```
src/stores/
├── index.ts              # 统一导出 + resetAllStores()
├── authStore.ts          # 认证状态（登录/登出/刷新 Token）
├── userStore.ts          # 用户信息（个人资料/修改密码）
├── sourceStore.ts        # Source 管理（CRUD/启用/禁用/Token 管理）
├── channelStore.ts       # Channel 管理（CRUD/启用/禁用/测试/渠道类型元数据）
└── messageStore.ts       # 消息管理（列表/筛选/详情/删除）
```

### 统一重置

登出时需要重置所有 Store，通过 `stores/index.ts` 统一管理：

```typescript
// src/stores/index.ts
export function resetAllStores() {
  useAuthStore().$reset()
  useUserStore().$reset()
  useSourceStore().$reset()
  useChannelStore().$reset()
  useMessageStore().$reset()
}
```

**注意：** 低频变动的元数据（如 `channelTypes`）在 `$reset()` 中保留不重置，优化体验。

### Store 组合

Store 可以使用其他 store:

```typescript
// src/stores/sourceStore.ts
import { defineStore } from 'pinia'
import { useAuthStore } from './authStore'

export const useSourceStore = defineStore('sourceStore', () => {
  const authStore = useAuthStore()

  async function fetchList() {
    if (!authStore.isAuthenticated) {
      throw new Error('必须登录才能查询')
    }
    // 查询逻辑
  }

  return { fetchList }
})
```

### 何时使用 Store

**使用 Pinia Store:**
- 认证状态(用户、令牌)
- 多组件共享数据
- 跨路由持久化的数据
- 复杂状态(多个方法)

**使用组件本地状态:**
- 表单输入
- UI 状态(模态框、下拉菜单)
- 单组件使用的数据
- 导航时重置的临时状态

**经验法则:** 从本地状态开始,当需要跨组件共享或跨路由持久化时提取到 store。

---

## 路由规范

### 路由配置

**分离路由定义和实例:**

```typescript
// src/router/routes.ts
export default [
  {
    path: '/',
    name: 'home',
    component: () => import('@/pages/index.vue')
  },
  {
    path: '/about',
    name: 'about',
    // 路由级别代码分割
    // 为这个路由生成单独的 chunk (About.[hash].js)
    // 访问时懒加载
    component: () => import('@/pages/about.vue')
  }
]
```

```typescript
// src/router/index.ts
import { createRouter, createWebHistory } from 'vue-router'
import routes from './routes'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

export default router
```

### 路由懒加载

**使用动态导入实现代码分割:**

```typescript
// ✅ 推荐: 懒加载
component: () => import('@/pages/about.vue')

// ❌ 不推荐: 同步导入
import AboutPage from '@/pages/about.vue'
component: AboutPage
```

### 路由命名

- 使用小写字母和连字符
- 名称应该描述页面内容
- 避免使用动词

```typescript
// ✅ 推荐
{ path: '/user-profile', name: 'user-profile' }
{ path: '/product-list', name: 'product-list' }

// ❌ 不推荐
{ path: '/getUserProfile', name: 'getUserProfile' }
{ path: '/products', name: 'showProducts' }
```

---

## 样式规范

### SCSS 使用

**使用 `lang="scss"` 启用 SCSS:**

```vue
<style lang="scss" scoped>
.container {
  padding: 1rem;
  
  .header {
    font-size: 1.5rem;
  }
}
</style>
```

### 设计变量

**集中管理设计令牌:**

```scss
// src/design/_colors.scss
$color-body-bg: #f9f7f5;
$color-text: #444;
$color-heading-text: #35495e;
$color-link-text: #39a275;
$color-input-border: lighten($color-heading-text, 50%);
$color-button-bg: $color-link-text;
$color-button-text: white;
```

```scss
// src/design/_sizes.scss
$size-grid-padding: 1.3rem;
$size-input-padding-vertical: 0.75em;
$size-input-padding-horizontal: 1em;
$size-button-padding-vertical: calc($size-grid-padding / 2);
```

```scss
// src/design/index.scss
@import 'colors';
@import 'sizes';
@import 'typography';
@import 'fonts';
@import 'durations';
@import 'layers';
```

### Vite 中全局注入 SCSS 变量

```typescript
// vite.config.ts
export default defineConfig({
  css: {
    preprocessorOptions: {
      scss: {
        additionalData: '@use "@/design/index.scss" as *;'
      }
    }
  }
})
```

**然后在组件中直接使用变量:**

```vue
<style lang="scss" scoped>
.button {
  background: $color-button-bg;
  padding: $size-button-padding;
}
</style>
```

### Scoped 样式

**使用 `scoped` 确保样式只应用于当前组件:**

```vue
<style scoped>
/* 这个 .button 只影响当前组件 */
.button {
  background: blue;
}
</style>
```

Vue 添加唯一属性(如 `data-v-7ba5bd90`)到元素和选择器,防止样式泄漏。

### 深度选择器

**使用 `:deep()` 样式化子组件元素:**

```vue
<style scoped>
/* 样式化子组件内部的元素 */
:deep(.child-class) {
  color: red;
}
</style>
```

**谨慎使用** - 优先通过 props 或 classes 传递给子组件。

### 全局 CSS

**全局样式应该最小化,只包含:**
- CSS 重置或规范化
- 基础元素样式(排版、链接)
- 工具类(如果不使用工具类框架)

```typescript
// src/main.ts
import './styles/global.scss'
```

### 为什么使用 Scoped 而不是 CSS Modules?

- Scoped 样式更简单,对大多数用例足够
- CSS Modules 提供更好的冲突保护但增加复杂性
- Scoped 样式提供了正确的平衡点

### 为什么使用 SCSS 而不是纯 CSS?

- SCSS 提供变量、嵌套和 mixins,使样式更可维护
- SCSS 是 CSS 的超集,学习曲线最小
- 任何有效的 CSS 都是有效的 SCSS

---

## TypeScript 规范

### 配置

```json
// tsconfig.json
{
  "files": [],
  "references": [
    { "path": "./tsconfig.node.json" },
    { "path": "./tsconfig.app.json" },
    { "path": "./tsconfig.vitest.json" }
  ]
}
```

```json
// tsconfig.app.json
{
  "extends": "@vue/tsconfig/tsconfig.dom.json",
  "include": ["auto-imports.d.ts", "component.d.ts", "env.d.ts", "src/**/*", "src/**/*.vue"],
  "exclude": ["src/**/__tests__/*"],
  "compilerOptions": {
    "composite": true,
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    }
  }
}
```

### 严格模式

```json
{
  "compilerOptions": {
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true
  }
}
```

### Interface vs Type

**使用 `interface` 定义对象形状(可能扩展):**

```typescript
interface User {
  id: string
  name: string
  email: string
}
```

**使用 `type` 定义联合类型、原始类型和工具类型:**

```typescript
type Status = 'pending' | 'active' | 'completed'
type ID = string | number
```

### 类型严格度

**目标: 完全类型覆盖,但实用主义优于完美主义**

- 少量使用 `any`,使用时记录原因
- `unknown` 类型通常是更好的选择,当类型真正未知时

### 从 JavaScript 迁移

**渐进式迁移策略:**

```json
{
  "compilerOptions": {
    "allowJs": true,
    "checkJs": false
  }
}
```

1. 启用 `allowJs`,禁用 `checkJs`
2. 一次一个文件从 `.js` 重命名为 `.ts`
3. 转换时修复每个文件的类型错误
4. 所有文件转换完成后,禁用 `allowJs`

---

## 工具与插件

### Vite 配置

```typescript
// vite.config.ts
import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'

export default defineConfig({
  plugins: [
    vue(),
    AutoImport({
      // 全局导入
      imports: ['vue', 'vue-router', '@vueuse/core', { '@unhead/vue': ['useHead'] }],
      dirs: ['@src/composables']
    }),
    Components({
      dirs: ['src/components', 'src/layouts']
    })
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    port: 8080
  },
  css: {
    preprocessorOptions: {
      scss: {
        additionalData: '@use "@/design/index.scss" as *;'
      }
    }
  }
})
```

### 路径别名

| 别名 | 路径 |
|------|------|
| `@` | `src/` |

**使用示例:**

```typescript
// ❌ 不推荐: 相对路径
import { useTheme } from '../../../composables/useTheme'

// ✅ 推荐: 使用别名
import { useTheme } from '@/composables/useTheme'
```

**添加新别名时,更新两个文件:**
- `vite.config.ts` - Vite 解析
- `tsconfig.app.json` - TypeScript 和 IDE 支持

### 自动导入

#### 组件自动注册

`src/components/` 和 `src/layouts/` 中的组件自动全局注册:

```vue
<template>
  <!-- BaseButton 从 src/components/BaseButton.vue 自动导入 -->
  <BaseButton>点击我</BaseButton>
  
  <!-- AppLayout 从 src/layouts/AppLayout.vue 自动导入 -->
  <AppLayout>
    <router-view />
  </AppLayout>
</template>

<script setup lang="ts">
// 不需要导入!
</script>
```

#### Vue API 自动导入

以下 API 自动导入,不需要手动导入:

```vue
<script setup lang="ts">
// 这些是自动导入的 - 不需要 import 语句
const count = ref(0)
const doubled = computed(() => count.value * 2)

watch(count, (newVal) => {
  console.log('Count 变化:', newVal)
})

onMounted(() => {
  console.log('组件已挂载')
})
</script>
```

**自动导入的 API:**
- Vue: `ref`, `computed`, `watch`, `onMounted`, 等
- Vue Router: `useRouter`, `useRoute`
- Pinia: `defineStore`, `storeToRefs`
- VueUse: 各种 composables

生成的类型声明在 `auto-imports.d.ts` 和 `components.d.ts`。

### Composables

**组合式函数封装可复用的有状态逻辑:**

```typescript
// src/composables/useTheme.ts
import { useColorMode } from '@vueuse/core'

type Theme = 'dark' | 'light'

function useTheme() {
  const themePreference = useColorMode()

  function setTheme(theme: Theme) {
    themePreference.value = theme
  }

  return { setTheme, themePreference }
}

export default useTheme
```

**命名规范:**
- 文件名: `use*.ts`
- 函数名: `use*`
- 返回: 对象包含状态和方法

**Composables 是 Vue 3 中 mixins 的替代品,提供更好的 TypeScript 支持和显式依赖。**

---

## 开发流程

### 首次设置

**确保已安装:**
- Node.js (至少最新 LTS)
- pnpm (推荐包管理器)

**安装 pnpm:**

```bash
# 使用 npm
npm install -g pnpm

# 或使用 Corepack (Node 16.13+ 包含)
corepack enable
corepack prepare pnpm@latest --activate
```

### 安装依赖

```bash
# 从 package.json 安装依赖
pnpm install
```

### 开发服务器

```bash
# 启动开发服务器
pnpm dev

# 启动开发服务器并自动打开浏览器
pnpm dev --open
```

Vite 的开发服务器几乎瞬间启动,具有闪电般的 HMR(热模块替换)。

### 环境变量

**开发时使用生产 API:**

```bash
# .env.local
VITE_API_BASE_URL=http://localhost:34123
```

**或在命令行设置:**

```bash
# 开发时使用本地后端
VITE_API_BASE_URL=http://localhost:3000 pnpm dev

# 开发时使用生产服务器
VITE_API_BASE_URL=https://api.example.com pnpm dev
```

**在代码中访问:**

```typescript
const apiBaseUrl = import.meta.env.VITE_API_BASE_URL
```

### OpenAPI 代码生成

**使用 @hey-api/openapi-ts 自动生成 API 客户端代码:**

```bash
# 生成 API 客户端和类型
pnpm api:generate
```

**配置:**

```typescript
// openapi-ts.config.ts
import { defineConfig } from '@hey-api/openapi-ts'

export default defineConfig({
  input: 'http://localhost:34123/openapi.json',
  output: 'src/api/generated',
  plugins: [
    '@hey-api/client-axios',
    '@hey-api/typescript',
    {
      name: '@hey-api/sdk',
      operations: { strategy: 'single' },
    },
  ],
})
```

**生成文件:**

| 文件 | 说明 |
|------|------|
| `src/api/generated/client.gen.ts` | Axios 客户端实例 |
| `src/api/generated/sdk.gen.ts` | API 请求函数 (SDK) |
| `src/api/generated/types.gen.ts` | TypeScript 类型定义 |

### API 层架构

**目录结构:**

```
src/api/
├── index.ts           # API 客户端配置和拦截器
└── generated/         # 自动生成的代码 (不要手动修改)
    ├── client.gen.ts  # Axios 客户端
    ├── sdk.gen.ts     # API 请求函数
    └── types.gen.ts   # 类型定义
```

**API 客户端配置:**

```typescript
// src/api/index.ts
import { createClient } from './generated/client'

export const apiClient = createClient({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 15000,
})
```

**请求拦截器 (自动添加 Token):**

```typescript
apiClient.instance.interceptors.request.use(
  async (config) => {
    const { accessToken } = useAuth()
    if (accessToken.value) {
      config.headers.Authorization = `Bearer ${accessToken.value}`
    }
    return config
  },
  (error) => Promise.reject(error),
)
```

**响应拦截器 (统一错误处理 + Token 刷新):**

```typescript
apiClient.instance.interceptors.response.use(
  // 响应成功: 检查业务状态码
  (response) => {
    const res = response.data
    if (res && typeof res === 'object' && 'code' in res) {
      if (res.code !== 0) {
        const error = new Error(res.msg || '请求失败')
        ;(error as any).code = res.code
        ;(error as any).response = res
        return Promise.reject(error)
      }
    }
    return response
  },
  // 响应错误: 处理 401 自动刷新 Token
  async (error: AxiosError) => {
    const originalRequest = error.config as any

    if (error.response?.status === 401 && !originalRequest.__isRetryRequest) {
      // 如果正在刷新,加入队列等待
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject })
        })
          .then((token: unknown) => {
            originalRequest.headers.Authorization = `Bearer ${token}`
            return apiClient.instance(originalRequest)
          })
          .catch((err: unknown) => Promise.reject(err))
      }

      // 开始刷新 Token
      originalRequest.__isRetryRequest = true
      isRefreshing = true

      try {
        const authStore = useAuthStore()
        const newToken = await authStore.refreshAccessToken()
        processQueue(null, newToken)
        originalRequest.headers.Authorization = `Bearer ${newToken}`
        return apiClient.instance(originalRequest)
      } catch (refreshError) {
        processQueue(refreshError, null)
        const { clearAuth } = useAuth()
        clearAuth()
        window.location.href = '/login'
        return Promise.reject(refreshError)
      } finally {
        isRefreshing = false
      }
    }

    const responseData = error.response?.data as { msg?: string } | undefined
    const errorMsg = responseData?.msg || error.message || '请求失败'
    console.error('API Error:', errorMsg)

    return Promise.reject(error)
  },
)
```

**业务状态码规范:**

| 状态码 | 说明 |
|--------|------|
| `0` | 成功 |
| 非 `0` | 失败,`msg` 字段包含错误信息 |

### 认证系统

**认证架构 (参考 vue-vben-admin):**

| 问题 | 方案 |
|------|------|
| Token 存储 | `localStorage` + Composable 响应式 |
| Token Key | `accessToken` / `refreshToken` |
| 错误提示 | `msg` 字段 + `console.error` |
| 401 处理 | 自动刷新 Token,刷新失败跳转登录 |

**Composable (useAuth.ts):**

```typescript
// src/composables/useAuth.ts
import { ref, computed } from 'vue'
import { apiClient } from '@/api'

const ACCESS_TOKEN_KEY = 'accessToken'
const REFRESH_TOKEN_KEY = 'refreshToken'

const accessToken = ref<string | null>(null)
const refreshToken = ref<string | null>(null)
const isAuthenticated = computed(() => !!accessToken.value)

export function useAuth() {
  function setAccessToken(token: string) {
    accessToken.value = token
    localStorage.setItem(ACCESS_TOKEN_KEY, token)
    apiClient.setConfig({
      headers: { Authorization: `Bearer ${token}` },
    })
  }

  function setRefreshToken(token: string) {
    refreshToken.value = token
    localStorage.setItem(REFRESH_TOKEN_KEY, token)
  }

  function clearAuth() {
    accessToken.value = null
    refreshToken.value = null
    localStorage.removeItem(ACCESS_TOKEN_KEY)
    localStorage.removeItem(REFRESH_TOKEN_KEY)
    apiClient.setConfig({
      headers: { Authorization: undefined },
    })
  }

  function initAuth() {
    const storedAccessToken = localStorage.getItem(ACCESS_TOKEN_KEY)
    const storedRefreshToken = localStorage.getItem(REFRESH_TOKEN_KEY)

    if (storedAccessToken) {
      accessToken.value = storedAccessToken
      apiClient.setConfig({
        headers: { Authorization: `Bearer ${storedAccessToken}` },
      })
    }

    if (storedRefreshToken) {
      refreshToken.value = storedRefreshToken
    }
  }

  return {
    accessToken,
    refreshToken,
    isAuthenticated,
    setAccessToken,
    setRefreshToken,
    clearAuth,
    initAuth,
  }
}
```

**Pinia Store (authStore.ts):**

```typescript
// src/stores/authStore.ts
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useAuth } from '@/composables/useAuth'
import { Sdk } from '@/api/generated/sdk.gen'

export const useAuthStore = defineStore('authStore', () => {
  const sdk = new Sdk()
  const { setAccessToken, setRefreshToken, clearAuth } = useAuth()
  const loginLoading = ref(false)

  async function login(username: string, password: string) {
    try {
      loginLoading.value = true
      const response = await sdk.login({
        body: { AuthCredentials: { username, password } },
      })

      if (response.data && response.data.code === 0 && response.data.data) {
        const data = response.data.data as {
          access_token: string
          refresh_token: string
          user: { id: number; username: string }
        }

        setAccessToken(data.access_token)
        setRefreshToken(data.refresh_token)

        return { success: true, user: data.user }
      }

      return {
        success: false,
        message: (response.data as { msg?: string })?.msg || '登录失败',
      }
    } catch {
      return { success: false, message: '网络错误,请稍后重试' }
    } finally {
      loginLoading.value = false
    }
  }

  async function logout() {
    try {
      await sdk.logout()
    } catch {
      // 忽略错误
    }
    clearAuth()
  }

  async function refreshAccessToken() {
    const { refreshToken } = useAuth()
    if (!refreshToken.value) {
      throw new Error('No refresh token available')
    }

    const response = await sdk.refreshToken({
      body: { refresh_token: refreshToken.value },
    })

    if (response.data && response.data.code === 0 && response.data.data) {
      const data = response.data.data as {
        access_token: string
        refresh_token: string
      }

      setAccessToken(data.access_token)
      setRefreshToken(data.refresh_token)

      return data.access_token
    }

    throw new Error('Refresh token failed')
  }

  return { login, logout, refreshAccessToken, loginLoading }
})
```

**应用初始化 (main.ts):**

```typescript
// src/main.ts
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'
import { createHead } from '@unhead/vue/client'

import App from '@/App.vue'
import router from '@/router/index'
import { useAuth } from '@/composables/useAuth'

const app = createApp(App)

const pinia = createPinia()
pinia.use(piniaPluginPersistedstate)
app.use(pinia)

app.use(router)

const head = createHead()
app.use(head)

// 初始化认证状态
const { initAuth } = useAuth()
initAuth()

app.mount('#app')
```

### 构建

```bash
# 类型检查 + 构建
pnpm build

# 仅构建
pnpm build-only

# 类型检查
pnpm type-check

# 预览生产构建
pnpm preview
```

### Lint 和 Format

```bash
# Lint 并自动修复
pnpm lint

# 格式化 src/ 中所有文件
pnpm format
```

### VS Code 推荐设置

**自动:**
- 保存时 Lint (ESLint)
- 保存时格式化 (Prettier)

**安装扩展:**
- ESLint
- Prettier
- Vue - Official
- TypeScript Vue Plugin

---

## 测试规范

### 单元测试 (Vitest)

**组件测试与组件放在一起:**

```typescript
// src/components/BaseButton.spec.ts
import { describe, it, expect } from 'vitest'
import { shallowMount } from '@vue/test-utils'
import BaseButton from './BaseButton.vue'

describe('BaseButton 组件', () => {
  it('渲染其内容', () => {
    const slotContent = '<strong>点击我!</strong>'
    const { element } = shallowMount(BaseButton, {
      slots: {
        default: slotContent
      }
    })
    expect(element.innerHTML).toContain(slotContent)
  })

  it('渲染默认内容', () => {
    const slotContent = ''
    const { element } = shallowMount(BaseButton, {
      slots: {
        default: slotContent
      }
    })
    expect(element.innerHTML).toContain('Submit')
  })
})
```

**运行单元测试:**

```bash
pnpm test:unit
```

### E2E 测试 (Playwright)

```bash
# 运行 E2E 测试
pnpm test:e2e

# 运行 E2E 测试带 UI
pnpm test:e2e:ui
```

### 测试文件组织

```
src/
├── components/
│   ├── BaseButton.vue
│   └── BaseButton.spec.ts    # 单元测试
│   ├── BaseInputText.vue
│   └── BaseInputText.spec.ts # 单元测试
```

---

## 性能优化

### 路由懒加载

```typescript
// ✅ 推荐: 动态导入
component: () => import('@/pages/about.vue')

// ❌ 不推荐: 同步导入
import AboutPage from '@/pages/about.vue'
component: AboutPage
```

### 组件懒加载

**对于大型组件或条件渲染的组件:**

```vue
<script setup lang="ts">
import { defineAsyncComponent } from 'vue'

const HeavyComponent = defineAsyncComponent(() =>
  import('@/components/HeavyComponent.vue')
)
</script>
```

### 计算属性缓存

```vue
<script setup lang="ts">
const count = ref(0)

// ✅ 推荐: 使用 computed 缓存结果
const doubled = computed(() => count.value * 2)

// ❌ 不推荐: 每次渲染都重新计算
function getDoubled() {
  return count.value * 2
}
</script>
```

### v-for 使用 key

```vue
<template>
  <!-- ✅ 推荐: 使用唯一 ID -->
  <div v-for="item in items" :key="item.id">
    {{ item.name }}
  </div>
  
  <!-- ❌ 不推荐: 使用索引 -->
  <div v-for="(item, index) in items" :key="index">
    {{ item.name }}
  </div>
</template>
```

### 条件渲染优化

```vue
<template>
  <!-- ✅ 推荐: v-if 用于条件渲染 -->
  <div v-if="isVisible">内容</div>
  
  <!-- ✅ 推荐: v-show 用于频繁切换 -->
  <div v-show="isVisible">内容</div>
</template>
```

---

## 附录

### Vue 3 风格指南

参考 [Vue 风格指南](https://vuejs.org/style-guide/)

**优先级 A (必要):**
- 组件名多单词
- 使用详细 props 定义
- v-for 设置 key
- 避免 v-if 和 v-for 一起使用

**优先级 B (强烈推荐):**
- 基础组件名以 Base 前缀
- 紧密耦合的组件使用单文件
- 组件/实例名使用 PascalCase
- Props 使用 camelCase

**优先级 C (推荐):**
- 多个 attribute 元素多行
- 组件模板简单表达式

### 常见问答

**为什么 ESLint 和 Prettier 分开?**

- ESLint 最擅长捕获错误和执行代码质量(未使用变量、缺少返回等)
- Prettier 最擅长格式化(缩进、换行、引号)
- 使用两者可以获得各自工具的最佳效果

**为什么不使用分号?**

- 这是风格选择
- JavaScript 的 ASI(自动分号插入)处理大多数情况
- 省略分号减少视觉噪音
- 如果偏好分号,更改 `.prettierrc.json` 中 `"semi": true`

**如何添加 Stylelint 用于 CSS?**

```bash
pnpm add -D stylelint stylelint-config-standard-scss
```

创建 `stylelint.config.js`:
```javascript
export default {
  extends: ['stylelint-config-standard-scss']
}
```

**为什么这么多配置文件?为什么不移到 `package.json`?**

- 单独文件更容易找到和编辑
- 支持通过 JavaScript 动态配置
- 更好的 IDE 支持和语法高亮
- 更清晰的 `package.json`,专注于依赖和脚本

### 参考资料

- [Vue 3 文档](https://vuejs.org/)
- [Vite 文档](https://vitejs.dev/)
- [Vue Router 文档](https://router.vuejs.org/)
- [Pinia 文档](https://pinia.vuejs.org/)
- [TypeScript 文档](https://www.typescriptlang.org/)
- [VueUse 文档](https://vueuse.org/)
- [ESLint 文档](https://eslint.org/)
- [Prettier 文档](https://prettier.io/)
