# StegSolve Go

Windows 优先的桌面隐写分析工具，基于 [Wails v2](https://wails.io/) + React/TypeScript 实现。

用于快速查看图像位平面、配置 LSB 多位提取，并导出原始二进制结果。行为参考公开的 StegSolve 位平面约定，实现为独立 Go 代码，**不复制**其 Java 源码、注释、文案或资产。

## 功能

- **打开 / 拖放** 单张图像：PNG、JPEG、BMP、GIF 首帧、WebP
- **位平面视图**
  - 原图与 `R/G/B/A × bit0…7`（0 = LSB，7 = MSB）
  - 位为 `1` → 白，`0` → 黑
  - 无有效 Alpha（全部为 255）时禁用 A 通道
  - 方向键在矩阵内移动：←→ 改位，↑↓ 改通道；`[` `]` 与前后按钮按线性顺序切换
- **LSB 提取**
  - RGBA 每通道 8 位复选、全选/清空、`R0/G0/B0/RGB0/RGBA0` 预设
  - 逐行 / 逐列、通道顺序可调、通道内 LSB→MSB 或 MSB→LSB
  - Hex / ASCII 预览（前 4096 字节）
  - 完整结果由 Go 流式写入 `.bin`（不经前端 Base64）
- **预览交互**：适应窗口 / 100% / 10%–1000% 缩放、平移、无平滑像素显示
- **界面**：简体中文（保留 RGB、LSB、MSB、Hex 等术语）

## 限制（首版）

- 单次仅加载一张图，≤ 5000 万像素
- 损坏 / 超限 / 不支持的文件不会替换当前图像
- 不包含安装器、ARM64 目标包、动画逐帧、反色、随机色图、立体图、文件结构分析等扩展功能

## 要求

| 组件 | 版本 |
|------|------|
| Go | 1.25+ |
| Node.js | 20+ |
| Wails CLI | [v2.13.0](https://github.com/wailsapp/wails/releases/tag/v2.13.0) |
| 目标平台 | Windows x64（WebView2，构建时 `-webview2 embed`） |

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.13.0
```

## 开发

```bash
# 克隆
git clone https://github.com/iselt/stegsolve-go.git
cd stegsolve-go

# 后端测试
go test ./...

# 前端
cd frontend && npm install && npm run build && cd ..

# 桌面开发（需本机 GUI）
wails dev
```

## 构建 Windows 便携 EXE

```bash
./scripts/build-windows.sh

# 等价命令
wails build -clean -trimpath -platform windows/amd64 -webview2 embed
```

产物：`build/bin/stegsolve-go.exe`（便携，无需安装器）。

- 优先使用系统已安装的 WebView2
- 缺失时由内嵌引导器安装（`-webview2 embed`）

也可在 macOS / Linux 上交叉编译同一目标；GitHub Actions（`windows-latest`）会执行相同构建并上传 artifact。

## 项目结构

```text
stegsolve-go/
├── app.go / main.go          # Wails 应用绑定与入口
├── internal/imagecore/       # 图像加载、位平面、LSB 提取核心与单测
├── frontend/                 # React + TypeScript UI
├── scripts/build-windows.sh  # Windows amd64 构建脚本
└── .github/workflows/        # CI：测试 + Windows 构建
```

### 后端绑定

| 方法 | 说明 |
|------|------|
| `OpenImage` | 原生打开对话框 |
| `LoadDroppedImage` | 拖放路径加载 |
| `RenderBitPlane` | 返回位平面 / 原图 PNG（base64） |
| `SaveBitPlane` | 导出当前位平面 PNG |
| `PreviewLSB` | Hex/ASCII 预览 |
| `SaveLSB` | 流式导出 `.bin` |

## 技术说明

- 分析使用解码后的**非预乘** RGBA8 像素值
- 位平面 PNG 由 Go 按需生成（无损）
- LSB 打包：第一个提取位装入输出字节 bit7，之后依次 bit6…0；尾部不足一字节时低位补零
- 所有请求由 Go 端再次校验 `imageId`、枚举与位掩码

## 许可证

源码以本仓库许可证为准。参考的 Java StegSolve 仓库未声明明确许可证，故未使用其代码。
