package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"

	"octotify/internal/client/ilink"
	"octotify/internal/handler/dto"
	"octotify/internal/model"
	"octotify/internal/sender"
	"octotify/pkg/aescipher"
	"octotify/pkg/xerr"
)

// ============================================================================
// TestChannelService_CreateChannel 创建推送渠道功能测试
// ============================================================================

// TestChannelService_CreateChannel 测试创建推送渠道功能
// 覆盖场景：正常创建、数据库插入错误
func TestChannelService_CreateChannel(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	ctx := context.Background()

	// 创建 SenderFactory（使用真实工厂，CreateChannel 内部不会调用 senderFactory.Create）
	ilinkClient := ilink.NewClient(logger)
	factory := sender.NewSenderFactory(logger, ilinkClient)

	// 预创建测试用户
	testUser := CreateTestUser(t, db, "create_channel_user", "Password1")

	tests := []struct {
		name        string                // 测试用例名称
		userID      int64                 // 用户 ID
		req         *dto.CreateChannelReq // 创建请求
		setup       func()                // 测试前置操作（可选）
		wantErrCode int                   // 期望的错误码（0 表示无错误）
		wantSuccess bool                  // 是否期望成功
	}{
		{
			name:   "成功：创建飞书渠道",
			userID: testUser.ID,
			req: &dto.CreateChannelReq{
				Type:   "feishu",
				Name:   "飞书-运维群",
				Config: dto.ChannelConfig{"webhook_url": "https://open.feishu.cn/open-apis/bot/v2/hook/test"},
			},
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:   "成功：创建钉钉渠道",
			userID: testUser.ID,
			req: &dto.CreateChannelReq{
				Type:   "dingtalk",
				Name:   "钉钉-告警群",
				Config: dto.ChannelConfig{"webhook": "https://oapi.dingtalk.com/robot/send?access_token=test"},
			},
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:   "失败：数据库连接已关闭，应返回 ErrChannelInsertFailed",
			userID: testUser.ID,
			req: &dto.CreateChannelReq{
				Type:   "webhook",
				Name:   "DB Error Channel",
				Config: dto.ChannelConfig{"url": "https://example.com"},
			},
			setup: func() {
				sqlDB, err := db.DB()
				require.NoError(t, err, "获取底层 sql.DB 失败")
				err = sqlDB.Close()
				require.NoError(t, err, "关闭数据库连接失败")
			},
			wantErrCode: xerr.ErrChannelInsertFailed.Code,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			svc := NewChannelService(db, logger, factory, ilinkClient)
			result, err := svc.CreateChannel(ctx, tt.userID, tt.req)

			if tt.wantSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.req.Name, result.Name)
				assert.Equal(t, tt.req.Type, result.Type)
				assert.Equal(t, tt.req.Config, result.Config)
				assert.Equal(t, tt.userID, result.UserID)
				assert.Equal(t, model.ChannelStatusActive, result.Status)
				assert.NotZero(t, result.ID, "渠道 ID 不应为零")
				assert.NotZero(t, result.CreatedAt, "CreatedAt 不应为零")
				assert.NotZero(t, result.UpdatedAt, "UpdatedAt 不应为零")
				assert.Equal(t, int64(0), result.LastUsedAt, "LastUsedAt 应为 0（未使用）")

				// 验证数据库中确实创建了渠道记录
				var channel model.Channel
				findErr := db.Where("name = ?", tt.req.Name).First(&channel).Error
				assert.NoError(t, findErr, "数据库中应存在该渠道")
				assert.Equal(t, tt.req.Name, channel.Name)
				assert.Equal(t, tt.req.Type, channel.Type)
				assert.Equal(t, model.ChannelStatusActive, channel.Status)
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
// TestChannelService_ListChannels 分页查询渠道列表功能测试
// ============================================================================

// TestChannelService_ListChannels 测试分页查询推送渠道列表功能
// 覆盖场景：正常分页查询、排除已删除渠道
func TestChannelService_ListChannels(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	ctx := context.Background()

	// 创建 SenderFactory（使用真实工厂，测试中不需要调用 Create）
	ilinkClient := ilink.NewClient(logger)
	factory := sender.NewSenderFactory(logger, ilinkClient)

	// 预创建测试用户和多个渠道
	testUser := CreateTestUser(t, db, "list_channels_user", "Password1")
	CreateTestChannel(t, db, testUser.ID, "feishu", "Active Channel 1")
	CreateTestChannel(t, db, testUser.ID, "dingtalk", "Active Channel 2")
	// 将第三个渠道标记为已删除
	deletedChannel := CreateTestChannel(t, db, testUser.ID, "webhook", "Deleted Channel")
	db.Model(&model.Channel{}).Where("id = ?", deletedChannel.ID).Update("status", model.ChannelStatusDeleted)

	tests := []struct {
		name        string       // 测试用例名称
		userID      int64        // 用户 ID
		pageReq     *dto.PageReq // 分页请求
		wantTotal   int64        // 期望的总数
		wantCount   int          // 期望的返回数量
		wantErrCode int          // 期望的错误码（0 表示无错误）
		wantSuccess bool         // 是否期望成功
	}{
		{
			name:        "成功：分页查询活跃渠道列表（应排除已删除渠道）",
			userID:      testUser.ID,
			pageReq:     &dto.PageReq{Page: 1, PageSize: 10},
			wantTotal:   2, // 只应返回 2 个活跃渠道
			wantCount:   2,
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:        "成功：分页查询第一页，限制每页 1 条",
			userID:      testUser.ID,
			pageReq:     &dto.PageReq{Page: 1, PageSize: 1},
			wantTotal:   2, // 总数仍为 2
			wantCount:   1, // 但只返回 1 条
			wantErrCode: 0,
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewChannelService(db, logger, factory, ilinkClient)
			list, total, err := svc.ListChannels(ctx, tt.userID, tt.pageReq)

			if tt.wantSuccess {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTotal, total, "总数应与预期一致")
				assert.Len(t, list, tt.wantCount, "返回数量应与预期一致")

				// 验证返回列表中不包含已删除的渠道
				for _, item := range list {
					assert.NotEqual(t, model.ChannelStatusDeleted, item.Status, "列表中不应包含已删除的渠道")
					assert.NotEqual(t, "Deleted Channel", item.Name, "列表中不应包含名为 'Deleted Channel' 的渠道")
				}

				// 如果只查询 1 条，验证按创建时间降序排列（最新优先）
				if tt.pageReq.PageSize == 1 {
					assert.Equal(t, "Active Channel 2", list[0].Name, "应返回最新创建的渠道")
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
// TestChannelService_UpdateChannel 更新渠道功能测试
// ============================================================================

// TestChannelService_UpdateChannel 测试更新推送渠道功能
// 覆盖场景：成功更新、渠道不存在、渠道已删除、权限不足（其他用户）
func TestChannelService_UpdateChannel(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	ctx := context.Background()

	// 创建 SenderFactory（使用真实工厂，测试中不需要调用 Create）
	ilinkClient := ilink.NewClient(logger)
	factory := sender.NewSenderFactory(logger, ilinkClient)

	// 预创建测试用户和渠道
	testUser := CreateTestUser(t, db, "update_channel_user", "Password1")
	otherUser := CreateTestUser(t, db, "other_channel_user", "Password1")
	testChannel := CreateTestChannel(t, db, testUser.ID, "feishu", "Original Channel")
	otherUserChannel := CreateTestChannel(t, db, otherUser.ID, "dingtalk", "Other User Channel")

	// 创建已删除的渠道
	deletedChannel := CreateTestChannel(t, db, testUser.ID, "webhook", "Deleted Update Channel")
	db.Model(&model.Channel{}).Where("id = ?", deletedChannel.ID).Update("status", model.ChannelStatusDeleted)

	tests := []struct {
		name        string                // 测试用例名称
		channelID   int64                 // 渠道 ID
		userID      int64                 // 操作用户 ID
		req         *dto.UpdateChannelReq // 更新请求
		wantErrCode int                   // 期望的错误码（0 表示无错误）
		wantSuccess bool                  // 是否期望成功
	}{
		{
			name:      "成功：更新渠道名称和配置",
			channelID: testChannel.ID,
			userID:    testUser.ID,
			req: &dto.UpdateChannelReq{
				Name:   "Updated Channel Name",
				Config: dto.ChannelConfig{"webhook_url": "https://open.feishu.cn/updated"},
			},
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:      "失败：渠道不存在，应返回 ErrChannelNotFound",
			channelID: 99999,
			userID:    testUser.ID,
			req: &dto.UpdateChannelReq{
				Name:   "Non-existent Channel",
				Config: dto.ChannelConfig{},
			},
			wantErrCode: xerr.ErrChannelNotFound.Code,
			wantSuccess: false,
		},
		{
			name:      "失败：渠道已删除，应返回 ErrChannelAlreadyDeleted",
			channelID: deletedChannel.ID,
			userID:    testUser.ID,
			req: &dto.UpdateChannelReq{
				Name:   "Updated Deleted Channel",
				Config: dto.ChannelConfig{},
			},
			wantErrCode: xerr.ErrChannelAlreadyDeleted.Code,
			wantSuccess: false,
		},
		{
			name:      "失败：其他用户尝试更新，应返回 ErrChannelNotFound（权限隔离）",
			channelID: otherUserChannel.ID,
			userID:    testUser.ID,
			req: &dto.UpdateChannelReq{
				Name:   "Hacked Channel Name",
				Config: dto.ChannelConfig{},
			},
			wantErrCode: xerr.ErrChannelNotFound.Code,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewChannelService(db, logger, factory, ilinkClient)
			err := svc.UpdateChannel(ctx, tt.userID, tt.channelID, tt.req)

			if tt.wantSuccess {
				assert.NoError(t, err)

				// 验证数据库中的渠道已更新
				var updatedChannel model.Channel
				findErr := db.Where("id = ?", tt.channelID).First(&updatedChannel).Error
				assert.NoError(t, findErr, "应能查询到渠道")
				assert.Equal(t, "Updated Channel Name", updatedChannel.Name, "名称应已更新")
			} else {
				assert.Error(t, err)
				appErr, ok := err.(*xerr.AppError)
				assert.True(t, ok, "错误类型应为 *xerr.AppError")
				assert.Equal(t, tt.wantErrCode, appErr.Code, "错误码不匹配")

				// 对于失败场景，验证渠道数据未被修改
				if tt.channelID != 99999 {
					var originalChannel model.Channel
					findErr := db.Where("id = ?", tt.channelID).First(&originalChannel).Error
					assert.NoError(t, findErr, "应能查询到渠道")
					// 确保名称没有被修改
					if tt.channelID == testChannel.ID {
						assert.Equal(t, "Original Channel", originalChannel.Name, "名称不应被修改")
					}
				}
			}
		})
	}
}

// ============================================================================
// TestChannelService_GetChannelByID 查询渠道详情功能测试
// ============================================================================

// TestChannelService_GetChannelByID 测试根据 ID 查询渠道详情功能
// 覆盖场景：成功获取详情、渠道不存在、渠道已删除
func TestChannelService_GetChannelByID(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	ctx := context.Background()

	// 创建 SenderFactory（使用真实工厂，测试中不需要调用 Create）
	ilinkClient := ilink.NewClient(logger)
	factory := sender.NewSenderFactory(logger, ilinkClient)

	// 预创建测试用户和渠道
	testUser := CreateTestUser(t, db, "get_channel_user", "Password1")
	testChannel := CreateTestChannel(t, db, testUser.ID, "feishu", "Get Detail Channel")

	// 创建已删除的渠道
	deletedChannel := CreateTestChannel(t, db, testUser.ID, "dingtalk", "Deleted Get Channel")
	db.Model(&model.Channel{}).Where("id = ?", deletedChannel.ID).Update("status", model.ChannelStatusDeleted)

	tests := []struct {
		name        string // 测试用例名称
		channelID   int64  // 渠道 ID
		userID      int64  // 用户 ID
		wantErrCode int    // 期望的错误码（0 表示无错误）
		wantSuccess bool   // 是否期望成功
	}{
		{
			name:        "成功：获取渠道详情",
			channelID:   testChannel.ID,
			userID:      testUser.ID,
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:        "失败：渠道不存在，应返回 ErrChannelNotFound",
			channelID:   99999,
			userID:      testUser.ID,
			wantErrCode: xerr.ErrChannelNotFound.Code,
			wantSuccess: false,
		},
		{
			name:        "失败：渠道已删除，应返回 ErrChannelNotFound",
			channelID:   deletedChannel.ID,
			userID:      testUser.ID,
			wantErrCode: xerr.ErrChannelNotFound.Code,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewChannelService(db, logger, factory, ilinkClient)
			result, err := svc.GetChannelByID(ctx, tt.userID, tt.channelID)

			if tt.wantSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, testChannel.ID, result.ID)
				assert.Equal(t, "Get Detail Channel", result.Name)
				assert.Equal(t, testUser.ID, result.UserID)
				assert.Equal(t, "feishu", result.Type)
				assert.Equal(t, model.ChannelStatusActive, result.Status)
				assert.NotZero(t, result.CreatedAt, "CreatedAt 不应为零")
				assert.NotZero(t, result.UpdatedAt, "UpdatedAt 不应为零")
				assert.Equal(t, int64(0), result.LastUsedAt, "LastUsedAt 应为 0（未使用）")
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
// TestChannelService_DisableChannel 停用渠道功能测试
// ============================================================================

// TestChannelService_DisableChannel 测试停用推送渠道功能
// 覆盖场景：成功停用活跃渠道、渠道已停用、渠道已删除
func TestChannelService_DisableChannel(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	ctx := context.Background()

	// 创建 SenderFactory（使用真实工厂，测试中不需要调用 Create）
	ilinkClient := ilink.NewClient(logger)
	factory := sender.NewSenderFactory(logger, ilinkClient)

	// 预创建测试用户和渠道
	testUser := CreateTestUser(t, db, "disable_channel_user", "Password1")

	// 创建活跃渠道
	activeChannel := CreateTestChannel(t, db, testUser.ID, "feishu", "Active Disable Channel")

	// 创建已停用渠道
	disabledChannel := CreateTestChannel(t, db, testUser.ID, "dingtalk", "Already Disabled Channel")
	db.Model(&model.Channel{}).Where("id = ?", disabledChannel.ID).Update("status", model.ChannelStatusDisabled)

	// 创建已删除渠道
	deletedChannel := CreateTestChannel(t, db, testUser.ID, "webhook", "Deleted Disable Channel")
	db.Model(&model.Channel{}).Where("id = ?", deletedChannel.ID).Update("status", model.ChannelStatusDeleted)

	tests := []struct {
		name        string // 测试用例名称
		channelID   int64  // 渠道 ID
		userID      int64  // 用户 ID
		wantErrCode int    // 期望的错误码（0 表示无错误）
		wantSuccess bool   // 是否期望成功
	}{
		{
			name:        "成功：停用活跃的渠道",
			channelID:   activeChannel.ID,
			userID:      testUser.ID,
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:        "失败：渠道已停用，应返回 ErrChannelAlreadyDisabled",
			channelID:   disabledChannel.ID,
			userID:      testUser.ID,
			wantErrCode: xerr.ErrChannelAlreadyDisabled.Code,
			wantSuccess: false,
		},
		{
			name:        "失败：渠道已删除，应返回 ErrChannelAlreadyDeleted",
			channelID:   deletedChannel.ID,
			userID:      testUser.ID,
			wantErrCode: xerr.ErrChannelAlreadyDeleted.Code,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewChannelService(db, logger, factory, ilinkClient)
			err := svc.DisableChannel(ctx, tt.userID, tt.channelID)

			if tt.wantSuccess {
				assert.NoError(t, err)

				// 验证数据库中渠道状态已更新为停用
				var updatedChannel model.Channel
				findErr := db.Where("id = ?", tt.channelID).First(&updatedChannel).Error
				assert.NoError(t, findErr, "应能查询到渠道")
				assert.Equal(t, model.ChannelStatusDisabled, updatedChannel.Status, "渠道状态应为已停用")
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
// TestChannelService_EnableChannel 启用渠道功能测试
// ============================================================================

// TestChannelService_EnableChannel 测试启用推送渠道功能
// 覆盖场景：成功启用已停用渠道、渠道已启用、渠道已删除
func TestChannelService_EnableChannel(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	ctx := context.Background()

	// 创建 SenderFactory（使用真实工厂，测试中不需要调用 Create）
	ilinkClient := ilink.NewClient(logger)
	factory := sender.NewSenderFactory(logger, ilinkClient)

	// 预创建测试用户和渠道
	testUser := CreateTestUser(t, db, "enable_channel_user", "Password1")

	// 创建已停用渠道
	disabledChannel := CreateTestChannel(t, db, testUser.ID, "feishu", "Disabled Enable Channel")
	db.Model(&model.Channel{}).Where("id = ?", disabledChannel.ID).Update("status", model.ChannelStatusDisabled)

	// 创建活跃渠道
	activeChannel := CreateTestChannel(t, db, testUser.ID, "dingtalk", "Already Active Channel")

	// 创建已删除渠道
	deletedChannel := CreateTestChannel(t, db, testUser.ID, "webhook", "Deleted Enable Channel")
	db.Model(&model.Channel{}).Where("id = ?", deletedChannel.ID).Update("status", model.ChannelStatusDeleted)

	tests := []struct {
		name        string // 测试用例名称
		channelID   int64  // 渠道 ID
		userID      int64  // 用户 ID
		wantErrCode int    // 期望的错误码（0 表示无错误）
		wantSuccess bool   // 是否期望成功
	}{
		{
			name:        "成功：启用已停用的渠道",
			channelID:   disabledChannel.ID,
			userID:      testUser.ID,
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:        "失败：渠道已启用，应返回 ErrChannelAlreadyEnabled",
			channelID:   activeChannel.ID,
			userID:      testUser.ID,
			wantErrCode: xerr.ErrChannelAlreadyEnabled.Code,
			wantSuccess: false,
		},
		{
			name:        "失败：渠道已删除，应返回 ErrChannelAlreadyDeleted",
			channelID:   deletedChannel.ID,
			userID:      testUser.ID,
			wantErrCode: xerr.ErrChannelAlreadyDeleted.Code,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewChannelService(db, logger, factory, ilinkClient)
			err := svc.EnableChannel(ctx, tt.userID, tt.channelID)

			if tt.wantSuccess {
				assert.NoError(t, err)

				// 验证数据库中渠道状态已更新为启用
				var updatedChannel model.Channel
				findErr := db.Where("id = ?", tt.channelID).First(&updatedChannel).Error
				assert.NoError(t, findErr, "应能查询到渠道")
				assert.Equal(t, model.ChannelStatusActive, updatedChannel.Status, "渠道状态应为已启用")
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
// TestChannelService_DeleteChannel 删除渠道功能测试
// ============================================================================

// TestChannelService_DeleteChannel 测试删除推送渠道功能（软删除）
// 覆盖场景：成功删除（级联删除 source_channel 关联）、渠道已删除、渠道不存在
func TestChannelService_DeleteChannel(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	ctx := context.Background()

	// 创建 SenderFactory（使用真实工厂，测试中不需要调用 Create）
	ilinkClient := ilink.NewClient(logger)
	factory := sender.NewSenderFactory(logger, ilinkClient)

	// 预创建测试用户、渠道和来源
	testUser := CreateTestUser(t, db, "delete_channel_user", "Password1")
	testChannel := CreateTestChannel(t, db, testUser.ID, "feishu", "Delete Channel")
	testSource := CreateTestSource(t, db, testUser.ID, "Delete Source", "src-delete-channel-test")

	// 绑定来源到渠道（创建关联关系）
	BindSourceToChannel(t, db, testSource.ID, testChannel.ID)

	// 创建已删除的渠道
	deletedChannel := CreateTestChannel(t, db, testUser.ID, "dingtalk", "Already Deleted Delete Channel")
	db.Model(&model.Channel{}).Where("id = ?", deletedChannel.ID).Update("status", model.ChannelStatusDeleted)

	tests := []struct {
		name        string // 测试用例名称
		channelID   int64  // 渠道 ID
		userID      int64  // 用户 ID
		wantErrCode int    // 期望的错误码（0 表示无错误）
		wantSuccess bool   // 是否期望成功
	}{
		{
			name:        "成功：删除渠道（应级联删除 source_channel 关联）",
			channelID:   testChannel.ID,
			userID:      testUser.ID,
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:        "失败：渠道已删除，应返回 ErrChannelAlreadyDeleted",
			channelID:   deletedChannel.ID,
			userID:      testUser.ID,
			wantErrCode: xerr.ErrChannelAlreadyDeleted.Code,
			wantSuccess: false,
		},
		{
			name:        "失败：渠道不存在，应返回 ErrChannelNotFound",
			channelID:   99999,
			userID:      testUser.ID,
			wantErrCode: xerr.ErrChannelNotFound.Code,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewChannelService(db, logger, factory, ilinkClient)
			err := svc.DeleteChannel(ctx, tt.userID, tt.channelID)

			if tt.wantSuccess {
				assert.NoError(t, err)

				// 验证渠道已被软删除（状态为 -1）
				var deletedChannel model.Channel
				findErr := db.Where("id = ?", tt.channelID).First(&deletedChannel).Error
				assert.NoError(t, findErr, "应能查询到渠道记录")
				assert.Equal(t, model.ChannelStatusDeleted, deletedChannel.Status, "渠道状态应为已删除")

				// 验证 source_channel 关联也被级联软删除
				var activeBindingCount int64
				db.Model(&model.SourceChannel{}).
					Where("channel_id = ? AND status = ?", tt.channelID, model.SourceChannelStatusActive).
					Count(&activeBindingCount)
				assert.Equal(t, int64(0), activeBindingCount, "所有渠道关联应被级联软删除")

				// 验证软删除后的关联记录仍然存在（状态为 -1）
				var deletedBindingCount int64
				db.Model(&model.SourceChannel{}).
					Where("channel_id = ? AND status = ?", tt.channelID, model.SourceChannelStatusDeleted).
					Count(&deletedBindingCount)
				assert.Equal(t, int64(1), deletedBindingCount, "应有 1 条已删除的关联记录")
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
// TestChannelService_GetChannelTypes 获取渠道类型元数据功能测试
// ============================================================================

// TestChannelService_GetChannelTypes 测试获取系统支持的渠道类型元数据功能
// 覆盖场景：返回渠道类型列表
func TestChannelService_GetChannelTypes(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)

	// 创建 SenderFactory（使用真实工厂，测试中不需要调用 Create）
	ilinkClient := ilink.NewClient(logger)
	factory := sender.NewSenderFactory(logger, ilinkClient)

	svc := NewChannelService(db, logger, factory, ilinkClient)
	result := svc.GetChannelTypes()

	// 验证返回的渠道类型列表非空
	assert.NotNil(t, result, "渠道类型列表不应为 nil")
	assert.NotEmpty(t, result, "渠道类型列表不应为空")

	// 验证当前支持的渠道类型（飞书）
	var hasFeishu bool
	for _, meta := range result {
		if meta.Type == "feishu" {
			hasFeishu = true
			assert.Equal(t, "飞书", meta.Name, "渠道类型名称应为 '飞书'")
			assert.NotEmpty(t, meta.Description, "渠道类型描述不应为空")
			assert.NotEmpty(t, meta.ConfigFields, "渠道配置字段不应为空")

			// 验证飞书渠道的配置字段定义
			fieldNames := make(map[string]bool)
			for _, field := range meta.ConfigFields {
				fieldNames[field.Name] = true
				assert.NotEmpty(t, field.Label, "字段标签不应为空")
				assert.NotEmpty(t, field.Type, "字段类型不应为空")
			}
			assert.True(t, fieldNames["webhook_url"], "应包含 webhook_url 字段")
			assert.True(t, fieldNames["secret"], "应包含 secret 字段")
		}
	}
	assert.True(t, hasFeishu, "渠道类型列表应包含飞书类型")
}

// ============================================================================
// TestChannelService_TestChannel 测试渠道连接功能
// ============================================================================

// TestChannelService_TestChannel 测试发送测试消息到指定渠道的功能
// 覆盖场景：成功测试活跃渠道、渠道不存在、渠道已删除、渠道已停用、发送器创建失败、发送失败
func TestChannelService_TestChannel(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	ctx := context.Background()

	// 预创建测试用户
	testUser := CreateTestUser(t, db, "test_channel_user", "Password1")

	// 创建共享的 ilink client 用于 service 构造
	ilinkClient := ilink.NewClient(logger)

	// 创建活跃的渠道
	activeChannel := CreateTestChannel(t, db, testUser.ID, "feishu", "Active Test Channel")
	// 创建未知渠道类型的渠道（用于测试发送器创建失败）
	unknownTypeChannel := CreateTestChannel(t, db, testUser.ID, "unknown_channel_type", "Unknown Type Channel")

	// 创建已删除的渠道
	deletedChannel := CreateTestChannel(t, db, testUser.ID, "dingtalk", "Deleted Test Channel")
	db.Model(&model.Channel{}).Where("id = ?", deletedChannel.ID).Update("status", model.ChannelStatusDeleted)

	// 创建已停用的渠道
	disabledChannel := CreateTestChannel(t, db, testUser.ID, "webhook", "Disabled Test Channel")
	db.Model(&model.Channel{}).Where("id = ?", disabledChannel.ID).Update("status", model.ChannelStatusDisabled)

	tests := []struct {
		name        string                      // 测试用例名称
		channelID   int64                       // 渠道 ID
		userID      int64                       // 用户 ID
		setupMock   func(*sender.SenderFactory) // Mock 设置
		wantErrCode int                         // 期望的错误码（0 表示无错误）
		wantSuccess bool                        // 是否期望成功
	}{
		{
			name:      "成功：测试活跃的渠道连接",
			channelID: activeChannel.ID,
			userID:    testUser.ID,
			setupMock: func(f *sender.SenderFactory) {
				f.Register("feishu", &sender.MockSender{SendFunc: func(ctx context.Context, config datatypes.JSON, title string, content string) error {
					return nil
				}})
			},
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:        "失败：渠道不存在，应返回 ErrChannelNotFound",
			channelID:   99999,
			userID:      testUser.ID,
			setupMock:   func(f *sender.SenderFactory) {},
			wantErrCode: xerr.ErrChannelNotFound.Code,
			wantSuccess: false,
		},
		{
			name:        "失败：渠道已删除，应返回 ErrChannelAlreadyDeleted",
			channelID:   deletedChannel.ID,
			userID:      testUser.ID,
			setupMock:   func(f *sender.SenderFactory) {},
			wantErrCode: xerr.ErrChannelAlreadyDeleted.Code,
			wantSuccess: false,
		},
		{
			name:        "失败：渠道已停用，应返回 ErrChannelAlreadyDisabled",
			channelID:   disabledChannel.ID,
			userID:      testUser.ID,
			setupMock:   func(f *sender.SenderFactory) {},
			wantErrCode: xerr.ErrChannelAlreadyDisabled.Code,
			wantSuccess: false,
		},
		{
			name:      "失败：发送器创建失败（不支持的渠道类型），应返回 ErrThirdPartyCallFailed",
			channelID: unknownTypeChannel.ID,
			userID:    testUser.ID,
			setupMock: func(f *sender.SenderFactory) {
				// 不注册任何发送器，让 Create 返回 ErrChannelInvalidType
			},
			wantErrCode: xerr.ErrThirdPartyCallFailed.Code,
			wantSuccess: false,
		},
		{
			name:      "失败：发送器发送失败，应返回 ErrThirdPartyCallFailed",
			channelID: activeChannel.ID,
			userID:    testUser.ID,
			setupMock: func(f *sender.SenderFactory) {
				f.Register("feishu", &sender.MockSender{SendFunc: func(ctx context.Context, config datatypes.JSON, title string, content string) error {
					return errors.New("feishu send failed")
				}})
			},
			wantErrCode: xerr.ErrThirdPartyCallFailed.Code,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 为每个子用例创建独立的发送器工厂
			testFactory := sender.NewSenderFactory(logger, ilinkClient)
			tt.setupMock(testFactory)

			svc := NewChannelService(db, logger, testFactory, ilinkClient)
			err := svc.TestChannel(ctx, tt.userID, tt.channelID)

			if tt.wantSuccess {
				assert.NoError(t, err)
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
// 微信ClawBot绑定流程测试
// ============================================================================

// newBindTestService 创建绑定测试用的 ChannelService
// BaseURL 指向模拟 iLink 服务器
func newBindTestService(t *testing.T, targetURL string) *ChannelService {
	t.Helper()
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	ilinkClient := ilink.NewClient(logger, ilink.WithBaseURL(targetURL))
	factory := sender.NewSenderFactory(logger, ilinkClient)
	return &ChannelService{
		db:            db,
		logger:        logger,
		senderFactory: factory,
		ilinkClient:   ilinkClient,
	}
}

// TestChannelService_StartBind 测试发起微信ClawBot扫码绑定
// 覆盖场景：成功获取二维码、iLink API 返回错误
func TestChannelService_StartBind(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		iLinkHandler http.HandlerFunc
		wantQRCode   string
		wantURL      string
		wantErr      bool
		wantErrCode  int
	}{
		{
			name: "成功：iLink API 返回 QRCode 和 QRCodeImgContent",
			iLinkHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{
					"qrcode":             "test-qrcode-abc123",
					"qrcode_img_content": "data:image/png;base64,iVBORw0KGgo=",
				})
			},
			wantQRCode: "test-qrcode-abc123",
			wantURL:    "data:image/png;base64,iVBORw0KGgo=",
			wantErr:    false,
		},
		{
			name: "失败：iLink API 返回 HTTP 500",
			iLinkHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("internal server error"))
			},
			wantQRCode:  "",
			wantURL:     "",
			wantErr:     true,
			wantErrCode: xerr.ErrQRCodeFetchFailed.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟 iLink 服务器
			iLinkServer := httptest.NewServer(tt.iLinkHandler)
			defer iLinkServer.Close()

			svc := newBindTestService(t, iLinkServer.URL)
			qrcode, qrcodeURL, err := svc.StartBind(ctx, 1)

			if tt.wantErr {
				assert.Error(t, err)
				appErr, ok := err.(*xerr.AppError)
				assert.True(t, ok, "错误类型应为 *xerr.AppError")
				assert.Equal(t, tt.wantErrCode, appErr.Code, "错误码不匹配")
				assert.Empty(t, qrcode)
				assert.Empty(t, qrcodeURL)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantQRCode, qrcode)
				assert.Equal(t, tt.wantURL, qrcodeURL)
			}
		})
	}
}

// TestChannelService_PollBindStatus 测试轮询微信ClawBot绑定状态
// 覆盖场景：pending、scanned、confirmed（含加密凭证验证）、expired、iLink API 失败
func TestChannelService_PollBindStatus(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		iLinkHandler    http.HandlerFunc
		wantStatus      string
		wantCreds       bool   // 是否期望返回凭证
		wantErr         bool   // 是否期望返回错误
		errContains     string // 错误信息应包含的子串
		wantILinkBotID  string
		wantILinkUserID string
	}{
		{
			name: "成功：iLink 返回 wait 状态",
			iLinkHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{
					"status": ilink.StatusWait,
				})
			},
			wantStatus: ilink.StatusWait,
			wantCreds:  false,
			wantErr:    false,
		},
		{
			name: "成功：iLink 返回 scanned 状态",
			iLinkHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{
					"status": ilink.StatusScanned,
				})
			},
			wantStatus: ilink.StatusScanned,
			wantCreds:  false,
			wantErr:    false,
		},
		{
			name: "成功：iLink 返回 confirmed 状态，包含加密凭证",
			iLinkHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{
					"status":        ilink.StatusConfirmed,
					"bot_token":     "test-bot-token-xyz",
					"ilink_bot_id":  "bot-789",
					"ilink_user_id": "user-012",
				})
			},
			wantStatus:      ilink.StatusConfirmed,
			wantCreds:       true,
			wantErr:         false,
			wantILinkBotID:  "bot-789",
			wantILinkUserID: "user-012",
		},
		{
			name: "成功：iLink 返回 expired 状态",
			iLinkHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{
					"status": ilink.StatusExpired,
				})
			},
			wantStatus: ilink.StatusExpired,
			wantCreds:  false,
			wantErr:    false,
		},
		{
			name: "失败：iLink API 返回 HTTP 500，应返回错误和空状态",
			iLinkHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("internal server error"))
			},
			wantStatus:  "",
			wantCreds:   false,
			wantErr:     true,
			errContains: "HTTP 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iLinkServer := httptest.NewServer(tt.iLinkHandler)
			defer iLinkServer.Close()

			svc := newBindTestService(t, iLinkServer.URL)
			status, credentials, err := svc.PollBindStatus(ctx, "test-qrcode")

			// 验证状态
			assert.Equal(t, tt.wantStatus, status)

			if tt.wantCreds {
				// 验证凭证结构完整（明文格式）
				assert.NoError(t, err)
				require.NotNil(t, credentials)
				assert.NotEmpty(t, credentials.BotToken, "BotToken 不应为空")
				assert.Equal(t, tt.wantILinkBotID, credentials.IlinkBotID)
				assert.Equal(t, tt.wantILinkUserID, credentials.IlinkUserID)
			} else {
				assert.Nil(t, credentials)
			}

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// TestChannelService_CheckActivationMessage 测试检查激活消息功能
// ============================================================================

func TestChannelService_CheckActivationMessage(t *testing.T) {
	logger := SetupTestLogger(t)
	ctx := context.Background()

	cipherB64, nonceB64, err := aescipher.GlobalEncryptBase64([]byte("test-bot-token"))
	require.NoError(t, err, "加密 bot_token 失败")

	invalidCipher := "invalid-cipher"
	invalidNonce := "invalid-nonce"

	tests := []struct {
		name             string
		cipherText       string
		nonce            string
		iLinkHandler     http.HandlerFunc
		wantResult       bool
		wantErr          bool
		errContains      string
	}{
		{
			name:       "成功：有激活消息",
			cipherText: cipherB64,
			nonce:      nonceB64,
			iLinkHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"ret": 0, "msgs": [{"from_user_id": "user@test.im.wechat", "message_type": 1}]}`))
			},
			wantResult:  true,
			wantErr:     false,
		},
		{
			name:       "成功：无激活消息",
			cipherText: cipherB64,
			nonce:      nonceB64,
			iLinkHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"ret": 0, "msgs": []}`))
			},
			wantResult:  false,
			wantErr:     false,
		},
		{
			name:        "失败：凭证解密失败",
			cipherText:  invalidCipher,
			nonce:       invalidNonce,
			iLinkHandler: nil,
			wantErr:     true,
			errContains: "凭证解密失败",
		},
		{
			name:       "失败：iLink API 调用失败",
			cipherText: cipherB64,
			nonce:      nonceB64,
			iLinkHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr:     true,
			errContains: "获取消息失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iLinkHandler := tt.iLinkHandler
			if iLinkHandler == nil {
				iLinkHandler = func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}
			}

			iLinkServer := httptest.NewServer(iLinkHandler)
			defer iLinkServer.Close()

			ilinkClient := ilink.NewClient(logger, ilink.WithBaseURL(iLinkServer.URL))
			factory := sender.NewSenderFactory(logger, ilinkClient)
			svc := NewChannelService(nil, logger, factory, ilinkClient)

			hasActivation, err := svc.CheckActivationMessage(ctx, tt.cipherText, tt.nonce)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResult, hasActivation)
			}
		})
	}
}
