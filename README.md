# Microblog Platform

Una plataforma de microblogging similar a Twitter construida con Go.

## Contenido
1. [Arquitectura](#arquitectura)
2. [Requisitos](#requisitos)
3. [Configuración](#configuración)
4. [API Documentation](#api-documentation)
5. [Tests](#tests)
6. [Escalabilidad](#escalabilidad)
7. [Desarrollo](#desarrollo)

## Arquitectura

### Visión General
La plataforma está construida siguiendo los principios de Clean Architecture:

```
microblog-platform/
├── cmd/
│   └── api/                # Punto de entrada de la aplicación
├── internal/
│   ├── models/            # Definiciones de modelos
│   ├── handlers/          # Manejadores HTTP
│   ├── repository/        # Capa de acceso a datos
│   └── service/           # Lógica de negocio
├── pkg/
│   ├── database/          # Configuración de base de datos
│   └── utils/            # Utilidades comunes
└── docker/               # Configuración de Docker
```

### Componentes Principales
1. **API Layer (Handlers)**
   - Manejo de requests HTTP
   - Validación de entrada
   - Respuestas HTTP
   - Rate limiting

2. **Service Layer**
   - Lógica de negocio
   - Validaciones
   - Coordinación entre repositorios

3. **Repository Layer**
   - Acceso a base de datos
   - Operaciones CRUD
   - Queries optimizadas

4. **Database**
   - MongoDB para almacenamiento
   - Índices optimizados para lecturas
   - Esquema flexible para futura evolución

### Diagrama de Arquitectura
```mermaid
graph TD
    A[API Layer] --> B[Service Layer]
    B --> C[Repository Layer]
    C --> D[MongoDB]
    E[Docker] --> D
    F[Health Check] --> A
```

## Requisitos

### Tecnologías Necesarias
- Go 1.23 o superior
- Docker y Docker Compose
- MongoDB
- Make (opcional)

### Dependencias Principales
```go
require (
    github.com/gin-gonic/gin v1.10.0
    go.mongodb.org/mongo-driver v1.17.1
    github.com/stretchr/testify v1.9.0
)
```

## Configuración

### Variables de Entorno
```env
PORT=8080
MONGODB_URI=mongodb://microblog_user:microblog_password@localhost:27017/microblog
MONGODB_DATABASE=microblog
```

### Docker
```yaml
version: "3.8"
services:
  mongodb:
    image: mongo:latest
    environment:
      MONGO_INITDB_ROOT_USERNAME: admin
      MONGO_INITDB_ROOT_PASSWORD: adminpassword
    ports:
      - "27017:27017"
    volumes:
      - mongodb_data:/data/db
```

## API Documentation

### Endpoints

#### Users
```
POST /api/v1/users
- Crear nuevo usuario
Request:
{
    "username": "string",
    "email": "string"
}
Response: 201 Created
{
    "id": "string",
    "username": "string",
    "email": "string"
}
```

```
GET /api/v1/users/:id
- Obtener usuario por ID
Response: 200 OK
{
    "id": "string",
    "username": "string",
    "email": "string",
    "followers_count": int
}
```

#### Following
```
POST /api/v1/users/:id/follow/:target_id
- Seguir a un usuario
Response: 200 OK
{
    "message": "Usuario seguido exitosamente"
}
```

```
POST /api/v1/users/:id/unfollow/:target_id
- Dejar de seguir a un usuario
Response: 200 OK
{
    "message": "Usuario dejado de seguir exitosamente"
}
```

#### Tweets
```
POST /api/v1/tweets
- Crear nuevo tweet
Request:
{
    "user_id": "string",
    "content": "string"
}
Response: 201 Created
{
    "id": "string",
    "user_id": "string",
    "content": "string",
    "created_at": "datetime"
}
```

```
GET /api/v1/users/:id/tweets
- Obtener tweets de un usuario
Response: 200 OK
{
    "count": int,
    "tweets": [
        {
            "id": "string",
            "content": "string",
            "created_at": "datetime"
        }
    ]
}
```

```
GET /api/v1/users/:id/timeline?page=1&limit=10
- Obtener timeline personalizado
Response: 200 OK
{
    "page": int,
    "limit": int,
    "count": int,
    "tweets": [
        {
            "id": "string",
            "user_id": "string",
            "content": "string",
            "created_at": "datetime"
        }
    ]
}
```

### Códigos de Error
- 400: Bad Request (validación fallida)
- 404: Not Found (recurso no encontrado)
- 500: Internal Server Error

## Tests

### Ejecutar Tests
```bash
# Todos los tests
go test ./...

# Tests específicos
go test -v ./internal/repository

# Cobertura
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Estructura de Tests
```go
// Unit tests
func TestUserRepository_Create(t *testing.T)
func TestTweetRepository_GetTimeline(t *testing.T)

// Integration tests
func TestUserTweetFlow(t *testing.T)
```

## Desarrollo

### Setup Local
```bash
# Clonar repositorio
git clone https://github.com/ffelixf/microblog-platform.git
cd microblog-platform

# Instalar dependencias
go mod download

# Iniciar MongoDB
docker-compose up -d

# Ejecutar aplicación
go run cmd/api/main.go
```

### Comandos Útiles
```bash
# Lint
go vet ./...

# Format
go fmt ./...

# Build
go build -o bin/api cmd/api/main.go
```
[![Test Coverage](https://img.shields.io/badge/coverage-80.3%25-green.svg)](docs/ARCHITECTURE.md#tests-y-calidad)
[![Go Version](https://img.shields.io/badge/go-1.23-blue.svg)](https://golang.org/doc/go1.23)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

# Enlaces Rápidos
- [Arquitectura](docs/ARCHITECTURE.md)
- [Configuración](docs/ARCHITECTURE.md#configuración-y-despliegue)
- [API Documentation](docs/ARCHITECTURE.md#api-documentation)
- [Tests](docs/ARCHITECTURE.md#tests-y-calidad)

# Quick Start

```bash
# Clonar el repositorio
git clone https://github.com/ffelixf/microblog-platform.git
cd microblog-platform

# Instalar dependencias
go mod download

# Iniciar servicios con Docker
docker-compose up -d

# Ejecutar la aplicación
go run cmd/api/main.go

# Ejecutar tests
go test ./...

# Ver cobertura de tests
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

# Ejemplos de Uso Rápido
```bash
# Crear un usuario
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com"}'

# Crear un tweet
curl -X POST http://localhost:8080/api/v1/tweets \
  -H "Content-Type: application/json" \
  -d '{"user_id":"<USER_ID>","content":"Hello, World!"}'

# Obtener timeline
curl http://localhost:8080/api/v1/users/<USER_ID>/timeline
```
## API Documentation

La documentación de la API está disponible en:
- Swagger UI: http://localhost:8080/swagger/index.html
- API Docs JSON: http://localhost:8080/swagger/doc.json