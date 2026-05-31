package repositories

import (
	"gorm.io/gorm"
	"task_api/internal/entities"
	"time"
)

type NotificationRepository interface {
	CreateNotification(notification entities.Notification) (entities.Notification, error)

	GetNotificationsByReceiverID(receiverID int) ([]entities.Notification, error)

	GetUnreadNotifications(receiverID int,) ([]entities.Notification, error)

	UpdateNotificationReadStatus(notificationID int,isRead bool,) error
    
    GetNotificationByID(notificationID int) (entities.Notification, error)

	MarkAllAsRead(receiverID int) error
}

type NotificationRepositoryImpl struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepositoryImpl {
	return &NotificationRepositoryImpl{db: db}
}

func (r *NotificationRepositoryImpl) CreateNotification(notification entities.Notification) (entities.Notification, error) {
	err := r.db.Create(&notification).Error
	if err != nil {
		return entities.Notification{}, err
	}
	return notification, nil
}

func (r *NotificationRepositoryImpl) GetNotificationsByReceiverID(receiverID int) ([]entities.Notification, error) {
	var notifications []entities.Notification
	err := r.db.Where("receiver_id = ?", receiverID).Order("created_at desc").Find(&notifications).Error
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *NotificationRepositoryImpl) GetUnreadNotifications(receiverID int) ([]entities.Notification, error) {
	var notifications []entities.Notification
	err := r.db.Where("receiver_id = ? AND is_read = ?", receiverID, false).Order("created_at desc").Find(&notifications).Error
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *NotificationRepositoryImpl) UpdateNotificationReadStatus(notificationID int, isRead bool) error {
	updates := map[string]interface{}{
		"is_read": isRead,
	}

	if isRead {
		updates["read_at"] = time.Now()
	} else {
		updates["read_at"] = nil
	}

	return r.db.
		Model(&entities.Notification{}).
		Where("id = ?", notificationID).
		Updates(updates).
		Error
}

func (r *NotificationRepositoryImpl) MarkAllAsRead(receiverID int) error {
	return r.db.
		Model(&entities.Notification{}).
		Where("receiver_id = ? AND is_read = ?", receiverID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": time.Now(),
		}).
		Error
}

func (r *NotificationRepositoryImpl) GetNotificationByID(notificationID int) (entities.Notification, error) {
	var notification entities.Notification

	err := r.db.
		Where("id = ?", notificationID).
		First(&notification).
		Error

	if err != nil {
		return entities.Notification{}, err
	}

	return notification, nil
}
