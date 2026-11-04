package models

import "time"

// ── Auth ─────────────────────────────────────────────────────────────────────

type User struct {
	ID           int32         `gorm:"primaryKey;autoIncrement;column:id"`
	Name         string        `gorm:"column:name;not null"`
	Email        string        `gorm:"column:email;uniqueIndex;not null"`
	Username     *string       `gorm:"column:username;uniqueIndex"`
	PasswordHash string        `gorm:"column:password_hash;not null"`
	Status       string        `gorm:"column:status;default:active"`
	CreatedAt    time.Time     `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time     `gorm:"column:updated_at;autoUpdateTime"`
	Settings     *UserSettings `gorm:"foreignKey:UserID"`
	Roles        []Role        `gorm:"many2many:user_roles;joinForeignKey:UserID;joinReferences:RoleID"`
	Permissions  []Permission  `gorm:"many2many:user_permissions;joinForeignKey:UserID;joinReferences:PermissionID"`
}

func (User) TableName() string { return "users" }

type UserSettings struct {
	ID     int32   `gorm:"primaryKey;autoIncrement;column:id"`
	UserID int32   `gorm:"column:user_id;uniqueIndex;not null"`
	Avatar *string `gorm:"column:avatar"`
}

func (UserSettings) TableName() string { return "user_settings" }

type Role struct {
	ID          int32        `gorm:"primaryKey;autoIncrement;column:id"`
	Name        string       `gorm:"column:name;uniqueIndex;not null"`
	GuardName   string       `gorm:"column:guard_name;default:api"`
	CreatedAt   time.Time    `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time    `gorm:"column:updated_at;autoUpdateTime"`
	Permissions []Permission `gorm:"many2many:role_permissions;joinForeignKey:RoleID;joinReferences:PermissionID"`
}

func (Role) TableName() string { return "roles" }

type Permission struct {
	ID        int32     `gorm:"primaryKey;autoIncrement;column:id"`
	Name      string    `gorm:"column:name;uniqueIndex;not null"`
	GuardName string    `gorm:"column:guard_name;default:api"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Permission) TableName() string { return "permissions" }

type RefreshToken struct {
	ID        int32     `gorm:"primaryKey;autoIncrement;column:id"`
	JTI       string    `gorm:"column:jti;uniqueIndex;not null"`
	UserID    int32     `gorm:"column:user_id;not null"`
	Revoked   bool      `gorm:"column:revoked;default:false"`
	ExpiresAt time.Time `gorm:"column:expires_at;not null"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }

type RevokedAccessToken struct {
	JTI string    `gorm:"primaryKey;column:jti"`
	Exp time.Time `gorm:"column:exp;not null"`
}

func (RevokedAccessToken) TableName() string { return "revoked_access_tokens" }

// ── Domain ───────────────────────────────────────────────────────────────────

type Trabajador struct {
	ID                 int32      `gorm:"primaryKey;autoIncrement;column:id"`
	Nombre             string     `gorm:"column:nombre;not null"`
	Apellido           string     `gorm:"column:apellido;not null;default:''"`
	SynergyResID       *int32     `gorm:"column:synergy_res_id;uniqueIndex"`
	DNI                *string    `gorm:"column:dni;uniqueIndex"`
	Email              *string    `gorm:"column:email"`
	Telefono           *string    `gorm:"column:telefono"`
	Estado             string     `gorm:"column:estado;default:activo"`
	EmpStat            *string    `gorm:"column:emp_stat"`
	Cargo              *string    `gorm:"column:cargo"`
	Departamento       *string    `gorm:"column:departamento"`
	Loc                *string    `gorm:"column:loc"`
	Ubicacion          *string    `gorm:"column:ubicacion"`
	Ciudad             *string    `gorm:"column:ciudad"`
	FechaIncorporacion *time.Time `gorm:"column:fecha_incorporacion"`
	FechaBaja          *time.Time `gorm:"column:fecha_baja"`
	Observaciones      string     `gorm:"column:observaciones;default:''"`
	CreatedAt          time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (Trabajador) TableName() string { return "trabajadores" }

type EstadoTelefono struct {
	ID        int32     `gorm:"primaryKey;autoIncrement;column:id"`
	Nombre    string    `gorm:"column:nombre;uniqueIndex;not null"`
	Color     *string   `gorm:"column:color"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (EstadoTelefono) TableName() string { return "estados_telefono" }

type Tipologia struct {
	ID     int32  `gorm:"primaryKey;autoIncrement;column:id"`
	Nombre string `gorm:"column:nombre;uniqueIndex;not null"`
}

func (Tipologia) TableName() string { return "tipologias" }

type Telefono struct {
	ID            int32               `gorm:"primaryKey;autoIncrement;column:id"`
	Numero        *string             `gorm:"column:numero"`
	Tipo          *string             `gorm:"column:tipo"`
	Plan          *string             `gorm:"column:plan"`
	Nplan         *string             `gorm:"column:nplan"`
	Pin           *string             `gorm:"column:pin"`
	Puk           *string             `gorm:"column:puk"`
	IMEI          *string             `gorm:"column:imei"`
	IMEI2         *string             `gorm:"column:imei2"`
	Marca         *string             `gorm:"column:marca"`
	Modelo        *string             `gorm:"column:modelo"`
	Linea         *string             `gorm:"column:linea"`
	Operadora     *string             `gorm:"column:operadora"`
	EstadoID      *int32              `gorm:"column:estado_id"`
	TipologiaID   *int32              `gorm:"column:tipologia_id"`
	TrabajadorID  *int32              `gorm:"column:trabajador_id"`
	Observaciones string              `gorm:"column:observaciones;default:''"`
	Activo        bool                `gorm:"column:activo;default:true"`
	CreatedAt     time.Time           `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time           `gorm:"column:updated_at;autoUpdateTime"`
	Estado        *EstadoTelefono     `gorm:"foreignKey:EstadoID"`
	Trabajador    *Trabajador         `gorm:"foreignKey:TrabajadorID"`
	Actuaciones   []ActuacionTelefono `gorm:"foreignKey:TelefonoID"`
}

func (Telefono) TableName() string { return "telefonos" }

type ActuacionTelefono struct {
	ID          int32     `gorm:"primaryKey;autoIncrement;column:id"`
	TelefonoID  int32     `gorm:"column:telefono_id;not null"`
	Descripcion string    `gorm:"column:descripcion;not null"`
	UserID      *int32    `gorm:"column:user_id"`
	Fecha       time.Time `gorm:"column:fecha;autoCreateTime"`
}

func (ActuacionTelefono) TableName() string { return "actuaciones_telefono" }

type DeviceBrand struct {
	ID     int32  `gorm:"primaryKey;autoIncrement;column:id"`
	Nombre string `gorm:"column:nombre;uniqueIndex;not null"`
}

func (DeviceBrand) TableName() string { return "device_brands" }

type DeviceModel struct {
	ID      int32        `gorm:"primaryKey;autoIncrement;column:id"`
	Nombre  string       `gorm:"column:nombre;not null"`
	BrandID int32        `gorm:"column:brand_id;not null"`
	Brand   *DeviceBrand `gorm:"foreignKey:BrandID"`
}

func (DeviceModel) TableName() string { return "device_models" }

type Dispositivo struct {
	ID            int32        `gorm:"primaryKey;autoIncrement;column:id"`
	Numero        *string      `gorm:"column:numero"`
	IMEI          *string      `gorm:"column:imei;uniqueIndex"`
	Serie         *string      `gorm:"column:serie"`
	SynergyResID  *int32       `gorm:"column:synergy_res_id;uniqueIndex"`
	BrandID       *int32       `gorm:"column:brand_id"`
	ModelID       *int32       `gorm:"column:model_id"`
	Grupo         *string      `gorm:"column:grupo"`
	Estado        string       `gorm:"column:estado;default:disponible"`
	TrabajadorID  *int32       `gorm:"column:trabajador_id"`
	Observaciones string       `gorm:"column:observaciones;default:''"`
	CreatedAt     time.Time    `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time    `gorm:"column:updated_at;autoUpdateTime"`
	Brand         *DeviceBrand `gorm:"foreignKey:BrandID"`
	Model         *DeviceModel `gorm:"foreignKey:ModelID"`
	Trabajador    *Trabajador  `gorm:"foreignKey:TrabajadorID"`
}

func (Dispositivo) TableName() string { return "dispositivos" }

// ── Learning management ─────────────────────────────────────────────────────

type Organization struct {
	ID         int32      `gorm:"primaryKey;autoIncrement"`
	Name       string     `gorm:"not null;size:160"`
	Slug       string     `gorm:"not null;size:100;uniqueIndex"`
	CreatedBy  int32      `gorm:"not null;index"`
	ArchivedAt *time.Time `gorm:"index"`
	CreatedAt  time.Time  `gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime"`
}

func (Organization) TableName() string { return "organizations" }

type OrganizationMembership struct {
	ID             int32     `gorm:"primaryKey;autoIncrement"`
	OrganizationID int32     `gorm:"not null;uniqueIndex:idx_org_member"`
	UserID         int32     `gorm:"not null;uniqueIndex:idx_org_member;index"`
	Role           string    `gorm:"not null;size:32"`
	Status         string    `gorm:"not null;default:active;size:24"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
}

func (OrganizationMembership) TableName() string { return "organization_memberships" }

type Course struct {
	ID             int32          `gorm:"primaryKey;autoIncrement"`
	OrganizationID int32          `gorm:"not null;uniqueIndex:idx_org_course_slug;index"`
	Title          string         `gorm:"not null;size:160"`
	Slug           string         `gorm:"not null;size:180;uniqueIndex:idx_org_course_slug"`
	Description    string         `gorm:"type:text;not null;default:''"`
	Status         string         `gorm:"not null;default:draft;size:24"`
	CreatedBy      int32          `gorm:"not null"`
	CreatedAt      time.Time      `gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime"`
	Modules        []CourseModule `gorm:"foreignKey:CourseID"`
}

func (Course) TableName() string { return "courses" }

type CourseModule struct {
	ID        int32     `gorm:"primaryKey;autoIncrement"`
	CourseID  int32     `gorm:"not null;uniqueIndex:idx_course_module_position"`
	Title     string    `gorm:"not null;size:160"`
	Position  int       `gorm:"not null;uniqueIndex:idx_course_module_position"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	Lessons   []Lesson  `gorm:"foreignKey:ModuleID"`
}

func (CourseModule) TableName() string { return "course_modules" }

type Lesson struct {
	ID        int32     `gorm:"primaryKey;autoIncrement"`
	ModuleID  int32     `gorm:"not null;uniqueIndex:idx_module_lesson_position"`
	Title     string    `gorm:"not null;size:160"`
	Kind      string    `gorm:"not null;size:24"`
	Content   string    `gorm:"type:text;not null;default:''"`
	Position  int       `gorm:"not null;uniqueIndex:idx_module_lesson_position"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (Lesson) TableName() string { return "lessons" }

type CourseTrainer struct {
	CourseID  int32     `gorm:"primaryKey;column:course_id"`
	TrainerID int32     `gorm:"primaryKey;column:trainer_id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (CourseTrainer) TableName() string { return "course_trainers" }

type Cohort struct {
	ID             int32      `gorm:"primaryKey;autoIncrement"`
	OrganizationID int32      `gorm:"not null;index:idx_cohort_org_status"`
	CourseID       int32      `gorm:"not null;index"`
	Name           string     `gorm:"not null;size:160"`
	StartsAt       *time.Time `gorm:"column:starts_at"`
	EndsAt         *time.Time `gorm:"column:ends_at"`
	Status         string     `gorm:"not null;size:24;default:draft;index:idx_cohort_org_status"`
	CreatedBy      int32      `gorm:"not null"`
	CreatedAt      time.Time  `gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime"`
}

func (Cohort) TableName() string { return "cohorts" }

type CohortInstructor struct {
	CohortID  int32     `gorm:"primaryKey;column:cohort_id"`
	TrainerID int32     `gorm:"primaryKey;column:trainer_id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (CohortInstructor) TableName() string { return "cohort_instructors" }

type CohortEnrollment struct {
	ID               int64      `gorm:"primaryKey;autoIncrement"`
	CohortID         int32      `gorm:"not null;index:idx_cohort_enrollment_status"`
	UserID           int32      `gorm:"not null;index"`
	Status           string     `gorm:"not null;size:24;default:active;index:idx_cohort_enrollment_status"`
	EnrolledBy       int32      `gorm:"not null"`
	EnrolledAt       time.Time  `gorm:"not null;autoCreateTime"`
	WithdrawnBy      *int32     `gorm:"column:withdrawn_by"`
	WithdrawnAt      *time.Time `gorm:"column:withdrawn_at"`
	WithdrawalReason *string    `gorm:"column:withdrawal_reason;size:500"`
}

func (CohortEnrollment) TableName() string { return "cohort_enrollments" }

type StudentLessonProgress struct {
	UserID      int32     `gorm:"primaryKey;column:user_id"`
	LessonID    int32     `gorm:"primaryKey;column:lesson_id"`
	CompletedAt time.Time `gorm:"not null;column:completed_at;autoCreateTime"`
}

func (StudentLessonProgress) TableName() string { return "student_lesson_progress" }

type AssessmentSubmission struct {
	ID             int64      `gorm:"primaryKey;autoIncrement"`
	OrganizationID int32      `gorm:"not null;index:idx_assessment_submissions_course_status"`
	CourseID       int32      `gorm:"not null;index:idx_assessment_submissions_course_status"`
	LessonID       int32      `gorm:"not null;index:idx_assessment_submissions_lesson"`
	StudentID      int32      `gorm:"not null;index:idx_assessment_submissions_student"`
	Attempt        int        `gorm:"column:attempt_number;not null"`
	ResponseText   string     `gorm:"not null;type:text;default:''"`
	ResponseURL    *string    `gorm:"column:response_url;size:2048"`
	Status         string     `gorm:"not null;size:24;default:submitted;index:idx_assessment_submissions_course_status"`
	Score          *float64   `gorm:"type:numeric(5,2)"`
	Feedback       *string    `gorm:"type:text"`
	SubmittedAt    time.Time  `gorm:"not null;column:submitted_at;autoCreateTime"`
	ReviewedBy     *int32     `gorm:"column:reviewed_by"`
	ReviewedAt     *time.Time `gorm:"column:reviewed_at"`
}

func (AssessmentSubmission) TableName() string { return "assessment_submissions" }

// ── API Clients ───────────────────────────────────────────────────────────────

// ApiClient stores external API client credentials (secret stored as SHA-256 hash, never raw).
type ApiClient struct {
	ID         string     `gorm:"primaryKey;type:varchar(36);column:id"`
	Name       string     `gorm:"column:name;not null"`
	ClientID   string     `gorm:"column:client_id;uniqueIndex;not null"`
	SecretHash string     `gorm:"column:secret_hash;not null"`
	Scopes     string     `gorm:"column:scopes;not null;default:'[]'"` // JSON array
	Active     bool       `gorm:"column:active;default:true"`
	LastUsedAt *time.Time `gorm:"column:last_used_at"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (ApiClient) TableName() string { return "api_clients" }

// ApiClientToken stores access/refresh token hashes (never raw tokens).
type ApiClientToken struct {
	ID               string    `gorm:"primaryKey;type:varchar(36);column:id"`
	ClientID         string    `gorm:"column:client_id;not null;index"`
	Scopes           string    `gorm:"column:scopes;not null;default:'[]'"` // JSON array
	AccessTokenHash  string    `gorm:"column:access_token_hash;uniqueIndex;not null"`
	RefreshTokenHash string    `gorm:"column:refresh_token_hash;uniqueIndex;not null"`
	AccessExpiresAt  time.Time `gorm:"column:access_expires_at;not null"`
	RefreshExpiresAt time.Time `gorm:"column:refresh_expires_at;not null"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (ApiClientToken) TableName() string { return "api_client_tokens" }

// ── Outbox ───────────────────────────────────────────────────────────────────

// OutboxMessage is a domain event persisted in the same transaction as the
// change that produced it. A background dispatcher delivers unprocessed rows to
// the in-process event bus and stamps ProcessedAt. Delivery is retried with
// backoff (NextAttemptAt) until it succeeds or the attempt budget is exhausted,
// at which point the row is dead-lettered (FailedAt set, LastError kept).
type OutboxMessage struct {
	ID            int64      `gorm:"primaryKey;autoIncrement;column:id"`
	EventName     string     `gorm:"column:event_name;not null;index"`
	Payload       []byte     `gorm:"column:payload;type:jsonb;not null"`
	OccurredAt    time.Time  `gorm:"column:occurred_at;not null"`
	ProcessedAt   *time.Time `gorm:"column:processed_at;index"`
	FailedAt      *time.Time `gorm:"column:failed_at;index"`
	Attempts      int        `gorm:"column:attempts;not null;default:0"`
	LastError     *string    `gorm:"column:last_error"`
	NextAttemptAt time.Time  `gorm:"column:next_attempt_at;not null;default:now();index"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (OutboxMessage) TableName() string { return "outbox_messages" }
