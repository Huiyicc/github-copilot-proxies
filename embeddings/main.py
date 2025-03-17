from fastapi import FastAPI, Depends, HTTPException, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from sentence_transformers import SentenceTransformer
from pydantic import BaseModel
from fastapi.middleware.cors import CORSMiddleware
import uvicorn
import tiktoken
import numpy as np
from scipy.interpolate import interp1d
from typing import List, Optional
from sklearn.preprocessing import PolynomialFeatures
from sklearn.decomposition import PCA
import torch
import os

# 接口秘钥环境变量传入
sk_key = os.environ.get('sk-key', 'sk-aaabbbcccdddeeefffggghhhiiijjjkkk')
# 是否自动进行维度操作的环境变量，默认为false
auto_dim = os.environ.get('auto_dim', 'false').lower() == 'true'
# 模型名称, 必须在models文件夹下有对应的模型文件夹
model_name = os.environ.get('model_name', 'bge-m3')

# 创建一个FastAPI实例
app = FastAPI()

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 创建一个HTTPBearer实例
security = HTTPBearer()

# 预加载模型
device = torch.device('cuda' if torch.cuda.is_available() else 'cpu') # 检测是否有GPU可用，如果有则使用cuda设备，否则使用cpu设备
if torch.cuda.is_available():
    print('本次加载模型的设备为GPU: ', torch.cuda.get_device_name(0))
else:
    print('本次加载模型的设备为CPU.')

print(f'加载模型: {model_name}')
model = SentenceTransformer(f'./models/{model_name}',device=device)

# 创建PCA降维模型
pca = None

class EmbeddingRequest(BaseModel):
    input: List[str]
    model: str
    dimensions: Optional[int] = 512

class EmbeddingResponse(BaseModel):
    data: list
    model: str
    object: str
    usage: dict

def num_tokens_from_string(string: str) -> int:
    """Returns the number of tokens in a text string."""
    encoding = tiktoken.get_encoding('cl100k_base')
    num_tokens = len(encoding.encode(string))
    return num_tokens

# 插值法
def interpolate_vector(vector, target_length):
    original_indices = np.arange(len(vector))
    target_indices = np.linspace(0, len(vector)-1, target_length)
    f = interp1d(original_indices, vector, kind='linear')
    return f(target_indices)

def expand_features(embedding, target_length):
    poly = PolynomialFeatures(degree=2)
    expanded_embedding = poly.fit_transform(embedding.reshape(1, -1))
    expanded_embedding = expanded_embedding.flatten()
    if len(expanded_embedding) > target_length:
        # 如果扩展后的特征超过目标长度，可以通过截断或其他方法来减少维度
        expanded_embedding = expanded_embedding[:target_length]
    elif len(expanded_embedding) < target_length:
        # 如果扩展后的特征少于目标长度，可以通过填充或其他方法来增加维度
        expanded_embedding = np.pad(expanded_embedding, (0, target_length - len(expanded_embedding)))
    return expanded_embedding

# 降维方法：使用PCA将向量从1024维降到512维
def reduce_dimensions(embeddings, target_dim=512):
    global pca
    
    # 将列表转换为numpy数组
    embeddings_array = np.array(embeddings)
    
    # 检查样本数量
    n_samples = embeddings_array.shape[0]
    n_features = embeddings_array.shape[1]
    
    # 如果只有一个样本，无法使用PCA，改用插值法
    if n_samples == 1:
        return [interpolate_vector(embeddings_array[0], target_dim)]
    
    # 确保目标维度不超过可能的最大值
    actual_target_dim = min(target_dim, n_samples, n_features)
    if actual_target_dim < target_dim:
        print(f"警告：目标维度{target_dim}超过了可能的最大值，已调整为{actual_target_dim}")
    
    # 如果是第一次运行或者输入维度变化，重新初始化PCA
    if pca is None or pca.n_components != actual_target_dim:
        pca = PCA(n_components=actual_target_dim)
        # 先拟合再转换
        reduced_embeddings = pca.fit_transform(embeddings_array)
    else:
        # 直接使用已训练的PCA模型转换
        reduced_embeddings = pca.transform(embeddings_array)
    
    # 如果实际降维后的维度小于目标维度，使用插值法扩展
    if actual_target_dim < target_dim:
        reduced_embeddings = [interpolate_vector(embedding, target_dim) for embedding in reduced_embeddings]
    
    return list(reduced_embeddings)

@app.post("/v1/embeddings", response_model=EmbeddingResponse)
async def get_embeddings(request: EmbeddingRequest, credentials: HTTPAuthorizationCredentials = Depends(security)):

    if credentials.credentials != sk_key:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid authorization code",
        )

    # 计算嵌入向量和tokens数量
    embeddings = [model.encode(text) for text in request.input]
    
    # 检查是否需要进行维度操作
    if auto_dim:
        # 检查向量维度
        embedding_dim = len(embeddings[0])
        
        # 如果维度大于512，则降维到512
        if embedding_dim > request.dimensions:
            embeddings = reduce_dimensions(embeddings, target_dim=request.dimensions)
        # 如果维度小于512，则使用插值法扩展到512
        elif embedding_dim < request.dimensions:
            embeddings = [interpolate_vector(embedding, request.dimensions) for embedding in embeddings]
    
    # 归一化处理
    embeddings = [embedding / np.linalg.norm(embedding) for embedding in embeddings]
    # 将numpy数组转换为列表
    embeddings = [embedding.tolist() for embedding in embeddings]
    prompt_tokens = sum(len(text.split()) for text in request.input)
    total_tokens = sum(num_tokens_from_string(text) for text in request.input)

    response = {
        "data": [
            {
                "embedding": embedding,
                "index": index,
                "object": "embedding"
            } for index, embedding in enumerate(embeddings)
        ],
        "model": model_name,
        "object": "list",
        "usage": {
            "prompt_tokens": prompt_tokens,
            "total_tokens": total_tokens,
        }
    }

    return response

if __name__ == "__main__":
 uvicorn.run("main:app", host='0.0.0.0', port=6008, workers=1)
