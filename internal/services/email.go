package services

import (
	"crypto/tls"
	"fmt"
	"multi-upload-api/internal/config"
	"strconv"

	"gopkg.in/gomail.v2"
)

type EmailService struct {
	config *config.Config
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{
		config: cfg,
	}
}

type ContactRequest struct {
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email" binding:"required,email"`
	Subject string `json:"subject" binding:"required"`
	Message string `json:"message" binding:"required"`
}

func (e *EmailService) SendContactEmail(req *ContactRequest) error {
	fmt.Println("[EmailService] Iniciando envio de e-mail via Zoho...")

	// Converter porta
	port, err := strconv.Atoi(e.config.SMTPPort)
	if err != nil {
		return fmt.Errorf("[EmailService] porta SMTP inválida: %v", err)
	}

	// Configurar dialer
	d := gomail.NewDialer(e.config.SMTPHost, port, e.config.SMTPUsername, e.config.SMTPPassword)
	d.SSL = false // STARTTLS
	d.TLSConfig = &tls.Config{
		ServerName: e.config.SMTPHost, // necessário para validar certificado
	}

	// Criar mensagem
	m := gomail.NewMessage()
	m.SetAddressHeader("From", e.config.FromEmail, e.config.FromName) // corrige acentos no nome
	m.SetHeader("To", e.config.ContactEmail)
	m.SetHeader("Reply-To", req.Email)
	m.SetHeader("Subject", fmt.Sprintf("Contato do Site - %s", req.Subject))

	// Corpo em HTML
	body := fmt.Sprintf(`
        <html>
        <body>
            <h2>Nova mensagem de contato do site</h2>
            <p><strong>Nome:</strong> %s</p>
            <p><strong>Email:</strong> %s</p>
            <p><strong>Assunto:</strong> %s</p>
            <p><strong>Mensagem:</strong></p>
            <p>%s</p>
        </body>
        </html>
    `, req.Name, req.Email, req.Subject, req.Message)

	m.SetBody("text/html", body)

	fmt.Printf("[EmailService] Tentando enviar email para: %s\n", e.config.ContactEmail)

	// Enviar email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("[EmailService] erro ao enviar email: %v", err)
	}

	fmt.Println("[EmailService] E-mail enviado com sucesso via Zoho!")
	return nil
}
