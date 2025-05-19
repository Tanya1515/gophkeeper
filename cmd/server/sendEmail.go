package main

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

func sendOTP(email string, tempPassword string) error {

	m := gomail.NewMessage()
	m.SetHeader("From", "gopherkeeper4@gmail.com")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Gophkeeper OTP")
	m.SetBody("text/plain", "One-time password for Gopherkeeper: "+tempPassword)

	d := gomail.NewDialer("smtp.gmail.com", 587, "gopherkeeper4@gmail.com", "cssc ddvu qgul wzhq")

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("error while sending OTP: %w", err)
	}
	return nil
}
