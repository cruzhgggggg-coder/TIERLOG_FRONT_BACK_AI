package models

import (
	"time"

	"gorm.io/gorm"
)

type UserRole string

const (
	RoleStudent  UserRole = "student"
	RoleLecturer UserRole = "lecturer"
)

type FeedbackCategory string

const (
	CategoryMinor FeedbackCategory = "Minor" // LOC
	CategoryMajor FeedbackCategory = "Major" // HOC
)

type FeedbackStatus string

const (
	StatusFixed     FeedbackStatus = "Fixed"
	StatusPending   FeedbackStatus = "Pending"
	StatusValidated FeedbackStatus = "Validated"
)

type User struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"not null;type:varchar(255)" json:"name"`
	Email     string         `gorm:"unique;not null;type:varchar(255)" json:"email"`
	Password  string         `gorm:"not null;type:varchar(255)" json:"-"`
	Role      UserRole       `gorm:"type:enum('student','lecturer');not null" json:"role"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Profiles
	Student  *Student  `gorm:"foreignKey:UserID" json:"student,omitempty"`
	Lecturer *Lecturer `gorm:"foreignKey:UserID" json:"lecturer,omitempty"`

	// AI Gateway Fields
	OpenAIKey       string `gorm:"type:varchar(255)" json:"openai_key"`
	GeminiKey       string `gorm:"type:varchar(255)" json:"gemini_key"`
	AnthropicKey    string `gorm:"type:varchar(255)" json:"anthropic_key"`
	NvidiaKey       string `gorm:"type:varchar(255)" json:"nvidia_key"`
	PreferredModel  string `gorm:"type:varchar(50);default:'default'" json:"preferred_model"`
	IsGatewayActive bool   `gorm:"default:false" json:"is_gateway_active"`
}

type RedeemCode struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Code      string         `gorm:"unique;not null;type:varchar(50)" json:"code"`
	IsUsed    bool           `gorm:"default:false" json:"is_used"`
	UsedBy    *uint64        `json:"used_by"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Lecturer struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64         `gorm:"not null" json:"user_id"`
	NIP       string         `gorm:"column:nip;unique;not null;type:varchar(20)" json:"nip"`
	Name      string         `gorm:"not null;type:varchar(100)" json:"name"`
	Faculty   string         `gorm:"type:varchar(100)" json:"faculty"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

type Student struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint64         `gorm:"not null" json:"user_id"`
	LecturerID  uint64         `gorm:"not null" json:"lecturer_id"`
	NIM         string         `gorm:"unique;not null;type:varchar(20)" json:"nim"`
	Name        string         `gorm:"not null;type:varchar(100)" json:"name"`
	Prodi       string         `gorm:"type:varchar(100)" json:"prodi"`
	ThesisTitle string         `gorm:"type:text" json:"thesis_title"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	User     *User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Lecturer *Lecturer `gorm:"foreignKey:LecturerID;constraint:OnDelete:RESTRICT" json:"lecturer,omitempty"`
}

type ConsultationLog struct {
	ID                 uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	StudentID          uint64         `gorm:"not null" json:"student_id"`
	AudioFilename      string         `gorm:"type:varchar(255)" json:"audio_filename"`
	TranscriptFilename string         `gorm:"type:varchar(255)" json:"transcript_filename"`
	TranscriptText     string         `gorm:"type:longtext" json:"transcript_text"`
	PaperFilename      string         `gorm:"type:varchar(255)" json:"paper_filename"`
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`

	Student       *Student       `gorm:"foreignKey:StudentID;constraint:OnDelete:CASCADE" json:"student,omitempty"`
	FeedbackItems []FeedbackItem `gorm:"foreignKey:ConsultationLogID;constraint:OnDelete:CASCADE" json:"feedback_items"`
}

type FeedbackItem struct {
	ID                 uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	ConsultationLogID uint64           `gorm:"column:consultation_log_id;not null" json:"log_id"`
	Content   string           `gorm:"type:text;not null" json:"content"`
	Category  FeedbackCategory `gorm:"type:enum('Minor','Major');not null" json:"category"`
	Status    FeedbackStatus   `gorm:"type:enum('Fixed','Pending','Validated');not null;default:'Pending'" json:"status"`
	CreatedAt time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt   `gorm:"index" json:"-"`
}
