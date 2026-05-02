package model

type SourceChannel struct {
	SourceID  int64 `gorm:"primaryKey" json:"source_id"`
	ChannelID int64 `gorm:"primaryKey" json:"channel_id"`
}

func (SourceChannel) TableName() string {
	return "source_channels"
}
