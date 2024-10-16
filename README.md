# Github Copilot 后端代理服务

[仅需四步](#快速使用步骤)即刻拥有完全离线的 `Copilot小助手` 同款服务，速度更快，更稳定，更安全。

借助其他FIM模型（如DeepSeek）来接管GitHub Copilot插件服务端, 廉价的模型+强大的补全插件相结合, 使得开发者可以更加高效的编写代码。  

✨现提供一个免费的公共服务端点: `mycopilot.noteo.cn`, 服务端代码会与此仓库版本保持一致, 响应速度较慢可用于测试服务有效性但不保证稳定性, 使用方式详见:[IDE设置方法](#ide设置方法)

## 特点

- [x] 支持使用Docker部署, 简单方便
- [x] 支持多种IDE,
  如: [VSCode](#vscode), [Jetbrains IDE系列](#jetbrains-ide系列), [Visual Studio 2022](#visual-studio-2022), [HBuilderX](#hbuilderx)
- [x] 支持任意符合 `OpenAI` 接口规范的模型, 和 `Ollama` 部署的本地模型
- [x] `GitHub Copilot` 插件各种API接口**全接管**, 无需担心插件升级导致服务失效
- [x] 代码补全请求防抖设置, 避免过度消耗 Tokens

## 如何使用?

> 在使用之前确保自己的环境是干净的, 也就是说不能使用过其他的激活服务, 可以先检查自己的环境变量将 `GITHUB` `COPILOT`
> 相关的环境变量删除, 然后将插件更新最新版本后重启IDE即可.

### 快速使用步骤

1. **部署服务**: 可以使用[下载文件直接部署使用](#下载文件直接部署使用) 或 使用[docker部署](#docker部署).
2. **配置IDE**: 详细参考下面的[IDE设置方法](#ide设置方法).
3. **修改本地hosts文件**: 具体参考[配置本机hosts文件](#配置本机hosts文件).
4. **信任SSL证书**: 具体自行搜索各个系统平台信任根证书操作 **(可选)**.
5. 重启IDE, 点击登录 `GitHub Copilot` 插件即可.

### Docker部署

**(推荐)** 懒人推荐使用此方案, 比较简单  
已经将自签证书的工作做完了, 只需要将 [docker-compose.yml](docker-compose.yml) 文件下载到本地, 将里面的
**模型API KEY 替换为你的**, 然后执行以下命令即可启动服务:

```shell
# 启动服务
docker-compose up -d

# 停止服务
docker-compose down

# 更新服务
1. docker-compose pull
2. docker-compose up -d

# 查看日志
docker-compose logs -f
```

镜像全部上传到阿里云容器镜像服务, 国内访问无惧.

### 下载文件直接部署使用

1. 下载最新版本的可执行文件
   访问 [releases](https://gitee.com/ripperTs/github-copilot-proxies/releases), 修改里面 `.env` 文件的配置项, 然后直接运行即可.
2. 如果希望绑定自己自有的域名, 可以参考: [自定义域名](#自定义域名) 配置, 然后将所有 `mycopilot.com` 相关的域名都修改为自己的域名.
3. 启动服务后然后按照[IDE设置方法](#ide设置方法)配置IDE.
4. 重启IDE,登录 `GitHub Copilot` 插件.

### 配置本机hosts文件

将下面hosts配置添加到本机hosts文件中, 以便访问本地服务:

```
127.0.0.1 mycopilot.com
127.0.0.1 api.mycopilot.com
127.0.0.1 copilot-proxy.mycopilot.com
127.0.0.1 copilot-telemetry-service.mycopilot.com
```

### 环境变量参数说明

详细参考: [环境变量参数说明](PARAM.md)

## IDE设置方法

### VSCode

1. 安装插件: `GitHub Copilot`
2. 修改 VSCode 的 settings.json 文件, 添加以下配置:

```json
"github.copilot.advanced": {
"authProvider": "github-enterprise",
"debug.overrideCAPIUrl": "http://api.mycopilot.com:1188",
"debug.overrideProxyUrl": "http://copilot-proxy.mycopilot.com:1188",
"debug.chatOverrideProxyUrl": "http://api.mycopilot.com:1188/chat/completions"
},
"github-enterprise.uri": "http://mycopilot.com:1188"
```

vscode 使用https有些问题, 并且直接使用ip好像也不行, 所以这里使用http的域名+端口的形式   
(不直接使用80端口是为了防止服务冲突), 形式不重要直接粘贴进去即可.

### Jetbrains IDE系列

1. 找到`设置` > `语言与框架` > `GitHub Copilot` > `Authentication Provider`
2. 填写的值为: `mycopilot.com`
3. 首次打开 `IDE` 应该会提示是否信任证书的弹窗, 点击**同意**即可, 如果已经配置了系统级别的信任证书可以忽略.

### Visual Studio 2022

1. 更新到最新版本（内置 Copilot 版本）至少是 `17.10.x` 以上
2. 首先开启 Github Enterprise 账户支持：工具-环境-账户-勾选“包含 Github Enterprise 服务器账户”
3. 然后点击添加 Github 账户，切换到 Github Enterprise 选项卡，输入 `https://mycopilot.com` 即可。

🚨 如果是默认自签证书的域名, 那么本次操作之前务必操作下 `信任根证书` 然后重启浏览器和IDE, 具体方法网上搜索下
证书文件 [mycopilot.crt](ssl/mycopilot.crt)  
🚧 Chat服务在代码选中后右键选择解释代码会报错, 解决方法是点击一下"在聊天窗口中继续"即可.

### HBuilderX

> 注意, 插件中的相关 domain 已经写死无法修改, 所以必须使用默认的 mycopilot.com 域名配置.

1. 下载 **[copilot-for-hbuilderx.zip](https://pan.quark.cn/s/eb7f501ad585)** 插件到本地
2. 将插件安装到 plugin目录下, 详细参考: [离线插件安装指南](https://hx.dcloud.net.cn/Tutorial/OfflineInstall)
3. 重启 Hbuilder X 后点击登录 `GitHub Copilot` 即可.

## 自定义域名

如果你有自己的域名或者不想使用默认的 `mycopilot.com` 域名, 你需要申请或自签一个https证书, 然后将证书文件路径配置到
`.env` 或 `docker-compose.yml` 文件中.

### 自有域名配置

将域名添加解析以下四个域名, 假设你的域名为 `yourdomain.com` (非必须是顶级域名), 则你需要解析的域名记录如下:

- `DEFAULT_BASE_URL`: `yourdomain.com`
- `API_BASE_URL`: `api.yourdomain.com`
- `PROXY_BASE_URL`: `copilot-proxy.yourdomain.com`
- `TELEMETRY_BASE_URL`: `copilot-telemetry-service.yourdomain.com`
- 以上四个域名都需要配置SSL证书, 通配符证书教程参考[免费通配符证书申请方法](#通配符证书申请方法).
- 以上几个域名前缀 (`api`, `copilot-proxy`, `copilot-telemetry-service`) 必须是一样的, 不可自定义修改, 否则会导致插件无法登录或正常使用.
- 最后将以上域名修改到对应的环境变量配置文件中.

### 没有域名自签本地证书

如果你没有域名, 可以随便想一个"假"域名, 然后直接修改 `hosts` 文件的方式进行解析, 然后使用自签证书即可.

## 支持的模型

> 大部分Chat模型都兼容, 因此下面列出的模型是支持 FIM 的模型, 也就是说支持补全功能.

| 模型名称 (区分大小写)                                                                                                                   | 类型      | 接入地址                                                                                                           | 说明                          |
|--------------------------------------------------------------------------------------------------------------------------------|---------|----------------------------------------------------------------------------------------------------------------|-----------------------------|
| [Qwen/Qwen2.5-Coder-7B-Instruct](https://docs.siliconflow.cn/features/fim)                                                     | 免费      | <details><summary>查看地址</summary>`https://api.siliconflow.cn/v1/completions`</details>                          | 硅基流动官方支持的 FIM 补全模型, 完美适配且免费 |
| [DeepSeek (API)](https://www.deepseek.com/)                                                                                    | 付费      | <details><summary>查看地址</summary>`https://api.deepseek.com/beta/v1/completions`</details>                       | 👍🏻完美适配且价格实惠, 推荐使用         |
| [deepseek-ai/DeepSeek-V2.5](https://docs.siliconflow.cn/features/fim)                                                          | 付费      | <details><summary>查看地址</summary>`https://api.siliconflow.cn/v1/completions`</details>                          | 硅基流动官方支持的 FIM 补全模型, 完美适配    |
| [deepseek-ai/DeepSeek-Coder-V2-Instruct](https://docs.siliconflow.cn/features/fim)                                             | 付费      | <details><summary>查看地址</summary>`https://api.siliconflow.cn/v1/completions`</details>                          | 硅基流动官方支持的 FIM 补全模型, 完美适配    |
| [codestral-latest (API)](https://docs.mistral.ai/api/#tag/fim)                                                                 | 免费 / 付费 | <details><summary>查看地址</summary>`https://api.mistral.ai/v1/fim/completions`</details>                          | Mistral 出品, 免费计划有非常严重的频率限制  |
| [stable-code](https://ollama.com/library/stable-code)                                                                          | 免费      | <details><summary>查看地址</summary>`http://127.0.0.1:11434/v1/chat/completions`</details>                         | Ollama部署本地的超小量级补全模型         |
| [codegemma](https://ollama.com/library/codegemma)                                                                              | 免费      | <details><summary>查看地址</summary>`http://127.0.0.1:11434/v1/chat/completions`</details>                         | Ollama部署本地的补全模型             |
| [codellama](https://ollama.com/library/codellama)                                                                              | 免费      | <details><summary>查看地址</summary>`http://127.0.0.1:11434/v1/chat/completions`</details>                         | Ollama部署本地的补全模型             |
| [qwen-coder-turbo-latest](https://help.aliyun.com/zh/model-studio/user-guide/qwen-coder?spm=a2c4g.11186623.0.0.a5234823I6LvAG) | 收费      | <details><summary>查看地址</summary>`https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions`</details> | 阿里通义代码补全模型                  |
| [mike/deepseek-coder-v2](https://ollama.com/mike/deepseek-coder-v2)                                                            | 免费      | <details><summary>查看地址</summary>`http://127.0.0.1:11434/api/generate`</details>                                | Ollama支持的 `suffix` 参数方式实现   |
| [deepseek-coder-v2](https://ollama.com/library/deepseek-coder-v2)                                                              | 免费      | <details><summary>查看地址</summary>`http://127.0.0.1:11434/api/generate`</details>                                | Ollama支持的 `suffix` 参数方式实现   |

**💡以上接入的模型除了 `DeepSeek` 模型与 `硅基流动` 模型之外, 效果均不理想, 这里仅做接入更多模型的Demo参考.**,
理论上后续如果有API支持标准的FIM补全, 都可以接入.

## 通配符证书申请方法

> 使用 [acme.sh](https://github.com/acmesh-official/acme.sh/wiki/%E8%AF%B4%E6%98%8E) 依旧可以申请通配符域名证书,
> 如果你的域名托管在 `cf` `腾讯云` `阿里云` 等等, 都可以使用他们的API来自动续期.

### 安装acme.sh

```shell
# 官方
curl https://get.acme.sh | sh -s email=617498836@qq.com

# 国内镜像
https://github.com/acmesh-official/acme.sh/wiki/Install-in-China

# 使环境变量立即生效
source ~/.bashrc

# 创建一个 alias，便于后续访问:
alias acme.sh=~/.acme.sh/acme.sh
```

### 操作步骤

我这里域名是托管在 `cf` 上的, 所以使用 `cf` 的API来申请证书, 你可以根据自己的情况来选择.

1. 配置dns秘钥

```shell
export CF_Email="110110110@qq.com"
export CF_Key="xxxxxxx"
```

2. 签发证书

```shell
acme.sh --issue --dns dns_cf -d supercopilot.top -d '*.supercopilot.top'
```

3. 安装证书

```shell
# 新建一个证书目录
mkdir -p /etc/nginx/cert_file/supercopilot.top

# 安装证书
acme.sh --install-cert -d supercopilot.top -d *.supercopilot.top \
		--key-file   /etc/nginx/cert_file/key.pem  \
		--fullchain-file /etc/nginx/cert_file/fullchain.pem
```

4. 修改对应的环境变量配置
    - CERT_FILE=/etc/nginx/cert_file/fullchain.pem
    - KEY_FILE=/etc/nginx/cert_file/key.pem

**如果你使用`宝塔`面板将会更加容易的申请, 因为面板中已经高度集成了此模块**

## 注意事项

1. 请勿将本服务用于商业用途, 仅供学习交流使用
2. 请勿将本服务用于非法用途, 一切后果自负

## 鸣谢

- [LoveA/copilot_take_over](https://gitee.com/LoveA/copilot_take_over)
- [override](https://github.com/linux-do/override)
