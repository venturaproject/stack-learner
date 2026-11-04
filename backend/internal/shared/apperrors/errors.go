// Package apperrors defines a transport-neutral error taxonomy shared by the
// domain and application layers. It has no knowledge of HTTP; the mapping from
// Kind to an HTTP status code lives in internal/infrastructure/http/httpx.
package apperrors

import "errors"

// Kind classifies an error so outer layers can translate it (HTTP status,
// gRPC code, CLI exit code, …) without inspecting error strings.
type Kind int

const (
	KindInternal Kind = iota
	KindNotFound
	KindInvalid
	KindConflict
	KindUnauthorized
	KindForbidden
)

// Error is the concrete application/domain error. It carries a Kind, a
// human-readable message and an optional wrapped cause (for errors.Is / As).
type Error struct {
	Kind Kind
	Msg  string
	Err  error
}

func (e *Error) Error() string {
	switch {
	case e.Msg != "" && e.Err != nil:
		return e.Msg + ": " + e.Err.Error()
	case e.Msg != "":
		return e.Msg
	case e.Err != nil:
		return e.Err.Error()
	default:
		return "error"
	}
}

func (e *Error) Unwrap() error { return e.Err }

func makeErr(k Kind, msg string) *Error { return &Error{Kind: k, Msg: msg} }

// Constructors — one per Kind.
func NotFound(msg string) *Error     { return makeErr(KindNotFound, msg) }
func Invalid(msg string) *Error      { return makeErr(KindInvalid, msg) }
func Conflict(msg string) *Error     { return makeErr(KindConflict, msg) }
func Unauthorized(msg string) *Error { return makeErr(KindUnauthorized, msg) }
func Forbidden(msg string) *Error    { return makeErr(KindForbidden, msg) }
func Internal(msg string) *Error     { return makeErr(KindInternal, msg) }

// BadRequest is a deprecated alias for Invalid, kept for call-site stability.
func BadRequest(msg string) *Error { return makeErr(KindInvalid, msg) }

// Wrap attaches a Kind and message to an existing error, preserving the cause.
func Wrap(k Kind, msg string, err error) *Error { return &Error{Kind: k, Msg: msg, Err: err} }

// KindOf returns the Kind of err, unwrapping as needed. Unknown errors are
// treated as KindInternal.
func KindOf(err error) Kind {
	var e *Error
	if errors.As(err, &e) {
		return e.Kind
	}
	return KindInternal
}
