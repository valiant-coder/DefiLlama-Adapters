package data

type ListParamInterface interface {
	GetOrder() string
	GetLimit() int
	Offset() int
	
	GetCountColumn() string
	CloseCounter() bool
	
	GetSummary() bool
	
	IsOnlyCount() bool
}

type ListParam struct {
	Current     int    `json:"current" validate:"required,gt=0" url:"current" ignore:"true"`            // 页面
	PageSize    int    `json:"pageSize" validate:"required,lte=1000,gt=0" url:"pageSize" ignore:"true"` // 页码
	Order       string `json:"order" url:"order" ignore:"true"`
	Keyword     string `json:"keyword" url:"keyword" ignore:"true"`
	CloseCount  bool   `json:"close_count" url:"close_count" ignore:"true"`
	SummaryInfo bool   `json:"summary_info" url:"summary_info" ignore:"true"`
	OnlyCount   bool   `json:"only_count" url:"only_count" ignore:"true"`
	CountColumn string `json:"count_column" url:"count_column" ignore:"true"`
}

func (p ListParam) Offset() int {
	
	return (p.Current - 1) * p.PageSize
}

func (p ListParam) GetPage() int {
	
	return p.Current
}

func (p ListParam) GetLimit() int {
	
	return p.PageSize
}

func (p ListParam) GetOrder() string {
	
	return p.Order
}

func (p ListParam) CloseCounter() bool {
	
	return p.CloseCount
}

func (p ListParam) GetSummary() bool {
	
	return p.SummaryInfo
}

func (p ListParam) IsOnlyCount() bool {
	
	return p.OnlyCount
}
func (p ListParam) GetCountColumn() string {
	
	// count(id)
	return p.CountColumn
}
