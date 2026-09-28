# 1. Requisitos Previos e Instalación de Herramientas

En este documento se explica cómo instalar todas las herramientas necesarias para ejecutar Meloop. Sigue los pasos para tu sistema operativo.

**Tiempo estimado:** 30-45 minutos

---

## Resumen de Herramientas Necesarias

| Herramienta | Versión | Uso | SO |
|------------|---------|-----|-----|
| **Git** | Latest | Control de versiones | Windows, Linux |
| **Go** | 1.26.6 | Microservicios backend | Windows, Linux |
| **Python** | 3.11 | Servicio de ML | Windows, Linux |
| **Node.js** | 18+ LTS | Backend NestJS | Windows, Linux |
| **Flutter** | 3.13.0+ | Aplicación móvil | Windows, Linux |
| **Docker** | Latest | Contenedorización | Windows, Linux |
| **Git Bash** | Included | Terminal (Windows only) | Windows |

---

## Instalación en Windows 11

### Git

1. Descarga desde [git-scm.com](https://git-scm.com/download/win)
2. Ejecuta el instalador
3. En la opción "Use Git from the Windows Command Prompt", selecciona: **"Git from the command line and also from 3rd-party software"**
4. Completa la instalación con opciones por defecto

**Validar instalación:**
```powershell
git --version
# Output: git version 2.x.x
```

---

### Go 1.26.6

1. Descarga desde [go.dev/dl](https://go.dev/dl) (búsca go1.26.6.windows-amd64.msi)
2. Ejecuta el instalador `.msi`
3. Sigue las instrucciones (instalará en `C:\Program Files\Go`)

**Validar instalación:**
```powershell
go version
# Output: go version go1.26.6 windows/amd64
```

**Configurar GOPATH (opcional pero recomendado):**
```powershell
# Abrir PowerShell como Administrador
setx GOPATH "$env:USERPROFILE\go"
setx GOBIN "$env:USERPROFILE\go\bin"

# Recargar PowerShell para aplicar cambios
```

---

### Python 3.11

1. Descarga desde [python.org](https://www.python.org/downloads/) (Python 3.11.x)
2. Ejecuta el instalador
3. **IMPORTANTE:** Marca "Add Python 3.11 to PATH" al inicio
4. Elige "Install Now" para instalación rápida

**Validar instalación:**
```powershell
python --version
# Output: Python 3.11.x

pip --version
# Output: pip 24.x.x from C:\...
```

**Crear un alias (opcional):**
```powershell
# Si prefieres usar 'python3' en lugar de 'python'
New-Alias -Name python3 -Value python -Scope CurrentUser
```

---

### Node.js 18+ LTS

1. Descarga desde [nodejs.org](https://nodejs.org/) (LTS de 20.x o 22.x)
2. Ejecuta el instalador `.msi`
3. Marca "Add to PATH" durante la instalación
4. Completa la instalación

**Validar instalación:**
```powershell
node --version
# Output: v20.x.x o v22.x.x

npm --version
# Output: 10.x.x o superior
```

---

### Flutter SDK 3.13.0+

1. Descarga desde [flutter.dev/docs/get-started/install/windows](https://flutter.dev/docs/get-started/install/windows)
2. Extrae el archivo en una carpeta sin espacios, por ejemplo:
   ```
   C:\src\flutter
   ```
3. Agrega Flutter al PATH:
   - Abre "Variables de Entorno" (busca "environment" en Inicio)
   - Haz clic en "Variables de Entorno"
   - En "Variables del sistema", haz clic en "Editar PATH"
   - Agrega: `C:\src\flutter\bin`

**Validar instalación:**
```powershell
flutter --version
# Output: Flutter 3.13.x

dart --version
# Output: Dart 3.1.x
```

**Ejecutar doctor:**
```powershell
flutter doctor
# Esto verificará todas las dependencias
```

**Nota:** Si `flutter doctor` reporta problemas con Android SDK, puedes ignorarlos por ahora si solo testearás en web/Windows.

---

### Docker Desktop

1. Descarga desde [docker.com/products/docker-desktop](https://www.docker.com/products/docker-desktop)
2. Ejecuta el instalador
3. Durante la instalación, marca:
   - ✅ "Use WSL 2 instead of Hyper-V"
   - ✅ "Install required Windows components for WSL 2"
4. Reinicia tu computadora cuando se solicite
5. Abre Docker Desktop después del reinicio

**Validar instalación:**
```powershell
docker --version
# Output: Docker version 24.x.x

docker-compose --version
# Output: Docker Compose version 2.20.x
```

**Verificar que Docker está ejecutándose:**
```powershell
docker ps
# Debería mostrar una tabla vacía, no errores
```

---

## Instalación en Linux (Ubuntu 22.04+)

### Git

```bash
sudo apt update
sudo apt install -y git

git --version
```

### Go 1.26.6

```bash
# Descarga
wget https://go.dev/dl/go1.26.6.linux-amd64.tar.gz

# Extrae en /usr/local
sudo tar -C /usr/local -xzf go1.26.6.linux-amd64.tar.gz

# Agrega al PATH (en ~/.bashrc o ~/.zshrc)
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Validar
go version
```

### Python 3.11

```bash
sudo apt install -y python3.11 python3.11-venv python3.11-dev

python3.11 --version
pip3.11 --version
```

### Node.js 18+ LTS

```bash
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs

node --version
npm --version
```

### 5️⃣ Flutter SDK 3.13.0+

```bash
# Descarga
git clone https://github.com/flutter/flutter.git ~/flutter -b stable

# Agrega al PATH
echo 'export PATH="$PATH:$HOME/flutter/bin"' >> ~/.bashrc
source ~/.bashrc

# Validar
flutter --version
dart --version
flutter doctor
```

### 6️⃣ Docker

```bash
sudo apt update
sudo apt install -y docker.io docker-compose

# Agrega usuario al grupo docker
sudo usermod -aG docker $USER
newgrp docker

# Validar
docker --version
docker-compose --version
docker ps
```

---

## Validación Completa

Una vez instaladas todas las herramientas, ejecuta este script de validación:

### Windows (PowerShell)
```powershell
Write-Host "=== VALIDACIÓN DE HERRAMIENTAS ===" -ForegroundColor Green

Write-Host "`nGit:"
git --version

Write-Host "`nGo:"
go version

Write-Host "`nPython:"
python --version

Write-Host "`nNode.js:"
node --version
npm --version

Write-Host "`nFlutter:"
flutter --version
dart --version

Write-Host "`nDocker:"
docker --version
docker-compose --version

Write-Host "`n✅ Todas las herramientas están instaladas correctamente"
```

### Linux (Bash)
```bash
echo "=== VALIDACIÓN DE HERRAMIENTAS ==="

echo -e "\nGit:"
git --version

echo -e "\nGo:"
go version

echo -e "\nPython:"
python3.11 --version

echo -e "\nNode.js:"
node --version
npm --version

echo -e "\nFlutter:"
flutter --version
dart --version

echo -e "\nDocker:"
docker --version
docker-compose --version

echo -e "\n✅ Todas las herramientas están instaladas correctamente"
```

---

## Configurar Git (Importante)

Si es tu primera vez usando Git, configura tu identidad:

```bash
git config --global user.name "Tu Nombre"
git config --global user.email "tu.email@ejemplo.com"

# Verificar
git config --list | grep user
```

---

## 📁 Estructura de Carpetas Recomendada

Se recomienda crear una carpeta para los proyectos:

### Windows
```powershell
mkdir "$env:USERPROFILE\Projects"
cd "$env:USERPROFILE\Projects"
```

### Linux
```bash
mkdir -p ~/Projects
cd ~/Projects
```

---

## 🚀 Próximo Paso

Una vez que todas las herramientas estén instaladas y validadas:

👉 **Ve a [2. Configuración del Entorno →](02-environment-setup.md)**

---

## Problemas Comunes

### El comando no se encuentra (Windows)
- **Causa:** La herramienta no está en PATH
- **Solución:** 
  1. Reinicia PowerShell/CMD después de instalar
  2. Reinicia tu computadora si el problema persiste
  3. Verifica manualmente la ruta en "Variables de Entorno"

### Python: "comando no encontrado"
- **Solución en Windows:** Usa `python` en lugar de `python3`
- **Solución en Linux:** Asegúrate de instalar `python3.11` específicamente

### Flutter doctor muestra errores
- Es normal si no tienes todas las plataformas (Android, iOS) configuradas
- Para desarrollo web/Windows, puedes ignorar las advertencias de Android/iOS

### Docker no se inicia en Windows
- Asegúrate de tener WSL 2 instalado
- Reinicia Docker Desktop
- Si persiste, reinicia tu computadora

---

**Última actualización:** Septiembre 2024  
**Versión:** 1.0
