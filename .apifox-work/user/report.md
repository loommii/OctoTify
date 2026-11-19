# 用户管理模块 Apifox 文档核对与 Mock 报告

- 项目：8214367（OctoTify），主分支 main
- 模块/文件夹：用户管理，folderId=84718780
- 接口数：3
- 工作目录：.apifox-work/user/
- 核对依据（实际阅读）：
  - backend/internal/server/router.go（/api/user 路由与 JWTAuth 中间件）
  - backend/internal/handler/user_handler.go
  - backend/internal/service/user_service.go
  - backend/internal/handler/dto/auth_dto.go
  - backend/internal/middleware/jwt_auth.go
  - backend/pkg/validator/custom.go（username/password 自定义校验）
  - backend/pkg/xerr/codes.go（业务错误码）
  - backend/pkg/response/response.go（统一响应包络，data omitempty）
  - backend/internal/server/openapi.go（206-272 行，仅参考）
  - docs/03-API-Specification.md、docs/05-Error-Codes.md

## 一、接口核对结论与修改

### 1) PUT /user/password — 修改密码（endpointId 452566407）
- method/path/name/tag/folderId：put /user/password、修改密码、用户管理、84718780 ✅
- 鉴权：securityScheme = BearerAuth(1229804)，required=true；router 使用 middleware.JWTAuth ✅
- 请求体：application/json → ChangePasswordReq(270530151)，required=true（old_password/new_password）✅
- 200 响应：response.Response(270530174)；handler 返回 response.Success(c, nil)，data 因 omitempty 省略 ✅
- 错误码：100000（绑定失败）、110402（旧密码错误）、110403（修改/落库失败）、111000/111001/111002/111003（JWT，HTTP 401）✅
- 已修复：
  - 删除错误码表 "100001 | 未登录或 Token 已过期（HTTP 401）"：JWT 中间件实际只返回 1110xx 且 HTTP 401；handler 内 100001 分支不可达，且 response.Fail 为 HTTP 200，原表述与实现冲突。
  - 补充 "## 响应"：成功时统一包络 data 字段省略（omitempty）。
- Mock：17845486（name=正常响应，conditions=[]，bodyData={"code":0,"msg":"请求成功"}）

### 2) GET /user/profile — 查询用户信息（endpointId 452566408）
- method/path/name/tag/folderId：get /user/profile、查询用户信息、用户管理、84718780 ✅
- 鉴权：BearerAuth(1229804)，required=true ✅
- 无 path/query/header 参数；无请求体 ✅
- 200 响应：response.Response + data → UserProfileResp(270532161)；与 handler 返回 dto.UserProfileResp{User{ID,Username,CreatedAt}} 一致 ✅
- 错误码：110800（用户不存在，含 user_id 解析失败）、110801（查询失败）、1110xx（HTTP 401）✅
- 已修复：删除错误码表 "100001" 行（同上原因）。
- Mock：17845489（data={"user":{"id":1,"username":"octotify","created_at_ts":1714636800000}}）

### 3) POST /user/register — 用户注册（endpointId 452566409）
- method/path/name/tag/folderId：post /user/register、用户注册、用户管理、84718780 ✅
- 鉴权：无（securityScheme={}），与 router 一致 ✅
- 请求体：application/json → RegisterReq(270530164)，required=true（username/password）✅
- 200 响应：response.Response + data → AuthResp(270530150)，与 service 返回 dto.AuthResp{access_token,refresh_token,user} 一致 ✅
- 错误码：100000（请求体格式/绑定失败，handler 不返回字段级明细）、110104（用户名已存在）、110105（注册失败）✅
- 描述中的 username(3-64，字母数字下划线)、password(8-64，含大小写字母和数字)、用户名唯一规则均与 validator 一致 ✅
- 无需修改。
- Mock：17845491（data=AuthResp 全字段示例，access_token/refresh_token 为 JWT 形式，user.created_at_ts 毫秒时间戳）

## 二、数据模型核对

| schemaId | 模型 | 结论 |
|---|---|---|
| 270530164 | RegisterReq (AuthCredentials) | 一致：username required/3-64/pattern ^[a-zA-Z0-9_]+$、password required/8-64，examples 正确；未修改 |
| 270530151 | ChangePasswordReq | 一致：old_password/new_password 均 required/8-64/password 规则（validator 对 old_password 同样套用 password 校验）、examples 正确；未修改 |
| 270532161 | UserProfileResp | 一致：{ user → UserDTO }；未修改 |
| 270530171 | UserDTO | 一致：id(int64)、username(string)、created_at_ts(int64 毫秒)；未修改 |

- 本模块数据模型与 DTO 完全一致，无 schema update。
- 共享模型检查（未修改，仅记录）：AuthResp(270530150) 与 dto.AuthResp 一致；response.Response(270530174) 的 data 无类型约束，符合统一包络 any+omitempty 实现。

## 三、Mock 清单（每接口恰好 1 条）

| endpointId | 接口 | mockId | 状态 |
|---|---|---|---|
| 452566407 | PUT /user/password | 17845486 | mock get 校验 bodyData 已正确落库 |
| 452566408 | GET /user/profile | 17845489 | mock get 校验 bodyData 已正确落库 |
| 452566409 | POST /user/register | 17845491 | mock get 校验 bodyData 已正确落库 |

- 均满足：name=正常响应、apiDetailId 正确、conditions=[]、response.code=200、delay=0、headers Content-Type=application/json、bodyType=json。
- 每个接口 mock list --http-api-id 复核 total=1。

## 四、自检证据

- endpoint 452566407/452566408 在 update 后再次 endpoint get：name/method/path/tags/folderId/requestBody/response/securityScheme/postProcessors 均保持不变，仅 description 按预期更新。
- 3 个接口 mock list --http-api-id 均为 1 条。
- 所有写入均先通过 apifox cli-schema validate（mock-create ×3、endpoint-update ×2）后执行。
- 未删除任何资源、未改 method/path/tags/folderId、未创建文件夹、未改代码或 .apifox/settings.json。

## 五、遗留问题 / 需 lead 决策

1. **（转交 auth 模块）** backend/internal/server/openapi.go 中登录接口描述写 "密码 8-128 个字符"，与实际 validator（8-64）冲突；本模块文档已按 8-64。login 接口文档不在本任务范围，建议由 doc-auth 核对修正。
2. **错误码不可达**：xerr.ErrChangePasswordOldEmpty(110400)、ErrChangePasswordNewEmpty(110401) 已定义，但 handler 用 ShouldBindJSON 统一返回 100000（请求参数格式错误），注册模块的 110100-110103 同理不可达。本次文档只列可达错误码；若要区分空值/格式错误，需改代码（超出文档任务范围）。
3. 两个鉴权接口原错误码表中的 100001 已移除；handler 内 user_id 缺失分支返回 100001 但实际不可达（中间件保证）。若需保留说明，建议代码侧统一为 401 语义或在中间件层归纳。
4. openapi.go 中路由路径带 /api 前缀，而 Apifox 文档按环境 baseURL 约定不带 /api（按任务要求保持现状，属全局约定，未改动）。
