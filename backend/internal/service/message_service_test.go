package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"octotify/internal/handler/dto"
	"octotify/internal/model"
	"octotify/internal/sender"
	"octotify/pkg/xerr"
)

// ============================================================================
// TestMessageService_ListMessages 测试消息列表查询功能
// ============================================================================

// TestMessageService_ListMessages 测试分页查询消息列表功能
// 覆盖场景：成功分页查询（按创建时间倒序）、排除已删除消息
func TestMessageService_ListMessages(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	ctx := context.Background()

	// 预创建测试用户和来源
	testUser := CreateTestUser(t, db, "list_messages_user", "Password1")
	channel := CreateTestChannel(t, db, testUser.ID, "webhook", "List Channel")
	testSource := CreateTestSource(t, db, testUser.ID, "List Source", "src-list-token")
	BindSourceToChannel(t, db, testSource.ID, channel.ID)

	// 预创建消息记录
	msg1 := &model.Message{
		SourceID:  testSource.ID,
		ChannelID: channel.ID,
		Title:     "Message 1",
		Content:   "Content 1",
		Status:    model.MessageStatusSuccess,
	}
	require.NoError(t, db.Create(msg1).Error)
	// 等待时间戳区分（确保创建时间有先后顺序）
	time.Sleep(2 * time.Millisecond)

	msg2 := &model.Message{
		SourceID:  testSource.ID,
		ChannelID: channel.ID,
		Title:     "Message 2",
		Content:   "Content 2",
		Status:    model.MessageStatusSuccess,
	}
	require.NoError(t, db.Create(msg2).Error)
	time.Sleep(2 * time.Millisecond)

	// 创建已删除的消息
	msg3 := &model.Message{
		SourceID:  testSource.ID,
		ChannelID: channel.ID,
		Title:     "Deleted Message",
		Content:   "Should be excluded",
		Status:    model.MessageStatusDeleted,
	}
	require.NoError(t, db.Create(msg3).Error)

	// 创建另一个用户的消息（应被用户隔离过滤）
	otherUser := CreateTestUser(t, db, "other_list_user", "Password1")
	otherSource := CreateTestSource(t, db, otherUser.ID, "Other Source", "src-other-token")
	otherMsg := &model.Message{
		SourceID:  otherSource.ID,
		ChannelID: channel.ID,
		Title:     "Other User Message",
		Content:   "Should be filtered",
		Status:    model.MessageStatusSuccess,
	}
	require.NoError(t, db.Create(otherMsg).Error)

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
			name:        "成功：分页查询消息列表（应排除已删除消息和其他用户的消息）",
			userID:      testUser.ID,
			pageReq:     &dto.PageReq{Page: 1, PageSize: 10},
			wantTotal:   2, // 只有 2 条非删除消息属于该用户
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
			svc := NewMessageService(db, logger, newMockSenderFactoryAdapter())
			list, total, err := svc.ListMessages(ctx, tt.userID, tt.pageReq)

			if tt.wantSuccess {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTotal, total, "总数应与预期一致")
				assert.Len(t, list, tt.wantCount, "返回数量应与预期一致")

				// 验证列表中不包含已删除消息
				for _, item := range list {
					assert.NotEqual(t, model.MessageStatusDeleted, item.Status, "列表中不应包含已删除消息")
					assert.NotEqual(t, "Deleted Message", item.Title, "列表中不应包含已删除的消息标题")
				}

				// 验证按创建时间倒序排列（最新的在前）
				if tt.wantCount >= 2 {
					assert.Equal(t, "Message 2", list[0].Title, "第一条消息应为最新创建的")
				}

				// 如果分页为 1 条，验证返回的是最新消息
				if tt.pageReq.PageSize == 1 {
					assert.Equal(t, "Message 2", list[0].Title, "分页限制为 1 时应返回最新消息")
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
// 辅助函数
// ============================================================================

// newMockSenderFactoryAdapter 创建适配后的 SenderFactory
func newMockSenderFactoryAdapter() *sender.SenderFactory {
	factory := sender.NewSenderFactory(nil)
	factory.Register("webhook", &sender.MockSender{SendFunc: func(ctx context.Context, config datatypes.JSON, title string, content string) error {
		return nil
	}})
	factory.Register("dingtalk", &sender.MockSender{SendFunc: func(ctx context.Context, config datatypes.JSON, title string, content string) error {
		return nil
	}})
	return factory
}

// intPtr 创建指向 int 的指针
func intPtr(v int) *int {
	return &v
}

// intPtr64 创建指向 int64 的指针
func intPtr64(v int64) *int64 {
	return &v
}

// containsSubstring 检查字符串是否包含子串
func containsSubstring(s, substr string) bool {
	return strings.Contains(s, substr)
}

// ============================================================================
// TestMessageService_FilterMessages 测试消息筛选功能
// ============================================================================

// TestMessageService_FilterMessages 测试多条件筛选消息记录功能
// 覆盖场景：按来源 ID、渠道 ID、状态、时间范围、关键词筛选
func TestMessageService_FilterMessages(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	ctx := context.Background()

	// 预创建测试用户
	testUser := CreateTestUser(t, db, "filter_messages_user", "Password1")

	// 创建两个来源
	source1 := CreateTestSource(t, db, testUser.ID, "Filter Source 1", "src-filter-1")
	source2 := CreateTestSource(t, db, testUser.ID, "Filter Source 2", "src-filter-2")

	// 创建两个渠道
	channel1 := CreateTestChannel(t, db, testUser.ID, "webhook", "Filter Channel 1")
	channel2 := CreateTestChannel(t, db, testUser.ID, "dingtalk", "Filter Channel 2")

	// 绑定来源到渠道
	BindSourceToChannel(t, db, source1.ID, channel1.ID)
	BindSourceToChannel(t, db, source1.ID, channel2.ID)
	BindSourceToChannel(t, db, source2.ID, channel1.ID)

	// 创建不同状态的消息记录
	msgs := []*model.Message{
		{
			SourceID:  source1.ID,
			ChannelID: channel1.ID,
			Title:     "Build Success",
			Content:   "Build #1 passed",
			Status:    model.MessageStatusSuccess,
		},
		{
			SourceID:  source1.ID,
			ChannelID: channel2.ID,
			Title:     "Deploy Failed",
			Content:   "Deploy #2 failed",
			Status:    model.MessageStatusFailed,
		},
		{
			SourceID:  source2.ID,
			ChannelID: channel1.ID,
			Title:     "Test Report",
			Content:   "All tests passed",
			Status:    model.MessageStatusSuccess,
		},
	}
	for _, msg := range msgs {
		require.NoError(t, db.Create(msg).Error)
	}

	// 创建已删除的消息（应被过滤）
	deletedMsg := &model.Message{
		SourceID:  source1.ID,
		ChannelID: channel1.ID,
		Title:     "Deleted Message",
		Content:   "Should not appear",
		Status:    model.MessageStatusDeleted,
	}
	require.NoError(t, db.Create(deletedMsg).Error)

	// 用于时间筛选的时间戳
	now := time.Now()
	startDate := now.Add(-1 * time.Hour).UnixMilli()
	endDate := now.Add(1 * time.Hour).UnixMilli()
	futureDate := now.Add(24 * time.Hour).UnixMilli()

	tests := []struct {
		name        string                // 测试用例名称
		userID      int64                 // 用户 ID
		filter      *dto.MessageFilterReq // 筛选请求
		wantTotal   int64                 // 期望的总数
		wantCount   int                   // 期望的返回数量
		wantErrCode int                   // 期望的错误码（0 表示无错误）
		wantSuccess bool                  // 是否期望成功
	}{
		{
			name:   "成功：按来源 ID 筛选",
			userID: testUser.ID,
			filter: &dto.MessageFilterReq{
				PageReq:  dto.PageReq{Page: 1, PageSize: 10},
				SourceID: &source1.ID,
			},
			wantTotal:   2, // source1 有 2 条非删除消息
			wantCount:   2,
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:   "成功：按渠道 ID 筛选",
			userID: testUser.ID,
			filter: &dto.MessageFilterReq{
				PageReq:   dto.PageReq{Page: 1, PageSize: 10},
				ChannelID: &channel1.ID,
			},
			wantTotal:   2, // channel1 有 2 条非删除消息
			wantCount:   2,
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:   "成功：按状态筛选（仅失败消息）",
			userID: testUser.ID,
			filter: &dto.MessageFilterReq{
				PageReq: dto.PageReq{Page: 1, PageSize: 10},
				Status:  intPtr(model.MessageStatusFailed),
			},
			wantTotal:   1,
			wantCount:   1,
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:   "成功：按时间范围筛选",
			userID: testUser.ID,
			filter: &dto.MessageFilterReq{
				PageReq:   dto.PageReq{Page: 1, PageSize: 10},
				StartDate: &startDate,
				EndDate:   &endDate,
			},
			wantTotal:   3, // 所有非删除消息都在这个时间范围内
			wantCount:   3,
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:   "成功：关键词搜索（匹配标题）",
			userID: testUser.ID,
			filter: &dto.MessageFilterReq{
				PageReq: dto.PageReq{Page: 1, PageSize: 10},
				Keyword: "Build",
			},
			wantTotal:   1,
			wantCount:   1,
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:   "成功：关键词搜索（匹配内容）",
			userID: testUser.ID,
			filter: &dto.MessageFilterReq{
				PageReq: dto.PageReq{Page: 1, PageSize: 10},
				Keyword: "tests",
			},
			wantTotal:   1,
			wantCount:   1,
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:   "成功：多条件组合筛选（来源 ID + 状态）",
			userID: testUser.ID,
			filter: &dto.MessageFilterReq{
				PageReq:  dto.PageReq{Page: 1, PageSize: 10},
				SourceID: &source1.ID,
				Status:   intPtr(model.MessageStatusSuccess),
			},
			wantTotal:   1,
			wantCount:   1,
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:   "成功：时间范围无匹配结果",
			userID: testUser.ID,
			filter: &dto.MessageFilterReq{
				PageReq:   dto.PageReq{Page: 1, PageSize: 10},
				StartDate: &futureDate,
				EndDate:   intPtr64(futureDate + 86400000),
			},
			wantTotal:   0,
			wantCount:   0,
			wantErrCode: 0,
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMessageService(db, logger, newMockSenderFactoryAdapter())
			list, total, err := svc.FilterMessages(ctx, tt.userID, tt.filter)

			if tt.wantSuccess {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTotal, total, "总数应与预期一致")
				assert.Len(t, list, tt.wantCount, "返回数量应与预期一致")

				// 验证不包含已删除消息
				for _, item := range list {
					assert.NotEqual(t, model.MessageStatusDeleted, item.Status, "筛选结果中不应包含已删除消息")
				}

				// 验证具体筛选条件
				if tt.filter.SourceID != nil {
					for _, item := range list {
						assert.Equal(t, *tt.filter.SourceID, item.SourceID, "来源 ID 应与筛选条件一致")
					}
				}
				if tt.filter.ChannelID != nil {
					for _, item := range list {
						assert.Equal(t, *tt.filter.ChannelID, item.ChannelID, "渠道 ID 应与筛选条件一致")
					}
				}
				if tt.filter.Status != nil {
					for _, item := range list {
						assert.Equal(t, *tt.filter.Status, item.Status, "状态应与筛选条件一致")
					}
				}
				if tt.filter.Keyword != "" {
					for _, item := range list {
						contains := containsSubstring(item.Title, tt.filter.Keyword) ||
							containsSubstring(item.Content, tt.filter.Keyword)
						assert.True(t, contains, "消息标题或内容应包含关键词")
					}
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
// TestMessageService_GetMessageByID 测试查询消息详情功能
// ============================================================================

// TestMessageService_GetMessageByID 测试查询单条消息详情功能
// 覆盖场景：成功查询（含来源/渠道信息）、消息不存在、消息已删除
func TestMessageService_GetMessageByID(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	ctx := context.Background()

	// 预创建测试用户、来源和渠道
	testUser := CreateTestUser(t, db, "detail_message_user", "Password1")
	channel := CreateTestChannel(t, db, testUser.ID, "webhook", "Detail Channel")
	testSource := CreateTestSource(t, db, testUser.ID, "Detail Source", "src-detail-token")
	BindSourceToChannel(t, db, testSource.ID, channel.ID)

	// 创建消息记录
	message := &model.Message{
		SourceID:  testSource.ID,
		ChannelID: channel.ID,
		Title:     "Detail Message",
		Content:   "Message content for detail",
		Status:    model.MessageStatusSuccess,
	}
	require.NoError(t, db.Create(message).Error)

	// 创建已删除的消息
	deletedMessage := &model.Message{
		SourceID:  testSource.ID,
		ChannelID: channel.ID,
		Title:     "Deleted Message",
		Content:   "Should not be found",
		Status:    model.MessageStatusDeleted,
	}
	require.NoError(t, db.Create(deletedMessage).Error)

	// 创建其他用户的消息
	otherUser := CreateTestUser(t, db, "other_detail_user", "Password1")
	otherSource := CreateTestSource(t, db, otherUser.ID, "Other Source", "src-other-detail")
	otherMessage := &model.Message{
		SourceID:  otherSource.ID,
		ChannelID: channel.ID,
		Title:     "Other User Message",
		Content:   "Should not be accessible",
		Status:    model.MessageStatusSuccess,
	}
	require.NoError(t, db.Create(otherMessage).Error)

	tests := []struct {
		name        string // 测试用例名称
		userID      int64  // 用户 ID
		messageID   int64  // 消息 ID
		wantErrCode int    // 期望的错误码（0 表示无错误）
		wantSuccess bool   // 是否期望成功
	}{
		{
			name:        "成功：查询消息详情，包含来源和渠道信息",
			userID:      testUser.ID,
			messageID:   message.ID,
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:        "失败：消息不存在，应返回 ErrNotFound",
			userID:      testUser.ID,
			messageID:   99999,
			wantErrCode: xerr.ErrNotFound.Code,
			wantSuccess: false,
		},
		{
			name:        "失败：消息已删除，应返回 ErrNotFound",
			userID:      testUser.ID,
			messageID:   deletedMessage.ID,
			wantErrCode: xerr.ErrNotFound.Code,
			wantSuccess: false,
		},
		{
			name:        "失败：查询其他用户的消息，应返回 ErrNotFound（权限隔离）",
			userID:      testUser.ID,
			messageID:   otherMessage.ID,
			wantErrCode: xerr.ErrNotFound.Code,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMessageService(db, logger, newMockSenderFactoryAdapter())
			result, err := svc.GetMessageByID(ctx, tt.userID, tt.messageID)

			if tt.wantSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, message.ID, result.ID)
				assert.Equal(t, "Detail Message", result.Title)
				assert.Equal(t, "Message content for detail", result.Content)
				assert.Equal(t, model.MessageStatusSuccess, result.Status)
				assert.Equal(t, testSource.ID, result.SourceID)
				assert.Equal(t, "Detail Source", result.SourceName)
				assert.Equal(t, channel.ID, result.ChannelID)
				assert.Equal(t, "Detail Channel", result.ChannelName)
				assert.Equal(t, "webhook", result.ChannelType)
				assert.NotZero(t, result.CreatedAt, "CreatedAt 不应为零")
				assert.NotZero(t, result.UpdatedAt, "UpdatedAt 不应为零")
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
// TestMessageService_DeleteMessage 测试删除消息功能
// ============================================================================

// TestMessageService_DeleteMessage 测试软删除消息记录功能
// 覆盖场景：成功软删除、消息已删除、消息不存在、权限不足
func TestMessageService_DeleteMessage(t *testing.T) {
	// 初始化共享测试基础设施
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	ctx := context.Background()

	// 预创建测试用户、来源和渠道
	testUser := CreateTestUser(t, db, "delete_message_user", "Password1")
	channel := CreateTestChannel(t, db, testUser.ID, "webhook", "Delete Channel")
	testSource := CreateTestSource(t, db, testUser.ID, "Delete Source", "src-delete-token")
	BindSourceToChannel(t, db, testSource.ID, channel.ID)

	// 创建待删除的消息
	message := &model.Message{
		SourceID:  testSource.ID,
		ChannelID: channel.ID,
		Title:     "Message to Delete",
		Content:   "This message will be deleted",
		Status:    model.MessageStatusSuccess,
	}
	require.NoError(t, db.Create(message).Error)

	// 创建已删除的消息
	alreadyDeleted := &model.Message{
		SourceID:  testSource.ID,
		ChannelID: channel.ID,
		Title:     "Already Deleted",
		Content:   "Should not be deletable again",
		Status:    model.MessageStatusDeleted,
	}
	require.NoError(t, db.Create(alreadyDeleted).Error)

	// 创建其他用户的消息
	otherUser := CreateTestUser(t, db, "other_delete_user", "Password1")
	otherSource := CreateTestSource(t, db, otherUser.ID, "Other Source", "src-other-delete")
	otherMessage := &model.Message{
		SourceID:  otherSource.ID,
		ChannelID: channel.ID,
		Title:     "Other User Message",
		Content:   "Should not be deletable",
		Status:    model.MessageStatusSuccess,
	}
	require.NoError(t, db.Create(otherMessage).Error)

	tests := []struct {
		name        string // 测试用例名称
		userID      int64  // 操作用户 ID
		messageID   int64  // 消息 ID
		wantErrCode int    // 期望的错误码（0 表示无错误）
		wantSuccess bool   // 是否期望成功
	}{
		{
			name:        "成功：软删除消息，状态变更为已删除",
			userID:      testUser.ID,
			messageID:   message.ID,
			wantErrCode: 0,
			wantSuccess: true,
		},
		{
			name:        "失败：消息已删除，应返回 ErrMessageAlreadyDeleted",
			userID:      testUser.ID,
			messageID:   alreadyDeleted.ID,
			wantErrCode: xerr.ErrMessageAlreadyDeleted.Code,
			wantSuccess: false,
		},
		{
			name:        "失败：消息不存在，应返回 ErrNotFound",
			userID:      testUser.ID,
			messageID:   99999,
			wantErrCode: xerr.ErrNotFound.Code,
			wantSuccess: false,
		},
		{
			name:        "失败：其他用户尝试删除，应返回 ErrNotFound（权限隔离）",
			userID:      testUser.ID,
			messageID:   otherMessage.ID,
			wantErrCode: xerr.ErrNotFound.Code,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewMessageService(db, logger, newMockSenderFactoryAdapter())
			err := svc.DeleteMessage(ctx, tt.userID, tt.messageID)

			if tt.wantSuccess {
				assert.NoError(t, err)

				// 验证数据库中消息状态已变更为已删除
				var deletedMsg model.Message
				findErr := db.Where("id = ?", tt.messageID).First(&deletedMsg).Error
				assert.NoError(t, findErr, "应能查询到消息记录")
				assert.Equal(t, model.MessageStatusDeleted, deletedMsg.Status, "消息状态应为已删除")

				// 验证删除后的消息无法再次通过 GetMessageByID 查询到（被排除）
				detail, detailErr := svc.GetMessageByID(ctx, tt.userID, tt.messageID)
				assert.Error(t, detailErr, "已删除的消息不应被查询到")
				assert.Nil(t, detail)
			} else {
				assert.Error(t, err)
				appErr, ok := err.(*xerr.AppError)
				assert.True(t, ok, "错误类型应为 *xerr.AppError")
				assert.Equal(t, tt.wantErrCode, appErr.Code, "错误码不匹配")

				// 对于非不存在的消息，验证原始状态未被修改
				if tt.messageID != 99999 {
					var originalMsg model.Message
					findErr := db.Where("id = ?", tt.messageID).First(&originalMsg).Error
					assert.NoError(t, findErr, "应能查询到消息记录")
					if tt.messageID == alreadyDeleted.ID {
						assert.Equal(t, model.MessageStatusDeleted, originalMsg.Status, "已删除消息状态不应改变")
					}
				}
			}
		})
	}
}

// TestMessageService_ListMessages_DBError 测试数据库查询错误场景
// 通过关闭数据库连接来模拟查询失败，验证 ErrMessageRecordFailed 路径
func TestMessageService_ListMessages_DBError(t *testing.T) {
	db := SetupTestDB(t)
	logger := SetupTestLogger(t)
	ctx := context.Background()

	// 关闭数据库连接，使 ListMessages 的查询阶段报错
	sqlDB, err := db.DB()
	require.NoError(t, err, "获取底层 sql.DB 失败")
	err = sqlDB.Close()
	require.NoError(t, err, "关闭数据库连接失败")

	svc := NewMessageService(db, logger, newMockSenderFactoryAdapter())
	list, total, err := svc.ListMessages(ctx, 1, &dto.PageReq{Page: 1, PageSize: 10})

	// 应该返回 ErrMessageRecordFailed
	assert.Error(t, err)
	assert.Nil(t, list)
	assert.Equal(t, int64(0), total)
	appErr, ok := err.(*xerr.AppError)
	assert.True(t, ok, "错误类型应为 *xerr.AppError")
	assert.Equal(t, xerr.ErrMessageRecordFailed.Code, appErr.Code, "应返回消息记录失败错误码")
}

// ============================================================================
// TestMessageService_PushMessage 测试推送消息功能
// ============================================================================

// pushMessageTestData 用于管理不同用例的测试数据创建
type pushMessageTestData struct {
	// setupData 在运行测试前创建所需的数据库记录
	setupData func(t *testing.T, db *gorm.DB, userID int64)
}

// TestMessageService_PushMessage 测试通过来源 Token 推送消息功能
// 覆盖场景：成功推送到多个渠道、来源不存在、无绑定渠道、发送器报错
func TestMessageService_PushMessage(t *testing.T) {
	tests := []struct {
		name           string                      // 测试用例
		sourceToken    string                      // 来源 Token
		req            *dto.PushMessageReq         // 推送请求
		setupMock      func(*sender.SenderFactory) // Mock 设置
		testData       *pushMessageTestData        // 测试数据设置（可选）
		wantErr        *xerr.AppError              // 期望的错误（nil 表示无错误）
		wantTotal      int                         // 期望的推送总数（0 表示不验证）
		wantSuccessCnt int                         // 期望的成功数
		wantFailedCnt  int                         // 期望的失败数
	}{
		{
			name:        "成功：推送到多个渠道（应返回 PushResponse，包含 success/failed 统计）",
			sourceToken: "src-push-token",
			req:         &dto.PushMessageReq{Title: "Test Title", Message: "Test Message"},
			setupMock: func(f *sender.SenderFactory) {
				f.Register("webhook", &sender.MockSender{SendFunc: func(ctx context.Context, config datatypes.JSON, title string, content string) error {
					return nil
				}})
				f.Register("dingtalk", &sender.MockSender{SendFunc: func(ctx context.Context, config datatypes.JSON, title string, content string) error {
					return nil
				}})
			},
			wantTotal:      2,
			wantSuccessCnt: 2,
			wantFailedCnt:  0,
		},
		{
			name:        "失败：来源 Token 不存在，应返回 ErrSourceNotFound",
			sourceToken: "non-existent-token",
			req:         &dto.PushMessageReq{Title: "Test Title", Message: "Test Message"},
			setupMock:   func(f *sender.SenderFactory) {},
			wantErr:     xerr.ErrSourceNotFound,
		},
		{
			name:        "失败：来源已禁用，应返回 ErrSourceNotFound",
			sourceToken: "src-disabled-push",
			req:         &dto.PushMessageReq{Title: "Test Title", Message: "Test Message"},
			setupMock:   func(f *sender.SenderFactory) {},
			wantErr:     xerr.ErrSourceNotFound,
		},
		{
			name:        "失败：来源未绑定任何渠道，应返回 ErrMessageNoChannels",
			sourceToken: "src-no-channels-push",
			req:         &dto.PushMessageReq{Title: "Test Title", Message: "Test Message"},
			setupMock:   func(f *sender.SenderFactory) {},
			wantErr:     xerr.ErrMessageNoChannels,
		},
		{
			name:        "成功：部分渠道推送成功，部分失败",
			sourceToken: "src-push-token",
			req:         &dto.PushMessageReq{Title: "Test Title", Message: "Test Message"},
			setupMock: func(f *sender.SenderFactory) {
				f.Register("webhook", &sender.MockSender{SendFunc: func(ctx context.Context, config datatypes.JSON, title string, content string) error {
					return nil
				}})
				f.Register("dingtalk", &sender.MockSender{SendFunc: func(ctx context.Context, config datatypes.JSON, title string, content string) error {
					return errors.New("dingtalk send failed")
				}})
			},
			wantTotal:      2,
			wantSuccessCnt: 1,
			wantFailedCnt:  1,
		},
		{
			name:        "成功：PushMessage 返回成功但所有渠道发送器创建失败",
			sourceToken: "src-unknown-type",
			req:         &dto.PushMessageReq{Title: "Test Title", Message: "Test Message"},
			setupMock: func(f *sender.SenderFactory) {
				// 不注册任何发送器，渠道类型也不匹配工厂内置的发送器
			},
			testData: &pushMessageTestData{
				setupData: func(t *testing.T, db *gorm.DB, userID int64) {
					t.Helper()
					src := CreateTestSource(t, db, userID, "Push Source Unknown", "src-unknown-type")
					ch1 := CreateTestChannel(t, db, userID, "unknown_type_1", "Unknown Channel 1")
					ch2 := CreateTestChannel(t, db, userID, "unknown_type_2", "Unknown Channel 2")
					BindSourceToChannel(t, db, src.ID, ch1.ID)
					BindSourceToChannel(t, db, src.ID, ch2.ID)
				},
			},
			// PushMessage 本身返回 nil，但所有渠道在结果中标记为失败
			wantTotal:      2,
			wantSuccessCnt: 0,
			wantFailedCnt:  2,
		},
	}

	// 默认测试数据设置（常规用例）
	defaultTestData := &pushMessageTestData{
		setupData: func(t *testing.T, db *gorm.DB, userID int64) {
			t.Helper()
			src := CreateTestSource(t, db, userID, "Push Source", "src-push-token")
			srcDisabled := CreateTestSource(t, db, userID, "Disabled Source", "src-disabled-push")
			db.Model(&model.Source{}).Where("id = ?", srcDisabled.ID).Update("status", model.SourceStatusDisabled)
			ch1 := CreateTestChannel(t, db, userID, "webhook", "Push Channel 1")
			ch2 := CreateTestChannel(t, db, userID, "dingtalk", "Push Channel 2")
			BindSourceToChannel(t, db, src.ID, ch1.ID)
			BindSourceToChannel(t, db, src.ID, ch2.ID)
			_ = CreateTestSource(t, db, userID, "No Channels Source", "src-no-channels-push")
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 每个子用例使用独立数据库
			testDB := SetupTestDB(t)
			testLogger := SetupTestLogger(t)
			testCtx := context.Background()

			// 创建该子用例独立的测试数据
			subUser := CreateTestUser(t, testDB, "push_user_sub", "Password1")

			// 使用用例自定义的测试数据设置，或默认设置
			if tt.testData != nil {
				tt.testData.setupData(t, testDB, subUser.ID)
			} else {
				defaultTestData.setupData(t, testDB, subUser.ID)
			}

			// 设置 Mock
			testFactory := sender.NewSenderFactory(testLogger)
			tt.setupMock(testFactory)

			svc := NewMessageService(testDB, testLogger, testFactory)
			result, err := svc.PushMessage(testCtx, tt.sourceToken, tt.req)

			if tt.wantErr != nil {
				// 验证错误场景
				assert.Error(t, err)
				assert.Nil(t, result)
				appErr, ok := err.(*xerr.AppError)
				assert.True(t, ok, "错误类型应为 *xerr.AppError")
				assert.Equal(t, tt.wantErr.Code, appErr.Code, "错误码不匹配")
			} else {
				// 验证成功场景（PushMessage 返回 nil，结果在 PushResponse 中）
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantTotal, result.Total, "推送总数应与预期一致")
				assert.Equal(t, tt.wantSuccessCnt, result.Success, "成功数应与预期一致")
				assert.Equal(t, tt.wantFailedCnt, result.Failed, "失败数应与预期一致")
				assert.Len(t, result.Results, tt.wantTotal, "推送结果数量应与总数一致")

				// 验证每个推送结果的结构
				for _, r := range result.Results {
					assert.NotZero(t, r.ChannelID, "渠道 ID 不应为零")
					assert.NotEmpty(t, r.ChannelName, "渠道名称不应为空")
				}

				// 对于全部失败的用例，验证错误信息
				if tt.wantSuccessCnt == 0 && tt.wantFailedCnt > 0 {
					for _, r := range result.Results {
						assert.False(t, r.Success, "渠道推送结果应为失败")
						assert.Contains(t, r.Error, "不支持的渠道类型", "错误信息应包含不支持的渠道类型")
					}
				}
			}
		})
	}
}
