package assessment

import (
	"context"
	"math"
	"strings"

	assessmentread "boilerplate-learner/internal/application/assessment/read"
	appshared "boilerplate-learner/internal/application/shared"
	domainassessment "boilerplate-learner/internal/domain/assessment"
	"boilerplate-learner/internal/domain/organizations"
	"boilerplate-learner/internal/shared/apperrors"
)

type PageQuery struct {
	Page    int
	PerPage int
	Search  string
}

type PageResult[T any] struct {
	Items      []T   `json:"items"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type Repository interface {
	ActiveOrganizationRole(context.Context, int32, int32) (string, error)
	ValidateStudentSubmissionAccess(context.Context, int32, int32, int32, int32) error
	CourseTrainerAssigned(context.Context, int32, int32, int32) (bool, error)
	CreateSubmission(context.Context, int32, int32, int32, int32, string, *string) (*domainassessment.Submission, error)
	ListStudentSubmissions(context.Context, int32, int32, int32, int32) ([]domainassessment.Submission, error)
	ListCourseSubmissions(context.Context, int32, int32, int32, PageQuery) ([]domainassessment.Submission, int64, error)
	ReviewSubmission(context.Context, int32, int32, int64, int32, float64, string) (*domainassessment.Submission, error)
}

type SubmitInput struct {
	ResponseText string `json:"response_text"`
	ResponseURL  string `json:"response_url"`
}

type ReviewInput struct {
	Score    float64 `json:"score"`
	Feedback string  `json:"feedback"`
}

type Service struct {
	repo Repository
	tx   appshared.TxManager
}

func NewService(repo Repository, tx appshared.TxManager) *Service {
	return &Service{repo: repo, tx: tx}
}

func (s *Service) Submit(ctx context.Context, userID, orgID, courseID, lessonID int32, input SubmitInput) (*assessmentread.Submission, error) {
	input.ResponseText = strings.TrimSpace(input.ResponseText)
	input.ResponseURL = strings.TrimSpace(input.ResponseURL)
	if err := domainassessment.ValidateResponse(input.ResponseText, input.ResponseURL); err != nil {
		return nil, err
	}
	return withinTx(ctx, s.tx, func(txctx context.Context) (*assessmentread.Submission, error) {
		role, err := s.repo.ActiveOrganizationRole(txctx, userID, orgID)
		if err != nil {
			return nil, err
		}
		if role != organizations.RoleStudent {
			return nil, apperrors.Forbidden("solo los estudiantes pueden enviar entregas")
		}
		if err := s.repo.ValidateStudentSubmissionAccess(txctx, orgID, courseID, lessonID, userID); err != nil {
			return nil, err
		}
		var responseURL *string
		if input.ResponseURL != "" {
			responseURL = &input.ResponseURL
		}
		submission, err := s.repo.CreateSubmission(txctx, orgID, courseID, lessonID, userID, input.ResponseText, responseURL)
		if err != nil {
			return nil, err
		}
		return toRead(submission), nil
	})
}

func (s *Service) ListStudentSubmissions(ctx context.Context, userID, orgID, courseID, lessonID int32) ([]assessmentread.Submission, error) {
	return withinTx(ctx, s.tx, func(txctx context.Context) ([]assessmentread.Submission, error) {
		role, err := s.repo.ActiveOrganizationRole(txctx, userID, orgID)
		if err != nil {
			return nil, err
		}
		if role != organizations.RoleStudent {
			return nil, apperrors.Forbidden("el historial personal solo está disponible para estudiantes")
		}
		if err := s.repo.ValidateStudentSubmissionAccess(txctx, orgID, courseID, lessonID, userID); err != nil {
			return nil, err
		}
		items, err := s.repo.ListStudentSubmissions(txctx, orgID, courseID, lessonID, userID)
		if err != nil {
			return nil, err
		}
		return toReadList(items), nil
	})
}

func (s *Service) ListCourseSubmissions(ctx context.Context, userID, orgID, courseID int32, query PageQuery) (PageResult[assessmentread.Submission], error) {
	query = normalizePage(query)
	return withinTx(ctx, s.tx, func(txctx context.Context) (PageResult[assessmentread.Submission], error) {
		role, err := s.repo.ActiveOrganizationRole(txctx, userID, orgID)
		if err != nil {
			return PageResult[assessmentread.Submission]{}, err
		}
		if role == organizations.RoleTrainer {
			assigned, err := s.repo.CourseTrainerAssigned(txctx, orgID, courseID, userID)
			if err != nil {
				return PageResult[assessmentread.Submission]{}, err
			}
			if !assigned {
				return PageResult[assessmentread.Submission]{}, apperrors.NotFound("curso no encontrado")
			}
		} else if role != organizations.RoleAdmin {
			return PageResult[assessmentread.Submission]{}, apperrors.Forbidden("no tienes permisos para revisar entregas")
		}
		items, total, err := s.repo.ListCourseSubmissions(txctx, orgID, courseID, 0, query)
		if err != nil {
			return PageResult[assessmentread.Submission]{}, err
		}
		result := PageResult[assessmentread.Submission]{Items: toReadList(items), Page: query.Page, PerPage: query.PerPage, Total: total}
		if total > 0 {
			result.TotalPages = int(math.Ceil(float64(total) / float64(query.PerPage)))
		}
		return result, nil
	})
}

func (s *Service) Review(ctx context.Context, userID, orgID, courseID int32, submissionID int64, input ReviewInput) (*assessmentread.Submission, error) {
	input.Feedback = strings.TrimSpace(input.Feedback)
	if err := domainassessment.ValidateReview(input.Score, input.Feedback); err != nil {
		return nil, err
	}
	return withinTx(ctx, s.tx, func(txctx context.Context) (*assessmentread.Submission, error) {
		role, err := s.repo.ActiveOrganizationRole(txctx, userID, orgID)
		if err != nil {
			return nil, err
		}
		if role == organizations.RoleTrainer {
			assigned, err := s.repo.CourseTrainerAssigned(txctx, orgID, courseID, userID)
			if err != nil {
				return nil, err
			}
			if !assigned {
				return nil, apperrors.NotFound("curso no encontrado")
			}
		} else if role != organizations.RoleAdmin {
			return nil, apperrors.Forbidden("no tienes permisos para revisar entregas")
		}
		submission, err := s.repo.ReviewSubmission(txctx, orgID, courseID, submissionID, userID, input.Score, input.Feedback)
		if err != nil {
			return nil, err
		}
		return toRead(submission), nil
	})
}

func normalizePage(query PageQuery) PageQuery {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PerPage < 1 {
		query.PerPage = 25
	}
	if query.PerPage > 100 {
		query.PerPage = 100
	}
	query.Search = strings.TrimSpace(query.Search)
	return query
}

func toRead(item *domainassessment.Submission) *assessmentread.Submission {
	if item == nil {
		return nil
	}
	return &assessmentread.Submission{
		ID: item.ID, OrganizationID: item.OrganizationID, CourseID: item.CourseID,
		LessonID: item.LessonID, LessonTitle: item.LessonTitle, StudentID: item.StudentID,
		StudentName: item.StudentName, StudentEmail: item.StudentEmail, Attempt: item.Attempt,
		ResponseText: item.ResponseText, ResponseURL: item.ResponseURL, Status: item.Status,
		Score: item.Score, Feedback: item.Feedback, SubmittedAt: item.SubmittedAt,
		ReviewedBy: item.ReviewedBy, ReviewedAt: item.ReviewedAt,
	}
}

func toReadList(items []domainassessment.Submission) []assessmentread.Submission {
	result := make([]assessmentread.Submission, 0, len(items))
	for i := range items {
		result = append(result, *toRead(&items[i]))
	}
	return result
}

func withinTx[T any](ctx context.Context, tx appshared.TxManager, action func(context.Context) (T, error)) (T, error) {
	var result T
	if tx == nil {
		return action(ctx)
	}
	err := tx.WithinTx(ctx, func(txctx context.Context) error {
		var err error
		result, err = action(txctx)
		return err
	})
	return result, err
}
