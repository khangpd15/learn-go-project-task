

package services

import (
	"errors"

	"task_api/internal/entities"
	"task_api/internal/repositories"
)

var (
	ErrInvalidNotificationID = errors.New("invalid notification id")
	ErrInvalidReceiverID     = errors.New("invalid receiver id")
	ErrInvalidSenderID       = errors.New("invalid sender id")
	ErrNotificationTypeEmpty = errors.New("notification type is required")
	ErrNotificationTitleEmpty = errors.New("notification title is required")
	ErrNotificationMessageEmpty = errors.New("notification message is required")
)

type NotificationService interface {
	CreateNotification(notification entities.Notification) (entities.Notification, error)
	GetNotificationsByReceiverID(receiverID int) ([]entities.Notification, error)
	GetUnreadNotifications(receiverID int) ([]entities.Notification, error)
	MarkAsRead(notificationID int) error
	MarkAsUnread(notificationID int) error
	MarkAllAsRead(receiverID int) error
}

type NotificationServiceImpl struct {
	notificationRepo repositories.NotificationRepository
}

func NewNotificationService(notificationRepo repositories.NotificationRepository) *NotificationServiceImpl {
	return &NotificationServiceImpl{
		notificationRepo: notificationRepo,
	}
}

func (s *NotificationServiceImpl) CreateNotification(notification entities.Notification) (entities.Notification, error) {
	if notification.SenderID <= 0 {
		return entities.Notification{}, ErrInvalidSenderID
	}

	if notification.ReceiverID != nil && *notification.ReceiverID <= 0 {
		return entities.Notification{}, ErrInvalidReceiverID
	}

	if notification.Type == "" {
		return entities.Notification{}, ErrNotificationTypeEmpty
	}

	if notification.Title == "" {
		return entities.Notification{}, ErrNotificationTitleEmpty
	}

	if notification.Message == "" {
		return entities.Notification{}, ErrNotificationMessageEmpty
	}

	return s.notificationRepo.CreateNotification(notification)
}

func (s *NotificationServiceImpl) GetNotificationsByReceiverID(receiverID int) ([]entities.Notification, error) {
	if receiverID <= 0 {
		return nil, ErrInvalidReceiverID
	}

	return s.notificationRepo.GetNotificationsByReceiverID(receiverID)
}

func (s *NotificationServiceImpl) GetUnreadNotifications(receiverID int) ([]entities.Notification, error) {
	if receiverID <= 0 {
		return nil, ErrInvalidReceiverID
	}

	return s.notificationRepo.GetUnreadNotifications(receiverID)
}

func (s *NotificationServiceImpl) MarkAsRead(notificationID int) error {
	if notificationID <= 0 {
		return ErrInvalidNotificationID
	}

	return s.notificationRepo.UpdateNotificationReadStatus(notificationID, true)
}

func (s *NotificationServiceImpl) MarkAsUnread(notificationID int) error {
	if notificationID <= 0 {
		return ErrInvalidNotificationID
	}

	return s.notificationRepo.UpdateNotificationReadStatus(notificationID, false)
}

func (s *NotificationServiceImpl) MarkAllAsRead(receiverID int) error {
	if receiverID <= 0 {
		return ErrInvalidReceiverID
	}

	return s.notificationRepo.MarkAllAsRead(receiverID)
}