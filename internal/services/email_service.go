package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"mime"
	"net/smtp"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/resend/resend-go/v2"
)

var (
	resendOnce    sync.Once
	resendClient  *resend.Client
	resendInitErr error
	smtpConfig    *SMTPConfig
)

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	FromEmail string
	FromName  string
}

func initResendClient() error {
	resendOnce.Do(func() {
		apiKey := strings.TrimSpace(os.Getenv("RESEND_API_KEY"))
		if apiKey == "" {
			resendInitErr = errors.New("RESEND_API_KEY not set")
			return
		}
		resendClient = resend.NewClient(apiKey)
	})
	return resendInitErr
}

func initSMTPConfig() {
	if smtpConfig != nil {
		return
	}

	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	if host == "" {
		return
	}

	port := 587 // default
	if portStr := os.Getenv("SMTP_PORT"); portStr != "" {
		if p, err := fmt.Sscanf(portStr, "%d", &port); err != nil || p <= 0 {
			log.Printf("Invalid SMTP_PORT value: %s, using default 587", portStr)
			port = 587
		}
	}

	smtpConfig = &SMTPConfig{
		Host:      host,
		Port:      port,
		Username:  strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		Password:  strings.TrimSpace(os.Getenv("SMTP_PASSWORD")),
		FromEmail: strings.TrimSpace(os.Getenv("SMTP_FROM_EMAIL")),
		FromName:  strings.TrimSpace(os.Getenv("SMTP_FROM_NAME")),
	}

	if smtpConfig.FromEmail == "" {
		smtpConfig.FromEmail = smtpConfig.Username
	}
}

func SendVerificationEmailAsync(ctx context.Context, toEmail string, token string, frontendURL string) <-chan error {
	if ctx == nil {
		ctx = context.Background()
	}
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		
		// Attempt with retry logic (3 attempts with exponential backoff)
		maxRetries := 3
		var lastErr error
		
		for attempt := 1; attempt <= maxRetries; attempt++ {
			if attempt > 1 {
				// Exponential backoff: 2s, 4s, 8s
				backoffDuration := time.Duration(1<<uint(attempt-1)) * time.Second
				log.Printf("Retrying email send to %s after %v (attempt %d/%d)", toEmail, backoffDuration, attempt, maxRetries)
				time.Sleep(backoffDuration)
			}
			
			err := sendVerificationEmailWithRetry(ctx, toEmail, token, frontendURL)
			if err == nil {
				log.Printf("Verification email sent successfully to %s on attempt %d", toEmail, attempt)
				return
			}
			
			lastErr = err
			log.Printf("Email send attempt %d/%d failed for %s: %v", attempt, maxRetries, toEmail, err)
		}
		
		// All retries failed - log critical error and notify admin
		log.Printf("CRITICAL: Failed to send verification email to %s after %d attempts. Error: %v", toEmail, maxRetries, lastErr)
		
		// Send admin notification in background (non-blocking)
		go notifyAdminOfEmailFailure(toEmail, lastErr)
		
		errCh <- fmt.Errorf("failed to send verification email after %d attempts: %w", maxRetries, lastErr)
	}()
	return errCh
}

// sendVerificationEmailWithRetry attempts to send verification email once
func sendVerificationEmailWithRetry(ctx context.Context, toEmail string, token string, frontendURL string) error {
	if err := initResendClient(); err != nil {
		log.Printf("resend init failed for %s: %v", toEmail, err)
		return err
	}
	
	sanitizedRecipient := strings.TrimSpace(toEmail)
	if sanitizedRecipient == "" {
		return errors.New("recipient email not provided")
	}
	
	fromEmail := strings.TrimSpace(os.Getenv("FROM_EMAIL"))
	if fromEmail == "" {
		return errors.New("FROM_EMAIL not set")
	}
	
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	
	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", strings.TrimSuffix(frontendURL, "/"), url.QueryEscape(token))
	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verify Your Email Address</title>
    <style>
        body { margin: 0; padding: 0; background-color: #f5f7fa; font-family: 'Helvetica Neue', Arial, sans-serif; color: #1f2937; }
        .wrapper { width: 100%%; table-layout: fixed; background-color: #f5f7fa; padding: 20px 0; }
        .main { background-color: #ffffff; margin: 0 auto; width: 100%%; max-width: 600px; border-radius: 12px; overflow: hidden; box-shadow: 0 10px 40px rgba(15, 23, 42, 0.08); }
        .header { background: linear-gradient(135deg, #1d4ed8 0%%, #9333ea 100%%); padding: 32px; text-align: center; }
        .header h1 { margin: 0; color: #ffffff; font-size: 24px; letter-spacing: 0.5px; }
        .content { padding: 36px 32px; }
        .content p { margin: 0 0 16px; }
        .button { display: inline-block; padding: 14px 28px; background-color: #1d4ed8; color: #ffffff; text-decoration: none; border-radius: 999px; font-weight: 600; letter-spacing: 0.3px; }
        .button:hover { background-color: #1e40af; }
        .link { word-break: break-all; color: #1d4ed8; }
        .footer { padding: 24px 32px 32px; background-color: #f9fafb; text-align: center; font-size: 12px; color: #6b7280; }
        @media only screen and (max-width: 620px) {
            .content { padding: 28px 20px; }
            .header { padding: 28px 20px; }
        }
    </style>
</head>
<body>
    <table class="wrapper" role="presentation" cellspacing="0" cellpadding="0" border="0">
        <tr>
            <td align="center">
                <table class="main" role="presentation" cellspacing="0" cellpadding="0" border="0">
                    <tr>
                        <td class="header">
                            <h1>Verify Your Email</h1>
                        </td>
                    </tr>
                    <tr>
                        <td class="content">
                            <p>Hello,</p>
                            <p>Thanks for joining us. Please confirm your email to activate your inventory account.</p>
                            <p style="text-align: center; margin: 32px 0;">
                                <a href="%s" class="button" target="_blank" rel="noopener">Verify Email Address</a>
                            </p>
                            <p>If the button above does not work, copy and paste this link into your browser:</p>
                            <p><a class="link" href="%s">%s</a></p>
                            <p>This link expires in 24 hours. If you did not request this, you can safely ignore this email.</p>
                        </td>
                    </tr>
                    <tr>
                        <td class="footer">
                            <p>You received this email because a new account was created using this address.</p>
                            <p>&copy; %d Inventory Management. All rights reserved.</p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>`, verificationURL, verificationURL, verificationURL, time.Now().Year())
	
	// Try Resend first
	log.Printf("sendVerificationEmail: attempting to send via Resend to %s with from=%s subject='Verify Your Email Address'", sanitizedRecipient, fromEmail)
	resp, err := resendClient.Emails.SendWithContext(ctx, &resend.SendEmailRequest{
		From:    fromEmail,
		To:      []string{sanitizedRecipient},
		Subject: "Verify Your Email Address",
		Html:    htmlContent,
	})

	if err == nil {
		log.Printf("sendVerificationEmail: Resend send succeeded for %s: email_id=%s", sanitizedRecipient, resp.Id)
		return nil
	}

	log.Printf("sendVerificationEmail: Resend send failed for %s: %v", sanitizedRecipient, err)

	// Try SMTP fallback if Resend fails
	initSMTPConfig()
	if smtpConfig != nil && smtpConfig.Host != "" {
		log.Printf("sendVerificationEmail: attempting SMTP fallback for %s", sanitizedRecipient)
		smtpErr := sendEmailViaSMTP(sanitizedRecipient, "Verify Your Email Address", htmlContent)
		if smtpErr == nil {
			log.Printf("sendVerificationEmail: SMTP fallback succeeded for %s", sanitizedRecipient)
			return nil
		}
		log.Printf("sendVerificationEmail: SMTP fallback failed for %s: %v", sanitizedRecipient, smtpErr)
		return fmt.Errorf("both Resend and SMTP failed: Resend error: %v, SMTP error: %v", err, smtpErr)
	}
	
	log.Printf("sendVerificationEmail: No SMTP configuration available for fallback")
	return fmt.Errorf("email send failed via Resend and no SMTP fallback configured: %w", err)
}

// notifyAdminOfEmailFailure sends a notification to admin about email failure
func notifyAdminOfEmailFailure(recipientEmail string, err error) {
	adminEmail := strings.TrimSpace(os.Getenv("ADMIN_EMAIL"))
	if adminEmail == "" {
		log.Printf("ADMIN_EMAIL not configured - cannot notify admin of email failure for %s", recipientEmail)
		return
	}
	
	subject := fmt.Sprintf("CRITICAL: Verification Email Failure for %s", recipientEmail)
	body := fmt.Sprintf(`
<html>
<body>
<h2>Critical Email Delivery Failure</h2>
<p><strong>Recipient:</strong> %s</p>
<p><strong>Error:</strong> %v</p>
<p><strong>Timestamp:</strong> %s</p>
<p><strong>Action Required:</strong> Please investigate email service configuration and consider manually resending verification email to the user.</p>
</body>
</html>
`, recipientEmail, err, time.Now().Format(time.RFC3339))
	
	// Try to send admin notification via SMTP
	initSMTPConfig()
	if smtpConfig != nil && smtpConfig.Host != "" {
		if err := sendEmailViaSMTP(adminEmail, subject, body); err != nil {
			log.Printf("Failed to notify admin about email failure: %v", err)
		} else {
			log.Printf("Admin notification sent successfully for failed email to %s", recipientEmail)
		}
	} else {
		log.Printf("Cannot notify admin - SMTP not configured")
	}
}

// sendEmailViaSMTP sends email using SMTP protocol
func sendEmailViaSMTP(recipient, subject, body string) error {
	if smtpConfig == nil || smtpConfig.Host == "" {
		return fmt.Errorf("SMTP not configured")
	}

	fromEmail := smtpConfig.FromEmail
	if fromEmail == "" {
		return fmt.Errorf("SMTP sender email not configured")
	}

	encodedSubject := mime.QEncoding.Encode("utf-8", subject)
	fromHeader := fromEmail
	if smtpConfig.FromName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", mime.QEncoding.Encode("utf-8", smtpConfig.FromName), fromEmail)
	}

	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("From: %s\r\n", fromHeader))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", recipient))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", encodedSubject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(body)

	addr := fmt.Sprintf("%s:%d", smtpConfig.Host, smtpConfig.Port)

	var auth smtp.Auth
	if smtpConfig.Username != "" {
		auth = smtp.PlainAuth("", smtpConfig.Username, smtpConfig.Password, smtpConfig.Host)
	}

	log.Printf("sendEmailViaSMTP: attempting to send to %s via %s", recipient, addr)
	if err := smtp.SendMail(addr, auth, fromEmail, []string{recipient}, msg.Bytes()); err != nil {
		return fmt.Errorf("failed to send email via SMTP: %w", err)
	}

	log.Printf("sendEmailViaSMTP: successfully sent email to %s", recipient)
	return nil
}
