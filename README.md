# 设备借还管理系统 API

基于 Go + Gin + GORM + SQLite 实现的内部设备借还管理 API 系统。

## 功能特性

- **用户模块**: JWT 登录认证，角色区分（管理员/普通员工）
- **设备管理**: 设备的录入、编辑、删除、查询（支持分类、状态、关键词筛选）
- **借还管理**: 借用申请、归还登记、逾期自动检测、借还记录查询
- **统计报表**: 管理员专属统计接口，包含在借数量、逾期数量、分类统计等
- **统一响应**: 所有接口返回标准 JSON 格式
- **权限控制**: 基于角色的访问控制（RBAC）

## 快速开始

### 启动服务

```bash
# 方式一：直接运行
go run main.go

# 方式二：编译后运行
go build -o equipment-borrow-system.exe
./equipment-borrow-system.exe
```

服务启动后访问：`http://localhost:8080`

### 默认账号

系统启动时自动创建以下测试账号：

| 用户名 | 密码 | 角色 |
|--------|------|------|
| admin | admin123 | 管理员 |
| employee | employee123 | 普通员工 |

## 统一响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

- `code`: 0 表示成功，非 0 表示错误
- `message`: 状态描述
- `data`: 响应数据（可选）

## API 接口列表

### 1. 认证接口

#### 登录
```
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

响应：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": 1,
      "username": "admin",
      "name": "系统管理员",
      "role": "admin",
      "created_at": "2024-01-01T00:00:00Z"
    }
  }
}
```

#### 注册
```
POST /api/auth/register
Content-Type: application/json

{
  "username": "zhangsan",
  "password": "123456",
  "name": "张三",
  "role": "employee"
}
```

### 2. 用户接口（需要登录）

#### 获取个人信息
```
GET /api/user/profile
Authorization: Bearer <token>
```

#### 获取所有用户（仅管理员）
```
GET /api/user/all
Authorization: Bearer <admin_token>
```

### 3. 设备管理接口

#### 查询设备列表
```
GET /api/device
Authorization: Bearer <token>

Query 参数：
- category: 设备分类（可选）
- status: 设备状态（可选，available/borrowed）
- keyword: 关键词搜索（可选，匹配名称/序列号）
```

#### 获取设备详情
```
GET /api/device/:id
Authorization: Bearer <token>
```

#### 创建设备（仅管理员）
```
POST /api/device
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "name": "MacBook Pro 14",
  "category": "笔记本电脑",
  "serial_number": "MBP-2024-001",
  "description": "Apple MacBook Pro 14寸 M3芯片"
}
```

#### 编辑设备（仅管理员）
```
PUT /api/device/:id
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "name": "MacBook Pro 14",
  "category": "笔记本电脑",
  "serial_number": "MBP-2024-001",
  "description": "新描述"
}
```

#### 删除设备（仅管理员）
```
DELETE /api/device/:id
Authorization: Bearer <admin_token>
```

### 4. 借还管理接口

#### 我的借还记录
```
GET /api/borrow/my
Authorization: Bearer <token>

Query 参数：
- status: 状态筛选（borrowed/returned/overdue_returned）
```

#### 所有借还记录（仅管理员）
```
GET /api/borrow/all
Authorization: Bearer <admin_token>

Query 参数：
- status: 状态筛选
- user_id: 用户ID筛选
- device_id: 设备ID筛选
```

#### 逾期记录（仅管理员）
```
GET /api/borrow/overdue
Authorization: Bearer <admin_token>
```

#### 获取借还记录详情
```
GET /api/borrow/:id
Authorization: Bearer <token>
```

#### 申请借用
```
POST /api/borrow
Authorization: Bearer <token>
Content-Type: application/json

{
  "device_id": 1,
  "expected_return": "2024-12-31T23:59:59Z"
}
```

#### 归还设备
```
POST /api/borrow/return
Authorization: Bearer <token>
Content-Type: application/json

{
  "record_id": 1
}
```

### 5. 统计接口（仅管理员）

#### 总览统计
```
GET /api/stats/overview
Authorization: Bearer <admin_token>
```

响应：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total_devices": 100,
    "available_devices": 80,
    "borrowed_devices": 20,
    "total_users": 50,
    "active_borrows": 20,
    "overdue_count": 3,
    "total_records": 500,
    "returned_records": 480,
    "borrow_rate": 20.0,
    "overdue_rate": 15.0
  }
}
```

#### 分类统计
```
GET /api/stats/category
Authorization: Bearer <admin_token>
```

#### 逾期详情
```
GET /api/stats/overdue-details
Authorization: Bearer <admin_token>
```

## 数据库设计

### users 表
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| username | string | 用户名（唯一） |
| password | string | 密码（bcrypt加密） |
| name | string | 姓名 |
| role | string | 角色（admin/employee） |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

### devices 表
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| name | string | 设备名称 |
| category | string | 设备分类 |
| serial_number | string | 序列号（唯一） |
| status | string | 状态（available/borrowed） |
| description | string | 设备描述 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

### borrow_records 表
| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| user_id | uint | 借用人ID |
| device_id | uint | 设备ID |
| borrow_date | datetime | 借出时间 |
| expected_return | datetime | 预计归还时间 |
| actual_return | datetime | 实际归还时间 |
| status | string | 状态（borrowed/returned/overdue_returned） |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

## 测试示例

### 使用 curl 测试

```bash
# 1. 登录获取 token
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.data.token')

# 2. 创建设备
curl -X POST http://localhost:8080/api/device \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "ThinkPad X1 Carbon",
    "category": "笔记本电脑",
    "serial_number": "TP-X1-001",
    "description": "联想ThinkPad X1 Carbon Gen 11"
  }'

# 3. 查询设备列表
curl http://localhost:8080/api/device?category=笔记本电脑 \
  -H "Authorization: Bearer $TOKEN"

# 4. 借用设备（使用员工账号）
EMP_TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"employee","password":"employee123"}' | jq -r '.data.token')

curl -X POST http://localhost:8080/api/borrow \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $EMP_TOKEN" \
  -d '{
    "device_id": 1,
    "expected_return": "2024-12-31T23:59:59Z"
  }'

# 5. 查看统计（管理员）
curl http://localhost:8080/api/stats/overview \
  -H "Authorization: Bearer $TOKEN"
```

## 项目结构

```
project33/
├── main.go                    # 入口文件
├── config/
│   └── database.go            # 数据库配置
├── controllers/
│   ├── user_controller.go     # 用户模块
│   ├── device_controller.go   # 设备管理
│   ├── borrow_controller.go   # 借还管理
│   └── stats_controller.go    # 统计模块
├── middleware/
│   └── auth.go                # 认证中间件
├── models/
│   └── models.go              # 数据模型
├── routes/
│   └── routes.go              # 路由配置
├── utils/
│   ├── response.go            # 响应格式
│   └── jwt.go                 # JWT工具
├── go.mod
├── go.sum
└── equipment.db               # SQLite数据库（自动创建）
```

## 技术栈

- **Go 1.20+**
- **Gin v1.9+**: Web 框架
- **GORM v1.26+**: ORM 框架
- **SQLite3**: 数据库
- **golang-jwt/v5**: JWT 认证
- **golang.org/x/crypto**: 密码加密（bcrypt）
