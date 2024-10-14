# 环境变量参数说明

## ENV默认参数说明

| 参数                       | 描述                                                                                   | 类型     | 默认值                                             |
|--------------------------|--------------------------------------------------------------------------------------|--------|-------------------------------------------------|
| ENV                      | 当前环境 (默认: production 表示生产环境, development 表示开发环境)                                     | string | production                                      |
| PORT                     | HTTP请求的端口号 ,非必要请勿更改                                                                  | int    | 1188                                            |
| HTTPS_PORT               | HTTPS请求的端口号 ,非必要请勿更改                                                                 | int    | 443                                             |
| HOST                     | 主机地址                                                                                 | int    | 0.0.0.0                                         |
| TOKEN_SALT               | JWT秘钥                                                                                | string | 7L3Gqrn24TUWzLwG                                |
| VS_COPILOT_CLIENT_ID     | VS2022登录GitHub Copilot插件所需的客户端ID                                                     | string | a200baed193bb2088a6e                            |
| VS_COPILOT_CLIENT_SECRET | VS2022登录GitHub Copilot插件所需的客户端秘钥                                                     | string |                                                 |
| CERT_FILE                | HTTPS域名证书                                                                            | string | ssl/mycopilot.crt                               |
| KEY_FILE                 | HTTPS域名证书秘钥                                                                          | string | ssl/mycopilot.key                               |
| CODEX_API_BASE           | 代码补全服务地址 , 详细参考[代码补全服务地址](#代码补全服务地址)                                                 | string | https://api.deepseek.com/beta/v1/completions    |
| CODEX_API_KEY            | 代码补全服务的API KEY                                                                       | string |                                                 |
| CODEX_API_MODEL_NAME     | 代码补全服务的模型名称                                                                          | string |                                                 |
| CODEX_MAX_TOKENS         | 代码补全模型的最大响应tokens, 如果是Ollama建议设置小一点, 避免直接补全一长串代码                                     | int    | 500                                             |
| CODEX_TEMPERATURE        | 代码补全模型温度超参数,deepseek模型官方推荐设置为1, 如果要跟随插件动态设置,请设置为-1 (默认值为 `1`, 可以调整为 `0.1-1.0` 之间的值.) | int    | 0                                               |
| CODEX_SERVICE_TYPE       | 代码补全模型类型, 用于兼容本地模型 <br/>可选值: `default` `ollama`                                      | string | default                                         |
| COPILOT_DEBOUNCE         | 补全防抖时间, 单位:毫秒                                                                        | int    | 200                                             |
| CHAT_API_BASE            | 对话服务请求地址, 理论支持任何符合 `OpenAI` 接口规范的模型                                                  | string | https://api.deepseek.com/v1/chat/completions    |
| CHAT_API_KEY             | 对话服务请求的API KEY                                                                       | string |                                                 |
| CHAT_API_MODEL_NAME      | 对话服务请求的模型名称                                                                          | string | deepseek-chat                                   |
| CHAT_MAX_TOKENS          | 对话模型的最大响应tokens , 常见的模型响应tokens是4k, 如果支持8k可以手动调整                                     | int    | 4096                                            |
| ~~CHAT_LOCALE~~          | 回答语言, 此参数在 `v0.0.8` 版本之后废弃                                                           | string | zh_CN                                           |
| DEFAULT_BASE_URL         | 默认的服务请求地址, 必须开启https. 可以替换任何二级域名, 但后续的服务域名必须与此域名有关                                   | string | https://mycopilot.com                           |
| API_BASE_URL             | 默认的API服务请求地址, 必须开启https.  域名 `api` 前缀必须固定                                            | string | https://api.mycopilot.com                       |
| PROXY_BASE_URL           | 默认的代理服务请求地址, 必须开启https.  域名 `copilot-proxy` 前缀必须固定                                   | string | https://copilot-proxy.mycopilot.com             |
| TELEMETRY_BASE_URL       | 默认的心跳服务请求地址, 必须开启https.  域名 `copilot-telemetry-service` 前缀必须固定                       | string | https://copilot-telemetry-service.mycopilot.com |

以上环境变量参数配置可以手动在以下几个地方更改进行覆盖默认的设置:

- 二进制文件同级目录下的 `.env` 文件, 如果没有可自行创建
- 系统的环境变量中设置, 例如 `export PORT=1188`
- `docker-compose.yml` 文件中的 `environment` 配置项

## 代码补全服务地址

兼容支持 `OpenAI` Chat 接口参数规范的所有地址, 下面是一些兼容常用的地址:

| 服务地址                                                               | 描述                                       |
|--------------------------------------------------------------------|------------------------------------------|
| https://api.deepseek.com/beta/v1/completions                       | DeepSeek 官方API, 这里使用Beta地址是为了 8k 的prompt |
| https://api.siliconflow.cn/v1/completions                          | 硅基流动 官方API                               |
| https://api.mistral.ai/v1/fim/completions                          | Mistral 官方API                            |
| http://127.0.0.1:11434/v1/chat/completions                         | Ollama的Chat对话接口                          |
| http://127.0.0.1:11434/api/generate                                | Ollama代码生成, 主要适配了 `suffix` 后缀参数的模型       |
| https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions | 阿里百炼平台API                                |