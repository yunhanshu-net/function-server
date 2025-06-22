# 文件上传功能实现文档

## 概述

本文档描述了function-server中文件上传功能的实现方案。采用前端直接上传到七牛云的方式，后端只负责提供上传token和相关配置信息。

## 技术架构

### 上传流程
1. 前端向后端请求上传token和配置信息
2. 后端生成七牛云上传token并返回完整的配置信息
3. 前端使用七牛JS SDK直接上传文件到七牛云
4. 上传完成后前端获得文件访问URL

### 优势
- **性能优化**: 文件直接上传到CDN，不经过后端服务器
- **减少服务器压力**: 后端只处理token生成，不处理文件流
- **上传速度快**: 利用七牛云的全球CDN加速
- **安全性**: 通过token控制上传权限和过期时间
- **配置完整**: 一次请求获取所有必要的配置信息

## API接口

### 获取上传Token

**接口地址**: `GET /api/v1/file/upload-token`

**请求参数**:
```
path_prefix: string (可选) - 文件路径前缀，如 "images"、"documents" 等
```

**响应数据**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "七牛上传token字符串",
    "domain": "http://cdn.geeleo.com",
    "bucket": "geeleo",
    "path_prefix": "user123/images",
    "expires_at": 1703127456,
    "region": "z0"
  }
}
```

**响应字段说明**:
- `token`: 七牛云上传token，用于授权上传
- `domain`: CDN域名，用于构建文件访问URL
- `bucket`: 存储桶名称
- `path_prefix`: 建议的文件路径前缀（已包含用户ID）
- `expires_at`: token过期时间戳（Unix时间戳）
- `region`: 存储区域代码（z0=华东，z1=华北，z2=华南等）

## 前端集成示例

### 1. 安装七牛JS SDK

```bash
npm install qiniu-js
```

### 2. 前端上传代码示例

```javascript
import * as qiniu from 'qiniu-js'

// 获取上传token和配置
async function getUploadConfig(pathPrefix = '') {
  const response = await fetch(`/api/v1/file/upload-token?path_prefix=${pathPrefix}`, {
    headers: {
      'Authorization': 'Bearer your-jwt-token'
    }
  })
  const result = await response.json()
  
  if (result.code !== 200) {
    throw new Error(result.message)
  }
  
  return result.data
}

// 上传文件
async function uploadFile(file, pathPrefix = '') {
  try {
    // 1. 获取上传配置
    const config = await getUploadConfig(pathPrefix)
    console.log('上传配置:', config)
    
    // 2. 生成文件key（路径）
    const timestamp = Date.now()
    const randomId = Math.random().toString(36).substr(2, 8)
    const fileExt = file.name.split('.').pop()
    const fileName = `${timestamp}_${randomId}.${fileExt}`
    const fileKey = `${config.path_prefix}/${fileName}`
    
    // 3. 配置上传参数
    const qiniuConfig = {
      useCdnDomain: true,
      region: qiniu.region[config.region] || qiniu.region.z0 // 根据返回的region设置
    }
    
    const putExtra = {
      fname: file.name,
      params: {},
      mimeType: file.type
    }
    
    // 4. 执行上传
    const observable = qiniu.upload(file, fileKey, config.token, putExtra, qiniuConfig)
    
    return new Promise((resolve, reject) => {
      observable.subscribe({
        next(res) {
          console.log('上传进度:', res.total.percent)
        },
        error(err) {
          console.error('上传失败:', err)
          reject(err)
        },
        complete(res) {
          // 使用返回的domain构建完整URL
          const fileUrl = `${config.domain}/${res.key}`
          console.log('上传成功:', fileUrl)
          resolve({
            key: res.key,
            url: fileUrl,
            hash: res.hash,
            bucket: config.bucket
          })
        }
      })
    })
    
  } catch (error) {
    console.error('上传文件失败:', error)
    throw error
  }
}

// 使用示例
document.getElementById('fileInput').addEventListener('change', async (event) => {
  const file = event.target.files[0]
  if (!file) return
  
  try {
    const result = await uploadFile(file, 'user-uploads')
    console.log('文件上传成功:', result)
    // result.url 就是完整的文件访问地址
  } catch (error) {
    console.error('上传失败:', error)
  }
})
```

### 3. React组件示例

```jsx
import React, { useState } from 'react'
import * as qiniu from 'qiniu-js'

const FileUpload = () => {
  const [uploading, setUploading] = useState(false)
  const [progress, setProgress] = useState(0)
  const [fileUrl, setFileUrl] = useState('')
  const [uploadConfig, setUploadConfig] = useState(null)

  const getUploadConfig = async (pathPrefix = '') => {
    const response = await fetch(`/api/v1/file/upload-token?path_prefix=${pathPrefix}`)
    const result = await response.json()
    
    if (result.code !== 200) {
      throw new Error(result.message)
    }
    
    return result.data
  }

  const handleFileUpload = async (file) => {
    setUploading(true)
    setProgress(0)
    
    try {
      // 获取上传配置
      const config = await getUploadConfig('uploads')
      setUploadConfig(config)
      
      // 生成文件key
      const timestamp = Date.now()
      const randomId = Math.random().toString(36).substr(2, 8)
      const fileExt = file.name.split('.').pop()
      const fileName = `${timestamp}_${randomId}.${fileExt}`
      const fileKey = `${config.path_prefix}/${fileName}`
      
      // 上传配置
      const qiniuConfig = {
        useCdnDomain: true,
        region: qiniu.region[config.region] || qiniu.region.z0
      }
      
      const putExtra = {
        fname: file.name,
        params: {},
        mimeType: file.type
      }
      
      // 执行上传
      const observable = qiniu.upload(file, fileKey, config.token, putExtra, qiniuConfig)
      
      observable.subscribe({
        next: (res) => {
          setProgress(res.total.percent)
        },
        error: (err) => {
          console.error('上传失败:', err)
          setUploading(false)
        },
        complete: (res) => {
          const url = `${config.domain}/${res.key}`
          setFileUrl(url)
          setUploading(false)
          console.log('上传成功:', url)
        }
      })
      
    } catch (error) {
      console.error('上传失败:', error)
      setUploading(false)
    }
  }

  return (
    <div>
      <input
        type="file"
        onChange={(e) => {
          const file = e.target.files[0]
          if (file) handleFileUpload(file)
        }}
        disabled={uploading}
      />
      
      {uploadConfig && (
        <div style={{marginTop: '10px', fontSize: '12px', color: '#666'}}>
          <p>存储桶: {uploadConfig.bucket}</p>
          <p>区域: {uploadConfig.region}</p>
          <p>路径前缀: {uploadConfig.path_prefix}</p>
          <p>Token过期时间: {new Date(uploadConfig.expires_at * 1000).toLocaleString()}</p>
        </div>
      )}
      
      {uploading && (
        <div>
          <div>上传中... {Math.round(progress)}%</div>
          <progress value={progress} max={100} />
        </div>
      )}
      
      {fileUrl && (
        <div>
          <p>上传成功!</p>
          <img src={fileUrl} alt="上传的文件" style={{maxWidth: '300px'}} />
          <p>文件URL: <a href={fileUrl} target="_blank" rel="noopener noreferrer">{fileUrl}</a></p>
        </div>
      )}
    </div>
  )
}

export default FileUpload
```

## 配置说明

### 后端配置

在 `configs/config.json` 中配置七牛云参数：

```json
{
  "qiniu": {
    "access_key": "your-access-key",
    "secret_key": "your-secret-key", 
    "bucket": "geeleo",
    "domain": "http://cdn.geeleo.com",
    "use_https": false
  }
}
```

### 文件路径规则

上传的文件会按以下规则组织路径：
- 格式: `{用户ID}/{路径前缀}/{时间戳}_{随机ID}.{扩展名}`
- 示例: `user123/uploads/1703123456789_abc12345.jpg`

这样可以确保：
- 多租户隔离（按用户ID分离）
- 文件名唯一性（时间戳+随机ID）
- 路径可读性（可选的路径前缀）

## 接口优势

### 扩展后的接口优势
1. **一次请求获取所有信息**: 前端无需硬编码域名、桶名等配置
2. **动态配置**: 后端可以根据不同环境返回不同的配置
3. **完整的URL构建**: 前端可以直接使用返回的domain构建完整URL
4. **Token过期时间**: 前端可以知道token何时过期，提前刷新
5. **区域信息**: 前端可以根据区域选择最优的上传节点

### 使用场景示例
```javascript
// 场景1: 图片上传
const imageResult = await uploadFile(imageFile, 'images')
// 生成路径: user123/images/1703123456789_abc12345.jpg

// 场景2: 文档上传  
const docResult = await uploadFile(docFile, 'documents')
// 生成路径: user123/documents/1703123456789_def67890.pdf

// 场景3: 临时文件上传
const tempResult = await uploadFile(tempFile, 'temp')
// 生成路径: user123/temp/1703123456789_ghi12345.tmp
```

## 安全考虑

1. **Token过期时间**: 上传token设置1小时过期时间
2. **用户隔离**: 文件路径包含用户ID，确保多租户隔离
3. **权限控制**: 需要登录才能获取上传token
4. **文件大小限制**: 可在七牛云控制台设置文件大小限制
5. **路径前缀验证**: 后端可以验证path_prefix参数的合法性

## 与function-go SDK集成

上传完成后，可以将文件URL传递给function-go SDK进行处理：

```javascript
// 上传文件后调用函数处理
const uploadResult = await uploadFile(file, 'input-files')
const processResult = await callFunction('image-processor', {
  input_file_url: uploadResult.url
})
```

这样就形成了完整的文件处理流程：文件上传 → 函数处理 → 结果输出。 