package commands

import (
	"fmt"
	"time"

	"boilerplate-learner/internal/infrastructure/persistence/models"

	"github.com/alexedwards/argon2id"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func sp(s string) *string { return &s }
func ip(i int32) *int32   { return &i }

func NewSeedCommand(db func() (*gorm.DB, error), log *zap.Logger) *cobra.Command {
	var fresh bool

	cmd := &cobra.Command{
		Use:   "seed",
		Short: "Carga datos de referencia y demostración",
		RunE: func(cmd *cobra.Command, args []string) error {
			conn, err := db()
			if err != nil {
				return err
			}
			return RunSeed(conn, log, fresh)
		},
	}

	cmd.Flags().BoolVar(&fresh, "fresh", false, "Trunca todas las tablas antes de sembrar")

	return cmd
}

func RunSeed(db *gorm.DB, log *zap.Logger, fresh bool) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if fresh {
			log.Warn("truncando tablas...")
			tables := []string{
				"dispositivos", "telefonos", "trabajadores",
				"tipologias", "estados_telefono",
				"device_models", "device_brands",
				"refresh_tokens", "user_settings",
				"role_permissions", "user_roles",
				"permissions", "roles", "users",
			}
			for _, t := range tables {
				if err := tx.Exec("TRUNCATE TABLE " + t + " RESTART IDENTITY CASCADE").Error; err != nil {
					return fmt.Errorf("truncate %s: %w", t, err)
				}
			}
		}
		if err := seedAuth(tx, log); err != nil {
			return err
		}
		if err := seedDomain(tx, log); err != nil {
			return err
		}

		log.Info("seed completado")
		return nil
	})
}

func seedAuth(db *gorm.DB, log *zap.Logger) error {
	permissions := []models.Permission{
		{Name: "users.view"}, {Name: "users.create"}, {Name: "users.edit"}, {Name: "users.delete"},
		{Name: "roles.view"}, {Name: "roles.manage"}, {Name: "roles.create"}, {Name: "roles.edit"}, {Name: "roles.delete"},
		{Name: "permissions.view"}, {Name: "permissions.create"}, {Name: "permissions.delete"},
		{Name: "trabajadores.view"}, {Name: "trabajadores.create"}, {Name: "trabajadores.edit"}, {Name: "trabajadores.delete"},
		{Name: "telefonos.view"}, {Name: "telefonos.create"}, {Name: "telefonos.edit"}, {Name: "telefonos.delete"},
		{Name: "dispositivos.view"}, {Name: "dispositivos.create"}, {Name: "dispositivos.edit"}, {Name: "dispositivos.delete"},
		{Name: "api_clients.manage"},
	}
	for i := range permissions {
		if err := db.Where(models.Permission{Name: permissions[i].Name}).FirstOrCreate(&permissions[i]).Error; err != nil {
			return fmt.Errorf("seed permission %s: %w", permissions[i].Name, err)
		}
	}

	adminRole := models.Role{Name: "admin"}
	if err := db.Where(models.Role{Name: adminRole.Name}).FirstOrCreate(&adminRole).Error; err != nil {
		return fmt.Errorf("seed admin role: %w", err)
	}
	if err := db.Model(&adminRole).Association("Permissions").Replace(permissions); err != nil {
		return fmt.Errorf("assign admin permissions: %w", err)
	}

	hash, err := argon2id.CreateHash("admin1234", argon2id.DefaultParams)
	if err != nil {
		return err
	}

	username := "admin"
	adminUser := models.User{
		Name:         "Administrator",
		Email:        "admin@example.com",
		Username:     &username,
		PasswordHash: hash,
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := db.Where(models.User{Email: adminUser.Email}).FirstOrCreate(&adminUser).Error; err != nil {
		return err
	}
	if err := db.Model(&adminUser).Association("Roles").Replace([]models.Role{adminRole}); err != nil {
		return fmt.Errorf("assign admin role: %w", err)
	}

	log.Info("auth seed completo", zap.String("email", "admin@example.com"), zap.String("password", "admin1234"))
	return nil
}

func seedDomain(db *gorm.DB, log *zap.Logger) error {
	sqls := []struct {
		label string
		sql   string
	}{
		{"tipologias", `
			INSERT INTO tipologias (id, nombre) VALUES
				(1,'Corporativo'),(2,'Operaciones'),(3,'Dirección'),(4,'Temporal')
			ON CONFLICT DO NOTHING;
			SELECT setval('tipologias_id_seq',(SELECT MAX(id) FROM tipologias));
		`},
		{"device_brands", `
			INSERT INTO device_brands (id, nombre) VALUES
				(1,'Apple'),(2,'Samsung'),(3,'Lenovo'),(4,'HP'),(5,'Dell')
			ON CONFLICT DO NOTHING;
			SELECT setval('device_brands_id_seq',(SELECT MAX(id) FROM device_brands));
		`},
	}

	for _, s := range sqls {
		if err := db.Exec(s.sql).Error; err != nil {
			return err
		}
		log.Info("sembrado", zap.String("tabla", s.label))
	}

	return seedDemoData(db, log)
}

// seedDemoData creates a realistic set of trabajadores / teléfonos / dispositivos
// so the dashboard and list views have something to show. It is skipped when the
// data already exists (run `seed --fresh` to rebuild).
func seedDemoData(db *gorm.DB, log *zap.Logger) error {
	var count int64
	if err := db.Model(&models.Trabajador{}).Count(&count).Error; err != nil {
		return fmt.Errorf("check demo data: %w", err)
	}
	if count > 0 {
		log.Info("demo data ya presente, se omite (usa --fresh para regenerar)")
		return nil
	}

	// Device models per brand.
	deviceModels := []models.DeviceModel{
		{Nombre: "iPhone 13", BrandID: 1}, {Nombre: "iPhone 14", BrandID: 1}, {Nombre: "MacBook Air M2", BrandID: 1},
		{Nombre: "Galaxy S22", BrandID: 2}, {Nombre: "Galaxy A53", BrandID: 2},
		{Nombre: "ThinkPad X1 Carbon", BrandID: 3}, {Nombre: "IdeaPad 5", BrandID: 3},
		{Nombre: "ProBook 450 G9", BrandID: 4}, {Nombre: "EliteBook 840", BrandID: 4},
		{Nombre: "Latitude 5540", BrandID: 5},
	}
	if err := db.Create(&deviceModels).Error; err != nil {
		return fmt.Errorf("device_models: %w", err)
	}

	// Trabajadores — incorporación spread over the last 6 months.
	ciudades := []string{"Madrid", "Barcelona", "Valencia", "Sevilla", "Bilbao"}
	deptos := []string{"Operaciones", "IT", "Ventas", "Dirección", "RRHH"}
	cargos := []string{"Técnico", "Responsable", "Comercial", "Director", "Administrativo"}
	now := time.Now()
	var trabajadores []models.Trabajador
	for i := 0; i < 16; i++ {
		fecha := now.AddDate(0, -(i % 6), -(i % 20))
		estado := "activo"
		var fechaBaja *time.Time
		if i%7 == 0 && i != 0 {
			estado = "baja"
			fb := now.AddDate(0, 0, -10)
			fechaBaja = &fb
		}
		trabajadores = append(trabajadores, models.Trabajador{
			Nombre:             fmt.Sprintf("Trabajador%02d", i+1),
			Apellido:           fmt.Sprintf("Apellido%02d", i+1),
			DNI:                sp(fmt.Sprintf("%08dX", 10000000+i)),
			Email:              sp(fmt.Sprintf("trabajador%02d@example.com", i+1)),
			Estado:             estado,
			Cargo:              sp(cargos[i%len(cargos)]),
			Departamento:       sp(deptos[i%len(deptos)]),
			Ciudad:             sp(ciudades[i%len(ciudades)]),
			Ubicacion:          sp(ciudades[i%len(ciudades)] + " - Oficina"),
			FechaIncorporacion: &fecha,
			FechaBaja:          fechaBaja,
		})
	}
	if err := db.Create(&trabajadores).Error; err != nil {
		return fmt.Errorf("trabajadores: %w", err)
	}

	// Teléfonos — varied estado, ~half assigned to a worker.
	operadoras := []string{"Movistar", "Vodafone", "Orange", "Yoigo"}
	planes := []string{"Empresa 20GB", "Empresa 50GB", "Ilimitado", "Básico 5GB"}
	estadoIDs := []int32{1, 1, 1, 2, 2, 3, 5} // Activo x3, Congelado x2, Baja, Sin datos
	var telefonos []models.Telefono
	for i := 0; i < 22; i++ {
		estadoID := estadoIDs[i%len(estadoIDs)]
		var trabID *int32
		activo := estadoID == 1
		if activo && i < len(trabajadores) {
			trabID = ip(trabajadores[i].ID)
		}
		telefonos = append(telefonos, models.Telefono{
			Numero:       sp(fmt.Sprintf("6%08d", 10000000+i)),
			Tipo:         sp("Smartphone"),
			Plan:         sp(planes[i%len(planes)]),
			IMEI:         sp(fmt.Sprintf("35%013d", 1000000000000+int64(i))),
			Marca:        sp([]string{"Apple", "Samsung"}[i%2]),
			Modelo:       sp([]string{"iPhone 13", "Galaxy S22"}[i%2]),
			Operadora:    sp(operadoras[i%len(operadoras)]),
			EstadoID:     ip(estadoID),
			TipologiaID:  ip(int32(i%4 + 1)),
			TrabajadorID: trabID,
			Activo:       activo,
		})
	}
	if err := db.Create(&telefonos).Error; err != nil {
		return fmt.Errorf("telefonos: %w", err)
	}

	// Dispositivos — laptops/tablets, ~half assigned.
	grupos := []string{"Portátiles", "Sobremesa", "Tablets", "Periféricos"}
	dispEstados := []string{"disponible", "disponible", "asignado"}
	var dispositivos []models.Dispositivo
	for i := 0; i < 15; i++ {
		brandID := int32(i%5 + 1)
		modelID := deviceModels[i%len(deviceModels)].ID
		estado := dispEstados[i%len(dispEstados)]
		var trabID *int32
		if estado == "asignado" && i < len(trabajadores) {
			trabID = ip(trabajadores[i].ID)
		}
		dispositivos = append(dispositivos, models.Dispositivo{
			Numero:       sp(fmt.Sprintf("DISP-%03d", i+1)),
			Serie:        sp(fmt.Sprintf("SN%08d", 90000000+i)),
			IMEI:         sp(fmt.Sprintf("49%013d", 2000000000000+int64(i))),
			BrandID:      ip(brandID),
			ModelID:      ip(modelID),
			Grupo:        sp(grupos[i%len(grupos)]),
			Estado:       estado,
			TrabajadorID: trabID,
		})
	}
	if err := db.Create(&dispositivos).Error; err != nil {
		return fmt.Errorf("dispositivos: %w", err)
	}

	log.Info("demo data sembrada",
		zap.Int("trabajadores", len(trabajadores)),
		zap.Int("telefonos", len(telefonos)),
		zap.Int("dispositivos", len(dispositivos)))
	return nil
}
