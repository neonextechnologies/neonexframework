# NeonEx Framework Documentation

ยินดีต้อนรับสู่เอกสารของ NeonEx Framework!

## 📚 Table of Contents

### Getting Started
- [Installation](./getting-started.md)
- [Quick Start](./quick-start.md)
- [Project Structure](./project-structure.md)
- [Core Management](./core-management.md)

### Core Concepts
- [Module System](./core/modules.md)
- [Dependency Injection](./core/dependency-injection.md)
- [Repository Pattern](./core/repository-pattern.md)

### Modules

#### CMS Module
- [Overview](./modules/cms/overview.md)
- [Pages Management](./modules/cms/pages.md)
- [Blog Posts](./modules/cms/posts.md)
- [Media Library](./modules/cms/media.md)

#### E-commerce Module
- [Overview](./modules/ecommerce/overview.md)
- [Products](./modules/ecommerce/products.md)
- [Shopping Cart](./modules/ecommerce/cart.md)
- [Orders](./modules/ecommerce/orders.md)
- [Payments](./modules/ecommerce/payments.md)

#### Admin Panel
- [Dashboard](./modules/admin/dashboard.md)
- [User Management](./modules/admin/users.md)
- [Settings](./modules/admin/settings.md)

#### Frontend
- [Theme System](./modules/frontend/themes.md)
- [Templates](./modules/frontend/templates.md)
- [Assets](./modules/frontend/assets.md)

### Development
- [Creating Modules](./development/creating-modules.md)
- [Testing](./development/testing.md)
- [Debugging](./development/debugging.md)
- [Best Practices](./development/best-practices.md)

### API Reference
- [REST API](./api/rest.md)
- [GraphQL](./api/graphql.md)
- [Authentication](./api/authentication.md)

### Deployment
- [Production Setup](./deployment/production.md)
- [Docker](./deployment/docker.md)
- [Environment Variables](./deployment/environment.md)

## 🚀 Quick Links

- [GitHub Repository](https://github.com/neonextechnologies/neonexframework)
- [NeonEx Core](https://github.com/neonextechnologies/neonexcore)
- [Examples](../examples/)

## 📖 About

NeonEx Framework เป็น full-stack Go framework ที่สร้างจาก NeonEx Core ออกแบบมาเพื่อพัฒนา:

- 🎨 **CMS** - Content Management System
- 👑 **Admin Panel** - ระบบจัดการหลังบ้าน
- 🛒 **E-commerce** - ระบบขายสินค้าออนไลน์
- 📱 **APIs** - RESTful & GraphQL APIs

## 🎯 Features

- ⚡ **High Performance** - สร้างด้วย Go เพื่อความเร็วสูงสุด
- 🎨 **Modular** - ระบบ module ที่ยืดหยุ่น
- 🔐 **Secure** - มีระบบความปลอดภัยในตัว
- 📦 **Complete** - มีทุกอย่างที่ต้องการในกล่องเดียว
- 🚀 **Production Ready** - พร้อมใช้งานจริง

## 💡 Examples

### Creating a Simple Page

```go
page := &cms.Page{
    Title:   "About Us",
    Slug:    "about",
    Content: "<h1>Welcome</h1><p>About our company...</p>",
    Status:  "published",
}
pageService.Create(ctx, page)
```

### Adding Products

```go
product := &ecommerce.Product{
    Name:  "Premium T-Shirt",
    SKU:   "TS-001",
    Price: 599.00,
    Stock: 100,
}
productService.Create(ctx, product)
```

### Managing Orders

```go
order := orderService.CreateFromCart(ctx, cart)
orderService.ProcessPayment(ctx, order.ID, paymentMethod)
```

## 🤝 Contributing

เราต้อนรับการมีส่วนร่วม! โปรดดูที่ [Contributing Guide](./CONTRIBUTING.md)

## 📄 License

MIT License - ดูรายละเอียดใน [LICENSE](../LICENSE)

## 💬 Support

- 📧 Email: support@neonexframework.dev
- 💬 Discussions: [GitHub Discussions](https://github.com/neonextechnologies/neonexframework/discussions)
- 🐛 Issues: [GitHub Issues](https://github.com/neonextechnologies/neonexframework/issues)
