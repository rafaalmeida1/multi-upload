.PHONY: help build up down logs clean restart create-user backup restore test

# Cores para output
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
NC := \033[0m # No Color

help: ## Mostrar esta ajuda
	@echo "$(GREEN)Multi Upload API - Comandos disponíveis:$(NC)"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(YELLOW)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Construir a aplicação
	@echo "$(GREEN)Construindo aplicação...$(NC)"
	docker-compose build

up: ## Subir todos os serviços
	@echo "$(GREEN)Subindo serviços...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)✅ Serviços iniciados!$(NC)"
	@echo "API: http://localhost:8080"
	@echo "Health: http://localhost:8080/health"

down: ## Parar todos os serviços
	@echo "$(YELLOW)Parando serviços...$(NC)"
	docker-compose down

logs: ## Ver logs dos serviços
	docker-compose logs -f

logs-api: ## Ver logs apenas da API
	docker-compose logs -f api

logs-db: ## Ver logs apenas do banco
	docker-compose logs -f postgres

clean: ## Limpar containers e imagens
	@echo "$(RED)Limpando containers e imagens...$(NC)"
	docker-compose down --rmi all --volumes --remove-orphans
	docker system prune -f

restart: down up ## Reiniciar todos os serviços

create-user: ## Criar novo usuário
	@echo "$(GREEN)Criando usuário...$(NC)"
	docker-compose exec api go run scripts/create_user.go

status: ## Ver status dos serviços
	docker-compose ps

health: ## Verificar saúde da API
	@echo "$(GREEN)Verificando saúde da API...$(NC)"
	@curl -s http://localhost:8080/health | jq . || echo "$(RED)API não está respondendo$(NC)"

backup-db: ## Fazer backup do banco de dados
	@echo "$(GREEN)Fazendo backup do banco...$(NC)"
	docker-compose exec postgres pg_dump -U postgres multiupload > backup-$(shell date +%Y%m%d_%H%M%S).sql
	@echo "$(GREEN)✅ Backup criado!$(NC)"

backup-files: ## Fazer backup dos arquivos
	@echo "$(GREEN)Fazendo backup dos arquivos...$(NC)"
	docker run --rm -v multi-upload-project_uploads_data:/data -v $(PWD):/backup alpine tar czf /backup/uploads-backup-$(shell date +%Y%m%d_%H%M%S).tar.gz -C /data .
	@echo "$(GREEN)✅ Backup dos arquivos criado!$(NC)"

restore-db: ## Restaurar banco de dados (especificar arquivo: make restore-db FILE=backup.sql)
	@if [ -z "$(FILE)" ]; then echo "$(RED)Especifique o arquivo: make restore-db FILE=backup.sql$(NC)"; exit 1; fi
	@echo "$(YELLOW)Restaurando banco de dados...$(NC)"
	docker-compose exec -T postgres psql -U postgres multiupload < $(FILE)
	@echo "$(GREEN)✅ Banco restaurado!$(NC)"

test: ## Executar testes
	go test ./...

dev: ## Executar em modo desenvolvimento (local)
	go run main.go

install: ## Instalar dependências
	go mod download
	go mod tidy

lint: ## Verificar código com golangci-lint
	golangci-lint run

format: ## Formatar código
	go fmt ./...

docker-clean: ## Limpar tudo do Docker (CUIDADO!)
	@echo "$(RED)⚠️  ATENÇÃO: Isso vai remover TODOS os volumes e dados!$(NC)"
	@read -p "Tem certeza? (y/N): " confirm && [ "$$confirm" = "y" ]
	docker-compose down -v --remove-orphans
	docker system prune -af --volumes

quick-test: up ## Teste rápido da API
	@echo "$(GREEN)Testando API...$(NC)"
	@sleep 5
	@curl -s http://localhost:8080/health | jq . && echo "$(GREEN)✅ API funcionando!$(NC)" || echo "$(RED)❌ API com problemas$(NC)"

# Comandos de desenvolvimento
dev-logs: ## Ver logs em tempo real durante desenvolvimento
	docker-compose up

dev-rebuild: ## Rebuild e restart para desenvolvimento
	docker-compose up --build -d api

# Comandos de produção
prod-deploy: ## Deploy em produção
	@echo "$(GREEN)Fazendo deploy em produção...$(NC)"
	docker-compose -f docker-compose.yml up -d --build
	@echo "$(GREEN)✅ Deploy concluído!$(NC)"

prod-backup: backup-db backup-files ## Backup completo para produção
