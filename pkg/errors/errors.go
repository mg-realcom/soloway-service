package errors

const (
	ErrMsgFailedCastRequest       = "failed to cast request"
	ErrMsgFailedInitBigQuery      = "can't init client BigQuery"
	ErrMsgFailedInitCloudStorage  = "can't init client CloudStorage"
	ErrMsgFailedInitGoogleSheets  = "can't init client GoogleSheets"
	ErrMsgFailedCloseBigQuery     = "can't close connection BigQuery"
	ErrMsgFailedCloseCloudStorage = "can't close connection CloudStorage"
)
