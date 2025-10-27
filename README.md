# Rate Limiter - Desafio GoExpert

Um rate limiter robusto e configurável implementado em Go, que pode limitar o número de requisições com base em endereço IP ou token de acesso.

## 🚀 Como Usar

### Pré-requisitos

- Go 1.24.1 ou superior
- Docker e Docker Compose (opcional)
- Redis (opcional)

### Crie o arquivo `.env`

Crie um arquivo `.env` na raiz do projeto usando o modelo fornecido em `env.example`:

```bash
cp env.example .env
```

Para mais detalhes sobre as configurações, veja a seção [⚙️ Configuração](#-configuração).

### Executando com Docker Compose (Recomendado)

```bash
# Iniciar a aplicação
docker-compose up -d
```

A aplicação estará disponível em http://localhost:8080

### Executando Localmente

```bash
# Instalar dependências
go mod download

# Executar a aplicação
go run ./cmd/server/main.go
```

### Executando Testes

```bash
go test ./...
```

## 🏗️ Arquitetura

```
rate-limiter/
├── cmd/
│   └── server/
│       └── main.go              # Aplicação principal
├── internal/
│   ├── config/
│   │   └── config.go            # Gerenciamento de configurações
│   ├── limiter/
│   │   ├── limiter.go           # Lógica do rate limiter
│   │   └── limiter_test.go      # Testes do limiter
│   ├── middleware/
│   │   ├── ratelimiter.go       # Middleware HTTP
│   │   └── ratelimiter_test.go  # Testes do middleware
│   └── storage/
│       ├── storage.go           # Interface de storage
│       ├── memory.go            # Implementação em memória
│       ├── memory_test.go       # Testes do memory storage
│       └── redis.go             # Implementação com Redis
├── docker-compose.yml
├── Dockerfile
└── README.md
```

## ⚙️ Configuração

As configurações são feitas através de variáveis de ambiente. Você pode criar um arquivo `.env` na raiz do projeto:

```env
SERVER_PORT=8080

# Configuração do Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Limitação por IP
RATE_LIMIT_IP=10
RATE_LIMIT_IP_BLOCK_TIME=300

# Limitação por Token (padrão)
RATE_LIMIT_TOKEN=100
RATE_LIMIT_TOKEN_BLOCK_TIME=300

# Configurações Específicas de Tokens
# Token 1
TOKEN_ONE=abc123
TOKEN_ONE_LIMIT=100
TOKEN_ONE_BLOCK_TIME=300

# Token 2
TOKEN_TWO=xyz789
TOKEN_TWO_LIMIT=50
TOKEN_TWO_BLOCK_TIME=600
```

### Configurações Específicas por Token

Você pode definir limites personalizados para tokens específicos usando o padrão:

```env
TOKEN_{NOME}={valor_do_token}
TOKEN_{NOME}_LIMIT={limite}
TOKEN_{NOME}_BLOCK_TIME={tempo_em_segundos}
```

Exemplo:
```env
TOKEN_PREMIUM=premium-key-123
TOKEN_PREMIUM_LIMIT=1000
TOKEN_PREMIUM_BLOCK_TIME=60

TOKEN_BASIC=basic-key-456
TOKEN_BASIC_LIMIT=10
TOKEN_BASIC_BLOCK_TIME=3600
```

## 🔧 API Endpoints

### Health Check
```bash
GET /health
```

Retorna o status da aplicação (não possui rate limiting).

**Resposta:**
```json
{
  "status": "healthy"
}
```

### Endpoint Principal
```bash
GET /
```

Endpoint com rate limiting aplicado.

**Headers (opcional):**
```
API_KEY: seu-token-aqui
```

**Resposta de Sucesso (200):**
```json
{
  "message": "Welcome to Rate Limiter API",
  "status": "ok"
}
```

**Resposta de Bloqueio (429):**
```json
{
  "error": "you have reached the maximum number of requests or actions allowed within a certain time frame"
}
```

## 📊 Exemplos de Uso

### Teste de Limitação por IP

```bash
# Fazer 15 requisições sem token (limite padrão de IP: 10)
for i in {1..15}; do
  curl http://localhost:8080/
  echo ""
done
```

As primeiras 10 requisições retornarão 200, as demais retornarão 429.

### Teste de Limitação por Token

```bash
# Requisição com token
curl -H "API_KEY: abc123" http://localhost:8080/
```

### Teste com Script de Carga

Temos o script `test_complete.sh` para testar o rate limiter com diferentes cenários:

```bash
# Executar script de teste completo
chmod +x test_complete.sh
./test_complete.sh
```

## 🧪 Testes

O projeto possui testes unitários completos para todas as camadas:

- **Storage Tests**: Testam as implementações de armazenamento (memória e Redis)
- **Limiter Tests**: Testam a lógica de rate limiting
- **Middleware Tests**: Testam a integração com HTTP

```bash
go test ./...
```


## Descrição do Desafio

### Objetivo:

Desenvolver um rate limiter em Go que possa ser configurado para limitar o número máximo de requisições por segundo com base em um endereço IP específico ou em um token de acesso.

### Descrição:

O objetivo deste desafio é criar um rate limiter em Go que possa ser utilizado para controlar o tráfego de requisições para um serviço web. O rate limiter deve ser capaz de limitar o número de requisições com base em dois critérios:

1. **Endereço IP**: O rate limiter deve restringir o número de requisições recebidas de um único endereço IP dentro de um intervalo de tempo definido.
2. **Token de Acesso**: O rate limiter deve também poderá limitar as requisições baseadas em um token de acesso único, permitindo diferentes limites de tempo de expiração para diferentes tokens. O Token deve ser informado no header no seguinte formato:
```
API_KEY: <TOKEN>
```
3. As configurações de limite do token de acesso devem se sobrepor as do IP. Ex: Se o limite por IP é de 10 req/s e a de um determinado token é de 100 req/s, o rate limiter deve utilizar as informações do token.

### Requisitos:

- O rate limiter deve poder trabalhar como um middleware que é injetado ao servidor web
- O rate limiter deve permitir a configuração do número máximo de requisições permitidas por segundo.
- O rate limiter deve ter ter a opção de escolher o tempo de bloqueio do IP ou do Token caso a quantidade de requisições tenha sido excedida.
- As configurações de limite devem ser realizadas via variáveis de ambiente ou em um arquivo “.env” na pasta raiz.
- Deve ser possível configurar o rate limiter tanto para limitação por IP quanto por token de acesso.
- O sistema deve responder adequadamente quando o limite é excedido:
    - Código HTTP: **429**
    - Mensagem: **you have reached the maximum number of requests or actions allowed within a certain time frame**
- Todas as informações de "limiter” devem ser armazenadas e consultadas de um banco de dados Redis. Você pode utilizar docker-compose para subir o Redis.
- Crie uma “strategy” que permita trocar facilmente o Redis por outro mecanismo de persistência.
- A lógica do limiter deve estar separada do middleware.

### Exemplos:

1. **Limitação por IP**: Suponha que o rate limiter esteja configurado para permitir no máximo 5 requisições por segundo por IP. Se o IP **192.168.1.1** enviar 6 requisições em um segundo, a sexta requisição deve ser bloqueada.
2. **Limitação por Token**: Se um token abc123 tiver um limite configurado de 10 requisições por segundo e enviar 11 requisições nesse intervalo, a décima primeira deve ser bloqueada.
3. Nos dois casos acima, as próximas requisições poderão ser realizadas somente quando o tempo total de expiração ocorrer. Ex: Se o tempo de expiração é de 5 minutos, determinado IP poderá realizar novas requisições somente após os 5 minutos.

### Dicas:

- Teste seu rate limiter sob diferentes condições de carga para garantir que ele funcione conforme esperado em situações de alto tráfego.

### Entrega:

- O código-fonte completo da implementação.
- Documentação explicando como o rate limiter funciona e como ele pode ser configurado.
- Testes automatizados demonstrando a eficácia e a robustez do rate limiter.
- Utilize docker/docker-compose para que possamos realizar os testes de sua aplicação.
- O servidor web deve responder na porta 8080.