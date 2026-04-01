# AMS Backend

Go-based backend for the Attendance Management System.

## 🚀 Features
- **RESTful API**: Built with Gin.
- **ORM**: GORM with PostgreSQL.
- **Auth**: JWT-based authentication and Role-Based Access Control (RBAC).
- **Documentation**: Swagger UI integrated.
- **Excel Support**: Bulk upload and export using `excelize`.

## 🛠 Setup
1. Ensure Go 1.22+ is installed.
2. Copy `.env.example` to `.env` and configure your database settings.
3. Run `go mod download`.
4. Start the server:
   ```bash
   go run cmd/server/main.go
   ```

## 📚 API Documentation
Once the server is running, visit:
`http://localhost:8080/swagger/index.html`

To update Swagger documentation:
```bash
go run github.com/swaggo/swag/cmd/swag init -g cmd/server/main.go
```

## 🗄 Project Structure
- `cmd/server`: Application entry point.
- `internal/models`: Database models.
- `internal/handlers`: HTTP request handlers.
- `internal/repository`: Data access layer.
- `internal/service`: Business logic layer.
- `internal/middleware`: Auth and role checks.
