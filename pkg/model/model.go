/*
 * Copyright 2022 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package model

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
	"gorm.io/rawsql"

	"github.com/cloudwego/cwgo/config"
	"github.com/cloudwego/cwgo/pkg/consts"
	"github.com/cloudwego/cwgo/pkg/model/templates"

	"gorm.io/gen"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB
	err error
)

func Model(c *config.ModelArgument) error {
	dialector := config.OpenTypeFuncMap[consts.DataBaseType(c.Type)]

	if c.SQLDir != "" {
		db, err = gorm.Open(rawsql.New(rawsql.Config{
			FilePath: []string{c.SQLDir},
		}))
	} else {
		db, err = gorm.Open(dialector(c.DSN))
	}
	if err != nil {
		return err
	}

	genConfig := gen.Config{
		OutPath:           c.OutPath,
		OutFile:           c.OutFile,
		ModelPkgPath:      c.ModelPkgName,
		WithUnitTest:      c.WithUnitTest,
		FieldNullable:     c.FieldNullable,
		FieldSignable:     c.FieldSignable,
		FieldWithIndexTag: c.FieldWithIndexTag,
		FieldWithTypeTag:  c.FieldWithTypeTag,
	}

	if len(c.ExcludeTables) > 0 || c.Type == string(consts.Sqlite) {
		genConfig.WithTableNameStrategy(func(tableName string) (targetTableName string) {
			if c.Type == string(consts.Sqlite) && strings.HasPrefix(tableName, "sqlite") {
				return ""
			}
			if len(c.ExcludeTables) > 0 {
				for _, table := range c.ExcludeTables {
					if tableName == table {
						return ""
					}
				}
			}
			return tableName
		})
	}

	g := gen.NewGenerator(genConfig)

	g.UseDB(db)
	g.WithImportPkgPath("github.com/shopspring/decimal")
	// 关键：全局类型覆盖 - 必须在 UseDB 之后调用
	// 使用正确的函数签名：func(columnType gorm.ColumnType) (dataType string)
	g.WithDataTypeMap(map[string]func(columnType gorm.ColumnType) (dataType string){
		"decimal": func(columnType gorm.ColumnType) (dataType string) {
			return "decimal.Decimal"
		},
		"tinyint": func(columnType gorm.ColumnType) (dataType string) {
			// columnType.ColumnType() 形如 "tinyint(1)" 或 "tinyint(4)"
			columnTypeStr, _ := columnType.ColumnType()
			if columnTypeStr == "tinyint(1)" {
				return "int8"
			}
			return "int32"
		},
	})

	_, err := genModels(g, db, c.Tables)
	if err != nil {
		return err
	}
	//if !c.OnlyModel {
	//g.ApplyBasic(models...)
	//}
	//

	g.Execute()
	err = genRepositories(g, c.Tables)
	if err != nil {
		return err
	}
	return nil
}

func genModels(g *gen.Generator, db *gorm.DB, tables []string) (models []interface{}, err error) {
	var tablesNameList []string
	if len(tables) == 0 {
		tablesNameList, err = db.Migrator().GetTables()
		if err != nil {
			return nil, fmt.Errorf("migrator get all tables fail: %w", err)
		}
	} else {
		tablesNameList = tables
	}

	models = make([]interface{}, len(tablesNameList))
	for i, tableName := range tablesNameList {
		models[i] = g.GenerateModel(tableName)
	}
	return models, nil
}

// genRepositories 生成 Repo 层代码
func genRepositories(g *gen.Generator, tables []string) error {
	outPath := consts.DefaultDbRepoDir
	// 创建 repo 目录
	if err := os.MkdirAll(consts.DefaultDbRepoDir, 0755); err != nil {
		return fmt.Errorf("create repo directory failed: %w", err)
	}
	filePath, _ := getModelOutputPath(g)
	pkgs, _ := packages.Load(&packages.Config{
		Mode: packages.NeedName,
		Dir:  filePath,
	})
	// 首先生成测试工具文件（只生成一次）
	testUtilFile := filepath.Join(outPath, "test_util_test.go")
	if err := generateFile(testUtilFile, templates.RepoTestUtilTemplate, map[string]interface{}{}); err != nil {
		return fmt.Errorf("generate test util file failed: %w", err)
	}

	// 为每个表生成 Repo
	for _, tableName := range tables {
		modelName := toCamelCase(tableName)
		modelNameLower := toLowerCamelCase(tableName)

		data := map[string]interface{}{
			"ModelName":      modelName,
			"ModelNameLower": modelNameLower,
			"ModelPkgPath":   pkgs[0].PkgPath,
			"IDType":         "int64",
		}
		// 生成实现文件
		implFile := filepath.Join(outPath, fmt.Sprintf("%s.repo.go", modelNameLower))
		if err := generateFile(implFile, templates.RepoImplTemplate, data); err != nil {
			return fmt.Errorf("generate implementation file failed: %w", err)
		}
		// 生成测试文件
		testFile := filepath.Join(outPath, fmt.Sprintf("%s.repo_test.go", modelNameLower))
		if err := generateFile(testFile, templates.RepoTestImplTemplate, data); err != nil {
			return fmt.Errorf("generate test implementation file failed: %w", err)
		}
	}

	return nil
}
func getModelOutputPath(g *gen.Generator) (outPath string, err error) {
	if strings.Contains(g.ModelPkgPath, string(os.PathSeparator)) {
		outPath, err = filepath.Abs(g.ModelPkgPath)
		if err != nil {
			return "", fmt.Errorf("cannot parse model pkg path: %w", err)
		}
	} else {
		outPath = filepath.Join(filepath.Dir(g.OutPath), g.ModelPkgPath)
	}
	return outPath + string(os.PathSeparator), nil
}

// generateFile 根据模板生成文件
func generateFile(filename, templateStr string, data map[string]interface{}) error {
	tmpl, err := template.New("").Parse(templateStr)
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, data)
}

// toCamelCase 将表名转换为 Go 结构体名
// 例如: wallets -> Wallets, user_profiles -> UserProfiles
func toCamelCase(tableName string) string {
	// 将下划线分隔的字符串转换为驼峰命名
	parts := strings.Split(tableName, "_")
	var result strings.Builder

	for _, part := range parts {
		if len(part) > 0 {
			// 首字母大写
			result.WriteString(strings.ToUpper(part[:1]))
			if len(part) > 1 {
				result.WriteString(part[1:])
			}
		}
	}

	return result.String()
}

// toLowerCamelCase 将表名转换为首字母小写的驼峰命名
// 例如: wallets -> wallets, user_profiles -> userProfiles
func toLowerCamelCase(tableName string) string {
	// 先转换为驼峰命名
	camel := toCamelCase(tableName)

	// 将首字母转为小写
	if len(camel) > 0 {
		return strings.ToLower(camel[:1]) + camel[1:]
	}

	return camel
}
