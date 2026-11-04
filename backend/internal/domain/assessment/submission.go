package assessment

import (
	"math"
	"net/url"
	"strings"
	"time"

	"boilerplate-learner/internal/shared/apperrors"
)

const (
	StatusSubmitted  = "submitted"
	StatusReviewed   = "reviewed"
	MaxResponseRunes = 20000
	MaxFeedbackRunes = 5000
)

type Submission struct {
	ID             int64
	OrganizationID int32
	CourseID       int32
	LessonID       int32
	StudentID      int32
	StudentName    string
	StudentEmail   string
	LessonTitle    string
	Attempt        int
	ResponseText   string
	ResponseURL    *string
	Status         string
	Score          *float64
	Feedback       *string
	SubmittedAt    time.Time
	ReviewedBy     *int32
	ReviewedAt     *time.Time
}

func ValidateResponse(text, rawURL string) error {
	text = strings.TrimSpace(text)
	rawURL = strings.TrimSpace(rawURL)
	if text == "" && rawURL == "" {
		return apperrors.Invalid("escribe una respuesta o añade un enlace")
	}
	if len([]rune(text)) > MaxResponseRunes {
		return apperrors.Invalid("la respuesta no puede superar 20000 caracteres")
	}
	if len(rawURL) > 2048 {
		return apperrors.Invalid("el enlace no puede superar 2048 caracteres")
	}
	if rawURL != "" {
		parsed, err := url.ParseRequestURI(rawURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return apperrors.Invalid("el enlace debe ser una URL http o https válida")
		}
	}
	return nil
}

func ValidateReview(score float64, feedback string) error {
	if math.IsNaN(score) || math.IsInf(score, 0) || score < 0 || score > 100 {
		return apperrors.Invalid("la nota debe estar entre 0 y 100")
	}
	if math.Abs(score*100-math.Round(score*100)) > 1e-8 {
		return apperrors.Invalid("la nota admite como máximo dos decimales")
	}
	if len([]rune(feedback)) > MaxFeedbackRunes {
		return apperrors.Invalid("la retroalimentación no puede superar 5000 caracteres")
	}
	return nil
}

func (s *Submission) Review(reviewerID int32, score float64, feedback string, reviewedAt time.Time) error {
	if s.Status != StatusSubmitted {
		return apperrors.Conflict("esta entrega ya fue revisada")
	}
	if err := ValidateReview(score, feedback); err != nil {
		return err
	}
	s.Status = StatusReviewed
	s.Score = &score
	feedback = strings.TrimSpace(feedback)
	s.Feedback = &feedback
	s.ReviewedBy = &reviewerID
	s.ReviewedAt = &reviewedAt
	return nil
}
