package domain

// SendResponse 发送响应
type SendResponse struct {
	NotificationID uint64     // 通知ID
	Status         SendStatus // 发送状态
}

// BatchSendResponse 批量发送响应
type BatchSendResponse struct {
	Results []SendResponse // 所有结果
}

// BatchSendAsyncResponse 批量异步发送响应
type BatchSendAsyncResponse struct {
	NotificationIDs []uint64 // 生成的通知ID列表
}
