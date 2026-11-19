# 用户认证模块 API 文档核对与 Mock 报告

- 项目：8214367（OctoTify），主分支 main
- 模块文件夹：84718775（用户认证）
- 接口数：3
- 任务：task-2（doc-auth）
- 核对基准：运行时 router / handler / dto / service / middleware / pkg/response / pkg/xerr / config（openapi.go 仅作参考，冲突处以运行时为准）

## 一、代码事实摘要（本模块）

| 接口 | 路由注册 | 鉴权 | 请求体 |
|---|---|---|---|
| POST /auth/login | router.go setupAuthRoutes 无中间件 | 无 | dto.LoginReq（AuthCredentials 内嵌） |
| POST /auth/refresh | 无中间件 | 无 | dto.RefreshReq |
| POST /auth/logout | middleware.JWTAuth(accessJWT) | Bearer Access Token | 无 |

- 统一响应：`{"code":0,"msg":"请求成功","data":...}`，成功 HTTP 200；`data` 为 omitempty，logout 无 data。
- 失败响应：`response.Fail` 统一 HTTP 200 + 非 0 code；仅 JWT 鉴权失败（JWTAuth）返回 HTTP 401 + 111000/111001/111002/111003，另 handler 中 100001 兜底。
- login 参数校验失败/JSON 错误统一 110200（"用户名或密码错误"，不区分原因）；refresh 同理统一 110300。
- 令牌有效期（backend/config/config.yaml）：access_ttl=1h，refresh_ttl=168h（7 天）。
- refresh 语义：校验 JWT 与 token_type=refresh → 查 RefreshToken 表（JTI）→ 已撤销 110301 → 重新签发 Access + Refresh（沿用同一 JTI，刷新过期时间，不新增记录）；旧 Refresh Token 在到期前仍可用。
- logout 语义：将该用户全部 RefreshToken 置为 revoked；Access Token 因 JWT 无状态在过期前仍有效。

## 二、逐接口核对结论

### 1) POST /auth/login（452566381，用户登录）
- 结论：**一致，未修改**
- method/path/tags(用户认证)/folderId(84718775)/name(用户登录)：正确
- 鉴权：无（auth 为空），描述"无需认证"正确
- path/query/header 参数：无，正确
- 请求体：引用 LoginReq(270530157)；username 3-64 字符、^[a-zA-Z0-9_]+$、required，password 8-64、需含大小写字母和数字、required，示例 octotify / P@ssw0rd123 —— 与 dto + pkg/validator/custom.go 完全一致
- 200 响应：Response(270530174) + data=AuthResp(270530150)，含 access_token、refresh_token、user{id,username,created_at_ts}，与 handler/service 返回一致
- 错误码：110200（用户名或密码错误，含 JSON/校验失败）、110201（登录失败/内部错误），HTTP 200 语义描述准确
- Mock：17845488

### 2) POST /auth/logout（452566382，退出登录）
- 结论：**已修复**
- 修复项：
  1. 新增 HTTP 401 响应定义（原来只有 200）。实际 JWTAuth 失败返回 401 + 111000/111001/111002/111003（或 100001），描述文本已列但响应定义缺失，已补齐（新建 response id 106259942，引用统一响应模型 270530174）。
  2. 描述补充"成功响应不包含 data 字段"（handler 为 response.Success(c, nil)）。
- 核对：requestBody=none 正确；auth=securityscheme(required) 与 JWTAuth 一致；200 响应仅 Response 包络（无 data）正确；撤销全部 Refresh Token 的描述正确；111000~111003 / 110304 / 100001 错误码与代码一致；"提交后旧 Refresh Token 刷新返回 110301" 与 service 一致
- 修改基于完整 endpoint get 结果，其他字段（method/path/tags/folderId/postProcessors/auth/parameters/原有 200 response id）均保留
- Mock：17845495

### 3) POST /auth/refresh（452566383，刷新令牌）
- 结论：**已修复**
- 修复项：
  1. 描述补充：请求体缺失、格式错误或 refresh_token 为空同样返回 110300（handler ShouldBindJSON 失败分支）
  2. 描述补充：响应 data 中同时返回用户信息（user）
  3. 错误码表 110300 说明补充"含请求体格式错误、令牌过期"（xerr 中 110302 从未被发出，JWT 过期走 110300）
- 核对：无鉴权正确；请求体引用 RefreshReq(270530163)，refresh_token 必填，与 dto 一致；200 响应 Response + AuthResp 正确；"同一 JTI 重新签发、旧 Refresh Token 到期前仍有效、已撤销返回 110301" 与 auth_service 一致；refresh TTL 7 天与 config 一致
- Mock：17845513

## 三、数据模型核对（负责范围）

| 模型 ID | 名称 | 结论 |
|---|---|---|
| 270530157 | LoginReq | 一致（字段、required、min/max、pattern、示例均与 dto.AuthCredentials + 校验器一致），未修改 |
| 270530163 | RefreshReq | 一致（refresh_token 必填、string、示例 JWT），未修改 |
| 270530150 | AuthResp | 一致（access_token/refresh_token/user，user 引用 UserDTO 270530171；DTO 无 omitempty 但文档未强制 required，不影响字段准确性），未修改 |

- 顺带核对（未修改）：UserDTO(270530171) 的 created_at_ts/id/username 与 dto.UserDTO 的 json tag 完全一致；统一响应模型 Response(270530174) 与 pkg/response.Response 一致（data 可选）。

## 四、Mock 交付（每接口恰好 1 条）

| 接口 | mockId | name | conditions | 响应 |
|---|---|---|---|---|
| POST /auth/login | 17845488 | 正常响应 | [] | HTTP 200 + Content-Type: application/json + `{"code":0,"msg":"请求成功","data":{"access_token":"eyJ...","refresh_token":"eyJ...","user":{"id":1,"username":"octotify","created_at_ts":1714636800000}}}` |
| POST /auth/logout | 17845495 | 正常响应 | [] | HTTP 200 + Content-Type: application/json + `{"code":0,"msg":"请求成功"}`（按 omitempty 省略 data） |
| POST /auth/refresh | 17845513 | 正常响应 | [] | HTTP 200 + Content-Type: application/json + `{"code":0,"msg":"请求成功","data":{"access_token":"eyJ...","refresh_token":"eyJ...","user":{"id":1,"username":"octotify","created_at_ts":1714636800000}}}` |

- 全部走 `cli-schema validate mock-create` 校验通过后创建；创建后用 `mock get <id>` 确认 bodyData 正确落库；`mock list --http-api-id` 复核各接口恰好 1 条。

## 五、遗留问题 / 需 lead 决策

1. backend/internal/server/openapi.go 的 login 描述写"密码 8-128 个字符"，与运行时 validator（8-64）冲突；本次接口文档按运行时 8-64 保留。若后续由 openapi.go 重新生成文档会覆盖本模块修正。
2. xerr 定义了 110302 ErrRefreshTokenExpired，但 auth_service 从不返回（JWT 过期统一 110300）；文档按实际行为未列 110302。如需区分"过期/无效"需改代码，超出文档范围。
3. 共享模型 Response(270530174) 的 data 未标 required（符合 omitempty 语义，未修改）；其他模块若强制 data 需自行处理。
4. logout 鉴权使用项目共享 security scheme 1229804（BearerAuth 组），未修改共享鉴权方案。
5. 本模块仅修改文档描述 + 新增 1 条 401 响应定义；未删除任何资源、未创建文件夹、未改 method/path/tags/folderId、未触碰其他模块资源。
