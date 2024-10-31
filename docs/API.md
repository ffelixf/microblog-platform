# API Documentation - Microblog Platform

## Contenido
- [General](#general)
- [Autenticación](#autenticación)
- [Endpoints](#endpoints)
  - [Users](#users)
  - [Tweets](#tweets)
  - [Timeline](#timeline)
  - [Health](#health)
- [Errores](#errores)
- [Ejemplos](#ejemplos)

## General

- Base URL: `http://localhost:8080`
- Formato: Todos los endpoints aceptan y devuelven datos en formato JSON
- Timestamps: Todos los timestamps están en formato ISO 8601

## Autenticación
Por simplicidad, no se requiere autenticación. El ID de usuario se envía como parte de las peticiones.

## Endpoints

### Users

#### Crear Usuario
```http
POST /api/v1/users

Request:
{
    "username": "string",     // requerido, único
    "email": "string"        // requerido, único
}

Response: 201 Created
{
    "id": "string",
    "username": "string",
    "email": "string",
    "created_at": "datetime",
    "updated_at": "datetime",
    "followers_count": 0
}

Errores:
- 400: Datos inválidos
- 409: Username o email ya existe
```

#### Obtener Usuario
```http
GET /api/v1/users/:id

Response: 200 OK
{
    "id": "string",
    "username": "string",
    "email": "string",
    "created_at": "datetime",
    "updated_at": "datetime",
    "followers_count": integer
}

Errores:
- 404: Usuario no encontrado
```

#### Seguir Usuario
```http
POST /api/v1/users/:id/follow/:target_id

Response: 200 OK
{
    "message": "Usuario seguido exitosamente",
    "user_id": "string",
    "following_id": "string"
}

Errores:
- 400: No se puede seguir a uno mismo
- 404: Usuario objetivo no encontrado
```

#### Dejar de Seguir Usuario
```http
POST /api/v1/users/:id/unfollow/:target_id

Response: 200 OK
{
    "message": "Usuario dejado de seguir exitosamente",
    "user_id": "string",
    "unfollowed_id": "string"
}

Errores:
- 404: Usuario objetivo no encontrado
```

#### Obtener Siguiendo
```http
GET /api/v1/users/:id/following

Response: 200 OK
{
    "count": integer,
    "following": [
        {
            "id": "string",
            "username": "string",
            "email": "string"
        }
    ]
}

Errores:
- 404: Usuario no encontrado
```

### Tweets

#### Crear Tweet
```http
POST /api/v1/tweets

Request:
{
    "user_id": "string",     // requerido
    "content": "string"      // requerido, max 280 caracteres
}

Response: 201 Created
{
    "id": "string",
    "user_id": "string",
    "content": "string",
    "created_at": "datetime"
}

Errores:
- 400: Contenido inválido o muy largo
- 404: Usuario no encontrado
```

#### Obtener Tweets de Usuario
```http
GET /api/v1/users/:id/tweets

Response: 200 OK
{
    "user_id": "string",
    "count": integer,
    "tweets": [
        {
            "id": "string",
            "content": "string",
            "created_at": "datetime"
        }
    ]
}

Errores:
- 404: Usuario no encontrado
```

### Timeline

#### Obtener Timeline
```http
GET /api/v1/users/:id/timeline?page=1&limit=10

Query Parameters:
- page: integer (default: 1)
- limit: integer (default: 10, max: 100)

Response: 200 OK
{
    "user_id": "string",
    "page": integer,
    "limit": integer,
    "count": integer,
    "tweets": [
        {
            "id": "string",
            "user_id": "string",
            "content": "string",
            "created_at": "datetime"
        }
    ]
}

Errores:
- 400: Parámetros de paginación inválidos
- 404: Usuario no encontrado
```

### Health

#### Health Check
```http
GET /health

Response: 200 OK
{
    "status": "ok",
    "timestamp": "datetime"
}
```

#### Database Health Check
```http
GET /health/db

Response: 200 OK
{
    "status": "ok",
    "database": {
        "status": "ok",
        "type": "mongodb"
    }
}
```

## Errores

### Formato de Error
```json
{
    "error": "Descripción del error"
}
```

### Códigos de Estado
- 200: Éxito
- 201: Recurso creado
- 400: Error de validación
- 404: Recurso no encontrado
- 409: Conflicto (duplicado)
- 429: Demasiadas peticiones
- 500: Error interno del servidor

## Ejemplos

### Flujo Típico de Uso

1. Crear usuarios:
```bash
# Crear primer usuario
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "user1",
    "email": "user1@example.com"
  }'

# Crear segundo usuario
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "user2",
    "email": "user2@example.com"
  }'
```

2. Seguir usuario:
```bash
curl -X POST http://localhost:8080/api/v1/users/<USER1_ID>/follow/<USER2_ID>
```

3. Crear tweet:
```bash
curl -X POST http://localhost:8080/api/v1/tweets \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "<USER_ID>",
    "content": "¡Hola Mundo!"
  }'
```

4. Obtener timeline:
```bash
curl http://localhost:8080/api/v1/users/<USER_ID>/timeline?page=1&limit=10
```

### Consideraciones
- Todos los IDs son strings en formato MongoDB ObjectID
- Los timestamps están en UTC
- La paginación comienza en 1
- El timeline está ordenado por fecha de creación descendente