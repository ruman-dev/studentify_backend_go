package notifications

type CreateRequest struct {
	Title         string `json:"title" validate:"required"`
	Description   string `json:"description"`
	Type          string `json:"type" validate:"omitempty,oneof=info warning alert reminder"`
	ReferenceType string `json:"reference_type"`
	ReferenceID   string `json:"reference_id"`
}

type Response struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	Type          string  `json:"type"`
	IsRead        bool    `json:"isRead"`
	ReadAt        *string `json:"readAt,omitempty"`
	ReferenceType string  `json:"referenceType"`
	ReferenceID   string  `json:"referenceId"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     string  `json:"updatedAt"`
}

type ListResponse struct {
	Items       []Response `json:"items"`
	UnreadCount int        `json:"unreadCount"`
}
