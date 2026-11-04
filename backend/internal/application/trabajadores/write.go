package trabajadores

import (
	"context"
	"time"

	appshared "boilerplate-learner/internal/application/shared"
	"boilerplate-learner/internal/application/trabajadores/read"
	domaintrab "boilerplate-learner/internal/domain/trabajadores"
	"boilerplate-learner/internal/shared/apperrors"
)

type CreateUseCase struct {
	repo    domaintrab.Repository
	queries read.Queries
	tx      appshared.TxManager
	pub     appshared.EventPublisher
}

func (uc *CreateUseCase) Execute(ctx context.Context, in CreateInput) (*read.Details, error) {
	data, err := in.toDomain()
	if err != nil {
		return nil, apperrors.Invalid("fecha_incorporacion inválida (formato YYYY-MM-DD)")
	}
	var id int32
	err = uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		t, err := domaintrab.NewTrabajador(data)
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
	repo    domaintrab.Repository
	queries read.Queries
	tx      appshared.TxManager
	pub     appshared.EventPublisher
}

func (uc *UpdateUseCase) Execute(ctx context.Context, id int32, in UpdateInput) (*read.Details, error) {
	data, err := in.toDomain()
	if err != nil {
		return nil, apperrors.Invalid("fecha inválida (formato YYYY-MM-DD)")
	}
	err = uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		t, err := uc.repo.FindByID(ctx, id)
		if err != nil {
			return err
		}
		if err := t.Update(data); err != nil {
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

type DarDeBajaUseCase struct {
	repo    domaintrab.Repository
	queries read.Queries
	tx      appshared.TxManager
	pub     appshared.EventPublisher
}

func (uc *DarDeBajaUseCase) Execute(ctx context.Context, id int32, fechaBaja *string) (*read.Details, error) {
	fecha := time.Now()
	if parsed, err := parseDate(fechaBaja); err != nil {
		return nil, apperrors.Invalid("fecha_baja inválida (formato YYYY-MM-DD)")
	} else if parsed != nil {
		fecha = *parsed
	}
	err := uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		t, err := uc.repo.FindByID(ctx, id)
		if err != nil {
			return err
		}
		if err := t.DarDeBaja(fecha); err != nil {
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

type DeleteUseCase struct {
	repo domaintrab.Repository
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
