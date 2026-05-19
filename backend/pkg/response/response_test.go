// Package response 提供统一响应处理的单元测试
package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"octotify/pkg/xerr"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// setupTestContext 创建用于测试的 Gin 上下文和响应记录器
func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return c, w
}

// TestSuccess 测试成功响应函数
func TestSuccess(t *testing.T) {
	c, w := setupTestContext()

	// 准备测试数据并调用 Success 函数
	testData := map[string]string{"key": "value"}
	Success(c, testData)

	// 验证 HTTP 状态码为 200
	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// 解析响应体并验证结构
	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// 验证业务状态码为 0（成功）
	if resp.Code != 0 {
		t.Errorf("expected code 0, got %d", resp.Code)
	}

	// 验证成功提示信息
	if resp.Msg != "请求成功" {
		t.Errorf("expected msg '请求成功', got '%s'", resp.Msg)
	}

	// 验证响应数据不为空
	if resp.Data == nil {
		t.Fatal("expected data to be non-nil")
	}

	// 验证数据内容正确
	dataMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be map, got %T", resp.Data)
	}

	if dataMap["key"] != "value" {
		t.Errorf("expected data key 'value', got '%v'", dataMap["key"])
	}
}

// TestSuccessWithNilData 测试成功响应中数据为空的情况
func TestSuccessWithNilData(t *testing.T) {
	c, w := setupTestContext()

	Success(c, nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Code != 0 {
		t.Errorf("expected code 0, got %d", resp.Code)
	}

	if resp.Msg != "请求成功" {
		t.Errorf("expected msg '请求成功', got '%s'", resp.Msg)
	}
}

// TestSuccessWithMsg 测试带自定义消息的成功响应
func TestSuccessWithMsg(t *testing.T) {
	c, w := setupTestContext()

	testData := []int{1, 2, 3}
	SuccessWithMsg(c, "操作成功", testData)

	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Code != 0 {
		t.Errorf("expected code 0, got %d", resp.Code)
	}

	// 验证自定义消息正确返回
	if resp.Msg != "操作成功" {
		t.Errorf("expected msg '操作成功', got '%s'", resp.Msg)
	}
}

// TestFail 测试失败响应函数
func TestFail(t *testing.T) {
	c, w := setupTestContext()

	Fail(c, 100001, "未登录")

	// 失败响应仍返回 HTTP 200，业务 code 在响应体中
	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Code != 100001 {
		t.Errorf("expected code 100001, got %d", resp.Code)
	}

	if resp.Msg != "未登录或Token已过期" {
		t.Errorf("expected msg '未登录或Token已过期', got '%s'", resp.Msg)
	}

	// 失败响应中 data 字段应为空
	if resp.Data != nil {
		t.Errorf("expected data to be nil, got %v", resp.Data)
	}
}

// TestUnauthorized 测试 JWT 鉴权失败响应
func TestUnauthorized(t *testing.T) {
	c, w := setupTestContext()

	Unauthorized(c, xerr.CodeUnauthorized)

	// 鉴权失败返回 HTTP 401
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status code %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Code != xerr.CodeUnauthorized {
		t.Errorf("expected code %d, got %d", xerr.CodeUnauthorized, resp.Code)
	}

	if resp.Msg != "未登录或Token已过期" {
		t.Errorf("expected msg '未登录或Token已过期', got '%s'", resp.Msg)
	}
}

// TestFailWithData 测试带附加数据的失败响应
func TestFailWithData(t *testing.T) {
	c, w := setupTestContext()

	errorDetails := map[string]string{"field": "username"}
	FailWithData(c, 100000, "参数错误", errorDetails)

	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Code != 100000 {
		t.Errorf("expected code 100000, got %d", resp.Code)
	}

	if resp.Msg != "请求参数错误" {
		t.Errorf("expected msg '请求参数错误', got '%s'", resp.Msg)
	}

	// 验证附加数据正确返回
	dataMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be map, got %T", resp.Data)
	}

	if dataMap["field"] != "username" {
		t.Errorf("expected data field 'username', got '%v'", dataMap["field"])
	}
}

// TestSuccessWithPage 测试分页成功响应
func TestSuccessWithPage(t *testing.T) {
	c, w := setupTestContext()

	list := []map[string]string{
		{"id": "1", "name": "item1"},
		{"id": "2", "name": "item2"},
	}

	SuccessWithPage(c, list, 10, 1, 20)

	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Code != 0 {
		t.Errorf("expected code 0, got %d", resp.Code)
	}

	// 验证分页数据结构
	dataMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be map, got %T", resp.Data)
	}

	if int(dataMap["total"].(float64)) != 10 {
		t.Errorf("expected total 10, got %v", dataMap["total"])
	}

	if int(dataMap["page"].(float64)) != 1 {
		t.Errorf("expected page 1, got %v", dataMap["page"])
	}

	if int(dataMap["page_size"].(float64)) != 20 {
		t.Errorf("expected pageSize 20, got %v", dataMap["page_size"])
	}
}

// TestHandleValidationError_WithValidationErrors 测试参数校验错误处理（多字段校验失败）
func TestHandleValidationError_WithValidationErrors(t *testing.T) {
	c, w := setupTestContext()

	validate := validator.New()

	// 定义测试结构体，包含多个校验规则
	type TestStruct struct {
		Name  string `validate:"required"`
		Email string `validate:"email"`
		Age   int    `validate:"min=18"`
	}

	testObj := TestStruct{
		Name:  "",              // 违反 required 规则
		Email: "invalid-email", // 违反 email 规则
		Age:   10,              // 违反 min=18 规则
	}

	err := validate.Struct(testObj)
	if err == nil {
		t.Fatal("expected validation error")
	}

	HandleValidationError(c, err)

	// 校验失败仍返回 HTTP 200
	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// 验证业务状态码为 ErrBadRequest
	if resp.Code != xerr.ErrBadRequest.Code {
		t.Errorf("expected code %d, got %d", xerr.ErrBadRequest.Code, resp.Code)
	}

	// 验证校验失败提示信息
	if resp.Msg != "请求参数校验失败" {
		t.Errorf("expected msg '请求参数校验失败', got '%s'", resp.Msg)
	}

	// 验证返回字段级错误列表
	errorsList, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected data to be []interface{}, got %T", resp.Data)
	}

	if len(errorsList) == 0 {
		t.Error("expected field errors, got empty list")
	}

	// 验证错误项包含 field 和 message 字段
	firstErr, ok := errorsList[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected first error to be map, got %T", errorsList[0])
	}
	if firstErr["field"] == nil {
		t.Error("expected error to have 'field' key")
	}
	if firstErr["message"] == nil {
		t.Error("expected error to have 'message' key")
	}
}

// TestHandleValidationError_WithNonValidationError 测试非校验错误的处理
func TestHandleValidationError_WithNonValidationError(t *testing.T) {
	c, w := setupTestContext()

	// 传入普通错误（非 validator.ValidationErrors 类型）
	// 这会触发 c.Error() 分支，不写入 JSON 响应体
	customErr := errors.New("some internal error")
	HandleValidationError(c, customErr)

	// c.Error() 不写入响应体，所以 body 应为空
	if w.Body.Len() > 0 {
		t.Errorf("expected empty body for non-validation error, got: %s", w.Body.String())
	}

	// 验证错误已附加到 Gin 上下文
	if len(c.Errors) == 0 {
		t.Error("expected error to be attached to context")
	}
}

// TestMsgForTag_Required 测试 required 校验标签的中文错误信息翻译
func TestMsgForTag_Required(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	type TestStruct struct {
		Name string `validate:"required"`
	}

	testObj := TestStruct{Name: ""}
	validate := validator.New()
	err := validate.Struct(testObj)

	if err == nil {
		t.Fatal("expected validation error")
	}

	HandleValidationError(c, err)

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	errorsList, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected data to be []interface{}, got %T", resp.Data)
	}

	if len(errorsList) > 0 {
		firstErr, ok := errorsList[0].(map[string]interface{})
		if !ok {
			t.Fatalf("expected first error to be map, got %T", errorsList[0])
		}
		// 验证 required 标签翻译为"不能为空"
		if firstErr["message"] != "不能为空" {
			t.Errorf("expected message '不能为空', got '%s'", firstErr["message"])
		}
	}
}

// TestMsgForTag_Email 测试 email 校验标签的中文错误信息翻译
func TestMsgForTag_Email(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	type TestStruct struct {
		Email string `validate:"email"`
	}

	testObj := TestStruct{Email: "not-an-email"}
	validate := validator.New()
	err := validate.Struct(testObj)

	if err == nil {
		t.Fatal("expected validation error")
	}

	HandleValidationError(c, err)

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	errorsList, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected data to be []interface{}, got %T", resp.Data)
	}

	if len(errorsList) > 0 {
		firstErr, ok := errorsList[0].(map[string]interface{})
		if !ok {
			t.Fatalf("expected first error to be map, got %T", errorsList[0])
		}
		// 验证 email 标签翻译为"邮箱格式不正确"
		if firstErr["message"] != "邮箱格式不正确" {
			t.Errorf("expected message '邮箱格式不正确', got '%s'", firstErr["message"])
		}
	}
}

// TestMsgForTag_Min 测试 min 校验标签的中文错误信息翻译
func TestMsgForTag_Min(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	type TestStruct struct {
		Name string `validate:"min=3"`
	}

	testObj := TestStruct{Name: "ab"}
	validate := validator.New()
	err := validate.Struct(testObj)

	if err == nil {
		t.Fatal("expected validation error")
	}

	HandleValidationError(c, err)

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	errorsList, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected data to be []interface{}, got %T", resp.Data)
	}

	if len(errorsList) > 0 {
		firstErr, ok := errorsList[0].(map[string]interface{})
		if !ok {
			t.Fatalf("expected first error to be map, got %T", errorsList[0])
		}
		// 验证 min 标签翻译为"长度不能小于 {param}"
		expected := "长度不能小于 3"
		if firstErr["message"] != expected {
			t.Errorf("expected message '%s', got '%s'", expected, firstErr["message"])
		}
	}
}

// TestMsgForTag_Max 测试 max 校验标签的中文错误信息翻译
func TestMsgForTag_Max(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	type TestStruct struct {
		Name string `validate:"max=5"`
	}

	testObj := TestStruct{Name: "this is too long"}
	validate := validator.New()
	err := validate.Struct(testObj)

	if err == nil {
		t.Fatal("expected validation error")
	}

	HandleValidationError(c, err)

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	errorsList, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected data to be []interface{}, got %T", resp.Data)
	}

	if len(errorsList) > 0 {
		firstErr, ok := errorsList[0].(map[string]interface{})
		if !ok {
			t.Fatalf("expected first error to be map, got %T", errorsList[0])
		}
		// 验证 max 标签翻译为"长度不能大于 {param}"
		expected := "长度不能大于 5"
		if firstErr["message"] != expected {
			t.Errorf("expected message '%s', got '%s'", expected, firstErr["message"])
		}
	}
}

// TestMsgForTag_OneOf 测试 oneof 校验标签的中文错误信息翻译
func TestMsgForTag_OneOf(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	type TestStruct struct {
		Color string `validate:"oneof=red green blue"`
	}

	testObj := TestStruct{Color: "yellow"}
	validate := validator.New()
	err := validate.Struct(testObj)

	if err == nil {
		t.Fatal("expected validation error")
	}

	HandleValidationError(c, err)

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	errorsList, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected data to be []interface{}, got %T", resp.Data)
	}

	if len(errorsList) > 0 {
		firstErr, ok := errorsList[0].(map[string]interface{})
		if !ok {
			t.Fatalf("expected first error to be map, got %T", errorsList[0])
		}
		// 验证 oneof 标签翻译为"值必须是以下之一: {param}"
		expected := "值必须是以下之一: red green blue"
		if firstErr["message"] != expected {
			t.Errorf("expected message '%s', got '%s'", expected, firstErr["message"])
		}
	}
}

// TestPageResult_JSONMarshaling 测试分页响应结构的 JSON 序列化
func TestPageResult_JSONMarshaling(t *testing.T) {
	pageResult := PageResult{
		List:     []string{"item1", "item2"},
		Total:    100,
		Page:     1,
		PageSize: 10,
	}

	data, err := json.Marshal(pageResult)
	if err != nil {
		t.Fatalf("failed to marshal PageResult: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal PageResult: %v", err)
	}

	// 验证 JSON 序列化后的字段值
	if result["total"].(float64) != 100 {
		t.Errorf("expected total 100, got %v", result["total"])
	}

	if result["page"].(float64) != 1 {
		t.Errorf("expected page 1, got %v", result["page"])
	}

	if result["page_size"].(float64) != 10 {
		t.Errorf("expected pageSize 10, got %v", result["page_size"])
	}
}

// TestResponse_JSONMarshaling 测试响应结构的 JSON 序列化
func TestResponse_JSONMarshaling(t *testing.T) {
	resp := Response{
		Code: 0,
		Msg:  "成功",
		Data: map[string]string{"key": "value"},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal Response: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal Response: %v", err)
	}

	if result["code"].(float64) != 0 {
		t.Errorf("expected code 0, got %v", result["code"])
	}

	if result["msg"] != "成功" {
		t.Errorf("expected msg '成功', got %v", result["msg"])
	}
}

// TestResponse_JSONMarshaling_NilData 测试响应结构在 Data 为空时的 JSON 序列化
func TestResponse_JSONMarshaling_NilData(t *testing.T) {
	resp := Response{
		Code: 100001,
		Msg:  "错误",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal Response: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal Response: %v", err)
	}

	if result["code"].(float64) != 100001 {
		t.Errorf("expected code 100001, got %v", result["code"])
	}
}
