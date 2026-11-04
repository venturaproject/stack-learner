package commands

import (
	"boilerplate-learner/internal/infrastructure/persistence"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func NewMigrateCommand(db func() (*gorm.DB, error), log *zap.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Aplica las migraciones de base de datos pendientes",
		RunE: func(cmd *cobra.Command, args []string) error {
			conn, err := db()
			if err != nil {
				return err
			}
			if err := persistence.Migrate(conn); err != nil {
				return err
			}
			log.Info("migraciones aplicadas")
			return nil
		},
	}
}
