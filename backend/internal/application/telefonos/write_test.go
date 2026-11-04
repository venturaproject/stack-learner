package telefonos

import (
	"context"
	"errors"
	"testing"

	telfread "boilerplate-learner/internal/application/telefonos/read"
	domainshared "boilerplate-learner/internal/domain/shared"
	domaintelf "boilerplate-learner/internal/domain/telefonos"
)

// ── Fakes ────────────────────────────────────────────────────────────────────

type fakeRepo struct {
	store   map[int32]*domaintelf.Telefono
	nextID  int32
	deleted []int32
}

func newFakeRepo() *fakeRepo { return &fakeRepo{store: map[int32]*domaintelf.Telefono{}, nextID: 1} }

func (r *fakeRepo) FindByID(_ context.Context, id int32) (*domaintelf.Telefono, error) {
	t, ok := r.store[id]
	if !ok {
		return nil, domaintelf.ErrNotFound
	}
	cp := *t
	return &cp, nil
}

func (r *fakeRepo) Save(_ context.Context, t *domaintelf.Telefono) error {
	if t.ID == 0 {
		t.ID = r.nextID
		r.nextID++
	}
	cp := *t
	r.store[t.ID] = &cp
	return nil
}

func (r *fakeRepo) Delete(_ context.Context, id int32) error {
	delete(r.store, id)
	r.deleted = append(r.deleted, id)
	return nil
}

type directTx struct{}

func (directTx) WithinTx(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }

type recordingPub struct{ events []domainshared.DomainEvent }

func (p *recordingPub) Publish(_ context.Context, evs ...domainshared.DomainEvent) error {
	p.events = append(p.events, evs...)
	return nil
}

type fakeQueries struct{ detail *telfread.Details }

func (q fakeQueries) List(context.Context, telfread.Filters) (*telfread.ListResult, error) {
	return &telfread.ListResult{}, nil
}
func (q fakeQueries) Details(context.Context, int32) (*telfread.Details, error) { return q.detail, nil }
func (q fakeQueries) Estados(context.Context) ([]telfread.EstadoSummary, error) { return nil, nil }
func (q fakeQueries) Tipologias(context.Context) ([]telfread.TipologiaSummary, error) {
	return nil, nil
}

// ── Tests ────────────────────────────────────────────────────────────────────

func TestCreatePublishesEventAfterSave(t *testing.T) {
	repo := newFakeRepo()
	pub := &recordingPub{}
	uc := &CreateUseCase{repo: repo, queries: fakeQueries{detail: &telfread.Details{}}, tx: directTx{}, pub: pub}

	if _, err := uc.Execute(context.Background(), CreateInput{Activo: true}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.store) != 1 {
		t.Fatalf("telefono not persisted")
	}
	if len(pub.events) != 1 || pub.events[0].EventName() != "telefono.created" {
		t.Fatalf("unexpected events: %+v", pub.events)
	}
}

func TestUpdateNotFoundPropagates(t *testing.T) {
	uc := &UpdateUseCase{repo: newFakeRepo(), queries: fakeQueries{}, tx: directTx{}, pub: &recordingPub{}}
	_, err := uc.Execute(context.Background(), 999, UpdateInput{})
	if !errors.Is(err, domaintelf.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestDeleteRemovesThenPublishes(t *testing.T) {
	repo := newFakeRepo()
	_ = repo.Save(context.Background(), &domaintelf.Telefono{Activo: true})
	pub := &recordingPub{}
	uc := &DeleteUseCase{repo: repo, tx: directTx{}, pub: pub}

	if err := uc.Execute(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.deleted) != 1 || repo.deleted[0] != 1 {
		t.Fatalf("row not deleted: %+v", repo.deleted)
	}
	if len(pub.events) != 1 || pub.events[0].EventName() != "telefono.deleted" {
		t.Fatalf("unexpected events: %+v", pub.events)
	}
}
