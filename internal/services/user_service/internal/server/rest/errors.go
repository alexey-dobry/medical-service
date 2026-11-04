package rest

import (
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// grpcErrorToHTTP преобразует ошибку gRPC auth_service в HTTP-статус, чтобы ошибки клиента (дубликат email, неверный ввод) не возвращались как 500.
func grpcErrorToHTTP(err error) int {
	switch status.Code(err) {
	case codes.AlreadyExists:
		return fiber.StatusConflict
	case codes.InvalidArgument:
		return fiber.StatusBadRequest
	case codes.NotFound:
		return fiber.StatusNotFound
	case codes.Unauthenticated:
		return fiber.StatusUnauthorized
	case codes.Unavailable, codes.DeadlineExceeded:
		return fiber.StatusServiceUnavailable
	default:
		return fiber.StatusInternalServerError
	}
}

func errorJSON(c *fiber.Ctx, code int, msg string) error {
	return c.Status(code).JSON(fiber.Map{"error": msg})
}
