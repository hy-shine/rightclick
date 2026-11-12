# 右键助手（RightClick）

macOS 访达（Finder）右键菜单增强工具。主程序的界面（主窗口、托盘）用 [mygo](https://github.com/egoist/mygo) 原生 UI 实现；Finder 里的菜单项由 Swift 写的 Finder Sync 扩展提供。

**状态：开发中。** Finder 扩展和主程序的右键动作已经接通；新建文件支持
Word、Excel、PowerPoint、RTF、WPS 文档、WPS 表格、WPS 演示（与 Office 同格式），以及 Text、Markdown、JSON、CSV、HTML、CSS、
JavaScript、Swift、Python、Go、Rust、Java、C、C++、SQL、TypeScript 和 Shell。新建页可以直接启用或停用类型，调整 Finder 顶层或
新建子菜单位置，并在编辑弹窗中修改默认文件名、文本内容或替换文档模板。
用户模板支持文本与上述办公文档，导入后保存原始内容；办公文档不会用空文件冒充。

**名称和 Bundle ID 都是占位。** `RightClick` 是模板名，`com.example.rightclick` 是模板默认值。签名或发布前要换成正式值。

## 开发

需要 macOS 和 Go。`go.mod` 声明 Go 1.27.3；本机 Go 不够新时，`go` 命令会自动下载匹配的工具链。

```sh
go test ./...        # 视图测试，不开窗口
go tool mygo dev     # 开发模式，带热重载
go tool mygo build   # 打包应用，输出到 build/
```

开发时可以用 `go tool mygo build -skip-dmg` 只生成应用，再执行
`sh extension/build.sh build/darwin-arm64/RightClick.app` 嵌入 Finder 扩展。
Finder 扩展更新后，需要重新加载扩展才能使用新菜单。

办公文档空白模板已随程序内嵌，运行时无需 Python。重新生成模板的命令：

```sh
uv run --with python-docx==1.2.0 --with openpyxl==3.1.5 --with python-pptx==1.0.2 python scripts/document_templates.py
```

mygo 固定在 v0.6.4，原因见私有决定记录 0004（不进 git）。

## 文档

私有项目文档不进 git。给 agent 的项目规则在 [AGENTS.md](AGENTS.md)。
