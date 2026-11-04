package telefonos

import (
	"context"

	"boilerplate-learner/internal/application/telefonos/read"
	"boilerplate-learner/internal/shared/apperrors"
)

type ListUseCase struct{ queries read.Queries }

func (uc *ListUseCase) Execute(ctx context.Context, f read.Filters) (*read.ListResult, error) {
	return uc.queries.List(ctx, f)
}

type GetUseCase struct{ queries read.Queries }

func (uc *GetUseCase) Execute(ctx context.Context, id int32) (*read.Details, error) {
	d, err := uc.queries.Details(ctx, id)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, apperrors.NotFound("teléfono no encontrado")
	}
	return d, nil
}

type ListEstadosUseCase struct{ queries read.Queries }

func (uc *ListEstadosUseCase) Execute(ctx context.Context) ([]read.EstadoSummary, error) {
	return uc.queries.Estados(ctx)
}

type ListTipologiasUseCase struct{ queries read.Queries }

func (uc *ListTipologiasUseCase) Execute(ctx context.Context) ([]read.TipologiaSummary, error) {
	return uc.queries.Tipologias(ctx)
}
