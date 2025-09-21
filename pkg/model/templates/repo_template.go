package templates

// RepoImplTemplate Repo 实现模板
const RepoImplTemplate = `package repo

import (
	"context"

	"{{.ModelPkgPath}}"

	"gorm.io/gorm"
)

// {{.ModelName}}Repo {{.ModelName}} 数据访问实现
type {{.ModelName}}Repo struct {
	db *gorm.DB
}

// New{{.ModelName}}Repo 创建 {{.ModelName}} Repo
func New{{.ModelName}}Repo(db *gorm.DB) *{{.ModelName}}Repo {
	return &{{.ModelName}}Repo{db: db}
}

// Create 创建记录
func (r *{{.ModelName}}Repo) Create(ctx context.Context, {{.ModelNameLower}} *model.{{.ModelName}}) error {
	return r.db.WithContext(ctx).Create({{.ModelNameLower}}).Error
}

// GetByID 根据ID获取记录
func (r *{{.ModelName}}Repo) GetByID(ctx context.Context, id {{.IDType}}) (*model.{{.ModelName}}, error) {
	var {{.ModelNameLower}} model.{{.ModelName}}
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&{{.ModelNameLower}}).Error
	if err != nil {
		return nil, err
	}
	return &{{.ModelNameLower}}, nil
}

// Update 更新记录
func (r *{{.ModelName}}Repo) Update(ctx context.Context, {{.ModelNameLower}} *model.{{.ModelName}}) error {
	return r.db.WithContext(ctx).Save({{.ModelNameLower}}).Error
}

// Delete 删除记录
func (r *{{.ModelName}}Repo) Delete(ctx context.Context, id {{.IDType}}) error {
	return r.db.WithContext(ctx).Delete(&model.{{.ModelName}}{}, id).Error
}

// List 获取记录列表
func (r *{{.ModelName}}Repo) List(ctx context.Context, offset, limit int) ([]*model.{{.ModelName}}, error) {
	var {{.ModelNameLower}}s []*model.{{.ModelName}}
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&{{.ModelNameLower}}s).Error
	return {{.ModelNameLower}}s, err
}

// Count 获取记录总数
func (r *{{.ModelName}}Repo) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.{{.ModelName}}{}).Count(&count).Error
	return count, err
}
`

// RepoTestImplTemplate Repo 单元测试模板
const RepoTestImplTemplate = `package repo

import (
	"context"
	"testing"
	"time"

	model "{{.ModelPkgPath}}"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestNew{{.ModelName}}Repo(t *testing.T) {
	db := setupTestDB(t)
	// 自动迁移表结构
	err := db.AutoMigrate(&model.{{.ModelName}}{})
	require.NoError(t, err)
	
	repo := New{{.ModelName}}Repo(db)
	
	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func Test{{.ModelName}}Repo_Create(t *testing.T) {
	db := setupTestDB(t)
	// 自动迁移表结构
	err := db.AutoMigrate(&model.{{.ModelName}}{})
	require.NoError(t, err)
	
	repo := New{{.ModelName}}Repo(db)
	ctx := context.Background()
	
	// 测试创建记录
	{{.ModelNameLower}} := &model.{{.ModelName}}{
		// 根据实际模型字段设置测试数据
		// 例如: Name: "test", Email: "test@example.com"
	}
	
	err := repo.Create(ctx, {{.ModelNameLower}})
	assert.NoError(t, err)
	assert.NotZero(t, {{.ModelNameLower}}.ID)
}

func Test{{.ModelName}}Repo_GetByID(t *testing.T) {
	db := setupTestDB(t)
	// 自动迁移表结构
	err := db.AutoMigrate(&model.{{.ModelName}}{})
	require.NoError(t, err)
	
	repo := New{{.ModelName}}Repo(db)
	ctx := context.Background()
	
	// 先创建一个记录
	{{.ModelNameLower}} := &model.{{.ModelName}}{
		// 根据实际模型字段设置测试数据
	}
	err := repo.Create(ctx, {{.ModelNameLower}})
	require.NoError(t, err)
	
	// 测试根据ID获取记录
	result, err := repo.GetByID(ctx, {{.ModelNameLower}}.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, {{.ModelNameLower}}.ID, result.ID)
}

func Test{{.ModelName}}Repo_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := New{{.ModelName}}Repo(db)
	ctx := context.Background()
	
	// 测试获取不存在的记录
	result, err := repo.GetByID(ctx, 999)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func Test{{.ModelName}}Repo_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := New{{.ModelName}}Repo(db)
	ctx := context.Background()
	
	// 先创建一个记录
	{{.ModelNameLower}} := &model.{{.ModelName}}{
		// 根据实际模型字段设置测试数据
	}
	err := repo.Create(ctx, {{.ModelNameLower}})
	require.NoError(t, err)
	
	// 更新记录
	{{.ModelNameLower}}.UpdatedAt = time.Now()
	err = repo.Update(ctx, {{.ModelNameLower}})
	assert.NoError(t, err)
}

func Test{{.ModelName}}Repo_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := New{{.ModelName}}Repo(db)
	ctx := context.Background()
	
	// 先创建一个记录
	{{.ModelNameLower}} := &model.{{.ModelName}}{
		// 根据实际模型字段设置测试数据
	}
	err := repo.Create(ctx, {{.ModelNameLower}})
	require.NoError(t, err)
	
	// 删除记录
	err = repo.Delete(ctx, {{.ModelNameLower}}.ID)
	assert.NoError(t, err)
	
	// 验证记录已被删除
	_, err = repo.GetByID(ctx, {{.ModelNameLower}}.ID)
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func Test{{.ModelName}}Repo_List(t *testing.T) {
	db := setupTestDB(t)
	repo := New{{.ModelName}}Repo(db)
	ctx := context.Background()
	
	// 创建多个测试记录
	for i := 0; i < 5; i++ {
		{{.ModelNameLower}} := &model.{{.ModelName}}{
			// 根据实际模型字段设置测试数据
		}
		err := repo.Create(ctx, {{.ModelNameLower}})
		require.NoError(t, err)
	}
	
	// 测试获取列表
	results, err := repo.List(ctx, 0, 10)
	assert.NoError(t, err)
	assert.Len(t, results, 5)
}

func Test{{.ModelName}}Repo_List_WithPagination(t *testing.T) {
	db := setupTestDB(t)
	repo := New{{.ModelName}}Repo(db)
	ctx := context.Background()
	
	// 创建多个测试记录
	for i := 0; i < 10; i++ {
		{{.ModelNameLower}} := &model.{{.ModelName}}{
			// 根据实际模型字段设置测试数据
		}
		err := repo.Create(ctx, {{.ModelNameLower}})
		require.NoError(t, err)
	}
	
	// 测试分页
	results, err := repo.List(ctx, 0, 5)
	assert.NoError(t, err)
	assert.Len(t, results, 5)
	
	results, err = repo.List(ctx, 5, 5)
	assert.NoError(t, err)
	assert.Len(t, results, 5)
}

func Test{{.ModelName}}Repo_Count(t *testing.T) {
	db := setupTestDB(t)
	repo := New{{.ModelName}}Repo(db)
	ctx := context.Background()
	
	// 创建多个测试记录
	for i := 0; i < 3; i++ {
		{{.ModelNameLower}} := &model.{{.ModelName}}{
			// 根据实际模型字段设置测试数据
		}
		err := repo.Create(ctx, {{.ModelNameLower}})
		require.NoError(t, err)
	}
	
	// 测试计数
	count, err := repo.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}
`

// RepoTestUtilTemplate Repo 测试工具模板
const RepoTestUtilTemplate = `package repo

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {
     _ = os.Setenv("C_PRE", "../../conf")
     _ = os.Setenv("GO_ENV", "dev")
}

// setupTestDB 设置测试数据库连接
func setupTestDB(t *testing.T) *gorm.DB {
	// 从环境变量或配置中获取MySQL连接信息
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		// 默认的测试数据库连接信息
		dsn = "test_user:test_password@tcp(127.0.0.1:3306)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
	}
	
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	
	return db
}
`
