package bridge

import "encoding/json"

type TransactionsListArgs struct {
	AccountID            string `json:"accountId,omitempty"`
	From                 string `json:"from"`
	To                   string `json:"to"`
	Limit                int    `json:"limit"`
	IncludeCategoryNames bool   `json:"includeCategoryNames,omitempty"`
}

type BudgetCategoriesArgs struct {
	Month string `json:"month"`
}

type Account struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	OffBudget bool   `json:"offbudget"`
	Closed    bool   `json:"closed"`
}

type AccountsListResponse struct {
	Accounts []json.RawMessage `json:"accounts"`
}

type AccountRow struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	OffBudget bool   `json:"offbudget"`
	Closed    bool   `json:"closed"`
}

type CategoriesListResponse struct {
	Categories []json.RawMessage `json:"categories"`
}

type CategoryRow struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	GroupID   string `json:"group_id"`
	GroupName string `json:"group_name"`
	Hidden    bool   `json:"hidden"`
	Archived  *bool  `json:"archived,omitempty"`
}

type TransactionsListResponse struct {
	Transactions []json.RawMessage `json:"transactions"`
}

type TransactionRow struct {
	Date              string      `json:"date"`
	Account           string      `json:"account"`
	PayeeName         string      `json:"payee_name"`
	Amount            interface{} `json:"amount"`
	Notes             string      `json:"notes"`
	CategoryName      string      `json:"category_name,omitempty"`
	CategoryGroupName string      `json:"category_group_name,omitempty"`
}

type BudgetSummaryResponse struct {
	Month  string          `json:"month"`
	Budget json.RawMessage `json:"budget"`
}

type BudgetCategoriesResponse struct {
	Month      string            `json:"month"`
	Categories []json.RawMessage `json:"categories"`
}

type BudgetSummaryRow struct {
	Income   interface{} `json:"income"`
	Budgeted interface{} `json:"budgeted"`
	Spent    interface{} `json:"spent"`
}

type AuthCheckResponse struct {
	OK      bool            `json:"ok"`
	Budgets json.RawMessage `json:"budgets"`
}
