# 花小龙打车 - H5乘客端

基于 Vue3 + Vite + Vant 构建的移动端H5乘客端应用。

## 功能特性

### 已实现功能

#### 1. 登录/注册模块
- 启动页（Logo展示）
- 手机号登录
- 短信验证码输入
- 新用户注册
- JWT Token 管理

#### 2. 首页地图模块
- 地图展示（高德地图集成）
- 地址搜索
- 当前位置定位
- 历史地址记录
- 快捷叫车入口

#### 3. 下单流程
- 目的地选择
- 车型选择（特惠快车/快车/专车）
- 价格预估
- 优惠券使用
- 订单备注

#### 4. 订单状态跟踪
- 等待接单页面（带动画效果）
- 司机已接单/即将到达
- 行程中（实时路线）
- 行程结束/待支付
- 支付成功评价

#### 5. 订单管理
- 订单列表（全部/进行中/已完成/已取消筛选）
- 订单详情（完整信息展示）
- 再次叫车功能

#### 6. 个人中心
- 用户信息展示
- 钱包余额管理（充值/提现）
- 交易记录查询
- 设置中心
- 退出登录

## 技术栈

- **框架**: Vue 3.5+ (Composition API)
- **构建工具**: Vite 6.x
- **UI组件库**: Vant 4.x (移动端)
- **状态管理**: Pinia 2.x
- **路由**: Vue Router 4.x
- **HTTP请求**: Axios
- **地图服务**: 高德地图 JS API 2.0
- **样式**: CSS Variables + 自定义主题

## 项目结构

```
web/user/
├── public/                 # 静态资源
│   ├── logo.png           # 应用Logo
│   └── ...
├── src/
│   ├── api/               # API接口
│   │   ├── request.js     # Axios封装
│   │   ├── auth.js        # 认证相关API
│   │   ├── user.js        # 用户相关API
│   │   └── order.js       # 订单相关API
│   ├── views/             # 页面组件
│   │   ├── Splash.vue     # 启动页
│   │   ├── login/         # 登录注册
│   │   ├── home/          # 首页
│   │   ├── order/         # 下单和订单状态
│   │   ├── orders/        # 订单列表和详情
│   │   └── profile/       # 个人中心
│   ├── stores/            # Pinia状态管理
│   ├── router/            # 路由配置
│   └── styles/            # 全局样式
├── index.html             # HTML模板
├── vite.config.js         # Vite配置
└── package.json           # 项目依赖
```

## 安装运行

### 环境要求
- Node.js >= 18.x
- npm 或 pnpm

### 安装依赖
```bash
cd web/user
npm install
# 或
pnpm install
```

### 开发模式
```bash
npm run dev
# 或
pnpm dev
```

访问: http://localhost:5174

### 生产构建
```bash
npm run build
# 或
pnpm build
```

## 后端对接

### API基础路径
- 开发环境代理: `http://localhost:8091`
- 生产环境: 根据实际部署地址配置

### 主要API接口

#### 认证模块 (无需Token)
- POST `/api/passenger/v1/auth/send-sms-code` - 发送验证码
- POST `/api/passenger/v1/auth/login-by-sms` - 短信登录
- POST `/api/passenger/v1/auth/refresh-token` - 刷新Token

#### 用户模块 (需要Token)
- GET `/api/passenger/v1/profile/me` - 获取用户信息
- POST `/api/passenger/v1/profile/real-name` - 实名认证

#### 订单模块 (需要Token)
- POST `/api/passenger/v1/orders/create` - 创建订单
- GET `/api/passenger/v1/orders/list` - 订单列表
- GET `/api/passenger/v1/orders/detail` - 订单详情
- GET `/api/passenger/v1/orders/status` - 轮询订单状态
- POST `/api/passenger/v1/orders/cancel` - 取消订单
- POST `/api/passenger/v1/orders/pay` - 发起支付
- GET `/api/passenger/v1/orders/payment-status` - 支付状态

## 移动端适配

- 使用 viewport 适配方案
- 设计稿基准: 375px
- 安全区域适配 (iPhone X 及以上)
- 触摸优化 (按钮最小点击区域 44px)

## 主题色配置

主色调采用紫色系:
- Primary: #7C3AED (紫色)
- Secondary: #F59E0B (橙黄色)
- Success: #10B981 (绿色)
- Danger: #EF4444 (红色)

## 注意事项

1. **地图Key配置**: 需要在 `Home.vue` 中配置高德地图的 Key 和安全密钥
2. **后端启动**: 确保后端服务在 8091 端口运行
3. **HTTPS**: 生产环境建议使用 HTTPS 以支持地理位置等功能

## 浏览器兼容性

- iOS Safari 12+
- Android Chrome 70+
- 微信内置浏览器
- 支付宝内置浏览器

## License

MIT
