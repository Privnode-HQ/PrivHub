# PrivHub Web `/api` 文档目录

本目录用于人工维护 PrivHub Web 后端 `/api` 接口文档。范围以 `router/api-router.go` 中 `SetApiRouter` 注册的非 AI Relay/Web 控制台接口为准，不包含 OpenAI/Anthropic 兼容转发、聊天补全、图像生成等 AI 调用 API。

## 文件约定

- `endpoints/` 下每个 Markdown 文件只描述一个 HTTP API。
- 每个文件开头必须包含 YAML frontmatter，至少写明 `method`、`path`、`auth`、`handler`、`source`、`request`、`response`。
- 正文必须用中文说明接口用途、鉴权方式、路径参数、查询参数、请求体字段、成功响应字段、错误响应和重要业务规则。
- 字段说明以当前 Go 控制器和模型为准；不得用脚本生成文档替代人工校对。

## 通用响应约定

多数控制器通过 `common.ApiSuccess` / `common.ApiError` 返回：

```json
{
  "success": true,
  "message": "",
  "data": {}
}
```

字段说明：

- `success`: `true` 表示业务成功，`false` 表示业务失败。多数失败仍使用 HTTP 200。
- `message`: 成功时通常为空字符串；失败时为可读错误原因。
- `data`: 接口特定数据。没有返回数据时可能为 `null`、缺省或空字符串。

少数历史支付接口、OAuth 跳转和健康检查不完全使用该包络。逐接口文件会单独注明。

## 通用分页参数

使用 `common.GetPageQuery` 的接口接受：

- `p`: 页码，从 `1` 开始；缺省为 `1`。
- `page_size`: 每页条数；缺省为系统 `ItemsPerPage`，最大 `100`。
- `ps`: 兼容字段，`page_size` 为空时使用。
- `size`: Token 列表兼容字段，`page_size` 和 `ps` 为空时使用。

分页响应 `data` 通常为：

- `page`: 当前页码。
- `page_size`: 实际每页条数。
- `total`: 符合条件的总条数。
- `items`: 当前页数据数组。

