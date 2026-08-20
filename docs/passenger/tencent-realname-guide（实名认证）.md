# 腾讯云实名认证接入指南

## 📋 功能概述

本模块为 XiaoLong-Ridy 乘客端提供**腾讯云身份证二要素核验**能力，确保用户提交的实名信息（姓名+身份证号）真实有效。

## 🎯 接入效果

### 接入前
```
用户提交姓名和身份证号 → 直接保存到数据库（无真实性校验）
```

### 接入后
```
用户提交姓名和身份证号 → 调用腾讯云 IdCardOCRVerification API 核验 
                       → 认证通过后保存到数据库
                       → 认证失败返回错误提示
```

## 📁 新增/修改的文件

### 新增文件
```
rpc/usersvc/internal/realname/
├── service.go                    # 实名认证接口定义
├── tencent_cloud_realname.go     # 腾讯云实现
└── service_test.go              # 单元测试

rpc/usersvc/etc/
└── usersvc.yaml.example         # 配置文件示例
```

### 修改文件
| 文件 | 修改内容 |
|------|---------|
| `rpc/usersvc/internal/config/config.go` | 添加 `TencentCloudConf` 配置结构 |
| `rpc/usersvc/internal/model/user.go` | 添加 `ErrRealNameVerifyFailed` 错误码 |
| `rpc/usersvc/internal/logic/errors.go` | 暴露新错误码 |
| `rpc/usersvc/internal/svc/service_context.go` | 注入 `RealNameVer` 依赖 |
| `rpc/usersvc/internal/logic/profile_logic.go` | 在 `SubmitRealName` 中增加核验逻辑 |
| `rpc/usersvc/main.go` | 初始化实名认证服务实例 |

## ⚙️ 配置说明

### 1. 安装依赖

```bash
go get github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/faceid
```

### 2. 配置腾讯云密钥

编辑 `rpc/usersvc/etc/usersvc.yaml`：

```yaml
tencentCloud:
  secretId: "your-tencent-cloud-secret-id"    # 必填：腾讯云 SecretId
  secretKey: "your-tencent-cloud-secret-key"   # 必填：腾讯云 SecretKey
  region: "ap-beijing"                         # 可选：地域，默认 ap-beijing
```

**推荐方式**：通过环境变量注入密钥（避免硬编码）：

```bash
export TENCENTCLOUD_SECRET_ID="your-secret-id"
export TENCENTCLOUD_SECRET_KEY="your-secret-key"
export TENCENTCLOUD_REGION="ap-beijing"
```

### 3. 开通腾讯云服务

1. 登录 [腾讯云控制台](https://console.cloud.tencent.com/)
2. 开通 **人脸核身** 服务（FaceID）
3. 获取 API 密钥（SecretId / SecretKey）
4. 确保账户余额充足（约 **0.5-1 元/次**）

## 🔧 工作原理

### 调用流程图

```
乘客 APP
    │
    ▼
POST /api/passenger/v1/profile/real-name
    │
    ▼
Passenger API (api/passenger)
    │  解析 JWT Token 获取 userID
    │  转发到 usersvc RPC
    ▼
usersvc RPC (rpc/usersvc)
    │
    ├─ 1. 参数校验（非空检查）
    │
    ├─ 2. 调用腾讯云实名认证
    │      └─ IdCardOCRVerification(Name, IdCard)
    │          ├─ Result = "0" ✅ 通过
    │          ├─ Result = "-1" ❌ 不一致
    │          └─ Result = 其他 ⚠️ 异常
    │
    ├─ 3. 认证通过 → 更新数据库
    │      UPDATE user SET real_name=?, id_card_no=? WHERE id=?
    │
    └─ 4. 返回更新后的用户信息
```

### 腾讯云返回码说明

| Result | 含义 | 是否收费 | 处理方式 |
|--------|------|---------|---------|
| `"0"` | 姓名和身份证号一致 | ✅ 收费 | 通过，保存数据 |
| `"-1"` | 姓名和身份证号不一致 | ✅ 收费 | 返回错误 |
| `"-2"` | 非法身份证号 | ❌ 不收费 | 返回错误 |
| `"-3"` | 非法姓名 | ❌ 不收费 | 返回错误 |
| `"-4"` | 证件库服务异常 | ❌ 不收费 | 返回错误 |
| `"-5"` | 证件库无此记录 | ❌ 不收费 | 返回错误 |
| `"-6"` | 权威比对系统升级中 | ❌ 不收费 | 提示稍后重试 |
| `"-7"` | 认证次数超限 | ❌ 不收费 | 提示明日再试 |

## 🧪 测试指南

### 本地开发模式（跳过核验）

不配置腾讯云密钥时，系统自动跳过实名认证：

```yaml
# rpc/usersvc/etc/usersvc.yaml
tencentCloud:
  secretId: ""       # 留空
  secretKey: ""      # 留空
  region: "ap-beijing"
```

日志输出：
```
未配置腾讯云密钥，实名认证将跳过核验
未配置实名认证服务，跳过核验
```

### 单元测试

```bash
cd rpc/usersvc
go test ./internal/realname/... -v
```

### 集成测试（需配置真实密钥）

```bash
# 1. 设置环境变量
export TENCENTCLOUD_SECRET_ID="test-secret-id"
export TENCENTCLOUD_SECRET_KEY="test-secret-key"

# 2. 启动服务
go run main.go -f etc/usersvc.yaml

# 3. 调用接口测试
curl -X POST http://localhost:8091/api/passenger/v1/profile/real-name \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-jwt-token>" \
  -d '{"realName": "张三", "idCardNo": "110101199001011234"}'
```

## 💡 最佳实践

### 1. 安全建议

- ✅ **必须**：使用环境变量或密钥管理服务存储 `SecretId` / `SecretKey`
- ❌ **禁止**：将密钥硬编码在代码或提交到 Git
- ✅ **推荐**：为不同环境（开发/测试/生产）使用不同的腾讯云子账号

### 2. 性能优化

- 腾讯云 API 响应时间约 **200-500ms**，建议：
  - 前端显示 loading 状态
  - 设置合理的超时时间（建议 5s）
  - 考虑添加重试机制（网络异常时）

### 3. 成本控制

- 仅 `Result = "0"` 或 `"-1"` 时收费
- 建议在前端做基础格式校验（身份证号长度、校验位等）
- 避免重复提交相同信息（可加幂等性检查）

### 4. 错误处理

前端应根据错误码给用户友好提示：

```javascript
// 示例错误处理
if (error.code === 'ErrRealNameVerifyFailed') {
  showError('实名认证失败，请检查姓名和身份证号是否正确');
}
```

## 🐛 常见问题

### Q1: 提示 "未开通服务"

**原因**：腾讯云账号未开通人脸核身服务  
**解决**：前往 [腾讯云控制台](https://console.cloud.tencent.com/faceid) 开通服务

### Q2: 提示 "CAM签名/鉴权错误"

**原因**：`SecretId` 或 `SecretKey` 错误  
**解决**：检查配置是否正确，确认密钥权限包含 FaceID 服务

### Q3: 提示 "账号已欠费"

**原因**：腾讯云账户余额不足  
**解决**：充值后再试

### Q4: 本地开发如何跳过核验？

**方法**：不配置 `secretId` 和 `secretKey`（留空即可），系统会自动跳过

## 📊 监控与日志

### 关键日志点

1. **调用开始**：
   ```
   实名认证开始: name=张三 idCardNo=110***1234
   ```

2. **认证结果**：
   ```
   实名认证结果: name=张三 result=0 description=姓名和身份证号一致
   ```

3. **认证失败**：
   ```
   实名认证未通过: result=-1 description=姓名和身份证号不一致
   ```

4. **调用异常**：
   ```
   腾讯云实名认证调用失败: <error details>
   ```

### 建议监控指标

- 实名认证成功率（目标 > 95%）
- 平均响应时间（目标 < 500ms）
- 日调用量及费用统计
- 失败原因分布（不一致/参数错误/系统异常）

## 🔗 相关资源

- [腾讯云人脸核身文档](https://cloud.tencent.com/document/product/1007/37980)
- [API Explorer 在线调试](https://console.cloud.tencent.com/api/explorer?Product=faceid&Version=2018-03-01&Action=IdCardOCRVerification)
- [Go SDK GitHub](https://github.com/TencentCloud/tencentcloud-sdk-go)

## 📝 更新日志

### v1.0.0 (2026-08-20)
- ✅ 初始版本
- ✅ 支持腾讯云二要素核验（姓名+身份证号）
- ✅ 支持配置开关（未配置时跳过核验）
- ✅ 支持环境变量注入密钥
- ✅ 完善的错误处理和日志记录
