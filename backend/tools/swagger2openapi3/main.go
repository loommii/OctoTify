package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
)

func main() {
	input := flag.String("i", "docs/swagger/swagger.json", "输入 Swagger 2.0 JSON 文件路径")
	output := flag.String("o", "docs/swagger/openapi3.json", "输出 OpenAPI 3.0 JSON 文件路径")
	flag.Parse()

	// 读取 Swagger 2.0 JSON
	data, err := os.ReadFile(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入文件失败: %v\n", err)
		os.Exit(1)
	}

	// 解析 Swagger 2.0
	var docV2 openapi2.T
	if err := json.Unmarshal(data, &docV2); err != nil {
		fmt.Fprintf(os.Stderr, "解析 Swagger 2.0 JSON 失败: %v\n", err)
		os.Exit(1)
	}

	// 转换为 OpenAPI 3.0
	docV3, err := openapi2conv.ToV3(&docV2)
	if err != nil {
		fmt.Fprintf(os.Stderr, "转换为 OpenAPI 3.0 失败: %v\n", err)
		os.Exit(1)
	}

	// 输出 OpenAPI 3.0 JSON
	outData, err := json.MarshalIndent(docV3, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "序列化 OpenAPI 3.0 JSON 失败: %v\n", err)
		os.Exit(1)
	}

	// 确保输出目录存在
	outDir := filepath.Dir(*output)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "创建输出目录失败: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*output, outData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "写入输出文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ OpenAPI 3.0 文档已生成: %s\n", *output)
}
