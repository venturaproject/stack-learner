package commands

import (
	"boilerplate-learner/internal/config"
	"boilerplate-learner/internal/infrastructure/persistence/models"
	"errors"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func NewMakeUserCommand(db func() (*gorm.DB, error), cfg *config.Config, log *zap.Logger) *cobra.Command {
	var (
		name     string
		email    string
		username string
		password string
		role     string
	)

	cmd := &cobra.Command{
		Use:   "make:user",
		Short: "Crea un nuevo usuario en la base de datos",
		Example: `  console make:user --name "Juan García" --email juan@example.com --password secret123
  console make:user --name Admin --email admin@example.com --password secret123 --role admin`,
		RunE: func(cmd *cobra.Command, args []string) error {
			conn, err := db()
			if err != nil {
				return err
			}
			return runMakeUser(conn, log, name, email, username, password, role)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Nombre completo del usuario (requerido)")
	cmd.Flags().StringVar(&email, "email", "", "Email del usuario (requerido)")
	cmd.Flags().StringVar(&username, "username", "", "Nombre de usuario (opcional)")
	cmd.Flags().StringVar(&password, "password", "", "Contraseña (requerido)")
	cmd.Flags().StringVar(&role, "role", "", "Nombre del rol a asignar (opcional)")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("email")
	_ = cmd.MarkFlagRequired("password")

	return cmd
}

func runMakeUser(db *gorm.DB, log *zap.Logger, name, email, username, password, roleName string) error {
	var existing models.User
	if err := db.Where("email = ?", email).First(&existing).Error; err == nil {
		return errors.New("ya existe un usuario con ese email")
	}

	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return err
	}

	if username == "" {
		username = email
	}

	user := models.User{
		Name:         name,
		Email:        email,
		Username:     &username,
		PasswordHash: hash,
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := db.Create(&user).Error; err != nil {
		return err
	}

	if roleName != "" {
		var r models.Role
		if err := db.Where("name = ?", roleName).First(&r).Error; err != nil {
			log.Warn("rol no encontrado, usuario creado sin rol", zap.String("role", roleName))
		} else {
			_ = db.Model(&user).Association("Roles").Replace([]models.Role{r})
		}
	}

	log.Info("usuario creado",
		zap.String("name", name),
		zap.String("email", email),
		zap.String("role", roleName),
	)
	return nil
}
