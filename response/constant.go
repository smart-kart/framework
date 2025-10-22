package response

// http status
const (
	statusSuccess  = "success"
	statusCreated  = "created"
	statusAccepted = "accepted"
	statusError    = "error"
)

// message
const (
	msgMethodNotAllowed = "Method Not Allowed : You tried to access with an invalid method. " +
		"The server knows the request method, but the target resource doesn't " +
		"support this method. eg: inputting an incorrect URL"
	msgServiceUnavailable  = "Service Unavailable : We're temporarily offline for maintenance. Please try again later"
	msgInternalServerError = "Internal Server Error : We had a problem with our server. Try again later"
	msgBadRequest          = "Bad Request : Your request is invalid"
	msgNotFound            = "Resource not found"
	msgConflict            = "Conflict with the current state of the resource"
	msgUnauthorized        = "Unauthorised : The Access Token or the API key or the Basic Auth entered is invalid. " +
		"Please also check if you are using the sandbox credentials on the production end point or vice versa"
	msgForbidden           = "Forbidden : You are not authorised to access the requested resource"
	msgUnprocessableEntity = "Unprocessable Entity : Please check the semantic erroneous"
	msgCanceled            = "Canceled: Client closed the request"
	msgTimeout             = "Gateway Timeout. Try again later"
	msgTooManyRequest      = "Too Many Request : " +
		"You have surpassed the rate limit or number of requests in a given time"
	msgNotImplemented = "Not Implemented : We're not support the functionality required to fulfil the request"
	msgUnavailable    = "Service Unavailable : Server is not ready to handle the request"
)

// grpc metadata [value should be in lower case]
const (
	mdHTTPStatusCode = "x-http-statuscode"
)

// http header key
const (
	headerTrailer          = "Trailer"
	headerTransferEncoding = "Transfer-Encoding"
	headerContentType      = "Content-Type"
	headerWWWAuthenticate  = "WWW-Authenticate"
	headerCacheControl     = "Cache-Control"
)

// http header value
const (
	ctApplicationJSON = "application/json"
	ccNoStore         = "no-store, max-age=0"
)

// verbose error code
const (
	ErrInvalidRequest ErrCode = 100 + iota
	ErrDBOperationFailed
	ErrSomethingWentWrong
	ErrInvalidPathParam
	ErrInvalidQueryParam
	ErrResourceNotFound
	ErrInvalidToken
	ErrTokenExpired
	ErrInvalidBasicAuth
	ErrEmptyBasicAuth
	ErrTooManyRequests
	ErrInvalidAPIKey
	ErrUnsupportedFileType
)

// response field
const (
	keyStrictResponse  = "strict_response"
	keyResponseMessage = "response_message"
)
