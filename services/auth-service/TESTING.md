# Pruebas funcionales en Kubernetes

Con el port-forward del API Gateway activo:

```powershell
kubectl port-forward -n student-acofre service/meloop-api-gateway-service 8080:8080
```

En otra ventana de Windows PowerShell 5.1, desde la raíz del repositorio:

```powershell
.\services\auth-service\test-auth-service.ps1
```

Para usar otra URL base:

```powershell
.\services\auth-service\test-auth-service.ps1 -BaseUrl "http://localhost:8080"
```

El script crea un usuario con nombre y correo únicos. No imprime contraseñas,
tokens ni cuerpos de respuesta. Informa pruebas aprobadas, fallidas y omitidas.
La recuperación de una cuenta registrada y el restablecimiento exitoso se
omiten porque requieren correo y un token que no puede obtenerse por una
interfaz de prueba segura. Cada ejecución deja la cuenta de prueba creada,
porque Auth Service no expone un endpoint para eliminarla; ejecútalo solo en un
entorno de pruebas autorizado.
