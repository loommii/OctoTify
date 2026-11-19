# 推送渠道管理模块 — 文档核对与 Mock 交付报告

- 项目：8214367（OctoTify），分支 main
- 模块文件夹：84718776（推送渠道管理），接口 9 个（未增删、未改 method/path/tag/folderId）
- 事实来源：`backend/internal/server/router.go`、`backend/internal/handler/channel_handler.go`、`backend/internal/service/channel_service.go`、`backend/internal/handler/dto/channel_dto.go`、`backend/internal/sender/*`、`backend/pkg/response/response.go`、`backend/pkg/xerr/codes.go`、`backend/pkg/validator/custom.go`（openapi.go 仅作参考，冲突以运行时代码为准）
- 工作目录：.apifox-work/channels/

## 一、接口核对与 Mock

| endpointId | METHOD path | 名称 | 核对结论 | mockId |
| --- | --- | --- | --- | --- |
| 452566384 | `GET /channel-types` | 获取渠道类型元数据 | 一致；修复错误码表格缺表头（Markdown 渲染失效） | 17845518 |
| 452566385 | `GET /channels` | 查看推送渠道列表 | 一致（page/page_size 默认值与范围、分页 data 结构、100000 分页校验错误码均正确） | 17845520 |
| 452566386 | `POST /channels` | 创建推送渠道 | 一致（type 枚举 5 种可创建、wechat/webhook 保留说明、config 必填字段与各 sender 一致） | 17845522 |
| 452566387 | `DELETE /channels/{id}` | 删除推送渠道 | 已修复:错误码补 100000（id 非数字）、110605 | 17845524 |
| 452566388 | `GET /channels/{id}` | 查看渠道详情 | 已修复:错误码补 100000（id 非数字） | 17845526 |
| 452566389 | `PUT /channels/{id}` | 编辑推送渠道 | 一致（type 不可改、100000/110603/110609/110605 齐全） | 17845527 |
| 452566390 | `PUT /channels/{id}/disable` | 停用推送渠道 | 已修复:错误码补 100000（id 非数字） | 17845528 |
| 452566391 | `PUT /channels/{id}/enable` | 启用推送渠道 | 已修复:错误码补 100000（id 非数字） | 17845529 |
| 452566392 | `POST /channels/{id}/test` | 测试渠道连接 | 已修复:错误码补 100000、110605 | 17845530 |

每条 Mock：name=`正常响应`，conditions=[]，response.code=200，bodyType=json，bodyData 为代码真实返回包络（`{"code":0,"msg":...}` + data）。已用 mock get 回读确认 bodyData 落库。

## 二、鉴权核对

本模块 9 个接口全部只挂 `middleware.JWTAuth`（router.go setupChannelRoutes + /channel-types），**没有** StepUpAuth（二次密码）或 SourceTokenAuth（来源令牌）。文档统一写 `Authorization: Bearer {access_token}`，准确。
JWT 失败码与 middleware/jwt_auth.go 一致：111000 未提供令牌、111001 格式错误、111002 令牌无效/过期、111003 令牌类型错误（均 HTTP 401）；100001 为 handler 内防御性用户态解析失败，保留在错误码表。

## 三、数据模型修改清单

| schemaId | 模型 | 修改内容 |
| --- | --- | --- |
| 270530152 | ChannelDTO | `config` 字段描述由"来源详情接口中未加载时可能为 null"（从来源模块复制残留）改为"渠道配置 JSON 对象，字段随渠道类型不同（feishu/dingtalk/telegram/email/gotify）" |

其余 5 个本模块模型核对为与 DTO 一致，未改动：
- 270530153 ChannelTypeMeta：type 枚举 feishu/dingtalk/telegram/email/gotify，config_fields → 270530154
- 270530154 ConfigField：name/label/type/required/placeholder，type 枚举 string/number/boolean/password/url（email skip_tls_verify 为 boolean，正确）
- 270530155 CreateChannelReq：type/name/config 均必填，name 1-128
- 270530169 UpdateChannelReq：name/config 必填，无 type（与 service 按原渠道类型校验 config 一致）
- 272572301 ChannelConfig：字段与各 sender 实际读取的配置项逐一比对一致
  - feishu: webhook_url(必), secret(可) — sender/feishu.go
  - dingtalk: webhook_url(必), secret(必) — sender/dingtalk.go
  - telegram: bot_token(必), chat_id(必), proxy(可) — sender/telegram.go
  - email: smtp_host/smtp_port/username/password/to(必), cc/from_name/skip_tls_verify(可) — sender/email.go
  - gotify: server_url/app_token(必), priority(可, 0-10) — sender/gotify.go

## 四、关键事实核对

- 可创建渠道类型为 ChannelTypeMetas 的 5 种：feishu/dingtalk/telegram/email/gotify。`ValidChannelTypes` 中 wechat/webhook 仍为 true，但 `validator.ValidateChannelConfig` 找不到 meta 会报"未知的渠道类型"，实际创建失败（100000），文档已按"保留未实现"表述。
- 已确认全模块文档/模型无 `wechat_clawbot`、`bind`、`activation` 等已删除微信 ClawBot 功能的残留。
- 业务失败统一 HTTP 200 + 业务 code（response.Fail），仅鉴权失败为 401；各接口错误码表已按 handler/service 实际 xerr 对齐。
- 列表：按创建时间倒序、排除 status=-1；分页 data 为 {list,total,page,page_size}，page 默认 1、page_size 默认 20 最大 100。
- 无 data 的接口（编辑/停用/启用/删除/测试）成功时响应仅 code+msg（Go omitempty），Mock 未附加 data。
- 删除为软删除（status=-1）并连带解绑 source_channel，文档描述一致。

## 五、遗留问题 / 需 lead 决策

1. 共享模型 270530173 PageResult 的 `list` 属性没有 type/items（无 schema），导致分页接口 data.list 的具体类型只能在各接口响应里用 allOf 覆盖；建议由 lead 统一决定是否补 `type: array`。
2. 共享模型 270530174 Response 的 `data` 为无类型 any，属全局设计，未改动。
3. 本模块接口错误码表中保留 100001（防御性），若要求"仅列实际可触发错误码"可删除。
4. 所有 9 条 Mock 的 bodyData 以 JSON 对象形式落库（validate 通过、mock get 回读正常）；如前端 Mock 服务要求 bodyData 为字符串，需统一批量转换，请 lead 决策。
## 六、自检结果

- mock list --http-api-id 逐个回查：9/9 接口恰好 1 条 Mock（另有 0 条残留，无需清理）。
- mock get 回读 9 条：bodyData 均按代码真实包络落库（编辑/停用/启用/删除/测试 5 个接口仅 code+msg，符合 omitempty）。
- 6 个被修改接口已 endpoint get 复验：name/method/path/tags/folderId 未变，错误码表生效，响应 $ref 完整保留。
- 数据模型 270530152 更新返回 success。
