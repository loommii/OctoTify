# 消息来源管理模块 · Apifox 文档核对与 Mock 报告

- 项目：OctoTify（projectId 8214367），分支 main，文件夹 84718779
- 任务：task-4（消息来源管理，9 个接口 + 7 个数据模型）
- 事实依据：`backend/internal/server/router.go`、`backend/internal/handler/source_handler.go`、`backend/internal/service/source_service.go`、`backend/internal/handler/dto/source_dto.go`、`backend/internal/handler/dto/channel_dto.go`、`backend/internal/middleware/{jwt_auth,stepup_auth}.go`、`backend/pkg/response/response.go`、`backend/pkg/xerr/{codes,http_status,i18n}.go`、`docs/03-API-Specification.md`、`docs/05-Error-Codes.md`（openapi.go 仅作参考，冲突时以运行时代码为准）
- 写入方式：endpoint get 完整对象 → 仅改 description 字段 → `cli-schema validate endpoint-update` → `endpoint update`；未改 method/path/tags/folderId，未删除任何资源。

## 一、通用事实（9 个接口一致）

- 所有接口挂载在 `/api/sources` 组下，均先经过 `JWTAuth`：`Authorization: Bearer {access_token}`；JWT 失败返回 **HTTP 401**，业务码 111000/111001/111002/111003（`xerr/commonHTTPStatus`）。
- 5 个敏感接口额外经过 `StepUpAuth`（查看令牌、重置令牌、停用、启用、删除）：请求体 `VerifyPasswordReq` = `{"password": "..."}`；缺 password/请求体非法 → 100000；密码校验失败 → 110200（中文 msg 为「用户名或密码错误」，响应 HTTP 200）。
- 除 JWT 外所有业务错误返回 **HTTP 200 + 非 0 code**。
- 路径参数 `id`：int64，必填，`ParseInt` 失败 → 100000「来源ID格式错误」。
- 统一响应包络 `{code,msg,data?}`（`pkg/response`）：成功 `code=0`；`data` 带 omitempty。
- 成功消息：创建/列表/详情/令牌接口 msg=`请求成功`；编辑=`更新成功`；停用=`已停用`；启用=`已启用`；删除=`已删除`。
- 来源状态：1-启用 / 2-停用 / -1-已删除；时间字段均为 `*_ts` 毫秒时间戳；列表分页 `page`（默认 1）/ `page_size`（默认 20，最大 100，越界由 `PageReq.Normalize` 自动修正）。
- Source Token 格式：`src` + UUID（去掉连字符），共 35 位，例如 `src0196a3b2c4d50000a1b2c3d4e5f67890`。

## 二、逐接口核对结论

| endpointId | 接口 | 文档结论 | 修改项 | Mock ID |
| --- | --- | --- | --- | --- |
| 452566398 | GET /sources 查看消息来源列表 | 一致 | 无需修改（参数 page/page_size 约束、分页响应 PageResult 结构、错误码 100000/110505、列表 token 为空字符串均与代码一致） | 17845487 |
| 452566399 | POST /sources 创建消息来源 | 已修复:错误码表 | 错误码 110501 原备注「含关联渠道不存在/查询失败」与代码不符（渠道校验失败实际直接返回 110603/110605）；改为 110501 创建来源失败，新增 110603 关联渠道不存在或不属于当前用户、110605 查询关联渠道失败 | 17845496 |
| 452566400 | DELETE /sources/{id} 删除消息来源 | 已修复:100000 说明 | 100000 说明补齐「来源ID格式错误」；StepUpAuth、软删除（status=-1）、级联解绑、错误码 110200/110503/110509/110506 均核对一致 | 17845512 |
| 452566401 | GET /sources/{id} 查看来源详情 | 已修复:错误码表 | 补充 100000 来源ID格式错误；响应结构 SourceDetailResponse{source,channels}、渠道仅基础字段（config=null、updated_at_ts 不保证有效）、110503/110505/110605 核对一致 | 17845514 |
| 452566402 | PUT /sources/{id} 编辑消息来源 | 已修复:描述/错误码 | 110510 原备注「含关联渠道校验失败」不成立（UpdateSource 不校验 channel_ids 归属/存在性），改为「更新来源失败」；参数说明补充「编辑时不会校验 channel_ids 中渠道的存在性和归属（与创建接口不同）」；成功响应为 {code:0,msg:"更新成功"} 无 data，与 SuccessWithMsg(msg,nil) 一致 | 17845516 |
| 452566403 | PUT /sources/{id}/disable 停用消息来源 | 已修复:100000 说明 | 100000 说明补齐「来源ID格式错误」；停用后推送失败返回 110503（message_service 仅查询 status=1 来源）、110507/110509/110510 均核对一致 | 17845519 |
| 452566404 | PUT /sources/{id}/enable 启用消息来源 | 已修复:100000 说明 | 100000 说明补齐「来源ID格式错误」；110508 已启用、110509、110510 核对一致 | 17845521 |
| 452566405 | POST /sources/{id}/token 查看来源令牌 | 已修复:Token 格式说明 | Token 格式由 `src{uuid}` 明确为「src + UUID（去掉连字符），共 35 位」；100000 说明补齐来源ID格式错误；响应 SourceTokenResponse{token}、110503/110505 核对一致 | 17845523 |
| 452566406 | PUT /sources/{id}/token 重置来源令牌 | 已修复:Token 格式说明 | 补充新 Token 格式说明；100000 说明补齐来源ID格式错误；110502 覆盖生成失败与更新失败、110503/110509/110505 核对一致；成功响应 {code:0,msg:"请求成功",data:{token}} | 17845525 |

所有接口的 method / path / name / tags（消息来源管理）/ folderId（84718779）核对无误；路径未加 `/api` 前缀（环境 baseURL 已含 /api），未做改动。

## 三、数据模型核对（7 个，均未修改）

| schemaId | 模型 | 结论 |
| --- | --- | --- |
| 270530156 | CreateSourceReq | 一致：name 必填 1-128；description 可选 maxLength 512；channel_ids 可选 int64 数组；引用关系正确 |
| 270530170 | UpdateSourceReq | 一致：字段与 CreateSourceReq 相同；channel_ids 语义（不传=保持、[]=解绑、非空=覆盖）与 service 的 nil 判断一致 |
| 270530165 | SourceDTO | 一致：id/user_id/name/token/description/status(1,2,-1)/created_at_ts；token 说明「创建返回真实值、列表固定空字符串」与代码一致 |
| 270530166 | SourceDetailDTO | 一致：在 SourceDTO 基础上增加 updated_at_ts、last_used_at_ts（0 表示未使用） |
| 270530167 | SourceDetailResponse | 一致：source 引用 270530166；channels 引用 ChannelDTO(270530152)，说明 config=null、updated_at_ts 不保证有效，与 GetSourceDetail 的 Select 字段一致 |
| 270530168 | SourceTokenResponse | 一致：{token:string} |
| 270530172 | VerifyPasswordReq | 一致：password 必填，8-64 位、含大小写与数字（与 dto binding:"required,password" 及 pkg/validator 规则一致） |

未修改任何共享模型（270530152 ChannelDTO、270530173 PageResult、270530174 Response 属其他模块/共享资源），只核对了引用是否指向正确模型。

## 四、Mock 期望（每接口 1 条，共 9 条）

- 统一结构：`{"name":"正常响应","apiDetailId":<接口ID>,"conditions":[],"response":{"code":200,"delay":0,"headers":[{"key":"Content-Type","value":"application/json"}],"bodyType":"json","bodyData":<真实响应包络>}}`
- 创建前均通过 `apifox cli-schema validate mock-create --file <f>`，创建后 `mock get` 抽查确认 bodyData 为 JSON 对象（非字符串二次编码）。
- 创建前 `mock list --http-api-id` 确认 9 个接口均无存量 Mock，故创建 9 条，无重复。
- bodyData 覆盖：列表返回 `{list,total,page,page_size}` 且列表项 token=""；创建返回完整 SourceDTO 含真实 token；无 data 的接口（编辑/停用/启用/删除）只返回 `{code,msg}`；详情返回 `{source,channels}` 且渠道 config=null；令牌接口返回 `{token}`。

## 五、遗留问题 / 需 lead 决策

1. **VerifyPasswordReq 的密码强度约束未被运行时强制执行**：文档/模型按 DTO 标注 8-64 位 + 大小写数字，但 `middleware.StepUpAuth` 只用匿名结构体校验 `password` 非空，实际不校验强度。当前文档保留 DTO 约束（与 dto 定义一致）；如需「文档=运行时」需改中间件（非本文档任务范围），建议 lead 决策。
2. **详情接口渠道字段的已知缺陷**：`GetSourceDetail` 只 Select 渠道基础列，导致 `config` 恒为 `null`、`updated_at_ts` 来自零值时间（约 -6795364578871）。已在接口描述中如实说明「不保证有效」，未修改代码。
3. **列表接口 token 字段恒为空字符串**（SourceDTO 无 omitempty，列表构造时未赋值）。已在模型与接口描述中说明，未改 DTO。
4. 编辑接口不校验 channel_ids 的渠道归属/存在性（创建接口会校验），存在把非法渠道 ID 写入关联表的风险，已在 452566402 描述中标注，建议后续在 service 层补齐校验。

## 六、自检结果（CLI 复核）

- **接口文档复核**：对修改过的 8 个接口重新 `endpoint get`，确认 description 中新增/修正的错误码与说明均已落库；method / path / name / tags / folderId(84718779) 全部保持原值，未改 method/path。
- **Mock 唯一性**：逐接口 `mock list --project 8214367 --http-api-id <id>`，9 个接口均返回 `total=1`，且均为 `name="正常响应"`，无重复 Mock。
- **Mock 落库内容**：逐条 `mock get <mockId>` 确认 `response.code=200、bodyType=json、headers[0]=Content-Type: application/json、conditions=[]`，且 `bodyData` 以 JSON 对象（非字符串）落库，`code=0`；分页/详情/令牌接口 data 字段结构正确，无 data 接口（编辑/停用/启用/删除）严格只含 code+msg。
- **共享模型复核**（只读，未修改）：270530174 Response 必填 code/msg、data 可选且带 omitempty 说明；270530173 PageResult 必填 list/total/page/page_size，与 `pkg/response` 实现一致；270530152 ChannelDTO 被详情接口正确引用。
- 项目级 `mock list --project 8214367` 在开工时为 0 条（确认无存量 Mock 需要保留或修正）。
