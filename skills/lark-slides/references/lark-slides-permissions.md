# slides 权限速查

各方法所需 scope。遇到权限不足时按 [`../../lark-shared/SKILL.md`](../../lark-shared/SKILL.md) 的流程处理（user 身份用 `auth login --domain slides`，bot 身份去开发者后台开通对应 scope）。

| 方法 | 所需 scope |
|------|-----------|
| `slides +create` | `slides:presentation:create`, `slides:presentation:write_only`（含 `@` 占位符时还需 `docs:document.media:upload`） |
| `slides +media-upload` | `docs:document.media:upload`（wiki URL 解析还需 `wiki:node:read`） |
| `slides +replace-slide` | `slides:presentation:update`（wiki URL 解析还需 `wiki:node:read`） |
| `xml_presentations.get` | `slides:presentation:read` |
| `xml_presentation.slide.create` | `slides:presentation:update` 或 `slides:presentation:write_only` |
| `xml_presentation.slide.delete` | `slides:presentation:update` 或 `slides:presentation:write_only` |
| `xml_presentation.slide.get` | `slides:presentation:read` |
| `xml_presentation.slide.replace` | `slides:presentation:update` |
