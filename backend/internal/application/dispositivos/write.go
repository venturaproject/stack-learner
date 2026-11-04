package dispositivos

import (
	"context"

	"boilerplate-learner/internal/application/dispositivos/read"
	appshared "boilerplate-learner/internal/application/shared"
	domaindisp "boilerplate-learner/internal/domain/dispositivos"
	"boilerplate-learner/internal/shared/apperrors"
)

type CreateUseCase struct {
	repo    domaindisp.Repository
	queries read.Queries
	tx      appshared.TxManager
	pub     appshared.EventPublisher
}

func (uc *CreateUseCase) Execute(ctx context.Context, in CreateInput) (*read.Details, error) {
	var id int32
	err := uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		d, err := domaindisp.NewDispositivo(in.toDomain())
		if err != nil {
			return err
		}
		if err := uc.repo.Save(ctx, d); err != nil {
			return err
		}
		id = d.ID
		return uc.pub.Publish(ctx, d.PullEvents()...)
	})
	if err != nil {
		return nil, err
	}
	return uc.queries.Details(ctx, id)
}

type UpdateUseCase struct {
	repo    domaindisp.Repository
	queries read.Queries
	tx      appshared.TxManager
	pub     appshared.EventPublisher
}

func (uc *UpdateUseCase) Execute(ctx context.Context, id int32, in UpdateInput) (*read.Details, error) {
	err := uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		d, err := uc.repo.FindByID(ctx, id)
		if err != nil {
			return err
		}
		if err := d.Update(in.toDomain()); err != nil {
			return err
		}
		if err := uc.repo.Save(ctx, d); err != nil {
			return err
		}
		return uc.pub.Publish(ctx, d.PullEvents()...)
	})
	if err != nil {
		return nil, err
	}
	return uc.queries.Details(ctx, id)
}

type EntregarUseCase struct {
	repo    domaindisp.Repository
	queries read.Queries
	tx      appshared.TxManager
	pub     appshared.EventPublisher
}

func (uc *EntregarUseCase) Execute(ctx context.Context, id int32, in EntregarInput) (*read.Details, error) {
	if in.TrabajadorID == 0 {
		return nil, apperrors.Invalid("trabajador_id es obligatorio")
	}
	err := uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		d, err := uc.repo.FindByID(ctx, id)
		if err != nil {
			return err
		}
		if err := d.Entregar(in.TrabajadorID); err != nil {
			return err
		}
		if err := uc.repo.Save(ctx, d); err != nil {
			return err
		}
		return uc.pub.Publish(ctx, d.PullEvents()...)
	})
	if err != nil {
		return nil, err
	}
	return uc.queries.Details(ctx, id)
}

type DeleteUseCase struct {
	repo domaindisp.Repository
	tx   appshared.TxManager
	pub  appshared.EventPublisher
}

func (uc *DeleteUseCase) Execute(ctx context.Context, id int32) error {
	return uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		d, err := uc.repo.FindByID(ctx, id)
		if err != nil {
			return err
		}
		d.Delete()
		if err := uc.repo.Delete(ctx, id); err != nil {
			return err
		}
		return uc.pub.Publish(ctx, d.PullEvents()...)
	})
}
