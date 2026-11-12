# 右键助手：agent 工作规则

Finder 右键助手。主程序用 mygo v0.6.4 原生 UI，Finder 菜单是 Swift 写的 Finder Sync 扩展。构建和测试命令见 README.md。

## 何时读哪份文档

私有项目文档（决定记录、功能清单）不进 git，文件名按编号或主题命名。用 `find . -name '<文件名>'` 查找。找不到时，停下来问用户，不要凭记忆补写。

- 改 mygo 版本，或怀疑某个 mygo API 的行为：读 `0004-pin-mygo-v0.6.4.md`，改版本时同步更新它。
- 改 Finder 扩展，或扩展与主程序之间的通信：读 `0002-finder-sync-extension.md`。
- 做签名、公证、分发：读 `0003-distribution-channel.md`。
- 增减功能或改优先级：读 `features.md`。
- 改 `ui` 视图代码：读 `.agents/skills/mygo-maintenance/SKILL.md`。

## 环境里查不到的规则

- 依赖写固定版本，如 `@v0.6.4`。`mygo init` 会拉最新版，生成后要改回。
- 主程序保持 `CGO_ENABLED=0` 可构建，mygo 本身也是这样。
- `mygo.json` 的 `name` 和 `identifier` 是占位值，改之前先问用户。
- 测试 UI 用 `ui.NewTester`，等待条件成立，不用固定时长的 sleep。
- 代码注释用英文（Go、Swift、shell）。文件名和目录名也用英文，不用中文。
- 图标不手绘：从 Lucide（ISC）取现成 SVG，放在 `extension/RightClickFinderSync/ActionIcons/`，许可文件随目录保留。图标要和动作的语义对得上；新增菜单项必须同时加图标，并让 `extension/tests/main.swift` 的图标断言通过。

## 做完的标准

改了 Go 代码后，以下命令都退出 0：

- `go test ./...`
- `CGO_ENABLED=0 go test ./...`
- `go vet ./...`
- `go tool mygo vet .`
