package rest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (s *RESTServer) handleGetRecords() fiber.Handler {
	return func(c *fiber.Ctx) error {
		patientIDParam := c.Params("patient_id")
		patientID, err := uuid.Parse(patientIDParam)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid doctor_id",
			})
		}

	}
}

func (s *RESTServer) handleGetRecord() fiber.Handler {
	return func(c *fiber.Ctx) error {

	}
}

func (s *RESTServer) handleGetFile() fiber.Handler {
	return func(c *fiber.Ctx) error {

	}
}

func (s *RESTServer) handleCreateRecord() fiber.Handler {
	return func(c *fiber.Ctx) error {

	}
}
