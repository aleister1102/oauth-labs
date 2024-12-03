package dto

type VisitRequest struct {
	URL       string `form:"url" binding:"required,http_url"`
	LabNumber int    `form:"lab" binding:"required,number,gt=0,lt=100"`
}
