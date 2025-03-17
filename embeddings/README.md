# 本地 Embeddings 模型API服务

下载本地 `Embedding模型` 并转为 `OpenAI` 接口格式的 API 服务。   

## 准备工作
- Python 3.9+
- 选择合适的模型文件 (根据效果自行测试), 程序支持自动提升维度或降级维度到指定维度(接口中传递的 `dimensions` 参数, 默认为512)
- 下载模型文件，放置在 `./models` 目录下, 国内下载可以去 [魔搭社区](https://www.modelscope.cn/models/BAAI/bge-m3), 速度不受影响

## 环境变量参数
- `sk-key`: 服务的 `API KEY`，默认为 `sk-aaabbbcccdddeeefffggghhhiiijjjkkk`
- `auto_dim`: 是否自动进行维度操作, 若为 `true` 则会自动提升或降级维度到512, 默认为 `false`
- `model_path`: 模型目录名称, 默认为 `bge-m3`, 注意必须在models文件夹下有对应的模型文件夹


## 运行服务
```shell
pip install -r requirements.txt
```

```shell
python app.py
```

```bash
curl --location --request POST 'http://127.0.0.1:6008/v1/embeddings' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer sk-aaabbbcccdddeeefffggghhhiiijjjkkk' \
--data-raw '{
    "input": [
        "解析当前项目"
    ],
    "model": "text-embedding-3-small",
    "dimensions": 512
}'
```
