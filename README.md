# Multi Upload API

Uma API completa e escalável para gerenciamento de imagens e vídeos, desenvolvida em Go com PostgreSQL.

## 🚀 Características

- **Escalável**: Arquitetura preparada para crescimento
- **Segura**: Todas as rotas protegidas por JWT
- **Performática**: Paginação eficiente e consultas otimizadas
- **Persistente**: Dados e arquivos mantidos entre deployments
- **Completa**: CRUD completo com ordenação personalizada

## 📋 Funcionalidades

- ✅ Upload de imagens e vídeos
- ✅ Listagem paginada com filtros
- ✅ Atualização de arquivos (substituição)
- ✅ Exclusão de arquivos
- ✅ Sistema de ordenação personalizada
- ✅ Novos uploads automaticamente em primeiro lugar
- ✅ Autenticação JWT
- ✅ Persistência de dados com Docker volumes
- ✅ API extremamente rápida e otimizada

## 🛠️ Tecnologias

- **Go 1.21** - Linguagem principal
- **Gin** - Framework web
- **PostgreSQL 15** - Banco de dados
- **JWT** - Autenticação
- **Docker & Docker Compose** - Containerização

## 📦 Instalação e Execução

### Pré-requisitos

- Docker
- Docker Compose

### 1. Clone o repositório

```bash
git clone <seu-repositorio>
cd multi-upload-project
```

### 2. Configure as variáveis de ambiente

Edite o arquivo `config.env` se necessário (valores padrão já configurados):

```env
DB_HOST=postgres
DB_PORT=5433
DB_USER=postgres
DB_PASSWORD=postgres123
DB_NAME=multiupload
JWT_SECRET=seu_jwt_secret_super_seguro_aqui_mude_em_producao
PORT=8082
UPLOAD_PATH=/app/uploads
```

### 3. Execute a aplicação

```bash
# Subir todos os serviços
docker-compose up -d

# Verificar se está funcionando
curl http://localhost:8082/health
```

### 4. Criar usuário padrão

```bash
# Executar script de criação de usuário
docker-compose exec api go run scripts/create_user.go
```

Ou localmente (se tiver Go instalado):

```bash
./scripts/create_user.sh
```

## 📚 Documentação da API

### Base URL

```
http://localhost:8082/api/v1
```

### Autenticação

Todas as rotas (exceto `/login`, `/gallery` e `/files/*`) requerem autenticação via JWT no header:

```
Authorization: Bearer <seu_token_jwt>
```

---

## 🔐 Rotas de Autenticação

### POST /login

Autentica o usuário e retorna um token JWT.

**Request:**
```json
{
  "username": "admin",
  "password": "senha123"
}
```

**Response (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "admin",
    "created_at": "2024-01-01T10:00:00Z",
    "updated_at": "2024-01-01T10:00:00Z"
  }
}
```

### GET /me

Retorna informações do usuário autenticado.

**Response (200):**
```json
{
  "id": 1,
  "username": "admin",
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:00:00Z"
}
```

---

## 📁 Rotas de Mídia

### POST /media/upload

Faz upload de um arquivo (imagem ou vídeo).

**Request:**
- Content-Type: `multipart/form-data`
- Field: `file` (arquivo)

**Tipos suportados:**
- Imagens: JPG, PNG, GIF, WebP, etc.
- Vídeos: MP4, AVI, MOV, etc.
- Tamanho máximo: 1GB (sem limitação para vídeos grandes)

**Response (201):**
```json
{
  "media": {
    "id": 1,
    "user_id": 1,
    "filename": "uuid-generated-name.jpg",
    "original_name": "minha-foto.jpg",
    "file_path": "2024/01/01/uuid-generated-name.jpg",
    "file_size": 1024000,
    "mime_type": "image/jpeg",
    "media_type": "image",
    "sort_order": 1,
    "created_at": "2024-01-01T10:00:00Z",
    "updated_at": "2024-01-01T10:00:00Z"
  },
  "message": "Arquivo enviado com sucesso"
}
```

### GET /media

Lista arquivos com paginação e filtros.

**Query Parameters:**
- `page` (int): Página (padrão: 1)
- `page_size` (int): Itens por página (padrão: 20, máx: 100)
- `type` (string): Filtrar por tipo (`image` ou `video`)
- `order_by` (string): Ordenação
  - `sort_order` (padrão): Ordem personalizada (novos uploads ficam em primeiro)
  - `created_at_desc`: Mais recentes primeiro
  - `created_at_asc`: Mais antigos primeiro
  - `filename_asc`: Nome A-Z
  - `filename_desc`: Nome Z-A
  - `size_asc`: Menor tamanho primeiro
  - `size_desc`: Maior tamanho primeiro

**Exemplos:**
```bash
GET /media?page=1&page_size=10&type=image&order_by=created_at_desc
```

**Response (200):**
```json
{
  "data": [
    {
      "id": 1,
      "user_id": 1,
      "filename": "uuid-name.jpg",
      "original_name": "foto.jpg",
      "file_path": "2024/01/01/uuid-name.jpg",
      "file_size": 1024000,
      "mime_type": "image/jpeg",
      "media_type": "image",
      "sort_order": 1,
      "created_at": "2024-01-01T10:00:00Z",
      "updated_at": "2024-01-01T10:00:00Z"
    }
  ],
  "total": 50,
  "page": 1,
  "page_size": 20,
  "total_pages": 3
}
```

### GET /media/:id

Busca um arquivo específico.

**Response (200):**
```json
{
  "id": 1,
  "user_id": 1,
  "filename": "uuid-name.jpg",
  "original_name": "foto.jpg",
  "file_path": "2024/01/01/uuid-name.jpg",
  "file_size": 1024000,
  "mime_type": "image/jpeg",
  "media_type": "image",
  "sort_order": 1,
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:00:00Z"
}
```

### PUT /media/:id

Atualiza propriedades de um arquivo.

**Request:**
```json
{
  "sort_order": 5
}
```

**Response (200):**
```json
{
  "id": 1,
  "sort_order": 5,
  "updated_at": "2024-01-01T11:00:00Z"
}
```

### PUT /media/:id/replace

Substitui um arquivo existente por um novo.

**Request:**
- Content-Type: `multipart/form-data`
- Field: `file` (novo arquivo)

**Response (200):**
```json
{
  "media": {
    "id": 1,
    "filename": "new-uuid-name.jpg",
    "original_name": "nova-foto.jpg",
    "file_size": 2048000,
    "updated_at": "2024-01-01T11:00:00Z"
  },
  "message": "Arquivo substituído com sucesso"
}
```

### DELETE /media/:id

Exclui um arquivo permanentemente.

**Response (200):**
```json
{
  "message": "Arquivo excluído com sucesso"
}
```

### POST /media/sort

Atualiza a ordem de múltiplos arquivos.

**Request:**
```json
{
  "media_ids": [3, 1, 2, 5, 4]
}
```

**Response (200):**
```json
{
  "message": "Ordem atualizada com sucesso"
}
```

---

## 🖼️ Galeria Pública

### GET /gallery

Lista todas as mídias publicamente (sem autenticação necessária). Ideal para ser usado em sites como galeria.

**Query Parameters:**
- `page` (int): Página (padrão: 1)
- `page_size` (int): Itens por página (padrão: 20, máx: 100)
- `type` (string): Filtrar por tipo (`image` ou `video`)
- `order_by` (string): Ordenação
  - `sort_order` (padrão): Ordem personalizada (novos uploads ficam em primeiro)
  - `created_at_desc`: Mais recentes primeiro
  - `created_at_asc`: Mais antigos primeiro
  - `filename_asc`: Nome A-Z
  - `filename_desc`: Nome Z-A
  - `size_asc`: Menor tamanho primeiro
  - `size_desc`: Maior tamanho primeiro

**Exemplos:**
```bash
GET /gallery?page=1&page_size=10&type=image&order_by=created_at_desc
```

**Response (200):**
```json
{
  "data": [
    {
      "id": 1,
      "user_id": 1,
      "filename": "uuid-name.jpg",
      "original_name": "foto.jpg",
      "file_path": "2024/01/01/uuid-name.jpg",
      "file_size": 1024000,
      "mime_type": "image/jpeg",
      "media_type": "image",
      "sort_order": 1,
      "created_at": "2024-01-01T10:00:00Z",
      "updated_at": "2024-01-01T10:00:00Z"
    }
  ],
  "total": 50,
  "page": 1,
  "page_size": 20,
  "total_pages": 3
}
```

**Para usar as imagens/vídeos:**
```bash
# URL completa para visualizar uma mídia
http://localhost:8082/api/v1/files/2024/01/01/uuid-name.jpg
```

---

## 📂 Rotas de Arquivos

### GET /files/*filepath

Serve arquivos estáticos (público, sem autenticação).

**Exemplo:**
```bash
GET /files/2024/01/01/uuid-name.jpg
```

Retorna o arquivo diretamente para visualização/download.

---

## 🏥 Health Check

### GET /health

Verifica se a API está funcionando.

**Response (200):**
```json
{
  "status": "ok",
  "message": "API funcionando corretamente"
}
```

---

## 🔧 Comandos Úteis

### Gerenciar containers

```bash
# Subir serviços
docker-compose up -d

# Ver logs
docker-compose logs -f api
docker-compose logs -f postgres

# Parar serviços
docker-compose down

# Rebuild da aplicação
docker-compose up --build -d api
```

### Backup e Restore

```bash
# Backup do banco
docker-compose exec postgres pg_dump -U postgres multiupload > backup.sql

# Restore do banco
docker-compose exec -T postgres psql -U postgres multiupload < backup.sql

# Backup dos arquivos (volumes são automaticamente persistentes)
docker run --rm -v multiupload_uploads_data:/data -v $(pwd):/backup alpine tar czf /backup/uploads-backup.tar.gz -C /data .
```

### Desenvolvimento

```bash
# Executar localmente (requer Go)
go mod download
go run main.go

# Executar testes
go test ./...

# Criar usuário
go run scripts/create_user.go
```

---

## 🐳 Volumes Docker

A aplicação usa volumes nomeados para persistência:

- `postgres_data`: Dados do PostgreSQL
- `uploads_data`: Arquivos de upload

**IMPORTANTE**: Estes volumes garantem que os dados NÃO sejam perdidos durante redeploys!

---

## 🔒 Segurança

- Todas as rotas protegidas por JWT
- Senhas criptografadas com bcrypt
- Validação de tipos de arquivo
- Suporte a vídeos grandes (até 1GB)
- Headers CORS configurados
- Usuários isolados (cada usuário vê apenas seus arquivos)

---

## ⚡ Performance

- Índices otimizados no banco de dados
- Paginação eficiente
- Pool de conexões configurado
- Consultas SQL otimizadas
- Middleware de cache-friendly
- Estrutura de arquivos organizada por data

---

## 🚀 Deploy em Produção

### Coolify

1. Conecte seu repositório Git
2. Configure as variáveis de ambiente
3. Certifique-se que os volumes estão mapeados
4. Deploy automático via Git hooks

### Variáveis de ambiente importantes:

```env
JWT_SECRET=um_secret_muito_seguro_e_longo_para_producao
ENVIRONMENT=production
```

---

## 📞 Suporte

Para dúvidas ou problemas:

1. Verifique os logs: `docker-compose logs -f`
2. Teste a conexão: `curl http://localhost:8082/health`
3. Verifique os volumes: `docker volume ls`

---

## 📄 Licença

Este projeto está licenciado sob a MIT License.

---

**Desenvolvido com ❤️ em Go**
