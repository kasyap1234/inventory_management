# Push Notifications Implementation Guide

## Overview

The push notification system is currently implemented as a stub. To enable full push notification functionality in production, you need to integrate with Firebase Cloud Messaging (FCM) and/or Apple Push Notification Service (APNs).

## Implementation Steps

### 1. Set up Firebase Cloud Messaging (FCM)

1. **Create a Firebase project:**
   - Go to [Firebase Console](https://console.firebase.google.com/)
   - Create a new project or use an existing one
   - Enable Cloud Messaging in the project settings

2. **Download service account credentials:**
   - Go to Project Settings > Service Accounts
   - Click "Generate New Private Key"
   - Save the JSON file securely (DO NOT commit to version control)

3. **Install Firebase Admin SDK:**
   ```bash
   go get firebase.google.com/go/v4
   ```

4. **Initialize FCM client in your application:**
   ```go
   import (
       firebase "firebase.google.com/go/v4"
       "firebase.google.com/go/v4/messaging"
       "google.golang.org/api/option"
   )

   // In your initialization code
   opt := option.WithCredentialsFile("path/to/serviceAccountKey.json")
   app, err := firebase.NewApp(context.Background(), nil, opt)
   if err != nil {
       log.Fatalf("error initializing app: %v", err)
   }

   fcmClient, err := app.Messaging(context.Background())
   if err != nil {
       log.Fatalf("error getting Messaging client: %v", err)
   }
   ```

### 2. Set up Device Token Management

Create a new table to store device tokens:

```sql
CREATE TABLE IF NOT EXISTS device_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_token VARCHAR(500) NOT NULL,
    device_type VARCHAR(50) NOT NULL CHECK (device_type IN ('android', 'ios', 'web')),
    device_name VARCHAR(255),
    app_version VARCHAR(50),
    is_active BOOLEAN DEFAULT true,
    last_used_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (tenant_id, user_id, device_token)
);

CREATE INDEX idx_device_tokens_user ON device_tokens (tenant_id, user_id, is_active);
CREATE INDEX idx_device_tokens_token ON device_tokens (device_token) WHERE is_active = true;
```

### 3. Create Device Token Repository

```go
// internal/repositories/device_token_repo.go
package repositories

import (
    "context"
    "github.com/google/uuid"
    "agromart2/internal/models"
)

type DeviceTokenRepository interface {
    RegisterToken(ctx context.Context, token *models.DeviceToken) error
    GetTokensByUser(ctx context.Context, tenantID, userID uuid.UUID) ([]*models.DeviceToken, error)
    GetActiveTokensByUser(ctx context.Context, tenantID, userID uuid.UUID) ([]*models.DeviceToken, error)
    DeactivateToken(ctx context.Context, tenantID uuid.UUID, deviceToken string) error
    UpdateLastUsed(ctx context.Context, tenantID uuid.UUID, deviceToken string) error
}
```

### 4. Update the Push Notification Delivery Method

Replace the stub implementation in `notification_delivery_service.go`:

```go
func (s *notificationDeliveryService) deliverPush(ctx context.Context, delivery *models.NotificationDelivery, notification *models.EnhancedNotification) error {
    // Get active device tokens for the user
    tokens, err := s.deviceTokenRepo.GetActiveTokensByUser(ctx, notification.TenantID, notification.UserID)
    if err != nil {
        return fmt.Errorf("failed to get device tokens: %w", err)
    }

    if len(tokens) == 0 {
        return fmt.Errorf("no active device tokens found for user")
    }

    // Prepare FCM message
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
                "event_type":      safeString(notification.EventType),
                "priority":        notification.Priority,
            },
        }

        // Platform-specific configurations
        if token.DeviceType == "android" {
            message.Android = &messaging.AndroidConfig{
                Priority: "high",
                Notification: &messaging.AndroidNotification{
                    Sound: "default",
                    Color: "#007bff",
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
                        Badge: getBadgeCount(ctx, notification.UserID),
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
            })
            
            // If token is invalid, deactivate it
            if messaging.IsInvalidArgument(err) || messaging.IsUnregistered(err) {
                s.deviceTokenRepo.DeactivateToken(ctx, notification.TenantID, token.DeviceToken)
            }
            continue
        }

        s.logger.InfoWithContext(ctx, "Push notification sent successfully", map[string]interface{}{
            "message_id":   response,
            "device_token": token.DeviceToken,
        })

        // Update last used timestamp
        s.deviceTokenRepo.UpdateLastUsed(ctx, notification.TenantID, token.DeviceToken)
    }

    return nil
}

func safeString(s *string) string {
    if s == nil {
        return ""
    }
    return *s
}
```

### 5. Client-Side Integration

#### Android (Kotlin/Java)
```kotlin
// Add dependencies in build.gradle
implementation 'com.google.firebase:firebase-messaging:23.0.0'

// Register for push notifications
FirebaseMessaging.getInstance().token.addOnCompleteListener { task ->
    if (task.isSuccessful) {
        val token = task.result
        // Send token to your backend
        registerDeviceToken(token, "android")
    }
}

// Handle incoming notifications
class MyFirebaseMessagingService : FirebaseMessagingService() {
    override fun onMessageReceived(remoteMessage: RemoteMessage) {
        // Handle notification
        val title = remoteMessage.notification?.title
        val body = remoteMessage.notification?.body
        val data = remoteMessage.data
        
        showNotification(title, body, data)
    }
}
```

#### iOS (Swift)
```swift
// AppDelegate.swift
import UserNotifications
import FirebaseMessaging

func application(_ application: UIApplication, 
                 didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {
    // Request notification permissions
    UNUserNotificationCenter.current().requestAuthorization(options: [.alert, .sound, .badge]) { granted, error in
        if granted {
            DispatchQueue.main.async {
                application.registerForRemoteNotifications()
            }
        }
    }
    
    return true
}

func application(_ application: UIApplication, 
                 didRegisterForRemoteNotificationsWithDeviceToken deviceToken: Data) {
    Messaging.messaging().apnsToken = deviceToken
    
    Messaging.messaging().token { token, error in
        if let token = token {
            // Send token to your backend
            self.registerDeviceToken(token, deviceType: "ios")
        }
    }
}
```

#### Web (JavaScript)
```javascript
// Initialize Firebase
import { initializeApp } from 'firebase/app';
import { getMessaging, getToken, onMessage } from 'firebase/messaging';

const firebaseConfig = {
    // Your Firebase configuration
};

const app = initializeApp(firebaseConfig);
const messaging = getMessaging(app);

// Request permission and get token
async function requestNotificationPermission() {
    try {
        const permission = await Notification.requestPermission();
        if (permission === 'granted') {
            const token = await getToken(messaging, { vapidKey: 'YOUR_VAPID_KEY' });
            // Send token to your backend
            await registerDeviceToken(token, 'web');
        }
    } catch (error) {
        console.error('Error getting notification permission:', error);
    }
}

// Handle foreground notifications
onMessage(messaging, (payload) => {
    console.log('Message received:', payload);
    // Display notification
    new Notification(payload.notification.title, {
        body: payload.notification.body,
        icon: '/icon.png'
    });
});
```

### 6. API Endpoints for Device Token Management

Add these endpoints to your API:

```go
// POST /api/v1/notifications/devices/register
func (h *NotificationHandlers) RegisterDevice(c echo.Context) error {
    var req struct {
        DeviceToken string `json:"device_token"`
        DeviceType  string `json:"device_type"`
        DeviceName  string `json:"device_name"`
        AppVersion  string `json:"app_version"`
    }
    
    if err := c.Bind(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
    }
    
    tenantID := c.Get("tenant_id").(uuid.UUID)
    userID := c.Get("user_id").(uuid.UUID)
    
    token := &models.DeviceToken{
        TenantID:    tenantID,
        UserID:      userID,
        DeviceToken: req.DeviceToken,
        DeviceType:  req.DeviceType,
        DeviceName:  req.DeviceName,
        AppVersion:  req.AppVersion,
        IsActive:    true,
    }
    
    if err := h.deviceTokenRepo.RegisterToken(c.Request().Context(), token); err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register device")
    }
    
    return c.JSON(http.StatusOK, map[string]string{"message": "Device registered successfully"})
}

// DELETE /api/v1/notifications/devices/:token
func (h *NotificationHandlers) UnregisterDevice(c echo.Context) error {
    deviceToken := c.Param("token")
    tenantID := c.Get("tenant_id").(uuid.UUID)
    
    if err := h.deviceTokenRepo.DeactivateToken(c.Request().Context(), tenantID, deviceToken); err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to unregister device")
    }
    
    return c.JSON(http.StatusOK, map[string]string{"message": "Device unregistered successfully"})
}
```

### 7. Environment Variables

Add these to your `.env` file:

```env
# Firebase Configuration
FIREBASE_CREDENTIALS_PATH=/path/to/serviceAccountKey.json
FCM_ENABLED=true

# Optional: for web push
FIREBASE_WEB_VAPID_KEY=your_vapid_key_here
```

### 8. Testing

1. **Test in development:**
   - Use Firebase Console's "Cloud Messaging" test tool
   - Send test messages to specific device tokens

2. **Test notification delivery:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/notifications/test-push \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "title": "Test Notification",
       "message": "This is a test push notification",
       "user_id": "uuid-here"
     }'
   ```

## Security Considerations

1. **Token Security:**
   - Store device tokens securely
   - Encrypt tokens at rest if required by compliance
   - Implement token rotation

2. **Rate Limiting:**
   - Implement rate limits per user/tenant
   - Prevent notification spam

3. **User Privacy:**
   - Allow users to opt-out of push notifications
   - Respect notification preferences
   - Implement "Do Not Disturb" hours

4. **Credentials:**
   - NEVER commit Firebase credentials to version control
   - Use environment variables or secret management systems
   - Rotate credentials periodically

## Monitoring and Analytics

1. **Track delivery metrics:**
   - Sent/delivered/failed counts
   - Click-through rates
   - Token refresh rates

2. **Set up alerts:**
   - High failure rates
   - Invalid token rates
   - FCM quota limits

3. **Log important events:**
   - Token registration/deregistration
   - Failed deliveries
   - Rate limit violations

## Cost Considerations

- FCM is free for most use cases
- Monitor your FCM usage in Firebase Console
- Be aware of quota limits:
  - Default: 500,000 messages per day
  - Higher limits available on request

## Additional Resources

- [Firebase Cloud Messaging Documentation](https://firebase.google.com/docs/cloud-messaging)
- [FCM Admin SDK Go](https://pkg.go.dev/firebase.google.com/go/v4/messaging)
- [APNs Documentation](https://developer.apple.com/documentation/usernotifications)
- [Web Push API](https://developer.mozilla.org/en-US/docs/Web/API/Push_API)
