package models

import (
	"time"
)

const (
	APPName    = "Photostock"
	APPVersion = "1.0"
	DBHost     = "localhost"
	DBPort     = "5432"
	DBName     = ""
	DBUser     = ""
	DBPassword = "QZxWHNawLaYUXJKQTMzlP2R7t9T8iMI5"
)

var Passphrase = "jM/0qr%HKU&!G%MdivH#A-{oInY*Nv20"

// Response is the type for response
type Response struct {
	Error   bool   `json:"error"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// User holds the user info
type JWT struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	Issuer    string    `json:"iss"`
	Audience  string    `json:"aud"`
	ExpiresAt int64     `json:"exp"`
	IssuedAt  int64     `json:"iat"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SubscriptionPlan struct {
	ID            int       `json:"id"`
	Title         string    `json:"title"`
	Terms         string    `json:"terms"`
	Status        bool      `json:"status"`
	DownloadLimit int       `json:"download_limit"`
	TimeLimit     string    `json:"time_limit"` // Representing INTERVAL as string
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type MediaCategory struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	ID               int               `json:"id"`
	Username         string            `json:"username"`
	Password         string            `json:"password,omitempty"`
	Name             string            `json:"name"`
	AvatarURL        string            `json:"avatar_url"`
	Status           bool              `json:"status"`
	Role             string            `json:"role"`
	Email            string            `json:"email"`
	Mobile           string            `json:"mobile"`
	TotalEarnings    float64           `json:"total_earnings"`
	TotalWithdraw    float64           `json:"total_withdraw"`
	TotalExpenses    float64           `json:"total_expenses"`
	Address          string            `json:"address"`
	SubscriptionID   *int              `json:"subscription_id"` // nullable FK
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
	SubscriptionPlan *SubscriptionPlan `json:"subscription_plan"`
}

type Subscription struct {
	ID                 int       `json:"id"`
	UserID             int       `json:"user_id"`
	SubscriptionPlanID int       `json:"subscription_plan_id"`
	PaymentStatus      string    `json:"payment_status"`
	PaymentAmount      float64   `json:"payment_amount"`
	PaymentTime        time.Time `json:"payment_time"`
	Status             bool      `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type Media struct {
	ID            int           `json:"id"`
	MediaUUID     string        `json:"media_uuid"`
	MediaTitle    string        `json:"media_title"`
	Description   string        `json:"description"`
	CategoryID    *int          `json:"category_id"` // nullable FK
	TotalEarnings float64       `json:"total_earnings"`
	LicenseType   int           `json:"license_type"` //premium = 0, free = 1
	MediaCategory MediaCategory `json:"media_category"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type UploadHistory struct {
	ID           int       `json:"id"`
	MediaUUID    string    `json:"media_id"`
	UserID       int       `json:"user_id"` //uploader
	UploadedAt time.Time `json:"uploaded_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
type DownloadHistory struct {
	ID           int       `json:"id"`
	MediaUUID    string    `json:"media_id"`
	UserID       int       `json:"user_id"` //downloader
	DownloadedAt time.Time `json:"downloaded_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
