<p align="center">
  <img src="resources/icon.png" width="128" alt="右键助手应用图标">
</p>

# 右键助手（RightClick）

让访达的右键菜单更好用：按类型和模板新建文件，拷贝文件名和路径，复制或移动到常用文件夹，用终端或任意应用打开，压缩解压，计算文件哈希。

[English](README.md) | 简体中文

## 功能

- **新建文件与从模板新建**：Word、Excel、PowerPoint、Markdown、JSON 和多种编程语言，也可以用你自己的模板。
- **用 X 打开**：终端、默认应用，或你添加的任意应用。
- **拷贝**：文件名、不含扩展名的文件名、路径。
- **复制到 / 移动到**与**常用文件夹**：把项目送到你选择的文件夹或常用文件夹。
- **压缩与哈希**：压缩为 ZIP、解压到此处、计算哈希、比较文件哈希。
- **其他**：显示或隐藏隐藏文件，彻底删除（默认关闭）。
- **菜单位置与语言**：每一项放在哪里可以选择，也可以拖动排序。支持简体中文、繁体中文和英文。

## 安装

需要 macOS 15.0 或更高版本。

1. 从 GitHub Releases 下载磁盘映像：Apple 芯片用 `RightClick-0.1.1-arm64.dmg`，Intel 芯片用 `RightClick-0.1.1-x86_64.dmg`。
2. 打开磁盘映像，把 RightClick 拖到「应用程序」文件夹。
3. 双击 RightClick，macOS 会弹出警告。关闭它。

如果 RightClick 仍然打不开，用下面两种方法之一。

**方法一：系统设置**

- 打开「系统设置 › 隐私与安全性」，在「安全性」里找到 RightClick，点「仍要打开」。输入密码或使用触控 ID，然后再点一次「仍要打开」。

**方法二：终端**

- 打开「终端」，粘贴下面这一行，按回车：

```sh
sudo xattr -dr com.apple.quarantine /Applications/RightClick.app
```

- 如果提示输入密码，输入即可。运行完后双击 RightClick。

然后在右键助手的窗口里按提示开启访达扩展。

## 从源码构建

需要 Go（版本见 `go.mod`）和 Xcode 或 Xcode 命令行工具。克隆本仓库，然后运行：

```sh
go tool mygo build -skip-dmg
sh extension/build.sh build/darwin-arm64/RightClick.app
open build/darwin-arm64/RightClick.app
```

## 许可证

右键助手以 [MIT 许可证](LICENSE) 发布。内置的 Lucide 图标采用 ISC 许可证，见 `THIRD_PARTY_LICENSES.txt`。
