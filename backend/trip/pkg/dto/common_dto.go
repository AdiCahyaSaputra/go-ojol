package dto

const (
	// Failed
	MESSAGE_ROLE_INVALID              = "invalid role for logon user"
	MESSAGE_PROFILE_CONTEXT_NOT_FOUND = "profile context not found, run this middleware after authentication middleware"
	MESSAGE_FAILED_PROSES_REQUEST     = "failed process request"
	MESSAGE_FAILED_TOKEN_NOT_FOUND    = "token not found"
	MESSAGE_FAILED_TOKEN_NOT_VALID    = "token not valid"
	MESSAGE_FAILED_DENIED_ACCESS      = "denied access"
	MESSAGE_FAILED_GET_DATA_FROM_BODY = "failed get data from body"
	MESSAGE_USER_ID_CLAIM_MISSING     = "can't find active session user id on the jwt claim"
	MESSAGE_INTERNAL_SERVER_ERROR     = "An error occurred on our server"

	// Success
	MESSAGE_SUCCESS_GET_DATA = "success get data"
)
