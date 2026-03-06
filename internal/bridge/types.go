package bridge

import "encoding/json"

type TransactionsListArgs struct {
	AccountID string `json:"accountId,omitempty"`
	From      string `json:"from"`
	To        string `json:"to"`
	Limit     int    `json:"limit"`
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

type TransactionsListResponse struct {
	Transactions []json.RawMessage `json:"transactions"`
}

type TransactionRow struct {
	Date      string      `json:"date"`
	Account   string      `json:"account"`
	PayeeName string      `json:"payee_name"`
	Amount    interface{} `json:"amount"`
	Notes     string      `json:"notes"`
}

type BudgetSummaryResponse struct {
	Month  string          `json:"month"`
	Budget json.RawMessage `json:"budget"`
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
