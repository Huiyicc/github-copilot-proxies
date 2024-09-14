# Github Copilot 后端代理服务 (本地部署搭建版)

借助其他FIM模型（如DeepSeek）来接管GitHub Copilot插件服务端, 廉价的模型+强大的补全插件相结合, 使得开发者可以更加高效的编写代码。   

理论上支持任何符合 `OpenAI` 接口格式的FIM模型API, 当然也可以自己实现一个。  

## 如何使用?
### Docker【推荐】
只需要将 [docker-compose.yml](docker-compose.yml) 文件下载到本地, 将里面的**模型API KEY 替换为你的**, 然后执行以下命令即可启动服务:
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

### 手动部署【不推荐,相当繁琐】
1. 下载最新版本的可执行文件
访问 [releases](https://gitee.com/ripperTs/github-copilot-proxies/releases) 下载最新版本的可执行文件, 然后执行以下命令启动服务即可.  
需要注意的是, 在启动服务之前添加 `.env` 文件到可执行文件同级目录, 内容参考 [.env.example](.env.example) 文件。  
2. 配置Nginx服务
3. 自签证书, 域名是`*.mycopilot.com`, 并启用https
4. 配置伪静态, 代理到本地服务端口, 内容参考文件: `[default.conf](nginx/conf.d/default.conf)`


## IDE设置方法
### VSCode
1. 安装插件: `GitHub Copilot`
2. 修改 VSCode 的 settings.json 文件, 添加以下配置:
```json
{
  "github.copilot.advanced": {
    "authProvider": "github-enterprise",
    "debug.overrideCAPIUrl": "https://api.mycopilot.com",
    "debug.overrideProxyUrl": "https://copilot-proxy.mycopilot.com",
    "debug.chatOverrideProxyUrl": "https://api.mycopilot.com/chat/completions",
  },
  "github-enterprise.uri": "https://mycopilot.com"
}
```

### Jetbrains IDE系列
1. 找到`设置` > `语言与框架` > `GitHub Copilot` > `Authentication Provider`
2. 填写的值为: `mycopilot.com`

### Visual Studio 2022
**Visual Studio 2022 版本 高于17.9 的用户无法使用, 降级到历史版本, 请访问: [Visual Studio 2022 降级长绿引导程序](https://learn.microsoft.com/zh-cn/visualstudio/releases/2022/release-history#evergreen-bootstrappers)** 选择 17.8 的版本即可.   

配置系统环境变量
```shell
CODESPACES=true
GITHUB_API_URL=https://api.mycopilot.com
GITHUB_SERVER_URL=https://mycopilot.com
GITHUB_TOKEN=YOUR_GITHUB_TOKEN
AGENT_DEBUG_OVERRIDE_PROXY_URL=https://copilot-proxy.mycopilot.com
GITHUB_USER=Copilot
AGENT_DEBUG_OVERRIDE_CAPI_URL=https://api.mycopilot.com
```

### HBuilderX
待续...

## 服务器部署使用
1. 删除 `docker-compose.yml` 文件中的 `copilot-nginx` 配置
2. `docker-compose.yml` 中的 `Copilot配置` 配置项中的域名替换你真实解析的域名.
3. 配置Nginx服务, 将指定域名解析到服务器IP, 并配置伪静态, 代理到本地服务端口, 内容参考文件: `[default.conf](nginx/conf.d/default.conf)`
4. 所有解析的域名需要启用https
5. `docker-compose up -d` 启动服务即可

## 注意事项
1. 请勿将本服务用于商业用途, 仅供学习交流使用
2. 请勿将本服务用于非法用途, 一切后果自负
3. 因自签证书问题, 浏览器访问可能会提示不安全甚至拦截, 建议使用无痕窗口访问可解决大多数问题.

## 鸣谢
- [LoveA/copilot_take_over](https://gitee.com/LoveA/copilot_take_over)
- [override](https://github.com/linux-do/override)
