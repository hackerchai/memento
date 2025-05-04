package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// AppConfig stores user-specific application configuration settings.
type AppConfig struct {
	bun.BaseModel `bun:"table:app_config,alias:ac"`

	ID                         uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	UserID                     uuid.UUID  `bun:"user_id,type:uuid,notnull,unique" json:"user_id"` // Unique constraint enforced by DB
	ScrapeImgOffline           bool       `bun:"scrape_img_offline,notnull,default:false" json:"scrape_img_offline"`
	LLMAutoGenTags             bool       `bun:"llm_auto_gen_tags,notnull,default:false" json:"llm_auto_gen_tags"`
	ExtractLinks               bool       `bun:"extract_links,notnull,default:false" json:"extract_links"`
	LLMProfileID               *uuid.UUID `bun:"llm_profile_id,type:uuid" json:"llm_profile_id,omitempty"`
	LLMProvider                *string    `bun:"llm_provider" json:"llm_provider,omitempty"`
	LLMAutoGenAbstract         bool       `bun:"llm_auto_gen_abstract,notnull,default:false" json:"llm_auto_gen_abstract"`
	CustomUserAgent            *string    `bun:"custom_user_agent" json:"custom_user_agent,omitempty"`
	CustomScrapeTimeoutSeconds *int       `bun:"custom_scrape_timeout_seconds" json:"custom_scrape_timeout_seconds,omitempty"`
	CustomScrapeRetryTimes     *int       `bun:"custom_scrape_retry_times" json:"custom_scrape_retry_times,omitempty"`
	CustomUserProxy            *string    `bun:"custom_user_proxy" json:"custom_user_proxy,omitempty"`
	BypassRefer                bool       `bun:"bypass_refer,notnull,default:false" json:"bypass_refer"`
	CreatedAt                  time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt                  time.Time  `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`

	// Relationships (optional, depending on query needs)
	// User *User `bun:"rel:belongs-to,join:user_id=id"`
}
