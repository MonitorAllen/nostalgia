// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"time"

	"github.com/google/uuid"
)

type Article struct {
	ID uuid.UUID `json:"id"`
	// 标题
	Title string `json:"title"`
	// 摘要
	Summary string `json:"summary"`
	// 内容
	Content string `json:"content"`
	// 浏览量
	Views int32 `json:"views"`
	// 点赞数
	Likes int32 `json:"likes"`
	// 是否发布
	IsPublish bool `json:"is_publish"`
	// 拥有者
	Owner     uuid.UUID `json:"owner"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`
}

type Comment struct {
	ID int64 `json:"id"`
	// 评论内容
	Content string `json:"content"`
	// 文章ID
	ArticleID uuid.UUID `json:"article_id"`
	// 父评论ID
	ParentID int64 `json:"parent_id"`
	Likes    int32 `json:"likes"`
	// 评论人ID
	FromUserID uuid.UUID `json:"from_user_id"`
	// 被评论人ID
	ToUserID  uuid.UUID `json:"to_user_id"`
	CreatedAt time.Time `json:"created_at"`
	DeletedAt time.Time `json:"deleted_at"`
}

type Session struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	RefreshToken string    `json:"refresh_token"`
	UserAgent    string    `json:"user_agent"`
	ClientIp     string    `json:"client_ip"`
	IsBlocked    bool      `json:"is_blocked"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type Tag struct {
	ID int64 `json:"id"`
	// 名称
	Name string `json:"name"`
	// 文章ID
	ArticleID uuid.UUID `json:"article_id"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID              uuid.UUID `json:"id"`
	Username        string    `json:"username"`
	HashedPassword  string    `json:"hashed_password"`
	FullName        string    `json:"full_name"`
	Email           string    `json:"email"`
	IsEmailVerified bool      `json:"is_email_verified"`
	// 介绍
	About     string    `json:"about"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at"`
}

type VerifyEmail struct {
	ID         int64     `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	Email      string    `json:"email"`
	SecretCode string    `json:"secret_code"`
	IsUsed     bool      `json:"is_used"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiredAt  time.Time `json:"expired_at"`
}
