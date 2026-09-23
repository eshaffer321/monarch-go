package monarch

import (
	"time"
)

// Account represents a financial account
type Account struct {
	ID                              string              `json:"id"`
	DisplayName                     string              `json:"displayName"`
	SyncDisabled                    bool                `json:"syncDisabled"`
	DeactivatedAt                   *time.Time          `json:"deactivatedAt,omitempty"`
	IsHidden                        bool                `json:"isHidden"`
	IsAsset                         bool                `json:"isAsset"`
	Mask                            string              `json:"mask"`
	CreatedAt                       time.Time           `json:"createdAt"`
	UpdatedAt                       time.Time           `json:"updatedAt"`
	DisplayLastUpdatedAt            time.Time           `json:"displayLastUpdatedAt"`
	CurrentBalance                  float64             `json:"currentBalance"`
	DisplayBalance                  float64             `json:"displayBalance"`
	IncludeInNetWorth               bool                `json:"includeInNetWorth"`
	HideFromList                    bool                `json:"hideFromList"`
	HideTransactionsFromReports     bool                `json:"hideTransactionsFromReports"`
	IncludeBalanceInNetWorth        bool                `json:"includeBalanceInNetWorth"`
	IncludeInGoalBalance            bool                `json:"includeInGoalBalance"`
	DataProvider                    string              `json:"dataProvider"`
	DataProviderAccountID           string              `json:"dataProviderAccountId"`
	IsManual                        bool                `json:"isManual"`
	TransactionsCount               int                 `json:"transactionsCount"`
	HoldingsCount                   int                 `json:"holdingsCount"`
	ManualInvestmentsTrackingMethod string              `json:"manualInvestmentsTrackingMethod"`
	Order                           int                 `json:"order"`
	LogoURL                         string              `json:"logoUrl"`
	Type                            *AccountTypeInfo    `json:"type"`
	Subtype                         *AccountSubtypeInfo `json:"subtype"`
	Credential                      *Credential         `json:"credential"`
	Institution                     *Institution        `json:"institution"`
}

// AccountTypeInfo represents account type information
type AccountTypeInfo struct {
	Name    string `json:"name"`
	Display string `json:"display"`
}

// AccountSubtypeInfo represents account subtype information
type AccountSubtypeInfo struct {
	Name    string `json:"name"`
	Display string `json:"display"`
}

// AccountType represents available account types
type AccountType struct {
	Type             *AccountTypeInfo      `json:"type"`
	Subtype          *AccountSubtypeInfo   `json:"subtype"`
	PossibleSubtypes []*AccountSubtypeInfo `json:"possibleSubtypes"`
}

// Credential represents account credentials
type Credential struct {
	ID                             string       `json:"id"`
	UpdateRequired                 bool         `json:"updateRequired"`
	DisconnectedFromDataProviderAt *time.Time   `json:"disconnectedFromDataProviderAt,omitempty"`
	DataProvider                   string       `json:"dataProvider"`
	Institution                    *Institution `json:"institution"`
}

// Institution represents a financial institution
type Institution struct {
	ID                 string `json:"id"`
	PlaidInstitutionID string `json:"plaidInstitutionId"`
	Name               string `json:"name"`
	Status             string `json:"status"`
	PrimaryColor       string `json:"primaryColor"`
	URL                string `json:"url"`
	LogoURL            string `json:"logoUrl"`
	// Credential fields
	CredentialID   string     `json:"credentialId"`
	UpdateRequired bool       `json:"updateRequired"`
	DisconnectedAt *time.Time `json:"disconnectedAt"`
	LastUpdatedAt  *time.Time `json:"lastUpdatedAt"`
	DataProvider   string     `json:"dataProvider"`
}

// Merchant represents a merchant
type Merchant struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Transaction represents a financial transaction
type Transaction struct {
	ID                 string               `json:"id"`
	Date               Date                 `json:"date"`
	Amount             float64              `json:"amount"`
	Pending            bool                 `json:"pending"`
	HideFromReports    bool                 `json:"hideFromReports"`
	PlaidName          string               `json:"plaidName"`
	Merchant           *Merchant            `json:"merchant"`
	Notes              string               `json:"notes"`
	HasSplits          bool                 `json:"hasSplits"`
	IsSplitTransaction bool                 `json:"isSplitTransaction"`
	IsRecurring        bool                 `json:"isRecurring"`
	NeedsReview        bool                 `json:"needsReview"`
	ReviewedAt         *Date                `json:"reviewedAt,omitempty"`
	ReviewedByUserID   string               `json:"reviewedByUserId"`
	CreatedAt          Date                 `json:"createdAt"`
	UpdatedAt          Date                 `json:"updatedAt"`
	Account            *Account             `json:"account"`
	Category           *TransactionCategory `json:"category"`
	Tags               []*Tag               `json:"tags"`
}

// TransactionDetails includes additional transaction information
type TransactionDetails struct {
	*Transaction
	OriginalMerchant    string               `json:"originalMerchant"`
	OriginalCategoryID  string               `json:"originalCategoryId"`
	OriginalCategory    *TransactionCategory `json:"originalCategory"`
	Splits              []*TransactionSplit  `json:"splits"`
	SimilarTransactions []*Transaction       `json:"similarTransactions"`
}

// TransactionSplit represents a split transaction
type TransactionSplit struct {
	ID         string               `json:"id"`
	Amount     float64              `json:"amount"`
	Merchant   *Merchant            `json:"merchant"`
	Notes      string               `json:"notes"`
	Category   *TransactionCategory `json:"category"`
	CategoryID string               `json:"categoryId"`
}

// TransactionCategory represents a transaction category
type TransactionCategory struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	Icon             string         `json:"icon"`
	Order            int            `json:"order"`
	SystemCategory   string         `json:"systemCategory"`
	IsSystemCategory bool           `json:"isSystemCategory"`
	IsDisabled       bool           `json:"isDisabled"`
	Group            *CategoryGroup `json:"group"`
	GroupID          string         `json:"groupId"`
}

// CategoryGroup represents a category group
type CategoryGroup struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	Order int    `json:"order"`
}

// Tag represents a transaction tag
type Tag struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Color            string `json:"color"`
	Order            int    `json:"order"`
	TransactionCount int    `json:"transactionCount"`
}

// Budget represents a budget entry
type Budget struct {
	ID                      string               `json:"id"`
	CategoryID              string               `json:"categoryId"`
	Category                *TransactionCategory `json:"category"`
	Amount                  float64              `json:"amount"`
	Rollover                bool                 `json:"rollover"`
	RolloverType            string               `json:"rolloverType"`
	RolloverAmount          float64              `json:"rolloverAmount"`
	StartDate               time.Time            `json:"startDate"`
	EndDate                 time.Time            `json:"endDate"`
	Spent                   float64              `json:"spent"`
	Remaining               float64              `json:"remaining"`
	PercentageComplete      float64              `json:"percentageComplete"`
}

// Goal represents a financial goal (goalsV2)
type Goal struct {
	ID                      string               `json:"id"`
	Name                    string               `json:"name"`
	Type                    string               `json:"type"`
	Amount                  float64              `json:"amount"`
	Priority                int                  `json:"priority"`
	TargetDate              *time.Time           `json:"targetDate,omitempty"`
	TargetAmount            float64              `json:"targetAmount"`
	CurrentAmount           float64              `json:"currentAmount"`
	ImageURL                *string              `json:"imageUrl,omitempty"`
	AccountID               *string              `json:"accountId,omitempty"`
	Account                 *Account             `json:"account,omitempty"`
	PercentageComplete      float64              `json:"percentageComplete"`
	MonthlyContribution     float64              `json:"monthlyContribution"`
	CreatedAt               time.Time            `json:"createdAt"`
	UpdatedAt               time.Time            `json:"updatedAt"`
}

// BudgetWithGoals extends Budget to include associated goals
type BudgetWithGoals struct {
	*Budget
	Goals []*Goal `json:"goals"`
}

// Cashflow represents comprehensive cashflow data
type Cashflow struct {
	StartDate       time.Time                `json:"startDate"`
	EndDate         time.Time                `json:"endDate"`
	Summary         *CashflowSummary         `json:"summary"`
	ByCategory      []*CashflowCategory      `json:"byCategory"`
	ByCategoryGroup []*CashflowCategoryGroup `json:"byCategoryGroup"`
	ByMerchant      []*CashflowMerchant      `json:"byMerchant"`
}

// CashflowCategory represents cashflow by category
type CashflowCategory struct {
	Category *TransactionCategory `json:"category"`
	Amount   float64              `json:"amount"`
}

// CashflowCategoryGroup represents cashflow by category group
type CashflowCategoryGroup struct {
	CategoryGroup *CategoryGroup `json:"categoryGroup"`
	Amount        float64        `json:"amount"`
}

// CashflowMerchant represents cashflow by merchant
type CashflowMerchant struct {
	Merchant *Merchant `json:"merchant"`
	Amount   float64   `json:"amount"`
}

// CashflowSummary represents cashflow summary
type CashflowSummary struct {
	StartDate   time.Time `json:"startDate,omitempty"`
	EndDate     time.Time `json:"endDate,omitempty"`
	Income      float64   `json:"income"`
	Expense     float64   `json:"expense"`
	Savings     float64   `json:"savings"`
	SavingsRate float64   `json:"savingsRate"`
}

// Holding represents an investment holding
type Holding struct {
	ID        string    `json:"id"`
	AccountID string    `json:"accountId"`
	Symbol    string    `json:"symbol"`
	Name      string    `json:"name"`
	Quantity  float64   `json:"quantity"`
	Price     float64   `json:"price"`
	Value     float64   `json:"value"`
	CostBasis float64   `json:"costBasis"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// AccountBalance represents account balance at a point in time
type AccountBalance struct {
	AccountID string  `json:"accountId"`
	Date      Date    `json:"date"`
	Balance   float64 `json:"balance"`
}

// AccountSnapshot represents account snapshot
type AccountSnapshot struct {
	Month        string  `json:"month"`
	Year         int     `json:"year"`
	Type         string  `json:"type"`
	Subtype      string  `json:"subtype"`
	TotalValue   float64 `json:"totalValue"`
	AccountCount int     `json:"accountCount"`
}

// AccountHistory represents account balance history
type AccountHistory struct {
	AccountID string          `json:"accountId"`
	Balances  []*BalanceEntry `json:"balances"`
}

// BalanceEntry represents a balance at a point in time
type BalanceEntry struct {
	Date    Date    `json:"date"`
	Balance float64 `json:"balance"`
	Synced  bool    `json:"synced"`
}

// RecurringTransaction represents a recurring transaction
type RecurringTransaction struct {
	ID            string               `json:"id"`
	Merchant      *Merchant            `json:"merchant"`
	Amount        float64              `json:"amount"`
	Frequency     string               `json:"frequency"`
	NextDate      Date                 `json:"nextDate"`
	LastDate      Date                 `json:"lastDate"`
	Category      *TransactionCategory `json:"category"`
	Account       *Account             `json:"account"`
	IsActive      bool                 `json:"isActive"`
	IsApproximate bool                 `json:"isApproximate"`
}

// Subscription represents subscription details
type Subscription struct {
	ID          string     `json:"id"`
	PlanType    string     `json:"planType"`
	Status      string     `json:"status"`
	StartDate   time.Time  `json:"startDate"`
	EndDate     *time.Time `json:"endDate,omitempty"`
	TrialEndsAt *time.Time `json:"trialEndsAt,omitempty"`
	CanceledAt  *time.Time `json:"canceledAt,omitempty"`
	Features    []string   `json:"features"`
}

// TransactionSummary represents transaction summary
type TransactionSummary struct {
	Avg        float64 `json:"avg"`
	Count      int     `json:"count"`
	Max        float64 `json:"max"`
	MaxExpense float64 `json:"maxExpense"`
	Sum        float64 `json:"sum"`
	SumIncome  float64 `json:"sumIncome"`
	SumExpense float64 `json:"sumExpense"`
	First      string  `json:"first"`
	Last       string  `json:"last"`
}

// TransactionList represents paginated transaction results
type TransactionList struct {
	Transactions []*Transaction `json:"transactions"`
	TotalCount   int            `json:"totalCount"`
	HasMore      bool           `json:"hasMore"`
	NextOffset   int            `json:"nextOffset"`
}

// Session represents an authenticated session
type Session struct {
	Token      string    `json:"token"`
	Cookie     string    `json:"cookie,omitempty"`
	UserID     string    `json:"userId"`
	Email      string    `json:"email"`
	ExpiresAt  time.Time `json:"expiresAt"`
	DeviceUUID string    `json:"deviceUuid"`
}

// Parameter structures

// CreateAccountParams for creating manual accounts
type CreateAccountParams struct {
	AccountType       string  `json:"accountType"`
	AccountSubtype    string  `json:"accountSubtype"`
	IsAsset           bool    `json:"isAsset"`
	AccountName       string  `json:"accountName"`
	CurrentBalance    float64 `json:"currentBalance"`
	IncludeInNetWorth bool    `json:"includeInNetWorth"`
}

// CreateInvestmentsAccountParams for creating a manual investments account with holdings.
type CreateInvestmentsAccountParams struct {
	Name                          string            `json:"name"`
	Subtype                       string            `json:"subtype"`
	ManualInvestmentsTrackingMethod string          `json:"manualInvestmentsTrackingMethod"`
	InitialHoldings               []InitialHolding  `json:"initialHoldings"`
}

// InitialHolding represents a holding to include when creating an investments account.
type InitialHolding struct {
	SecurityID string  `json:"securityId"`
	Quantity   float64 `json:"quantity"`
}

// UpdateAccountParams for updating accounts
type UpdateAccountParams struct {
	DisplayName                 *string  `json:"displayName,omitempty"`
	IncludeInNetWorth           *bool    `json:"includeInNetWorth,omitempty"`
	HideFromList                *bool    `json:"hideFromList,omitempty"`
	HideTransactionsFromReports *bool    `json:"hideTransactionsFromReports,omitempty"`
	CurrentBalance              *float64 `json:"currentBalance,omitempty"`
}

// CreateTransactionParams for creating transactions
type CreateTransactionParams struct {
	Date                Date      `json:"date"`
	AccountID           string    `json:"accountId"`
	Amount              float64   `json:"amount"`
	Merchant            *Merchant `json:"merchant"`
	CategoryID          string    `json:"categoryId"`
	Notes               string    `json:"notes"`
	ShouldUpdateBalance *bool     `json:"shouldUpdateBalance,omitempty"`
}

// UpdateTransactionParams for updating transactions
type UpdateTransactionParams struct {
	Date            *Date    `json:"date,omitempty"`
	AccountID       *string  `json:"accountId,omitempty"`
	Amount          *float64 `json:"amount,omitempty"`
	Merchant        *string  `json:"merchant,omitempty"`
	CategoryID      *string  `json:"categoryId,omitempty"`
	Notes           *string  `json:"notes,omitempty"`
	HideFromReports *bool    `json:"hideFromReports,omitempty"`
	NeedsReview     *bool    `json:"needsReview,omitempty"`
}

// CreateCategoryParams for creating categories
type CreateCategoryParams struct {
	Name               string    `json:"name"`
	GroupID            string    `json:"groupId"`
	Icon               string    `json:"icon"`
	RolloverEnabled    bool      `json:"rolloverEnabled"`
	RolloverType       string    `json:"rolloverType"`        // "monthly", etc.
	RolloverStartMonth time.Time `json:"rolloverStartMonth"`
}

// SnapshotParams for account snapshots
type SnapshotParams struct {
	StartDate    time.Time `json:"startDate"`
	Timeframe    string    `json:"timeframe"` // "year" or "month"
	AccountTypes []string  `json:"accountTypes,omitempty"`
	GroupBy      string    `json:"groupBy,omitempty"`
}

// CashflowParams for cashflow queries
type CashflowParams struct {
	StartDate  time.Time `json:"startDate"`
	EndDate    time.Time `json:"endDate"`
	Limit      int       `json:"limit,omitempty"`
	AccountIDs []string  `json:"accountIds,omitempty"`
}

// CashflowSummaryParams for cashflow summary
type CashflowSummaryParams struct {
	StartDate      time.Time `json:"startDate"`
	EndDate        time.Time `json:"endDate"`
	Interval       string    `json:"interval"` // "day", "week", "month", "year"
	CategoryID     string    `json:"categoryId,omitempty"`
	AccountsFilter []string  `json:"accountsFilter,omitempty"`
}

// SubscriptionDetails represents subscription information
type SubscriptionDetails struct {
	ID                    string `json:"id"`
	PaymentSource         string `json:"paymentSource"`
	ReferralCode          string `json:"referralCode"`
	IsOnFreeTrial         bool   `json:"isOnFreeTrial"`
	HasPremiumEntitlement bool   `json:"hasPremiumEntitlement"`
}

// AggregateSnapshot represents a daily balance snapshot
type AggregateSnapshot struct {
	Date    string  `json:"date"`
	Balance float64 `json:"balance"`
}

// AggregateSnapshotsParams for aggregate snapshots query
type AggregateSnapshotsParams struct {
	StartDate   *time.Time `json:"startDate,omitempty"`
	EndDate     *time.Time `json:"endDate,omitempty"`
	AccountType string     `json:"accountType,omitempty"`
}
