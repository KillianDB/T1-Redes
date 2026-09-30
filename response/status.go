package response

type Status struct {
	Code    int
	Message string
}

var (
	OK = Status{
		Code:    200,
		Message: "OK",
	}

	BadRequest = Status{
		Code:    400,
		Message: "Bad Request",
	}

	Unauthorized = Status{
		Code:    401,
		Message: "Unauthorized",
	}

	Forbidden = Status{
		Code:    403,
		Message: "Forbidden",
	}

	NotFound = Status{
		Code:    404,
		Message: "Not Found",
	}

	InternalServerError = Status{
		Code:    500,
		Message: "Internal Server Error",
	}
)
