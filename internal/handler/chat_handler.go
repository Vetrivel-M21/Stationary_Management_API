package handler

import (
	"strconv"
	"stationery-management/internal/domain"
	"stationery-management/internal/service"
	"stationery-management/pkg/response"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatSvc service.ChatService
}

func NewChatHandler(chatSvc service.ChatService) *ChatHandler {
	return &ChatHandler{chatSvc: chatSvc}
}

func (h *ChatHandler) GetChatMessages(c *gin.Context) {
	reqIDStr := c.Param("id")
	reqID, err := strconv.ParseUint(reqIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid request ID")
		return
	}

	messages, err := h.chatSvc.GetChatMessages(uint(reqID))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.JSONSuccess(c, 200, "Chat messages retrieved", messages)
}

func (h *ChatHandler) SendChatMessage(c *gin.Context) {
	reqIDStr := c.Param("id")
	reqID, err := strconv.ParseUint(reqIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid request ID")
		return
	}

	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "User unauthenticated")
		return
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		response.Unauthorized(c, "Invalid user token session")
		return
	}

	var dto domain.SendChatMessageDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}

	msg, err := h.chatSvc.SendChatMessage(uint(reqID), userID, dto.Message)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.JSONSuccess(c, 201, "Message sent successfully", msg)
}
