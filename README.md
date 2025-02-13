# Github Copilot 后端代理服务

借助其他支持 `FIM` 接口的模型（如DeepSeek）来接管GitHub Copilot插件服务端, 廉价的模型+强大的补全插件相结合, 使得开发者可以更加高效的编写代码。

✨ **免费的公共服务端点: [mycopilot.noteo.cn](https://mycopilot.noteo.cn/help), 感谢 [硅基流动](https://cloud.siliconflow.cn/i/NO6ShUc3) 提供免费的模型服务**  

🚨 **破坏性更新提示: v0.1.0 版本更新之后需要做配置调整, 具体参考: [升级指南](https://gitee.com/ripperTs/github-copilot-proxies/releases/tag/v0.1.0) 部分内容.**

## 功能特性

- [x] 使用 `docker-compose` 部署, 简单方便
- [x] 多种IDE, 如: [VSCode](#vscode), [Jetbrains IDE系列](#jetbrains-ide系列), [Visual Studio 2022](#visual-studio-2022), [HBuilderX](#hbuilderx)
- [x] 任意符合 `OpenAI` 接口规范的模型, 和 `Ollama` 部署的本地模型
- [x] `GitHub Copilot` 插件各种API接口**全接管**, 无需担心插件升级导致服务失效
- [x] 代码补全请求防抖设置和自定义 prompt 精简, 避免过度消耗 Tokens
- [x] 使用 Github Copilot 官方服务, 参考: [使用GitHub Copilot官方服务](#使用github-copilot官方服务)
- [x] VSCode 对话编辑模式
- [x] 代码补全APIKEY支持多个轮询, 避免限频
- [x] 无需自有域名, 自动配置和续签 `Let's Encrypt` SSL证书 (每 60 天自动更新一次证书, 自动重载 https 服务)
- [x] 局域网共享, 可多台电脑共享一个服务端, 参考: [局域网共享方案](#局域网共享方案)
- [x] 完全纯离线部署, 无需任何外部网络支持, 参考: [纯内网离线部署方案](#纯内网离线部署方案)
- [ ] Ollama 部署的 Embeddings 模型支持

## 如何使用?

**在使用之前确保自己的环境是干净的, 也就是说不能使用过其他的激活服务, 可以先检查自己的环境变量将 `GITHUB` `COPILOT` 相关的环境变量删除, 然后将插件更新最新版本后重启IDE即可.**    

**⚠️ 如果你本地有使用科学上网工具, 那必须将域名 `copilot.supercopilot.top` 系列域名添加直连名单中, 否则无法正常使用!** 


### 快速使用步骤

1. **部署服务**: 可以使用[下载文件直接部署使用](#下载文件直接部署使用) 或 使用[docker部署](#docker部署).
2. **配置IDE**: 详细参考下面的[IDE设置方法](#ide设置方法).
3. **重启IDE**: 点击登录 `GitHub Copilot` 插件即可.

### Docker部署

**(推荐)** 懒人推荐使用此方案, 比较简单  
**模型API KEY 替换为你的**, 然后执行以下命令即可启动服务:  

```shell
# 启动服务
docker-compose up -d

# 停止服务
docker-compose down

# 更新服务
1. docker-compose pull
2. docker-compose down
3. docker-compose up -d

# 查看日志
docker-compose logs -f
```

镜像全部上传到阿里云容器镜像服务, 每个版本都有对应的镜像可使用或回滚.  

### 下载文件直接部署使用

1. 下载最新版本的对应系统的可执行文件访问 [releases](https://gitee.com/ripperTs/github-copilot-proxies/releases).
2. 在可执行文件同级目录下创建 `.env` 文件, 参考 [.env.example](.env.example) 文件配置.
3. 启动服务后然后按照[IDE设置方法](#ide设置方法)配置IDE.
4. 重启IDE,登录 `GitHub Copilot` 插件.

### 自有服务器部署

1. 使用 `docker-compose` 或下载可执行文件运行起程序 (如果已有 nginx, 避免 443 端口占用可直接修改其他端口, 后面借助nginx 反向代理实现 https)   
2. 解析四个域名到服务器IP, 假设你的域名是: `domain.com`, 那么你需要解析的域名分别是 (**特别注意: 域名前缀不可变**): 
```
domain.com  
api.domain.com  
copilot-proxy.domain.com  
copilot-telemetry-service.domain.com  
```   
3. 将四个域名全部配置好 `SSL` 证书   
4. 配置 Nginx 反向代理或伪静态规则, 参考配置如下:  
```nginx
location ^~ /
{
    proxy_pass http://127.0.0.1:1188/;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header REMOTE-HOST $remote_addr;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "Upgrade";
    proxy_http_version 1.1;
    # proxy_hide_header Upgrade;

    add_header X-Cache $upstream_cache_status;
    
    proxy_redirect off;
    proxy_buffering off;
    proxy_max_temp_file_size 0;
    client_max_body_size 10m;
    client_body_buffer_size 128k;
    proxy_connect_timeout 90;
    proxy_send_timeout 90;
    proxy_read_timeout 90;
    proxy_buffer_size 4k;
    proxy_buffers 4 32k;
    proxy_busy_buffers_size 64k;
    proxy_temp_file_write_size 64k;

    #Set Nginx Cache
    
    
    set $static_filer5CIeZff 0;
    if ( $uri ~* "\.(gif|png|jpg|css|js|woff|woff2)$" )
    {
    	set $static_filer5CIeZff 1;
    	expires 1m;
        }
    if ( $static_filer5CIeZff = 0 )
    {
    add_header Cache-Control no-cache;
    }
}
```   
5. 最后将以上域名修改到对应的环境变量配置文件中.   
6. 最终使用 https 方式访问四个域名必须是正常的, 不能有任何问题, 否则插件无法正常使用.  

### 环境变量参数说明

详细参考: [环境变量参数说明](PARAM.md)

## IDE设置方法

### VSCode

1. 安装插件: `GitHub Copilot`
2. 修改 VSCode 的 settings.json 文件, 添加以下配置:

```json
"github.copilot.advanced": {
  "authProvider": "github-enterprise",
  "debug.overrideCAPIUrl": "https://api.copilot.supercopilot.top",
  "debug.overrideProxyUrl": "https://copilot-proxy.copilot.supercopilot.top",
  "debug.chatOverrideProxyUrl": "https://api.copilot.supercopilot.top/chat/completions",
  "debug.overrideFastRewriteEngine": "v1/engines/copilot-centralus-h100",
  "debug.overrideFastRewriteUrl": "https://api.copilot.supercopilot.top"
},
"github-enterprise.uri": "https://copilot.supercopilot.top"
```

### Jetbrains IDE系列

1. 找到`设置` > `语言与框架` > `GitHub Copilot` > `Authentication Provider`
2. 填写的值为: `copilot.supercopilot.top`

### Visual Studio 2022

1. 更新到最新版本（内置 Copilot 版本）至少是 `17.10.x` 以上
2. 首先开启 Github Enterprise 账户支持：工具-环境-账户-勾选“包含 Github Enterprise 服务器账户”
3. 然后点击添加 Github 账户，切换到 Github Enterprise 选项卡，输入 `https://copilot.supercopilot.top` 即可。

🚧 Chat服务在代码选中后右键选择解释代码会报错, 解决方法是点击一下"在聊天窗口中继续"即可.

### HBuilderX

⚠️ 注意, 插件中的相关 domain 已经写死无法修改, 所以必须使用默认的 `copilot.supercopilot.top` 域名配置.

1. 下载 **[copilot-for-hbuilderx-v1.zip](https://pan.quark.cn/s/70e6849970e5)** 插件到本地
2. 将插件安装到 plugin目录下, 详细参考: [离线插件安装指南](https://hx.dcloud.net.cn/Tutorial/OfflineInstall)
3. 重启 Hbuilder X 后点击登录 `GitHub Copilot` 即可.


## 局域网共享方案
如果是局域网多台电脑共用一个服务端，只需要更改hosts文件指向到内网服务器 ip 即可, 例如：   
在局域网服务器(192.168.80.40)部署了copilot-proxies服务，那么局域网内其他机器仅需要修改host为以下即可，可以工作。 ( [@pennbay](https://gitee.com/pennbay) 提供实测反馈 )       
```
192.168.80.40 copilot.supercopilot.top
192.168.80.40 api.copilot.supercopilot.top
192.168.80.40 copilot-proxy.copilot.supercopilot.top
192.168.80.40 copilot-telemetry-service.copilot.supercopilot.top
```

## 纯内网离线部署方案
`v0.1.0` 版本之后 ssl 证书调整为从网络上下载同步, 这对于纯内网部署造成了一些困难, 下面我提供一个简单的方案你需要做如下操作:   

- 从外网下载最新证书文件, 远程下载地址参考 [certificate.go](pkg/certificate/certificate.go), 注意证书最长只有 60 天, 需要手动更新.
- 将两个证书文件放在 `/cert` 目录下.
- 因为也无法连接公共 DNS 服务器, 所以也需要更改本机 hosts 文件, 将以下域名手动指向到本机的 `127.0.0.1`:  
  - `copilot.supercopilot.top`   
  - `api.copilot.supercopilot.top`   
  - `copilot-proxy.copilot.supercopilot.top`   
  - `copilot-telemetry-service.copilot.supercopilot.top`
- 启动主程序.

还有一种方案, 依旧使用 `v0.1.0` 之前版本的自签证书, 但这会在未来 `GitHub Copilot` 插件更新后可能无法正常使用.

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

## 使用Github Copilot官方服务

> 前提条件: 必须有官方正版的 `GitHub Copilot` 订阅权限, 否则无法使用.

**应用场景:**

- 适用于 **"月抛"** 的Github账号, 避免每个月切换Github账号后都要重复登录多个IDE中的插件操作, 只需要更改环境变量中的
  `COPILOT_GHU_TOKEN` 参数一处即可.
- 适用于 **"多人共享"** 的Github账号, 共享者只需要使用此服务即可, 不需要告知Github账号密码.

### 使用方法

- 设置环境变量参数 `COPILOT_CLIENT_TYPE=github` (设置此参数后其他的Copilot相关配置都可以不用设置了, 因为这里已经使用了官方的服务).
- 启动服务访问 [https://copilot.supercopilot.top/github/login/device/code](https://copilot.supercopilot.top/github/login/device/code) 获取 `ghu_` 的参数
- 将获取到的 `ghu_` 参数填写到 `COPILOT_GHU_TOKEN` 环境变量中.
- 重启服务, 重启IDE即可.

### 全代理模式
> 即所有请求走依旧服务端,然后由服务端发起请求到github, 在多人共享账号的情况下所有请求全部统一出口, 可以略微降低被风控的情况.

- 设置环境变量参数 `COPILOT_PROXY_ALL=true` (默认值为 `false`).
- 重启服务即可.
- 全代理模式的 `/embeddings`和 `/chunks` 接口即将推出.

**🚨 全代理模式有封号的风险, 请自行甄别谨慎使用.** 补全和对话接口的请求频率都有阀值限制的, 共享人数过多肯定会触发风控.

## Embeddings模型配置

> 目前仅 VSCode 最新版本的 `Github Copilot Chat` 插件支持使用 Embeddings 模型, 其他IDE可以不用考虑.

插件默认使用 `512维` 的Embeddings模型, 为了方便项目借助阿里的模型, 文档: [API-KEY的获取与配置](https://help.aliyun.com/zh/dashscope/developer-reference/acquisition-and-configuration-of-api-key), 获取后填写环境变量 `DASHSCOPE_API_KEY` 即可.   
注意: 阿里的Embedding模型是收费的, 但是有免费额度, 详细参考阿里的文档.   

后续将继续测试其他维度的模型和本地 `Ollama` 部署Embeddings模型进行测试, 可以关注下后续的更新. 


## 问题排查
如果本地部署遇到了 `无法登录` `无法对话` `无法补全` 等问题, 可以参考下面的排查方法:
1. 确认最新版本服务
2. `ping copilot.supercopilot.top` 是否指向 `127.0.0.1` ，如果不是则表明受到了代理工具影响
3. 访问 [https://copilot.supercopilot.top](https://copilot.supercopilot.top/help) 是否出现帮助页面
4. 检查目录下是否含有 `cert` 目录
5. 特别注意检查 https 端口不可被占用

## 注意事项

1. 请勿将本服务用于商业用途, 仅供学习交流使用
2. 请勿将本服务用于非法用途, 一切后果自负

## 鸣谢

- [copilot_take_over](https://gitee.com/LoveA/copilot_take_over)
- [override](https://github.com/linux-do/override)
- [硅基流动](https://cloud.siliconflow.cn/i/NO6ShUc3)
