package utils

import (
	"fmt"
	"os"
	"strconv"
	"time"
	"log"
    "gopkg.in/gomail.v2"
)

//for production
// resendRequest defines the payload for Resend API

// type resendRequest struct {
// 	From    string `json:"from"`
// 	To      string `json:"to"`
// 	Subject string `json:"subject"`
// 	HTML    string `json:"html"`
// }

// // SendEmail is a general utility to send emails via Resend API
// func SendEmail(to, subject, body string) error {
// 	apiKey := os.Getenv("RESEND_API_KEY")
// 	if apiKey == "" {
// 		return fmt.Errorf("missing RESEND_API_KEY environment variable")
// 	}

// 	reqBody := resendRequest{
// 		From:    "Auth Service <no-reply@yourdomain.com>", // Use your verified domain
// 		To:      to,
// 		Subject: subject,
// 		HTML:    body,
// 	}

// 	jsonData, err := json.Marshal(reqBody)
// 	if err != nil {
// 		return err
// 	}

// 	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		return err
// 	}

// 	req.Header.Set("Authorization", "Bearer "+apiKey)
// 	req.Header.Set("Content-Type", "application/json")

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode >= 400 {
// 		return fmt.Errorf("resend API returned status %d", resp.StatusCode)
// 	}

// 	return nil
// }

// // SendWelcomeEmail sends a welcome email to the newly registered user only
// func SendWelcomeEmail(to, username string) error {
// 	subject := "Welcome to Our Platform 🎉"
// 	body := fmt.Sprintf(`
// 	<!DOCTYPE html>
// 	<html lang="en">
// 	<head>
// 		<meta charset="UTF-8" />
// 		<meta name="viewport" content="width=device-width, initial-scale=1.0" />
// 		<title>Welcome</title>
// 		<style>
// 			body {
// 				font-family: Arial, sans-serif;
// 				background-color: #f9fafb;
// 				color: #111827;
// 				padding: 20px;
// 			}
// 			.container {
// 				max-width: 600px;
// 				margin: 0 auto;
// 				background: white;
// 				border-radius: 10px;
// 				padding: 30px;
// 				box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
// 			}
// 			h1 {
// 				color: #2563eb;
// 			}
// 			.footer {
// 				margin-top: 30px;
// 				font-size: 13px;
// 				color: #6b7280;
// 			}
// 		</style>
// 	</head>
// 	<body>
// 		<div class="container">
// 			<h1>Welcome, %s 👋</h1>
// 			<p>We’re excited to have you join our community!</p>
// 			<p>Your account has been successfully created. You can now log in and start exploring our platform’s features built just for you.</p>
// 			<p>If you have any questions, simply reply to this email — we’re always happy to help.</p>
// 			<br/>
// 			<p>Cheers,<br/>The Team</p>
// 			<div class="footer">
// 				<p>© 2025 Auth Service. All rights reserved.</p>
// 			</div>
// 		</div>
// 	</body>
// 	</html>
// 	`, username)

// 	return SendEmail(to, subject, body)
// }
//for development
func SendWelcomeEmail(to, name string) error {
	// Load SMTP credentials from environment variables
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")

	if smtpHost == "" || smtpPortStr == "" || smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("missing SMTP configuration environment variables")
	}

	port, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		return fmt.Errorf("invalid SMTP_PORT: %v", err)
	}

	// Create the HTML message
	body := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<title>Welcome</title>
			<style>
				body {
					font-family: Arial, sans-serif;
					background-color: #f7f9fc;
					color: #333;
					padding: 20px;
				}
				.container {
					max-width: 600px;
					margin: 0 auto;
					background: white;
					border-radius: 12px;
					padding: 30px;
					box-shadow: 0 4px 12px rgba(0,0,0,0.08);
				}
				h1 {
					color: #0061ff;
					font-size: 24px;
				}
				.button {
					display: inline-block;
					padding: 10px 20px;
					background-color: #0061ff;
					color: white;
					border-radius: 6px;
					text-decoration: none;
					margin-top: 20px;
				}
				.footer {
					margin-top: 30px;
					font-size: 12px;
					color: #888;
				}
			</style>
		</head>
		<body>
			<div class="container">
				<h1>Welcome, %s 👋</h1>
				<p>We’re thrilled to have you at <strong>Divya Packing</strong>!</p>
				<p>Your account has been successfully created. You can now log in and start exploring our secure authentication platform.</p>
				<a href="https://yourdomain.com/login" class="button">Get Started</a>
				<div class="footer">
					<p>If you didn’t create this account, please ignore this email.</p>
					<p>© 2025 Divya Packing. All rights reserved.</p>
				</div>
			</div>
		</body>
		</html>
	`, name)

	// Prepare message
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("Divya Packing <%s>", smtpUser))
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Welcome to Divya Packing 🎉")
	m.SetBody("text/html", body)

	// Set up dialer
	d := gomail.NewDialer(smtpHost, port, smtpUser, smtpPass)

	// Retry logic
	const maxRetries = 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := d.DialAndSend(m)
		if err == nil {
			log.Printf("📧 Welcome email sent successfully to %s", to)
			return nil
		}

		log.Printf("[mailer] send failed (attempt %d/%d): %v", attempt, maxRetries, err)
		if attempt < maxRetries {
			time.Sleep(10 * time.Second)
		}
	}

	return fmt.Errorf("failed to send welcome email after %d attempts", maxRetries)
}