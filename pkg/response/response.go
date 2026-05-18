package response

type Response struct {
	Success bool         `json:"success"`
	Message string       `json:"message,omitempty"`
	Data    any          `json:"data,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
	Meta    *Meta        `json:"meta,omitempty"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func OK(data any) Response {
	return Response{Success: true, Data: data}
}

func OKWithMeta(data any, meta Meta) Response {
	return Response{Success: true, Data: data, Meta: &meta}
}

func OKWithMessage(data any, message string) Response {
	return Response{Success: true, Message: message, Data: data}
}

func Created(data any) Response {
	return Response{Success: true, Message: "created", Data: data}
}

func Deleted() Response {
	return Response{Success: true, Message: "deleted"}
}

func Health() Response {
	return Response{Success: true, Message: "ok"}
}

func Error(code, message string) Response {
	return Response{
		Success: false,
		Error:   &ErrorDetail{Code: code, Message: message},
	}
}

func ErrorWithDetails(code, message, details string) Response {
	return Response{
		Success: false,
		Error:   &ErrorDetail{Code: code, Message: message, Details: details},
	}
}
