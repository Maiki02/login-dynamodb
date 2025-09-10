# 🤖 GitHub Copilot Instructions - Login DynamoDB Backend

## 🎯 Objetivo del Proyecto

Este repositorio tiene un **objetivo específico y limitado**: crear un backend de autenticación usando **Go + AWS Lambda + DynamoDB** con arquitectura cebolla.

### ✅ Funcionalidades PERMITIDAS:
- **Registro de usuarios** (POST /register)
- **Login de usuarios** (POST /login) 
- **Refresh de tokens** (POST /refresh-token)
- **Gestión JWT** (access + refresh tokens)
- **Validaciones de entrada**
- **Manejo de errores**

### ❌ Funcionalidades NO permitidas:
- Recuperación de contraseña
- Verificación de email
- Gestión de roles complejos
- CRUD de otras entidades
- Sistemas de notificaciones
- Integración con servicios externos (excepto AWS)

## 🏗️ Arquitectura Obligatoria: Patrón Cebolla

**IMPORTANTE**: Siempre seguir esta estructura de capas:

```
Main → Handlers → Services → Repositories → DynamoDB
```

### 📋 Reglas de Arquitectura:

1. **🎯 Main Layer** (`cmd/api/main.go`)
   - Solo configuración e inyección de dependencias
   - Detección Lambda vs Local
   - NO lógica de negocio

2. **🌐 Handler Layer** (`internal/handlers/`)
   - Solo manejo HTTP (request/response)
   - Validación de formato JSON
   - Llamadas directas al Service Layer
   - NO lógica de negocio

3. **⚙️ Service Layer** (`internal/services/`)
   - **NÚCLEO de la lógica de negocio**
   - Validaciones de negocio
   - Orquestación de operaciones
   - Transformación de datos
   - Interface para Repository

4. **📊 Repository Layer** (`internal/repositories/`)
   - Solo acceso a datos
   - Queries DynamoDB
   - Mapeo modelo ↔ DB
   - Implementa interfaces del Service

5. **🗄️ Database Layer** (DynamoDB)
   - **SOLO DynamoDB** (no MongoDB)

## 🚫 Reglas Estrictas de Código

### Dependency Inversion
```go
// ✅ CORRECTO: Service depende de interface
type UserService struct {
    userRepo UserRepository // Interface
}

// ❌ INCORRECTO: Service depende de implementación
type UserService struct {
    userRepo *DynamoUserRepository // Implementación concreta
}
```

### Separación de Capas
```go
// ❌ NUNCA: Handler llamando Repository directamente
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
    h.userRepo.Create(user) // ¡MAL!
}

// ✅ SIEMPRE: Handler → Service → Repository
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
    h.userService.CreateUser(req) // ¡BIEN!
}
```

## 📁 Estructura de Archivos

### Antes de crear UN NUEVO archivo:
1. **ANALIZAR** si la funcionalidad puede ir en un archivo existente
2. **PREGUNTAR** al usuario si no estás seguro
3. **REUTILIZAR** código existente cuando sea posible

### Naming Conventions:
- **Handlers**: `{entity}_handler.go` (ej: `user_handler.go`)
- **Services**: `{entity}_service.go` (ej: `user_service.go`)  
- **Repositories**: `{entity}_repository.go` (ej: `user_repository.go`)
- **Models**: `{entity}.go` (ej: `user.go`)

## 🛠️ Stack Tecnológico OBLIGATORIO

### ✅ Permitido:
- **Go 1.23.1+**
- **AWS Lambda Go** (`github.com/aws/aws-lambda-go`)
- **DynamoDB** (AWS SDK v2)
- **Gorilla Mux** (routing)
- **JWT** (`github.com/golang-jwt/jwt/v5`)
- **bcrypt** (hash contraseñas)
- **Estructuras existentes** del proyecto

### ❌ Prohibido:
- MongoDB (migrar todo a DynamoDB)
- ORM pesados
- Frameworks web complejos
- Librerías no necesarias

## 🔍 Mejores Prácticas Modernas

### Error Handling
```go
// ✅ Errores específicos y informativos
var (
    ErrUserNotFound = errors.New("usuario no encontrado")
    ErrInvalidCredentials = errors.New("credenciales inválidas")
)

// ✅ Wrap errors con contexto
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}
```

### Context Usage
```go
// ✅ Pasar context en operaciones de DB
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
    // usar ctx en operaciones DynamoDB
}
```

### Validation
```go
// ✅ Validar en Service Layer
func (s *UserService) CreateUser(req *request.CreateUserRequest) error {
    if err := s.validateCreateUserRequest(req); err != nil {
        return err
    }
    // continuar con lógica
}
```

### DynamoDB Patterns
```go
// ✅ Usar SDK v2 de AWS
import (
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
    "github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
)
```

## 🤔 Cuando Preguntar

**SIEMPRE preguntar antes de:**
1. Crear un nuevo archivo
2. Agregar funcionalidades fuera del scope
3. Cambiar la estructura de arquitectura
4. Modificar interfaces existentes
5. Agregar nuevas dependencias

**Preguntas típicas:**
- "¿Debería crear `{filename}.go` o integrar esto en `{existing_file}.go`?"
- "¿Esta funcionalidad está dentro del scope del proyecto?"
- "¿Prefieres que esto vaya en Service o Repository layer?"

## 🎯 Objetivos de Calidad

1. **Performance**: Optimizado para cold starts de Lambda
2. **Security**: JWT seguro, validación de inputs
3. **Maintainability**: Código limpio y bien documentado
4. **Testability**: Fácil de testear por capas
5. **Scalability**: Preparado para escalar en AWS

## 📝 Comentarios en Código

```go
// ✅ Comentarios útiles y concisos
// CreateUser creates a new user with company association
func (s *UserService) CreateUser(req *request.CreateUserRequest) error {

// ❌ Comentarios obvios
// This function creates a user
func (s *UserService) CreateUser(req *request.CreateUserRequest) error {
```

---

## 🚨 RECORDATORIO IMPORTANTE

**Este proyecto tiene un scope muy específico: sistema de autenticación básico**. 

No agregar funcionalidades adicionales sin preguntar primero. Mantener el código simple, limpio y enfocado en el objetivo principal.

**En caso de duda: PREGUNTAR SIEMPRE antes de proceder.**