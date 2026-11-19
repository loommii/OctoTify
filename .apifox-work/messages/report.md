# 消息管理模块 Apifox 文档核对与 Mock 报告

- 项目：8214367（OctoTify），主分支 main
- 模块：消息管理（folderId=84718777），共 4 个接口
- 任务：task-6，操作人：doc-messages
- 依据代码：router.go、message_handler.go、message_service.go、dto/message_dto.go、dto/source_dto.go(PageReq)、pkg/response/response.go、pkg/xerr/*、middleware/jwt_auth.go、middleware/error_handler.go、sender/factory.go
- 结论：4 个接口 method/path/name/tags/folderId 均正确；鉴权均为 JWT Bearer（无 StepUpAuth / SourceToken）；列表与筛选均按来源归属做用户隔离、排除已删除消息、按 created_at 倒序；时间字段均为 *_ts 毫秒时间戳。共修复 1 个接口响应描述 + 2 个数据模型，创建 4 条 Mock。

## 一、接口核对结果

### 1) GET /messages 查看消息列表（endpointId 452566393）
- 结论：一致，无需修改。
- 核对点：query 仅 page（默认1，min1）/page_size（默认20，1~100）与 dto.PageReq 一致；200 data 为 PageResult{list:MessageDTO[],total,page,page_size}；错误码 100000 参数校验、110704 查询失败、JWT 111000~111003（HTTP 401）与代码一致；描述中的用户隔离、倒序、排除已删除均正确。
- Mock：17845490（正常响应，分页 data，2 条真实感消息，status 200/300，毫秒时间戳）

### 2) DELETE /messages/{id} 删除消息（endpointId 452566394）
- 结论：已修复：200 响应描述（原「删除成功」未说明响应体形态）。
- 修复后：「删除成功（响应 msg 为「已删除」，无 data 字段）」——代码为 response.SuccessWithMsg(c,"已删除",nil)，data 因 omitempty 省略。
- 其余核对点：path id（int64、必填、min1）；软删除置 status=-1；重复删除 110705；不存在 100003；ID 非法 100000；记录失败 110704；JWT 111000~111003（HTTP 401）均与代码一致。
- Mock：17845492（正常响应，bodyData 严格为 {"code":0,"msg":"已删除"}，无 data）

### 3) GET /messages/{id} 查看消息详情（endpointId 452566395）
- 结论：一致，无需修改。
- 核对点：path id（int64、必填、min1）；200 data 引用 MessageDetailDTO（id/source_id/source_name/channel_id/channel_name/channel_type/title/content/status/created_at_ts/updated_at_ts）；排除已删除、用户隔离正确；错误码 100000/100003/110704/JWT 与代码一致。
- Mock：17845494（正常响应，含来源名、渠道名、channel_type=dingtalk）

### 4) GET /messages/filter 筛选消息（endpointId 452566396）
- 结论：一致，无需修改。
- 核对点：query 覆盖 MessageFilterReq 全部 form 字段：source_id、channel_id、status、start_date、end_date、keyword、page、page_size；默认值 page=1/page_size=20（handler 先赋默认再 bind）与文档一致；AND 关系、关键词匹配标题+内容、排除已删除、倒序均正确；错误码一致。
- Mock：17845511（正常响应，筛选后分页 data）

## 二、数据模型修改清单
1. MessageDTO（270530158）：已修复 —— 为 id/source_id/channel_id/created_at_ts/updated_at_ts 补充 `format: int64`（与 Go int64 及毫秒时间戳一致）；字段集合与 status 枚举（100/200/300/-1）与 DTO 完全一致，未改字段。
2. MessageDetailDTO（270530159）：已修复 ——
   - channel_type 原枚举仅 feishu/dingtalk/telegram/email/gotify，遗漏 wechat、webhook；现修正为 wechat、telegram、dingtalk、email、webhook、feishu、gotify（依据 sender/factory.go 内置 7 类 sender），描述同步；
   - 为 id/source_id/channel_id/created_at_ts/updated_at_ts 补充 `format: int64`。

## 三、Mock 创建清单（每接口恰好 1 条「正常响应」，conditions=[]）
| endpointId | 接口 | mockId |
| --- | --- | --- |
| 452566393 | GET /messages | 17845490 |
| 452566394 | DELETE /messages/{id} | 17845492 |
| 452566395 | GET /messages/{id} | 17845494 |
| 452566396 | GET /messages/filter | 17845511 |
- 均已 mock get 校验 bodyData 落库正确；mock list --http-api-id 复核每接口 total=1。
- 所有写入均先 cli-schema validate 通过；未删除任何资源、未改 method/path/folderId、未创建文件夹、未触碰其他模块。

## 四、遗留问题 / 需 lead 决策
1. 共享模型 PageResult（270530173）：list 无类型定义、total/page/page_size 为 integer 但无 format:int64（Go 为 int64/int）。按约定未修改共享模型，请共享模型负责人补充；当前接口已通过 allOf 覆盖 list items，展示不受影响。
2. 接口错误码表中列有 100001（未登录或 Token 已过期，HTTP 401），但 JWT 中间件实际返回 111000~111003，handler 内 response.Fail(100001) 分支在正常中间件链路下不可达。属轻微冗余表述，未批量改动三个接口，建议统一规范后由 lead 决定是否移除。
3. 列表/筛选返回的 MessageDTO 中 status 理论上不会出现 -1（已删除被查询条件排除），但 DTO 注释含 -1，模型枚举保留 -1；如需严格收敛可在共享规范中统一。

## 五、验收补充修复（2026-11-19 晚）
按 lead 验收反馈：业务码 100001 仅由 StepUpAuth（来源模块二次验证）返回，消息管理 4 个接口仅使用 JWTAuth（实际返回 111000~111003，HTTP 401），故 100001 不可达。
- 已从 4 个接口 description 的错误码表中移除 `| 100001 | 未登录或 Token 已过期（HTTP 401） |` 行，其它内容未动。
- 流程：逐个 endpoint get → 仅改 description → cli-schema validate endpoint-update（4/4 valid）→ endpoint update（4/4 success）→ endpoint get 复核。
- 复核结果：4 个接口 name/method/path/folderId/tags/parameters/requestBody/responses 与修改前完全一致，description 中已无 100001；仅 description 与 updatedAt 变化。Mock 未受影响。
