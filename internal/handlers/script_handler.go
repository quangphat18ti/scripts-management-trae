package handlers

import (
	"scripts-management/internal/models"
	"scripts-management/internal/services"
	"scripts-management/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ScriptHandler struct {
	scriptService *services.ScriptService
}

func NewScriptHandler(scriptService *services.ScriptService) *ScriptHandler {
	return &ScriptHandler{
		scriptService: scriptService,
	}
}

func (h *ScriptHandler) CreateScript(c *fiber.Ctx) error {
	var req models.CreateScriptRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	user := c.Locals("user").(*utils.JWTClaims)
	userID, err := primitive.ObjectIDFromHex(user.UserID.Hex())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	script, err := h.scriptService.CreateScript(c.Context(), userID, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(script)
}

func (h *ScriptHandler) GetScript(c *fiber.Ctx) error {
	scriptID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid script ID",
		})
	}

	user := c.Locals("user").(*utils.JWTClaims)
	userID, err := primitive.ObjectIDFromHex(user.UserID.Hex())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	script, err := h.scriptService.GetScriptByID(c.Context(), userID, scriptID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(script)
}

func (h *ScriptHandler) GetUserScripts(c *fiber.Ctx) error {
	user := c.Locals("user").(*utils.JWTClaims)
	userID, err := primitive.ObjectIDFromHex(user.UserID.Hex())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	scripts, err := h.scriptService.GetUserScripts(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(scripts)
}

func (h *ScriptHandler) UpdateScript(c *fiber.Ctx) error {
	scriptID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid script ID",
		})
	}

	var req models.UpdateScriptRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	user := c.Locals("user").(*utils.JWTClaims)
	userID, err := primitive.ObjectIDFromHex(user.UserID.Hex())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	script, err := h.scriptService.UpdateScript(c.Context(), userID, scriptID, &req)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(script)
}

func (h *ScriptHandler) DeleteScript(c *fiber.Ctx) error {
	scriptID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid script ID",
		})
	}

	user := c.Locals("user").(*utils.JWTClaims)
	userID, err := primitive.ObjectIDFromHex(user.UserID.Hex())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	if err := h.scriptService.DeleteScript(c.Context(), userID, scriptID); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Script deleted successfully",
	})
}

func (h *ScriptHandler) ShareScript(c *fiber.Ctx) error {
	scriptID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid script ID",
		})
	}

	var req models.ShareScriptRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	user := c.Locals("user").(*utils.JWTClaims)
	userID, err := primitive.ObjectIDFromHex(user.UserID.Hex())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	if err := h.scriptService.ShareScript(c.Context(), userID, scriptID, req.UserID); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Script shared successfully",
	})
}

func (h *ScriptHandler) RevokeShare(c *fiber.Ctx) error {
	scriptID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid script ID",
		})
	}

	targetUserID := c.Params("userId")
	if targetUserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	user := c.Locals("user").(*utils.JWTClaims)
	userID, err := primitive.ObjectIDFromHex(user.UserID.Hex())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	if err := h.scriptService.RevokeShare(c.Context(), userID, scriptID, targetUserID); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Share revoked successfully",
	})
}
