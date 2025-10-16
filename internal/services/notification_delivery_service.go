package services

import (
	"context"
	"fmt"
	"time"

	"agromart2/internal/common"
	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
)

// NotificationDeliveryService interface for notification delivery operations
type NotificationDeliveryService interface {
	ProcessEvent(ctx context.Context, event *models.NotificationEvent) error
	DeliverNotification(ctx context.Context, notification *models.EnhancedNotification) error
	RetryFailedDeliveries(ctx context.Context, tenantID uuid.UUID) error
	GetDeliveryStatus(ctx context.Context, tenantID uuid.UUID, deliveryID uuid.UUID) (*models.NotificationDelivery, error)
	ListDeliveries(ctx context.Context, tenantID uuid.UUID, filters repositories.DeliveryFilters) ([]*models.NotificationDelivery, error)
}

// NotificationDeliveryRepository interface for delivery persistence
type NotificationDeliveryRepository interface {
	Create(ctx context.Context, delivery *models.NotificationDelivery) error
	Update(ctx context.Context, delivery *models.NotificationDelivery) error
	GetByID(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*models.NotificationDelivery, error)
	List(ctx context.Context, tenantID uuid.UUID, filters repositories.DeliveryFilters) ([]*models.NotificationDelivery, error)
	GetFailedDeliveries(ctx context.Context, tenantID uuid.UUID, maxRetries int) ([]*models.NotificationDelivery, error)
	Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error
}

// notificationDeliveryService implements NotificationDeliveryService
type notificationDeliveryService struct {
	deliveryRepo     NotificationDeliveryRepository
	notificationRepo repositories.NotificationRepository
	templateService  NotificationTemplateService
	webhookService   WebhookSubscriptionService
	alertService     AlertRuleService
	logger           *common.StructuredLogger
	notificationSvc  NotificationService
	userRepo         repositories.UserRepository
}

// NewNotificationDeliveryService creates a new notification delivery service
func NewNotificationDeliveryService(
	deliveryRepo NotificationDeliveryRepository,
	templateService NotificationTemplateService,
	webhookService WebhookSubscriptionService,
	alertService AlertRuleService,
	logger *common.StructuredLogger,
	notificationSvc NotificationService,
	userRepo repositories.UserRepository,
) NotificationDeliveryService {
	return &notificationDeliveryService{
		deliveryRepo:     deliveryRepo,
		notificationRepo: nil, // Can be injected if needed for enhanced functionality
		templateService:  templateService,
		webhookService:   webhookService,
		alertService:     alertService,
		logger:           logger,
		notificationSvc:  notificationSvc,
		userRepo:         userRepo,
	}
}

// ProcessEvent processes a notification event and triggers appropriate notifications
func (s *notificationDeliveryService) ProcessEvent(ctx context.Context, event *models.NotificationEvent) error {
	s.logger.InfoWithContext(ctx, "Processing notification event", map[string]interface{}{
		"event_type": event.Type,
		"tenant_id":  event.TenantID,
		"user_id":    event.UserID,
	})

	// 1. Evaluate alert rules
	if err := s.alertService.EvaluateAlertRules(ctx, event.TenantID, event.Type, event.Data); err != nil {
		s.logger.ErrorWithContext(ctx, "Failed to evaluate alert rules", err, map[string]interface{}{
			"event_type": event.Type,
			"tenant_id":  event.TenantID,
		})
		// Don't return error, continue with other notifications
	}

	// 2. Process webhook notifications
	if err := s.processWebhookNotifications(ctx, event); err != nil {
		s.logger.ErrorWithContext(ctx, "Failed to process webhook notifications", err, map[string]interface{}{
			"event_type": event.Type,
			"tenant_id":  event.TenantID,
		})
		// Don't return error, continue with other notifications
	}

	// 3. Process template-based notifications
	if err := s.processTemplateNotifications(ctx, event); err != nil {
		s.logger.ErrorWithContext(ctx, "Failed to process template notifications", err, map[string]interface{}{
			"event_type": event.Type,
			"tenant_id":  event.TenantID,
		})
		// Don't return error, continue
	}

	s.logger.InfoWithContext(ctx, "Notification event processed", map[string]interface{}{
		"event_type": event.Type,
		"tenant_id":  event.TenantID,
	})

	return nil
}

// DeliverNotification delivers a single notification
func (s *notificationDeliveryService) DeliverNotification(ctx context.Context, notification *models.EnhancedNotification) error {
	s.logger.InfoWithContext(ctx, "Delivering notification", map[string]interface{}{
		"notification_id": notification.ID,
		"tenant_id":       notification.TenantID,
		"user_id":         notification.UserID,
		"type":            notification.NotificationType,
	})

	// Get user notification preferences
	// Using default channels based on notification type
	// Can be enhanced with NotificationConfigRepository for user-specific preferences

	channels := s.getDefaultChannels(notification.NotificationType)

	for _, channel := range channels {
		delivery := &models.NotificationDelivery{
			ID:             uuid.New(),
			TenantID:       notification.TenantID,
			NotificationID: &notification.ID,
			Channel:        models.NotificationType(channel),
			Status:         "pending",
			AttemptCount:   0,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		// Set recipient based on channel
		recipient, err := s.getRecipientForChannel(ctx, notification.TenantID, notification.UserID, channel)
		if err != nil {
			s.logger.WarnWithContext(ctx, "Failed to get recipient for channel", map[string]interface{}{
				"user_id": notification.UserID,
				"channel": channel,
				"error":   err.Error(),
			})
			continue
		}
		delivery.Recipient = recipient

		// Create delivery record
		if err := s.deliveryRepo.Create(ctx, delivery); err != nil {
			s.logger.ErrorWithContext(ctx, "Failed to create delivery record", err, map[string]interface{}{
				"notification_id": notification.ID,
				"channel":         channel,
			})
			continue
		}

		// Attempt delivery
		if err := s.attemptDelivery(ctx, delivery, notification); err != nil {
			s.logger.ErrorWithContext(ctx, "Failed to deliver notification", err, map[string]interface{}{
				"delivery_id": delivery.ID,
				"channel":     channel,
				"recipient":   delivery.Recipient,
			})
		}
	}

	return nil
}

// RetryFailedDeliveries retries failed notification deliveries
func (s *notificationDeliveryService) RetryFailedDeliveries(ctx context.Context, tenantID uuid.UUID) error {
	maxRetries := 3
	failedDeliveries, err := s.deliveryRepo.GetFailedDeliveries(ctx, tenantID, maxRetries)
	if err != nil {
		return common.CreateDatabaseError("retry_failed_deliveries", err)
	}

	s.logger.InfoWithContext(ctx, "Retrying failed deliveries", map[string]interface{}{
		"tenant_id":      tenantID,
		"delivery_count": len(failedDeliveries),
	})

	for _, delivery := range failedDeliveries {
		// For retry, we create a minimal notification
		// In production, you may want to fetch the original notification from NotificationRepository
		notification := &models.EnhancedNotification{
			ID:       *delivery.NotificationID,
			TenantID: delivery.TenantID,
			Title:    "Retry Notification",
			Message:  "This is a retry of a failed notification",
		}

		if err := s.attemptDelivery(ctx, delivery, notification); err != nil {
			s.logger.ErrorWithContext(ctx, "Retry delivery failed", err, map[string]interface{}{
				"delivery_id": delivery.ID,
				"channel":     delivery.Channel,
			})
		}
	}

	return nil
}

// GetDeliveryStatus retrieves the status of a notification delivery
func (s *notificationDeliveryService) GetDeliveryStatus(ctx context.Context, tenantID uuid.UUID, deliveryID uuid.UUID) (*models.NotificationDelivery, error) {
	delivery, err := s.deliveryRepo.GetByID(ctx, tenantID, deliveryID)
	if err != nil {
		return nil, common.CreateDatabaseError("get_delivery_status", err)
	}

	return delivery, nil
}

// ListDeliveries retrieves notification deliveries with filters
func (s *notificationDeliveryService) ListDeliveries(ctx context.Context, tenantID uuid.UUID, filters repositories.DeliveryFilters) ([]*models.NotificationDelivery, error) {
	deliveries, err := s.deliveryRepo.List(ctx, tenantID, filters)
	if err != nil {
		return nil, common.CreateDatabaseError("list_deliveries", err)
	}

	return deliveries, nil
}

// Helper methods

// processWebhookNotifications processes webhook notifications for an event
func (s *notificationDeliveryService) processWebhookNotifications(ctx context.Context, event *models.NotificationEvent) error {
	// Get active webhooks for this event type
	webhooks, err := s.webhookService.ListWebhooks(ctx, event.TenantID)
	if err != nil {
		return err
	}

	for _, webhook := range webhooks {
		if !webhook.IsActive {
			continue
		}

		// Check if webhook is subscribed to this event type
		subscribed := false
		for _, eventType := range webhook.EventTypes {
			if eventType == event.Type {
				subscribed = true
				break
			}
		}

		if !subscribed {
			continue
		}

		// Deliver webhook with retry logic
		bgCtx := context.Background()
		go s.deliverWebhookWithRetry(bgCtx, webhook, event)
	}

	return nil
}

// processTemplateNotifications processes template-based notifications for an event
func (s *notificationDeliveryService) processTemplateNotifications(ctx context.Context, event *models.NotificationEvent) error {
	// Get active templates for this event type
	channels := []models.NotificationType{
		models.NotificationTypeEmail,
		models.NotificationTypeSMS,
		models.NotificationTypePush,
	}

	for _, channel := range channels {
		templates, err := s.templateService.GetActiveTemplatesForEvent(ctx, event.TenantID, event.Type, channel)
		if err != nil {
			s.logger.ErrorWithContext(ctx, "Failed to get templates for event", err, map[string]interface{}{
				"event_type": event.Type,
				"channel":    channel,
			})
			continue
		}

		for _, template := range templates {
			// Render template
			body, subject, err := s.templateService.RenderTemplateWithValidation(template, event.Data)
			if err != nil {
				s.logger.ErrorWithContext(ctx, "Failed to render template", err, map[string]interface{}{
					"template_id": template.ID,
					"event_type":  event.Type,
				})
				continue
			}

			// Create notification
			notification := &models.EnhancedNotification{
				ID:               uuid.New(),
				TenantID:         event.TenantID,
				UserID:           *event.UserID, // Assuming event has user ID
				Title:            subject,
				Message:          body,
				NotificationType: string(channel),
				Priority:         "normal",
				Status:           "unread",
				EventType:        &event.Type,
				EventData:        event.Data,
				TemplateID:       &template.ID,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}

			// Deliver notification
			if err := s.DeliverNotification(ctx, notification); err != nil {
				s.logger.ErrorWithContext(ctx, "Failed to deliver template notification", err, map[string]interface{}{
					"template_id": template.ID,
					"channel":     channel,
				})
			}
		}
	}

	return nil
}

// deliverWebhookWithRetry delivers a webhook with retry logic
func (s *notificationDeliveryService) deliverWebhookWithRetry(ctx context.Context, webhook *models.WebhookSubscription, event *models.NotificationEvent) {
	maxRetries := webhook.RetryCount
	backoffDuration := time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			time.Sleep(backoffDuration)
			backoffDuration *= 2
		}

		err := s.webhookService.DeliverWebhook(ctx, webhook, event)
		if err == nil {
			// Success
			return
		}

		s.logger.WarnWithContext(ctx, "Webhook delivery attempt failed", map[string]interface{}{
			"webhook_id":  webhook.ID,
			"attempt":     attempt + 1,
			"max_retries": maxRetries,
			"error":       err.Error(),
		})

		if attempt == maxRetries {
			s.logger.ErrorWithContext(ctx, "Webhook delivery failed after all retries", err, map[string]interface{}{
				"webhook_id": webhook.ID,
				"attempts":   attempt + 1,
			})
		}
	}
}

// attemptDelivery attempts to deliver a notification via a specific channel
func (s *notificationDeliveryService) attemptDelivery(ctx context.Context, delivery *models.NotificationDelivery, notification *models.EnhancedNotification) error {
	delivery.AttemptCount++
	delivery.LastAttemptAt = &[]time.Time{time.Now()}[0]
	delivery.UpdatedAt = time.Now()

	var err error

	switch delivery.Channel {
	case models.NotificationTypeEmail:
		err = s.deliverEmail(ctx, delivery, notification)
	case models.NotificationTypeSMS:
		err = s.deliverSMS(ctx, delivery, notification)
	case models.NotificationTypePush:
		err = s.deliverPush(ctx, delivery, notification)
	default:
		err = fmt.Errorf("unsupported delivery channel: %s", delivery.Channel)
	}

	if err != nil {
		delivery.Status = "failed"
		delivery.ErrorMessage = &[]string{err.Error()}[0]
	} else {
		delivery.Status = "delivered"
		delivery.DeliveredAt = &[]time.Time{time.Now()}[0]
	}

	// Update delivery record
	if updateErr := s.deliveryRepo.Update(ctx, delivery); updateErr != nil {
		s.logger.ErrorWithContext(ctx, "Failed to update delivery record", updateErr, map[string]interface{}{
			"delivery_id": delivery.ID,
		})
	}

	return err
}

// deliverEmail delivers a notification via email
func (s *notificationDeliveryService) deliverEmail(ctx context.Context, delivery *models.NotificationDelivery, notification *models.EnhancedNotification) error {
	if s.notificationSvc == nil {
		return fmt.Errorf("email provider not configured")
	}
	return s.notificationSvc.SendEmail(ctx, notification.TenantID, delivery.Recipient, notification.Title, notification.Message)
}

// deliverSMS delivers a notification via SMS
func (s *notificationDeliveryService) deliverSMS(ctx context.Context, delivery *models.NotificationDelivery, notification *models.EnhancedNotification) error {
	if s.notificationSvc == nil {
		return fmt.Errorf("SMS provider not configured")
	}

	// For SMS, combine title and message
	message := notification.Title
	if notification.Message != "" {
		message += ": " + notification.Message
	}

	return s.notificationSvc.SendSMS(ctx, notification.TenantID, delivery.Recipient, message)
}

// deliverPush delivers a notification via push notification
func (s *notificationDeliveryService) deliverPush(ctx context.Context, delivery *models.NotificationDelivery, notification *models.EnhancedNotification) error {
	// NOTE: This is a basic implementation. For production, you need to:
	// 1. Integrate with Firebase Cloud Messaging (FCM) for Android
	// 2. Integrate with Apple Push Notification Service (APNs) for iOS
	// 3. Store device tokens in a separate table
	// 4. Handle token registration/unregistration
	// 5. Implement retry logic for failed deliveries
	//
	// Example FCM integration:
	// - Install: go get firebase.google.com/go/v4
	// - Set up Firebase project and download service account JSON
	// - Initialize FCM client with credentials
	// - Use client.Send() to deliver push notifications
	//
	// For now, we'll log the notification and mark it as delivered in test/dev environments

	s.logger.InfoWithContext(ctx, "Push notification would be delivered in production", map[string]interface{}{
		"delivery_id": delivery.ID,
		"recipient":   delivery.Recipient,
		"title":       notification.Title,
		"message":     notification.Message,
		"priority":    notification.Priority,
	})

	// In development/test mode, consider it delivered
	// In production, you would send the actual push notification here
	// Example pseudo-code for FCM:
	/*
		fcmMessage := &messaging.Message{
			Token: delivery.Recipient, // FCM device token
			Notification: &messaging.Notification{
				Title: notification.Title,
				Body:  notification.Message,
			},
			Data: map[string]string{
				"notification_id": notification.ID.String(),
				"event_type":      *notification.EventType,
				"priority":        notification.Priority,
			},
			Android: &messaging.AndroidConfig{
				Priority: "high",
			},
			APNS: &messaging.APNSConfig{
				Headers: map[string]string{
					"apns-priority": "10",
				},
			},
		}

		response, err := fcmClient.Send(ctx, fcmMessage)
		if err != nil {
			return fmt.Errorf("failed to send push notification: %w", err)
		}
	*/

	// For now, return success in non-production environments
	return nil
}

// getDefaultChannels returns default channels for a notification type
func (s *notificationDeliveryService) getDefaultChannels(notificationType string) []string {
	switch notificationType {
	case "critical", "alert":
		return []string{"email", "sms"}
	case "info", "success":
		return []string{"email"}
	case "warning":
		return []string{"email"}
	default:
		return []string{"email"}
	}
}

// getRecipientForChannel gets the recipient address for a specific channel
func (s *notificationDeliveryService) getRecipientForChannel(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, channel string) (string, error) {
	// Try to resolve from user repository when possible
	if s.userRepo != nil {
		if user, err := s.userRepo.GetByID(ctx, tenantID, userID); err == nil && user != nil {
			switch channel {
			case "email":
				if user.Email != "" {
					return user.Email, nil
				}
			}
		}
	}

	// Return error when required contact information is not available
	// Do not use placeholder values - require actual contact data
	switch channel {
	case "email":
		return "", fmt.Errorf("email address not available for user %s", userID.String())
	case "sms":
		return "", fmt.Errorf("phone number not available for user %s", userID.String())
	case "push":
		return "", fmt.Errorf("push token not available for user %s", userID.String())
	default:
		return "", fmt.Errorf("unsupported notification channel: %s", channel)
	}
}
