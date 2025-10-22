package response

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spf13/cast"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smart-kart/framework/logger"
	"github.com/smart-kart/framework/utils/generic"
	protov1 "github.com/smart-kart/proto/gen/go/proto/v1"
)

// generic data type being used only in place of grpc status error
type t struct{}

// Remarks field for the Err object
type Remarks string

// ErrCode error code series
type ErrCode int32

// ErrType validation error type
type ErrType string

// variables get injected from service initialization using Load()
//
//nolint:gochecknoglobals // required to load variables once
var (
	// validation error codes w.r.t type & field
	_validationErrCode = make(map[ErrType]map[string]ErrCode)

	// error code - message
	_errMsg = make(map[ErrCode]string)

	// Empty generic data object being used only in place of grpc status error
	Empty = t{}
)

// LoadErrCode load service error codes
func LoadErrCode(em map[ErrCode]string, vec map[ErrType]map[string]ErrCode) {
	// fix: panic: assignment to entry in nil map
	if em == nil {
		em = make(map[ErrCode]string)
	}

	if vec == nil {
		vec = make(map[ErrType]map[string]ErrCode)
	}

	// adding general error code - msg to the service list
	for k, v := range ErrMsg {
		em[k] = v
	}

	_errMsg = em
	_validationErrCode = vec
}

// RegisterErrMsg registers error messages at package initialization
func RegisterErrMsg(em map[ErrCode]string) {
	for k, v := range em {
		_errMsg[k] = v
	}
}

// RegisterFieldErrCode registers field error codes at package initialization
func RegisterFieldErrCode(vec map[ErrType]map[string]ErrCode) {
	for errType, fields := range vec {
		if _validationErrCode[errType] == nil {
			_validationErrCode[errType] = make(map[string]ErrCode)
		}
		for field, code := range fields {
			_validationErrCode[errType][field] = code
		}
	}
}

// GetValidationErrCode return custom error code against json field
// with an underlined type
func GetValidationErrCode(errType ErrType, jsonTag string) ErrCode {
	return _validationErrCode[errType][jsonTag]
}

// GetErrMsg return message defined for the custom error code
func GetErrMsg(errCode ErrCode) string {
	return _errMsg[errCode]
}

// ReadGRPCError parse and return gRPC native error (code XX, message,
// details [if any] (array of custom Err object - code XXXX, message, remarks [if any]))
func ReadGRPCError(err error) *protov1.GRPCError {
	// define empty object
	v := new(protov1.GRPCError)

	// check whether error is nil
	if err == nil {
		return v
	}

	// get the Status representation of err
	st := status.Convert(err)

	// convert s's status as an spb.Status proto message
	pb := st.Proto()

	// transform details (anypb) to Err object
	details := make([]*protov1.Err, 0)
	var mErr error

	for _, v := range pb.GetDetails() {
		var m protov1.Err

		mErr = anypb.UnmarshalTo(v, &m, proto.UnmarshalOptions{})
		if mErr != nil {
			logger.RestrictedGet().Error("failed to unmarshal error detail", mErr, "detail", v)
			continue // ok to skip failed detail, we used to concentrate more on code & message
		}

		details = append(details, &m)
	}

	return &protov1.GRPCError{
		Code:    pb.GetCode(),
		Message: pb.GetMessage(),
		Details: details,
	}
}

// ExtractGRPCError filter out and return only gRPC error
func ExtractGRPCError(_ any, err error) error {
	return err
}

// buildErr construct Err object with errCode, errMsg and optional remarks
func buildErr(errCode ErrCode, args ...any) *protov1.Err {
	var remarks Remarks

	// iterate args and identify Remarks
	for _, arg := range args {
		if v, ok := arg.(Remarks); ok {
			remarks = v
		}
	}

	return &protov1.Err{
		Code:    int32(errCode),
		Message: _errMsg[errCode],
		Remarks: string(remarks),
	}
}

// FormatErr construct Err object with errCode, formatted errMsg
func FormatErr(errCode ErrCode, args ...any) *protov1.Err {
	return &protov1.Err{
		Code:    int32(errCode),
		Message: fmt.Sprintf(_errMsg[errCode], args...),
	}
}

// FormatErrWithRemarks construct Err object with errCode, formatted errMsg and mandate remarks
func FormatErrWithRemarks(errCode ErrCode, remarks Remarks, args ...any) *protov1.Err {
	err := FormatErr(errCode, args...)
	err.Remarks = string(remarks)

	return err
}

// parseErr returns slice of Err object
func parseErr(args ...any) []*protov1.Err {
	e := make([]*protov1.Err, 0)

	for _, arg := range args {
		switch v := arg.(type) {
		case ErrCode:
			e = append(e, buildErr(v, args...))

		case []*protov1.Err:
			e = v

		case []protov1.Err:
			for i := range v {
				e = append(e, &v[i])
			}

		case protov1.Err:
			e = []*protov1.Err{&v}

		case *protov1.Err:
			e = []*protov1.Err{v}
		}
	}

	return e
}

/*
	Go 1.18 introduces `any` as an alias to `interface{}`

	Generics are not a replacement for interfaces.
	Generics are designed to work with interfaces and make Go more type-safe,
	and can also be used to eliminate code repetition (i.e _, ok := val.(*struct)).

	[T any]
		- T : type parameter
		- any : constraint on type
*/

// Success grpc-code: 0 ; http-code: 200
// The request has succeeded.
// @param context: relevant server context
// @param res: data to be consumed
func Success[T any](_ context.Context, res T) (T, error) {
	return res, nil
}

// Created grpc-code: 0 ; http-code: 201
// The request has been fulfilled and has resulted in one or more new resources being created.
// @param context: relevant server context
// @param res: data to be consumed
//
//nolint:unparam,errcheck // trade-off and better API experience
func Created[T any](ctx context.Context, res T) (T, error) {
	grpc.SetHeader(ctx, metadata.Pairs(mdHTTPStatusCode, cast.ToString(http.StatusCreated)))
	return res, nil
}

// Accepted grpc-code: 0 ; http-code: 202
// The request has been accepted for processing, but the processing has not been completed.
// @param context: relevant server context
// @param res: data to be consumed
//
//nolint:unparam,errcheck // trade-off and better API experience
func Accepted[T any](ctx context.Context, res T) (T, error) {
	grpc.SetHeader(ctx, metadata.Pairs(mdHTTPStatusCode, cast.ToString(http.StatusAccepted)))
	return res, nil
}

// e build common error response
// @param context: relevant server context
// @param res: rpc method return type; always be an empty or nil value [using for the trade-off]
// @param code: unsigned 32-bit error code as defined in the gRPC spec
// @param msg: associated message with the gRPC code
// args...
// @type errCode: custom four-digit [XXXX] series error code
// @type Err(s): custom err object to tell what exactly happened
func e[T any](_ context.Context, res T, code codes.Code, msg string, args ...any) (T, error) {
	// build error details
	details := parseErr(args...)

	// create status with code & msg
	st := status.New(code, msg)

	// add details
	var stErr error
	for _, detail := range details {
		st, stErr = st.WithDetails(detail)
		if stErr != nil {
			// fallback
			return generic.ReturnZero(res), status.New(codes.Internal, msgInternalServerError).Err()
		}
	}

	return generic.ReturnZero(res), st.Err()
}

// StrictError build error response with strict error object
// @param context: relevant server context
// @param res: rpc method return type; always be an empty or nil value [using for the trade-off]
// @param code: unsigned 32-bit error code as defined in the gRPC spec
// @param msg: associated message with the gRPC code
// @param strictErr: strict error object to use as HTTP response body
func StrictError[T any](_ context.Context, res T, code codes.Code, msg string, strictErr *protov1.StrictErr,
) (T, error) {
	// create status with code & msg
	st := status.New(code, msg)

	// add details
	st, stErr := st.WithDetails(strictErr)
	if stErr != nil {
		// fallback
		return generic.ReturnZero(res), status.New(codes.Internal, msgInternalServerError).Err()
	}

	return generic.ReturnZero(res), st.Err()
}

// GRPCError handle already constructed grpc status error, fallback to InternalError.
// @param context: relevant server context
// @param res: rpc method return type; always be an empty or nil value [using for the trade-off]
// @param err: grpc status error
func GRPCError[T any](ctx context.Context, res T, err error) (T, error) {
	// check whether grpc status error
	if _, ok := status.FromError(err); ok {
		return generic.ReturnZero(res), err
	}

	// fallback to internal error
	return InternalError(ctx, res)
}

// Canceled grpc-code: 1 ; http-code: 499
// The client has closed the connection while the server is still processing the request.
// @param context: relevant server context
// @param res: rpc method return type; always be an empty or nil value [using for the trade-off]
// args...
// @type errCode: custom four-digit [XXXX] series error code
// @type Err(s): custom err object to tell what exactly happened
func Canceled[T any](ctx context.Context, res T, args ...any) (T, error) {
	return e(ctx, res, codes.Canceled, msgCanceled, args...)
}

// Unknown grpc-code: 2 ; http-code: 500
// @param context: relevant server context
// @param res: rpc method return type; always be an empty or nil value [using for the trade-off]
// args...
// @type errCode: custom four-digit [XXXX] series error code
// @type Err(s): custom err object to tell what exactly happened
func Unknown[T any](ctx context.Context, res T, args ...any) (T, error) {
	return e(ctx, res, codes.Internal, msgInternalServerError, args...)
}

// InvalidArgument grpc-code: 3 ; http-code: 400
// The server cannot or will not process the request due to something that is perceived to be a client error
// (e.g., malformed request syntax, invalid request message framing, or deceptive request routing).
// @param context: relevant server context
// @param res: rpc method return type; always be an empty or nil value [using for the trade-off]
// args...
// @type errCode: custom four-digit [XXXX] series error code
// @type Err(s): custom err object to tell what exactly happened
func InvalidArgument[T any](ctx context.Context, res T, args ...any) (T, error) {
	return e(ctx, res, codes.InvalidArgument, msgBadRequest, args...)
}

// DeadlineExceeded grpc-code: 4 ; http-code: 504
// The server, while acting as a gateway or proxy, did not receive a timely response from an
// upstream server it needed to access in order to complete the request.
// @param context: relevant server context
// @param res: rpc method return type; always be an empty or nil value [using for the trade-off]
// args...
// @type errCode: custom four-digit [XXXX] series error code
// @type Err(s): custom err object to tell what exactly happened
func DeadlineExceeded[T any](ctx context.Context, res T, args ...any) (T, error) {
	return e(ctx, res, codes.DeadlineExceeded, msgTimeout, args...)
}

// NotFound grpc-code: 5 ; http-code: 404
// The origin server did not find a current representation for the target resource.
// @param context: relevant server context
// @param res: rpc method return type; always be an empty or nil value [using for the trade-off]
// args...
// @type errCode: custom four-digit [XXXX] series error code
// @type Err(s): custom err object to tell what exactly happened
func NotFound[T any](ctx context.Context, res T, args ...any) (T, error) {
	return e(ctx, res, codes.NotFound, msgNotFound, args...)
}

// AlreadyExists grpc-code: 6 ; http-code: 409
// The request could not be completed due to a conflict with the current state of the resource.
// Ex: resource already exists or should be in this state to process the resource.
// @param context: relevant server context
// @param res: rpc method return type; always be an empty or nil value [using for the trade-off]
// args...
// @type errCode: custom four-digit [XXXX] series error code
// @type Err(s): custom err object to tell what exactly happened
func AlreadyExists[T any](ctx context.Context, res T, args ...any) (T, error) {
	return e(ctx, res, codes.AlreadyExists, msgConflict, args...)
}

// PermissionDenied grpc-code: 7 ; http-code: 403
// The server understood the request but refuses to authorize it.
// Primarily due to a lack of permission to access the requested resource.
// @param context: relevant server context
// @param res: rpc method return type; always be an empty or nil value [using for the trade-off]
// args...
// @type errCode: custom four-digit [XXXX] series error code
// @type Err(s): custom err object to tell what exactly happened
func PermissionDenied[T any](ctx context.Context, res T, args ...any) (T, error) {
	return e(ctx, res, codes.PermissionDenied, msgForbidden, args...)
}

// ResourceExhausted grpc-code: 8 ; http-code: 429
// The user has sent too many requests in a given amount of time ("rate limiting").
// @param context: relevant server context
// @param res: rpc method return type; always be an empty or nil value [using for the trade-off]
// args...
// @type errCode: custom four-digit [XXXX] series error code
// @type Err(s): custom err object to tell what exactly happened
func ResourceExhausted[T any](ctx context.Context, res T, args ...any) (T, error) {
	return e(ctx, res, codes.ResourceExhausted, msgTooManyRequest, args...)
}

// FailedPrecondition grpc-code: 9 ; http-code: 400
// The server cannot or will not process the request due to something that is perceived to be a client error
// (e.g., malformed request syntax, invalid request message framing, or deceptive request routing).
// @param context: relevant server context
// @param res: rpc method return type; always be an empty or nil value [using for the trade-off]
// args...
// @type errCode: custom four-digit [XXXX] series error code
// @type Err(s): custom err object to tell what exactly happened
func FailedPrecondition[T any](ctx context.Context, res T, args ...any) (T, error) {
	return e(ctx, res, codes.FailedPrecondition, msgBadRequest, args...)
}

// Aborted grpc-code: 10 ; http-code: 409
// The request could not be completed due to a conflict with the current state of the resource.
// Ex: resource already exists or should be in this state to process the resource.
// @param context: relevant server context
// @param res: rpc method return type; always be an empty or nil value [using for the trade-off]
// args...
// @type errCode: custom four-digit [XXXX] series error code
// @type Err(s): custom err object to tell what exactly happened
func Aborted[T any](ctx context.Context, res T, args ...any) (T, error) {
	return e(ctx, res, codes.Aborted, msgConflict, args...)
}

// OutOfRange grpc-code: 11 ; http-code: 400
// The server cannot or will not process the request due to something that is perceived to be a client error
// (e.g., malformed request syntax, invalid request message framing, or deceptive request routing).
// @param context: relevant server context
// @param res: rpc method return type; always be an empty or nil value [using for the trade-off]
// args...
// @type errCode: custom four-digit [XXXX] series error code
// @type Err(s): custom err object to tell what exactly happened
func OutOfRange[T any](ctx context.Context, res T, args ...any) (T, error) {
	return e(ctx, res, codes.OutOfRange, msgBadRequest, args...)
}

// Unimplemented grpc-code: 12 ; http-code: 501
// The server does not support the functionality required to fulfil the request.
// @param context: relevant server context
// @param res: rpc method return type; always be an empty or nil value [using for the trade-off]
// args...
// @type errCode: custom four-digit [XXXX] series error code
// @type Err(s): custom err object to tell what exactly happened
func Unimplemented[T any](ctx context.Context, res T, args ...any) (T, error) {
	return e(ctx, res, codes.Unimplemented, msgNotImplemented, args...)
}

// InternalError grpc-code: 13 ; http-code: 500
// @param context: relevant server context
// @param res: rpc method return type; always be an empty or nil value [using for the trade-off]
// args...
// @type errCode: custom four-digit [XXXX] series error code
// @type Err(s): custom err object to tell what exactly happened
func InternalError[T any](ctx context.Context, res T, args ...any) (T, error) {
	return e(ctx, res, codes.Internal, msgInternalServerError, args...)
}

// Unavailable grpc-code: 14 ; http-code: 503
// The server is currently unable to handle the request due to a temporary overload
// or scheduled maintenance, which will likely be alleviated after some delay.
// @param context: relevant server context
// @param res: rpc method return type; always be an empty or nil value [using for the trade-off]
// args...
// @type errCode: custom four-digit [XXXX] series error code
// @type Err(s): custom err object to tell what exactly happened
func Unavailable[T any](ctx context.Context, res T, args ...any) (T, error) {
	return e(ctx, res, codes.Unavailable, msgUnavailable, args...)
}

// DataLoss grpc-code: 15 ; http-code: 500
//
// @param context: relevant server context
// @param res: rpc method return type; always be an empty or nil value [using for the trade-off]
// args...
// @type errCode: custom four-digit [XXXX] series error code
// @type Err(s): custom err object to tell what exactly happened
func DataLoss[T any](ctx context.Context, res T, args ...any) (T, error) {
	return e(ctx, res, codes.DataLoss, msgInternalServerError, args...)
}

// Unauthenticated grpc-code: 16 ; http-code: 401
// The request has not been applied because it lacks valid authentication credentials.
// @param context: relevant server context
// @param res: rpc method return type; always be an empty or nil value [using for the trade-off]
// args...
// @type errCode: custom four-digit [XXXX] series error code
// @type Err(s): custom err object to tell what exactly happened
func Unauthenticated[T any](ctx context.Context, res T, args ...any) (T, error) {
	return e(ctx, res, codes.Unauthenticated, msgUnauthorized, args...)
}
