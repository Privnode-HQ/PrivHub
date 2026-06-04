---
method: GET
path: /api/admin/service-accounts/
auth: admin
handler: controller.GetAdminServiceAccounts
source: router/api-router.go:28
request:
  query_params:
    - keyword
    - p
    - page_size
response:
  success_http_status: 200
  envelope: common
  data: PageInfo<AdminServiceAccountView>
---

# GET `/api/admin/service-accounts/`

分页列出当前管理员可管理的 Admin Service Account。

## 查询参数字段

- `keyword`: 字符串，可选。按名称、目标管理员等服务层支持的字段搜索。
- `p`: 页码，从 `1` 开始。
- `page_size`: 每页条数，最大 `100`。

## 成功响应字段

- `success`: `true`。
- `message`: 空字符串。
- `data.page`: 当前页码。
- `data.page_size`: 实际每页条数。
- `data.total`: 总条数。
- `data.items`: Service Account 视图数组。
- `data.items[].id`: 数据库 ID。
- `data.items[].service_account_id`: 对外展示的 ASA ID，前缀为 `asa_`。
- `data.items[].name`: 名称。
- `data.items[].description`: 描述。
- `data.items[].user_id`: 绑定管理员用户 ID。
- `data.items[].username`: 绑定管理员用户名。
- `data.items[].user_cah_id`: 绑定管理员 CAH ID。
- `data.items[].user_role`: 绑定管理员角色。
- `data.items[].created_by_id`: 创建者用户 ID。
- `data.items[].created_by_username`: 创建者用户名。
- `data.items[].created_by_cah_id`: 创建者 CAH ID。
- `data.items[].status`: 状态，`1` 启用，`2` 禁用。
- `data.items[].scopes`: 权限范围字符串，当前为 `admin:api`。
- `data.items[].allow_ips`: IP 白名单文本，空或 `null` 表示不限制。
- `data.items[].created_time`: 创建时间 Unix 秒。
- `data.items[].accessed_time`: 最近访问时间 Unix 秒，未访问时为 `0`。
- `data.items[].expires_at`: 过期时间 Unix 秒。
- `data.items[].deleted_at`: 删除时间对象；未删除时为空。
- `data.items[].expired`: 布尔值，当前时间是否已超过 `expires_at`。
- `data.items[].target`: 绑定管理员摘要对象。
- `data.items[].target.id`: 绑定管理员用户 ID。
- `data.items[].target.username`: 绑定管理员用户名。
- `data.items[].target.cah_id`: 绑定管理员 CAH ID。
- `data.items[].target.role`: 绑定管理员角色。
- `data.items[].created_by`: 创建者摘要对象，字段同 `target`。

## 失败响应

- `success`: `false`。
- `message`: 数据库查询或权限过滤错误。

