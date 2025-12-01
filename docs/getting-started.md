# Getting Started with NeonEx Framework

## Introduction

NeonEx Framework เป็น full-stack Go framework ที่ออกแบบมาเพื่อการพัฒนาแอปพลิเคชันเว็บที่ทันสมัย รวดเร็ว และครบครัน

## Prerequisites

ก่อนเริ่มต้น คุณต้องติดตั้งสิ่งเหล่านี้:

- **Go 1.21 หรือสูงกว่า** - [ดาวน์โหลด](https://go.dev/dl/)
- **PostgreSQL 14+** - [ดาวน์โหลด](https://www.postgresql.org/download/)
- **Git** - สำหรับ version control

## Installation

### 1. Clone Repository

```bash
git clone https://github.com/neonextechnologies/neonexframework.git
cd neonexframework
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Setup Environment

สร้างไฟล์ `.env` จาก template:

```bash
cp .env.example .env
```

แก้ไขไฟล์ `.env` ตามการตั้งค่าของคุณ:

```env
# Database
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=neonexframework
DB_USER=postgres
DB_PASSWORD=your_password

# Application
APP_PORT=8080
JWT_SECRET=your-secret-key
```

### 4. Create Database

สร้าง database ใน PostgreSQL:

```sql
CREATE DATABASE neonexframework;
```

### 5. Run Application

```bash
go run main.go
```

หรือใช้ Makefile:

```bash
make run
```

## First Steps

### Access the Application

เมื่อ server เริ่มทำงาน คุณสามารถเข้าถึงได้ที่:

- **Homepage**: http://localhost:8080
- **Admin Panel**: http://localhost:8080/admin
- **API**: http://localhost:8080/api/v1
- **Health Check**: http://localhost:8080/health

### Default Credentials

สำหรับ Admin Panel:

- **Email**: admin@example.com
- **Password**: admin123

### Test API

ทดสอบ API ด้วย curl:

```bash
# Register a new user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123"
  }'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

## Project Structure

```
neonexframework/
├── core/                 # NeonEx Core (framework core)
├── modules/              # Application modules
│   ├── cms/             # Content Management System
│   ├── ecommerce/       # E-commerce functionality
│   ├── admin/           # Admin panel
│   └── frontend/        # Frontend support
├── public/              # Static assets
├── templates/           # HTML templates
├── storage/             # Storage (logs, cache, uploads)
├── config/              # Configuration files
├── docs/                # Documentation
└── main.go             # Application entry point
```

## Development

### Hot Reload

สำหรับการพัฒนา ใช้ hot reload:

```bash
make dev
```

หรือ:

```bash
air
```

### Running Tests

```bash
make test
```

### Code Generation

สร้าง module ใหม่:

```bash
go run main.go make:module blog
```

สร้าง model:

```bash
go run main.go make:model Article
```

## Next Steps

- [Module Development](./module-development.md)
- [CMS Guide](./cms-guide.md)
- [E-commerce Setup](./ecommerce-setup.md)
- [API Reference](./api-reference.md)
- [Deployment](./deployment.md)

## Getting Help

- **Documentation**: [docs/](../docs)
- **Issues**: [GitHub Issues](https://github.com/neonextechnologies/neonexframework/issues)
- **Email**: support@neonexframework.dev
