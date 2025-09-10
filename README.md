# 🚀 Login DynamoDB - Backend API

Un backend moderno en **Go** diseñado para manejar autenticación de usuarios, construido para ejecutarse como **AWS Lambda Functions** y usando **DynamoDB** como base de datos principal.

## 📋 Descripción del Proyecto

Este proyecto es una migración y optimización de un sistema de backend existente que originalmente usaba MongoDB. La nueva versión está específicamente diseñada para:

- ✅ **Despliegue en AWS Lambda** - Serverless architecture
- ✅ **DynamoDB** como base de datos principal
- ✅ **Gestión completa de sesiones** con JWT
- ✅ **Arquitectura escalable** y moderna

## 🎯 Funcionalidades Principales

### 🔐 Gestión de Sesiones
- **Registro de usuarios** (`POST /register`)
- **Login de usuarios** (`POST /login`) 
- **Refresh de tokens** (`POST /refresh-token`)
- **Autenticación JWT** con tokens de acceso y refresh

### 🏗️ Arquitectura

Este proyecto implementa el **Patrón Cebolla (Onion Architecture)** para garantizar una arquitectura escalable, mantenible y testeable.

#### 🧅 Capas de la Arquitectura (de exterior a interior):

```
┌─────────────────────────────────────────┐
│               MAIN LAYER                │  ← Punto de entrada
├─────────────────────────────────────────┤
│              HANDLER LAYER              │  ← HTTP Handlers
├─────────────────────────────────────────┤
│              SERVICE LAYER              │  ← Lógica de Negocio
├─────────────────────────────────────────┤
│            REPOSITORY LAYER             │  ← Acceso a Datos
├─────────────────────────────────────────┤
│             DATABASE LAYER              │  ← DynamoDB
└─────────────────────────────────────────┘
```

**Características de cada capa:**

- **🎯 Main Layer** (`cmd/api/main.go`)
  - Punto de entrada de la aplicación
  - Configuración de Lambda vs Local
  - Inicialización de dependencias
  - Inyección de dependencias

- **🌐 Handler Layer** (`internal/handlers/`)
  - Manejo de requests/responses HTTP
  - Validación de entrada
  - Serialización/Deserialización JSON
  - Manejo de errores HTTP
  - **No contiene lógica de negocio**

- **⚙️ Service Layer** (`internal/services/`)
  - **Núcleo de la lógica de negocio**
  - Orquestación de operaciones
  - Validaciones de negocio
  - Transformación de datos
  - Independiente de frameworks externos

- **📊 Repository Layer** (`internal/repositories/`)
  - Abstracción de acceso a datos
  - Implementación de interfaces
  - Queries y operaciones CRUD
  - Mapeo entre modelos y BD

- **🗄️ Database Layer** (DynamoDB)
  - Almacenamiento persistente
  - Esquemas y índices
  - Consistencia de datos

#### 🔄 Flujo de Dependencias

```
Main → Handlers → Services → Repositories → Database
  ↑                ↓           ↓              ↓
  └── Interfaces ←─┴─────────←─┴──────────────┘
```

**Ventajas del Patrón Cebolla:**
- ✅ **Testabilidad**: Cada capa puede testearse independientemente
- ✅ **Mantenibilidad**: Cambios en capas externas no afectan el core
- ✅ **Escalabilidad**: Fácil agregar nuevas funcionalidades
- ✅ **Flexibilidad**: Intercambiar implementaciones fácilmente
- ✅ **Separación de responsabilidades**: Cada capa tiene un propósito específico

#### 📋 Principios de la Arquitectura

**1. 🎯 Dependency Inversion**
```go
// Service depende de interface, no de implementación
type UserService struct {
    userRepo UserRepository  // Interface, no implementación concreta
}

type UserRepository interface {
    Create(user *models.User) error
    GetByEmail(email string) (*models.User, error)
}
```

**2. 🔒 Encapsulación de Capas**
```go
// ❌ MAL: Handler accediendo directamente al Repository
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    user := parseRequest(r)
    h.userRepo.Create(user) // ¡Saltar el Service Layer!
}

// ✅ BIEN: Handler usa Service Layer
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    req := parseRequest(r)
    err := h.userService.CreateUser(req) // Service maneja la lógica
}
```

**3. 🚫 Regla de Dependencias**
- **Las capas internas NO conocen las externas**
- **Solo se comunican a través de interfaces**
- **El flujo de control y datos va hacia adentro**

**4. 🎪 Inyección de Dependencias en main.go**
```go
func main() {
    // Inicializar desde la capa más interna hacia afuera
    db := dynamo.NewConnection()
    userRepo := repositories.NewUserRepository(db)
    userService := services.NewUserService(userRepo)
    userHandler := handlers.NewUserHandler(userService)
    
    // Configurar rutas
    router.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
}
```

## 🛠️ Stack Tecnológico

### Backend
- **Go 1.23.1** - Lenguaje principal
- **Gorilla Mux** - Router HTTP
- **AWS Lambda Go** - Para funciones serverless
- **JWT (golang-jwt/jwt/v5)** - Manejo de tokens
- **bcrypt** - Hash de contraseñas

### Base de Datos
- **DynamoDB** - Base de datos NoSQL de AWS
- **MongoDB Driver** (en proceso de migración)

### AWS Services
- **AWS Lambda** - Compute serverless
- **API Gateway** - Gestión de APIs
- **DynamoDB** - Base de datos
- **IAM** - Gestión de permisos

## 📁 Estructura del Proyecto

### 🏛️ Organización por Capas de Arquitectura Cebolla

```
├── cmd/                          # 🎯 MAIN LAYER
│   ├── api/
│   │   └── main.go              # ← Punto de entrada y configuración
│   ├── middlewares/
│   │   └── middlewares.go       # ← Middlewares HTTP
│   └── routes/
│       └── routes.go            # ← Definición de rutas y DI
│
├── internal/                     # 🌐 CAPAS INTERNAS
│   ├── handlers/                # ← HANDLER LAYER
│   │   ├── session.go          # ← Handlers de autenticación
│   │   ├── user.go             # ← Handlers de usuario
│   │   └── *.go                # ← Otros handlers
│   │
│   ├── services/                # ← SERVICE LAYER (Core Business)
│   │   ├── session.go          # ← Lógica de autenticación
│   │   ├── user.go             # ← Lógica de usuario
│   │   └── *.go                # ← Otros servicios
│   │
│   ├── repositories/            # ← REPOSITORY LAYER
│   │   ├── user.go             # ← Acceso a datos de usuario
│   │   └── *.go                # ← Otros repositorios
│   │
│   ├── models/                  # ← DOMAIN MODELS
│   │   ├── user.go             # ← Entidades del dominio
│   │   └── *.go                # ← Otros modelos
│   │
│   └── db/                      # 🗄️ DATABASE LAYER
│       └── dynamo.go           # ← Conexión DynamoDB (a implementar)
│
├── pkg/                         # 🛠️ UTILIDADES COMPARTIDAS
│   ├── jwt/
│   │   └── jwt.go              # ← Utilidades JWT
│   ├── request/
│   │   └── *.go                # ← DTOs de entrada
│   ├── response/
│   │   └── *.go                # ← DTOs de salida
│   ├── validations/
│   │   └── *.go                # ← Validaciones reutilizables
│   └── consts/
│       └── *.go                # ← Constantes de la aplicación
│
├── go.mod                       # Dependencias Go
└── go.sum                       # Checksums
```

### 🔄 Flujo de Datos por Capas

#### Ejemplo: Login de Usuario

```
1. 🌐 Handler Layer (session.go)
   ├── Recibe HTTP Request
   ├── Valida formato JSON
   ├── Extrae datos del request
   └── Llama al Service Layer
        ↓
2. ⚙️ Service Layer (session.go)
   ├── Valida lógica de negocio
   ├── Hashea/Compara contraseñas
   ├── Genera tokens JWT
   └── Llama al Repository Layer
        ↓
3. 📊 Repository Layer (user.go)
   ├── Construye queries DynamoDB
   ├── Ejecuta operaciones CRUD
   ├── Mapea resultados a modelos
   └── Retorna datos al Service
        ↓
4. 🗄️ Database Layer (DynamoDB)
   ├── Ejecuta query en tabla
   ├── Aplica índices y filtros
   └── Retorna datos raw
```

## 🚀 Configuración y Despliegue

### Prerrequisitos
- Go 1.23.1+
- AWS CLI configurado
- Cuenta de AWS con permisos para Lambda y DynamoDB

### Variables de Entorno
```bash
# Para desarrollo local
PORT=9000
MONGO_URI=mongodb://localhost:27017  # Temporal durante migración
JWT_SECRET=tu_jwt_secret_key

# Para AWS Lambda
LAMBDA_SERVER_PORT=true  # Indica ejecución en Lambda
AWS_REGION=us-east-1
DYNAMODB_TABLE_USERS=users
```

### Instalación Local

1. **Clonar el repositorio**
```bash
git clone <repository-url>
cd Login-DynamoDB
```

2. **Instalar dependencias**
```bash
go mod download
```

3. **Configurar variables de entorno**
```bash
cp .env.example .env
# Editar .env con tus configuraciones
```

4. **Ejecutar en modo desarrollo**
```bash
go run cmd/api/main.go
```

### Despliegue en AWS Lambda

1. **Compilar para Linux**
```bash
GOOS=linux GOARCH=amd64 go build -o bootstrap cmd/api/main.go
```

2. **Crear archivo ZIP**
```bash
zip lambda-deployment.zip bootstrap
```

3. **Desplegar usando AWS CLI**
```bash
aws lambda create-function \
  --function-name login-dynamodb-api \
  --runtime provided.al2 \
  --role arn:aws:iam::ACCOUNT:role/lambda-execution-role \
  --handler bootstrap \
  --zip-file fileb://lambda-deployment.zip
```

## 📡 API Endpoints

### Autenticación

#### Registro de Usuario
```http
POST /api/register
Content-Type: application/json

{
  "name": "Juan",
  "last_name": "Pérez",
  "email": "juan@example.com",
  "password": "password123",
  "company_name": "Mi Empresa"
}
```

#### Login
```http
POST /api/login
Content-Type: application/json

{
  "email": "juan@example.com",
  "password": "password123"
}
```

**Respuesta:**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "token_type": "Bearer",
    "expires_in": 3600
  }
}
```

#### Refresh Token
```http
POST /api/refresh-token
Authorization: Bearer <refresh_token>
```

## 🔧 Estado del Proyecto

### ✅ Completado
- [x] Estructura base del proyecto
- [x] Configuración para AWS Lambda
- [x] Handlers de sesión básicos
- [x] Modelos de usuario
- [x] Sistema JWT
- [x] Validaciones básicas

### 🚧 En Desarrollo
- [ ] **Migración completa a DynamoDB**
  - [ ] Reemplazar MongoDB por DynamoDB
  - [ ] Adaptar repositorios para DynamoDB
  - [ ] Configurar índices y queries
- [ ] **Testing**
  - [ ] Unit tests para servicios
  - [ ] Integration tests para handlers
  - [ ] Tests de DynamoDB
- [ ] **Documentación**
  - [ ] Swagger/OpenAPI documentation
  - [ ] Postman collection

### 🎯 Próximas Funcionalidades
- [ ] Recuperación de contraseña
- [ ] Verificación de email
- [ ] Rate limiting
- [ ] Logs estructurados
- [ ] Métricas y monitoring

## 🤝 Contribución

Este proyecto está en desarrollo activo. Para contribuir:

1. Fork el proyecto
2. Crea una rama para tu feature (`git checkout -b feature/AmazingFeature`)
3. Commit tus cambios (`git commit -m 'Add some AmazingFeature'`)
4. Push a la rama (`git push origin feature/AmazingFeature`)
5. Abre un Pull Request

## 📄 Licencia

Este proyecto está bajo la Licencia MIT - ver el archivo [LICENSE](LICENSE) para más detalles.

## 👨‍💻 Autor

**Tu Nombre** - [GitHub](https://github.com/Maiki02)

---

## 🎯 Objetivos del Proyecto

Este backend está siendo desarrollado con los siguientes objetivos en mente:

1. **Serverless First**: Diseñado desde cero para funcionar óptimamente en AWS Lambda
2. **Escalabilidad**: Uso de DynamoDB para manejar cargas de trabajo variables
3. **Seguridad**: Implementación robusta de JWT y validaciones
4. **Mantenibilidad**: Clean Architecture para facilitar el mantenimiento
5. **Performance**: Optimizado para cold starts mínimos en Lambda

---

*Este README será actualizado conforme el proyecto evolucione hacia su versión final con DynamoDB.*
