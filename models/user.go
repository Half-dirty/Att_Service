package models

import (
	"time"

	"gorm.io/gorm"
)

// User представляет пользователя системы.
type User struct {
	gorm.Model

	JestID      string `gorm:"unique;size:20;column:jest_id"`
	StoragePath string `gorm:"size:255;column:storage_path"`

	// Имя (в трёх падежах)
	NameInIp string `gorm:"size:100;column:name_in_ip"`
	NameInRp string `gorm:"size:100;column:name_in_rp"`
	NameInDp string `gorm:"size:100;column:name_in_dp"`

	// Фамилия (в трёх падежах)
	SurnameInIp string `gorm:"size:100;column:surname_in_ip"`
	SurnameInRp string `gorm:"size:100;column:surname_in_rp"`
	SurnameInDp string `gorm:"size:100;column:surname_in_dp"`

	// Отчество (в трёх падежах)
	LastnameInIp string `gorm:"size:100;column:lastname_in_ip"`
	LastnameInRp string `gorm:"size:100;column:lastname_in_rp"`
	LastnameInDp string `gorm:"size:100;column:lastname_in_dp"`

	Email    string `gorm:"size:100;not null;unique;column:email"`
	Password string `gorm:"size:100;not null;column:password"`
	Role     string `gorm:"size:100;not null;column:role"`

	MobilePhone string `gorm:"size:20;column:mobile_phone"`
	WorkPhone   string `gorm:"size:20;column:work_phone"`
	Mail        string `gorm:"size:100;column:mail"`
	Sex         string `gorm:"size:10;column:sex"`

	Snils string `gorm:"size:20;column:snils"`

	Status        string `gorm:"size:100;default:'pending';column:status"`
	Confirmed     bool   `gorm:"default:false;column:confirmed"` // новое поле
	DeclineReason string `gorm:"type:text;column:rejection_reason"`

	RefreshToken string `gorm:"type:text;column:refresh_token"`
}

type Application struct {
	ID                        uint   `gorm:"primaryKey"`
	UserID                    uint   `gorm:"not null"` // внешний ключ
	ApplicationType           string `gorm:"size:255"`
	ApplicationNumber         string `gorm:"size:255"`
	NativeLanguage            string `gorm:"size:255"`
	Citizenship               string `gorm:"size:255"`
	MaritalStatus             string `gorm:"size:255"`
	Organization              string `gorm:"size:255"`
	JobPosition               string `gorm:"size:255"`
	RequestedCategory         string `gorm:"size:255"`
	BasisForAttestation       string `gorm:"size:255"`
	ExistingCategory          string `gorm:"size:255"`
	ExistingCategoryTerm      string `gorm:"size:255"`
	WorkExperience            string `gorm:"size:255"`
	CurrentPositionExperience string `gorm:"size:255"`
	AwardsInfo                string `gorm:"size:255"`
	TrainingInfo              string `gorm:"size:255"`
	Memberships               string `gorm:"size:255"`
	Consent                   bool
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

// UserDocument представляет загруженный пользователем документ.
type UserDocument struct {
	gorm.Model
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"user_id"`
	DocumentName string    `gorm:"not null" json:"document_name"`
	DocumentType string    `gorm:"not null" json:"document_type"`
	FilePath     string    `gorm:"not null" json:"file_path"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// DocumentInfo описывает информацию о документе для вывода.
type DocumentInfo struct {
	Filename string `json:"filename"`
	URL      string `json:"url"`
}

// UserDocumentsResponse описывает структуру ответа для документов пользователя.
type UserDocumentsResponse struct {
	UserID    uint           `json:"userId"`
	Documents []DocumentInfo `json:"documents"`
}

// EmailVerification представляет код подтверждения email.
type EmailVerification struct {
	ID        uint      `gorm:"primaryKey"`
	Email     string    `gorm:"not null"`
	Link      string    `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Passport struct {
	gorm.Model
	ID                   uint   `gorm:"primaryKey"`
	UserID               uint   `gorm:"not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	PassportSeries       string `gorm:"size:20"`
	PassportNumber       string `gorm:"size:20"`
	PassportIssuedBy     string `gorm:"size:200"`
	PassportIssueDate    time.Time
	PassportDivisionCode string `gorm:"size:20"`
	BirthDate            time.Time
	BirthPlace           string `gorm:"size:200"`
	RegistrationAddress  string `gorm:"size:200"`
}

type EducationDocument struct {
	gorm.Model
	ID                 uint `gorm:"primaryKey"`
	UserID             uint `gorm:"not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	InstitutionName    string
	CityName           string
	DiplomaSeries      string
	DiplomaRegNumber   string
	IssueDate          time.Time
	SpecialtyCode      string
	QualificationLevel string
}

/*
имя
фамилия
отчество
почта
пароль
номер телефона

-----
пол
и тд
*/
