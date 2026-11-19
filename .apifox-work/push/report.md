# 消息推送模块 · Apifox 文档核对与 Mock 报告

- 项目：8214367（OctoTify，main）
- 模块：消息推送（folderId=84718778）
- 接口数：1
- 工作目录：.apifox-work/push/
- 核对基准：运行时代码（router/handler/dto/service）优先于 openapi.go

## 接口清单与结论

| endpointId | method path | name | 文档结论 | mockId |
|---|---|---|---|---|
| 452566397 | POST /push/{token} | 推送消息 | 已修复：描述/响应示例/模型 required | 17845485 |

## 452566397 POST /push/{token} 推送消息

### 代码事实（已实际阅读）
- 路由：`backend/internal/server/router.go:99` `push.POST("/:token", s.pushHandler.PushMessage)`，挂载于 `/api` group；**无 JWT / StepUpAuth 中间件**，鉴权凭据为路径中的来源 Token（`src` + UUID 去横线，共 35 字符）。
- Handler：`backend/internal/handler/push_handler.go:26-49`：`c.Param("token")` → `ShouldBindJSON(dto.PushMessageReq)` → `messageService.PushMessage` → `response.Success`。
- Service：`backend/internal/service/message_service.go:356-434` `PushMessage`（按 token + status=1 查来源 → 查绑定渠道 → 无渠道 110703 / 全删 110707 / 全停 110706 → 并发推送），`:439-578` `pushToChannels`（errgroup，单渠道 30s 超时，单渠道失败不影响整体，最终 `code` 仍为 0）。
- DTO：`backend/internal/handler/dto/message_dto.go:4-7` PushMessageReq{`title`,`message` 均 `required,min=1`}；`:47-62` PushResult{channel_id,channel_name,message_id,success,error omitempty}、PushResponse{total,success,failed,results}。
- 响应封装：`backend/pkg/response/response.go:20-26` Success → HTTP 200 + `{code:0,msg:"请求成功",data}`；业务错误统一 HTTP 200（`pkg/xerr/http_status.go`）。
- 错误码（xerr）：100000 参数校验失败（data 字段级错误）、110503 来源不存在/停用/删除、110505 查询来源失败、110605 查询渠道失败、110703 未绑定渠道、110706 渠道均停用、110707 渠道均已删除、110702 并发推送失败。

### 核对结论
路径 `/push/{token}`、method POST、name 推送消息、tags [消息推送]、folderId 84718778、path 参数 `token`(string/required/示例 `src0196a3b2c4d50000a1b2c3d4e5f67890`)、请求体 `$ref PushMessageReq(270530160)`、200 响应 `allOf[Response(270530174) + data:PushResponse(270530161)]` **均与代码一致，未改动这些结构**。请求体字段为 `title`/`message`（不是 content），模型 270530160 已正确含 required 与 minLength=1。

### 修改项
1. **接口描述重写完善**（仍保留原有准确信息）：
   - 明确「来源已删除（status=-1）」同样返回 110503（service 只查询 status=1）；
   - 补充无可用渠道时判定顺序：存在已删除渠道返回 110707，否则 110706（已删除优先）；
   - 补充 `message_id=0`（消息记录创建失败）、`success+failed=total`、`total` 为启用渠道数、results[].error 语义；
   - 补充支持渠道类型 wechat/telegram/dingtalk/email/webhook/feishu/gotify；
   - 明确业务错误统一 HTTP 200、通过 code 区分。
2. **新增 200 响应示例「正常响应」**（responseExamples，data 为 JSON 字符串），内容与 Mock bodyData 完全一致，避免文档只有 schema 无真实示例。
3. **数据模型 PushResponse(270530161)**：补 `required: [total,success,failed,results]`（Go 非指针且无 omitempty，始终返回），属性展示顺序调整为 total/success/failed/results，描述补充固定返回说明。
4. **数据模型 PushResult(270530162)**：补 `required: [channel_id,channel_name,message_id,success]`（error 保持可选，对应 omitempty），展示顺序调整，描述补充 message_id=0 语义。

### 未改动 / 保持现状
- 未改 method、path、tags、folderId、endpointId；未删除任何资源；未创建文件夹。
- `auth` 保持 `noauth`：运行时无 Authorization 中间件，凭据在 path，Apifox 鉴权组件仅支持 header/query，挂载会误导为 header 传递；鉴权机制已在描述「认证方式」中写明为 Source Token 鉴权（对应 `@Security SourceTokenAuth`）。
- 请求体未加 /api 前缀（baseURL 已含 /api）。

### Mock
- mock id：**17845485**，name=正常响应，apiDetailId=452566397，conditions=[]，HTTP 200，Content-Type application/json，bodyType=json。
- bodyData（已 `mock get` 验证落库）：
  ```json
  {"code":0,"msg":"请求成功","data":{"total":2,"success":2,"failed":0,"results":[{"channel_id":1,"channel_name":"钉钉-运维群","message_id":1024,"success":true},{"channel_id":3,"channel_name":"Telegram-告警通知","message_id":1025,"success":true}]}}
  ```
- `mock list --http-api-id 452566397` 结果：恰好 1 条（创建前为 0 条，未重复创建）。

### 写入流程
- 全部写入均先 `apifox cli-schema validate`（endpoint-update / schema-update / mock-create）通过后再执行 update/create；update 基于 `endpoint get` 完整对象构造。
- 已再次 `endpoint get` / `schema get` 复核，修改均生效。

## 遗留问题 / 需 lead 决策
1. **SourceTokenAuth 鉴权组件未挂载**：项目存在鉴权组件 `SourceTokenAuth(1229805)`，但本接口凭据在 path 中，Apifox 鉴权组件只支持 header/query；是否要为全项目统一挂载（可能在文档中显示为 header 传递）需 lead 决策。当前以描述说明 + noauth 处理。
2. **共享模型未改**：Response(270530174) 的 `data` 无类型定义（所有接口共用），属共享模型，未修改。
3. **110511（来源已停用，无法推送）** 在 push 流程中未被使用（停用来源直接返回 110503），文档未列该码，符合运行时代码；若产品希望区分「已停用」与「不存在」，需后端调整。
4. **110702** 代码路径存在（errgroup `g.Wait()`），但各 goroutine 恒返回 nil，实际几乎不可达；文档保留该错误码。
