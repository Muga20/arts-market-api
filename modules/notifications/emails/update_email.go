package emails

import (
	"fmt"
	"net/smtp"

	"github.com/muga20/artsMarket/config"
	"github.com/muga20/artsMarket/pkg/logs/handlers"
)

// SendEmailVerificationEmail sends an email verification email with an embedded link
func SendEmailVerificationEmail(toEmail string, verificationToken string, responseHandler *handlers.ResponseHandler) error {
	// Get SMTP configuration from the config package
	smtpHost := config.Envs.SMTPHost
	smtpPort := config.Envs.SMTPPort
	fromEmail := config.Envs.SMTPUser
	fromPassword := config.Envs.SMTPPassword
	clientURL := config.Envs.ClientURL

	// Generate the verification link
	verifyLink := fmt.Sprintf("%s/verify-email?token=%s", clientURL, verificationToken)

	// Set up authentication information
	auth := smtp.PlainAuth("", fromEmail, fromPassword, smtpHost)

	// Compose the email
	subject := "Verify Your New Email Address"
	body := fmt.Sprintf(`
		<html>
		<body>
			<p>You recently updated your email address. Please verify your new email by clicking the link below:</p>
			<p><a href="%s" style="color: blue; text-decoration: none;">Verify Your Email</a></p>
			<p>This link will expire in 24 hours. If you did not request this change, please contact support.</p>
		</body>
		</html>`, verifyLink)

	// Format the email message
	message := []byte(fmt.Sprintf("Subject: %s\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s", subject, body))

	// Send the email
	err := smtp.SendMail(fmt.Sprintf("%s:%s", smtpHost, smtpPort), auth, fromEmail, []string{toEmail}, message)
	if err != nil {
		// Handle the error using the provided responseHandler
		return responseHandler.Handle(nil, nil, fmt.Errorf("failed to send email verification email: %v", err))
	}
	return nil
}
