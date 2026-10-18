package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/x-reconstruction/api/internal/domain"
	"github.com/x-reconstruction/api/internal/recommendation"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(dbURL string) (*PostgresRepository, error) {
	if !strings.Contains(dbURL, "sslmode=") {
		if strings.Contains(dbURL, "?") {
			dbURL += "&sslmode=disable"
		} else {
			dbURL += "?sslmode=disable"
		}
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	maxOpen := 20
	if v := os.Getenv("DB_MAX_OPEN_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxOpen = n
		}
	}
	maxIdle := 10
	if v := os.Getenv("DB_MAX_IDLE_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxIdle = n
		}
	}

	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(15 * time.Minute)
	db.SetConnMaxIdleTime(3 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	repo := &PostgresRepository{db: db}
	repo.ensureSearchSchema(ctx)
	repo.ensureRBACSchema(ctx)
	repo.ensureOutboxSchema(ctx)

	return repo, nil
}

func (r *PostgresRepository) ensureSearchSchema(ctx context.Context) {
	stmts := []string{
		`CREATE EXTENSION IF NOT EXISTS pg_trgm;`,
		`DO $$ 
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'posts' AND column_name = 'search_vector'
			) THEN
				ALTER TABLE posts
				ADD COLUMN search_vector tsvector
				GENERATED ALWAYS AS (to_tsvector('simple', COALESCE(content, ''))) STORED;
			END IF;
		END $$;`,
		`CREATE INDEX IF NOT EXISTS idx_posts_search_vector ON posts USING GIN (search_vector);`,
		`CREATE INDEX IF NOT EXISTS idx_users_username_trgm ON users USING GIN (username gin_trgm_ops);`,
		`CREATE INDEX IF NOT EXISTS idx_user_profiles_display_name_trgm ON user_profiles USING GIN (display_name gin_trgm_ops);`,
		`CREATE INDEX IF NOT EXISTS idx_posts_search_keyset ON posts (created_at DESC, id DESC) WHERE deleted_at IS NULL;`,
		`CREATE TABLE IF NOT EXISTS recent_searches (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			query VARCHAR(255) NOT NULL,
			searched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(user_id, query)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_recent_searches_user ON recent_searches (user_id, searched_at DESC);`,
	}
	for _, stmt := range stmts {
		_, _ = r.db.ExecContext(ctx, stmt)
	}
}

func (r *PostgresRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *PostgresRepository) CreateUser(ctx context.Context, u *domain.User) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}

	if u.Role == "" || u.Role == "user" {
		u.Role = "member"
	}
	if strings.ToLower(strings.TrimSpace(u.Email)) == "afrizaahmad18@gmail.com" {
		u.Role = "owner"
		u.IsVerified = true
	}

	queryUser := `
		INSERT INTO users (id, username, email, password_hash, role, is_verified, verified_type, is_private, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING created_at, updated_at;
	`
	err = tx.QueryRowContext(ctx, queryUser, u.ID, u.Username, u.Email, u.PasswordHash, u.Role, u.IsVerified, u.VerifiedType, u.IsPrivate).Scan(&u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return err
	}

	queryProfile := `
		INSERT INTO user_profiles (user_id, display_name, bio, avatar_url, banner_url, location, website)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`
	_, err = tx.ExecContext(ctx, queryProfile, u.ID, u.DisplayName, u.Bio, u.AvatarURL, u.BannerURL, u.Location, u.Website)
	if err != nil {
		return err
	}

	// Assign default role in user_roles
	targetRoleKey := "member"
	if u.Role == "owner" {
		targetRoleKey = "owner"
	}
	_, _ = tx.ExecContext(ctx, `
		INSERT INTO user_roles (user_id, role_id, reason)
		SELECT $1, id, 'Default role assignment'
		FROM roles WHERE key = $2
		ON CONFLICT (user_id, role_id) DO NOTHING;
	`, u.ID, targetRoleKey)

	// Assign default membership
	targetPlan := "free"
	if u.Role == "owner" {
		targetPlan = "mvp"
	}
	_, _ = tx.ExecContext(ctx, `
		INSERT INTO user_memberships (user_id, plan_id, status)
		VALUES ($1, $2, 'active')
		ON CONFLICT (user_id, plan_id) DO NOTHING;
	`, u.ID, targetPlan)

	return tx.Commit()
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID) (*domain.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.role, u.is_verified, u.verified_type, u.is_private, u.created_at, u.updated_at,
		       p.display_name, p.bio, p.avatar_url, p.banner_url, p.location, p.website,
		       p.followers_count, p.following_count, p.posts_count,
		       COALESCE((SELECT TRUE FROM follows f WHERE f.follower_id = $2 AND f.following_id = u.id), FALSE) as is_following,
		       COALESCE((SELECT TRUE FROM follows f WHERE f.follower_id = u.id AND f.following_id = $2), FALSE) as is_followed_by
		FROM users u
		JOIN user_profiles p ON p.user_id = u.id
		WHERE u.id = $1 AND u.deleted_at IS NULL;
	`
	var u domain.User
	var vID interface{} = uuid.Nil
	if viewerID != nil {
		vID = *viewerID
	}

	err := r.db.QueryRowContext(ctx, query, id, vID).Scan(
		&u.ID, &u.Username, &u.Email, &u.Role, &u.IsVerified, &u.VerifiedType, &u.IsPrivate, &u.CreatedAt, &u.UpdatedAt,
		&u.DisplayName, &u.Bio, &u.AvatarURL, &u.BannerURL, &u.Location, &u.Website,
		&u.FollowersCount, &u.FollowingCount, &u.PostsCount,
		&u.IsFollowing, &u.IsFollowedBy,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &u, err
}

func (r *PostgresRepository) GetUserByUsername(ctx context.Context, username string, viewerID *uuid.UUID) (*domain.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.role, u.is_verified, u.verified_type, u.is_private, u.created_at, u.updated_at,
		       p.display_name, p.bio, p.avatar_url, p.banner_url, p.location, p.website,
		       p.followers_count, p.following_count, p.posts_count,
		       COALESCE((SELECT TRUE FROM follows f WHERE f.follower_id = $2 AND f.following_id = u.id), FALSE) as is_following,
		       COALESCE((SELECT TRUE FROM follows f WHERE f.follower_id = u.id AND f.following_id = $2), FALSE) as is_followed_by
		FROM users u
		JOIN user_profiles p ON p.user_id = u.id
		WHERE LOWER(u.username) = LOWER($1) AND u.deleted_at IS NULL;
	`
	var u domain.User
	var vID interface{} = uuid.Nil
	if viewerID != nil {
		vID = *viewerID
	}

	err := r.db.QueryRowContext(ctx, query, username, vID).Scan(
		&u.ID, &u.Username, &u.Email, &u.Role, &u.IsVerified, &u.VerifiedType, &u.IsPrivate, &u.CreatedAt, &u.UpdatedAt,
		&u.DisplayName, &u.Bio, &u.AvatarURL, &u.BannerURL, &u.Location, &u.Website,
		&u.FollowersCount, &u.FollowingCount, &u.PostsCount,
		&u.IsFollowing, &u.IsFollowedBy,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &u, err
}

func (r *PostgresRepository) GetUserByIdentifier(ctx context.Context, identifier string) (*domain.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.role, u.is_verified, u.verified_type, u.is_private, u.created_at, u.updated_at,
		       p.display_name, p.bio, p.avatar_url, p.banner_url, p.location, p.website,
		       p.followers_count, p.following_count, p.posts_count
		FROM users u
		JOIN user_profiles p ON p.user_id = u.id
		WHERE (LOWER(u.username) = LOWER($1) OR LOWER(u.email) = LOWER($1)) AND u.deleted_at IS NULL;
	`
	var u domain.User
	err := r.db.QueryRowContext(ctx, query, identifier).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role, &u.IsVerified, &u.VerifiedType, &u.IsPrivate, &u.CreatedAt, &u.UpdatedAt,
		&u.DisplayName, &u.Bio, &u.AvatarURL, &u.BannerURL, &u.Location, &u.Website,
		&u.FollowersCount, &u.FollowingCount, &u.PostsCount,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &u, err
}

func (r *PostgresRepository) UpdateProfile(ctx context.Context, u *domain.User) error {
	query := `
		UPDATE user_profiles
		SET display_name = $2, bio = $3, avatar_url = $4, banner_url = $5, location = $6, website = $7, updated_at = NOW()
		WHERE user_id = $1;
	`
	_, err := r.db.ExecContext(ctx, query, u.ID, u.DisplayName, u.Bio, u.AvatarURL, u.BannerURL, u.Location, u.Website)
	return err
}

func (r *PostgresRepository) CreateSession(ctx context.Context, s *domain.Session) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}

	ipStr := s.IPAddress
	if host, _, err := net.SplitHostPort(ipStr); err == nil {
		ipStr = host
	}

	var ip interface{} = nil
	if net.ParseIP(ipStr) != nil {
		ip = ipStr
	}

	query := `
		INSERT INTO sessions (id, user_id, token_hash, ip_address, user_agent, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW());
	`
	_, err := r.db.ExecContext(ctx, query, s.ID, s.UserID, s.TokenHash, ip, s.UserAgent, s.ExpiresAt)
	return err
}

func (r *PostgresRepository) GetSessionByTokenHash(ctx context.Context, hash string) (*domain.Session, error) {
	query := `
		SELECT id, user_id, token_hash, COALESCE(ip_address::text, ''), user_agent, expires_at, created_at
		FROM sessions
		WHERE token_hash = $1 AND expires_at > NOW();
	`
	var s domain.Session
	err := r.db.QueryRowContext(ctx, query, hash).Scan(&s.ID, &s.UserID, &s.TokenHash, &s.IPAddress, &s.UserAgent, &s.ExpiresAt, &s.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &s, err
}

func (r *PostgresRepository) DeleteSession(ctx context.Context, hash string) error {
	query := `DELETE FROM sessions WHERE token_hash = $1;`
	_, err := r.db.ExecContext(ctx, query, hash)
	return err
}

func (r *PostgresRepository) CreatePost(ctx context.Context, p *domain.Post) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}

	query := `
		INSERT INTO posts (id, author_id, content, reply_to_id, quote_post_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING created_at, updated_at;
	`
	err = tx.QueryRowContext(ctx, query, p.ID, p.AuthorID, p.Content, p.ReplyToID, p.QuotePostID).Scan(&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return err
	}

	// Update author post count
	_, err = tx.ExecContext(ctx, `UPDATE user_profiles SET posts_count = posts_count + 1 WHERE user_id = $1;`, p.AuthorID)
	if err != nil {
		return err
	}

	// If reply, increment reply count
	if p.ReplyToID != nil {
		_, err = tx.ExecContext(ctx, `UPDATE posts SET replies_count = replies_count + 1 WHERE id = $1;`, *p.ReplyToID)
		if err != nil {
			return err
		}
	}

	// Insert media relations
	for i, m := range p.Media {
		_, err = tx.ExecContext(ctx, `INSERT INTO post_media (post_id, media_id, position) VALUES ($1, $2, $3);`, p.ID, m.ID, i)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) GetPostByID(ctx context.Context, postID uuid.UUID, viewerID *uuid.UUID) (*domain.Post, error) {
	query := `
		SELECT p.id, p.author_id, p.content, p.reply_to_id, p.quote_post_id,
		       p.replies_count, p.reposts_count, p.likes_count, p.views_count, p.created_at, p.updated_at,
		       u.username, u.is_verified, u.verified_type,
		       pr.display_name, pr.avatar_url,
		       COALESCE((SELECT TRUE FROM post_likes pl WHERE pl.post_id = p.id AND pl.user_id = $2), FALSE) as is_liked,
		       COALESCE((SELECT TRUE FROM post_reposts prp WHERE prp.post_id = p.id AND prp.user_id = $2), FALSE) as is_reposted,
		       COALESCE((SELECT TRUE FROM post_bookmarks pb WHERE pb.post_id = p.id AND pb.user_id = $2), FALSE) as is_bookmarked
		FROM posts p
		JOIN users u ON u.id = p.author_id
		JOIN user_profiles pr ON pr.user_id = p.author_id
		WHERE p.id = $1 AND p.deleted_at IS NULL;
	`
	var vID interface{} = uuid.Nil
	if viewerID != nil {
		vID = *viewerID
	}

	var p domain.Post
	var author domain.User

	err := r.db.QueryRowContext(ctx, query, postID, vID).Scan(
		&p.ID, &p.AuthorID, &p.Content, &p.ReplyToID, &p.QuotePostID,
		&p.RepliesCount, &p.RepostsCount, &p.LikesCount, &p.ViewsCount, &p.CreatedAt, &p.UpdatedAt,
		&author.Username, &author.IsVerified, &author.VerifiedType,
		&author.DisplayName, &author.AvatarURL,
		&p.IsLiked, &p.IsReposted, &p.IsBookmarked,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	author.ID = p.AuthorID
	p.Author = &author
	single := []domain.Post{p}
	_ = r.attachMediaToPosts(ctx, single)
	p.Media = single[0].Media
	return &p, nil
}

func (r *PostgresRepository) DeletePost(ctx context.Context, postID, authorID uuid.UUID) error {
	query := `UPDATE posts SET deleted_at = NOW() WHERE id = $1 AND author_id = $2;`
	res, err := r.db.ExecContext(ctx, query, postID, authorID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrUnauthorized
	}
	return nil
}

func (r *PostgresRepository) GetFeed(ctx context.Context, feedType string, viewerID *uuid.UUID, cursor string, limit int) ([]domain.Post, string, error) {
	engine := recommendation.NewEngine(r)
	res, err := engine.GenerateFeed(ctx, recommendation.FeedOptions{
		ViewerID: viewerID,
		FeedType: feedType,
		Cursor:   cursor,
		Limit:    limit,
	})
	if err != nil {
		return nil, "", err
	}
	if len(res.Posts) > 0 {
		_ = r.attachMediaToPosts(ctx, res.Posts)
	}
	return res.Posts, res.NextCursor, nil
}

func (r *PostgresRepository) scanPostRows(rows *sql.Rows) ([]domain.Post, error) {
	var posts []domain.Post
	for rows.Next() {
		var p domain.Post
		var author domain.User
		err := rows.Scan(
			&p.ID, &p.AuthorID, &p.Content, &p.ReplyToID, &p.QuotePostID,
			&p.RepliesCount, &p.RepostsCount, &p.LikesCount, &p.ViewsCount, &p.CreatedAt, &p.UpdatedAt,
			&author.Username, &author.IsVerified, &author.VerifiedType,
			&author.DisplayName, &author.AvatarURL,
			&p.IsLiked, &p.IsReposted, &p.IsBookmarked,
		)
		if err != nil {
			return nil, err
		}
		author.ID = p.AuthorID
		p.Author = &author
		posts = append(posts, p)
	}
	return posts, nil
}

func (r *PostgresRepository) attachQuotesToPosts(ctx context.Context, posts []domain.Post, viewerID *uuid.UUID) error {
	if len(posts) == 0 {
		return nil
	}
	quoteIDs := make([]uuid.UUID, 0)
	for _, p := range posts {
		if p.QuotePostID != nil && *p.QuotePostID != uuid.Nil {
			quoteIDs = append(quoteIDs, *p.QuotePostID)
		}
	}
	if len(quoteIDs) == 0 {
		return nil
	}

	var vID interface{} = uuid.Nil
	if viewerID != nil {
		vID = *viewerID
	}

	query := `
		SELECT p.id, p.author_id, p.content, p.reply_to_id, p.quote_post_id,
		       p.replies_count, p.reposts_count, p.likes_count, COALESCE(p.quotes_count, 0), p.views_count, p.created_at, p.updated_at,
		       u.username, u.is_verified, u.verified_type,
		       pr.display_name, pr.avatar_url,
		       COALESCE((SELECT TRUE FROM post_likes pl WHERE pl.post_id = p.id AND pl.user_id = $2), FALSE) as is_liked,
		       COALESCE((SELECT TRUE FROM post_reposts prp WHERE prp.post_id = p.id AND prp.user_id = $2), FALSE) as is_reposted,
		       COALESCE((SELECT TRUE FROM post_bookmarks pb WHERE pb.post_id = p.id AND pb.user_id = $2), FALSE) as is_bookmarked
		FROM posts p
		JOIN users u ON u.id = p.author_id
		JOIN user_profiles pr ON pr.user_id = p.author_id
		WHERE p.id = ANY($1) AND p.deleted_at IS NULL;
	`
	rows, err := r.db.QueryContext(ctx, query, pq.Array(quoteIDs), vID)
	if err != nil {
		return err
	}
	defer rows.Close()

	quotedMap := make(map[uuid.UUID]*domain.Post)
	quotedList := make([]domain.Post, 0)
	for rows.Next() {
		var qp domain.Post
		var author domain.User
		err := rows.Scan(
			&qp.ID, &qp.AuthorID, &qp.Content, &qp.ReplyToID, &qp.QuotePostID,
			&qp.RepliesCount, &qp.RepostsCount, &qp.LikesCount, &qp.QuotesCount, &qp.ViewsCount, &qp.CreatedAt, &qp.UpdatedAt,
			&author.Username, &author.IsVerified, &author.VerifiedType,
			&author.DisplayName, &author.AvatarURL,
			&qp.IsLiked, &qp.IsReposted, &qp.IsBookmarked,
		)
		if err == nil {
			author.ID = qp.AuthorID
			qp.Author = &author
			quotedMap[qp.ID] = &qp
			quotedList = append(quotedList, qp)
		}
	}

	if len(quotedList) > 0 {
		_ = r.attachMediaToPosts(ctx, quotedList)
		for _, qp := range quotedList {
			if ptr, ok := quotedMap[qp.ID]; ok {
				ptr.Media = qp.Media
			}
		}
	}

	for i := range posts {
		if posts[i].QuotePostID != nil {
			if qp, ok := quotedMap[*posts[i].QuotePostID]; ok {
				posts[i].QuotePost = qp
			}
		}
	}
	return nil
}

func (r *PostgresRepository) attachParentPosts(ctx context.Context, posts []domain.Post) error {
	if len(posts) == 0 {
		return nil
	}
	parentIDs := make([]uuid.UUID, 0)
	for _, p := range posts {
		if p.ReplyToID != nil && *p.ReplyToID != uuid.Nil {
			parentIDs = append(parentIDs, *p.ReplyToID)
		}
	}
	if len(parentIDs) == 0 {
		return nil
	}

	query := `
		SELECT p.id, p.author_id, u.username, pr.display_name
		FROM posts p
		JOIN users u ON u.id = p.author_id
		JOIN user_profiles pr ON pr.user_id = p.author_id
		WHERE p.id = ANY($1);
	`
	rows, err := r.db.QueryContext(ctx, query, pq.Array(parentIDs))
	if err != nil {
		return err
	}
	defer rows.Close()

	parentMap := make(map[uuid.UUID]*domain.Post)
	for rows.Next() {
		var parentID, authorID uuid.UUID
		var username, displayName string
		if err := rows.Scan(&parentID, &authorID, &username, &displayName); err == nil {
			parentMap[parentID] = &domain.Post{
				ID:       parentID,
				AuthorID: authorID,
				Author: &domain.User{
					ID:          authorID,
					Username:    username,
					DisplayName: displayName,
				},
			}
		}
	}

	for i := range posts {
		if posts[i].ReplyToID != nil {
			if parent, ok := parentMap[*posts[i].ReplyToID]; ok {
				posts[i].ParentPost = parent
			}
		}
	}
	return nil
}

func (r *PostgresRepository) GetUserPosts(ctx context.Context, username, tab string, viewerID *uuid.UUID, cursor string, limit int) ([]domain.Post, string, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	var vID interface{} = uuid.Nil
	if viewerID != nil {
		vID = *viewerID
	}

	var targetUser domain.User
	err := r.db.QueryRowContext(ctx, `
		SELECT u.id, u.username, u.is_verified, u.verified_type, pr.display_name, pr.avatar_url
		FROM users u
		JOIN user_profiles pr ON pr.user_id = u.id
		WHERE LOWER(u.username) = LOWER($1) AND u.deleted_at IS NULL;
	`, username).Scan(
		&targetUser.ID, &targetUser.Username, &targetUser.IsVerified, &targetUser.VerifiedType,
		&targetUser.DisplayName, &targetUser.AvatarURL,
	)
	if err != nil {
		return []domain.Post{}, "", nil
	}

	normTab := strings.ToLower(strings.TrimSpace(tab))
	var cur *recommendation.FeedCursor
	if cursor != "" {
		cur, _ = recommendation.DecodeCursor(cursor)
	}

	var posts []domain.Post
	var query string
	var args []interface{}

	switch normTab {
	case "replies", "reply":
		// Tab 2: User's replies to other posts
		query = `
			SELECT p.id, p.author_id, p.content, p.reply_to_id, p.quote_post_id,
			       p.replies_count, p.reposts_count, p.likes_count, COALESCE(p.quotes_count, 0), p.views_count, p.created_at, p.updated_at,
			       u.username, u.is_verified, u.verified_type,
			       pr.display_name, pr.avatar_url,
			       COALESCE((SELECT TRUE FROM post_likes pl WHERE pl.post_id = p.id AND pl.user_id = $2), FALSE) as is_liked,
			       COALESCE((SELECT TRUE FROM post_reposts prp WHERE prp.post_id = p.id AND prp.user_id = $2), FALSE) as is_reposted,
			       COALESCE((SELECT TRUE FROM post_bookmarks pb WHERE pb.post_id = p.id AND pb.user_id = $2), FALSE) as is_bookmarked
			FROM posts p
			JOIN users u ON u.id = p.author_id
			JOIN user_profiles pr ON pr.user_id = p.author_id
			WHERE p.author_id = $1 AND p.reply_to_id IS NOT NULL AND p.deleted_at IS NULL
		`
		args = []interface{}{targetUser.ID, vID}
		if cur != nil {
			query += ` AND (p.created_at < $3 OR (p.created_at = $3 AND p.id < $4))`
			args = append(args, cur.CreatedAt, cur.PostID)
		}
		query += fmt.Sprintf(` ORDER BY p.created_at DESC, p.id DESC LIMIT $%d;`, len(args)+1)
		args = append(args, limit)

	case "reposts", "repost":
		// Tab 3: Original posts reposted by the target user with reposter attribution
		query = `
			SELECT p.id, p.author_id, p.content, p.reply_to_id, p.quote_post_id,
			       p.replies_count, p.reposts_count, p.likes_count, COALESCE(p.quotes_count, 0), p.views_count, p.created_at, p.updated_at,
			       u.username, u.is_verified, u.verified_type,
			       pr.display_name, pr.avatar_url,
			       COALESCE((SELECT TRUE FROM post_likes pl WHERE pl.post_id = p.id AND pl.user_id = $2), FALSE) as is_liked,
			       COALESCE((SELECT TRUE FROM post_reposts prp WHERE prp.post_id = p.id AND prp.user_id = $2), FALSE) as is_reposted,
			       COALESCE((SELECT TRUE FROM post_bookmarks pb WHERE pb.post_id = p.id AND pb.user_id = $2), FALSE) as is_bookmarked,
			       rep.created_at as reposted_at
			FROM post_reposts rep
			JOIN posts p ON p.id = rep.post_id
			JOIN users u ON u.id = p.author_id
			JOIN user_profiles pr ON pr.user_id = p.author_id
			WHERE rep.user_id = $1 AND p.deleted_at IS NULL
		`
		args = []interface{}{targetUser.ID, vID}
		if cur != nil {
			query += ` AND (rep.created_at < $3 OR (rep.created_at = $3 AND p.id < $4))`
			args = append(args, cur.CreatedAt, cur.PostID)
		}
		query += fmt.Sprintf(` ORDER BY rep.created_at DESC, p.id DESC LIMIT $%d;`, len(args)+1)
		args = append(args, limit)

	case "media":
		// Tab 4: User's posts containing media attachments
		query = `
			SELECT p.id, p.author_id, p.content, p.reply_to_id, p.quote_post_id,
			       p.replies_count, p.reposts_count, p.likes_count, COALESCE(p.quotes_count, 0), p.views_count, p.created_at, p.updated_at,
			       u.username, u.is_verified, u.verified_type,
			       pr.display_name, pr.avatar_url,
			       COALESCE((SELECT TRUE FROM post_likes pl WHERE pl.post_id = p.id AND pl.user_id = $2), FALSE) as is_liked,
			       COALESCE((SELECT TRUE FROM post_reposts prp WHERE prp.post_id = p.id AND prp.user_id = $2), FALSE) as is_reposted,
			       COALESCE((SELECT TRUE FROM post_bookmarks pb WHERE pb.post_id = p.id AND pb.user_id = $2), FALSE) as is_bookmarked
			FROM posts p
			JOIN users u ON u.id = p.author_id
			JOIN user_profiles pr ON pr.user_id = p.author_id
			WHERE p.author_id = $1 AND p.deleted_at IS NULL
			  AND EXISTS (SELECT 1 FROM post_media pm WHERE pm.post_id = p.id)
		`
		args = []interface{}{targetUser.ID, vID}
		if cur != nil {
			query += ` AND (p.created_at < $3 OR (p.created_at = $3 AND p.id < $4))`
			args = append(args, cur.CreatedAt, cur.PostID)
		}
		query += fmt.Sprintf(` ORDER BY p.created_at DESC, p.id DESC LIMIT $%d;`, len(args)+1)
		args = append(args, limit)

	default:
		// Tab 1: Postingan (original posts and quote posts authored by user, excluding pure replies)
		query = `
			SELECT p.id, p.author_id, p.content, p.reply_to_id, p.quote_post_id,
			       p.replies_count, p.reposts_count, p.likes_count, COALESCE(p.quotes_count, 0), p.views_count, p.created_at, p.updated_at,
			       u.username, u.is_verified, u.verified_type,
			       pr.display_name, pr.avatar_url,
			       COALESCE((SELECT TRUE FROM post_likes pl WHERE pl.post_id = p.id AND pl.user_id = $2), FALSE) as is_liked,
			       COALESCE((SELECT TRUE FROM post_reposts prp WHERE prp.post_id = p.id AND prp.user_id = $2), FALSE) as is_reposted,
			       COALESCE((SELECT TRUE FROM post_bookmarks pb WHERE pb.post_id = p.id AND pb.user_id = $2), FALSE) as is_bookmarked
			FROM posts p
			JOIN users u ON u.id = p.author_id
			JOIN user_profiles pr ON pr.user_id = p.author_id
			WHERE p.author_id = $1 AND p.reply_to_id IS NULL AND p.deleted_at IS NULL
		`
		args = []interface{}{targetUser.ID, vID}
		if cur != nil {
			query += ` AND (p.created_at < $3 OR (p.created_at = $3 AND p.id < $4))`
			args = append(args, cur.CreatedAt, cur.PostID)
		}
		query += fmt.Sprintf(` ORDER BY p.created_at DESC, p.id DESC LIMIT $%d;`, len(args)+1)
		args = append(args, limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return []domain.Post{}, "", err
	}
	defer rows.Close()

	posts = make([]domain.Post, 0)
	for rows.Next() {
		var p domain.Post
		var author domain.User
		var repostedAt *time.Time

		var scanErr error
		if normTab == "reposts" || normTab == "repost" {
			scanErr = rows.Scan(
				&p.ID, &p.AuthorID, &p.Content, &p.ReplyToID, &p.QuotePostID,
				&p.RepliesCount, &p.RepostsCount, &p.LikesCount, &p.QuotesCount, &p.ViewsCount, &p.CreatedAt, &p.UpdatedAt,
				&author.Username, &author.IsVerified, &author.VerifiedType,
				&author.DisplayName, &author.AvatarURL,
				&p.IsLiked, &p.IsReposted, &p.IsBookmarked,
				&repostedAt,
			)
		} else {
			scanErr = rows.Scan(
				&p.ID, &p.AuthorID, &p.Content, &p.ReplyToID, &p.QuotePostID,
				&p.RepliesCount, &p.RepostsCount, &p.LikesCount, &p.QuotesCount, &p.ViewsCount, &p.CreatedAt, &p.UpdatedAt,
				&author.Username, &author.IsVerified, &author.VerifiedType,
				&author.DisplayName, &author.AvatarURL,
				&p.IsLiked, &p.IsReposted, &p.IsBookmarked,
			)
		}
		if scanErr != nil {
			return []domain.Post{}, "", scanErr
		}
		author.ID = p.AuthorID
		p.Author = &author

		if normTab == "reposts" || normTab == "repost" {
			repUser := targetUser
			p.Reposter = &repUser
			p.RepostedAt = repostedAt
		}

		posts = append(posts, p)
	}

	if len(posts) > 0 {
		_ = r.attachMediaToPosts(ctx, posts)
		_ = r.attachQuotesToPosts(ctx, posts, viewerID)
		_ = r.attachParentPosts(ctx, posts)
	}

	nextCursor := ""
	if len(posts) == limit {
		lastPost := posts[len(posts)-1]
		if (normTab == "reposts" || normTab == "repost") && lastPost.RepostedAt != nil {
			nextCursor = recommendation.EncodeCursor(*lastPost.RepostedAt, lastPost.ID)
		} else {
			nextCursor = recommendation.EncodeCursor(lastPost.CreatedAt, lastPost.ID)
		}
	}

	return posts, nextCursor, nil
}

func (r *PostgresRepository) GetPostReplies(ctx context.Context, postID uuid.UUID, viewerID *uuid.UUID, cursor string, limit int, sortBy string) ([]domain.Post, string, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	var vID interface{} = uuid.Nil
	if viewerID != nil {
		vID = *viewerID
	}

	var cur *recommendation.FeedCursor
	if cursor != "" {
		cur, _ = recommendation.DecodeCursor(cursor)
	}

	query := `
		SELECT p.id, p.author_id, p.content, p.reply_to_id, p.quote_post_id,
		       p.replies_count, p.reposts_count, p.likes_count, COALESCE(p.quotes_count, 0), p.views_count, p.created_at, p.updated_at,
		       u.username, u.is_verified, u.verified_type,
		       pr.display_name, pr.avatar_url,
		       COALESCE((SELECT TRUE FROM post_likes pl WHERE pl.post_id = p.id AND pl.user_id = $2), FALSE) as is_liked,
		       COALESCE((SELECT TRUE FROM post_reposts prp WHERE prp.post_id = p.id AND prp.user_id = $2), FALSE) as is_reposted,
		       COALESCE((SELECT TRUE FROM post_bookmarks pb WHERE pb.post_id = p.id AND pb.user_id = $2), FALSE) as is_bookmarked
		FROM posts p
		JOIN users u ON u.id = p.author_id
		JOIN user_profiles pr ON pr.user_id = p.author_id
		WHERE p.reply_to_id = $1 AND p.deleted_at IS NULL
	`
	args := []interface{}{postID, vID}
	if cur != nil {
		query += ` AND (p.created_at > $3 OR (p.created_at = $3 AND p.id > $4))`
		args = append(args, cur.CreatedAt, cur.PostID)
	}

	switch strings.ToLower(sortBy) {
	case "recent", "terkini":
		query += fmt.Sprintf(` ORDER BY p.created_at DESC, p.id DESC LIMIT $%d;`, len(args)+1)
	case "likes", "suka":
		query += fmt.Sprintf(` ORDER BY p.likes_count DESC, p.created_at DESC, p.id DESC LIMIT $%d;`, len(args)+1)
	default: // "relevant" / "relevan"
		query += fmt.Sprintf(` ORDER BY (CASE WHEN u.is_verified THEN 100 ELSE 0 END + p.likes_count * 2 + p.replies_count) DESC, p.created_at DESC, p.id DESC LIMIT $%d;`, len(args)+1)
	}
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return []domain.Post{}, "", err
	}
	defer rows.Close()

	posts := make([]domain.Post, 0)
	for rows.Next() {
		var p domain.Post
		var author domain.User
		err := rows.Scan(
			&p.ID, &p.AuthorID, &p.Content, &p.ReplyToID, &p.QuotePostID,
			&p.RepliesCount, &p.RepostsCount, &p.LikesCount, &p.QuotesCount, &p.ViewsCount, &p.CreatedAt, &p.UpdatedAt,
			&author.Username, &author.IsVerified, &author.VerifiedType,
			&author.DisplayName, &author.AvatarURL,
			&p.IsLiked, &p.IsReposted, &p.IsBookmarked,
		)
		if err != nil {
			return []domain.Post{}, "", err
		}
		author.ID = p.AuthorID
		p.Author = &author
		posts = append(posts, p)
	}
	if len(posts) > 0 {
		_ = r.attachMediaToPosts(ctx, posts)
		_ = r.attachQuotesToPosts(ctx, posts, viewerID)
		_ = r.attachParentPosts(ctx, posts)
	}

	nextCursor := ""
	if len(posts) == limit {
		lastPost := posts[len(posts)-1]
		nextCursor = recommendation.EncodeCursor(lastPost.CreatedAt, lastPost.ID)
	}

	return posts, nextCursor, nil
}

func (r *PostgresRepository) GetPostQuotes(ctx context.Context, postID uuid.UUID, viewerID *uuid.UUID, cursor string, limit int) ([]domain.Post, string, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	var vID interface{} = uuid.Nil
	if viewerID != nil {
		vID = *viewerID
	}

	var cur *recommendation.FeedCursor
	if cursor != "" {
		cur, _ = recommendation.DecodeCursor(cursor)
	}

	query := `
		SELECT p.id, p.author_id, p.content, p.reply_to_id, p.quote_post_id,
		       p.replies_count, p.reposts_count, p.likes_count, COALESCE(p.quotes_count, 0), p.views_count, p.created_at, p.updated_at,
		       u.username, u.is_verified, u.verified_type,
		       pr.display_name, pr.avatar_url,
		       COALESCE((SELECT TRUE FROM post_likes pl WHERE pl.post_id = p.id AND pl.user_id = $2), FALSE) as is_liked,
		       COALESCE((SELECT TRUE FROM post_reposts prp WHERE prp.post_id = p.id AND prp.user_id = $2), FALSE) as is_reposted,
		       COALESCE((SELECT TRUE FROM post_bookmarks pb WHERE pb.post_id = p.id AND pb.user_id = $2), FALSE) as is_bookmarked
		FROM posts p
		JOIN users u ON u.id = p.author_id
		JOIN user_profiles pr ON pr.user_id = p.author_id
		WHERE p.quote_post_id = $1 AND p.deleted_at IS NULL
	`
	args := []interface{}{postID, vID}
	if cur != nil {
		query += ` AND (p.created_at < $3 OR (p.created_at = $3 AND p.id < $4))`
		args = append(args, cur.CreatedAt, cur.PostID)
	}
	query += fmt.Sprintf(` ORDER BY p.created_at DESC, p.id DESC LIMIT $%d;`, len(args)+1)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return []domain.Post{}, "", err
	}
	defer rows.Close()

	posts := make([]domain.Post, 0)
	for rows.Next() {
		var p domain.Post
		var author domain.User
		err := rows.Scan(
			&p.ID, &p.AuthorID, &p.Content, &p.ReplyToID, &p.QuotePostID,
			&p.RepliesCount, &p.RepostsCount, &p.LikesCount, &p.QuotesCount, &p.ViewsCount, &p.CreatedAt, &p.UpdatedAt,
			&author.Username, &author.IsVerified, &author.VerifiedType,
			&author.DisplayName, &author.AvatarURL,
			&p.IsLiked, &p.IsReposted, &p.IsBookmarked,
		)
		if err != nil {
			return []domain.Post{}, "", err
		}
		author.ID = p.AuthorID
		p.Author = &author
		posts = append(posts, p)
	}
	if len(posts) > 0 {
		_ = r.attachMediaToPosts(ctx, posts)
		_ = r.attachQuotesToPosts(ctx, posts, viewerID)
	}

	nextCursor := ""
	if len(posts) == limit {
		lastPost := posts[len(posts)-1]
		nextCursor = recommendation.EncodeCursor(lastPost.CreatedAt, lastPost.ID)
	}

	return posts, nextCursor, nil
}

func (r *PostgresRepository) LikePost(ctx context.Context, postID, userID uuid.UUID) (bool, int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `INSERT INTO post_likes (post_id, user_id, created_at) VALUES ($1, $2, NOW()) ON CONFLICT (post_id, user_id) DO NOTHING;`, postID, userID)
	if err != nil {
		return false, 0, err
	}

	rows, _ := res.RowsAffected()
	var count int
	if rows > 0 {
		err = tx.QueryRowContext(ctx, `UPDATE posts SET likes_count = (SELECT COUNT(*) FROM post_likes WHERE post_id = $1) WHERE id = $1 RETURNING likes_count;`, postID).Scan(&count)
		if err != nil {
			return false, 0, err
		}
	} else {
		err = tx.QueryRowContext(ctx, `SELECT likes_count FROM posts WHERE id = $1;`, postID).Scan(&count)
		if err != nil {
			return false, 0, err
		}
	}

	return true, count, tx.Commit()
}

func (r *PostgresRepository) UnlikePost(ctx context.Context, postID, userID uuid.UUID) (bool, int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `DELETE FROM post_likes WHERE post_id = $1 AND user_id = $2;`, postID, userID)
	if err != nil {
		return false, 0, err
	}
	rows, _ := res.RowsAffected()
	var count int
	if rows > 0 {
		err = tx.QueryRowContext(ctx, `UPDATE posts SET likes_count = (SELECT COUNT(*) FROM post_likes WHERE post_id = $1) WHERE id = $1 RETURNING likes_count;`, postID).Scan(&count)
		if err != nil {
			return false, 0, err
		}
	} else {
		err = tx.QueryRowContext(ctx, `SELECT likes_count FROM posts WHERE id = $1;`, postID).Scan(&count)
		if err != nil {
			return false, 0, err
		}
	}

	return false, count, tx.Commit()
}

func (r *PostgresRepository) RepostPost(ctx context.Context, postID, userID uuid.UUID) (bool, int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `INSERT INTO post_reposts (post_id, user_id, created_at) VALUES ($1, $2, NOW()) ON CONFLICT (post_id, user_id) DO NOTHING;`, postID, userID)
	if err != nil {
		return false, 0, err
	}

	rows, _ := res.RowsAffected()
	var count int
	if rows > 0 {
		err = tx.QueryRowContext(ctx, `UPDATE posts SET reposts_count = (SELECT COUNT(*) FROM post_reposts WHERE post_id = $1) WHERE id = $1 RETURNING reposts_count;`, postID).Scan(&count)
		if err != nil {
			return false, 0, err
		}
	} else {
		err = tx.QueryRowContext(ctx, `SELECT reposts_count FROM posts WHERE id = $1;`, postID).Scan(&count)
		if err != nil {
			return false, 0, err
		}
	}

	return true, count, tx.Commit()
}

func (r *PostgresRepository) UndoRepost(ctx context.Context, postID, userID uuid.UUID) (bool, int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `DELETE FROM post_reposts WHERE post_id = $1 AND user_id = $2;`, postID, userID)
	if err != nil {
		return false, 0, err
	}
	rows, _ := res.RowsAffected()
	var count int
	if rows > 0 {
		err = tx.QueryRowContext(ctx, `UPDATE posts SET reposts_count = (SELECT COUNT(*) FROM post_reposts WHERE post_id = $1) WHERE id = $1 RETURNING reposts_count;`, postID).Scan(&count)
		if err != nil {
			return false, 0, err
		}
	} else {
		err = tx.QueryRowContext(ctx, `SELECT reposts_count FROM posts WHERE id = $1;`, postID).Scan(&count)
		if err != nil {
			return false, 0, err
		}
	}

	return false, count, tx.Commit()
}

func (r *PostgresRepository) BookmarkPost(ctx context.Context, postID, userID uuid.UUID) (bool, int, error) {
	_, err := r.db.ExecContext(ctx, `INSERT INTO post_bookmarks (post_id, user_id, created_at) VALUES ($1, $2, NOW()) ON CONFLICT DO NOTHING;`, postID, userID)
	return true, 1, err
}

func (r *PostgresRepository) RemoveBookmark(ctx context.Context, postID, userID uuid.UUID) (bool, int, error) {
	_, err := r.db.ExecContext(ctx, `DELETE FROM post_bookmarks WHERE post_id = $1 AND user_id = $2;`, postID, userID)
	return false, 0, err
}

func (r *PostgresRepository) FollowUser(ctx context.Context, followerID, followingID uuid.UUID) (bool, int, error) {
	if followerID == followingID {
		return false, 0, errors.New("cannot follow yourself")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `INSERT INTO follows (follower_id, following_id, created_at) VALUES ($1, $2, NOW()) ON CONFLICT DO NOTHING;`, followerID, followingID)
	if err != nil {
		return false, 0, err
	}
	rows, _ := res.RowsAffected()
	var followersCount int
	if rows > 0 {
		_ = tx.QueryRowContext(ctx, `UPDATE user_profiles SET followers_count = followers_count + 1 WHERE user_id = $1 RETURNING followers_count;`, followingID).Scan(&followersCount)
		_, _ = tx.ExecContext(ctx, `UPDATE user_profiles SET following_count = following_count + 1 WHERE user_id = $1;`, followerID)
	} else {
		_ = tx.QueryRowContext(ctx, `SELECT followers_count FROM user_profiles WHERE user_id = $1;`, followingID).Scan(&followersCount)
	}

	return true, followersCount, tx.Commit()
}

func (r *PostgresRepository) UnfollowUser(ctx context.Context, followerID, followingID uuid.UUID) (bool, int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `DELETE FROM follows WHERE follower_id = $1 AND following_id = $2;`, followerID, followingID)
	if err != nil {
		return false, 0, err
	}
	rows, _ := res.RowsAffected()
	var followersCount int
	if rows > 0 {
		_ = tx.QueryRowContext(ctx, `UPDATE user_profiles SET followers_count = GREATEST(followers_count - 1, 0) WHERE user_id = $1 RETURNING followers_count;`, followingID).Scan(&followersCount)
		_, _ = tx.ExecContext(ctx, `UPDATE user_profiles SET following_count = GREATEST(following_count - 1, 0) WHERE user_id = $1;`, followerID)
	} else {
		_ = tx.QueryRowContext(ctx, `SELECT followers_count FROM user_profiles WHERE user_id = $1;`, followingID).Scan(&followersCount)
	}

	return false, followersCount, tx.Commit()
}

func (r *PostgresRepository) attachMediaToPosts(ctx context.Context, posts []domain.Post) error {
	if len(posts) == 0 {
		return nil
	}
	postIDs := make([]uuid.UUID, len(posts))
	postIndexMap := make(map[uuid.UUID]int)
	for i, p := range posts {
		postIDs[i] = p.ID
		postIndexMap[p.ID] = i
	}

	query := `
		SELECT pm.post_id, mo.id, mo.uploader_id, mo.storage_key, mo.url, mo.media_type, mo.mime_type,
		       mo.file_size_bytes, mo.width, mo.height, mo.duration_seconds, mo.created_at
		FROM post_media pm
		JOIN media_objects mo ON mo.id = pm.media_id
		WHERE pm.post_id = ANY($1)
		ORDER BY pm.position ASC;
	`
	rows, err := r.db.QueryContext(ctx, query, pq.Array(postIDs))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var postID uuid.UUID
		var m domain.MediaObject
		if err := rows.Scan(
			&postID, &m.ID, &m.UploaderID, &m.StorageKey, &m.URL, &m.MediaType, &m.MimeType,
			&m.FileSizeBytes, &m.Width, &m.Height, &m.DurationSeconds, &m.CreatedAt,
		); err == nil {
			if idx, ok := postIndexMap[postID]; ok {
				posts[idx].Media = append(posts[idx].Media, m)
			}
		}
	}
	return nil
}

func (r *PostgresRepository) SearchPostsAdvanced(ctx context.Context, params domain.SearchFilterParams, viewerID *uuid.UUID) ([]domain.Post, string, error) {
	limit := params.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	var vID interface{} = uuid.Nil
	if viewerID != nil {
		vID = *viewerID
	}

	var args []interface{}
	argIdx := 1

	args = append(args, vID) // $1 = viewerID
	argIdx++

	var whereClauses []string
	whereClauses = append(whereClauses, "p.deleted_at IS NULL")

	if viewerID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf(`p.author_id NOT IN (
			SELECT blocked_id FROM blocks WHERE blocker_id = $%d
			UNION
			SELECT blocker_id FROM blocks WHERE blocked_id = $%d
		)`, 1, 1))
	}

	// Filter query text
	var textArgIdx int
	if params.Query != "" {
		args = append(args, params.Query)
		textArgIdx = argIdx
		argIdx++
		whereClauses = append(whereClauses, fmt.Sprintf(`(
			(p.search_vector IS NOT NULL AND p.search_vector @@ plainto_tsquery('simple', $%d))
			OR p.content ILIKE '%%' || $%d || '%%'
		)`, textArgIdx, textArgIdx))
	}

	// from:username
	if params.FromUsername != "" {
		args = append(args, params.FromUsername)
		whereClauses = append(whereClauses, fmt.Sprintf("u.username ILIKE $%d", argIdx))
		argIdx++
	}

	// to:username
	if params.ToUsername != "" {
		args = append(args, params.ToUsername)
		whereClauses = append(whereClauses, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM posts rp
			JOIN users ru ON ru.id = rp.author_id
			WHERE rp.id = p.reply_to_id AND ru.username ILIKE $%d
		)`, argIdx))
		argIdx++
	}

	// mentions
	if params.Mentions != "" {
		args = append(args, params.Mentions)
		whereClauses = append(whereClauses, fmt.Sprintf(`p.content ILIKE '%%' || $%d || '%%'`, argIdx))
		argIdx++
	}

	// has_media or tab == "media"
	if params.HasMedia || params.Tab == "media" {
		whereClauses = append(whereClauses, "EXISTS (SELECT 1 FROM post_media pm WHERE pm.post_id = p.id)")
	}

	// min_likes
	if params.MinLikes > 0 {
		args = append(args, params.MinLikes)
		whereClauses = append(whereClauses, fmt.Sprintf("p.likes_count >= $%d", argIdx))
		argIdx++
	}

	// since
	if params.SinceDate != nil {
		args = append(args, *params.SinceDate)
		whereClauses = append(whereClauses, fmt.Sprintf("p.created_at >= $%d", argIdx))
		argIdx++
	}

	// until
	if params.UntilDate != nil {
		args = append(args, *params.UntilDate)
		whereClauses = append(whereClauses, fmt.Sprintf("p.created_at <= $%d", argIdx))
		argIdx++
	}

	// Keyset cursor pagination
	if params.Cursor != "" {
		if feedCursor, err := recommendation.DecodeCursor(params.Cursor); err == nil && feedCursor != nil {
			args = append(args, feedCursor.CreatedAt, feedCursor.PostID)
			whereClauses = append(whereClauses, fmt.Sprintf("(p.created_at, p.id) < ($%d, $%d)", argIdx, argIdx+1))
			argIdx += 2
		}
	}

	// Order by
	var orderClause string
	if params.Tab == "latest" {
		orderClause = "ORDER BY p.created_at DESC, p.id DESC"
	} else if params.Tab == "top" && params.Query != "" {
		orderClause = fmt.Sprintf(`ORDER BY (
			COALESCE(ts_rank_cd(p.search_vector, plainto_tsquery('simple', $%d), 32), 0.0) * 5.0
			+ LN(1.0 + p.likes_count) * 1.5
			+ LN(1.0 + p.reposts_count) * 2.0
			+ EXP(-EXTRACT(EPOCH FROM (NOW() - p.created_at)) / 86400.0) * 3.0
			+ CASE WHEN u.is_verified THEN 1.5 ELSE 0.0 END
		) DESC, p.created_at DESC, p.id DESC`, textArgIdx)
	} else {
		orderClause = "ORDER BY p.created_at DESC, p.id DESC"
	}

	args = append(args, limit+1)
	limitArgIdx := argIdx

	sqlQuery := fmt.Sprintf(`
		SELECT p.id, p.author_id, p.content, p.reply_to_id, p.quote_post_id,
		       p.replies_count, p.reposts_count, p.likes_count, p.views_count, p.created_at, p.updated_at,
		       u.username, u.is_verified, u.verified_type,
		       pr.display_name, pr.avatar_url,
		       COALESCE((SELECT TRUE FROM post_likes pl WHERE pl.post_id = p.id AND pl.user_id = $1), FALSE) as is_liked,
		       COALESCE((SELECT TRUE FROM post_reposts prp WHERE prp.post_id = p.id AND prp.user_id = $1), FALSE) as is_reposted,
		       COALESCE((SELECT TRUE FROM post_bookmarks pb WHERE pb.post_id = p.id AND pb.user_id = $1), FALSE) as is_bookmarked
		FROM posts p
		JOIN users u ON u.id = p.author_id
		JOIN user_profiles pr ON pr.user_id = p.author_id
		WHERE %s
		%s
		LIMIT $%d;
	`, strings.Join(whereClauses, " AND "), orderClause, limitArgIdx)

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var posts []domain.Post
	for rows.Next() {
		var p domain.Post
		var author domain.User
		err := rows.Scan(
			&p.ID, &p.AuthorID, &p.Content, &p.ReplyToID, &p.QuotePostID,
			&p.RepliesCount, &p.RepostsCount, &p.LikesCount, &p.ViewsCount, &p.CreatedAt, &p.UpdatedAt,
			&author.Username, &author.IsVerified, &author.VerifiedType,
			&author.DisplayName, &author.AvatarURL,
			&p.IsLiked, &p.IsReposted, &p.IsBookmarked,
		)
		if err != nil {
			return nil, "", err
		}
		author.ID = p.AuthorID
		p.Author = &author
		posts = append(posts, p)
	}

	var nextCursor string
	if len(posts) > limit {
		last := posts[limit-1]
		nextCursor = recommendation.EncodeCursor(last.CreatedAt, last.ID)
		posts = posts[:limit]
	}

	_ = r.attachMediaToPosts(ctx, posts)

	return posts, nextCursor, nil
}

func (r *PostgresRepository) SearchPosts(ctx context.Context, query string, viewerID *uuid.UUID, cursor string, limit int) ([]domain.Post, string, error) {
	return r.SearchPostsAdvanced(ctx, domain.SearchFilterParams{
		Query:  query,
		Tab:    "top",
		Cursor: cursor,
		Limit:  limit,
	}, viewerID)
}

func (r *PostgresRepository) SearchUsersAdvanced(ctx context.Context, query string, viewerID *uuid.UUID, cursor string, limit int) ([]domain.User, string, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	var vID interface{} = uuid.Nil
	if viewerID != nil {
		vID = *viewerID
	}

	trimmed := strings.TrimSpace(query)
	sqlQuery := `
		SELECT u.id, u.username, u.email, u.role, u.is_verified, u.verified_type, u.is_private, u.created_at, u.updated_at,
		       p.display_name, p.bio, p.avatar_url, p.banner_url, p.location, p.website,
		       p.followers_count, p.following_count, p.posts_count,
		       COALESCE((SELECT TRUE FROM follows f WHERE f.follower_id = $2 AND f.following_id = u.id), FALSE) as is_following
		FROM users u
		JOIN user_profiles p ON p.user_id = u.id
		WHERE (
			u.username ILIKE '%' || $1 || '%'
			OR p.display_name ILIKE '%' || $1 || '%'
			OR similarity(u.username, $1) > 0.2
			OR similarity(p.display_name, $1) > 0.2
		) AND u.deleted_at IS NULL
		ORDER BY (
			CASE WHEN LOWER(u.username) = LOWER($1) THEN 10.0 ELSE 0.0 END
			+ similarity(u.username, $1) * 4.0
			+ similarity(p.display_name, $1) * 2.5
			+ LN(1.0 + p.followers_count) * 0.5
			+ CASE WHEN u.is_verified THEN 1.0 ELSE 0.0 END
		) DESC
		LIMIT $3;
	`
	rows, err := r.db.QueryContext(ctx, sqlQuery, trimmed, vID, limit+1)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		err := rows.Scan(
			&u.ID, &u.Username, &u.Email, &u.Role, &u.IsVerified, &u.VerifiedType, &u.IsPrivate, &u.CreatedAt, &u.UpdatedAt,
			&u.DisplayName, &u.Bio, &u.AvatarURL, &u.BannerURL, &u.Location, &u.Website,
			&u.FollowersCount, &u.FollowingCount, &u.PostsCount,
			&u.IsFollowing,
		)
		if err != nil {
			return nil, "", err
		}
		users = append(users, u)
	}

	var nextCursor string
	if len(users) > limit {
		nextCursor = users[limit-1].ID.String()
		users = users[:limit]
	}

	return users, nextCursor, nil
}

func (r *PostgresRepository) SearchUsers(ctx context.Context, query string, viewerID *uuid.UUID, limit int) ([]domain.User, error) {
	users, _, err := r.SearchUsersAdvanced(ctx, query, viewerID, "", limit)
	return users, err
}

func (r *PostgresRepository) GetSearchSuggestions(ctx context.Context, query string, viewerID *uuid.UUID, limit int) ([]string, []domain.User, error) {
	if limit <= 0 || limit > 10 {
		limit = 5
	}
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return nil, nil, nil
	}

	users, _, _ := r.SearchUsersAdvanced(ctx, trimmed, viewerID, "", limit)

	rows, err := r.db.QueryContext(ctx, `
		SELECT tag FROM hashtags
		WHERE tag ILIKE '%' || $1 || '%'
		ORDER BY posts_count DESC
		LIMIT $2;
	`, strings.TrimPrefix(trimmed, "#"), limit)

	var queries []string
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var tag string
			if err := rows.Scan(&tag); err == nil {
				queries = append(queries, "#"+tag)
			}
		}
	}

	return queries, users, nil
}

func (r *PostgresRepository) GetRecentSearches(ctx context.Context, userID uuid.UUID, limit int) ([]domain.RecentSearch, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	query := `
		SELECT id, user_id, query, searched_at
		FROM recent_searches
		WHERE user_id = $1
		ORDER BY searched_at DESC
		LIMIT $2;
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.RecentSearch
	for rows.Next() {
		var s domain.RecentSearch
		if err := rows.Scan(&s.ID, &s.UserID, &s.Query, &s.SearchedAt); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

func (r *PostgresRepository) AddRecentSearch(ctx context.Context, userID uuid.UUID, query string) error {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return nil
	}
	sqlQuery := `
		INSERT INTO recent_searches (user_id, query, searched_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id, query)
		DO UPDATE SET searched_at = NOW();
	`
	_, err := r.db.ExecContext(ctx, sqlQuery, userID, trimmed)
	return err
}

func (r *PostgresRepository) DeleteRecentSearch(ctx context.Context, userID uuid.UUID, searchID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM recent_searches WHERE id = $1 AND user_id = $2;`, searchID, userID)
	return err
}

func (r *PostgresRepository) ClearRecentSearches(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM recent_searches WHERE user_id = $1;`, userID)
	return err
}

func (r *PostgresRepository) CreateNotification(ctx context.Context, n *domain.Notification) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	query := `
		INSERT INTO notifications (id, recipient_id, actor_id, type, post_id, is_read, entity_type, entity_id, created_at)
		SELECT $1, $2, $3, $4, $5, $6, $7, $8, NOW()
		WHERE NOT EXISTS (
			SELECT 1 FROM notifications
			WHERE recipient_id = $2
			  AND actor_id = $3
			  AND type = $4
			  AND (
				(post_id IS NULL AND $5::uuid IS NULL) OR
				(post_id = $5::uuid)
			  )
			  AND created_at > NOW() - INTERVAL '30 seconds'
		);
	`
	res, err := r.db.ExecContext(ctx, query, n.ID, n.RecipientID, n.ActorID, n.Type, n.PostID, n.IsRead, n.EntityType, n.EntityID)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return errors.New("duplicate_notification_skipped")
	}
	return nil
}

func (r *PostgresRepository) GetNotifications(ctx context.Context, userID uuid.UUID, cursor string, limit int) ([]domain.Notification, string, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	query := `
		SELECT n.id, n.recipient_id, n.actor_id, n.type, n.post_id, n.is_read, n.created_at,
		       u.username, u.is_verified, u.verified_type,
		       p.display_name, p.avatar_url,
		       n.entity_type, n.entity_id,
		       po.content
		FROM notifications n
		JOIN users u ON u.id = n.actor_id
		JOIN user_profiles p ON p.user_id = n.actor_id
		LEFT JOIN posts po ON po.id = n.post_id
		WHERE n.recipient_id = $1
		ORDER BY n.created_at DESC
		LIMIT $2;
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var notifs []domain.Notification
	for rows.Next() {
		var n domain.Notification
		var actor domain.User
		var entityType sql.NullString
		var entityID uuid.NullUUID
		var postContent sql.NullString
		err := rows.Scan(
			&n.ID, &n.RecipientID, &n.ActorID, &n.Type, &n.PostID, &n.IsRead, &n.CreatedAt,
			&actor.Username, &actor.IsVerified, &actor.VerifiedType,
			&actor.DisplayName, &actor.AvatarURL,
			&entityType, &entityID,
			&postContent,
		)
		if err != nil {
			return nil, "", err
		}
		if entityType.Valid {
			n.EntityType = entityType.String
		}
		if entityID.Valid {
			n.EntityID = &entityID.UUID
		}
		if n.PostID != nil && postContent.Valid {
			n.Post = &domain.Post{
				ID:      *n.PostID,
				Content: postContent.String,
			}
		}
		actor.ID = n.ActorID
		n.Actor = &actor
		notifs = append(notifs, n)
	}
	return notifs, "", nil
}

func (r *PostgresRepository) MarkNotificationsRead(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE notifications SET is_read = TRUE WHERE recipient_id = $1;`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *PostgresRepository) CreateMediaObject(ctx context.Context, m *domain.MediaObject) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	query := `
		INSERT INTO media_objects (id, uploader_id, storage_key, url, media_type, mime_type, file_size_bytes, width, height, duration_seconds, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW());
	`
	_, err := r.db.ExecContext(ctx, query, m.ID, m.UploaderID, m.StorageKey, m.URL, m.MediaType, m.MimeType, m.FileSizeBytes, m.Width, m.Height, m.DurationSeconds)
	return err
}

func (r *PostgresRepository) GetMediaObjectByID(ctx context.Context, id uuid.UUID) (*domain.MediaObject, error) {
	query := `
		SELECT id, uploader_id, storage_key, url, media_type, mime_type, file_size_bytes, width, height, duration_seconds, created_at
		FROM media_objects
		WHERE id = $1;
	`
	var m domain.MediaObject
	err := r.db.QueryRowContext(ctx, query, id).Scan(&m.ID, &m.UploaderID, &m.StorageKey, &m.URL, &m.MediaType, &m.MimeType, &m.FileSizeBytes, &m.Width, &m.Height, &m.DurationSeconds, &m.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &m, err
}

// ==========================================
// SUBSCRIPTIONS & ENTITLEMENTS (SECTION 6A)
// ==========================================

func (r *PostgresRepository) GetSubscriptionPlans(ctx context.Context) ([]domain.SubscriptionPlan, error) {
	query := `
		SELECT id, name, description, monthly_price_cents, annual_price_cents, currency, max_post_length,
		       can_get_verified_badge, can_edit_posts, can_undo_posts, max_video_minutes, max_media_count,
		       reduced_ads, is_active, created_at, updated_at
		FROM subscription_plans
		WHERE is_active = TRUE
		ORDER BY monthly_price_cents ASC;
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []domain.SubscriptionPlan
	for rows.Next() {
		var p domain.SubscriptionPlan
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.MonthlyPriceCents, &p.AnnualPriceCents, &p.Currency, &p.MaxPostLength,
			&p.CanGetVerifiedBadge, &p.CanEditPosts, &p.CanUndoPosts, &p.MaxVideoMinutes, &p.MaxMediaCount,
			&p.ReducedAds, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}
	return plans, rows.Err()
}

func (r *PostgresRepository) GetSubscriptionPlanByID(ctx context.Context, id string) (*domain.SubscriptionPlan, error) {
	query := `
		SELECT id, name, description, monthly_price_cents, annual_price_cents, currency, max_post_length,
		       can_get_verified_badge, can_edit_posts, can_undo_posts, max_video_minutes, max_media_count,
		       reduced_ads, is_active, created_at, updated_at
		FROM subscription_plans
		WHERE id = $1;
	`
	var p domain.SubscriptionPlan
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.MonthlyPriceCents, &p.AnnualPriceCents, &p.Currency, &p.MaxPostLength,
		&p.CanGetVerifiedBadge, &p.CanEditPosts, &p.CanUndoPosts, &p.MaxVideoMinutes, &p.MaxMediaCount,
		&p.ReducedAds, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return &p, err
}

func (r *PostgresRepository) GetActiveSubscription(ctx context.Context, userID uuid.UUID) (*domain.Subscription, error) {
	query := `
		SELECT s.id, s.user_id, s.plan_id, s.billing_cycle, s.status, s.current_period_start,
		       s.current_period_end, s.cancel_at_period_end, s.created_at, s.updated_at,
		       p.name, p.description, p.monthly_price_cents, p.annual_price_cents, p.currency,
		       p.max_post_length, p.can_get_verified_badge, p.can_edit_posts, p.can_undo_posts,
		       p.max_video_minutes, p.max_media_count, p.reduced_ads
		FROM subscriptions s
		JOIN subscription_plans p ON s.plan_id = p.id
		WHERE s.user_id = $1 AND s.status = 'active' AND s.current_period_end > NOW()
		ORDER BY s.created_at DESC
		LIMIT 1;
	`
	var s domain.Subscription
	var p domain.SubscriptionPlan
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&s.ID, &s.UserID, &s.PlanID, &s.BillingCycle, &s.Status, &s.CurrentPeriodStart,
		&s.CurrentPeriodEnd, &s.CancelAtPeriodEnd, &s.CreatedAt, &s.UpdatedAt,
		&p.Name, &p.Description, &p.MonthlyPriceCents, &p.AnnualPriceCents, &p.Currency,
		&p.MaxPostLength, &p.CanGetVerifiedBadge, &p.CanEditPosts, &p.CanUndoPosts,
		&p.MaxVideoMinutes, &p.MaxMediaCount, &p.ReducedAds,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // No active subscription
	}
	if err != nil {
		return nil, err
	}
	p.ID = s.PlanID
	s.Plan = &p
	return &s, nil
}

func (r *PostgresRepository) CreateSubscription(ctx context.Context, sub *domain.Subscription) error {
	if sub.ID == uuid.Nil {
		sub.ID = uuid.New()
	}
	query := `
		INSERT INTO subscriptions (id, user_id, plan_id, billing_cycle, status, current_period_start, current_period_end, cancel_at_period_end, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING created_at, updated_at;
	`
	return r.db.QueryRowContext(ctx, query, sub.ID, sub.UserID, sub.PlanID, sub.BillingCycle, sub.Status, sub.CurrentPeriodStart, sub.CurrentPeriodEnd, sub.CancelAtPeriodEnd).Scan(&sub.CreatedAt, &sub.UpdatedAt)
}

func (r *PostgresRepository) CancelSubscription(ctx context.Context, subID, userID uuid.UUID) error {
	query := `
		UPDATE subscriptions
		SET cancel_at_period_end = TRUE, updated_at = NOW()
		WHERE id = $1 AND user_id = $2;
	`
	res, err := r.db.ExecContext(ctx, query, subID, userID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) GetUserEntitlements(ctx context.Context, userID uuid.UUID) (*domain.Entitlements, error) {
	sub, err := r.GetActiveSubscription(ctx, userID)
	if err != nil {
		return nil, err
	}

	ent := &domain.Entitlements{
		PlanID:           "free",
		MaxPostLength:    280,
		HasVerifiedBadge: false,
		BadgeType:        "none",
		CanEditPosts:     false,
		CanUndoPosts:     false,
		MaxVideoMinutes:  2,
		MaxMediaCount:    4,
		ReducedAds:       false,
	}

	if sub != nil && sub.Plan != nil {
		ent.PlanID = sub.PlanID
		ent.MaxPostLength = sub.Plan.MaxPostLength
		ent.HasVerifiedBadge = sub.Plan.CanGetVerifiedBadge
		if sub.Plan.CanGetVerifiedBadge {
			ent.BadgeType = "blue"
		}
		ent.CanEditPosts = sub.Plan.CanEditPosts
		ent.CanUndoPosts = sub.Plan.CanUndoPosts
		ent.MaxVideoMinutes = sub.Plan.MaxVideoMinutes
		ent.MaxMediaCount = sub.Plan.MaxMediaCount
		ent.ReducedAds = sub.Plan.ReducedAds
	}

	// Also check if user has custom badge (e.g. staff/gold) in users table
	var isVerified bool
	var verifiedType string
	qUser := `SELECT is_verified, verified_type FROM users WHERE id = $1;`
	if err := r.db.QueryRowContext(ctx, qUser, userID).Scan(&isVerified, &verifiedType); err == nil && isVerified {
		ent.HasVerifiedBadge = true
		ent.BadgeType = verifiedType
	}

	return ent, nil
}

func (r *PostgresRepository) CreatePaymentOrder(ctx context.Context, order *domain.PaymentOrder) error {
	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}
	metaJSON, _ := json.Marshal(order.Metadata)
	if len(metaJSON) == 0 || string(metaJSON) == "null" {
		metaJSON = []byte("{}")
	}
	query := `
		INSERT INTO payment_orders (id, user_id, plan_id, billing_cycle, amount_cents, currency, provider, provider_order_id, provider_invoice_url, status, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb, NOW(), NOW())
		RETURNING created_at, updated_at;
	`
	return r.db.QueryRowContext(ctx, query, order.ID, order.UserID, order.PlanID, order.BillingCycle, order.AmountCents, order.Currency, order.Provider, order.ProviderOrderID, order.ProviderInvoiceURL, order.Status, string(metaJSON)).Scan(&order.CreatedAt, &order.UpdatedAt)
}

func (r *PostgresRepository) GetPaymentOrderByID(ctx context.Context, orderID uuid.UUID) (*domain.PaymentOrder, error) {
	query := `
		SELECT id, user_id, plan_id, billing_cycle, amount_cents, currency, provider, provider_order_id, provider_invoice_url, status, metadata, created_at, updated_at
		FROM payment_orders
		WHERE id = $1;
	`
	var o domain.PaymentOrder
	var provOrderID, invoiceURL sql.NullString
	var metaBytes []byte
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&o.ID, &o.UserID, &o.PlanID, &o.BillingCycle, &o.AmountCents, &o.Currency, &o.Provider,
		&provOrderID, &invoiceURL, &o.Status, &metaBytes, &o.CreatedAt, &o.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if provOrderID.Valid {
		o.ProviderOrderID = provOrderID.String
	}
	if invoiceURL.Valid {
		o.ProviderInvoiceURL = invoiceURL.String
	}
	if len(metaBytes) > 0 {
		_ = json.Unmarshal(metaBytes, &o.Metadata)
	}
	return &o, err
}

func (r *PostgresRepository) UpdatePaymentOrderStatus(ctx context.Context, orderID uuid.UUID, status, invoiceURL string) error {
	query := `
		UPDATE payment_orders
		SET status = $2, provider_invoice_url = COALESCE(NULLIF($3, ''), provider_invoice_url), updated_at = NOW()
		WHERE id = $1;
	`
	_, err := r.db.ExecContext(ctx, query, orderID, status, invoiceURL)
	return err
}

func (r *PostgresRepository) GetBillingHistory(ctx context.Context, userID uuid.UUID) ([]domain.PaymentOrder, []domain.Invoice, error) {
	ordersQuery := `
		SELECT id, user_id, plan_id, billing_cycle, amount_cents, currency, provider, provider_order_id, provider_invoice_url, status, created_at, updated_at
		FROM payment_orders
		WHERE user_id = $1
		ORDER BY created_at DESC;
	`
	rows, err := r.db.QueryContext(ctx, ordersQuery, userID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var orders []domain.PaymentOrder
	for rows.Next() {
		var o domain.PaymentOrder
		var pID, invURL sql.NullString
		if err := rows.Scan(&o.ID, &o.UserID, &o.PlanID, &o.BillingCycle, &o.AmountCents, &o.Currency, &o.Provider, &pID, &invURL, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, nil, err
		}
		if pID.Valid {
			o.ProviderOrderID = pID.String
		}
		if invURL.Valid {
			o.ProviderInvoiceURL = invURL.String
		}
		orders = append(orders, o)
	}

	invQuery := `
		SELECT id, order_id, user_id, amount_cents, currency, status, pdf_url, created_at
		FROM invoices
		WHERE user_id = $1
		ORDER BY created_at DESC;
	`
	invRows, err := r.db.QueryContext(ctx, invQuery, userID)
	if err != nil {
		return orders, nil, nil
	}
	defer invRows.Close()

	var invoices []domain.Invoice
	for invRows.Next() {
		var inv domain.Invoice
		var pdfURL sql.NullString
		if err := invRows.Scan(&inv.ID, &inv.OrderID, &inv.UserID, &inv.AmountCents, &inv.Currency, &inv.Status, &pdfURL, &inv.CreatedAt); err == nil {
			if pdfURL.Valid {
				inv.PDFURL = pdfURL.String
			}
			invoices = append(invoices, inv)
		}
	}

	return orders, invoices, nil
}

func (r *PostgresRepository) CreateInvoice(ctx context.Context, inv *domain.Invoice) error {
	if inv.ID == uuid.Nil {
		inv.ID = uuid.New()
	}
	query := `
		INSERT INTO invoices (id, order_id, user_id, amount_cents, currency, status, pdf_url, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING created_at;
	`
	return r.db.QueryRowContext(ctx, query, inv.ID, inv.OrderID, inv.UserID, inv.AmountCents, inv.Currency, inv.Status, inv.PDFURL).Scan(&inv.CreatedAt)
}

func (r *PostgresRepository) GetUserBadges(ctx context.Context, userID uuid.UUID) ([]domain.VerificationBadge, error) {
	query := `
		SELECT id, user_id, badge_type, display_label, description, color_token, status, expires_at, created_at, updated_at
		FROM verification_badges
		WHERE user_id = $1 AND status = 'approved'
		ORDER BY created_at DESC;
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var badges []domain.VerificationBadge
	for rows.Next() {
		var b domain.VerificationBadge
		if err := rows.Scan(&b.ID, &b.UserID, &b.BadgeType, &b.DisplayLabel, &b.Description, &b.ColorToken, &b.Status, &b.ExpiresAt, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		badges = append(badges, b)
	}
	return badges, rows.Err()
}

func (r *PostgresRepository) AssignVerificationBadge(ctx context.Context, badge *domain.VerificationBadge) error {
	if badge.ID == uuid.Nil {
		badge.ID = uuid.New()
	}
	query := `
		INSERT INTO verification_badges (id, user_id, badge_type, display_label, description, color_token, status, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		ON CONFLICT (user_id, badge_type) DO UPDATE
		SET status = EXCLUDED.status, updated_at = NOW()
		RETURNING created_at, updated_at;
	`
	return r.db.QueryRowContext(ctx, query, badge.ID, badge.UserID, badge.BadgeType, badge.DisplayLabel, badge.Description, badge.ColorToken, badge.Status, badge.ExpiresAt).Scan(&badge.CreatedAt, &badge.UpdatedAt)
}

func (r *PostgresRepository) CreateVerificationApplication(ctx context.Context, app *domain.VerificationApplication) error {
	if app.ID == uuid.Nil {
		app.ID = uuid.New()
	}
	query := `
		INSERT INTO verification_applications (id, user_id, badge_type, category, document_url, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING created_at;
	`
	return r.db.QueryRowContext(ctx, query, app.ID, app.UserID, app.BadgeType, app.Category, app.DocumentURL, app.Status).Scan(&app.CreatedAt)
}

func (r *PostgresRepository) GetVerificationApplication(ctx context.Context, userID uuid.UUID) (*domain.VerificationApplication, error) {
	query := `
		SELECT id, user_id, badge_type, category, document_url, status, review_notes, reviewed_by, reviewed_at, created_at
		FROM verification_applications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1;
	`
	var app domain.VerificationApplication
	var docURL, notes sql.NullString
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&app.ID, &app.UserID, &app.BadgeType, &app.Category, &docURL, &app.Status, &notes,
		&app.ReviewedBy, &app.ReviewedAt, &app.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if docURL.Valid {
		app.DocumentURL = docURL.String
	}
	if notes.Valid {
		app.ReviewNotes = notes.String
	}
	return &app, err
}

func (r *PostgresRepository) UpdateUserVerification(ctx context.Context, userID uuid.UUID, isVerified bool, verifiedType string) error {
	query := `
		UPDATE users
		SET is_verified = $2, verified_type = $3, updated_at = NOW()
		WHERE id = $1;
	`
	_, err := r.db.ExecContext(ctx, query, userID, isVerified, verifiedType)
	return err
}

// ==========================================
// SOCIAL GRAPH IMPLEMENTATION
// ==========================================

func (r *PostgresRepository) GetFollowers(ctx context.Context, userID, viewerID uuid.UUID, limit int, cursor string) ([]domain.User, string, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	query := `
		SELECT u.id, u.username, u.email, u.role, u.is_verified, u.verified_type, u.is_private, u.created_at, u.updated_at,
		       p.display_name, p.bio, p.avatar_url, p.banner_url, p.location, p.website,
		       p.followers_count, p.following_count, p.posts_count,
		       COALESCE((SELECT TRUE FROM follows f WHERE f.follower_id = $2 AND f.following_id = u.id), FALSE) as is_following,
		       COALESCE((SELECT TRUE FROM follows f WHERE f.follower_id = u.id AND f.following_id = $2), FALSE) as is_followed_by
		FROM follows fl
		JOIN users u ON u.id = fl.follower_id
		JOIN user_profiles p ON p.user_id = u.id
		WHERE fl.following_id = $1 AND u.deleted_at IS NULL
		ORDER BY fl.created_at DESC
		LIMIT $3;
	`

	rows, err := r.db.QueryContext(ctx, query, userID, viewerID, limit)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(
			&u.ID, &u.Username, &u.Email, &u.Role, &u.IsVerified, &u.VerifiedType, &u.IsPrivate, &u.CreatedAt, &u.UpdatedAt,
			&u.DisplayName, &u.Bio, &u.AvatarURL, &u.BannerURL, &u.Location, &u.Website,
			&u.FollowersCount, &u.FollowingCount, &u.PostsCount,
			&u.IsFollowing, &u.IsFollowedBy,
		); err != nil {
			return nil, "", err
		}
		users = append(users, u)
	}

	return users, "", nil
}

func (r *PostgresRepository) GetFollowing(ctx context.Context, userID, viewerID uuid.UUID, limit int, cursor string) ([]domain.User, string, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	query := `
		SELECT u.id, u.username, u.email, u.role, u.is_verified, u.verified_type, u.is_private, u.created_at, u.updated_at,
		       p.display_name, p.bio, p.avatar_url, p.banner_url, p.location, p.website,
		       p.followers_count, p.following_count, p.posts_count,
		       COALESCE((SELECT TRUE FROM follows f WHERE f.follower_id = $2 AND f.following_id = u.id), FALSE) as is_following,
		       COALESCE((SELECT TRUE FROM follows f WHERE f.follower_id = u.id AND f.following_id = $2), FALSE) as is_followed_by
		FROM follows fl
		JOIN users u ON u.id = fl.following_id
		JOIN user_profiles p ON p.user_id = u.id
		WHERE fl.follower_id = $1 AND u.deleted_at IS NULL
		ORDER BY fl.created_at DESC
		LIMIT $3;
	`

	rows, err := r.db.QueryContext(ctx, query, userID, viewerID, limit)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(
			&u.ID, &u.Username, &u.Email, &u.Role, &u.IsVerified, &u.VerifiedType, &u.IsPrivate, &u.CreatedAt, &u.UpdatedAt,
			&u.DisplayName, &u.Bio, &u.AvatarURL, &u.BannerURL, &u.Location, &u.Website,
			&u.FollowersCount, &u.FollowingCount, &u.PostsCount,
			&u.IsFollowing, &u.IsFollowedBy,
		); err != nil {
			return nil, "", err
		}
		users = append(users, u)
	}

	return users, "", nil
}

func (r *PostgresRepository) GetMutualConnections(ctx context.Context, userID, viewerID uuid.UUID, limit int) ([]domain.User, error) {
	if limit <= 0 || limit > 20 {
		limit = 5
	}

	query := `
		SELECT u.id, u.username, u.email, u.role, u.is_verified, u.verified_type, u.is_private, u.created_at, u.updated_at,
		       p.display_name, p.bio, p.avatar_url, p.banner_url, p.location, p.website,
		       p.followers_count, p.following_count, p.posts_count,
		       TRUE as is_following, FALSE as is_followed_by
		FROM follows f_viewer
		JOIN follows f_target ON f_viewer.following_id = f_target.following_id
		JOIN users u ON u.id = f_viewer.following_id
		JOIN user_profiles p ON p.user_id = u.id
		WHERE f_viewer.follower_id = $1 AND f_target.follower_id = $2 AND u.deleted_at IS NULL
		LIMIT $3;
	`

	rows, err := r.db.QueryContext(ctx, query, viewerID, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(
			&u.ID, &u.Username, &u.Email, &u.Role, &u.IsVerified, &u.VerifiedType, &u.IsPrivate, &u.CreatedAt, &u.UpdatedAt,
			&u.DisplayName, &u.Bio, &u.AvatarURL, &u.BannerURL, &u.Location, &u.Website,
			&u.FollowersCount, &u.FollowingCount, &u.PostsCount,
			&u.IsFollowing, &u.IsFollowedBy,
		); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

func (r *PostgresRepository) GetSocialRelationship(ctx context.Context, viewerID, targetID uuid.UUID) (*domain.SocialRelationship, error) {
	rel := &domain.SocialRelationship{
		TargetUserID: targetID,
	}

	q := `
		SELECT 
			u.is_private,
			COALESCE((SELECT TRUE FROM follows WHERE follower_id = $1 AND following_id = $2), FALSE),
			COALESCE((SELECT TRUE FROM follows WHERE follower_id = $2 AND following_id = $1), FALSE),
			COALESCE((SELECT TRUE FROM blocks WHERE blocker_id = $1 AND blocked_id = $2), FALSE),
			COALESCE((SELECT TRUE FROM blocks WHERE blocker_id = $2 AND blocked_id = $1), FALSE),
			COALESCE((SELECT TRUE FROM mutes WHERE muter_id = $1 AND muted_id = $2), FALSE),
			COALESCE((SELECT TRUE FROM follow_requests WHERE requester_id = $1 AND target_id = $2 AND status = 'pending'), FALSE)
		FROM users u
		WHERE u.id = $2 AND u.deleted_at IS NULL;
	`
	err := r.db.QueryRowContext(ctx, q, viewerID, targetID).Scan(
		&rel.IsPrivate,
		&rel.IsFollowing,
		&rel.IsFollowedBy,
		&rel.IsBlocked,
		&rel.IsBlockedBy,
		&rel.IsMuted,
		&rel.HasPendingFollowRequest,
	)
	if err != nil {
		return nil, err
	}
	rel.IsMutual = rel.IsFollowing && rel.IsFollowedBy
	return rel, nil
}

func (r *PostgresRepository) BlockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	if blockerID == blockedID {
		return errors.New("cannot block yourself")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert block
	_, err = tx.ExecContext(ctx, `INSERT INTO blocks (blocker_id, blocked_id, created_at) VALUES ($1, $2, NOW()) ON CONFLICT DO NOTHING;`, blockerID, blockedID)
	if err != nil {
		return err
	}

	// Remove mutual follows
	res1, _ := tx.ExecContext(ctx, `DELETE FROM follows WHERE follower_id = $1 AND following_id = $2;`, blockerID, blockedID)
	r1, _ := res1.RowsAffected()
	if r1 > 0 {
		_, _ = tx.ExecContext(ctx, `UPDATE user_profiles SET following_count = GREATEST(following_count - 1, 0) WHERE user_id = $1;`, blockerID)
		_, _ = tx.ExecContext(ctx, `UPDATE user_profiles SET followers_count = GREATEST(followers_count - 1, 0) WHERE user_id = $1;`, blockedID)
	}

	res2, _ := tx.ExecContext(ctx, `DELETE FROM follows WHERE follower_id = $1 AND following_id = $2;`, blockedID, blockerID)
	r2, _ := res2.RowsAffected()
	if r2 > 0 {
		_, _ = tx.ExecContext(ctx, `UPDATE user_profiles SET following_count = GREATEST(following_count - 1, 0) WHERE user_id = $1;`, blockedID)
		_, _ = tx.ExecContext(ctx, `UPDATE user_profiles SET followers_count = GREATEST(followers_count - 1, 0) WHERE user_id = $1;`, blockerID)
	}

	// Cancel any pending follow requests
	_, _ = tx.ExecContext(ctx, `DELETE FROM follow_requests WHERE (requester_id = $1 AND target_id = $2) OR (requester_id = $2 AND target_id = $1);`, blockerID, blockedID)

	return tx.Commit()
}

func (r *PostgresRepository) UnblockUser(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM blocks WHERE blocker_id = $1 AND blocked_id = $2;`, blockerID, blockedID)
	return err
}

func (r *PostgresRepository) MuteUser(ctx context.Context, muterID, mutedID uuid.UUID) error {
	if muterID == mutedID {
		return errors.New("cannot mute yourself")
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO mutes (muter_id, muted_id, created_at) VALUES ($1, $2, NOW()) ON CONFLICT DO NOTHING;`, muterID, mutedID)
	return err
}

func (r *PostgresRepository) UnmuteUser(ctx context.Context, muterID, mutedID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM mutes WHERE muter_id = $1 AND muted_id = $2;`, muterID, mutedID)
	return err
}

func (r *PostgresRepository) IsBlocked(ctx context.Context, userA, userB uuid.UUID) (bool, error) {
	var blocked bool
	q := `SELECT EXISTS(SELECT 1 FROM blocks WHERE (blocker_id = $1 AND blocked_id = $2) OR (blocker_id = $2 AND blocked_id = $1));`
	err := r.db.QueryRowContext(ctx, q, userA, userB).Scan(&blocked)
	return blocked, err
}

// ==========================================
// FOLLOW REQUESTS
// ==========================================

func (r *PostgresRepository) CreateFollowRequest(ctx context.Context, req *domain.FollowRequest) error {
	if req.ID == uuid.Nil {
		req.ID = uuid.New()
	}
	q := `
		INSERT INTO follow_requests (id, requester_id, target_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, 'pending', NOW(), NOW())
		ON CONFLICT (requester_id, target_id) DO UPDATE SET status = 'pending', updated_at = NOW()
		RETURNING created_at, updated_at;
	`
	return r.db.QueryRowContext(ctx, q, req.ID, req.RequesterID, req.TargetID).Scan(&req.CreatedAt, &req.UpdatedAt)
}

func (r *PostgresRepository) GetFollowRequests(ctx context.Context, targetID uuid.UUID) ([]domain.FollowRequest, error) {
	q := `
		SELECT fr.id, fr.requester_id, fr.target_id, fr.status, fr.created_at, fr.updated_at,
		       u.username, p.display_name, p.avatar_url, u.is_verified, u.verified_type
		FROM follow_requests fr
		JOIN users u ON u.id = fr.requester_id
		JOIN user_profiles p ON p.user_id = u.id
		WHERE fr.target_id = $1 AND fr.status = 'pending'
		ORDER BY fr.created_at DESC;
	`
	rows, err := r.db.QueryContext(ctx, q, targetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reqs []domain.FollowRequest
	for rows.Next() {
		var req domain.FollowRequest
		var u domain.User
		if err := rows.Scan(
			&req.ID, &req.RequesterID, &req.TargetID, &req.Status, &req.CreatedAt, &req.UpdatedAt,
			&u.Username, &u.DisplayName, &u.AvatarURL, &u.IsVerified, &u.VerifiedType,
		); err != nil {
			return nil, err
		}
		u.ID = req.RequesterID
		req.Requester = &u
		reqs = append(reqs, req)
	}
	return reqs, nil
}

func (r *PostgresRepository) RespondFollowRequest(ctx context.Context, requestID, targetID uuid.UUID, accept bool) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var req domain.FollowRequest
	q := `SELECT id, requester_id, target_id, status FROM follow_requests WHERE id = $1 AND target_id = $2;`
	if err := tx.QueryRowContext(ctx, q, requestID, targetID).Scan(&req.ID, &req.RequesterID, &req.TargetID, &req.Status); err != nil {
		return err
	}

	newStatus := "rejected"
	if accept {
		newStatus = "accepted"
		// Insert follow
		res, err := tx.ExecContext(ctx, `INSERT INTO follows (follower_id, following_id, created_at) VALUES ($1, $2, NOW()) ON CONFLICT DO NOTHING;`, req.RequesterID, req.TargetID)
		if err != nil {
			return err
		}
		rows, _ := res.RowsAffected()
		if rows > 0 {
			_, _ = tx.ExecContext(ctx, `UPDATE user_profiles SET followers_count = followers_count + 1 WHERE user_id = $1;`, req.TargetID)
			_, _ = tx.ExecContext(ctx, `UPDATE user_profiles SET following_count = following_count + 1 WHERE user_id = $1;`, req.RequesterID)
		}

		// Notify requester that their request was accepted
		_, _ = tx.ExecContext(ctx, `
			INSERT INTO notifications (id, recipient_id, actor_id, type, is_read, entity_type, entity_id, created_at)
			VALUES ($1, $2, $3, 'follow_accept', FALSE, 'user', $4, NOW());
		`, uuid.New(), req.RequesterID, req.TargetID, req.TargetID)
	}

	_, err = tx.ExecContext(ctx, `UPDATE follow_requests SET status = $1, updated_at = NOW() WHERE id = $2;`, newStatus, req.ID)
	if err != nil {
		return err
	}

	// Mark the pending follow_request notification for the target as read
	_, _ = tx.ExecContext(ctx, `
		UPDATE notifications SET is_read = TRUE
		WHERE recipient_id = $1 AND (entity_id = $2 OR (actor_id = $3 AND type = 'follow_request'));
	`, targetID, req.ID, req.RequesterID)

	return tx.Commit()
}

func (r *PostgresRepository) CancelFollowRequest(ctx context.Context, requesterID, targetID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM follow_requests WHERE requester_id = $1 AND target_id = $2 AND status = 'pending';`, requesterID, targetID)
	return err
}

// ==========================================
// DIRECT MESSAGING IMPLEMENTATION
// ==========================================

func (r *PostgresRepository) GetConversations(ctx context.Context, userID uuid.UUID, limit int, cursor string) ([]domain.Conversation, error) {
	if limit <= 0 || limit > 50 {
		limit = 25
	}

	query := `
		SELECT c.id, c.is_group, COALESCE(c.name, ''), c.created_at, c.updated_at,
		       other_u.id, other_u.username, other_p.display_name, other_p.avatar_url, other_u.is_verified, other_u.verified_type,
		       COALESCE(
		           (SELECT COUNT(*) FROM messages m
		            JOIN conversation_members cm ON cm.conversation_id = m.conversation_id AND cm.user_id = $1
		            WHERE m.conversation_id = c.id AND m.created_at > cm.last_read_at AND m.sender_id != $1), 0) AS unread_count
		FROM conversations c
		JOIN conversation_members my_m ON my_m.conversation_id = c.id AND my_m.user_id = $1
		LEFT JOIN conversation_members other_m ON other_m.conversation_id = c.id AND other_m.user_id != $1
		LEFT JOIN users other_u ON other_u.id = other_m.user_id
		LEFT JOIN user_profiles other_p ON other_p.user_id = other_u.id
		WHERE EXISTS (SELECT 1 FROM messages m WHERE m.conversation_id = c.id)
		ORDER BY c.updated_at DESC
		LIMIT $2;
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var convs []domain.Conversation
	for rows.Next() {
		var c domain.Conversation
		var otherUser domain.User
		var oID *uuid.UUID
		var oName, oAvatar, oVerifiedType *string
		var oUsername *string
		var oVerified *bool

		if err := rows.Scan(
			&c.ID, &c.IsGroup, &c.Name, &c.CreatedAt, &c.UpdatedAt,
			&oID, &oUsername, &oName, &oAvatar, &oVerified, &oVerifiedType,
			&c.UnreadCount,
		); err != nil {
			return nil, err
		}

		if oID != nil {
			otherUser.ID = *oID
			if oUsername != nil {
				otherUser.Username = *oUsername
			}
			if oName != nil {
				otherUser.DisplayName = *oName
			}
			if oAvatar != nil {
				otherUser.AvatarURL = *oAvatar
			}
			if oVerified != nil {
				otherUser.IsVerified = *oVerified
			}
			if oVerifiedType != nil {
				otherUser.VerifiedType = *oVerifiedType
			}
			c.OtherMember = &otherUser
		}

		// Fetch last message preview
		var lastMsg domain.Message
		qMsg := `SELECT id, conversation_id, sender_id, content, message_type, created_at FROM messages WHERE conversation_id = $1 ORDER BY created_at DESC LIMIT 1;`
		if err := r.db.QueryRowContext(ctx, qMsg, c.ID).Scan(&lastMsg.ID, &lastMsg.ConversationID, &lastMsg.SenderID, &lastMsg.Content, &lastMsg.MessageType, &lastMsg.CreatedAt); err == nil {
			c.LastMessage = &lastMsg
		}

		convs = append(convs, c)
	}

	return convs, nil
}

func (r *PostgresRepository) GetOrCreateConversation(ctx context.Context, userA, userB uuid.UUID) (*domain.Conversation, error) {
	// Look for existing 1-on-1 conversation
	qFind := `
		SELECT c.id, c.is_group, COALESCE(c.name, ''), c.created_at, c.updated_at
		FROM conversations c
		JOIN conversation_members m1 ON m1.conversation_id = c.id AND m1.user_id = $1
		JOIN conversation_members m2 ON m2.conversation_id = c.id AND m2.user_id = $2
		WHERE c.is_group = FALSE
		LIMIT 1;
	`
	var c domain.Conversation
	err := r.db.QueryRowContext(ctx, qFind, userA, userB).Scan(&c.ID, &c.IsGroup, &c.Name, &c.CreatedAt, &c.UpdatedAt)
	if err == nil {
		return &c, nil
	}

	// Create new conversation in transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	newID := uuid.New()
	qCreate := `INSERT INTO conversations (id, is_group, created_at, updated_at) VALUES ($1, FALSE, NOW(), NOW()) RETURNING created_at, updated_at;`
	if err := tx.QueryRowContext(ctx, qCreate, newID).Scan(&c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, err
	}
	c.ID = newID
	c.IsGroup = false

	// Add members
	qMember := `INSERT INTO conversation_members (conversation_id, user_id, joined_at, last_read_at) VALUES ($1, $2, NOW(), NOW());`
	if _, err := tx.ExecContext(ctx, qMember, newID, userA); err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, qMember, newID, userB); err != nil {
		return nil, err
	}

	return &c, tx.Commit()
}

func (r *PostgresRepository) GetConversationByID(ctx context.Context, convID, userID uuid.UUID) (*domain.Conversation, error) {
	q := `
		SELECT c.id, c.is_group, COALESCE(c.name, ''), c.created_at, c.updated_at,
		       other_u.id, other_u.username, other_p.display_name, other_p.avatar_url, other_u.is_verified, other_u.verified_type
		FROM conversations c
		JOIN conversation_members my_m ON my_m.conversation_id = c.id AND my_m.user_id = $2
		LEFT JOIN conversation_members other_m ON other_m.conversation_id = c.id AND other_m.user_id != $2
		LEFT JOIN users other_u ON other_u.id = other_m.user_id
		LEFT JOIN user_profiles other_p ON other_p.user_id = other_u.id
		WHERE c.id = $1;
	`
	var c domain.Conversation
	var otherUser domain.User
	var oID *uuid.UUID
	var oName, oAvatar, oVerifiedType *string
	var oUsername *string
	var oVerified *bool

	err := r.db.QueryRowContext(ctx, q, convID, userID).Scan(
		&c.ID, &c.IsGroup, &c.Name, &c.CreatedAt, &c.UpdatedAt,
		&oID, &oUsername, &oName, &oAvatar, &oVerified, &oVerifiedType,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if oID != nil {
		otherUser.ID = *oID
		if oUsername != nil {
			otherUser.Username = *oUsername
		}
		if oName != nil {
			otherUser.DisplayName = *oName
		}
		if oAvatar != nil {
			otherUser.AvatarURL = *oAvatar
		}
		if oVerified != nil {
			otherUser.IsVerified = *oVerified
		}
		if oVerifiedType != nil {
			otherUser.VerifiedType = *oVerifiedType
		}
		c.OtherMember = &otherUser
	}

	return &c, nil
}

func (r *PostgresRepository) GetMessages(ctx context.Context, convID, userID uuid.UUID, limit int, beforeID *uuid.UUID) ([]domain.Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	// Verify user is a member of this conversation
	var isMember bool
	_ = r.db.QueryRowContext(ctx, `SELECT TRUE FROM conversation_members WHERE conversation_id = $1 AND user_id = $2;`, convID, userID).Scan(&isMember)
	if !isMember {
		return nil, errors.New("forbidden: not a member of conversation")
	}

	query := `
		SELECT m.id, m.conversation_id, m.sender_id, m.content, m.message_type, m.client_message_id, m.media_id, m.reply_to_id, m.created_at,
		       u.username, p.display_name, p.avatar_url, u.is_verified, u.verified_type,
		       mo.url, mo.media_type, mo.file_size_bytes, mo.width, mo.height
		FROM messages m
		JOIN users u ON u.id = m.sender_id
		JOIN user_profiles p ON p.user_id = u.id
		LEFT JOIN media_objects mo ON mo.id = m.media_id
		WHERE m.conversation_id = $1 AND ($2::uuid IS NULL OR m.created_at < (SELECT created_at FROM messages WHERE id = $2))
		ORDER BY m.created_at ASC
		LIMIT $3;
	`

	rows, err := r.db.QueryContext(ctx, query, convID, beforeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []domain.Message
	for rows.Next() {
		var m domain.Message
		var u domain.User
		var clientMsgID, replyToID sql.NullString
		var mediaURL, mediaType sql.NullString
		var mediaSize sql.NullInt64
		var mediaW, mediaH sql.NullInt32

		if err := rows.Scan(
			&m.ID, &m.ConversationID, &m.SenderID, &m.Content, &m.MessageType, &clientMsgID, &m.MediaID, &replyToID, &m.CreatedAt,
			&u.Username, &u.DisplayName, &u.AvatarURL, &u.IsVerified, &u.VerifiedType,
			&mediaURL, &mediaType, &mediaSize, &mediaW, &mediaH,
		); err != nil {
			return nil, err
		}

		u.ID = m.SenderID
		m.Sender = &u
		if clientMsgID.Valid {
			m.ClientMessageID = clientMsgID.String
		}

		if mediaURL.Valid && m.MediaID != nil {
			var wPtr, hPtr *int
			if mediaW.Valid {
				wVal := int(mediaW.Int32)
				wPtr = &wVal
			}
			if mediaH.Valid {
				hVal := int(mediaH.Int32)
				hPtr = &hVal
			}
			m.Media = &domain.MediaObject{
				ID:            *m.MediaID,
				URL:           mediaURL.String,
				MediaType:     mediaType.String,
				FileSizeBytes: mediaSize.Int64,
				Width:         wPtr,
				Height:        hPtr,
			}
		}

		msgs = append(msgs, m)
	}

	return msgs, nil
}

func (r *PostgresRepository) CreateMessage(ctx context.Context, msg *domain.Message) error {
	if msg.ID == uuid.Nil {
		msg.ID = uuid.New()
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Idempotency: If client_message_id is provided, check if already sent
	if msg.ClientMessageID != "" {
		var existingID uuid.UUID
		qCheck := `SELECT id, created_at FROM messages WHERE conversation_id = $1 AND client_message_id = $2;`
		if err := tx.QueryRowContext(ctx, qCheck, msg.ConversationID, msg.ClientMessageID).Scan(&existingID, &msg.CreatedAt); err == nil {
			msg.ID = existingID
			return nil // already processed, return existing
		}
	}

	q := `
		INSERT INTO messages (id, conversation_id, sender_id, client_message_id, content, message_type, media_id, reply_to_id, created_at)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, COALESCE(NULLIF($6, ''), 'text'), $7, $8, NOW())
		RETURNING created_at;
	`
	err = tx.QueryRowContext(ctx, q, msg.ID, msg.ConversationID, msg.SenderID, msg.ClientMessageID, msg.Content, msg.MessageType, msg.MediaID, msg.ReplyToID).Scan(&msg.CreatedAt)
	if err != nil {
		return err
	}

	// Update conversation updated_at
	_, _ = tx.ExecContext(ctx, `UPDATE conversations SET updated_at = NOW() WHERE id = $1;`, msg.ConversationID)

	// Update sender's last_read_at
	_, _ = tx.ExecContext(ctx, `UPDATE conversation_members SET last_read_at = NOW(), last_read_message_id = $1 WHERE conversation_id = $2 AND user_id = $3;`, msg.ID, msg.ConversationID, msg.SenderID)

	return tx.Commit()
}

func (r *PostgresRepository) MarkConversationRead(ctx context.Context, convID, userID uuid.UUID, messageID *uuid.UUID) error {
	query := `
		UPDATE conversation_members
		SET last_read_at = NOW(), last_read_message_id = COALESCE($3, last_read_message_id)
		WHERE conversation_id = $1 AND user_id = $2;
	`
	_, err := r.db.ExecContext(ctx, query, convID, userID, messageID)
	return err
}

// ==========================================
// NOTIFICATIONS WITH TABS & UNREAD COUNT
// ==========================================

func (r *PostgresRepository) GetNotificationsWithTab(ctx context.Context, userID uuid.UUID, tab string, cursor string, limit int) ([]domain.Notification, string, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	filterClause := ""
	if tab == "verified" {
		filterClause = "AND actor_u.is_verified = TRUE"
	} else if tab == "mentions" {
		filterClause = "AND n.type IN ('reply', 'mention')"
	}

	query := fmt.Sprintf(`
		SELECT n.id, n.recipient_id, n.actor_id, n.type, n.post_id, n.entity_type, n.entity_id, n.is_read, n.created_at,
		       actor_u.username, actor_p.display_name, actor_p.avatar_url, actor_u.is_verified, actor_u.verified_type,
		       p.content
		FROM notifications n
		JOIN users actor_u ON actor_u.id = n.actor_id
		JOIN user_profiles actor_p ON actor_p.user_id = actor_u.id
		LEFT JOIN posts p ON p.id = n.post_id
		WHERE n.recipient_id = $1 %s
		ORDER BY n.created_at DESC
		LIMIT $2;
	`, filterClause)

	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var notifs []domain.Notification
	for rows.Next() {
		var n domain.Notification
		var actor domain.User
		var postContent, entityType sql.NullString
		var entityID *uuid.UUID

		if err := rows.Scan(
			&n.ID, &n.RecipientID, &n.ActorID, &n.Type, &n.PostID, &entityType, &entityID, &n.IsRead, &n.CreatedAt,
			&actor.Username, &actor.DisplayName, &actor.AvatarURL, &actor.IsVerified, &actor.VerifiedType,
			&postContent,
		); err != nil {
			return nil, "", err
		}

		actor.ID = n.ActorID
		n.Actor = &actor
		if entityType.Valid {
			n.EntityType = entityType.String
		}
		n.EntityID = entityID

		if postContent.Valid && n.PostID != nil {
			n.Post = &domain.Post{
				ID:      *n.PostID,
				Content: postContent.String,
			}
		}

		notifs = append(notifs, n)
	}

	return notifs, "", nil
}

func (r *PostgresRepository) GetUnreadNotificationsCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	q := `SELECT COUNT(*) FROM notifications WHERE recipient_id = $1 AND is_read = FALSE;`
	err := r.db.QueryRowContext(ctx, q, userID).Scan(&count)
	return count, err
}

func (r *PostgresRepository) MarkNotificationRead(ctx context.Context, notifID, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE notifications SET is_read = TRUE, read_at = NOW() WHERE id = $1 AND recipient_id = $2;`, notifID, userID)
	return err
}

// ==========================================
// TRANSACTIONAL OUTBOX EVENTS
// ==========================================

func (r *PostgresRepository) CreateOutboxEvent(ctx context.Context, event *domain.OutboxEvent) error {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	payloadJSON, _ := json.Marshal(event.Payload)
	q := `
		INSERT INTO outbox_events (id, event_type, aggregate_type, aggregate_id, payload, status, retry_count, created_at)
		VALUES ($1, $2, $3, $4, $5, 'pending', 0, NOW())
		RETURNING created_at;
	`
	return r.db.QueryRowContext(ctx, q, event.ID, event.EventType, event.AggregateType, event.AggregateID, payloadJSON).Scan(&event.CreatedAt)
}

func (r *PostgresRepository) ensureOutboxSchema(ctx context.Context) {
	stmts := []string{
		`ALTER TABLE outbox_events ADD COLUMN IF NOT EXISTS last_error TEXT NULL;`,
		`ALTER TABLE outbox_events ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMPTZ NOT NULL DEFAULT NOW();`,
		`CREATE INDEX IF NOT EXISTS idx_outbox_poll_ready ON outbox_events(next_retry_at ASC) WHERE status IN ('pending', 'retrying');`,
		`CREATE INDEX IF NOT EXISTS idx_outbox_retention_cleanup ON outbox_events(published_at) WHERE status = 'published';`,
	}
	for _, stmt := range stmts {
		_, _ = r.db.ExecContext(ctx, stmt)
	}
}

func (r *PostgresRepository) GetPendingOutboxEvents(ctx context.Context, batchSize int) ([]domain.OutboxEvent, error) {
	if batchSize <= 0 || batchSize > 100 {
		batchSize = 25
	}
	q := `
		SELECT id, event_type, aggregate_type, aggregate_id, payload, status, retry_count, created_at, last_error, next_retry_at
		FROM outbox_events
		WHERE status IN ('pending', 'retrying') AND next_retry_at <= NOW()
		ORDER BY next_retry_at ASC
		LIMIT $1;
	`
	rows, err := r.db.QueryContext(ctx, q, batchSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.OutboxEvent
	for rows.Next() {
		var e domain.OutboxEvent
		var payloadBytes []byte
		var lastError sql.NullString
		var nextRetryAt sql.NullTime
		if err := rows.Scan(
			&e.ID, &e.EventType, &e.AggregateType, &e.AggregateID, &payloadBytes,
			&e.Status, &e.RetryCount, &e.CreatedAt, &lastError, &nextRetryAt,
		); err != nil {
			return nil, err
		}
		if lastError.Valid {
			e.LastError = &lastError.String
		}
		if nextRetryAt.Valid {
			e.NextRetryAt = &nextRetryAt.Time
		}
		_ = json.Unmarshal(payloadBytes, &e.Payload)
		events = append(events, e)
	}
	return events, nil
}

func (r *PostgresRepository) MarkOutboxEventPublished(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE outbox_events SET status = 'published', published_at = NOW(), last_error = NULL WHERE id = $1;`, id)
	return err
}

func (r *PostgresRepository) MarkOutboxEventFailed(ctx context.Context, id uuid.UUID, errMsg string, nextRetryAt *time.Time, isDLQ bool) error {
	status := "retrying"
	if isDLQ {
		status = "failed"
	}
	nextTime := time.Now().Add(10 * time.Second)
	if nextRetryAt != nil {
		nextTime = *nextRetryAt
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE outbox_events 
		SET status = $1, retry_count = retry_count + 1, last_error = $2, next_retry_at = $3 
		WHERE id = $4;
	`, status, errMsg, nextTime, id)
	return err
}

func (r *PostgresRepository) CleanupOldOutboxEvents(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM outbox_events 
		WHERE status = 'published' AND published_at < $1;
	`, cutoff)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ==============================================================================
// CandidateProvider & Recommendation Implementation (Postgres)
// ==============================================================================

func (r *PostgresRepository) GetFollowingCandidates(ctx context.Context, viewerID uuid.UUID, limit int) ([]domain.Post, error) {
	if limit <= 0 || limit > 100 {
		limit = 80
	}
	query := `
		SELECT p.id, p.author_id, p.content, p.reply_to_id, p.quote_post_id,
		       p.replies_count, p.reposts_count, p.likes_count, p.views_count, p.created_at, p.updated_at,
		       u.username, u.is_verified, u.verified_type,
		       pr.display_name, pr.avatar_url,
		       COALESCE((SELECT TRUE FROM post_likes pl WHERE pl.post_id = p.id AND pl.user_id = $1), FALSE) as is_liked,
		       COALESCE((SELECT TRUE FROM post_reposts prp WHERE prp.post_id = p.id AND prp.user_id = $1), FALSE) as is_reposted,
		       COALESCE((SELECT TRUE FROM post_bookmarks pb WHERE pb.post_id = p.id AND pb.user_id = $1), FALSE) as is_bookmarked
		FROM posts p
		JOIN follows f ON f.following_id = p.author_id AND f.follower_id = $1
		JOIN users u ON u.id = p.author_id
		JOIN user_profiles pr ON pr.user_id = p.author_id
		WHERE p.deleted_at IS NULL AND p.created_at >= NOW() - INTERVAL '72 hours'
		ORDER BY p.created_at DESC, p.id DESC
		LIMIT $2;
	`
	rows, err := r.db.QueryContext(ctx, query, viewerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanPostRows(rows)
}

func (r *PostgresRepository) GetSocialEngagementCandidates(ctx context.Context, viewerID uuid.UUID, limit int) ([]domain.Post, error) {
	if limit <= 0 || limit > 100 {
		limit = 40
	}
	query := `
		SELECT DISTINCT p.id, p.author_id, p.content, p.reply_to_id, p.quote_post_id,
		       p.replies_count, p.reposts_count, p.likes_count, p.views_count, p.created_at, p.updated_at,
		       u.username, u.is_verified, u.verified_type,
		       pr.display_name, pr.avatar_url,
		       COALESCE((SELECT TRUE FROM post_likes pl WHERE pl.post_id = p.id AND pl.user_id = $1), FALSE) as is_liked,
		       COALESCE((SELECT TRUE FROM post_reposts prp WHERE prp.post_id = p.id AND prp.user_id = $1), FALSE) as is_reposted,
		       COALESCE((SELECT TRUE FROM post_bookmarks pb WHERE pb.post_id = p.id AND pb.user_id = $1), FALSE) as is_bookmarked
		FROM posts p
		JOIN post_likes pl2 ON pl2.post_id = p.id
		JOIN follows f ON f.following_id = pl2.user_id AND f.follower_id = $1
		JOIN users u ON u.id = p.author_id
		JOIN user_profiles pr ON pr.user_id = p.author_id
		WHERE p.deleted_at IS NULL AND p.created_at >= NOW() - INTERVAL '48 hours'
		ORDER BY p.created_at DESC
		LIMIT $2;
	`
	rows, err := r.db.QueryContext(ctx, query, viewerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanPostRows(rows)
}

func (r *PostgresRepository) GetTopicCandidates(ctx context.Context, viewerID *uuid.UUID, topics []string, limit int) ([]domain.Post, error) {
	if len(topics) == 0 {
		return nil, nil
	}
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	var vID interface{} = uuid.Nil
	if viewerID != nil {
		vID = *viewerID
	}

	cleaned := make([]string, 0, len(topics))
	for _, t := range topics {
		clean := strings.ToLower(strings.TrimPrefix(t, "#"))
		if clean != "" {
			cleaned = append(cleaned, clean)
		}
	}
	if len(cleaned) == 0 {
		return nil, nil
	}

	query := `
		SELECT DISTINCT p.id, p.author_id, p.content, p.reply_to_id, p.quote_post_id,
		       p.replies_count, p.reposts_count, p.likes_count, p.views_count, p.created_at, p.updated_at,
		       u.username, u.is_verified, u.verified_type,
		       pr.display_name, pr.avatar_url,
		       COALESCE((SELECT TRUE FROM post_likes pl WHERE pl.post_id = p.id AND pl.user_id = $1), FALSE) as is_liked,
		       COALESCE((SELECT TRUE FROM post_reposts prp WHERE prp.post_id = p.id AND prp.user_id = $1), FALSE) as is_reposted,
		       COALESCE((SELECT TRUE FROM post_bookmarks pb WHERE pb.post_id = p.id AND pb.user_id = $1), FALSE) as is_bookmarked
		FROM posts p
		JOIN post_hashtags ph ON ph.post_id = p.id
		JOIN hashtags h ON h.id = ph.hashtag_id
		JOIN users u ON u.id = p.author_id
		JOIN user_profiles pr ON pr.user_id = p.author_id
		WHERE p.deleted_at IS NULL AND LOWER(h.tag) = ANY($2)
		ORDER BY p.created_at DESC
		LIMIT $3;
	`
	rows, err := r.db.QueryContext(ctx, query, vID, cleaned, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanPostRows(rows)
}

func (r *PostgresRepository) GetTrendingCandidates(ctx context.Context, limit int) ([]domain.Post, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	query := `
		SELECT p.id, p.author_id, p.content, p.reply_to_id, p.quote_post_id,
		       p.replies_count, p.reposts_count, p.likes_count, p.views_count, p.created_at, p.updated_at,
		       u.username, u.is_verified, u.verified_type,
		       pr.display_name, pr.avatar_url,
		       FALSE as is_liked, FALSE as is_reposted, FALSE as is_bookmarked
		FROM posts p
		JOIN users u ON u.id = p.author_id
		JOIN user_profiles pr ON pr.user_id = p.author_id
		WHERE p.deleted_at IS NULL AND p.created_at >= NOW() - INTERVAL '72 hours'
		ORDER BY (p.likes_count + p.reposts_count * 2 + p.replies_count * 3) DESC, p.created_at DESC
		LIMIT $1;
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanPostRows(rows)
}

func (r *PostgresRepository) GetFreshCandidates(ctx context.Context, limit int) ([]domain.Post, error) {
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	query := `
		SELECT p.id, p.author_id, p.content, p.reply_to_id, p.quote_post_id,
		       p.replies_count, p.reposts_count, p.likes_count, p.views_count, p.created_at, p.updated_at,
		       u.username, u.is_verified, u.verified_type,
		       pr.display_name, pr.avatar_url,
		       FALSE as is_liked, FALSE as is_reposted, FALSE as is_bookmarked
		FROM posts p
		JOIN users u ON u.id = p.author_id
		JOIN user_profiles pr ON pr.user_id = p.author_id
		WHERE p.deleted_at IS NULL
		ORDER BY p.created_at DESC, p.id DESC
		LIMIT $1;
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanPostRows(rows)
}

func (r *PostgresRepository) GetExplorationCandidates(ctx context.Context, limit int) ([]domain.Post, error) {
	if limit <= 0 || limit > 100 {
		limit = 15
	}
	query := `
		SELECT p.id, p.author_id, p.content, p.reply_to_id, p.quote_post_id,
		       p.replies_count, p.reposts_count, p.likes_count, p.views_count, p.created_at, p.updated_at,
		       u.username, u.is_verified, u.verified_type,
		       pr.display_name, pr.avatar_url,
		       FALSE as is_liked, FALSE as is_reposted, FALSE as is_bookmarked
		FROM posts p
		JOIN users u ON u.id = p.author_id
		JOIN user_profiles pr ON pr.user_id = p.author_id
		WHERE p.deleted_at IS NULL AND pr.followers_count < 1000 AND p.reply_to_id IS NULL
		ORDER BY p.created_at DESC
		LIMIT $1;
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanPostRows(rows)
}

func (r *PostgresRepository) GetFollowingPosts(ctx context.Context, viewerID uuid.UUID, cursor *recommendation.FeedCursor, limit int) ([]domain.Post, string, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	var rows *sql.Rows
	var err error

	if cursor != nil {
		query := `
			SELECT p.id, p.author_id, p.content, p.reply_to_id, p.quote_post_id,
			       p.replies_count, p.reposts_count, p.likes_count, p.views_count, p.created_at, p.updated_at,
			       u.username, u.is_verified, u.verified_type,
			       pr.display_name, pr.avatar_url,
			       COALESCE((SELECT TRUE FROM post_likes pl WHERE pl.post_id = p.id AND pl.user_id = $1), FALSE) as is_liked,
			       COALESCE((SELECT TRUE FROM post_reposts prp WHERE prp.post_id = p.id AND prp.user_id = $1), FALSE) as is_reposted,
			       COALESCE((SELECT TRUE FROM post_bookmarks pb WHERE pb.post_id = p.id AND pb.user_id = $1), FALSE) as is_bookmarked
			FROM posts p
			JOIN follows f ON f.following_id = p.author_id AND f.follower_id = $1
			JOIN users u ON u.id = p.author_id
			JOIN user_profiles pr ON pr.user_id = p.author_id
			WHERE p.deleted_at IS NULL AND (p.created_at, p.id) < ($2, $3)
			ORDER BY p.created_at DESC, p.id DESC
			LIMIT $4;
		`
		rows, err = r.db.QueryContext(ctx, query, viewerID, cursor.CreatedAt, cursor.PostID, limit+1)
	} else {
		query := `
			SELECT p.id, p.author_id, p.content, p.reply_to_id, p.quote_post_id,
			       p.replies_count, p.reposts_count, p.likes_count, p.views_count, p.created_at, p.updated_at,
			       u.username, u.is_verified, u.verified_type,
			       pr.display_name, pr.avatar_url,
			       COALESCE((SELECT TRUE FROM post_likes pl WHERE pl.post_id = p.id AND pl.user_id = $1), FALSE) as is_liked,
			       COALESCE((SELECT TRUE FROM post_reposts prp WHERE prp.post_id = p.id AND prp.user_id = $1), FALSE) as is_reposted,
			       COALESCE((SELECT TRUE FROM post_bookmarks pb WHERE pb.post_id = p.id AND pb.user_id = $1), FALSE) as is_bookmarked
			FROM posts p
			JOIN follows f ON f.following_id = p.author_id AND f.follower_id = $1
			JOIN users u ON u.id = p.author_id
			JOIN user_profiles pr ON pr.user_id = p.author_id
			WHERE p.deleted_at IS NULL
			ORDER BY p.created_at DESC, p.id DESC
			LIMIT $2;
		`
		rows, err = r.db.QueryContext(ctx, query, viewerID, limit+1)
	}

	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	posts, err := r.scanPostRows(rows)
	if err != nil {
		return nil, "", err
	}

	hasMore := len(posts) > limit
	if len(posts) > limit {
		posts = posts[:limit]
	}

	var nextCursor string
	if hasMore && len(posts) > 0 {
		last := posts[len(posts)-1]
		nextCursor = recommendation.EncodeCursor(last.CreatedAt, last.ID)
	}

	return posts, nextCursor, nil
}

func (r *PostgresRepository) GetViewerInteractions(ctx context.Context, viewerID uuid.UUID) (
	followedAuthors map[uuid.UUID]bool,
	mutualAuthors map[uuid.UUID]bool,
	likedPosts map[uuid.UUID]bool,
	mutedAuthors map[uuid.UUID]bool,
	blockedAuthors map[uuid.UUID]bool,
	hiddenPosts map[uuid.UUID]bool,
	userTopics map[string]float64,
	err error,
) {
	followedAuthors = make(map[uuid.UUID]bool)
	mutualAuthors = make(map[uuid.UUID]bool)
	likedPosts = make(map[uuid.UUID]bool)
	mutedAuthors = make(map[uuid.UUID]bool)
	blockedAuthors = make(map[uuid.UUID]bool)
	hiddenPosts = make(map[uuid.UUID]bool)
	userTopics = make(map[string]float64)

	// 1. Followed and Mutual accounts
	followRows, err := r.db.QueryContext(ctx, `
		SELECT f1.following_id, (f2.follower_id IS NOT NULL) as is_mutual
		FROM follows f1
		LEFT JOIN follows f2 ON f2.follower_id = f1.following_id AND f2.following_id = $1
		WHERE f1.follower_id = $1;
	`, viewerID)
	if err == nil {
		defer followRows.Close()
		for followRows.Next() {
			var fID uuid.UUID
			var isMutual bool
			if err := followRows.Scan(&fID, &isMutual); err == nil {
				followedAuthors[fID] = true
				if isMutual {
					mutualAuthors[fID] = true
				}
			}
		}
	}

	// 2. Liked Posts (Recent 500)
	likeRows, err := r.db.QueryContext(ctx, `SELECT post_id FROM post_likes WHERE user_id = $1 ORDER BY created_at DESC LIMIT 500;`, viewerID)
	if err == nil {
		defer likeRows.Close()
		for likeRows.Next() {
			var pID uuid.UUID
			if err := likeRows.Scan(&pID); err == nil {
				likedPosts[pID] = true
			}
		}
	}

	// 3. Muted Accounts
	muteRows, err := r.db.QueryContext(ctx, `SELECT muted_id FROM mutes WHERE muter_id = $1;`, viewerID)
	if err == nil {
		defer muteRows.Close()
		for muteRows.Next() {
			var mID uuid.UUID
			if err := muteRows.Scan(&mID); err == nil {
				mutedAuthors[mID] = true
			}
		}
	}

	// 4. Bilateral Blocks
	blockRows, err := r.db.QueryContext(ctx, `
		SELECT blocked_id FROM blocks WHERE blocker_id = $1
		UNION
		SELECT blocker_id FROM blocks WHERE blocked_id = $1;
	`, viewerID)
	if err == nil {
		defer blockRows.Close()
		for blockRows.Next() {
			var bID uuid.UUID
			if err := blockRows.Scan(&bID); err == nil {
				blockedAuthors[bID] = true
			}
		}
	}

	// 5. Hidden Posts from Feed Feedback
	feedRows, err := r.db.QueryContext(ctx, `SELECT post_id FROM feed_feedback WHERE user_id = $1 AND feedback_type IN ('hide', 'not_interested');`, viewerID)
	if err == nil {
		defer feedRows.Close()
		for feedRows.Next() {
			var pID uuid.UUID
			if err := feedRows.Scan(&pID); err == nil {
				hiddenPosts[pID] = true
			}
		}
	}

	// 6. User Topic Preferences
	topicRows, err := r.db.QueryContext(ctx, `SELECT topic, affinity_score FROM user_topic_preferences WHERE user_id = $1 AND is_muted = FALSE;`, viewerID)
	if err == nil {
		defer topicRows.Close()
		for topicRows.Next() {
			var topic string
			var score float64
			if err := topicRows.Scan(&topic, &score); err == nil {
				userTopics[topic] = score
			}
		}
	}

	return followedAuthors, mutualAuthors, likedPosts, mutedAuthors, blockedAuthors, hiddenPosts, userTopics, nil
}

func (r *PostgresRepository) GetActiveWeights(ctx context.Context) (domain.RankingWeights, string, error) {
	var version string
	var weightsJSON []byte
	err := r.db.QueryRowContext(ctx, `SELECT version, weights FROM ranking_experiments WHERE is_active = TRUE ORDER BY created_at DESC LIMIT 1;`).Scan(&version, &weightsJSON)
	if err != nil {
		return domain.DefaultRankingWeights(), "default_baseline", nil
	}

	var w domain.RankingWeights
	if err := json.Unmarshal(weightsJSON, &w); err != nil {
		return domain.DefaultRankingWeights(), "default_baseline", nil
	}
	return w, version, nil
}

func (r *PostgresRepository) RecordFeedFeedback(ctx context.Context, feedback domain.FeedFeedback) error {
	if feedback.ID == uuid.Nil {
		feedback.ID = uuid.New()
	}
	query := `
		INSERT INTO feed_feedback (id, user_id, post_id, feedback_type, topic, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (user_id, post_id, feedback_type) DO NOTHING;
	`
	_, err := r.db.ExecContext(ctx, query, feedback.ID, feedback.UserID, feedback.PostID, feedback.FeedbackType, feedback.Topic)
	return err
}

func (r *PostgresRepository) RecordFeedImpressions(ctx context.Context, impressions []domain.FeedImpression) error {
	if len(impressions) == 0 {
		return nil
	}

	for _, imp := range impressions {
		impID := imp.ID
		if impID == uuid.Nil {
			impID = uuid.New()
		}
		var uID interface{} = nil
		if imp.UserID != nil {
			uID = *imp.UserID
		}
		_, _ = r.db.ExecContext(ctx, `
			INSERT INTO feed_impressions (id, user_id, post_id, feed_type, dwell_time_ms, created_at)
			VALUES ($1, $2, $3, $4, $5, NOW());
		`, impID, uID, imp.PostID, imp.FeedType, imp.DwellTimeMs)
	}
	return nil
}

func (r *PostgresRepository) GetUserTopicPreferences(ctx context.Context, userID uuid.UUID) ([]domain.UserTopicPreference, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, topic, affinity_score, is_explicit, is_muted, updated_at
		FROM user_topic_preferences
		WHERE user_id = $1
		ORDER BY affinity_score DESC;
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prefs []domain.UserTopicPreference
	for rows.Next() {
		var p domain.UserTopicPreference
		if err := rows.Scan(&p.ID, &p.UserID, &p.Topic, &p.AffinityScore, &p.IsExplicit, &p.IsMuted, &p.UpdatedAt); err != nil {
			return nil, err
		}
		prefs = append(prefs, p)
	}
	return prefs, nil
}

func (r *PostgresRepository) UpdateUserTopicPreference(ctx context.Context, pref domain.UserTopicPreference) error {
	if pref.ID == uuid.Nil {
		pref.ID = uuid.New()
	}
	query := `
		INSERT INTO user_topic_preferences (id, user_id, topic, affinity_score, is_explicit, is_muted, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (user_id, topic) DO UPDATE
		SET affinity_score = EXCLUDED.affinity_score,
		    is_explicit = EXCLUDED.is_explicit,
		    is_muted = EXCLUDED.is_muted,
		    updated_at = NOW();
	`
	_, err := r.db.ExecContext(ctx, query, pref.ID, pref.UserID, pref.Topic, pref.AffinityScore, pref.IsExplicit, pref.IsMuted)
	return err
}

func (r *PostgresRepository) ResetPersonalization(ctx context.Context, userID uuid.UUID) error {
	_, err1 := r.db.ExecContext(ctx, `DELETE FROM feed_feedback WHERE user_id = $1;`, userID)
	_, err2 := r.db.ExecContext(ctx, `DELETE FROM user_topic_preferences WHERE user_id = $1 AND is_explicit = FALSE;`, userID)
	if err1 != nil {
		return err1
	}
	return err2
}

func categorizeTopic(tag string) string {
	lower := strings.ToLower(tag)
	if strings.Contains(lower, "react") || strings.Contains(lower, "golang") || strings.Contains(lower, "ai") ||
		strings.Contains(lower, "next") || strings.Contains(lower, "code") || strings.Contains(lower, "tech") ||
		strings.Contains(lower, "dev") || strings.Contains(lower, "postgres") || strings.Contains(lower, "web") {
		return "Technology · Trending"
	}
	if strings.Contains(lower, "sport") || strings.Contains(lower, "ball") || strings.Contains(lower, "cup") ||
		strings.Contains(lower, "game") || strings.Contains(lower, "match") {
		return "Sports · Trending"
	}
	if strings.Contains(lower, "news") || strings.Contains(lower, "world") || strings.Contains(lower, "politic") {
		return "News · Trending"
	}
	if strings.Contains(lower, "music") || strings.Contains(lower, "movie") || strings.Contains(lower, "film") {
		return "Entertainment · Trending"
	}
	return "Trending"
}

func (r *PostgresRepository) GetTrendingTopics(ctx context.Context, limit int) ([]domain.TrendingTopic, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	// Rolling 24-hour window with unique author anti-spam weighting and engagement velocity
	query := `
		SELECT 
			h.tag,
			COUNT(DISTINCT p.id) as recent_posts,
			COUNT(DISTINCT p.author_id) as unique_authors,
			COALESCE(SUM(p.likes_count + p.reposts_count * 2 + p.replies_count * 3), 0) as total_engagement,
			COALESCE(MAX(h.posts_count), 0) as lifetime_posts
		FROM hashtags h
		JOIN post_hashtags ph ON ph.hashtag_id = h.id
		JOIN posts p ON p.id = ph.post_id
		WHERE p.deleted_at IS NULL AND p.created_at >= NOW() - INTERVAL '24 hours'
		GROUP BY h.id, h.tag
		HAVING COUNT(DISTINCT p.id) > 0
		ORDER BY (
			(COUNT(DISTINCT p.author_id) * 3.0) +
			(COUNT(DISTINCT p.id) * 1.5) +
			(COALESCE(SUM(p.likes_count + p.reposts_count * 2 + p.replies_count * 3), 0) * 0.1) -
			CASE WHEN COUNT(DISTINCT p.id) > COUNT(DISTINCT p.author_id) * 2 THEN (COUNT(DISTINCT p.id) - COUNT(DISTINCT p.author_id) * 2) * 2.0 ELSE 0 END
		) DESC
		LIMIT $1;
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []domain.TrendingTopic
	for rows.Next() {
		var tag string
		var recentPosts, uniqueAuthors, totalEngagement, lifetimePosts int
		if err := rows.Scan(&tag, &recentPosts, &uniqueAuthors, &totalEngagement, &lifetimePosts); err == nil {
			cleanTag := strings.TrimPrefix(tag, "#")
			rawVelocity := float64(uniqueAuthors)*3.0 + float64(recentPosts)*1.5 + float64(totalEngagement)*0.1
			if recentPosts > uniqueAuthors*2 {
				rawVelocity -= float64(recentPosts-uniqueAuthors*2) * 2.0
			}
			if rawVelocity < 0 {
				rawVelocity = 0
			}

			displayCount := recentPosts
			if displayCount < lifetimePosts {
				displayCount = lifetimePosts
			}

			topics = append(topics, domain.TrendingTopic{
				Category:      categorizeTopic(cleanTag),
				Topic:         "#" + cleanTag,
				PostsCount:    displayCount,
				VelocityScore: rawVelocity,
			})
		}
	}

	// Fallback to active non-empty hashtags if rolling window has fewer than required
	if len(topics) < limit {
		remaining := limit - len(topics)
		existingTags := make(map[string]bool)
		for _, t := range topics {
			existingTags[strings.ToLower(t.Topic)] = true
		}

		rows2, err := r.db.QueryContext(ctx, `
			SELECT tag, posts_count
			FROM hashtags
			WHERE posts_count > 0
			ORDER BY posts_count DESC
			LIMIT $1;
		`, remaining*2)
		if err == nil {
			defer rows2.Close()
			for rows2.Next() {
				if len(topics) >= limit {
					break
				}
				var tag string
				var count int
				if err := rows2.Scan(&tag, &count); err == nil {
					hashTopic := "#" + strings.TrimPrefix(tag, "#")
					if !existingTags[strings.ToLower(hashTopic)] {
						existingTags[strings.ToLower(hashTopic)] = true
						topics = append(topics, domain.TrendingTopic{
							Category:      categorizeTopic(tag),
							Topic:         hashTopic,
							PostsCount:    count,
							VelocityScore: float64(count),
						})
					}
				}
			}
		}
	}

	if topics == nil {
		topics = []domain.TrendingTopic{}
	}
	return topics, nil
}

func (r *PostgresRepository) GetNewPosts(ctx context.Context, feedType string, viewerID *uuid.UUID, sinceCursor string, limit int) ([]domain.Post, int, string, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	cursorObj, err := recommendation.DecodeCursor(sinceCursor)
	if err != nil || cursorObj == nil {
		return []domain.Post{}, 0, "", nil
	}

	var posts []domain.Post
	var totalNew int

	if feedType == "following" {
		if viewerID == nil {
			return []domain.Post{}, 0, "", nil
		}

		// Count total new eligible posts
		countQuery := `
			SELECT COUNT(DISTINCT p.id)
			FROM posts p
			JOIN follows f ON f.following_id = p.author_id AND f.follower_id = $1
			WHERE (p.created_at, p.id) > ($2, $3)
			  AND p.deleted_at IS NULL
			  AND p.author_id NOT IN (
				  SELECT blocked_id FROM user_blocks WHERE blocker_id = $1
				  UNION
				  SELECT blocker_id FROM user_blocks WHERE blocked_id = $1
			  )
			  AND p.author_id NOT IN (SELECT muted_id FROM user_mutes WHERE muter_id = $1)
		`
		_ = r.db.QueryRowContext(ctx, countQuery, *viewerID, cursorObj.CreatedAt, cursorObj.PostID).Scan(&totalNew)

		if totalNew > 0 {
			query := `
				SELECT p.id, p.author_id, p.content, p.reply_to_id, p.quote_post_id,
				       p.replies_count, p.reposts_count, p.likes_count, p.views_count, p.created_at, p.updated_at,
				       u.username, u.is_verified, u.verified_type,
				       pr.display_name, pr.avatar_url,
				       COALESCE((SELECT TRUE FROM post_likes pl WHERE pl.post_id = p.id AND pl.user_id = $1), FALSE) as is_liked,
				       COALESCE((SELECT TRUE FROM post_reposts prp WHERE prp.post_id = p.id AND prp.user_id = $1), FALSE) as is_reposted,
				       COALESCE((SELECT TRUE FROM post_bookmarks pb WHERE pb.post_id = p.id AND pb.user_id = $1), FALSE) as is_bookmarked
				FROM posts p
				JOIN follows f ON f.following_id = p.author_id AND f.follower_id = $1
				JOIN users u ON u.id = p.author_id
				JOIN user_profiles pr ON pr.user_id = p.author_id
				WHERE (p.created_at, p.id) > ($2, $3)
				  AND p.deleted_at IS NULL
				  AND p.author_id NOT IN (
					  SELECT blocked_id FROM user_blocks WHERE blocker_id = $1
					  UNION
					  SELECT blocker_id FROM user_blocks WHERE blocked_id = $1
				  )
				  AND p.author_id NOT IN (SELECT muted_id FROM user_mutes WHERE muter_id = $1)
				ORDER BY p.created_at DESC, p.id DESC
				LIMIT $4;
			`
			rows, err := r.db.QueryContext(ctx, query, *viewerID, cursorObj.CreatedAt, cursorObj.PostID, limit)
			if err == nil {
				defer rows.Close()
				posts, _ = r.scanPostRows(rows)
			}
		}
	} else {
		// For You feed mode
		var vID interface{} = uuid.Nil
		if viewerID != nil {
			vID = *viewerID
		}

		countQuery := `
			SELECT COUNT(DISTINCT p.id)
			FROM posts p
			WHERE (p.created_at, p.id) > ($2, $3)
			  AND p.deleted_at IS NULL
			  AND ($1 = '00000000-0000-0000-0000-000000000000'::uuid OR (
				  p.author_id NOT IN (
					  SELECT blocked_id FROM user_blocks WHERE blocker_id = $1
					  UNION
					  SELECT blocker_id FROM user_blocks WHERE blocked_id = $1
				  )
				  AND p.author_id NOT IN (SELECT muted_id FROM user_mutes WHERE muter_id = $1)
			  ))
		`
		_ = r.db.QueryRowContext(ctx, countQuery, vID, cursorObj.CreatedAt, cursorObj.PostID).Scan(&totalNew)

		if totalNew > 0 {
			query := `
				SELECT p.id, p.author_id, p.content, p.reply_to_id, p.quote_post_id,
				       p.replies_count, p.reposts_count, p.likes_count, p.views_count, p.created_at, p.updated_at,
				       u.username, u.is_verified, u.verified_type,
				       pr.display_name, pr.avatar_url,
				       COALESCE((SELECT TRUE FROM post_likes pl WHERE pl.post_id = p.id AND pl.user_id = $1), FALSE) as is_liked,
				       COALESCE((SELECT TRUE FROM post_reposts prp WHERE prp.post_id = p.id AND prp.user_id = $1), FALSE) as is_reposted,
				       COALESCE((SELECT TRUE FROM post_bookmarks pb WHERE pb.post_id = p.id AND pb.user_id = $1), FALSE) as is_bookmarked
				FROM posts p
				JOIN users u ON u.id = p.author_id
				JOIN user_profiles pr ON pr.user_id = p.author_id
				WHERE (p.created_at, p.id) > ($2, $3)
				  AND p.deleted_at IS NULL
				  AND ($1 = '00000000-0000-0000-0000-000000000000'::uuid OR (
					  p.author_id NOT IN (
						  SELECT blocked_id FROM user_blocks WHERE blocker_id = $1
						  UNION
						  SELECT blocker_id FROM user_blocks WHERE blocked_id = $1
					  )
					  AND p.author_id NOT IN (SELECT muted_id FROM user_mutes WHERE muter_id = $1)
				  ))
				ORDER BY p.created_at DESC, p.id DESC
				LIMIT $4;
			`
			rows, err := r.db.QueryContext(ctx, query, vID, cursorObj.CreatedAt, cursorObj.PostID, limit)
			if err == nil {
				defer rows.Close()
				posts, _ = r.scanPostRows(rows)
			}
		}
	}

	var nextCursor string
	if len(posts) > 0 {
		_ = r.attachMediaToPosts(ctx, posts)
		nextCursor = recommendation.EncodeCursor(posts[0].CreatedAt, posts[0].ID)
	}

	return posts, totalNew, nextCursor, nil
}

func (r *PostgresRepository) GetWhoToFollow(ctx context.Context, viewerID *uuid.UUID, limit int) ([]domain.UserRecommendation, error) {
	if limit <= 0 || limit > 30 {
		limit = 5
	}

	var recommendations []domain.UserRecommendation

	if viewerID != nil {
		// 1. Personalized candidate retrieval via 2nd-degree social network (Friends-of-Friends) & topic match
		query := `
			WITH my_follows AS (
				SELECT following_id FROM follows WHERE follower_id = $1
			),
			excluded AS (
				SELECT $1::uuid as user_id -- self
				UNION
				SELECT following_id FROM my_follows -- already followed
				UNION
				SELECT blocked_id FROM user_blocks WHERE blocker_id = $1 -- blocked
				UNION
				SELECT blocker_id FROM user_blocks WHERE blocked_id = $1 -- blocking viewer
				UNION
				SELECT muted_id FROM user_mutes WHERE muter_id = $1 -- muted
				UNION
				SELECT target_user_id FROM dismissed_user_recommendations WHERE user_id = $1 AND dismissed_until > NOW() -- dismissed
			),
			friend_of_friends AS (
				SELECT 
					f2.following_id as user_id,
					COUNT(DISTINCT f2.follower_id) as mutual_count,
					ARRAY_AGG(DISTINCT u_mutual.username) FILTER (WHERE u_mutual.username IS NOT NULL) as mutual_usernames
				FROM my_follows mf
				JOIN follows f2 ON f2.follower_id = mf.following_id
				JOIN users u_mutual ON u_mutual.id = f2.follower_id
				WHERE f2.following_id NOT IN (SELECT user_id FROM excluded)
				GROUP BY f2.following_id
			)
			SELECT 
				u.id, u.username, u.email, u.is_verified, u.verified_type, u.created_at,
				pr.display_name, pr.avatar_url, pr.bio, pr.location, pr.website,
				pr.followers_count, pr.following_count, pr.is_private,
				COALESCE(fof.mutual_count, 0) as mutual_count,
				COALESCE(fof.mutual_usernames, '{}'::text[]) as mutual_usernames
			FROM users u
			JOIN user_profiles pr ON pr.user_id = u.id
			LEFT JOIN friend_of_friends fof ON fof.user_id = u.id
			WHERE u.id NOT IN (SELECT user_id FROM excluded)
			ORDER BY 
				COALESCE(fof.mutual_count, 0) DESC,
				pr.followers_count DESC,
				u.created_at DESC
			LIMIT $2;
		`
		rows, err := r.db.QueryContext(ctx, query, *viewerID, limit)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var u domain.User
				var mutualCount int
				var mutualNames []string
				err := rows.Scan(
					&u.ID, &u.Username, &u.Email, &u.IsVerified, &u.VerifiedType, &u.CreatedAt,
					&u.DisplayName, &u.AvatarURL, &u.Bio, &u.Location, &u.Website,
					&u.FollowersCount, &u.FollowingCount, &u.IsPrivate,
					&mutualCount, pq.Array(&mutualNames),
				)
				if err == nil {
					reason := "Suggested for you"
					if mutualCount > 0 && len(mutualNames) > 0 {
						if mutualCount == 1 {
							reason = fmt.Sprintf("Followed by @%s", mutualNames[0])
						} else {
							reason = fmt.Sprintf("Followed by @%s + %d others", mutualNames[0], mutualCount-1)
						}
					} else if u.FollowersCount > 50 {
						reason = "Popular on X"
					}

					score := float64(mutualCount)*3.0 + float64(u.FollowersCount)*0.01
					recommendations = append(recommendations, domain.UserRecommendation{
						User:        u,
						Reason:      reason,
						Score:       score,
						MutualCount: mutualCount,
						MutualNames: mutualNames,
					})
				}
			}
		}
	}

	// 2. Fallback / Unauthenticated popular discovery
	if len(recommendations) < limit {
		remaining := limit - len(recommendations)
		existingIDs := make(map[uuid.UUID]bool)
		if viewerID != nil {
			existingIDs[*viewerID] = true
		}
		for _, r := range recommendations {
			existingIDs[r.User.ID] = true
		}

		query := `
			SELECT 
				u.id, u.username, u.email, u.is_verified, u.verified_type, u.created_at,
				pr.display_name, pr.avatar_url, pr.bio, pr.location, pr.website,
				pr.followers_count, pr.following_count, pr.is_private
			FROM users u
			JOIN user_profiles pr ON pr.user_id = u.id
			WHERE pr.is_private = FALSE
			ORDER BY pr.followers_count DESC, u.created_at DESC
			LIMIT $1;
		`
		rows, err := r.db.QueryContext(ctx, query, remaining*2)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				if len(recommendations) >= limit {
					break
				}
				var u domain.User
				if err := rows.Scan(
					&u.ID, &u.Username, &u.Email, &u.IsVerified, &u.VerifiedType, &u.CreatedAt,
					&u.DisplayName, &u.AvatarURL, &u.Bio, &u.Location, &u.Website,
					&u.FollowersCount, &u.FollowingCount, &u.IsPrivate,
				); err == nil {
					if !existingIDs[u.ID] {
						existingIDs[u.ID] = true
						recommendations = append(recommendations, domain.UserRecommendation{
							User:        u,
							Reason:      "Popular on X",
							Score:       float64(u.FollowersCount),
							MutualCount: 0,
						})
					}
				}
			}
		}
	}

	if recommendations == nil {
		recommendations = []domain.UserRecommendation{}
	}
	return recommendations, nil
}

func (r *PostgresRepository) DismissUserRecommendation(ctx context.Context, viewerID, targetID uuid.UUID, duration time.Duration) error {
	if duration <= 0 {
		duration = 7 * 24 * time.Hour // default 7 days
	}
	dismissedUntil := time.Now().Add(duration)

	query := `
		INSERT INTO dismissed_user_recommendations (user_id, target_user_id, dismissed_until, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id, target_user_id)
		DO UPDATE SET dismissed_until = EXCLUDED.dismissed_until, created_at = NOW();
	`
	_, err := r.db.ExecContext(ctx, query, viewerID, targetID, dismissedUntil)
	return err
}

func (r *PostgresRepository) ReconcileCounters(ctx context.Context, dryRun bool, batchSize int) (*domain.CounterReconciliationResult, error) {
	if batchSize <= 0 || batchSize > 1000 {
		batchSize = 250
	}

	result := &domain.CounterReconciliationResult{
		DryRun:  dryRun,
		Details: []domain.CounterMismatchDetail{},
	}

	query := `
		SELECT p.id,
		       p.likes_count,
		       (SELECT COUNT(*) FROM post_likes pl WHERE pl.post_id = p.id) as actual_likes,
		       p.reposts_count,
		       (SELECT COUNT(*) FROM post_reposts prp WHERE prp.post_id = p.id) as actual_reposts,
		       p.replies_count,
		       (SELECT COUNT(*) FROM posts rep WHERE rep.reply_to_id = p.id AND rep.deleted_at IS NULL) as actual_replies,
		       COALESCE(p.quotes_count, 0) as reported_quotes,
		       (SELECT COUNT(*) FROM posts q WHERE q.quote_post_id = p.id AND q.deleted_at IS NULL) as actual_quotes
		FROM posts p
		WHERE p.deleted_at IS NULL
		ORDER BY p.created_at DESC
		LIMIT $1;
	`
	rows, err := r.db.QueryContext(ctx, query, batchSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var postID uuid.UUID
		var reportedLikes, actualLikes int
		var reportedReposts, actualReposts int
		var reportedReplies, actualReplies int
		var reportedQuotes, actualQuotes int

		err := rows.Scan(
			&postID,
			&reportedLikes, &actualLikes,
			&reportedReposts, &actualReposts,
			&reportedReplies, &actualReplies,
			&reportedQuotes, &actualQuotes,
		)
		if err != nil {
			return nil, err
		}

		result.TotalScanned++

		hasMismatch := false
		if reportedLikes != actualLikes {
			result.MismatchesFound++
			hasMismatch = true
			result.Details = append(result.Details, domain.CounterMismatchDetail{
				PostID:        postID,
				Type:          "likes",
				ReportedCount: reportedLikes,
				ActualCount:   actualLikes,
				Fixed:         !dryRun,
			})
		}
		if reportedReposts != actualReposts {
			result.MismatchesFound++
			hasMismatch = true
			result.Details = append(result.Details, domain.CounterMismatchDetail{
				PostID:        postID,
				Type:          "reposts",
				ReportedCount: reportedReposts,
				ActualCount:   actualReposts,
				Fixed:         !dryRun,
			})
		}
		if reportedReplies != actualReplies {
			result.MismatchesFound++
			hasMismatch = true
			result.Details = append(result.Details, domain.CounterMismatchDetail{
				PostID:        postID,
				Type:          "replies",
				ReportedCount: reportedReplies,
				ActualCount:   actualReplies,
				Fixed:         !dryRun,
			})
		}
		if reportedQuotes != actualQuotes {
			result.MismatchesFound++
			hasMismatch = true
			result.Details = append(result.Details, domain.CounterMismatchDetail{
				PostID:        postID,
				Type:          "quotes",
				ReportedCount: reportedQuotes,
				ActualCount:   actualQuotes,
				Fixed:         !dryRun,
			})
		}

		if hasMismatch && !dryRun {
			_, updateErr := r.db.ExecContext(ctx, `
				UPDATE posts
				SET likes_count = $2,
				    reposts_count = $3,
				    replies_count = $4,
				    quotes_count = $5
				WHERE id = $1;
			`, postID, actualLikes, actualReposts, actualReplies, actualQuotes)
			if updateErr == nil {
				result.MismatchesFixed++
			}
		}
	}

	return result, nil
}



