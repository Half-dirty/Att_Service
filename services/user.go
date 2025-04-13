package services

func SendVerificationEmail(to, link string) error {
	// m := gomail.NewMessage()
	// m.SetHeader("From", "youremail@example.com")
	// m.SetHeader("To", to)
	// m.SetHeader("Subject", "Email Confirmation")
	// body := fmt.Sprintf("Click <a href=\"%s\">here</a> to confirm your email.", link)
	// m.SetBody("text/html", body)

	// d := gomail.NewDialer("smtp.example.com", 587, "youremail@example.com", "password")
	// return d.DialAndSend(m)
	return nil
}
