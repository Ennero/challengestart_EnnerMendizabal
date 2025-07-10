# API de Mocks para Servicios REST

## Objetivos

### Objetivo General

- Implementar y diseñar una API REST robusta en Go con Fiber que permita al usuario gestionar y simular dinámicamente respuestas personalizadas para endpoints de servicios externos, facilitando las pruebas en el desarrollo.

### Objetivos Específicos

- Definir Extructura de Mocks: Establceder una estructura clara y flexible para almacenar las configuraciones de los mocks, incluyendo ruta, metodo HTTP, parámetros de URL, headers, cuepro de la solicitud, código de estado tipo de contenido y contenido de la respuesta.

- Implementar un Módulo de Configuración de Mocks: Desarrollar los endpoins para la gestión completa de las configuraciones de mocks

- Desarrollar un enrutamiento dinámico de Mocks: Crear un mecanismo que intercepte solcitudes entrantes, compare características con las configuraciones de los mocks almacenados y devuelva la respuesta predefinida correspondiente.

- Asegurar la Calidad del Código y Documentación: Aplicar buenas prácticas de desarrollo y proporcionar una documentación clara para la instalación y suo de la API.

## Alcances del Sistema
Este programa se encargará de proporcionar una API para la simulación de servicios REST cubriendo las siguiente funcionalidades:

### Gestión de Configuración de Mocks
- **Creación de Mocks** `POST /configure-mock`
  - Permite registrar una nueva configuración de mock.
  
  - Sporta la definición de route, method, queryParams, headers y requerstBody.

  - Permite la especifiación del statusCode, contentType y body de la respuesta simulada.
  
- **Listado de Mocks** `GET /configure-mock`
  - Devuelve una lista de todas las configuraciones de mocks activas.
- **Eliminación de Mocks** `DELETE /configure-mock/:id)`
  - Permite eliminar una configuración de mock específica utilizando su ID único.

### Ejecución de Mocks
- **Enrutamiento Genérico**
  - La API interceptará cualquier solicitud HTTP que no corresponda a sus rutas de administración.
-  **Coincidencia de Mocks**
   -  A partir de cada solicitud entrante, el sistema buscará la configuración de mock más apropiada basándose en:

      -  La ruta exacta.

      -  Método HTTP.

      -  Parámetros de URL.

      -  Encabezados.
  
      -  Cuerpo de la solicitud.
  
-  **Devolución de Respuesta**
   -  Si se encuentra un mock que coincida, la API responderá con el ``statusCode``, ``contentType`` y ``body`` definidos en la configuración del mock.
-  
   -  Si no se encuentra ninguna coincidencia, la API devolverá un ``404 Not Found``.




### Gestión de Configuración de Mocks



## Requisitos del Sistema

### Hardware
- **Memoria RAM:** 512 MB (Recomendado 1 GB+ para desarrollo).

- **Espacio en Disco:** 200 MB libre (para el código fuente, entorno Go, y dependencias).

- **Procesador:** Cualquier CPU medianamente moderna.

### Software
- **Sistema Operativo:** Compatible con Go (Linux, macOS, Windows).

- **Go:** Versión 1.21 o superior.

- **Git:** Para clonar el repositorio.

- **Herramientas de Cliente HTTP:** cURL, Postman, Insomnia o similar para interactuar con la API.

- **IDE/Editor:** Visual Studio Code (recomendado) con extensiones para Go, u otro editor/IDE de preferencia.

- **Terminal/Consola:** Para compilar y ejecutar el backend.

## Instalación y Ejecución
 1. **Clonar el Repositorio:**
    ```bash
    git clone https://github.com/Ennero/challengestart_EnnerMendizabal.git 
    
    cd mock-api/backend
    ```
2.  **Descargar dependencias:**
    ```bash
    go mod tidy
    ```
3.  **Ejecutar la aplicación:**
    ```bash
    go run main.go
    ```
La API se ejecutará en `http://localhost:3000` de forma predeterminada, pero de la siguiente forma se puede inciar con otro puerto:

```bash
# Para linux o mac
PORT=8080 go run backend/main.go

# Para CMD de windows
set PORT=8080 && go run backend/main.go

# Para PowerShell de Windows
$env:PORT="8080"; go run backend/main.go
```


## Ejemplo de uso


