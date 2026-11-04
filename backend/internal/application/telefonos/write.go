package telefonos

import (
	"context"

	appshared "boilerplate-learner/internal/application/shared"
	"boilerplate-learner/internal/application/telefonos/read"
	domaintelf "boilerplate-learner/internal/domain/telefonos"
	"boilerplate-learner/internal/shared/apperrors"
)

type CreateUseCase struct {
	repo    domaintelf.Repository
	queries read.Queries
	tx      appshared.TxManager
	pub     appshared.EventPublisher
}

func (uc *CreateUseCase) Execute(ctx context.Context, in CreateInput) (*read.Details, error) {
	var id int32
	err := uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		t, err := domaintelf.NewTelefono(in.toDomain())
		if err != nil {
			return err
		}
		if err := uc.repo.Save(ctx, t); err != nil {
			return err
		}
		id = t.ID
		return uc.pub.Publish(ctx, t.PullEvents()...)
	})
	if err != nil {
		return nil, err
	}
	return uc.queries.Details(ctx, id)
}

type UpdateUseCase struct {
	repo    domaintelf.Repository
	queries read.Queries
	tx      appshared.TxManager
	pub     appshared.EventPublisher
}

func (uc *UpdateUseCase) Execute(ctx context.Context, id int32, in UpdateInput) (*read.Details, error) {
	err := uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		t, err := uc.repo.FindByID(ctx, id)
		if err != nil {
			return err
		}
		if err := t.Update(in.toDomain()); err != nil {
			return err
		}
		if err := uc.repo.Save(ctx, t); err != nil {
			return err
		}
		return uc.pub.Publish(ctx, t.PullEvents()...)
	})
	if err != nil {
		return nil, err
	}
	return uc.queries.Details(ctx, id)
}

type AsignarUseCase struct {
	repo    domaintelf.Repository
	queries read.Queries
	tx      appshared.TxManager
	pub     appshared.EventPublisher
}

func (uc *AsignarUseCase) Execute(ctx context.Context, id int32, in AsignarInput) (*read.Details, error) {
	if in.TrabajadorID == 0 {
		return nil, apperrors.Invalid("trabajador_id es obligatorio")
	}
	err := uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		t, err := uc.repo.FindByID(ctx, id)
		if err != nil {
			return err
		}
		if err := t.Asignar(in.TrabajadorID); err != nil {
			return err
		}
		if err := uc.repo.Save(ctx, t); err != nil {
			return err
		}
		return uc.pub.Publish(ctx, t.PullEvents()...)
	})
	if err != nil {
		return nil, err
	}
	return uc.queries.Details(ctx, id)
}

type DesactivarUseCase struct {
	repo    domaintelf.Repository
	queries read.Queries
	tx      appshared.TxManager
	pub     appshared.EventPublisher
}

func (uc *DesactivarUseCase) Execute(ctx context.Context, id int32) (*read.Details, error) {
	err := uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		t, err := uc.repo.FindByID(ctx, id)
		if err != nil {
			return err
		}
		t.Desactivar()
		if err := uc.repo.Save(ctx, t); err != nil {
			return err
		}
		return uc.pub.Publish(ctx, t.PullEvents()...)
	})
	if err != nil {
		return nil, err
	}
	return uc.queries.Details(ctx, id)
}

type DeleteUseCase struct {
	repo domaintelf.Repository
	tx   appshared.TxManager
	pub  appshared.EventPublisher
}

func (uc *DeleteUseCase) Execute(ctx context.Context, id int32) error {
	return uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		t, err := uc.repo.FindByID(ctx, id)
		if err != nil {
			return err
		}
		t.Delete()
		if err := uc.repo.Delete(ctx, id); err != nil {
			return err
		}
		return uc.pub.Publish(ctx, t.PullEvents()...)
	})
}
