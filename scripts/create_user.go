package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"multi-upload-api/internal/config"
	"multi-upload-api/internal/database"
	"multi-upload-api/internal/models"
	"multi-upload-api/internal/repository"
	"os"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"golang.org/x/term"
)

func loadOptionalEnvFiles(files ...string) {
	for _, file := range files {
		if err := godotenv.Load(file); err == nil {
			return
		} else if !errors.Is(err, os.ErrNotExist) {
			log.Printf("Aviso: não foi possível carregar %s: %v", file, err)
		}
	}
}

func main() {
	fmt.Println("=== Script de Criação de Usuário ===")

	// Carregar variáveis de ambiente
	loadOptionalEnvFiles("config.env", ".env")

	// Configurar aplicação
	cfg := config.Load()

	// Conectar ao banco de dados
	db, err := database.Connect(cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("Erro ao conectar com o banco de dados: %v", err)
	}
	defer db.Close()

	// Executar migrations
	if err := database.RunMigrations(cfg.DatabaseURL()); err != nil {
		log.Fatalf("Erro ao executar migrations: %v", err)
	}

	// Inicializar repositório
	userRepo := repository.NewUserRepository(db)

	reader := bufio.NewReader(os.Stdin)

	// Solicitar username
	fmt.Print("Digite o nome de usuário: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Erro ao ler username: %v", err)
	}
	username = strings.TrimSpace(username)

	if username == "" {
		log.Fatal("Username não pode estar vazio")
	}

	// Verificar se usuário já existe
	existingUser, err := userRepo.GetByUsername(username)
	if err == nil && existingUser != nil {
		fmt.Printf("Usuário '%s' já existe!\n", username)
		fmt.Print("Deseja atualizar a senha? (s/N): ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "s" && response != "sim" {
			fmt.Println("Operação cancelada.")
			return
		}
	}

	// Solicitar senha
	fmt.Print("Digite a senha: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatalf("Erro ao ler senha: %v", err)
	}
	password := string(passwordBytes)
	fmt.Println() // Nova linha após a senha

	if len(password) < 6 {
		log.Fatal("Senha deve ter pelo menos 6 caracteres")
	}

	// Confirmar senha
	fmt.Print("Confirme a senha: ")
	confirmPasswordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatalf("Erro ao ler confirmação da senha: %v", err)
	}
	confirmPassword := string(confirmPasswordBytes)
	fmt.Println() // Nova linha após a confirmação

	if password != confirmPassword {
		log.Fatal("Senhas não coincidem")
	}

	// Criar usuário
	user := &models.User{
		Username: username,
	}

	if err := user.HashPassword(password); err != nil {
		log.Fatalf("Erro ao criptografar senha: %v", err)
	}

	// Se usuário já existe, atualizar senha
	if existingUser != nil {
		// Como não temos um método Update no repositório, vamos fazer manualmente
		query := `UPDATE users SET password = $1, updated_at = CURRENT_TIMESTAMP WHERE username = $2`
		_, err := db.Exec(query, user.Password, username)
		if err != nil {
			log.Fatalf("Erro ao atualizar usuário: %v", err)
		}
		fmt.Printf("✅ Senha do usuário '%s' atualizada com sucesso!\n", username)
	} else {
		// Criar novo usuário
		if err := userRepo.Create(user); err != nil {
			log.Fatalf("Erro ao criar usuário: %v", err)
		}
		fmt.Printf("✅ Usuário '%s' criado com sucesso!\n", username)
	}

	fmt.Println("\n=== Informações do Usuário ===")
	fmt.Printf("ID: %d\n", user.ID)
	fmt.Printf("Username: %s\n", user.Username)
	fmt.Printf("Criado em: %s\n", user.CreatedAt.Format("2006-01-02 15:04:05"))

	fmt.Println("\n🔑 Use estas credenciais para fazer login na API:")
	fmt.Printf("Username: %s\n", username)
	fmt.Println("Senha: [a senha que você digitou]")
}
