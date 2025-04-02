package emails

import (
	"fmt"
	"net/smtp"

	"github.com/muga20/artsMarket/config"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
)

// SendPasswordResetEmail sends a password reset email with an embedded reset link
func SendPasswordResetEmail(toEmail string, resetToken string, responseHandler *handlers.ResponseHandler) error {
	// Get SMTP configuration from the config package
	smtpHost := config.Envs.SMTPHost
	smtpPort := config.Envs.SMTPPort
	fromEmail := config.Envs.SMTPUser
	fromPassword := config.Envs.SMTPPassword
	clientURL := config.Envs.ClientURL

	// Generate the reset link
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", clientURL, resetToken)

	// Set up authentication information
	auth := smtp.PlainAuth("", fromEmail, fromPassword, smtpHost)

	// Compose the email
	subject := "Password Reset Request"
	body := fmt.Sprintf(`
		<html>
		<body>
			<p>You requested a password reset. Click the link below to reset your password:</p>
			<p><a href="%s" style="color: blue; text-decoration: none;">Reset Your Password</a></p>
			<p>If you did not request this, please ignore this email.</p>
		</body>
		</html>`, resetLink)

	// Format the email message
	message := []byte(fmt.Sprintf("Subject: %s\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s", subject, body))

	// Send the email
	err := smtp.SendMail(fmt.Sprintf("%s:%s", smtpHost, smtpPort), auth, fromEmail, []string{toEmail}, message)
	if err != nil {
		// Handle the error using the provided responseHandler
		return responseHandler.Handle(nil, nil, fmt.Errorf("failed to send password reset email: %v", err))
	}
	return nil
}
