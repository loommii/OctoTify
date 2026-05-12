package server

import (
	"net/http"

	"github.com/ruiborda/go-swagger-generator/v2/src/middleware"
	"github.com/ruiborda/go-swagger-generator/v2/src/openapi"
	"github.com/ruiborda/go-swagger-generator/v2/src/openapi_spec/mime"
	"github.com/ruiborda/go-swagger-generator/v2/src/swagger"

	"octotify/internal/handler/dto"
	"octotify/pkg/response"
)

// setupOpenAPI 配置 OpenAPI 3.0 文档和 Swagger UI
func (s *Server) setupOpenAPI() {
	// 注册 Swagger UI 中间件
	s.engine.Use(middleware.SwaggerGin(middleware.SwaggerConfig{
		Enabled:  true,
		JSONPath: "/openapi.json",
		UIPath:   "/swagger/",
		Title:    "OctoTify API",
	}))

	doc := swagger.Swagger()

	// 基础信息
	doc.Info(func(info openapi.Info) {
		info.Title("OctoTify API").
			Version("1.0.0").
			Description("OctoTify 是一个消息总线平台，支持多种消息来源和推送渠道。\n\n" +
				"核心功能：消息来源管理、推送渠道管理、消息推送与记录。")
	})

	// 服务器地址
	doc.Server("http://localhost:34123", func(server openapi.Server) {
		server.Description("本地开发服务器")
	})

	// 安全方案定义
	doc.ComponentSecurityScheme("BearerAuth", func(s openapi.SecurityScheme) {
		s.Type("http").Scheme("bearer").BearerFormat("JWT").
			Description("输入格式：Bearer {access_token}，例如：Bearer eyJhbGciOiJSUzI1NiIs...")
	})

	doc.ComponentSecurityScheme("SourceTokenAuth", func(s openapi.SecurityScheme) {
		s.Type("http").Scheme("bearer").
			Description("推送消息时使用 Source Token，输入格式：Bearer src{uuid}，例如：Bearer src0196a3b2c4d50000a1b2c3d4e5f67890")
	})

	// 注册所有 DTO 为组件 Schema
	s.registerComponentSchemas(doc)

	// 注册所有路由
	s.registerPingRoute(doc)
	s.registerAuthRoutes(doc)
	s.registerUserRoutes(doc)
	s.registerSourceRoutes(doc)
	s.registerChannelRoutes(doc)
	s.registerMessageRoutes(doc)
	s.registerPushRoutes(doc)
}

// registerComponentSchemas 注册所有 DTO 为 OpenAPI 组件 Schema
func (s *Server) registerComponentSchemas(doc openapi.SwaggerDocBuilder) {
	// 认证相关
	_, _ = doc.SchemaFromDTO(&dto.AuthCredentials{})
	_, _ = doc.SchemaFromDTO(&dto.RegisterReq{})
	_, _ = doc.SchemaFromDTO(&dto.LoginReq{})
	_, _ = doc.SchemaFromDTO(&dto.UserDTO{})
	_, _ = doc.SchemaFromDTO(&dto.RefreshReq{})
	_, _ = doc.SchemaFromDTO(&dto.ChangePasswordReq{})
	_, _ = doc.SchemaFromDTO(&dto.AuthResp{})
	_, _ = doc.SchemaFromDTO(&dto.UserProfileResp{})

	// 来源相关
	_, _ = doc.SchemaFromDTO(&dto.SourceBaseReq{})
	_, _ = doc.SchemaFromDTO(&dto.CreateSourceReq{})
	_, _ = doc.SchemaFromDTO(&dto.UpdateSourceReq{})
	_, _ = doc.SchemaFromDTO(&dto.SourceDTO{})
	_, _ = doc.SchemaFromDTO(&dto.SourceDetailDTO{})
	_, _ = doc.SchemaFromDTO(&dto.SourceDetailResponse{})
	_, _ = doc.SchemaFromDTO(&dto.SourceTokenResponse{})
	_, _ = doc.SchemaFromDTO(&dto.VerifyPasswordReq{})

	// 渠道相关
	_, _ = doc.SchemaFromDTO(&dto.ConfigField{})
	_, _ = doc.SchemaFromDTO(&dto.ChannelTypeMeta{})
	_, _ = doc.SchemaFromDTO(&dto.CreateChannelReq{})
	_, _ = doc.SchemaFromDTO(&dto.UpdateChannelReq{})
	_, _ = doc.SchemaFromDTO(&dto.ChannelDTO{})
	_, _ = doc.SchemaFromDTO(&dto.StartBindResp{})
	_, _ = doc.SchemaFromDTO(&dto.GetBindStatusReq{})
	_, _ = doc.SchemaFromDTO(&dto.BindStatusResp{})
	_, _ = doc.SchemaFromDTO(&dto.BindCredentialsDTO{})

	// 消息相关
	_, _ = doc.SchemaFromDTO(&dto.PushMessageReq{})
	_, _ = doc.SchemaFromDTO(&dto.MessageDTO{})
	_, _ = doc.SchemaFromDTO(&dto.MessageDetailDTO{})
	_, _ = doc.SchemaFromDTO(&dto.MessageFilterReq{})
	_, _ = doc.SchemaFromDTO(&dto.PushResult{})
	_, _ = doc.SchemaFromDTO(&dto.PushResponse{})

	// 分页和响应
	_, _ = doc.SchemaFromDTO(&dto.PageReq{})
	_, _ = doc.SchemaFromDTO(&response.Response{})
	_, _ = doc.SchemaFromDTO(&response.PageResult{})
	_, _ = doc.SchemaFromDTO(&response.FieldError{})
}

// registerPingRoute 注册健康检查路由
func (s *Server) registerPingRoute(doc openapi.SwaggerDocBuilder) {
	var _ = doc.Path("/ping").
		Get(func(op openapi.Operation) {
			op.Summary("健康检查").
				Description("系统健康检查接口，返回服务器名称和运行状态。").
				Tag("系统").
				OperationID("Ping").
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()
}

// registerAuthRoutes 注册用户认证路由
func (s *Server) registerAuthRoutes(doc openapi.SwaggerDocBuilder) {
	// POST /api/auth/login
	var _ = doc.Path("/api/auth/login").
		Post(func(op openapi.Operation) {
			op.Summary("用户登录").
				Description("使用用户名和密码登录系统，登录成功后返回 Access Token 和 Refresh Token。\n\n"+
					"## 密码要求\n"+
					"- 8-128 个字符\n"+
					"- 必须包含小写字母\n"+
					"- 必须包含大写字母\n"+
					"- 必须包含数字").
				Tag("用户认证").
				OperationID("Login").
				RequestBody(func(rb openapi.RequestBody) {
					rb.Description("登录请求参数").
						Required(true).
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&dto.LoginReq{})
						})
				}).
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("登录成功，返回认证信息").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()

	// POST /api/auth/refresh
	var _ = doc.Path("/api/auth/refresh").
		Post(func(op openapi.Operation) {
			op.Summary("刷新令牌").
				Description("使用 Refresh Token 获取新的 Access Token 和 Refresh Token。\n\n"+
					"## 使用场景\n"+
					"1. Access Token 过期（1 小时后）时调用此接口\n"+
					"2. 前端应在请求收到 401 响应时自动调用此接口").
				Tag("用户认证").
				OperationID("RefreshToken").
				RequestBody(func(rb openapi.RequestBody) {
					rb.Description("刷新令牌请求参数").
						Required(true).
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&dto.RefreshReq{})
						})
				}).
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("刷新成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()

	// POST /api/auth/logout
	var _ = doc.Path("/api/auth/logout").
		Post(func(op openapi.Operation) {
			op.Summary("退出登录").
				Description("撤销当前用户的所有 Refresh Token，退出登录后需要重新登录才能访问系统。").
				Tag("用户认证").
				OperationID("Logout").
				Security("BearerAuth").
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("退出成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				}).
				Response(http.StatusUnauthorized, func(r openapi.Response) {
					r.Description("未提供认证令牌或认证令牌无效或已过期")
				})
		}).
		Doc()
}

// registerUserRoutes 注册用户管理路由
func (s *Server) registerUserRoutes(doc openapi.SwaggerDocBuilder) {
	// POST /api/user/register
	var _ = doc.Path("/api/user/register").
		Post(func(op openapi.Operation) {
			op.Summary("用户注册").
				Description("注册新用户账号，注册成功后自动登录并返回 Access Token 和 Refresh Token。").
				Tag("用户管理").
				OperationID("Register").
				RequestBody(func(rb openapi.RequestBody) {
					rb.Description("注册请求参数").
						Required(true).
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&dto.RegisterReq{})
						})
				}).
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("注册成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()

	// PUT /api/user/password
	var _ = doc.Path("/api/user/password").
		Put(func(op openapi.Operation) {
			op.Summary("修改密码").
				Description("修改当前用户的登录密码，修改成功后所有 Refresh Token 将被撤销，需要重新登录。").
				Tag("用户管理").
				OperationID("ChangePassword").
				Security("BearerAuth").
				RequestBody(func(rb openapi.RequestBody) {
					rb.Description("修改密码请求参数").
						Required(true).
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&dto.ChangePasswordReq{})
						})
				}).
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("修改成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()

	// GET /api/user/profile
	var _ = doc.Path("/api/user/profile").
		Get(func(op openapi.Operation) {
			op.Summary("查询用户信息").
				Description("获取当前登录用户的个人信息，包括用户 ID、用户名和创建时间。").
				Tag("用户管理").
				OperationID("GetUserProfile").
				Security("BearerAuth").
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("查询成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()
}

// registerSourceRoutes 注册消息来源管理路由
func (s *Server) registerSourceRoutes(doc openapi.SwaggerDocBuilder) {
	// POST /api/sources
	var _ = doc.Path("/api/sources").
		Post(func(op openapi.Operation) {
			op.Summary("创建消息来源").
				Description("创建一个新的消息来源，系统自动生成 Source Token，用于外部系统推送消息。").
				Tag("消息来源管理").
				OperationID("CreateSource").
				Security("BearerAuth").
				RequestBody(func(rb openapi.RequestBody) {
					rb.Description("创建消息来源请求参数").
						Required(true).
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&dto.CreateSourceReq{})
						})
				}).
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("创建成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()

	// GET /api/sources
	var _ = doc.Path("/api/sources").
		Get(func(op openapi.Operation) {
			op.Summary("查看消息来源列表").
				Description("分页查询当前用户的所有消息来源列表，返回来源基本信息（不包含 Token）。").
				Tag("消息来源管理").
				OperationID("ListSources").
				Security("BearerAuth").
				QueryParameter("page", func(p openapi.Parameter) {
					p.Description("页码，从 1 开始，默认 1").
						Schema(func(s openapi.Schema) {
							s.Type("integer").Default(1).Minimum(1, false)
						})
				}).
				QueryParameter("page_size", func(p openapi.Parameter) {
					p.Description("每页条数，默认 20，最大 100").
						Schema(func(s openapi.Schema) {
							s.Type("integer").Default(20).Minimum(1, false).Maximum(100, false)
						})
				}).
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("查询成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()

	// PUT /api/sources/{id}
	var _ = doc.Path("/api/sources/{id}").
		Put(func(op openapi.Operation) {
			op.Summary("编辑消息来源").
				Description("编辑已有消息来源的名称和描述，不影响已绑定的渠道和推送功能。").
				Tag("消息来源管理").
				OperationID("UpdateSource").
				Security("BearerAuth").
				PathParameter("id", func(p openapi.Parameter) {
					p.Description("来源ID").
						Required(true).
						Schema(func(s openapi.Schema) {
							s.Type("integer").Format("int64").Minimum(1, false)
						})
				}).
				RequestBody(func(rb openapi.RequestBody) {
					rb.Description("编辑消息来源请求参数").
						Required(true).
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&dto.UpdateSourceReq{})
						})
				}).
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("编辑成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()

	// GET /api/sources/{id}
	var _ = doc.Path("/api/sources/{id}").
		Get(func(op openapi.Operation) {
			op.Summary("查看来源详情").
				Description("查询指定消息来源的详细信息，包含来源 Token、使用时间和已绑定的有效渠道列表。").
				Tag("消息来源管理").
				OperationID("GetSourceDetail").
				Security("BearerAuth").
				PathParameter("id", func(p openapi.Parameter) {
					p.Description("来源ID").
						Required(true).
						Schema(func(s openapi.Schema) {
							s.Type("integer").Format("int64").Minimum(1, false)
						})
				}).
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("查询成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()

	// POST /api/sources/{id}/token
	var _ = doc.Path("/api/sources/{id}/token").
		Post(func(op openapi.Operation) {
			op.Summary("查看来源令牌").
				Description("查询指定消息来源的 Token，需要密码二次验证以确保安全性。").
				Tag("消息来源管理").
				OperationID("GetSourceToken").
				Security("BearerAuth").
				PathParameter("id", func(p openapi.Parameter) {
					p.Description("来源ID").
						Required(true).
						Schema(func(s openapi.Schema) {
							s.Type("integer").Format("int64").Minimum(1, false)
						})
				}).
				RequestBody(func(rb openapi.RequestBody) {
					rb.Description("密码验证请求参数").
						Required(true).
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&dto.VerifyPasswordReq{})
						})
				}).
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("查看成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()

	// PUT /api/sources/{id}/token
	var _ = doc.Path("/api/sources/{id}/token").
		Put(func(op openapi.Operation) {
			op.Summary("重置来源令牌").
				Description("重置指定消息来源的 Token，需要密码二次验证，旧 Token 立即失效。").
				Tag("消息来源管理").
				OperationID("ResetSourceToken").
				Security("BearerAuth").
				PathParameter("id", func(p openapi.Parameter) {
					p.Description("来源ID").
						Required(true).
						Schema(func(s openapi.Schema) {
							s.Type("integer").Format("int64").Minimum(1, false)
						})
				}).
				RequestBody(func(rb openapi.RequestBody) {
					rb.Description("密码验证请求参数").
						Required(true).
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&dto.VerifyPasswordReq{})
						})
				}).
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("重置成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()

	// PUT /api/sources/{id}/disable
	var _ = doc.Path("/api/sources/{id}/disable").
		Put(func(op openapi.Operation) {
			op.Summary("停用消息来源").
				Description("停用指定消息来源，需要密码二次验证，停用后该来源无法推送消息。").
				Tag("消息来源管理").
				OperationID("DisableSource").
				Security("BearerAuth").
				PathParameter("id", func(p openapi.Parameter) {
					p.Description("来源ID").
						Required(true).
						Schema(func(s openapi.Schema) {
							s.Type("integer").Format("int64").Minimum(1, false)
						})
				}).
				RequestBody(func(rb openapi.RequestBody) {
					rb.Description("密码验证请求参数").
						Required(true).
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&dto.VerifyPasswordReq{})
						})
				}).
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("停用成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()

	// PUT /api/sources/{id}/enable
	var _ = doc.Path("/api/sources/{id}/enable").
		Put(func(op openapi.Operation) {
			op.Summary("启用消息来源").
				Description("启用指定消息来源，需要密码二次验证，恢复消息推送功能。").
				Tag("消息来源管理").
				OperationID("EnableSource").
				Security("BearerAuth").
				PathParameter("id", func(p openapi.Parameter) {
					p.Description("来源ID").
						Required(true).
						Schema(func(s openapi.Schema) {
							s.Type("integer").Format("int64").Minimum(1, false)
						})
				}).
				RequestBody(func(rb openapi.RequestBody) {
					rb.Description("密码验证请求参数").
						Required(true).
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&dto.VerifyPasswordReq{})
						})
				}).
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("启用成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()

	// DELETE /api/sources/{id}
	var _ = doc.Path("/api/sources/{id}").
		Delete(func(op openapi.Operation) {
			op.Summary("删除消息来源").
				Description("软删除指定消息来源及其关联渠道关系，需要密码二次验证。删除后来源和所有关联数据不可恢复。").
				Tag("消息来源管理").
				OperationID("DeleteSource").
				Security("BearerAuth").
				PathParameter("id", func(p openapi.Parameter) {
					p.Description("来源ID").
						Required(true).
						Schema(func(s openapi.Schema) {
							s.Type("integer").Format("int64").Minimum(1, false)
						})
				}).
				RequestBody(func(rb openapi.RequestBody) {
					rb.Description("密码验证请求参数").
						Required(true).
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&dto.VerifyPasswordReq{})
						})
				}).
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("删除成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()
}

// registerChannelRoutes 注册推送渠道管理路由
func (s *Server) registerChannelRoutes(doc openapi.SwaggerDocBuilder) {
	// GET /api/channel-types
	var _ = doc.Path("/api/channel-types").
		Get(func(op openapi.Operation) {
			op.Summary("获取渠道类型元数据").
				Description("获取系统支持的所有推送渠道类型及其配置字段定义，用于前端动态渲染创建渠道表单。").
				Tag("推送渠道管理").
				OperationID("GetChannelTypes").
				Security("BearerAuth").
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("获取成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()

	// POST /api/channels
	var _ = doc.Path("/api/channels").
		Post(func(op openapi.Operation) {
			op.Summary("创建推送渠道").
				Description("创建一个新的推送渠道，配置渠道类型、名称和凭证信息。\n\n"+
					"## 支持的渠道类型\n"+
					"- **wechat**: 企业微信群机器人\n"+
					"- **telegram**: Telegram Bot\n"+
					"- **dingtalk**: 钉钉群机器人\n"+
					"- **email**: 邮件推送\n"+
					"- **webhook**: 自定义 Webhook\n"+
					"- **feishu**: 飞书自定义机器人").
				Tag("推送渠道管理").
				OperationID("CreateChannel").
				Security("BearerAuth").
				RequestBody(func(rb openapi.RequestBody) {
					rb.Description("创建推送渠道请求参数").
						Required(true).
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&dto.CreateChannelReq{})
						})
				}).
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("创建成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()

	// GET /api/channels
	var _ = doc.Path("/api/channels").
		Get(func(op openapi.Operation) {
			op.Summary("查看推送渠道列表").
				Description("分页查询当前用户的所有推送渠道列表，返回渠道基本信息。").
				Tag("推送渠道管理").
				OperationID("ListChannels").
				Security("BearerAuth").
				QueryParameter("page", func(p openapi.Parameter) {
					p.Description("页码，从 1 开始，默认 1").
						Schema(func(s openapi.Schema) {
							s.Type("integer").Default(1).Minimum(1, false)
						})
				}).
				QueryParameter("page_size", func(p openapi.Parameter) {
					p.Description("每页条数，默认 20，最大 100").
						Schema(func(s openapi.Schema) {
							s.Type("integer").Default(20).Minimum(1, false).Maximum(100, false)
						})
				}).
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("查询成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()

	// PUT /api/channels/{id}
	var _ = doc.Path("/api/channels/{id}").
		Put(func(op openapi.Operation) {
			op.Summary("编辑推送渠道").
				Description("编辑推送渠道的名称和配置信息，不影响已推送的消息。").
				Tag("推送渠道管理").
				OperationID("UpdateChannel").
				Security("BearerAuth").
				PathParameter("id", func(p openapi.Parameter) {
					p.Description("渠道ID").
						Required(true).
						Schema(func(s openapi.Schema) {
							s.Type("integer").Format("int64").Minimum(1, false)
						})
				}).
				RequestBody(func(rb openapi.RequestBody) {
					rb.Description("编辑推送渠道请求参数").
						Required(true).
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&dto.UpdateChannelReq{})
						})
				}).
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("编辑成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()

	// GET /api/channels/{id}
	var _ = doc.Path("/api/channels/{id}").
		Get(func(op openapi.Operation) {
			op.Summary("查看渠道详情").
				Description("查询指定推送渠道的详细信息，包括渠道类型、名称、配置和状态。").
				Tag("推送渠道管理").
				OperationID("GetChannelDetail").
				Security("BearerAuth").
				PathParameter("id", func(p openapi.Parameter) {
					p.Description("渠道ID").
						Required(true).
						Schema(func(s openapi.Schema) {
							s.Type("integer").Format("int64").Minimum(1, false)
						})
				}).
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("查询成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()

	// POST /api/channels/{id}/test
	var _ = doc.Path("/api/channels/{id}/test").
		Post(func(op openapi.Operation) {
			op.Summary("测试渠道连接").
				Description("发送测试消息到指定渠道，验证渠道配置是否正确。").
				Tag("推送渠道管理").
				OperationID("TestChannel").
				Security("BearerAuth").
				PathParameter("id", func(p openapi.Parameter) {
					p.Description("渠道ID").
						Required(true).
						Schema(func(s openapi.Schema) {
							s.Type("integer").Format("int64").Minimum(1, false)
						})
				}).
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("测试成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()

	// PUT /api/channels/{id}/disable
	var _ = doc.Path("/api/channels/{id}/disable").
		Put(func(op openapi.Operation) {
			op.Summary("停用推送渠道").
				Description("停用指定推送渠道，停用后该渠道不再接收消息推送。").
				Tag("推送渠道管理").
				OperationID("DisableChannel").
				Security("BearerAuth").
				PathParameter("id", func(p openapi.Parameter) {
					p.Description("渠道ID").
						Required(true).
						Schema(func(s openapi.Schema) {
							s.Type("integer").Format("int64").Minimum(1, false)
						})
				}).
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("停用成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()

	// PUT /api/channels/{id}/enable
	var _ = doc.Path("/api/channels/{id}/enable").
		Put(func(op openapi.Operation) {
			op.Summary("启用推送渠道").
				Description("启用指定推送渠道，恢复消息推送功能。").
				Tag("推送渠道管理").
				OperationID("EnableChannel").
				Security("BearerAuth").
				PathParameter("id", func(p openapi.Parameter) {
					p.Description("渠道ID").
						Required(true).
						Schema(func(s openapi.Schema) {
							s.Type("integer").Format("int64").Minimum(1, false)
						})
				}).
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("启用成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()

	// DELETE /api/channels/{id}
	var _ = doc.Path("/api/channels/{id}").
		Delete(func(op openapi.Operation) {
			op.Summary("删除推送渠道").
				Description("软删除指定推送渠道，同时解除与所有来源的关联关系。删除后渠道和所有关联数据不可恢复。").
				Tag("推送渠道管理").
				OperationID("DeleteChannel").
				Security("BearerAuth").
				PathParameter("id", func(p openapi.Parameter) {
					p.Description("渠道ID").
						Required(true).
						Schema(func(s openapi.Schema) {
							s.Type("integer").Format("int64").Minimum(1, false)
						})
				}).
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("删除成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()

	// POST /api/channels/wechat-clawbot/bind
	var _ = doc.Path("/api/channels/wechat-clawbot/bind").
		Post(func(op openapi.Operation) {
			op.Summary("发起微信ClawBot扫码绑定").
				Description("获取微信ClawBot绑定二维码，用户扫码后完成绑定。").
				Tag("微信ClawBot绑定").
				OperationID("StartBind").
				Security("BearerAuth").
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("获取成功，返回 qrcode_url 和 qrcode").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()

	// POST /api/channels/wechat-clawbot/bind/status
	var _ = doc.Path("/api/channels/wechat-clawbot/bind/status").
		Post(func(op openapi.Operation) {
			op.Summary("查询微信ClawBot绑定状态").
				Description("通过 qrcode 调用 iLink API 查询绑定状态（长轮询，40s 超时）。\n\n"+
					"## 返回状态\n"+
					"- pending: 仍在等待中（超时返回）\n"+
					"- scanned: 用户已扫码，待确认\n"+
					"- confirmed: 绑定成功，返回凭证\n"+
					"- expired: 二维码已过期").
				Tag("微信ClawBot绑定").
				OperationID("GetBindStatus").
				Security("BearerAuth").
				RequestBody(func(rb openapi.RequestBody) {
					rb.Description("查询绑定状态请求参数").
						Required(true).
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&dto.GetBindStatusReq{})
						})
				}).
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("查询成功，返回绑定状态").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()
}

// registerMessageRoutes 注册消息管理路由
func (s *Server) registerMessageRoutes(doc openapi.SwaggerDocBuilder) {
	// GET /api/messages
	var _ = doc.Path("/api/messages").
		Get(func(op openapi.Operation) {
			op.Summary("查看消息列表").
				Description("查看当前用户的消息记录列表，按创建时间倒序排列。").
				Tag("消息管理").
				OperationID("ListMessages").
				Security("BearerAuth").
				QueryParameter("page", func(p openapi.Parameter) {
					p.Description("页码，从 1 开始，默认 1").
						Schema(func(s openapi.Schema) {
							s.Type("integer").Default(1).Minimum(1, false)
						})
				}).
				QueryParameter("page_size", func(p openapi.Parameter) {
					p.Description("每页条数，默认 20，最大 100").
						Schema(func(s openapi.Schema) {
							s.Type("integer").Default(20).Minimum(1, false).Maximum(100, false)
						})
				}).
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("查询成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()

	// GET /api/messages/filter
	var _ = doc.Path("/api/messages/filter").
		Get(func(op openapi.Operation) {
			op.Summary("筛选消息").
				Description("根据来源、渠道、状态、时间范围等条件筛选消息，支持多条件组合查询。").
				Tag("消息管理").
				OperationID("FilterMessages").
				Security("BearerAuth").
				QueryParameter("source_id", func(p openapi.Parameter) {
					p.Description("来源 ID（可选）").
						Schema(func(s openapi.Schema) {
							s.Type("integer").Format("int64")
						})
				}).
				QueryParameter("channel_id", func(p openapi.Parameter) {
					p.Description("渠道 ID（可选）").
						Schema(func(s openapi.Schema) {
							s.Type("integer").Format("int64")
						})
				}).
				QueryParameter("status", func(p openapi.Parameter) {
					p.Description("推送状态（可选）：100-待推送 200-成功 300-失败").
						Schema(func(s openapi.Schema) {
							s.Type("integer")
						})
				}).
				QueryParameter("start_date", func(p openapi.Parameter) {
					p.Description("开始时间（可选，Unix 毫秒时间戳）").
						Schema(func(s openapi.Schema) {
							s.Type("integer").Format("int64")
						})
				}).
				QueryParameter("end_date", func(p openapi.Parameter) {
					p.Description("结束时间（可选，Unix 毫秒时间戳）").
						Schema(func(s openapi.Schema) {
							s.Type("integer").Format("int64")
						})
				}).
				QueryParameter("keyword", func(p openapi.Parameter) {
					p.Description("关键词（可选，搜索标题和内容）").
						Schema(func(s openapi.Schema) {
							s.Type("string")
						})
				}).
				QueryParameter("page", func(p openapi.Parameter) {
					p.Description("页码，从 1 开始，默认 1").
						Schema(func(s openapi.Schema) {
							s.Type("integer").Default(1).Minimum(1, false)
						})
				}).
				QueryParameter("page_size", func(p openapi.Parameter) {
					p.Description("每页条数，默认 20，最大 100").
						Schema(func(s openapi.Schema) {
							s.Type("integer").Default(20).Minimum(1, false).Maximum(100, false)
						})
				}).
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("查询成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()

	// GET /api/messages/{id}
	var _ = doc.Path("/api/messages/{id}").
		Get(func(op openapi.Operation) {
			op.Summary("查看消息详情").
				Description("查看单条消息的详细信息，包含消息内容、来源名称、渠道名称和推送状态。").
				Tag("消息管理").
				OperationID("GetMessageDetail").
				Security("BearerAuth").
				PathParameter("id", func(p openapi.Parameter) {
					p.Description("消息ID").
						Required(true).
						Schema(func(s openapi.Schema) {
							s.Type("integer").Format("int64").Minimum(1, false)
						})
				}).
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("查询成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()

	// DELETE /api/messages/{id}
	var _ = doc.Path("/api/messages/{id}").
		Delete(func(op openapi.Operation) {
			op.Summary("删除消息").
				Description("删除单条消息记录（软删除），删除后消息不再显示在列表中。").
				Tag("消息管理").
				OperationID("DeleteMessage").
				Security("BearerAuth").
				PathParameter("id", func(p openapi.Parameter) {
					p.Description("消息ID").
						Required(true).
						Schema(func(s openapi.Schema) {
							s.Type("integer").Format("int64").Minimum(1, false)
						})
				}).
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("删除成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()
}

// registerPushRoutes 注册消息推送路由
func (s *Server) registerPushRoutes(doc openapi.SwaggerDocBuilder) {
	// POST /api/push/{token}
	var _ = doc.Path("/api/push/{token}").
		Post(func(op openapi.Operation) {
			op.Summary("推送消息").
				Description("外部系统通过 Source Token 推送消息到平台，消息会被并发推送到所有绑定的有效渠道。\n\n"+
					"## 认证方式\n"+
					"使用 Source Token 进行认证，而非用户 Access Token。\n"+
					"Token 格式：src{uuid}，例如：src0196a3b2c4d50000a1b2c3d4e5f67890").
				Tag("消息推送").
				OperationID("PushMessage").
				Security("SourceTokenAuth").
				PathParameter("token", func(p openapi.Parameter) {
					p.Description("来源 Token，格式为 src{uuid}").
						Required(true).
						Schema(func(s openapi.Schema) {
							s.Type("string")
						})
				}).
				RequestBody(func(rb openapi.RequestBody) {
					rb.Description("推送消息请求参数").
						Required(true).
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&dto.PushMessageReq{})
						})
				}).
				Response(http.StatusOK, func(r openapi.Response) {
					r.Description("推送成功").
						Content(mime.ApplicationJSON, func(mt openapi.MediaType) {
							mt.SchemaFromDTO(&response.Response{})
						})
				})
		}).
		Doc()
}
