package services

import (
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
	// Converter porta para int
	port, err := strconv.Atoi(e.config.SMTPPort)
	if err != nil {
		return fmt.Errorf("porta SMTP inválida: %v", err)
	}

	// Configurar dialer
	d := gomail.NewDialer(e.config.SMTPHost, port, e.config.SMTPUsername, e.config.SMTPPassword)

	// Criar mensagem
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("JAM Locacao de Guindastes <%s>", e.config.FromEmail))
	m.SetHeader("To", e.config.ContactEmail)
	m.SetHeader("Reply-To", req.Email)
	m.SetHeader("Subject", fmt.Sprintf("Contato do Site - %s", req.Subject))

	// Corpo do email em HTML
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Nova mensagem de contato do site</h2>
			<p><strong>Nome:</strong> %s</p>
			<p><strong>Email:</strong> %s</p>
			<p><strong>Assunto:</strong> %s</p>
			<p><strong>Mensagem:</strong></p>
			<p>%s</p>
			<hr>
			<p><em>Esta mensagem foi enviada através do formulário de contato do site.</em></p>
		</body>
		</html>
	`, req.Name, req.Email, req.Subject, req.Message)

	m.SetBody("text/html", body)

	// Enviar email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("erro ao enviar email: %v", err)
	}

	return nil
}
