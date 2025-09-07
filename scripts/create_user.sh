#!/bin/bash

# Script para criar usuário padrão
echo "=== Script de Criação de Usuário ==="

# Verificar se está no diretório correto
if [ ! -f "go.mod" ]; then
    echo "❌ Execute este script a partir do diretório raiz do projeto"
    exit 1
fi

# Executar o script Go
echo "🚀 Executando criação de usuário..."
go run scripts/create_user.go

echo "✅ Script executado!"
