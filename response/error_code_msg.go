//nolint:gochecknoglobals // expected to be at global level
package response

// ErrMsg error code - message [General]
var ErrMsg = map[ErrCode]string{
	ErrInvalidRequest: "Failure on decoding the json request. " +
		"The format for Json is not correct, please check the syntax & case sensitivity " +
		"for any missing brackets, semi-colons, capitalizations, or any other formatting errors.",
	ErrDBOperationFailed:  "Something went wrong with this API",
	ErrSomethingWentWrong: "Something went wrong with this API",
	ErrInvalidPathParam:   "Invalid parameter in the URL path.\nExample : /xyz/my_value/abc Here my_value is invalid",
	ErrInvalidQueryParam:  "Invalid parameter in the URL path.\nExample : /xyz?sort=my_value Here my_value is invalid",
	ErrResourceNotFound: "The resource you are trying to retrieve is not present. " +
		"This error code happens when trying to Get or Read a data " +
		"point (eg: transaction, user, KYB, etc) that does not exist.",
	ErrInvalidToken:     "Invalid access token. Please check your authentication flow and try again.",
	ErrTokenExpired:     "Access token has expired. Please generate a new token to continue.",
	ErrInvalidBasicAuth: "Invalid basic auth. Please check your authentication flow and try again.",
	ErrEmptyBasicAuth:   "Empty basic auth. Please check your authentication flow and try again.",
	ErrTooManyRequests: "Too Many Requests: Exceeded the allowable rate of requests." +
		" Please wait and try again later.",
	ErrInvalidAPIKey:       "Invalid api key. Please check your authentication flow and try again.",
	ErrUnsupportedFileType: "Unsupported file type. Please check the file type and try again.",
}
