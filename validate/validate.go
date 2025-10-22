package validate

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/smart-kart/framework/response"
	protov1 "github.com/smart-kart/proto/gen/go/proto/v1"
)

// global object
//
//nolint:gochecknoglobals // validate is thread safe to create singleton object
var (
	_v *validator.Validate
)

// creates a new instance of 'validate' with JSON tag support.
//
//nolint:gochecknoinits // since we have common `Request` func, `_v` should initialize at first place
func init() {
	// Validate is designed to be thread-safe and used as a singleton instance.
	// create new instance of 'validate' with sane defaults.
	_v = validator.New()

	// register a function to get json names for StructFields.
	_v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}

		return name
	})
}

// format parse all ValidationErrors and prepare detailed Err object
func format(errType response.ErrType, verr validator.ValidationErrors) []*protov1.Err {
	errs := make([]*protov1.Err, 0)

	for _, f := range verr {
		code := response.GetValidationErrCode(errType, f.Field())
		err := &protov1.Err{
			Code:    int32(code),
			Message: response.GetErrMsg(code),
			// exact field in tree view (i.e root.element.key)
			Remarks: f.Namespace()[strings.Index(f.Namespace(), ".")+1:],
		}
		errs = append(errs, err)
	}

	return errs
}

// Request validate `struct` fields with validator pkg
// @param ctx: context
// @param req: struct to be validated
// @param errType: validation ErrType to pick respective field's ErrCode
//
// @return error: gRPC error response if validation failed
func Request(ctx context.Context, req any, errType response.ErrType) error {
	err := _v.StructCtx(ctx, req)
	if err != nil {
		// check for any validator errors
		vErr := validator.ValidationErrors{}
		if ok := errors.As(err, &vErr); ok {
			validationErrors := format(errType, vErr)
			_, err = response.InvalidArgument(ctx, response.Empty, validationErrors)

			return err
		}

		// fallback to internal error
		_, err = response.InternalError(ctx, response.Empty, response.ErrSomethingWentWrong)

		return err
	}

	return nil
}

// RegisterCustomValidators registers custom validation functions
func RegisterCustomValidators(customValidator map[string]validator.Func) error {
	// register the custom validators provided
	for key, val := range customValidator {
		err := _v.RegisterValidation(key, val)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetValidator returns the validator instance for custom usage
func GetValidator() *validator.Validate {
	return _v
}
