package dispositivos

import (
	"context"

	"boilerplate-learner/internal/application/dispositivos/read"
	domaindisp "boilerplate-learner/internal/domain/dispositivos"
	"boilerplate-learner/internal/shared/apperrors"
)

func brandSummary(b *domaindisp.Brand) read.BrandSummary {
	return read.BrandSummary{ID: b.ID, Nombre: b.Nombre}
}

func modelSummary(m *domaindisp.Model) read.ModelSummary {
	return read.ModelSummary{ID: m.ID, Nombre: m.Nombre, BrandID: m.BrandID}
}

type ListBrandsUseCase struct{ catalog domaindisp.CatalogRepository }

func (uc *ListBrandsUseCase) Execute(ctx context.Context) ([]read.BrandSummary, error) {
	bs, err := uc.catalog.ListBrands(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]read.BrandSummary, 0, len(bs))
	for i := range bs {
		out = append(out, brandSummary(&bs[i]))
	}
	return out, nil
}

type CreateBrandUseCase struct{ catalog domaindisp.CatalogRepository }

func (uc *CreateBrandUseCase) Execute(ctx context.Context, in BrandInput) (*read.BrandSummary, error) {
	if in.Nombre == "" {
		return nil, apperrors.Invalid("el nombre es obligatorio")
	}
	b, err := uc.catalog.CreateBrand(ctx, in.Nombre)
	if err != nil {
		return nil, err
	}
	s := brandSummary(b)
	return &s, nil
}

type UpdateBrandUseCase struct{ catalog domaindisp.CatalogRepository }

func (uc *UpdateBrandUseCase) Execute(ctx context.Context, id int32, in BrandInput) (*read.BrandSummary, error) {
	if in.Nombre == "" {
		return nil, apperrors.Invalid("el nombre es obligatorio")
	}
	b, err := uc.catalog.UpdateBrand(ctx, id, in.Nombre)
	if err != nil {
		return nil, err
	}
	s := brandSummary(b)
	return &s, nil
}

type DeleteBrandUseCase struct{ catalog domaindisp.CatalogRepository }

func (uc *DeleteBrandUseCase) Execute(ctx context.Context, id int32) error {
	return uc.catalog.DeleteBrand(ctx, id)
}

type ListModelsUseCase struct{ catalog domaindisp.CatalogRepository }

func (uc *ListModelsUseCase) Execute(ctx context.Context, brandID *int32) ([]read.ModelSummary, error) {
	ms, err := uc.catalog.ListModels(ctx, brandID)
	if err != nil {
		return nil, err
	}
	out := make([]read.ModelSummary, 0, len(ms))
	for i := range ms {
		out = append(out, modelSummary(&ms[i]))
	}
	return out, nil
}

type CreateModelUseCase struct{ catalog domaindisp.CatalogRepository }

func (uc *CreateModelUseCase) Execute(ctx context.Context, in ModelInput) (*read.ModelSummary, error) {
	if in.Nombre == "" || in.BrandID == 0 {
		return nil, apperrors.Invalid("nombre y brand_id son obligatorios")
	}
	m, err := uc.catalog.CreateModel(ctx, in.Nombre, in.BrandID)
	if err != nil {
		return nil, err
	}
	s := modelSummary(m)
	return &s, nil
}

type UpdateModelUseCase struct{ catalog domaindisp.CatalogRepository }

func (uc *UpdateModelUseCase) Execute(ctx context.Context, id int32, in ModelInput) (*read.ModelSummary, error) {
	if in.Nombre == "" || in.BrandID == 0 {
		return nil, apperrors.Invalid("nombre y brand_id son obligatorios")
	}
	m, err := uc.catalog.UpdateModel(ctx, id, in.Nombre, in.BrandID)
	if err != nil {
		return nil, err
	}
	s := modelSummary(m)
	return &s, nil
}

type DeleteModelUseCase struct{ catalog domaindisp.CatalogRepository }

func (uc *DeleteModelUseCase) Execute(ctx context.Context, id int32) error {
	return uc.catalog.DeleteModel(ctx, id)
}
