package handlers

import (
	"scripts-management/internal/models"
	"scripts-management/internal/services"
	"scripts-management/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProcessHandler struct {
	processService *services.ProcessService
}

func NewProcessHandler(processService *services.ProcessService) *ProcessHandler {
	return &ProcessHandler{
		processService: processService,
	}
}

func (h *ProcessHandler) RunScript(c *fiber.Ctx) error {
	scriptID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid script ID",
		})
	}

	var req models.RunScriptRequest
	if err := c.BodyParser(&req); err != nil {
		// Nếu không có body hoặc parse lỗi, sử dụng args rỗng
		req.Args = []string{}
	}

	user := c.Locals("user").(*utils.JWTClaims)
	userID, err := primitive.ObjectIDFromHex(user.UserID.Hex())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	return h.processService.RunScript(c, userID, scriptID, req.Args)
}

func (h *ProcessHandler) StopProcess(c *fiber.Ctx) error {
	processID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid process ID",
		})
	}

	user := c.Locals("user").(*utils.JWTClaims)
	userID, err := primitive.ObjectIDFromHex(user.UserID.Hex())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	if err := h.processService.StopProcess(c.Context(), userID, processID); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Process stopped successfully",
	})
}

func (h *ProcessHandler) GetProcesses(c *fiber.Ctx) error {
	user := c.Locals("user").(*utils.JWTClaims)
	userID, err := primitive.ObjectIDFromHex(user.UserID.Hex())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	processes, err := h.processService.GetProcesses(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(processes)
}
