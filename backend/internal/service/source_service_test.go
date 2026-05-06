package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"octotify/internal/handler/dto"
	"octotify/internal/model"
	"octotify/pkg/xerr"
)

// ============================================================================
// TestSourceService_CreateSource 创建消息来源功能测试
// ============================================================================

// TestSourceService_CreateSource 测试创建消息来源功能
// 覆盖场景：正常创建（带渠道绑定）、正常创建（不带渠道绑定）、数据库插入错误
func TestSourceService_CreateSource(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	ctx := context.Background()

	// 预创建测试用户和渠道
	testUser := CreateTestUser(t, db, "create_source_user", "Password1")
	channel1 := CreateTestChannel(t, db, testUser.ID, "webhook", "Test Channel 1")
	channel2 := CreateTestChannel(t, db, testUser.ID, "dingtalk", "Test Channel 2")

	tests := []struct {
		name        string            // 测试用例名称
		userID      int64             // 用户 ID
		req         *dto.CreateSourceReq // 创建请求
		setup       func()            // 测试前置操作（可选）
		wantErrCode int               // 期望的错误码（0 表示无错误）
		wantSuccess bool              // 是否期望成功
	}{
		{
			name:   "成功：创建消息来源并绑定多个渠道",
			userID: testUser.ID,
			req: &dto.CreateSourceReq{
				SourceBaseReq: dto.SourceBaseReq{
					Name:        "Test Source",
					Description: "Test source description",
				},
				ChannelIDs: []int64{channel1.ID, channel2.ID},
			},
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:   "成功：创建消息来源但不绑定渠道",
			userID: testUser.ID,
			req: &dto.CreateSourceReq{
				SourceBaseReq: dto.SourceBaseReq{
					Name:        "Source Without Channels",
					Description: "Source with no channel bindings",
				},
				ChannelIDs: []int64{},
			},
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:   "失败：数据库连接已关闭，应返回 ErrSourceTokenFailed",
			userID: testUser.ID,
			req: &dto.CreateSourceReq{
				SourceBaseReq: dto.SourceBaseReq{
					Name:        "DB Error Source",
					Description: "Should fail",
				},
				ChannelIDs: []int64{channel1.ID},
			},
			setup: func() {
				sqlDB, err := db.DB()
				require.NoError(t, err, "获取底层 sql.DB 失败")
				err = sqlDB.Close()
				require.NoError(t, err, "关闭数据库连接失败")
			},
			wantErrCode: xerr.ErrSourceTokenFailed.Code,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			svc := NewSourceService(db, logger)
			result, err := svc.CreateSource(ctx, tt.userID, tt.req)

			if tt.wantSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.req.Name, result.Name)
				assert.Equal(t, tt.req.Description, result.Description)
				assert.Equal(t, tt.userID, result.UserID)
				assert.NotEmpty(t, result.Token, "Source Token 不应为空")
				assert.True(t, len(result.Token) == 35, "Token 长度应为 35 位（src + 32 位 UUID）")
				assert.Equal(t, model.SourceStatusActive, result.Status)
				assert.NotZero(t, result.CreatedAt, "CreatedAt 不应为零")

				// 验证数据库中确实创建了来源记录
				var source model.Source
				findErr := db.Where("name = ?", tt.req.Name).First(&source).Error
				assert.NoError(t, findErr, "数据库中应存在该来源")
				assert.Equal(t, tt.req.Name, source.Name)
				assert.Equal(t, tt.userID, source.UserID)
				assert.Equal(t, model.SourceStatusActive, source.Status)

				// 如果请求中包含渠道绑定，验证关联记录已创建
				if len(tt.req.ChannelIDs) > 0 {
					var bindingCount int64
					db.Model(&model.SourceChannel{}).Where("source_id = ?", source.ID).Count(&bindingCount)
					assert.Equal(t, int64(len(tt.req.ChannelIDs)), bindingCount, "应创建对应数量的渠道绑定记录")
				}
			} else {
				assert.Error(t, err)
				assert.Nil(t, result)
				appErr, ok := err.(*xerr.AppError)
				assert.True(t, ok, "错误类型应为 *xerr.AppError")
				assert.Equal(t, tt.wantErrCode, appErr.Code, "错误码不匹配")
			}
		})
	}
}

// ============================================================================
// TestSourceService_ListSources 列表查询功能测试
// ============================================================================

// TestSourceService_ListSources 测试分页查询消息来源列表功能
// 覆盖场景：正常分页查询、排除已删除来源
func TestSourceService_ListSources(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	ctx := context.Background()

	// 预创建测试用户和多个来源
	testUser := CreateTestUser(t, db, "list_sources_user", "Password1")
	CreateTestSource(t, db, testUser.ID, "Active Source 1", "src-token-1")
	CreateTestSource(t, db, testUser.ID, "Active Source 2", "src-token-2")
	CreateTestSource(t, db, testUser.ID, "Deleted Source", "src-token-3")
	// 将第三个来源标记为已删除
	db.Model(&model.Source{}).Where("name = ?", "Deleted Source").Update("status", model.SourceStatusDeleted)

	tests := []struct {
		name        string         // 测试用例名称
		userID      int64          // 用户 ID
		pageReq     *dto.PageReq   // 分页请求
		wantTotal   int64          // 期望的总数
		wantCount   int            // 期望的返回数量
		wantErrCode int            // 期望的错误码（0 表示无错误）
		wantSuccess bool           // 是否期望成功
	}{
		{
			name:      "成功：分页查询活跃来源列表（应排除已删除来源）",
			userID:    testUser.ID,
			pageReq:   &dto.PageReq{Page: 1, PageSize: 10},
			wantTotal: 2,  // 只应返回 2 个活跃来源
			wantCount: 2,
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:      "成功：分页查询第一页，限制每页 1 条",
			userID:    testUser.ID,
			pageReq:   &dto.PageReq{Page: 1, PageSize: 1},
			wantTotal: 2,  // 总数仍为 2
			wantCount: 1,  // 但只返回 1 条
			wantErrCode: 0,
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewSourceService(db, logger)
			list, total, err := svc.ListSources(ctx, tt.userID, tt.pageReq)

			if tt.wantSuccess {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTotal, total, "总数应与预期一致")
				assert.Len(t, list, tt.wantCount, "返回数量应与预期一致")

				// 验证返回列表中不包含已删除的来源
				for _, item := range list {
					assert.NotEqual(t, model.SourceStatusDeleted, item.Status, "列表中不应包含已删除的来源")
					assert.NotEqual(t, "Deleted Source", item.Name, "列表中不应包含名为 'Deleted Source' 的来源")
				}

				// 如果只查询 1 条，验证按创建时间降序排列（最新优先）
				if tt.pageReq.PageSize == 1 {
					assert.Equal(t, "Active Source 2", list[0].Name, "应返回最新创建的来源")
				}
			} else {
				assert.Error(t, err)
				assert.Nil(t, list)
				assert.Equal(t, int64(0), total)
				appErr, ok := err.(*xerr.AppError)
				assert.True(t, ok, "错误类型应为 *xerr.AppError")
				assert.Equal(t, tt.wantErrCode, appErr.Code, "错误码不匹配")
			}
		})
	}
}

// ============================================================================
// TestSourceService_UpdateSource 更新消息来源功能测试
// ============================================================================

// TestSourceService_UpdateSource 测试更新消息来源功能
// 覆盖场景：成功更新、来源不存在、来源已删除、权限不足
func TestSourceService_UpdateSource(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	ctx := context.Background()

	// 预创建测试用户和来源
	testUser := CreateTestUser(t, db, "update_source_user", "Password1")
	otherUser := CreateTestUser(t, db, "other_user", "Password1")
	testSource := CreateTestSource(t, db, testUser.ID, "Original Source", "src-update-token")
	otherUserSource := CreateTestSource(t, db, otherUser.ID, "Other User Source", "src-other-user-token")
	
	// 创建已删除的来源
	deletedSource := CreateTestSource(t, db, testUser.ID, "Deleted Source", "src-deleted-token")
	db.Model(&model.Source{}).Where("id = ?", deletedSource.ID).Update("status", model.SourceStatusDeleted)

	tests := []struct {
		name        string            // 测试用例名称
		sourceID    int64             // 来源 ID
		userID      int64             // 操作用户 ID
		req         *dto.UpdateSourceReq // 更新请求
		setup       func()            // 测试前置操作（可选）
		wantErrCode int               // 期望的错误码（0 表示无错误）
		wantSuccess bool              // 是否期望成功
	}{
		{
			name:     "成功：更新来源名称和描述",
			sourceID: testSource.ID,
			userID:   testUser.ID,
			req: &dto.UpdateSourceReq{
				SourceBaseReq: dto.SourceBaseReq{
					Name:        "Updated Source Name",
					Description: "Updated description",
				},
				ChannelIDs: nil, // 不更新渠道绑定
			},
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:     "失败：来源不存在，应返回 ErrSourceNotFound",
			sourceID: 99999,
			userID:   testUser.ID,
			req: &dto.UpdateSourceReq{
				SourceBaseReq: dto.SourceBaseReq{
					Name:        "Non-existent Source",
					Description: "Should fail",
				},
			},
			wantErrCode: xerr.ErrSourceNotFound.Code,
			wantSuccess: false,
		},
		{
			name:     "失败：来源已删除，应返回 ErrSourceAlreadyDeleted",
			sourceID: deletedSource.ID,
			userID:   testUser.ID,
			req: &dto.UpdateSourceReq{
				SourceBaseReq: dto.SourceBaseReq{
					Name:        "Updated Deleted Source",
					Description: "Should fail",
				},
			},
			wantErrCode: xerr.ErrSourceAlreadyDeleted.Code,
			wantSuccess: false,
		},
		{
			name:     "失败：其他用户尝试更新，应返回 ErrSourceNotFound（权限隔离）",
			sourceID: otherUserSource.ID,
			userID:   testUser.ID,
			req: &dto.UpdateSourceReq{
				SourceBaseReq: dto.SourceBaseReq{
					Name:        "Hacked Source Name",
					Description: "Should fail",
				},
			},
			wantErrCode: xerr.ErrSourceNotFound.Code,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			svc := NewSourceService(db, logger)
			err := svc.UpdateSource(ctx, tt.sourceID, tt.userID, tt.req)

			if tt.wantSuccess {
				assert.NoError(t, err)

				// 验证数据库中的来源已更新
				var updatedSource model.Source
				findErr := db.Where("id = ?", tt.sourceID).First(&updatedSource).Error
				assert.NoError(t, findErr, "应能查询到来源")
				assert.Equal(t, "Updated Source Name", updatedSource.Name, "名称应已更新")
				assert.Equal(t, "Updated description", updatedSource.Description, "描述应已更新")
			} else {
				assert.Error(t, err)
				appErr, ok := err.(*xerr.AppError)
				assert.True(t, ok, "错误类型应为 *xerr.AppError")
				assert.Equal(t, tt.wantErrCode, appErr.Code, "错误码不匹配")

				// 对于失败场景，验证来源数据未被修改
				if tt.sourceID != 99999 {
					var originalSource model.Source
					findErr := db.Where("id = ?", tt.sourceID).First(&originalSource).Error
					assert.NoError(t, findErr, "应能查询到来源")
					// 确保名称没有被修改
					if tt.sourceID == testSource.ID {
						assert.Equal(t, "Original Source", originalSource.Name, "名称不应被修改")
					}
				}
			}
		})
	}
}

// ============================================================================
// TestSourceService_GetSourceDetail 查询来源详情功能测试
// ============================================================================

// TestSourceService_GetSourceDetail 测试查询消息来源详情功能
// 覆盖场景：成功获取详情（含绑定渠道）、来源不存在、来源已删除
func TestSourceService_GetSourceDetail(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	ctx := context.Background()

	// 预创建测试用户、来源和渠道
	testUser := CreateTestUser(t, db, "detail_source_user", "Password1")
	testSource := CreateTestSource(t, db, testUser.ID, "Detail Source", "src-detail-token")
	channel1 := CreateTestChannel(t, db, testUser.ID, "webhook", "Detail Channel 1")
	channel2 := CreateTestChannel(t, db, testUser.ID, "dingtalk", "Detail Channel 2")
	
	// 绑定渠道
	BindSourceToChannel(t, db, testSource.ID, channel1.ID)
	BindSourceToChannel(t, db, testSource.ID, channel2.ID)

	// 创建已删除的来源
	deletedSource := CreateTestSource(t, db, testUser.ID, "Deleted Detail Source", "src-deleted-detail-token")
	db.Model(&model.Source{}).Where("id = ?", deletedSource.ID).Update("status", model.SourceStatusDeleted)

	tests := []struct {
		name        string            // 测试用例名称
		sourceID    int64             // 来源 ID
		userID      int64             // 用户 ID
		wantErrCode int               // 期望的错误码（0 表示无错误）
		wantSuccess bool              // 是否期望成功
		wantChannelCount int          // 期望的绑定渠道数量
	}{
		{
			name:             "成功：获取来源详情，包含绑定的渠道列表",
			sourceID:         testSource.ID,
			userID:           testUser.ID,
			wantErrCode:      0,
			wantSuccess:      true,
			wantChannelCount: 2,
		},
		{
			name:             "失败：来源不存在，应返回 ErrSourceNotFound",
			sourceID:         99999,
			userID:           testUser.ID,
			wantErrCode:      xerr.ErrSourceNotFound.Code,
			wantSuccess:      false,
			wantChannelCount: 0,
		},
		{
			name:             "失败：来源已删除，应返回 ErrSourceNotFound",
			sourceID:         deletedSource.ID,
			userID:           testUser.ID,
			wantErrCode:      xerr.ErrSourceNotFound.Code,
			wantSuccess:      false,
			wantChannelCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewSourceService(db, logger)
			result, err := svc.GetSourceDetail(ctx, tt.sourceID, tt.userID)

			if tt.wantSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotNil(t, result.Source)
				assert.Equal(t, testSource.ID, result.Source.ID)
				assert.Equal(t, "Detail Source", result.Source.Name)
				assert.Equal(t, testUser.ID, result.Source.UserID)
				assert.Equal(t, "src-detail-token", result.Source.Token)
				assert.Equal(t, model.SourceStatusActive, result.Source.Status)
				assert.NotZero(t, result.Source.CreatedAt, "CreatedAt 不应为零")
				assert.NotZero(t, result.Source.UpdatedAt, "UpdatedAt 不应为零")
				assert.Equal(t, int64(0), result.Source.LastUsedAt, "LastUsedAt 应为 0（未使用）")

				// 验证绑定的渠道列表
				assert.Len(t, result.Channels, tt.wantChannelCount, "绑定渠道数量应与预期一致")
				
				// 验证渠道信息完整
				channelNames := make(map[string]bool)
				for _, ch := range result.Channels {
					channelNames[ch.Name] = true
					assert.NotZero(t, ch.ID, "渠道 ID 不应为零")
					assert.Equal(t, testUser.ID, ch.UserID, "渠道 UserID 应匹配")
				}
				assert.True(t, channelNames["Detail Channel 1"], "应包含 Detail Channel 1")
				assert.True(t, channelNames["Detail Channel 2"], "应包含 Detail Channel 2")
			} else {
				assert.Error(t, err)
				assert.Nil(t, result)
				appErr, ok := err.(*xerr.AppError)
				assert.True(t, ok, "错误类型应为 *xerr.AppError")
				assert.Equal(t, tt.wantErrCode, appErr.Code, "错误码不匹配")
			}
		})
	}
}

// ============================================================================
// TestSourceService_GetSourceToken 查询来源令牌功能测试
// ============================================================================

// TestSourceService_GetSourceToken 测试查询来源令牌功能
// 覆盖场景：成功获取令牌、来源不存在或已删除
func TestSourceService_GetSourceToken(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	ctx := context.Background()

	// 预创建测试用户和来源
	testUser := CreateTestUser(t, db, "token_source_user", "Password1")
	testSource := CreateTestSource(t, db, testUser.ID, "Token Source", "src-get-token-test")

	// 创建已删除的来源
	deletedSource := CreateTestSource(t, db, testUser.ID, "Deleted Token Source", "src-deleted-token-test")
	db.Model(&model.Source{}).Where("id = ?", deletedSource.ID).Update("status", model.SourceStatusDeleted)

	tests := []struct {
		name        string      // 测试用例名称
		sourceID    int64       // 来源 ID
		userID      int64       // 用户 ID
		wantToken   string      // 期望的令牌
		wantErrCode int         // 期望的错误码（0 表示无错误）
		wantSuccess bool        // 是否期望成功
	}{
		{
			name:        "成功：获取来源令牌",
			sourceID:    testSource.ID,
			userID:      testUser.ID,
			wantToken:   "src-get-token-test",
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:        "失败：来源不存在或已删除，应返回 ErrSourceNotFound",
			sourceID:    deletedSource.ID,
			userID:      testUser.ID,
			wantToken:   "",
			wantErrCode: xerr.ErrSourceNotFound.Code,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewSourceService(db, logger)
			token, err := svc.GetSourceToken(ctx, tt.sourceID, tt.userID)

			if tt.wantSuccess {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantToken, token, "返回的令牌应与预期一致")
			} else {
				assert.Error(t, err)
				assert.Empty(t, token, "失败时令牌应为空")
				appErr, ok := err.(*xerr.AppError)
				assert.True(t, ok, "错误类型应为 *xerr.AppError")
				assert.Equal(t, tt.wantErrCode, appErr.Code, "错误码不匹配")
			}
		})
	}
}

// ============================================================================
// TestSourceService_ResetSourceToken 重置来源令牌功能测试
// ============================================================================

// TestSourceService_ResetSourceToken 测试重置来源令牌功能
// 覆盖场景：成功重置令牌、来源不存在、来源已删除
func TestSourceService_ResetSourceToken(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	ctx := context.Background()

	// 预创建测试用户和来源
	testUser := CreateTestUser(t, db, "reset_token_user", "Password1")
	testSource := CreateTestSource(t, db, testUser.ID, "Reset Token Source", "src-old-token")

	// 创建已删除的来源
	deletedSource := CreateTestSource(t, db, testUser.ID, "Deleted Reset Source", "src-deleted-reset-token")
	db.Model(&model.Source{}).Where("id = ?", deletedSource.ID).Update("status", model.SourceStatusDeleted)

	tests := []struct {
		name        string      // 测试用例名称
		sourceID    int64       // 来源 ID
		userID      int64       // 用户 ID
		wantErrCode int         // 期望的错误码（0 表示无错误）
		wantSuccess bool        // 是否期望成功
	}{
		{
			name:        "成功：重置来源令牌",
			sourceID:    testSource.ID,
			userID:      testUser.ID,
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:        "失败：来源不存在，应返回 ErrSourceNotFound",
			sourceID:    99999,
			userID:      testUser.ID,
			wantErrCode: xerr.ErrSourceNotFound.Code,
			wantSuccess: false,
		},
		{
			name:        "失败：来源已删除，应返回 ErrSourceAlreadyDeleted",
			sourceID:    deletedSource.ID,
			userID:      testUser.ID,
			wantErrCode: xerr.ErrSourceAlreadyDeleted.Code,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewSourceService(db, logger)
			newToken, err := svc.ResetSourceToken(ctx, tt.sourceID, tt.userID)

			if tt.wantSuccess {
				assert.NoError(t, err)
				assert.NotEmpty(t, newToken, "新令牌不应为空")
				assert.NotEqual(t, "src-old-token", newToken, "新令牌应与旧令牌不同")
				assert.True(t, len(newToken) == 35, "Token 长度应为 35 位")

				// 验证数据库中令牌已更新
				var updatedSource model.Source
				findErr := db.Where("id = ?", tt.sourceID).First(&updatedSource).Error
				assert.NoError(t, findErr, "应能查询到来源")
				assert.Equal(t, newToken, updatedSource.Token, "数据库中的令牌应已更新")
			} else {
				assert.Error(t, err)
				assert.Empty(t, newToken, "失败时新令牌应为空")
				appErr, ok := err.(*xerr.AppError)
				assert.True(t, ok, "错误类型应为 *xerr.AppError")
				assert.Equal(t, tt.wantErrCode, appErr.Code, "错误码不匹配")
			}
		})
	}
}

// ============================================================================
// TestSourceService_DisableSource 停用消息来源功能测试
// ============================================================================

// TestSourceService_DisableSource 测试停用消息来源功能
// 覆盖场景：成功停用活跃来源、来源已停用、来源已删除
func TestSourceService_DisableSource(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	ctx := context.Background()

	// 预创建测试用户和来源
	testUser := CreateTestUser(t, db, "disable_source_user", "Password1")
	
	// 创建活跃来源
	activeSource := CreateTestSource(t, db, testUser.ID, "Active Disable Source", "src-active-disable")
	
	// 创建已停用来源
	disabledSource := CreateTestSource(t, db, testUser.ID, "Already Disabled Source", "src-disabled-disable")
	db.Model(&model.Source{}).Where("id = ?", disabledSource.ID).Update("status", model.SourceStatusDisabled)
	
	// 创建已删除来源
	deletedSource := CreateTestSource(t, db, testUser.ID, "Deleted Disable Source", "src-deleted-disable")
	db.Model(&model.Source{}).Where("id = ?", deletedSource.ID).Update("status", model.SourceStatusDeleted)

	tests := []struct {
		name        string      // 测试用例名称
		sourceID    int64       // 来源 ID
		userID      int64       // 用户 ID
		wantErrCode int         // 期望的错误码（0 表示无错误）
		wantSuccess bool        // 是否期望成功
	}{
		{
			name:        "成功：停用活跃的来源",
			sourceID:    activeSource.ID,
			userID:      testUser.ID,
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:        "失败：来源已停用，应返回 ErrSourceAlreadyDisabled",
			sourceID:    disabledSource.ID,
			userID:      testUser.ID,
			wantErrCode: xerr.ErrSourceAlreadyDisabled.Code,
			wantSuccess: false,
		},
		{
			name:        "失败：来源已删除，应返回 ErrSourceAlreadyDeleted",
			sourceID:    deletedSource.ID,
			userID:      testUser.ID,
			wantErrCode: xerr.ErrSourceAlreadyDeleted.Code,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewSourceService(db, logger)
			err := svc.DisableSource(ctx, tt.sourceID, tt.userID)

			if tt.wantSuccess {
				assert.NoError(t, err)

				// 验证数据库中来源状态已更新为停用
				var updatedSource model.Source
				findErr := db.Where("id = ?", tt.sourceID).First(&updatedSource).Error
				assert.NoError(t, findErr, "应能查询到来源")
				assert.Equal(t, model.SourceStatusDisabled, updatedSource.Status, "来源状态应为已停用")
			} else {
				assert.Error(t, err)
				appErr, ok := err.(*xerr.AppError)
				assert.True(t, ok, "错误类型应为 *xerr.AppError")
				assert.Equal(t, tt.wantErrCode, appErr.Code, "错误码不匹配")
			}
		})
	}
}

// ============================================================================
// TestSourceService_EnableSource 启用消息来源功能测试
// ============================================================================

// TestSourceService_EnableSource 测试启用消息来源功能
// 覆盖场景：成功停用已停用的来源、来源已启用、来源已删除
func TestSourceService_EnableSource(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	ctx := context.Background()

	// 预创建测试用户和来源
	testUser := CreateTestUser(t, db, "enable_source_user", "Password1")
	
	// 创建已停用来源
	disabledSource := CreateTestSource(t, db, testUser.ID, "Disabled Enable Source", "src-disabled-enable")
	db.Model(&model.Source{}).Where("id = ?", disabledSource.ID).Update("status", model.SourceStatusDisabled)
	
	// 创建活跃来源
	activeSource := CreateTestSource(t, db, testUser.ID, "Already Active Source", "src-active-enable")
	
	// 创建已删除来源
	deletedSource := CreateTestSource(t, db, testUser.ID, "Deleted Enable Source", "src-deleted-enable")
	db.Model(&model.Source{}).Where("id = ?", deletedSource.ID).Update("status", model.SourceStatusDeleted)

	tests := []struct {
		name        string      // 测试用例名称
		sourceID    int64       // 来源 ID
		userID      int64       // 用户 ID
		wantErrCode int         // 期望的错误码（0 表示无错误）
		wantSuccess bool        // 是否期望成功
	}{
		{
			name:        "成功：启用已停用的来源",
			sourceID:    disabledSource.ID,
			userID:      testUser.ID,
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:        "失败：来源已启用，应返回 ErrSourceAlreadyEnabled",
			sourceID:    activeSource.ID,
			userID:      testUser.ID,
			wantErrCode: xerr.ErrSourceAlreadyEnabled.Code,
			wantSuccess: false,
		},
		{
			name:        "失败：来源已删除，应返回 ErrSourceAlreadyDeleted",
			sourceID:    deletedSource.ID,
			userID:      testUser.ID,
			wantErrCode: xerr.ErrSourceAlreadyDeleted.Code,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewSourceService(db, logger)
			err := svc.EnableSource(ctx, tt.sourceID, tt.userID)

			if tt.wantSuccess {
				assert.NoError(t, err)

				// 验证数据库中来源状态已更新为启用
				var updatedSource model.Source
				findErr := db.Where("id = ?", tt.sourceID).First(&updatedSource).Error
				assert.NoError(t, findErr, "应能查询到来源")
				assert.Equal(t, model.SourceStatusActive, updatedSource.Status, "来源状态应为已启用")
			} else {
				assert.Error(t, err)
				appErr, ok := err.(*xerr.AppError)
				assert.True(t, ok, "错误类型应为 *xerr.AppError")
				assert.Equal(t, tt.wantErrCode, appErr.Code, "错误码不匹配")
			}
		})
	}
}

// ============================================================================
// TestSourceService_DeleteSource 删除消息来源功能测试
// ============================================================================

// TestSourceService_DeleteSource 测试删除消息来源功能（软删除）
// 覆盖场景：成功删除（级联删除渠道绑定）、来源已删除、来源不存在
func TestSourceService_DeleteSource(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	ctx := context.Background()

	// 预创建测试用户、来源和渠道
	testUser := CreateTestUser(t, db, "delete_source_user", "Password1")
	testSource := CreateTestSource(t, db, testUser.ID, "Delete Source", "src-delete-token")
	channel1 := CreateTestChannel(t, db, testUser.ID, "webhook", "Delete Channel 1")
	channel2 := CreateTestChannel(t, db, testUser.ID, "dingtalk", "Delete Channel 2")
	
	// 绑定渠道
	BindSourceToChannel(t, db, testSource.ID, channel1.ID)
	BindSourceToChannel(t, db, testSource.ID, channel2.ID)

	// 创建已删除的来源
	deletedSource := CreateTestSource(t, db, testUser.ID, "Already Deleted Source", "src-deleted-delete")
	db.Model(&model.Source{}).Where("id = ?", deletedSource.ID).Update("status", model.SourceStatusDeleted)

	tests := []struct {
		name        string      // 测试用例名称
		sourceID    int64       // 来源 ID
		userID      int64       // 用户 ID
		wantErrCode int         // 期望的错误码（0 表示无错误）
		wantSuccess bool        // 是否期望成功
	}{
		{
			name:        "成功：删除来源（应级联删除渠道绑定）",
			sourceID:    testSource.ID,
			userID:      testUser.ID,
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:        "失败：来源已删除，应返回 ErrSourceAlreadyDeleted",
			sourceID:    deletedSource.ID,
			userID:      testUser.ID,
			wantErrCode: xerr.ErrSourceAlreadyDeleted.Code,
			wantSuccess: false,
		},
		{
			name:        "失败：来源不存在，应返回 ErrSourceNotFound",
			sourceID:    99999,
			userID:      testUser.ID,
			wantErrCode: xerr.ErrSourceNotFound.Code,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewSourceService(db, logger)
			err := svc.DeleteSource(ctx, tt.sourceID, tt.userID)

			if tt.wantSuccess {
				assert.NoError(t, err)

				// 验证来源已被软删除（状态为 -1）
				var deletedSource model.Source
				findErr := db.Where("id = ?", tt.sourceID).First(&deletedSource).Error
				assert.NoError(t, findErr, "应能查询到来源记录")
				assert.Equal(t, model.SourceStatusDeleted, deletedSource.Status, "来源状态应为已删除")

				// 验证渠道绑定也被级联软删除
				var activeBindingCount int64
				db.Model(&model.SourceChannel{}).
					Where("source_id = ? AND status = ?", tt.sourceID, model.SourceChannelStatusActive).
					Count(&activeBindingCount)
				assert.Equal(t, int64(0), activeBindingCount, "所有渠道绑定应被级联软删除")

				// 验证软删除后的绑定记录仍然存在（状态为 -1）
				var deletedBindingCount int64
				db.Model(&model.SourceChannel{}).
					Where("source_id = ? AND status = ?", tt.sourceID, model.SourceChannelStatusDeleted).
					Count(&deletedBindingCount)
				assert.Equal(t, int64(2), deletedBindingCount, "应有 2 条已删除的绑定记录")
			} else {
				assert.Error(t, err)
				appErr, ok := err.(*xerr.AppError)
				assert.True(t, ok, "错误类型应为 *xerr.AppError")
				assert.Equal(t, tt.wantErrCode, appErr.Code, "错误码不匹配")
			}
		})
	}
}

// ============================================================================
// TestSourceService_VerifyPassword 验证用户密码功能测试
// ============================================================================

// TestSourceService_VerifyPassword 测试用户密码验证功能
// 覆盖场景：正确密码验证通过、错误密码验证失败
func TestSourceService_VerifyPassword(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	ctx := context.Background()

	// 预创建测试用户（密码为 Password1）
	testUser := CreateTestUser(t, db, "verify_password_user", "Password1")

	tests := []struct {
		name        string      // 测试用例名称
		userID      int64       // 用户 ID
		password    string      // 待验证的密码
		wantErrCode int         // 期望的错误码（0 表示无错误）
		wantSuccess bool        // 是否期望成功
	}{
		{
			name:        "成功：验证正确的密码",
			userID:      testUser.ID,
			password:    "Password1",
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:        "失败：验证错误的密码，应返回 ErrUnauthorized",
			userID:      testUser.ID,
			password:    "WrongPassword123",
			wantErrCode: xerr.ErrUnauthorized.Code,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewSourceService(db, logger)
			err := svc.VerifyPassword(ctx, tt.userID, tt.password)

			if tt.wantSuccess {
				assert.NoError(t, err, "正确密码应验证通过")
			} else {
				assert.Error(t, err)
				appErr, ok := err.(*xerr.AppError)
				assert.True(t, ok, "错误类型应为 *xerr.AppError")
				assert.Equal(t, tt.wantErrCode, appErr.Code, "错误码不匹配")
			}
		})
	}
}
