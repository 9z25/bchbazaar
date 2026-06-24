package main

import (
	"fmt"
	"log"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
)

func rawEmailAddress(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}

	addr, err := mail.ParseAddress(value)
	if err != nil {
		return "", err
	}

	return addr.Address, nil
}

func sendEmail(to string, subject string, body string) error {
	toEmail, err := rawEmailAddress(to)
	if err != nil {
		return fmt.Errorf("invalid recipient email %q: %w", to, err)
	}

	if toEmail == "" {
		return nil
	}

	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	port := strings.TrimSpace(os.Getenv("SMTP_PORT"))
	user := strings.TrimSpace(os.Getenv("SMTP_USER"))
	pass := strings.ReplaceAll(strings.TrimSpace(os.Getenv("SMTP_PASS")), " ", "")
	fromRaw := strings.TrimSpace(os.Getenv("SMTP_FROM"))

	if fromRaw == "" {
		fromRaw = user
	}

	fromEmail, err := rawEmailAddress(fromRaw)
	if err != nil {
		return fmt.Errorf("invalid SMTP_FROM %q: %w", fromRaw, err)
	}

	if host == "" || port == "" || user == "" || pass == "" || fromEmail == "" {
		log.Println("email skipped: SMTP env vars not configured")
		return nil
	}

	subject = strings.ReplaceAll(subject, "\r", "")
	subject = strings.ReplaceAll(subject, "\n", "")

	auth := smtp.PlainAuth("", user, pass, host)

	fromHeader := (&mail.Address{
		Name:    "BCHBazaar",
		Address: fromEmail,
	}).String()

	toHeader := (&mail.Address{
		Address: toEmail,
	}).String()

	message := strings.Join([]string{
		"From: " + fromHeader,
		"To: " + toHeader,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=\"utf-8\"",
		"",
		body,
	}, "\r\n")

	return smtp.SendMail(
		host+":"+port,
		auth,
		fromEmail,          // raw email only
		[]string{toEmail},  // raw email only
		[]byte(message),
	)
}

type OrderEmailData struct {
	OrderID       uint64
	ListingTitle  string
	Amount        float64
	Currency      string
	BuyerEmail    string
	SellerEmail   string
	BuyerUsername string
	SellerUsername string
}

func getOrderEmailData(orderID any) (*OrderEmailData, error) {
	var data OrderEmailData

	err := db.QueryRow(`
		SELECT
			o.id,
			l.title,
			o.amount,
			o.currency,
			COALESCE(b.email, ''),
			COALESCE(s.email, ''),
			COALESCE(b.username, ''),
			COALESCE(s.username, '')
		FROM orders o
		JOIN listings l ON l.id = o.listing_id
		JOIN users s ON s.id = l.user_id
		LEFT JOIN users b ON b.id = o.buyer_user_id
		WHERE o.id = ?
	`, orderID).Scan(
		&data.OrderID,
		&data.ListingTitle,
		&data.Amount,
		&data.Currency,
		&data.BuyerEmail,
		&data.SellerEmail,
		&data.BuyerUsername,
		&data.SellerUsername,
	)

	if err != nil {
		return nil, err
	}

	return &data, nil
}

func sendOrderEmailAsync(orderID any, event string) {
	go func() {
		if err := sendOrderEmail(orderID, event); err != nil {
			log.Println("email notification failed:", err)
		}
	}()
}

func sendOrderEmail(orderID any, event string) error {
	data, err := getOrderEmailData(orderID)
	if err != nil {
		return err
	}

	price := fmt.Sprintf("%.8f %s", data.Amount, data.Currency)

	switch event {
	case "paid":
		return sendEmail(
			data.SellerEmail,
			"New paid order on BCHBazaar",
			fmt.Sprintf(
				"Your listing has been paid.\n\nItem: %s\nOrder: #%d\nAmount: %s\n\nNext step: ship the item and add tracking.",
				data.ListingTitle,
				data.OrderID,
				price,
			),
		)

	case "shipped":
		return sendEmail(
			data.BuyerEmail,
			"Your BCHBazaar order has shipped",
			fmt.Sprintf(
				"Your order has shipped.\n\nItem: %s\nOrder: #%d\nAmount: %s\n\nCheck BCHBazaar for tracking details.",
				data.ListingTitle,
				data.OrderID,
				price,
			),
		)

	case "completed":
		buyerErr := sendEmail(
			data.BuyerEmail,
			"Your BCHBazaar order is complete",
			fmt.Sprintf(
				"Your order is complete.\n\nItem: %s\nOrder: #%d\nAmount: %s",
				data.ListingTitle,
				data.OrderID,
				price,
			),
		)

		sellerErr := sendEmail(
			data.SellerEmail,
			"Your BCHBazaar sale is complete",
			fmt.Sprintf(
				"Your sale is complete.\n\nItem: %s\nOrder: #%d\nAmount: %s",
				data.ListingTitle,
				data.OrderID,
				price,
			),
		)

		if buyerErr != nil {
			return buyerErr
		}
		return sellerErr
	}

	return nil
}