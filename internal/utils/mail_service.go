package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"strconv"

	gomail "gopkg.in/mail.v2"
)

func GenerateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", n.Int64()), nil
}

func SendOTP(to, otp string) error {
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	if from == "" || password == "" || host == "" || port == "" {
		return fmt.Errorf("SMTP_EMAIL, SMTP_PASSWORD, and SMTP_HOST are required")
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("convert smtp port to int: %w", err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Your OTP Code")
	m.SetBody("text/html", fmt.Sprintf(`
		<h2>Your Verification Code</h2>
		<p>Your OTP is:</p>
		<h1>%s</h1>
		<p>This OTP expires in 5 minutes.</p>
	`, otp))

	d := gomail.NewDialer(host, portInt, from, password)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("dial and send: %w", err)
	}
	return nil
}
