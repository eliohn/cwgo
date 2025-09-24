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

package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/cloudwego/cwgo/pkg/consts"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

type ModelArgument struct {
	DSN               string
	Type              string
	Tables            []string
	ExcludeTables     []string
	OnlyModel         bool
	OutPath           string
	OutFile           string
	WithUnitTest      bool
	ModelPkgName      string
	FieldNullable     bool
	FieldSignable     bool
	FieldWithIndexTag bool
	FieldWithTypeTag  bool
	SQLDir            string
	ConfigFile        string               // 配置文件路径
	FieldMappings     []FieldMappingConfig // YAML 配置的字段映射
}

// FieldMappingConfig YAML 配置中的字段映射
type FieldMappingConfig struct {
	FieldKey string `yaml:"field_key"`        // 字段键，如 "admin_audit_log.status"
	Type     string `yaml:"type"`             // 目标类型
	Import   string `yaml:"import,omitempty"` // 导入包
}

func NewModelArgument() *ModelArgument {
	return &ModelArgument{
		OutPath:       consts.DefaultDbOutDir,
		OutFile:       consts.DefaultDbOutFile,
		FieldMappings: make([]FieldMappingConfig, 0),
	}
}

func (c *ModelArgument) ParseCli(ctx *cli.Context) error {
	c.DSN = ctx.String(consts.DSN)
	c.Type = strings.ToLower(ctx.String(consts.DBType))
	c.Tables = ctx.StringSlice(consts.Tables)
	c.ExcludeTables = ctx.StringSlice(consts.ExcludeTables)
	c.OnlyModel = ctx.Bool(consts.OnlyModel)
	c.OutPath = ctx.String(consts.OutDir)
	c.OutFile = ctx.String(consts.OutFile)
	c.WithUnitTest = ctx.Bool(consts.UnitTest)
	c.ModelPkgName = ctx.String(consts.ModelPkgName)
	c.FieldNullable = ctx.Bool(consts.Nullable)
	c.FieldSignable = ctx.Bool(consts.Signable)
	c.FieldWithIndexTag = ctx.Bool(consts.IndexTag)
	c.FieldWithTypeTag = ctx.Bool(consts.TypeTag)
	c.SQLDir = ctx.String(consts.SQLDir)
	c.ConfigFile = ctx.String(consts.ConfigFile)

	// 解析配置文件
	if c.ConfigFile != "" {
		if err := c.parseConfigFile(); err != nil {
			return fmt.Errorf("parse config file failed: %w", err)
		}
	}

	return nil
}

// parseConfigFile 解析 YAML 配置文件
func (c *ModelArgument) parseConfigFile() error {
	data, err := os.ReadFile(c.ConfigFile)
	if err != nil {
		return fmt.Errorf("read config file failed: %w", err)
	}

	var config struct {
		FieldMapping map[string]FieldMappingConfig `yaml:"fieldMapping"`
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("unmarshal config file failed: %w", err)
	}

	// 将 YAML 配置转换为内部格式
	for fieldKey, fieldConfig := range config.FieldMapping {
		fieldConfig.FieldKey = fieldKey
		c.FieldMappings = append(c.FieldMappings, fieldConfig)
	}

	return nil
}
