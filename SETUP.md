# NeonEx Framework - Quick Setup Guide

## สิ่งที่ทำเสร็จแล้ว ✅

### 1. โครงสร้างพื้นฐาน
- ✅ Clone neonexcore มาเป็น `core/` directory
- ✅ สร้าง main.go สำหรับ NeonEx Framework
- ✅ สร้าง go.mod และ dependencies
- ✅ สร้าง .env.example, .gitignore, Makefile

### 2. Modules ที่สร้างแล้ว
- ✅ **Frontend Module** - Template engine, asset management, theme system
- ✅ **CMS Module** - Pages, Posts, Categories, Tags, Media management
- ✅ **E-commerce Module** - Products, Cart, Orders, Payments, Coupons, Reviews

### 3. โครงสร้าง Directories
```
neonexframework/
├── core/                    # NeonEx Core (cloned)
├── modules/
│   ├── frontend/           # Frontend support
│   ├── cms/                # Content management
│   ├── ecommerce/          # E-commerce
│   └── admin/              # Admin panel (ยังไม่เสร็จ)
├── public/                 # Static assets
├── templates/              # HTML templates  
├── storage/                # Logs, cache, uploads
├── config/                 # Configuration
├── docs/                   # Documentation
└── tests/                  # Unit/Integration/E2E tests
```

## ปัญหาที่พบ ⚠️

### Core Update Strategy
NeonEx Framework ใช้ neonexcore เป็น dependency โดยจัดเก็บใน `/core` directory

**การ Update Core ในอนาคต:**
```bash
# ใช้ script อัตโนมัติ (แนะนำ)
make update-core

# หรือ manual
# Windows
powershell -ExecutionPolicy Bypass -File scripts/update-core.ps1

# Linux/Mac
bash scripts/update-core.sh
```

Script จะทำการ:
1. ✅ Backup core เดิม
2. ✅ Clone neonexcore version ใหม่
3. ✅ Clean ไฟล์ที่ไม่จำเป็น (docs, examples, .git)
4. ✅ Update dependencies
5. ✅ Run tests
6. ✅ Restore backup ถ้า test failed

### Internal Package Access (แก้ไขแล้ว)
Go ไม่อนุญาตให้เข้าถึง `internal/` packages จากภายนอก module

**วิธีแก้:**

### Option 1: แยก main.go ออกมาเป็น external package (แนะนำ)
สร้าง cmd/neonexframework/main.go:

```go
package main

import (
	"neonexcore/cmd/core"
	// Import your custom modules here
)

func main() {
	core.Run()
}
```

### Option 2: แก้ไขโครงสร้าง neonexcore
ย้าย internal packages ออกมาเป็น public packages:
- `internal/core` → `pkg/core`
- `internal/config` → `pkg/config`

### Option 3: Fork และแก้ไข neonexcore
Fork neonexcore repository และแก้ไขโครงสร้างให้รองรับการใช้งานจากภายนอก

## ขั้นตอนถัดไป 📝

### 1. แก้ไขปัญหา Internal Package (เลือก Option ข้างบน)

### 2. เพิ่ม Admin Panel Module
```bash
# สร้างโครงสร้าง admin module
modules/admin/
├── dashboard.go
├── settings.go
├── analytics.go
└── templates/
```

### 3. Register Modules ใน main.go
```go
import (
	"neonexframework/modules/frontend"
	"neonexframework/modules/cms"
	"neonexframework/modules/ecommerce"
)

// Register framework modules
core.ModuleMap["frontend"] = func() core.Module { return frontend.New() }
core.ModuleMap["cms"] = func() core.Module { return cms.New() }
core.ModuleMap["ecommerce"] = func() core.Module { return ecommerce.New() }
```

### 4. Setup Database
```bash
# สร้าง database
createdb neonexframework

# แก้ไข .env
cp .env.example .env
# Edit DB settings

# Run migrations
go run main.go
```

### 5. Build และ Run
```bash
# Build
make build

# Run
make run

# Development with hot reload
make dev
```

## Features ที่พร้อมใช้งาน

### CMS Features
- ✅ Page Management (CRUD)
- ✅ Blog Posts
- ✅ Categories & Tags
- ✅ Media Library
- ✅ SEO Management

### E-commerce Features
- ✅ Product Catalog
- ✅ Shopping Cart
- ✅ Order Management
- ✅ Payment Processing
- ✅ Coupon System
- ✅ Product Reviews

### Frontend Features
- ✅ Theme System
- ✅ Template Engine
- ✅ Asset Management (CSS/JS)
- ✅ Static File Serving

## API Endpoints

### CMS APIs
```
GET    /api/v1/cms/pages
GET    /api/v1/cms/pages/:id
POST   /api/v1/cms/pages
PUT    /api/v1/cms/pages/:id
DELETE /api/v1/cms/pages/:id

GET    /api/v1/cms/posts
...
```

### E-commerce APIs
```
GET    /api/v1/ecommerce/products
GET    /api/v1/ecommerce/products/:id
POST   /api/v1/ecommerce/cart/items
GET    /api/v1/ecommerce/orders
...
```

## Documentation

- [Getting Started](./docs/getting-started.md)
- [Module Development](./docs/module-development.md)
- [API Reference](./docs/api-reference.md)
- [Deployment Guide](./docs/deployment.md)

## ติดต่อ & Support

- GitHub: https://github.com/neonextechnologies/neonexframework
- Email: support@neonexframework.dev

---

**Note**: Framework นี้ยังอยู่ในช่วงพัฒนา (v0.1.0) แนะนำให้แก้ไขปัญหา internal package access ก่อนใช้งานจริง
