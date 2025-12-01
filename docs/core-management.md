# Core Management Guide

## Overview

NeonEx Framework ใช้ [neonexcore](https://github.com/neonextechnologies/neonexcore) เป็น core dependency โดยจัดเก็บไว้ใน `/core` directory

## What's Cleaned

เพื่อให้ NeonEx Framework มีขนาดเล็กและเหมาะสมกับการใช้งาน ได้มีการลบไฟล์ที่ไม่จำเป็นออกจาก neonexcore:

### ไฟล์และ Directories ที่ถูกลบ:
- ✅ `docs/` - เอกสารทั้งหมด (58+ files)
- ✅ `examples/` - ตัวอย่างโค้ด
- ✅ `.git/` - Git repository
- ✅ `.gitbook.yaml` - GitBook config
- ✅ `.air.toml` - Hot reload config
- ✅ `dev.sh` - Development script
- ✅ `.env.example` - Environment template
- ✅ `LICENSE` - License file (มีใน framework root)
- ✅ `Makefile` - Build config (มีใน framework root)
- ✅ Original `README.md` - เปลี่ยนเป็น core-specific README

### ไฟล์ที่เก็บไว้:
- ✅ `internal/` - Core framework code
- ✅ `modules/` - Built-in modules (user, admin)
- ✅ `pkg/` - Public packages
- ✅ `go.mod`, `go.sum` - Go modules
- ✅ `main.go` - Entry point (ไม่ได้ใช้โดยตรง)
- ✅ `.gitattributes`, `.gitignore` - Git config

## Statistics

**Before cleaning:**
- 220+ files
- 74 directories

**After cleaning:**
- ~160 files
- ~61 directories

**Space saved:** ~27% reduction in file count

## Automatic Updates

### Using Make Command

```bash
make update-core
```

### Using Scripts Directly

**Windows (PowerShell):**
```powershell
powershell -ExecutionPolicy Bypass -File scripts/update-core.ps1
```

**Linux/Mac (Bash):**
```bash
bash scripts/update-core.sh
```

## Update Process

The update script performs these steps:

1. **Backup** - สำรองไฟล์ core เดิมไปที่ `core.backup/`
2. **Clone** - Clone neonexcore ใหม่จาก GitHub
3. **Clean** - ลบไฟล์ที่ไม่จำเป็นออก
4. **Create README** - สร้าง README ใหม่สำหรับ core
5. **Update Dependencies** - รัน `go mod tidy`
6. **Test** - รัน tests ทั้งหมด
7. **Restore or Complete**:
   - ถ้า tests ผ่าน: ถามว่าจะลบ backup หรือไม่
   - ถ้า tests ไม่ผ่าน: restore จาก backup อัตโนมัติ

## Manual Update

ถ้าต้องการ update manually:

```bash
# 1. Backup
mv core core.backup

# 2. Clone
git clone https://github.com/neonextechnologies/neonexcore.git core

# 3. Clean
cd core
rm -rf docs examples .gitbook.yaml .git README.md
rm -f .air.toml dev.sh .env.example LICENSE Makefile

# 4. Create README
cat > README.md << 'EOF'
# NeonEx Core
See ../core/README.md for update instructions
EOF

cd ..

# 5. Test
go mod tidy
go test ./...

# 6. If success, remove backup
rm -rf core.backup
```

## Version Tracking

Current version: **v0.1.0**

Check for updates:
- Repository: https://github.com/neonextechnologies/neonexcore
- Releases: https://github.com/neonextechnologies/neonexcore/releases

## Important Notes

### ⚠️ Do NOT Modify Core Files

Core directory เป็น dependency ไม่ควรแก้ไขโดยตรง เพราะ:
- จะหายเมื่อ update core
- ทำให้ maintenance ยาก
- อาจทำให้ merge conflicts

### ✅ Instead, Use Wrapper Packages

สร้าง wrapper packages ใน `neonexframework/pkg/`:

```go
// neonexframework/pkg/myutils/utils.go
package myutils

import "neonexcore/pkg/somepackage"

func MyCustomFunction() {
    // Extend core functionality here
}
```

### 📝 Track Your Changes

ถ้าต้องการแก้ไข core:
1. Fork neonexcore
2. แก้ไขใน fork
3. Update core URL ใน update script

## Troubleshooting

### Update ไม่สำเร็จ

```bash
# Restore manual
rm -rf core
mv core.backup core
```

### Conflicts หลัง Update

```bash
# Check differences
diff -r core.backup core

# Revert specific files if needed
cp core.backup/some/file core/some/file
```

### Want Original Docs

Original documentation อยู่ที่:
- https://github.com/neonextechnologies/neonexcore/tree/main/docs

หรือ clone เพื่อดู:
```bash
git clone https://github.com/neonextechnologies/neonexcore.git temp-core
cd temp-core/docs
# Read docs...
cd ../..
rm -rf temp-core
```

## Questions?

- Framework Issues: https://github.com/neonextechnologies/neonexframework/issues
- Core Issues: https://github.com/neonextechnologies/neonexcore/issues
- Email: support@neonexframework.dev
