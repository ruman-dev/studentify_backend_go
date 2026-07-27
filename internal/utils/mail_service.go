package utils

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	gomail "gopkg.in/mail.v2"
)

func GenerateOTP() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func SendOTP(to string, otp string) error {
	m := gomail.NewMessage()

	m.SetHeader("From", os.Getenv("SMTP_EMAIL"))
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Your OTP Code")

	body := fmt.Sprintf(`
		<h2>Your Verification Code</h2>
		<p>Your OTP is:</p>
		<h1>%s</h1>
		<p>This OTP expires in 5 minutes.</p>
	`, otp)

	m.SetBody("text/html", body)

	d := gomail.NewDialer(
		os.Getenv("SMTP_HOST"),
		587,
		os.Getenv("SMTP_EMAIL"),
		os.Getenv("SMTP_PASSWORD"),
	)

	return d.DialAndSend(m)
}
