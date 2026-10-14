package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/x-reconstruction/api/internal/domain"
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

	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(15 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return &PostgresRepository{db: db}, nil
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
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	var vID interface{} = uuid.Nil
	if viewerID != nil {
		vID = *viewerID
	}

	var query string
	if feedType == "following" && viewerID != nil {
		query = `
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
	} else {
		query = `
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
			WHERE p.deleted_at IS NULL
			ORDER BY p.created_at DESC, p.id DESC
			LIMIT $2;
		`
	}

	rows, err := r.db.QueryContext(ctx, query, vID, limit)
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

	return posts, "", nil
}

func (r *PostgresRepository) GetUserPosts(ctx context.Context, username, tab string, viewerID *uuid.UUID, cursor string, limit int) ([]domain.Post, string, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	var vID interface{} = uuid.Nil
	if viewerID != nil {
		vID = *viewerID
	}

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
		WHERE LOWER(u.username) = LOWER($1) AND p.deleted_at IS NULL
		ORDER BY p.created_at DESC, p.id DESC
		LIMIT $3;
	`
	rows, err := r.db.QueryContext(ctx, query, username, vID, limit)
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
	return posts, "", nil
}

func (r *PostgresRepository) GetPostReplies(ctx context.Context, postID uuid.UUID, viewerID *uuid.UUID, cursor string, limit int) ([]domain.Post, string, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	var vID interface{} = uuid.Nil
	if viewerID != nil {
		vID = *viewerID
	}

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
		WHERE p.reply_to_id = $1 AND p.deleted_at IS NULL
		ORDER BY p.created_at ASC
		LIMIT $3;
	`
	rows, err := r.db.QueryContext(ctx, query, postID, vID, limit)
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
	return posts, "", nil
}

func (r *PostgresRepository) LikePost(ctx context.Context, postID, userID uuid.UUID) (bool, int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, 0, err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `INSERT INTO post_likes (post_id, user_id, created_at) VALUES ($1, $2, NOW()) ON CONFLICT DO NOTHING;`, postID, userID)
	if err != nil {
		return false, 0, err
	}

	var count int
	err = tx.QueryRowContext(ctx, `UPDATE posts SET likes_count = likes_count + 1 WHERE id = $1 RETURNING likes_count;`, postID).Scan(&count)
	if err != nil {
		return false, 0, err
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
		err = tx.QueryRowContext(ctx, `UPDATE posts SET likes_count = GREATEST(likes_count - 1, 0) WHERE id = $1 RETURNING likes_count;`, postID).Scan(&count)
		if err != nil {
			return false, 0, err
		}
	} else {
		_ = tx.QueryRowContext(ctx, `SELECT likes_count FROM posts WHERE id = $1;`, postID).Scan(&count)
	}

	return false, count, tx.Commit()
}

func (r *PostgresRepository) RepostPost(ctx context.Context, postID, userID uuid.UUID) (bool, int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, 0, err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `INSERT INTO post_reposts (post_id, user_id, created_at) VALUES ($1, $2, NOW()) ON CONFLICT DO NOTHING;`, postID, userID)
	if err != nil {
		return false, 0, err
	}

	var count int
	err = tx.QueryRowContext(ctx, `UPDATE posts SET reposts_count = reposts_count + 1 WHERE id = $1 RETURNING reposts_count;`, postID).Scan(&count)
	if err != nil {
		return false, 0, err
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
		err = tx.QueryRowContext(ctx, `UPDATE posts SET reposts_count = GREATEST(reposts_count - 1, 0) WHERE id = $1 RETURNING reposts_count;`, postID).Scan(&count)
		if err != nil {
			return false, 0, err
		}
	} else {
		_ = tx.QueryRowContext(ctx, `SELECT reposts_count FROM posts WHERE id = $1;`, postID).Scan(&count)
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

func (r *PostgresRepository) SearchPosts(ctx context.Context, query string, viewerID *uuid.UUID, cursor string, limit int) ([]domain.Post, string, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	var vID interface{} = uuid.Nil
	if viewerID != nil {
		vID = *viewerID
	}

	sqlQuery := `
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
		WHERE p.content ILIKE '%' || $1 || '%' AND p.deleted_at IS NULL
		ORDER BY p.created_at DESC
		LIMIT $3;
	`
	rows, err := r.db.QueryContext(ctx, sqlQuery, query, vID, limit)
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
	return posts, "", nil
}

func (r *PostgresRepository) SearchUsers(ctx context.Context, query string, viewerID *uuid.UUID, limit int) ([]domain.User, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	var vID interface{} = uuid.Nil
	if viewerID != nil {
		vID = *viewerID
	}

	sqlQuery := `
		SELECT u.id, u.username, u.email, u.role, u.is_verified, u.verified_type, u.is_private, u.created_at, u.updated_at,
		       p.display_name, p.bio, p.avatar_url, p.banner_url, p.location, p.website,
		       p.followers_count, p.following_count, p.posts_count,
		       COALESCE((SELECT TRUE FROM follows f WHERE f.follower_id = $2 AND f.following_id = u.id), FALSE) as is_following
		FROM users u
		JOIN user_profiles p ON p.user_id = u.id
		WHERE (u.username ILIKE '%' || $1 || '%' OR p.display_name ILIKE '%' || $1 || '%') AND u.deleted_at IS NULL
		ORDER BY p.followers_count DESC
		LIMIT $3;
	`
	rows, err := r.db.QueryContext(ctx, sqlQuery, query, vID, limit)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *PostgresRepository) CreateNotification(ctx context.Context, n *domain.Notification) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	query := `
		INSERT INTO notifications (id, recipient_id, actor_id, type, post_id, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW());
	`
	_, err := r.db.ExecContext(ctx, query, n.ID, n.RecipientID, n.ActorID, n.Type, n.PostID, n.IsRead)
	return err
}

func (r *PostgresRepository) GetNotifications(ctx context.Context, userID uuid.UUID, cursor string, limit int) ([]domain.Notification, string, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	query := `
		SELECT n.id, n.recipient_id, n.actor_id, n.type, n.post_id, n.is_read, n.created_at,
		       u.username, u.is_verified, u.verified_type,
		       p.display_name, p.avatar_url
		FROM notifications n
		JOIN users u ON u.id = n.actor_id
		JOIN user_profiles p ON p.user_id = n.actor_id
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
		err := rows.Scan(
			&n.ID, &n.RecipientID, &n.ActorID, &n.Type, &n.PostID, &n.IsRead, &n.CreatedAt,
			&actor.Username, &actor.IsVerified, &actor.VerifiedType,
			&actor.DisplayName, &actor.AvatarURL,
		)
		if err != nil {
			return nil, "", err
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
	}

	_, err = tx.ExecContext(ctx, `UPDATE follow_requests SET status = $1, updated_at = NOW() WHERE id = $2;`, newStatus, req.ID)
	if err != nil {
		return err
	}

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
		SELECT c.id, c.is_group, c.name, c.created_at, c.updated_at,
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
		SELECT c.id, c.is_group, c.name, c.created_at, c.updated_at
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
		SELECT c.id, c.is_group, c.name, c.created_at, c.updated_at,
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

func (r *PostgresRepository) GetPendingOutboxEvents(ctx context.Context, batchSize int) ([]domain.OutboxEvent, error) {
	if batchSize <= 0 || batchSize > 100 {
		batchSize = 25
	}
	q := `
		SELECT id, event_type, aggregate_type, aggregate_id, payload, status, retry_count, created_at
		FROM outbox_events
		WHERE status = 'pending'
		ORDER BY created_at ASC
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
		if err := rows.Scan(&e.ID, &e.EventType, &e.AggregateType, &e.AggregateID, &payloadBytes, &e.Status, &e.RetryCount, &e.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(payloadBytes, &e.Payload)
		events = append(events, e)
	}
	return events, nil
}

func (r *PostgresRepository) MarkOutboxEventPublished(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE outbox_events SET status = 'published', published_at = NOW() WHERE id = $1;`, id)
	return err
}

