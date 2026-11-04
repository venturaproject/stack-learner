package telefonos

import (
	"context"

	appshared "boilerplate-learner/internal/application/shared"
	domaintelf "boilerplate-learner/internal/domain/telefonos"
	"boilerplate-learner/internal/shared/apperrors"
)

type AddActuacionUseCase struct {
	repo    domaintelf.Repository
	actRepo domaintelf.ActuacionRepository
	tx      appshared.TxManager
}

func (uc *AddActuacionUseCase) Execute(ctx context.Context, telefonoID int32, in ActuacionInput, userID *int32) (*domaintelf.Actuacion, error) {
	if in.Descripcion == "" {
		return nil, apperrors.Invalid("la descripción es obligatoria")
	}
	var a *domaintelf.Actuacion
	err := uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		if _, err := uc.repo.FindByID(ctx, telefonoID); err != nil {
			return err
		}
		created, err := uc.actRepo.Add(ctx, telefonoID, in.Descripcion, userID)
		if err != nil {
			return err
		}
		a = created
		return nil
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

type RemoveActuacionUseCase struct {
	actRepo domaintelf.ActuacionRepository
}

func (uc *RemoveActuacionUseCase) Execute(ctx context.Context, id int32) error {
	return uc.actRepo.Remove(ctx, id)
}
