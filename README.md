# Github Copilot 后端代理服务 (本地部署搭建版)

借助其他FIM模型（如DeepSeek）来接管GitHub Copilot插件服务端, 廉价的模型+强大的补全插件相结合, 使得开发者可以更加高效的编写代码。   

理论上支持任何符合 `OpenAI` 接口格式的FIM模型API, 当然也可以自己实现一个。

## 如何使用?
> 在使用之前确保自己的环境是干净的, 也就是说不能使用过其他的激活服务, 可以先检查自己的环境变量将 `GITHUB` `COPILOT` 相关的环境变量删除, 然后将插件更新最新版本后重启IDE即可.

### Docker【推荐】
已经将nginx和服务端及自签证书的工作全部做完了, 只需要将 [docker-compose.yml](docker-compose.yml) 文件下载到本地, 将里面的**模型API KEY 替换为你的**, 然后执行以下命令即可启动服务:
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

### 配置本机hosts文件
将下面hosts配置添加到本机hosts文件中, 以便访问本地服务:
```
127.0.0.1 mycopilot.com
127.0.0.1 api.mycopilot.com
127.0.0.1 copilot-proxy.mycopilot.com
127.0.0.1 copilot-telemetry-service.mycopilot.com
```

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
  "github.copilot.advanced": {
    "authProvider": "github-enterprise",
    "debug.overrideCAPIUrl": "http://api.mycopilot.com:1188",
    "debug.overrideProxyUrl": "http://copilot-proxy.mycopilot.com:1188",
    "debug.chatOverrideProxyUrl": "http://api.mycopilot.com/chat/completions:1188"
  },
  "github-enterprise.uri": "http://mycopilot.com:1188"
```
**vscode 使用https有些问题, 并且直接使用ip好像也不行, 所以这里使用http的域名+端口的形式 (不直接使用80端口是为了防止服务冲突), 形式不重要直接粘贴进去即可.**

### Jetbrains IDE系列
1. 找到`设置` > `语言与框架` > `GitHub Copilot` > `Authentication Provider`
2. 填写的值为: `mycopilot.com`
3. 如果已经配置了系统级别的信任证书, 可以忽略下面步骤, 直接在IDE中信任即可.
![Xnip2024-09-14_13-08-17.png](docs/Xnip2024-09-14_13-08-17.png)

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
> 用于多人共享使用方案, 如果是个人使用还是推荐使用Docker部署, 然后 `hosts` 文件里面的ip配置改为服务器ip即可.

1. 删除 `docker-compose.yml` 文件中的 `copilot-nginx` 配置
2. `docker-compose.yml` 中的 `Copilot配置` 配置项中的域名替换你真实解析的域名.
3. 配置Nginx服务, 将指定域名解析到服务器IP, 并配置伪静态, 代理到本地服务端口, 内容参考文件: `[default.conf](nginx/conf.d/default.conf)`
4. 所有解析的域名需要启用https
5. `docker-compose up -d` 启动服务即可

## 信任证书
> 在正式使用之前, 推荐您信任证书, 否则vscode会出现各种各样的问题.

### Windows操作
1. 双击证书文件 [mycopilot.crt](nginx/ssl/mycopilot.crt) , 点击安装证书
2. 选择 `本地计算机` > `下一步`
3. 选择 `将所有的证书放入下列存储` > `浏览` > `受信任的发布者` > `确定` > `下一步` > `完成`

### MacOS操作
1. 打开钥匙串访问
2. 将 [mycopilot.crt](nginx/ssl/mycopilot.crt) 文件拖拽到“系统”钥匙串列表中。
3. 双击导入的证书,展开"信任"部分, 将"使用此证书时"选项改为"始终信任"。
4. 关闭窗口,系统会要求输入管理员密码以确认更改。

## 注意事项
1. 请勿将本服务用于商业用途, 仅供学习交流使用
2. 请勿将本服务用于非法用途, 一切后果自负

## 鸣谢
- [LoveA/copilot_take_over](https://gitee.com/LoveA/copilot_take_over)
- [override](https://github.com/linux-do/override)
