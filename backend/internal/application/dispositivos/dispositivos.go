// Package dispositivos is the application layer for the Dispositivo aggregate
// and its device catalog (brands/models). Write use cases run in a TxManager
// transaction and publish domain events through the outbox EventPublisher.
package dispositivos

import (
	"boilerplate-learner/internal/application/dispositivos/read"
	appshared "boilerplate-learner/internal/application/shared"
	domaindisp "boilerplate-learner/internal/domain/dispositivos"
)

type UseCases struct {
	List     *ListUseCase
	Get      *GetUseCase
	Create   *CreateUseCase
	Update   *UpdateUseCase
	Delete   *DeleteUseCase
	Entregar *EntregarUseCase

	ListBrands  *ListBrandsUseCase
	CreateBrand *CreateBrandUseCase
	UpdateBrand *UpdateBrandUseCase
	DeleteBrand *DeleteBrandUseCase
	ListModels  *ListModelsUseCase
	CreateModel *CreateModelUseCase
	UpdateModel *UpdateModelUseCase
	DeleteModel *DeleteModelUseCase
}

func NewUseCases(
	repo domaindisp.Repository,
	catalog domaindisp.CatalogRepository,
	queries read.Queries,
	tx appshared.TxManager,
	pub appshared.EventPublisher,
) *UseCases {
	return &UseCases{
		List:        &ListUseCase{queries: queries},
		Get:         &GetUseCase{queries: queries},
		Create:      &CreateUseCase{repo: repo, queries: queries, tx: tx, pub: pub},
		Update:      &UpdateUseCase{repo: repo, queries: queries, tx: tx, pub: pub},
		Delete:      &DeleteUseCase{repo: repo, tx: tx, pub: pub},
		Entregar:    &EntregarUseCase{repo: repo, queries: queries, tx: tx, pub: pub},
		ListBrands:  &ListBrandsUseCase{catalog: catalog},
		CreateBrand: &CreateBrandUseCase{catalog: catalog},
		UpdateBrand: &UpdateBrandUseCase{catalog: catalog},
		DeleteBrand: &DeleteBrandUseCase{catalog: catalog},
		ListModels:  &ListModelsUseCase{catalog: catalog},
		CreateModel: &CreateModelUseCase{catalog: catalog},
		UpdateModel: &UpdateModelUseCase{catalog: catalog},
		DeleteModel: &DeleteModelUseCase{catalog: catalog},
	}
}

// ── Input DTOs ───────────────────────────────────────────────────────────────

type CreateInput struct {
	Numero        *string `json:"numero"`
	IMEI          *string `json:"imei"`
	Serie         *string `json:"serie"`
	SynergyResID  *int32  `json:"synergy_res_id"`
	BrandID       *int32  `json:"brand_id"`
	ModelID       *int32  `json:"model_id"`
	Grupo         *string `json:"grupo"`
	Estado        string  `json:"estado"`
	TrabajadorID  *int32  `json:"trabajador_id"`
	Observaciones string  `json:"observaciones"`
}

func (in CreateInput) toDomain() domaindisp.CreateData {
	return domaindisp.CreateData{
		Numero: in.Numero, IMEI: in.IMEI, Serie: in.Serie, SynergyResID: in.SynergyResID,
		BrandID: in.BrandID, ModelID: in.ModelID, Grupo: in.Grupo, Estado: in.Estado,
		TrabajadorID: in.TrabajadorID, Observaciones: in.Observaciones,
	}
}

type UpdateInput struct {
	Numero        *string `json:"numero"`
	IMEI          *string `json:"imei"`
	Serie         *string `json:"serie"`
	SynergyResID  *int32  `json:"synergy_res_id"`
	BrandID       *int32  `json:"brand_id"`
	ModelID       *int32  `json:"model_id"`
	Grupo         *string `json:"grupo"`
	Estado        *string `json:"estado"`
	TrabajadorID  *int32  `json:"trabajador_id"`
	Observaciones *string `json:"observaciones"`
}

func (in UpdateInput) toDomain() domaindisp.UpdateData {
	return domaindisp.UpdateData{
		Numero: in.Numero, IMEI: in.IMEI, Serie: in.Serie, SynergyResID: in.SynergyResID,
		BrandID: in.BrandID, ModelID: in.ModelID, Grupo: in.Grupo, Estado: in.Estado,
		TrabajadorID: in.TrabajadorID, Observaciones: in.Observaciones,
	}
}

type EntregarInput struct {
	TrabajadorID int32 `json:"trabajador_id"`
}

type BrandInput struct {
	Nombre string `json:"nombre"`
}

type ModelInput struct {
	Nombre  string `json:"nombre"`
	BrandID int32  `json:"brand_id"`
}
