# 验收标准（lead）

## A. 完整性
- [ ] 接口总数仍为 30，ID 集合不变（无删除/重建）
- [ ] 每个接口恰好 1 条 Mock 期望，共 30 条；apiDetailId 覆盖全部 30 个接口
- [ ] 无重复 Mock、无跨模块误改

## B. 文档与代码一致（抽查 + 全量字段级）
- [ ] method / path / name / tag / folderId 与 router.go 一致
- [ ] path/query/header 参数与 DTO 的 binding/form 标签一致（名称、类型、必填、范围、默认值）
- [ ] 请求体字段与 DTO json 标签 + binding 一致，引用正确模型
- [ ] 200 响应 data 结构与 handler 返回一致（含分页 list/total/page/page_size、时间戳 *_ts）
- [ ] 错误码表与 pkg/xerr + service 返回一致；鉴权描述准确（JWT / StepUpAuth / SourceTokenAuth）
- [ ] 无已删除功能的残留（如 wechat_clawbot / bind-activation）

## C. Mock 质量
- [ ] bodyData 为响应包络 {"code":0,"msg":"请求成功","data":...}
- [ ] data 结构与响应 schema/DTO 一致，字段值合理（token 形如 src...、时间戳为毫秒）
- [ ] 无 data 的接口正确省略 data
- [ ] mock 名称统一「正常响应」，conditions=[]，code=200

## D. 合规
- [ ] 未删除任何资源、未创建文件夹、未改 .apifox/settings.json
- [ ] .apifox-work/<模块>/report.md 存在且内容完整（每接口 mock id）
- [ ] 任务板 7 个子任务均为 completed
