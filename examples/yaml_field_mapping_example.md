# YAML 字段映射配置示例

## 功能说明

cwgo 现在支持通过 YAML 配置文件来定义字段类型映射，可以将数据库字段映射到自定义的 Go 类型，比如枚举类型。这种方式比命令行参数更加灵活和可维护。

## 配置文件格式

### YAML 配置结构

```yaml
fieldMapping:
  table_name.field_name:
    type: "CustomType"
    import: "package/path"  # 可选，如果需要导入包
```

### 示例配置文件

```yaml
# field_mapping_config.yaml
fieldMapping:
  admin_audit_log.status:
    type: "enum.AdminAuditLogStatus"
    import: "goframe/model/enum"
  admin_audit_log.type:
    type: "enum.AdminAuditLogType"
    import: "goframe/model/enum"
  users.status:
    type: "enum.UserStatus"
    import: "your-project/enums"
  orders.state:
    type: "enum.OrderState"
    import: "your-project/enums"
  products.category:
    type: "enum.ProductCategory"
    import: "your-project/enums"
```

## 使用方法

### 1. 创建配置文件

创建 `field_mapping_config.yaml` 文件，定义字段映射：

```yaml
fieldMapping:
  users.status:
    type: "enum.UserStatus"
    import: "your-project/enums"
  orders.state:
    type: "enum.OrderState"
    import: "your-project/enums"
```

### 2. 定义枚举类型

在项目中定义相应的枚举类型：

```go
// enums/user_status.go
package enums

type UserStatus int8

const (
    UserStatusDisabled UserStatus = iota
    UserStatusEnabled
    UserStatusPending
)

func (s UserStatus) String() string {
    switch s {
    case UserStatusDisabled:
        return "disabled"
    case UserStatusEnabled:
        return "enabled"
    case UserStatusPending:
        return "pending"
    default:
        return "unknown"
    }
}

// enums/order_state.go
package enums

type OrderState int8

const (
    OrderStatePending OrderState = iota
    OrderStatePaid
    OrderStateShipped
    OrderStateCompleted
)

func (o OrderState) String() string {
    switch o {
    case OrderStatePending:
        return "pending"
    case OrderStatePaid:
        return "paid"
    case OrderStateShipped:
        return "shipped"
    case OrderStateCompleted:
        return "completed"
    default:
        return "unknown"
    }
}
```

### 3. 生成模型

使用配置文件生成模型：

```bash
cwgo model \
  --db_type mysql \
  --dsn "user:password@tcp(localhost:3306)/database" \
  --config_file ./field_mapping_config.yaml \
  --out_dir ./biz/dal/model
```

### 4. 生成的模型代码

生成的模型将自动应用字段映射：

```go
// biz/dal/model/user.go
package model

import (
    "time"
    "your-project/enums"  // 自动添加的导入
)

type User struct {
    ID        int64           `gorm:"column:id;primaryKey" json:"id"`
    Name      string          `gorm:"column:name" json:"name"`
    Status    enums.UserStatus `gorm:"column:status" json:"status"`  // 映射为枚举类型
    CreatedAt time.Time       `gorm:"column:created_at" json:"created_at"`
}

// biz/dal/model/order.go
package model

import (
    "time"
    "your-project/enums"  // 自动添加的导入
)

type Order struct {
    ID        int64                `gorm:"column:id;primaryKey" json:"id"`
    UserID    int64                `gorm:"column:user_id" json:"user_id"`
    State     enums.OrderState     `gorm:"column:state" json:"state"`  // 映射为枚举类型
    Amount    decimal.Decimal      `gorm:"column:amount" json:"amount"`
    CreatedAt time.Time            `gorm:"column:created_at" json:"created_at"`
}
```

## 命令行参数

### 配置文件参数

```bash
--config_file value    # 指定 YAML 配置文件路径
```

### 其他参数

```bash
--field_mapping value  # 命令行方式指定字段映射（与配置文件二选一）
```

## 工作原理

1. **生成阶段**：cwgo 先生成标准的 GORM 模型文件
2. **后处理阶段**：读取 YAML 配置文件，对生成的 Go 文件进行后处理
3. **字段替换**：根据配置将指定字段的类型替换为自定义类型
4. **导入管理**：自动添加必要的导入包

## 支持的映射类型

- **枚举类型**（推荐）
- **自定义结构体**
- **基础类型的别名**
- **其他自定义类型**

## 配置规则

1. **字段键格式**：`table_name.field_name`
2. **类型格式**：`package.TypeName` 或 `TypeName`
3. **导入格式**：完整的包路径
4. **表名匹配**：不区分大小写

## 注意事项

1. **表名匹配**：配置文件中的表名会与生成的文件名进行匹配
2. **字段名匹配**：通过正则表达式匹配结构体字段定义
3. **导入管理**：自动添加必要的导入包，避免重复导入
4. **类型安全**：确保映射的类型在目标包中存在

## 优势

1. **配置集中**：所有字段映射集中在一个 YAML 文件中
2. **版本控制**：配置文件可以纳入版本控制
3. **团队协作**：团队成员可以共享相同的映射配置
4. **易于维护**：修改映射配置不需要重新生成代码
5. **类型安全**：编译时检查类型正确性

## 示例项目结构

```
project/
├── field_mapping_config.yaml    # 字段映射配置
├── enums/                       # 枚举类型定义
│   ├── user_status.go
│   └── order_state.go
├── biz/dal/model/               # 生成的模型文件
│   ├── user.go
│   └── order.go
└── ...
```
