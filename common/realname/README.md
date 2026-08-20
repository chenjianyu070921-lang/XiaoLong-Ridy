# common/realname - 实名认证公共包

## 📦 包概述

`common/realname` 是 XiaoLong-Ridy 项目的**实名认证公共模块**，提供统一的身份证二要素核验能力。

### 核心特性

- ✅ **接口抽象**：定义 `RealNameVerifier` 接口，支持多种实现
- ✅ **腾讯云市场集成**：基于云市场 HTTP API + HMAC-SHA1 签名认证
- ✅ **灵活配置**：支持 YAML 配置 + 环境变量注入
- ✅ **开发友好**：未配置时自动跳过核验（本地开发模式）
- ✅ **完整测试**：包含 Mock 实现和单元测试

---

## 🚀 快速开始

### 1. 无需额外安装依赖

本包仅使用 Go 标准库和项目已有依赖：
- `crypto/hmac`, `crypto/sha1` - HMAC-SHA1 签名
- `encoding/base64` - Base64 编码
- `net/http` - HTTP 客户端
- `github.com/zeromicro/go-zero/core/logx` - 日志

### 2. 在你的服务中使用

#### 步骤一：添加配置

在你的服务的 `etc/xxx.yaml` 中添加：

```yaml
tencentCloud:
  secretId: "your-tencent-cloud-market-secret-id"    # 云市场分配的密钥 Id
  secretKey: "your-tencent-cloud-market-secret-key"   # 云市场分配的密钥 Key
  region: "ap-beijing"                                # 可选，默认 ap-beijing
```

或通过环境变量（推荐）：

```bash
# Windows PowerShell
$env:TENCENTCLOUD_SECRET_ID = "your-secret-id"
$env:TENCENTCLOUD_SECRET_KEY = "your-secret-key"
$env:TENCENTCLOUD_REGION = "ap-beijing"

# Linux/Mac
export TENCENTCLOUD_SECRET_ID="your-secret-id"
export TENCENTCLOUD_SECRET_KEY="your-secret-key"
export TENCENTCLOUD_REGION="ap-beijing"
```

#### 步骤二：初始化实例

在服务的 `main.go` 或初始化函数中：

```go
import (
    "XiaoLong-Ridy/common/realname"
)

// 初始化实名认证服务
func initRealNameService(cfg realname.TencentCloudConfig) realname.RealNameVerifier {
    return realname.NewTencentCloudRealNameVerifier(cfg)
}
```

#### 步骤三：在业务逻辑中调用

```go
import (
    "XiaoLong-Ridy/common/realname"
)

type YourLogic struct {
    realNameVer realname.RealNameVerifier
}

func (l *YourLogic) SubmitRealName(name, idCardNo string) error {
    // 检查是否配置了实名认证
    if l.realNameVer == nil {
        log.Println("未配置实名认证，跳过核验")
        // 直接保存到数据库...
        return nil
    }

    // 调用腾讯云进行核验
    result, err := l.realNameVer.Verify(ctx, name, idCardNo)
    if err != nil {
        return fmt.Errorf("实名认证调用失败: %w", err)
    }

    // 判断结果
    if result.Result != "0" {
        return fmt.Errorf("实名认证失败: %s", result.Description)
    }

    // 认证通过，保存到数据库...
    return nil
}
```

---

## 📁 文件结构

```
common/realname/
├── service.go              # 接口定义（RealNameVerifier, VerifyResult）
├── tencent_cloud.go         # 腾讯云市场实现（HTTP API + HMAC-SHA1）
├── service_test.go         # 单元测试（含 Mock 实现）
└── README.md               # 本文档
```

---

## 🎯 API 接口文档

### 腾讯云市场实名认证 API

**接口地址**：`https://{region}.cloudmarket-apigw.com/service-18c38npd/idcard/VerifyIdcardv2`

**请求方式**：`POST`

**Content-Type**：`application/json`

**认证方式**：HMAC-SHA1 签名（Header: Authorization）

### 请求参数（Body - JSON 格式）

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| `cardNo` | string | 是 | 身份证号（18位） |
| `realName` | string | 是 | 真实姓名（中文） |

**请求示例**：

```json
{
  "cardNo": "110101199001011234",
  "realName": "张三"
}
```

### 请求头（Headers）

| Header | 类型 | 必填 | 描述 |
|--------|------|------|------|
| `Authorization` | string | 是 | HMAC-SHA1 签名字符串（JSON格式） |
| `X-Date` | string | 是 | GMT 时间戳（RFC1123 格式） |
| `Content-Type` | string | 是 | 固定值：`application/json` |

**签名算法**：

```go
// 1. 生成时间戳
datetime = time.Now().In(timeLocation).Format("Mon, 02 Jan 2006 15:04:05 GMT")

// 2. 构建签名字符串
signStr = fmt.Sprintf("x-date: %s", datetime)

// 3. 计算 HMAC-SHA1
mac = hmac.New(sha1.New, []byte(secretKey))
mac.Write([]byte(signStr))
sign = base64.StdEncoding.EncodeToString(mac.Sum(nil))

// 4. 构建 Authorization
auth = fmt.Sprintf(`{"id":"%s", "x-date":"%s", "signature":"%s"}`, secretId, datetime, sign)
```

### 响应参数（JSON）

**成功响应示例**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "isMatch": true
  }
}
```

**失败响应示例**：

```json
{
  "code": -1,
  "message": "姓名与身份证号不匹配"
}
```

| 字段 | 类型 | 描述 |
|------|------|------|
| `code` | int | 业务码（0=成功，其他=失败） |
| `message` | string | 结果描述 |
| `data.isMatch` | bool | 是否匹配（仅 code=0 时有效） |

---

## 📐 接口定义

### RealNameVerifier 接口

```go
type RealNameVerifier interface {
    Verify(ctx context.Context, name, idCardNo string) (*VerifyResult, error)
}
```

**参数说明**：
- `ctx`: 上下文（用于日志、超时控制等）
- `name`: 真实姓名（中文）
- `idCardNo`: 身份证号（18位）

**返回值**：
- `*VerifyResult`: 认证结果（见下方）
- `error`: 网络错误或系统异常

### VerifyResult 结构体

```go
type VerifyResult struct {
    Result      string  // 结果码: "0"=通过, "-1"=不一致, 其他=异常
    Description string  // 结果描述（如"姓名和身份证号一致"）
}
```

**结果码说明**：

| Result | 含义 | 处理建议 |
|--------|------|---------|
| `"0"` | 姓名和身份证号一致 ✅ | 通过，继续业务流程 |
| `"-1"` | 姓名和身份证号不一致 ❌ | 提示用户检查信息 |
| 其他值 | 系统异常/参数错误 ⚠️ | 检查参数或稍后重试 |

### TencentCloudConfig 配置结构体

```go
type TencentCloudConfig struct {
    SecretID  string  // 云市场分配的密钥 Id
    SecretKey string  // 云市场分配的密钥 Key
    Region    string  // 地域（默认 ap-beijing）
}
```

### NewTencentCloudRealNameVerifier 构造函数

```go
func NewTencentCloudRealNameVerifier(cfg TencentCloudConfig) *TencentCloudRealNameVerifier
```

**特性**：
- 若 `SecretID` 或 `SecretKey` 为空，返回 `nil`
- 返回 `nil` 时表示**未启用实名认证**（兼容本地开发）
- 自动设置 10 秒超时

---

## 💡 使用示例

### 示例 1：在 usersvc 中使用（已集成）

参见：`rpc/usersvc/main.go`, `rpc/usersvc/internal/logic/profile_logic.go`

### 示例 2：在 driversvc 中使用（司机实名认证）

```go
// rpc/driversvc/main.go
import commonRealName "XiaoLong-Ridy/common/realname"

func newServiceContext(c config.Config) *svc.ServiceContext {
    // ... 其他初始化
    
    // 初始化实名认证
    realNameVer := commonRealName.NewTencentCloudRealNameVerifier(c.TencentCloud)
    
    return svc.NewServiceContext(c, ..., realNameVer)
}
```

```go
// rpc/driversvc/internal/logic/driver_logic.go
func (l *SubmitDriverRealNameLogic) SubmitDriverRealName(...) error {
    // 调用公共包的接口
    result, err := l.svcCtx.RealNameVer.Verify(l.ctx, name, idCardNo)
    if err != nil {
        return err
    }
    
    if result.Result != "0" {
        return errors.New("实名认证失败")
    }
    
    // 保存到数据库...
    return nil
}
```

### 示例 3：自定义实现（如阿里云）

```go
package myrealname

import (
    "context"
    "XiaoLong-Ridy/common/realname"
)

// AliyunRealNameVerifier 阿里云实名认证实现
type AliyunRealNameVerifier struct {
    client *aliyun.Client
}

func (v *AliyunRealNameVerifier) Verify(ctx context.Context, name, idCardNo string) (*realname.VerifyResult, error) {
    // 调用阿里云 API...
    return &realname.VerifyResult{
        Result:      "0",
        Description: "认证通过",
    }, nil
}
```

---

## 🧪 测试

### 运行单元测试

```bash
cd common/realname
go test -v ./...
```

### 使用 Mock 进行测试

```go
func TestYourLogic(t *testing.T) {
    // 创建 Mock 实例
    mockVerifier := &MockRealNameVerifier{
        MockResult: &realname.VerifyResult{
            Result:      "0",
            Description: "姓名和身份证号一致",
        },
    }
    
    // 注入到你的逻辑中
    logic := &YourLogic{
        realNameVer: mockVerifier,
    }
    
    // 执行测试
    err := logic.SubmitRealName("张三", "110101199001011234")
    assert.NoError(t, err)
}
```

---

## ⚙️ 配置最佳实践

### 1. 生产环境（必须配置）

```yaml
# etc/prod.yaml
tencentCloud:
  secretId: "${TENCENTCLOUD_SECRET_ID}"      # 从环境变量读取
  secretKey: "${TENCENTCLOUD_SECRET_KEY}"     # 从环境变量读取
  region: "ap-beijing"
```

### 2. 开发环境（可选跳过）

```yaml
# etc/dev.yaml
tencentCloud:
  secretId: ""     # 留空即可跳过
  secretKey: ""
  region: "ap-beijing"
```

### 3. 测试环境（使用 Mock）

```go
// 在测试代码中使用 Mock 实现
mockVerifier := &MockRealNameVerifier{...}
logic := &YourLogic{realNameVer: mockVerifier}
```

---

## 🔒 安全注意事项

### ⚠️ 必须遵守

1. **禁止硬编码密钥**
   ```go
   // ❌ 错误做法
   cfg.SecretID = "axMeMqAqC7dSjc"
   
   // ✅ 正确做法
   cfg.SecretID = os.Getenv("TENCENTCLOUD_SECRET_ID")
   ```

2. **使用最小权限原则**
   - 为不同环境创建不同的云市场密钥
   - 定期轮换密钥
   - 不要将密钥提交到 Git/GitHub

3. **敏感信息脱敏**
   ```go
   log.Printf("实名认证结果: name=%s idCardNo=%s", name, maskIdCard(idCardNo))
   ```

4. **费用控制**
   - 监控日调用量（免费额度：1000次/年）
   - 设置费用告警阈值
   - 避免重复提交相同信息

---

## 📊 监控与日志

### 关键日志点

```go
// 1. 初始化日志
log.Printf("实名认证服务初始化: enabled=%v", realNameVer != nil)

// 2. API 调用响应
logx.WithContext(ctx).Infof("实名认证API响应: status=%d body=%s", resp.StatusCode, body)

// 3. 异常情况
logx.WithContext(ctx).Errorf("调用实名认证API失败: %v", err)
```

### 建议监控指标

| 指标 | 目标值 | 告警阈值 |
|------|--------|---------|
| 成功率 | > 95% | < 90% |
| 平均响应时间 | < 1000ms | > 3000ms |
| 日调用量 | - | > 900 次（接近免费额度上限） |

---

## 🐛 常见问题

### Q1: 如何在本地开发时跳过核验？

**A**: 不配置 `secretId` 和 `secretKey`（留空），`NewTencentCloudRealNameVerifier` 会返回 `nil`。

### Q2: 如何切换到其他云服务商？

**A**: 实现 `RealNameVerifier` 接口即可：

```go
type MyCustomVerifier struct{}

func (v *MyCustomVerifier) Verify(ctx context.Context, name, idCardNo string) (*realname.VerifyResult, error) {
    // 你的实现...
}
```

### Q3: 免费额度是多少？

**A**: 
- 当前套餐：**1000 次/年**（有效期至 2027-08-20）
- 超出后按量计费（约 0.5 元/次）
- 建议在前端做基础格式校验以减少无效调用

### Q4: 支持哪些地域？

**A**: 
- 默认：`ap-beijing`
- 根据实际情况调整（通常不需要修改）

### Q5: 签名错误怎么办？

**A**: 检查以下项：
- SecretId 和 SecretKey 是否正确
- 时间戳是否为 GMT 格式
- HMAC-SHA1 算法是否正确实现

---

## 🔗 相关资源

- [腾讯云市场 - 身份证实名认证产品页](https://market.cloud.tencent.com/products/xxxx)
- [腾讯云市场 API 签名文档](https://cloud.tencent.com/document/product/xxx)
- [HMAC-SHA1 签名算法说明](https://tools.ietf.org/html/rfc2104)

---

## 📝 更新日志

### v1.1.0 (2026-08-20)
- 🔧 **重构**：从 FaceID SDK 改为腾讯云市场 HTTP API
- 🔄 **变更**：请求体格式从 Form 改为 JSON
- ✨ **新增**：HMAC-SHA1 签名认证机制
- 📝 **完善**：完整的 API 接口文档和示例代码

### v1.0.0 (2026-08-20)
- ✅ 初始版本发布
- ✅ 支持腾讯云二要素核验
- ✅ 接口抽象设计，支持多实现
- ✅ 完整的单元测试和文档
