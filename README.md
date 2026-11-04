<p align="center">
  <img src="doc/header.png" alt="Learner — plataforma de aprendizaje multi-organización" width="100%" />
</p>

<h1 align="center">Learner</h1>

<p align="center">
  Plataforma de aprendizaje para academias, centros de formación y equipos de empresa.
</p>

<p align="center">
  <a href="doc/product.md">Producto</a> ·
  <a href="doc/architecture.md">Arquitectura</a> ·
  <a href="doc/architecture-reference.md">Guía técnica</a> ·
  <a href="https://github.com/venturaproject/stack-learner">GitHub</a>
</p>

---

## Qué es Learner

Learner es un LMS **multi-organización** para crear cursos, organizar cohortes,
acompañar el progreso del alumnado y evaluar entregas. El backend y la aplicación web
comparten una API; la autorización y el aislamiento por organización se validan en el
servidor.

El proyecto está en desarrollo activo. El alcance implementado y las funciones que aún
están pendientes se detallan en [doc/product.md](doc/product.md).

## Funcionalidades

- Organizaciones y membresías con roles de administrador, docente y estudiante.
- Cursos con módulos y lecciones ordenables, y estados de borrador/publicado.
- Cohortes con docentes asignados, matrículas y registro histórico de bajas.
- Progreso individual con desbloqueo secuencial de lecciones.
- Actividades `task`: entregas de texto y/o enlace, intentos con historial, cola de
  revisión, nota de 0 a 100 y retroalimentación.
- Interfaz web responsive, disponible en español e inglés.

Las invitaciones por correo, el alta de cuentas desde la plataforma y la exportación de
progreso siguen pendientes.

El proyecto parte de una plantilla técnica y todavía conserva algunos módulos
heredados que no pertenecen al alcance LMS; se irán sustituyendo o retirando de forma
gradual.

## Stack

| Área | Tecnologías |
| --- | --- |
| API | Go 1.27 · Gin · GORM |
| Persistencia | PostgreSQL 16 · migraciones SQL |
| Cache y servicios | Redis 7 |
| Web | React · TypeScript · Vite · Tailwind CSS |
| Entorno local | Docker Compose · Nginx |

## Empezar en desarrollo

Requisitos: Docker y Docker Compose.

```bash
make setup
make up
make migrate
make seed
```

Abre [http://localhost:8092/login](http://localhost:8092/login). El usuario local creado
por `make seed` es `admin@example.com` y la contraseña es `admin1234` (**solo para
desarrollo**). Tras iniciar sesión, entra en **Cursos y formación** y crea una
organización para empezar.

| Servicio | Dirección local |
| --- | --- |
| Aplicación (proxy Nginx) | `http://localhost:8092` |
| Frontend de desarrollo | `http://localhost:5174` |
| API | `http://localhost:8011/api` |
| Salud de la API | `http://localhost:8011/api/health` |
| PostgreSQL | `localhost:5445` |
| Redis | `localhost:6380` |

`make up` aplica las migraciones pendientes al iniciar la API; `make migrate` permite
ejecutarlas explícitamente. No uses las credenciales locales en un despliegue real y
cambia `JWT_SECRET` antes de exponer el servicio.

## Comprobaciones

Los chequeos del backend se ejecutan en el contenedor:

```bash
make test
make test-integration
make vet
make staticcheck
make arch-lint
```

Frontend:

```bash
docker compose -f compose.dev.yml exec frontend pnpm build
docker compose -f compose.dev.yml exec frontend pnpm lint
docker compose -f compose.dev.yml exec frontend pnpm test
```

## Arquitectura y documentación

El backend sigue una estructura modular con límites de dominio, casos de uso y
adaptadores HTTP/persistencia. Los detalles y las decisiones vigentes están en:

- [Alcance y permisos del producto](doc/product.md)
- [Arquitectura del backend](doc/architecture.md)
- [Referencia técnica y compromisos actuales](doc/architecture-reference.md)
- [DDD, hexagonal y SOLID](doc/solid-ddd-hexagonal.md)

## Licencia

Todavía no se ha definido una licencia para el repositorio.
