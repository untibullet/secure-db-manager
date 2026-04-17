package domain

// Filter — карта фильтров для List-методов репозитория.
// Ключи — хардкодированные имена колонок (AD-5), никогда не пользовательский ввод.
type Filter map[string]any

// Paging задаёт параметры пагинации.
type Paging struct {
	Limit  int
	Offset int
}

// PageMeta содержит метаданные страницы для ответов API.
type PageMeta struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}
