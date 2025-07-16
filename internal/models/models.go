package models

import (
	"time"
)

const (
	APIEndPoint = "http://localhost:8080/"
	APPName     = "Photostock"
	APPVersion  = "1.0"
	DBHost      = "localhost"
	DBPort      = "5432"
	DBName      = ""
	DBUser      = ""
	DBPassword  = "QZxWHNawLaYUXJKQTMzlP2R7t9T8iMI5"
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
	TermsList     []string  `json:"terms"`
	Terms         string    `json:"concat_terms"` // stored in DB
	Status        bool      `json:"status"`
	Price         int       `json:"price"`
	DownloadLimit int       `json:"download_limit"`
	ExpiresAt     int       `json:"expires_at"` // stored as interval in DB
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}


type MediaCategory struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	ThumbnailURL string    `json:"thumbnail_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type User struct {
	ID             int           `json:"id"`
	Username       string        `json:"username"`
	Password       string        `json:"password,omitempty"`
	Name           string        `json:"name"`
	AvatarURL      string        `json:"avatar_url"`
	AvatarID       string        `json:"avatar_id"`
	Status         bool          `json:"status"`
	Role           string        `json:"role"`
	Email          string        `json:"email"`
	Mobile         string        `json:"mobile"`
	TotalEarnings  float64       `json:"total_earnings"`
	TotalWithdraw  float64       `json:"total_withdraw"`
	TotalExpenses  float64       `json:"total_expenses"`
	Address        string        `json:"address"`
	SubscriptionID *int          `json:"subscription_id"` // nullable FK
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	CurrentPlan    *Subscription `json:"current_plan"`
}

type Subscription struct {
	ID                 int               `json:"id"`
	UserID             int               `json:"user_id"`
	SubscriptionPlanID int               `json:"subscription_plan_id"`
	PlanDetails        *SubscriptionPlan `json:"plan_details"`
	PaymentStatus      string            `json:"payment_status"`
	PaymentAmount      float64           `json:"payment_amount"`
	PaymentTime        time.Time         `json:"payment_time"`
	TotalDownloads     int               `json:"total_downloads"`
	Status             bool              `json:"status"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
}

type Media struct {
	ID            int           `json:"id"`
	MediaUUID     string        `json:"media_uuid"`
	MediaURL      string        `json:"media_url"`
	MediaTitle    string        `json:"media_title"`
	Description   string        `json:"description"`
	CategoryID    int           `json:"category_id"` // foreign key of media_categories
	TotalEarnings float64       `json:"total_earnings"`
	LicenseType   int           `json:"license_type"` //premium = 0, free = 1
	MediaCategory MediaCategory `json:"media_category"`
	UploaderID    int           `json:"uploader_id"` //foreign key of users table
	UploaderName  string        `json:"uploader_name"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type UploadHistory struct {
	ID         int       `json:"id"`
	MediaUUID  string    `json:"media_id"`
	UserID     int       `json:"user_id"` //uploader
	FileType string `json:"file_type"`
	FileExt string `json:"file_ext"`
	FileName string `json:"file_name"`
	FileSize string `json:"file_size"`
	Resolution string `json:"resolution"`
	UploadedAt time.Time `json:"uploaded_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
type DownloadHistory struct {
	ID         int       `json:"id"`
	MediaUUID  string    `json:"media_id"`
	UserID     int       `json:"user_id"` //uploader
	FileType string `json:"file_type"`
	FileExt string `json:"file_ext"`
	FileName string `json:"file_name"`
	FileSize string `json:"file_size"`
	Resolution string `json:"resolution"`
	DownloadedAt time.Time `json:"downloaded_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}