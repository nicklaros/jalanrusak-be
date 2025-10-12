package messaging

import (
	"context"
	"fmt"

	"github.com/nicklaros/jalanrusak-be/core/ports/external"
)

// ConsoleEmailService implements EmailService by printing emails to console (for development)
type ConsoleEmailService struct{}

// NewConsoleEmailService creates a new console-based email service
func NewConsoleEmailService() external.EmailService {
	return &ConsoleEmailService{}
}

// SendPasswordResetEmail prints the password reset email to console
func (s *ConsoleEmailService) SendPasswordResetEmail(ctx context.Context, to, name, resetToken string) error {
	fmt.Println("========================================")
	fmt.Println("📧 PASSWORD RESET EMAIL (Console)")
	fmt.Println("========================================")
	fmt.Printf("To: %s <%s>\n", name, to)
	fmt.Println("Subject: Reset Your Password")
	fmt.Println("----------------------------------------")
	fmt.Printf("Hi %s,\n\n", name)
	fmt.Println("You requested to reset your password. Use the token below:")
	fmt.Printf("\nReset Token: %s\n\n", resetToken)
	fmt.Println("This token will expire in 1 hour.")
	fmt.Println("If you didn't request this, please ignore this email.")
	fmt.Println("========================================")
	return nil
}

// SendWelcomeEmail prints the welcome email to console
func (s *ConsoleEmailService) SendWelcomeEmail(ctx context.Context, to, name string) error {
	fmt.Println("========================================")
	fmt.Println("📧 WELCOME EMAIL (Console)")
	fmt.Println("========================================")
	fmt.Printf("To: %s <%s>\n", name, to)
	fmt.Println("Subject: Welcome to JalanRusak!")
	fmt.Println("----------------------------------------")
	fmt.Printf("Hi %s,\n\n", name)
	fmt.Println("Welcome to JalanRusak! Your account has been created successfully.")
	fmt.Println("Thank you for joining us.")
	fmt.Println("========================================")
	return nil
}

// SendPasswordChangedEmail prints the password changed notification to console
func (s *ConsoleEmailService) SendPasswordChangedEmail(ctx context.Context, to, name string) error {
	fmt.Println("========================================")
	fmt.Println("📧 PASSWORD CHANGED EMAIL (Console)")
	fmt.Println("========================================")
	fmt.Printf("To: %s <%s>\n", name, to)
	fmt.Println("Subject: Your Password Was Changed")
	fmt.Println("----------------------------------------")
	fmt.Printf("Hi %s,\n\n", name)
	fmt.Println("Your password has been changed successfully.")
	fmt.Println("If you didn't make this change, please contact support immediately.")
	fmt.Println("========================================")
	return nil
}
