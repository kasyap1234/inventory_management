package services

import (
	"context"
	"fmt"
	"time"

	"agromart2/internal/common"
	"agromart2/internal/models"
	"agromart2/internal/repositories"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
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
	deviceTokenRepo  repositories.DeviceTokenRepository
	fcmClient        *messaging.Client
	fcmEnabled       bool
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
	deviceTokenRepo repositories.DeviceTokenRepository,
	firebaseApp *firebase.App,
	fcmEnabled bool,
) NotificationDeliveryService {
	var fcmClient *messaging.Client
	if firebaseApp != nil && fcmEnabled {
		var err error
		fcmClient, err = firebaseApp.Messaging(context.Background())
		if err != nil {
			logger.Error("Failed to initialize FCM client", err)
			fcmEnabled = false
		}
	}

	return &notificationDeliveryService{
		deliveryRepo:     deliveryRepo,
		notificationRepo: nil, // Can be injected if needed for enhanced functionality
		templateService:  templateService,
		webhookService:   webhookService,
		alertService:     alertService,
		logger:           logger,
		notificationSvc:  notificationSvc,
		userRepo:         userRepo,
		deviceTokenRepo:  deviceTokenRepo,
		fcmClient:        fcmClient,
		fcmEnabled:       fcmEnabled,
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
	// Check if FCM is enabled
	if !s.fcmEnabled || s.fcmClient == nil {
		s.logger.InfoWithContext(ctx, "FCM not enabled, push notification would be delivered in production", map[string]interface{}{
			"delivery_id": delivery.ID,
			"recipient":   delivery.Recipient,
			"title":       notification.Title,
			"message":     notification.Message,
		})
		return nil // Return success in dev/test mode
	}

	// Check if device token repository is available
	if s.deviceTokenRepo == nil {
		return fmt.Errorf("device token repository not configured")
	}

	// Get active device tokens for the user
	tokens, err := s.deviceTokenRepo.GetActiveTokensByUser(ctx, notification.TenantID, notification.UserID)
	if err != nil {
		return fmt.Errorf("failed to get device tokens: %w", err)
	}

	if len(tokens) == 0 {
		return fmt.Errorf("no active device tokens found for user")
	}

	// Track successful deliveries
	successCount := 0

	// Prepare and send FCM message to each device
	for _, token := range tokens {
		message := &messaging.Message{
			Token: token.DeviceToken,
			Notification: &messaging.Notification{
				Title: notification.Title,
				Body:  notification.Message,
			},
			Data: map[string]string{
				"notification_id": notification.ID.String(),
				"tenant_id":       notification.TenantID.String(),
				"priority":        notification.Priority,
			},
		}

		// Add event type if available
		if notification.EventType != nil {
			message.Data["event_type"] = *notification.EventType
		}

		// Platform-specific configurations
		if token.DeviceType == "android" {
			message.Android = &messaging.AndroidConfig{
				Priority: "high",
				Notification: &messaging.AndroidNotification{
					Sound:       "default",
					Color:       "#007bff",
					ChannelID:   "default",
					ClickAction: "FLUTTER_NOTIFICATION_CLICK",
				},
			}
		} else if token.DeviceType == "ios" {
			message.APNS = &messaging.APNSConfig{
				Headers: map[string]string{
					"apns-priority": "10",
				},
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{
						Sound: "default",
						Badge: s.getBadgeCount(ctx, notification.UserID),
						Alert: &messaging.ApsAlert{
							Title: notification.Title,
							Body:  notification.Message,
						},
					},
				},
			}
		}

		// Send the message
		response, err := s.fcmClient.Send(ctx, message)
		if err != nil {
			s.logger.ErrorWithContext(ctx, "Failed to send push notification", err, map[string]interface{}{
				"device_token": token.DeviceToken,
				"device_type":  token.DeviceType,
				"user_id":      notification.UserID,
			})

			// If token is invalid or unregistered, deactivate it
			if messaging.IsInvalidArgument(err) || messaging.IsUnregistered(err) {
				if deactivateErr := s.deviceTokenRepo.DeactivateToken(ctx, notification.TenantID, token.DeviceToken); deactivateErr != nil {
					s.logger.WarnWithContext(ctx, "Failed to deactivate invalid token", map[string]interface{}{
						"device_token": token.DeviceToken,
						"error":        deactivateErr.Error(),
					})
				}
			}
			continue
		}

		s.logger.InfoWithContext(ctx, "Push notification sent successfully", map[string]interface{}{
			"message_id":   response,
			"device_token": token.DeviceToken,
			"device_type":  token.DeviceType,
			"user_id":      notification.UserID,
		})

		// Update last used timestamp
		if err := s.deviceTokenRepo.UpdateLastUsed(ctx, notification.TenantID, token.DeviceToken); err != nil {
			s.logger.WarnWithContext(ctx, "Failed to update token last used", map[string]interface{}{
				"device_token": token.DeviceToken,
				"error":        err.Error(),
			})
		}

		successCount++
	}

	// Return error if no devices were successfully notified
	if successCount == 0 {
		return fmt.Errorf("failed to deliver push notification to any device")
	}

	return nil
}

// getBadgeCount returns the unread notification count for badge display
func (s *notificationDeliveryService) getBadgeCount(ctx context.Context, userID uuid.UUID) *int {
	// This is a placeholder - in production, you would query the actual unread count
	// from the notifications table
	badge := 0
	return &badge
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
