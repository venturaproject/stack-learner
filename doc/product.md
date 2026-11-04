# Learner — alcance de producto

## Visión

Plataforma self-hosted y multi-organización para que academias y equipos creen cursos,
organicen cohortes, acompañen a sus estudiantes y evalúen entregas desde una API común
y una aplicación web responsive.

## Personas y permisos

- **Administrador de plataforma:** gestiona la instalación y las organizaciones.
- **Administrador de organización:** administra miembros, cursos y cohortes de su
  organización.
- **Docente:** puede crear cursos (queda asignado automáticamente), gestiona sus cursos
  asignados, acompaña estudiantes y califica entregas de esos cursos.
- **Estudiante:** consulta su ruta, completa lecciones y entrega actividades.

Una persona puede pertenecer a varias organizaciones y tener distintos roles en cada
una. Las rutas reciben el ámbito de organización solicitado, pero el backend siempre
valida una membresía activa antes de operar; nunca confía en un `organization_id` sin
comprobar esa relación.

## MVP por etapas

### Etapa 1 — organización y catálogo

- Organizaciones y membresías con roles.
- Alta como miembros de cuentas existentes; las invitaciones por correo y el alta de
  cuentas nuevas siguen pendientes.
- Cursos con módulos y lecciones ordenables.
- Estados de borrador/publicado y asignación de docentes.

### Etapa 2 — cohortes y aprendizaje

- Cohortes con fechas, docentes y estudiantes inscritos. El backend y la interfaz de
  gestión de cohortes ya están implementados. El seguimiento personal de finalización de
  lecciones y el desbloqueo secuencial básico también están implementados.
- Lecciones de contenido y recursos; el modelo admite `text`, `video`, `link`, `file`,
  `task` y `quiz`. La evaluación implementada corresponde a `task`; `quiz` todavía no
  tiene un flujo de evaluación propio.
- Paneles iniciales de administración, docente y estudiante.

### Etapa 3 — evaluación

- Actividades `task` y entregas de texto o enlace, con intentos múltiples e historial.
- Cola de revisión para administradores y docentes asignados al curso; nota de 0 a 100 y
  retroalimentación. Cada intento solo puede revisarse una vez.
- Exportación sencilla de progreso (pendiente).

### Después del MVP

Certificados PDF, carga masiva CSV, notificaciones por email, analítica avanzada,
gamificación, app móvil, videoconferencia e integraciones externas.

## Principios de dominio y seguridad

- Cada dato educativo pertenece a una organización; autorización y aislamiento se
  validan en el backend en cada caso de uso y consulta.
- La membresía une usuario, organización, rol y estado; el rol no es global.
- Actualmente los cursos publicados son inmutables desde los endpoints de edición,
  incluso si todavía no tienen estudiantes. Si se incorpora edición/versionado, deberá
  preservar el progreso o crear una revisión cuando existan matrículas activas.
- El progreso registra finalización por estudiante y unidad de contenido, no se infiere
  solamente desde el navegador.
- Las entregas conservan historial de envíos y de evaluación.
- Los límites entre organizaciones tendrán pruebas negativas explícitas.

## Contextos candidatos

`organizations` (organizaciones y membresías), `accounts` (identidad y acceso),
`curriculum` (cursos, módulos y lecciones), `cohorts` (cohortes e inscripciones),
`learning` (progreso y desbloqueos) y `assessment` (actividades, entregas y revisión).
En la implementación actual, las reglas de cohorte viven en `domain/cohorts`, mientras
los casos de uso de cohortes y progreso se coordinan desde `application/learning`; las
reglas y casos de uso de entregas se separan entre `domain/assessment` y
`application/assessment`.

Son límites iniciales sujetos a evolución al implementar los casos de uso; no implican
crear microservicios. La arquitectura de referencia es un monolito modular con puertos
y adaptadores.

## Criterios para empezar cada etapa

Cada módulo deberá incluir reglas de autorización por organización, validaciones de
invariantes, migración/modelo, endpoints documentados y pruebas de aplicación. UI y
API avanzan con datos reales; no se dará por terminado un módulo si es solo una pantalla
con datos simulados.

## Primera implementación disponible

El backend expone los siguientes casos de uso bajo `/api`:

- `GET/POST /organizations`: listar las organizaciones del usuario y crear una
  organización (quien la crea recibe `org_admin`).
- `PUT/DELETE /organizations/:orgID`: actualizar o archivar una organización; solo
  `org_admin`. El archivado conserva cursos, contenidos y membresías.
- `POST /organizations/:orgID/restore`: restaurar una organización archivada; solo
  `org_admin`. Las operaciones habituales sobre organizaciones archivadas se rechazan.
- `POST /organizations/:orgID/members`: añadir un usuario existente por email con rol
  `org_admin`, `trainer` o `student`; solo un administrador de esa organización puede
  hacerlo.
- `GET/POST /organizations/:orgID/students` y
  `DELETE /organizations/:orgID/students/:studentID`: listar, añadir como miembro a una
  cuenta existente y desactivar la membresía de un estudiante en la organización;
  `org_admin` o `trainer` pueden consultar la lista y solo `org_admin` puede añadir o
  desactivar miembros. Esta baja de membresía no equivale a retirar al estudiante de una
  cohorte; las bajas de cohorte tienen su propio endpoint e historial.
- `GET/POST /organizations/:orgID/courses`: consultar el catálogo de la organización
  y crear cursos en borrador (administrador o docente).
- `PUT/DELETE /organizations/:orgID/courses/:courseID`: editar o eliminar cursos en
  borrador (administrador o docente); los cursos publicados se conservan sin cambios.
- `GET /organizations/:orgID/courses/:courseID`: consultar estructura; los estudiantes
  solo pueden leer cursos publicados vinculados a una cohorte propia activa o finalizada.
- `POST .../:courseID/modules` y `POST .../:moduleID/lessons`: construir la ruta del
  curso.
- `PUT/DELETE .../:courseID/modules/:moduleID` y
  `PUT/DELETE .../:moduleID/lessons/:lessonID`: editar o quitar elementos mientras el
  curso está en borrador. Eliminar un módulo también elimina sus lecciones.
- `PUT .../:courseID/modules/order` y `PUT .../:moduleID/lessons/order`: reordenar
  enviando la lista completa de IDs en el campo `ids`.
- `GET/POST/DELETE .../:courseID/instructors`: consultar docentes activos de la
  organización y asignarlos/quitarles del curso; solo `org_admin` puede modificar
  asignaciones. Los docentes solo pueden consultar y modificar los cursos que tienen
  asignados.
- `POST .../:courseID/publish`: publicar cuando el curso ya contiene al menos una
  lección.
- `GET /organizations/:orgID/cohorts` y
  `GET /organizations/:orgID/courses/:courseID/cohorts`: listar cohortes visibles para
  el rol; los administradores ven todas, los docentes las asignadas y los estudiantes
  las cohortes activas o finalizadas donde conservan una matrícula activa.
- `POST /organizations/:orgID/courses/:courseID/cohorts`: crear cohorte en borrador
  asociada a un curso publicado; solo `org_admin`.
- `PUT /organizations/:orgID/cohorts/:cohortID` y
  `PUT /organizations/:orgID/cohorts/:cohortID/status`: editar borradores y transicionar
  estados; solo `org_admin`. Los estados terminales no se reactivan.
- `GET/POST/DELETE /organizations/:orgID/cohorts/:cohortID/instructors`: consultar,
  asignar y quitar docentes; solo `org_admin` puede cambiar asignaciones.
- `GET/POST/DELETE /organizations/:orgID/cohorts/:cohortID/enrollments`: consultar el
  historial, inscribir y dar de baja estudiantes. Administradores y docentes asignados
  a esa cohorte gestionan matrículas; solo se inscribe en cohortes activas. Las bajas se
  conservan y permiten una nueva matrícula histórica.
- `GET /organizations/:orgID/courses/:courseID/progress` y
  `POST /organizations/:orgID/courses/:courseID/lessons/:lessonID/complete`: consultar y
  actualizar el progreso propio. Solo estudiantes con matrícula activa en una cohorte
  activa o finalizada pueden usar estas operaciones; completar una lección es idempotente
  y no permite modificar el progreso de otra persona. Las lecciones se desbloquean en el
  orden del curso y el contenido de las lecciones bloqueadas no se incluye en la respuesta.
- `POST/GET /organizations/:orgID/courses/:courseID/lessons/:lessonID/submissions`: enviar
  intentos de una lección `task` y consultar el historial propio. Solo estudiantes con
  acceso desbloqueado a la actividad; se admite texto, enlace o ambos. Se requiere al
  menos un intento antes de poder completar la lección `task`.
- `GET /organizations/:orgID/courses/:courseID/submissions` y
  `POST /organizations/:orgID/courses/:courseID/submissions/:submissionID/review`: cola de
  revisión y calificación de un intento con nota de 0 a 100 y retroalimentación. Solo
  administradores de la organización y docentes asignados al curso; una revisión guardada
  no se sobrescribe y los intentos posteriores mantienen su propio historial.

Los listados paginados de organizaciones, estudiantes, cursos, cohortes, matrículas y la
cola de entregas aceptan `page` y `per_page` (máximo 100), y responden con la colección
correspondiente y metadatos `pagination`. Los listados de organizaciones, estudiantes,
cursos, cohortes y entregas también aceptan `search`; el historial personal de intentos
se devuelve completo.

Toda ruta con organización valida en servidor la membresía activa y aplica el scope de
organización a la consulta. La interfaz incluye páginas de listado, alta y edición de
organizaciones, cursos, cohortes y estudiantes, así como la gestión de matrícula y
catálogo/estructura de cursos. Los estudiantes solo pueden listar y consultar cursos
 publicados vinculados a una cohorte activa o finalizada donde mantengan una matrícula
 activa. El progreso personal registra finalizaciones de lecciones y su resumen, y libera
 el contenido siguiente cuando se completa la lección previa. Las entregas y la revisión
 se realizan desde lecciones `task`, con texto/enlace, intentos históricos y nota con
 retroalimentación; la exportación de progreso sigue pendiente. La matrícula de
 estudiantes apunta a cuentas ya existentes; no crea cuentas de usuario.
