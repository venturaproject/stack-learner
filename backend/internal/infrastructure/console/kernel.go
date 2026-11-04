package console

import (
	"boilerplate-learner/internal/config"
	"boilerplate-learner/internal/infrastructure/console/commands"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Kernel struct {
	root *cobra.Command
}

func NewKernel(cfg *config.Config, log *zap.Logger, db func() (*gorm.DB, error)) *Kernel {
	root := &cobra.Command{
		Use:           "console",
		Short:         "CLI de administración de boilerplate-learner",
		Long:          "Herramientas de administración: migraciones, seeds, gestión de usuarios, etc.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(
		commands.NewMigrateCommand(db, log),
		commands.NewSeedCommand(db, log),
		commands.NewMakeUserCommand(db, cfg, log),
	)

	return &Kernel{root: root}
}

func (k *Kernel) Execute() error {
	return k.root.Execute()
}
