package services

import (
	"fmt"
	"log"
	"net/smtp"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

type Mailer struct {
	config Config
}

func NewMailer(cfg Config) *Mailer {
	return &Mailer{
		config: cfg,
	}
}

func (m *Mailer) SendHTMLEmail(to, subject, htmlBody string) error {

	headers := map[string]string{
		"From":         m.config.From,
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=\"UTF-8\"",
	}

	var msg string
	for k, v := range headers {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += "\r\n" + htmlBody

	auth := smtp.PlainAuth(m.config.From, m.config.Username, m.config.Password, m.config.Host)

	addr := fmt.Sprintf("%s:%s", m.config.Host, m.config.Port)

	err := smtp.SendMail(addr, auth, m.config.From, []string{to}, []byte(msg))
	if err != nil {
		log.Printf("Gagal mengirim email HTML ke %s: %v", to, err)
		return fmt.Errorf("gagal mengirim email HTML: %w", err)
	}

	return nil
}

func BuildOTPEmailBody(otpCode string, expiryMinutes int) string {
	return fmt.Sprintf(`
        <!DOCTYPE html>
        <html>
        <head>
            <meta charset="utf-8">
            <title>Kode OTP Reset Kata Sandi Anda</title>
            <style>
                body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
                .container { max-width: 600px; margin: 20px auto; padding: 20px; border: 1px solid #ddd; border-radius: 5px; }
                .header { background-color: #f8f8f8; padding: 10px 0; text-align: center; border-bottom: 1px solid #ddd; }
                .content { padding: 20px; text-align: center; }
                .otp-code { font-size: 2em; font-weight: bold; color: #007bff; margin: 20px 0; padding: 10px; background-color: #e9f5ff; border-radius: 5px; display: inline-block;}
                .footer { font-size: 0.8em; color: #777; text-align: center; margin-top: 20px; border-top: 1px solid #ddd; padding-top: 10px; }
            </style>
        </head>
        <body>
            <div class="container">
                <div class="header">
                    <h2>Verifikasi Kode untuk Reset Kata Sandi</h2>
                </div>
                <div class="content">
                    <p>Kami menerima permintaan untuk mereset kata sandi akun Anda.</p>
                    <p>Masukkan kode verifikasi berikut pada halaman verifikasi OTP:</p>
                    <p class="otp-code">%s</p>
                    <p>Kode ini akan kedaluwarsa dalam **%d menit**.</p>
                    <p>Jika Anda tidak meminta reset kata sandi, abaikan email ini.</p>
                    <p>Terima kasih,</p>
                    <p>Tim Toko Bulan</p>
                </div>
                <div class="footer">
                    <p>&copy; 2025 Toko Bulan. Semua hak dilindungi.</p>
                </div>
            </div>
        </body>
        </html>
    `, otpCode, expiryMinutes)
}
