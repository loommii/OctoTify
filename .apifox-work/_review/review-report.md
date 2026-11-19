# Apifox 文档 + Mock 同步验收报告（技术主管）

项目：OctoTify（ID 8214367，分支 main） · 代码基线：main@c6d551f · 验收时间：2026-11-19

## 一、任务分派
7 个模块各 1 个独立子智能体（共享任务板 task-1~task-7，写入范围 `.apifox-work/<模块>/`）：

| 任务 | 智能体 | 模块 | 文件夹 | 接口数 | 状态 |
|---|---|---|---|---|---|
| task-1 | doc-system | 系统（健康检查） | 系统 85273266 | 1 | completed |
| task-2 | doc-auth | 用户认证 | 用户认证 84718775 | 3 | completed |
| task-3 | doc-user | 用户管理 | 用户管理 84718780 | 3 | completed |
| task-4 | doc-sources | 消息来源管理 | 消息来源管理 84718779 | 9 | completed |
| task-5 | doc-channels | 推送渠道管理 | 推送渠道管理 84718776 | 9 | completed |
| task-6 | doc-messages | 消息管理 | 消息管理 84718777 | 4 | completed |
| task-7 | doc-push | 消息推送 | 消息推送 84718778 | 1 | completed |
| task-8 | lead | 汇总验收 | — | 30 | completed |

## 二、验收结论：通过

### A. 完整性（自动化校验）
- 接口总数 before/after 均为 30，接口 ID 集合完全一致；method/path/name/folderId 零变更（无删除重建）。
- Mock 总数 30，覆盖 30/30 接口，每接口恰好 1 条，无重复、无孤儿 Mock。
- 文件夹仍为原有 8 个（根目录 + 7 模块），无新增/删除。

### B. 文档 ↔ 代码一致性（自动化交叉校验）
1. **请求体**：30 个接口的 requestBody 引用与 handler 中 `ShouldBindJSON` 的 DTO 类型逐一一致（28 个 DTO 模型）。
2. **Query 参数**：GET /sources、/channels、/messages 的 page/page_size，/messages/filter 的 8 个筛选参数与 `PageReq`/`MessageFilterReq` 的 form 标签完全一致，无多列/漏列。
3. **响应结构**：每个接口 200 响应 data 引用与 handler/service 实际返回类型一致（AuthResp / UserProfileResp / SourceDTO / SourceDetailResponse / SourceTokenResponse / ChannelDTO / MessageDTO / MessageDetailDTO / PushResponse / PingResp）。
4. **数据模型**：28 个 Apifox 数据模型的属性名集合与 Go DTO json 标签（含嵌入结构体展开）完全一致，0 缺失 0 多余。
5. **错误码**：30 个接口描述中出现的所有 6 位业务码均在 `pkg/xerr/codes.go` 中存在；无已删除功能残留（wechat_clawbot / bind / activation / 110901-110903 / 200001-200002 均为 0 命中）。
6. **Mock 内容**：30 条 mock 的 bodyData 全部通过「响应包络 + data 结构 vs 响应 schema」递归校验：code=0、msg 正确、必填字段齐全、无 schema 外字段；无 data 的接口正确省略 data（logout/delete/更新/启停/测试等 12 个）。

### C. 本次文档修正（25 个接口 + 8 个数据模型）
- 系统 1：/ping 修 auth=noauth、补 X-Request-ID 响应头、响应示例；补「不含 /api 前缀」注意小节。
- 认证 2：logout 补 401 响应定义；refresh 补 110300 语义与 data 含 user。
- 用户 2：password/profile 移除不可达 100001；补 data omitempty 说明。
- 来源 8：错误码表修正（110501/110510/100000/110603/110605 等）、Token 格式（src+32 位 UUID）、编辑不校验 channel_ids 等语义。
- 渠道 6：/channel-types 错误码表格修复；delete/detail/disable/enable/test 补 100000/110605；确认无 ClawBot 残留、5 种可创建类型 + wechat/webhook 保留。
- 消息 1 + 4：DELETE 响应 msg=「已删除」无 data；4 个接口移除不可达 100001。
- 推送 1：重写描述（110503 覆盖已删除来源、110706/110707 优先级、message_id=0、业务错误统一 HTTP 200）+ 响应示例。
- 数据模型 8：PingResp、Response（required）、PageResult（required/int64/示例）、PushResponse、PushResult、MessageDTO、MessageDetailDTO、ChannelDTO。

### D. Mock 规格
统一规格：name=`正常响应`、conditions=[]、HTTP 200、bodyType=json、Content-Type=application/json、bodyData 为真实响应包络与合理示例数据（token 形如 src…、时间字段毫秒 *_ts、分页 list/total/page/page_size）。

## 三、遗留事项（需用户/后续决策，非本次阻塞）
1. `backend/internal/server/openapi.go` 与运行时不一致：login 描述写密码 8-128（实际 validator 8-64）；openapi.go 路由带 /api 前缀而 Apifox 走环境 baseURL 约定。若将来用 openapi.go 重新生成将会覆盖部分修正，建议同步修 openapi.go。
2. `/ping` 注册在 gin 根路径，而环境 baseURL 含 /api：Apifox 按 baseURL+path 拼接调试会得到 /api/ping（404）。已在描述中加注意说明；如需精确可调试，建议为系统模块单独建 server 或用根地址环境。
3. PageResult.list 仍为无类型（Go 设计为 any，由各分页接口用 allOf 覆盖元素类型）；Response.data 为无类型 any（全局设计）。
4. 文档与实际运行时的小差异已按代码如实记录：StepUpAuth 只校验密码非空（未强制强度）；来源详情 config 恒 null、updated_at_ts 为零值；渠道 Mock bodyData 以 JSON 对象存储（mock 服务正常）。
5. 代码侧建议（不在本次范围）：编辑来源未校验 channel_ids 归属；xerr 中 ClawBot 遗留码（110901-110903/200001-200002）可清理；110302/110511 等定义未使用。

## 四、产物
- 各模块报告：`.apifox-work/<模块>/report.md`
- 验收数据：`.apifox-work/_review/`（baseline/after/final 快照、30 条 mock 详情、30 个接口详情、28 个模型详情）
