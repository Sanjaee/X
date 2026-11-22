package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/x-reconstruction/api/internal/domain"
)

// CreateCommunity creates a new community, initializes rules, and sets owner as first member
func (r *PostgresRepository) CreateCommunity(ctx context.Context, c *domain.Community, rules []domain.CommunityRule) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	c.MembersCount = 1
	c.PostsCount = 0

	query := `
		INSERT INTO communities (
			id, slug, name, description, avatar_url, banner_url,
			access_model, visibility, owner_id, members_count, posts_count,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11,
			NOW(), NOW()
		);
	`
	_, err = tx.ExecContext(ctx, query,
		c.ID, c.Slug, c.Name, c.Description, c.AvatarURL, c.BannerURL,
		c.AccessModel, c.Visibility, c.OwnerID, c.MembersCount, c.PostsCount,
	)
	if err != nil {
		return err
	}

	// Insert owner as community_member with role 'owner'
	memQuery := `
		INSERT INTO community_members (
			id, community_id, user_id, role, status, joined_at
		) VALUES (
			gen_random_uuid(), $1, $2, 'owner', 'active', NOW()
		);
	`
	if _, err = tx.ExecContext(ctx, memQuery, c.ID, c.OwnerID); err != nil {
		return err
	}

	// Insert rules
	for i, rule := range rules {
		ruleNum := rule.RuleNumber
		if ruleNum <= 0 {
			ruleNum = i + 1
		}
		ruleQuery := `
			INSERT INTO community_rules (
				id, community_id, rule_number, title, description, created_at
			) VALUES (
				gen_random_uuid(), $1, $2, $3, $4, NOW()
			);
		`
		if _, err = tx.ExecContext(ctx, ruleQuery, c.ID, ruleNum, rule.Title, rule.Description); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) GetCommunityByID(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID) (*domain.Community, error) {
	query := `
		SELECT c.id, c.slug, c.name, c.description, c.avatar_url, c.banner_url,
		       c.access_model, c.visibility, c.owner_id, c.members_count, c.posts_count,
		       c.created_at, c.updated_at, c.deleted_at,
		       u.username, COALESCE(p.display_name, u.username), COALESCE(p.avatar_url, '')
		FROM communities c
		JOIN users u ON c.owner_id = u.id
		LEFT JOIN user_profiles p ON u.id = p.user_id
		WHERE c.id = $1 AND c.deleted_at IS NULL;
	`
	var c domain.Community
	var owner domain.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.Slug, &c.Name, &c.Description, &c.AvatarURL, &c.BannerURL,
		&c.AccessModel, &c.Visibility, &c.OwnerID, &c.MembersCount, &c.PostsCount,
		&c.CreatedAt, &c.UpdatedAt, &c.DeletedAt,
		&owner.Username, &owner.DisplayName, &owner.AvatarURL,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	owner.ID = c.OwnerID
	c.Owner = &owner

	// Check viewer relationship if authenticated
	if viewerID != nil {
		_ = r.db.QueryRowContext(ctx, `
			SELECT role, status FROM community_members WHERE community_id = $1 AND user_id = $2;
		`, c.ID, *viewerID).Scan(&c.ViewerRole, &c.ViewerStatus)

		if c.ViewerStatus == "" {
			// Check if pending join request
			var reqStatus string
			_ = r.db.QueryRowContext(ctx, `
				SELECT status FROM community_join_requests WHERE community_id = $1 AND user_id = $2 AND status = 'pending';
			`, c.ID, *viewerID).Scan(&reqStatus)
			if reqStatus == "pending" {
				c.ViewerStatus = "pending"
			}
		}
	}

	return &c, nil
}

func (r *PostgresRepository) GetCommunityBySlug(ctx context.Context, slug string, viewerID *uuid.UUID) (*domain.Community, error) {
	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, `SELECT id FROM communities WHERE slug = $1 AND deleted_at IS NULL;`, strings.ToLower(slug)).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return r.GetCommunityByID(ctx, id, viewerID)
}

func (r *PostgresRepository) GetCommunities(ctx context.Context, search string, limit, offset int, viewerID *uuid.UUID) ([]domain.Community, int, error) {
	if limit <= 0 {
		limit = 20
	}
	where := "WHERE c.deleted_at IS NULL AND c.visibility = 'public'"
	var args []interface{}
	idx := 1

	if search != "" {
		where += fmt.Sprintf(" AND (c.name ILIKE $%d OR c.slug ILIKE $%d OR c.description ILIKE $%d)", idx, idx, idx)
		args = append(args, "%"+search+"%")
		idx++
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM communities c %s;", where)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT c.id, c.slug, c.name, c.description, c.avatar_url, c.banner_url,
		       c.access_model, c.visibility, c.owner_id, c.members_count, c.posts_count,
		       c.created_at, c.updated_at,
		       u.username, COALESCE(p.display_name, u.username), COALESCE(p.avatar_url, '')
		FROM communities c
		JOIN users u ON c.owner_id = u.id
		LEFT JOIN user_profiles p ON u.id = p.user_id
		%s
		ORDER BY c.members_count DESC, c.created_at DESC
		LIMIT $%d OFFSET $%d;
	`, where, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []domain.Community
	for rows.Next() {
		var c domain.Community
		var owner domain.User
		if err := rows.Scan(
			&c.ID, &c.Slug, &c.Name, &c.Description, &c.AvatarURL, &c.BannerURL,
			&c.AccessModel, &c.Visibility, &c.OwnerID, &c.MembersCount, &c.PostsCount,
			&c.CreatedAt, &c.UpdatedAt,
			&owner.Username, &owner.DisplayName, &owner.AvatarURL,
		); err != nil {
			return nil, 0, err
		}
		owner.ID = c.OwnerID
		c.Owner = &owner
		list = append(list, c)
	}

	// Enrich with viewer membership if viewerID present
	if viewerID != nil && len(list) > 0 {
		for i := range list {
			_ = r.db.QueryRowContext(ctx, `
				SELECT role, status FROM community_members WHERE community_id = $1 AND user_id = $2;
			`, list[i].ID, *viewerID).Scan(&list[i].ViewerRole, &list[i].ViewerStatus)
		}
	}

	return list, total, nil
}

func (r *PostgresRepository) GetUserCommunities(ctx context.Context, userID uuid.UUID) ([]domain.Community, error) {
	query := `
		SELECT c.id, c.slug, c.name, c.description, c.avatar_url, c.banner_url,
		       c.access_model, c.visibility, c.owner_id, c.members_count, c.posts_count,
		       c.created_at, c.updated_at, m.role, m.status
		FROM community_members m
		JOIN communities c ON m.community_id = c.id
		WHERE m.user_id = $1 AND m.status = 'active' AND c.deleted_at IS NULL
		ORDER BY m.joined_at DESC;
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.Community
	for rows.Next() {
		var c domain.Community
		if err := rows.Scan(
			&c.ID, &c.Slug, &c.Name, &c.Description, &c.AvatarURL, &c.BannerURL,
			&c.AccessModel, &c.Visibility, &c.OwnerID, &c.MembersCount, &c.PostsCount,
			&c.CreatedAt, &c.UpdatedAt, &c.ViewerRole, &c.ViewerStatus,
		); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

func (r *PostgresRepository) UpdateCommunity(ctx context.Context, c *domain.Community) error {
	query := `
		UPDATE communities SET
			name = $2,
			description = $3,
			avatar_url = $4,
			banner_url = $5,
			access_model = $6,
			visibility = $7,
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL;
	`
	_, err := r.db.ExecContext(ctx, query,
		c.ID, c.Name, c.Description, c.AvatarURL, c.BannerURL,
		c.AccessModel, c.Visibility,
	)
	return err
}

func (r *PostgresRepository) DeleteCommunity(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE communities SET deleted_at = NOW() WHERE id = $1;`, id)
	return err
}

// ==========================================
// MEMBERSHIP OPERATIONS
// ==========================================

func (r *PostgresRepository) GetCommunityMember(ctx context.Context, commID, userID uuid.UUID) (*domain.CommunityMember, error) {
	query := `
		SELECT id, community_id, user_id, role, status, suspended_until, joined_at
		FROM community_members
		WHERE community_id = $1 AND user_id = $2;
	`
	var m domain.CommunityMember
	err := r.db.QueryRowContext(ctx, query, commID, userID).Scan(
		&m.ID, &m.CommunityID, &m.UserID, &m.Role, &m.Status, &m.SuspendedUntil, &m.JoinedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *PostgresRepository) AddCommunityMember(ctx context.Context, m *domain.CommunityMember) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if m.Role == "" {
		m.Role = "member"
	}
	if m.Status == "" {
		m.Status = "active"
	}

	query := `
		INSERT INTO community_members (
			id, community_id, user_id, role, status, joined_at
		) VALUES (
			$1, $2, $3, $4, $5, NOW()
		)
		ON CONFLICT (community_id, user_id) DO UPDATE SET
			status = 'active',
			suspended_until = NULL;
	`
	if _, err := tx.ExecContext(ctx, query, m.ID, m.CommunityID, m.UserID, m.Role, m.Status); err != nil {
		return err
	}

	// Increment members_count
	_, err = tx.ExecContext(ctx, `
		UPDATE communities SET members_count = (
			SELECT COUNT(*) FROM community_members WHERE community_id = $1 AND status = 'active'
		) WHERE id = $1;
	`, m.CommunityID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresRepository) RemoveCommunityMember(ctx context.Context, commID, userID uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `DELETE FROM community_members WHERE community_id = $1 AND user_id = $2;`, commID, userID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE communities SET members_count = (
			SELECT COUNT(*) FROM community_members WHERE community_id = $1 AND status = 'active'
		) WHERE id = $1;
	`, commID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresRepository) UpdateCommunityMemberRole(ctx context.Context, commID, userID uuid.UUID, role string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE community_members SET role = $3 WHERE community_id = $1 AND user_id = $2;
	`, commID, userID, role)
	return err
}

func (r *PostgresRepository) UpdateCommunityMemberStatus(ctx context.Context, commID, userID uuid.UUID, status string, suspendedUntil *time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		UPDATE community_members SET status = $3, suspended_until = $4 WHERE community_id = $1 AND user_id = $2;
	`, commID, userID, status, suspendedUntil)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE communities SET members_count = (
			SELECT COUNT(*) FROM community_members WHERE community_id = $1 AND status = 'active'
		) WHERE id = $1;
	`, commID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresRepository) GetCommunityMembers(ctx context.Context, commID uuid.UUID, limit, offset int) ([]domain.CommunityMember, int, error) {
	if limit <= 0 {
		limit = 20
	}
	var total int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM community_members WHERE community_id = $1;
	`, commID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT m.id, m.community_id, m.user_id, m.role, m.status, m.suspended_until, m.joined_at,
		       u.username, u.role, u.is_verified,
		       COALESCE(p.display_name, u.username), COALESCE(p.avatar_url, ''), COALESCE(p.name_style, '')
		FROM community_members m
		JOIN users u ON m.user_id = u.id
		LEFT JOIN user_profiles p ON u.id = p.user_id
		WHERE m.community_id = $1
		ORDER BY 
			CASE m.role 
				WHEN 'owner' THEN 1 
				WHEN 'admin' THEN 2 
				WHEN 'moderator' THEN 3 
				ELSE 4 
			END,
			m.joined_at ASC
		LIMIT $2 OFFSET $3;
	`
	rows, err := r.db.QueryContext(ctx, query, commID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []domain.CommunityMember
	for rows.Next() {
		var m domain.CommunityMember
		var u domain.User
		if err := rows.Scan(
			&m.ID, &m.CommunityID, &m.UserID, &m.Role, &m.Status, &m.SuspendedUntil, &m.JoinedAt,
			&u.Username, &u.Role, &u.IsVerified,
			&u.DisplayName, &u.AvatarURL, &u.NameStyle,
		); err != nil {
			return nil, 0, err
		}
		u.ID = m.UserID
		m.User = &u
		list = append(list, m)
	}

	return list, total, nil
}

// ==========================================
// JOIN REQUESTS & INVITES
// ==========================================

func (r *PostgresRepository) CreateCommunityJoinRequest(ctx context.Context, req *domain.CommunityJoinRequest) error {
	if req.ID == uuid.Nil {
		req.ID = uuid.New()
	}
	query := `
		INSERT INTO community_join_requests (
			id, community_id, user_id, status, response_note, created_at
		) VALUES (
			$1, $2, $3, 'pending', $4, NOW()
		);
	`
	_, err := r.db.ExecContext(ctx, query, req.ID, req.CommunityID, req.UserID, req.ResponseNote)
	return err
}

func (r *PostgresRepository) GetCommunityJoinRequests(ctx context.Context, commID uuid.UUID, status string, limit, offset int) ([]domain.CommunityJoinRequest, error) {
	if limit <= 0 {
		limit = 20
	}
	if status == "" {
		status = "pending"
	}
	query := `
		SELECT r.id, r.community_id, r.user_id, r.status, r.response_note, r.reviewed_at, r.reviewed_by, r.created_at,
		       u.username, COALESCE(p.display_name, u.username), COALESCE(p.avatar_url, '')
		FROM community_join_requests r
		JOIN users u ON r.user_id = u.id
		LEFT JOIN user_profiles p ON u.id = p.user_id
		WHERE r.community_id = $1 AND r.status = $2
		ORDER BY r.created_at ASC
		LIMIT $3 OFFSET $4;
	`
	rows, err := r.db.QueryContext(ctx, query, commID, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.CommunityJoinRequest
	for rows.Next() {
		var req domain.CommunityJoinRequest
		var u domain.User
		if err := rows.Scan(
			&req.ID, &req.CommunityID, &req.UserID, &req.Status, &req.ResponseNote, &req.ReviewedAt, &req.ReviewedBy, &req.CreatedAt,
			&u.Username, &u.DisplayName, &u.AvatarURL,
		); err != nil {
			return nil, err
		}
		u.ID = req.UserID
		req.User = &u
		list = append(list, req)
	}

	return list, nil
}

func (r *PostgresRepository) ReviewCommunityJoinRequest(ctx context.Context, reqID, reviewerID uuid.UUID, status, note string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var commID, userID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		SELECT community_id, user_id FROM community_join_requests WHERE id = $1 AND status = 'pending';
	`, reqID).Scan(&commID, &userID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE community_join_requests SET
			status = $2,
			response_note = $3,
			reviewed_at = NOW(),
			reviewed_by = $4
		WHERE id = $1;
	`, reqID, status, note, reviewerID)
	if err != nil {
		return err
	}

	if status == "approved" {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO community_members (
				id, community_id, user_id, role, status, joined_at
			) VALUES (
				gen_random_uuid(), $1, $2, 'member', 'active', NOW()
			)
			ON CONFLICT (community_id, user_id) DO UPDATE SET
				status = 'active';
		`, commID, userID)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE communities SET members_count = (
				SELECT COUNT(*) FROM community_members WHERE community_id = $1 AND status = 'active'
			) WHERE id = $1;
		`, commID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) CancelCommunityJoinRequest(ctx context.Context, commID, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE community_join_requests SET status = 'cancelled'
		WHERE community_id = $1 AND user_id = $2 AND status = 'pending';
	`, commID, userID)
	return err
}

// ==========================================
// RULES & POSTS
// ==========================================

func (r *PostgresRepository) GetCommunityRules(ctx context.Context, commID uuid.UUID) ([]domain.CommunityRule, error) {
	query := `
		SELECT id, community_id, rule_number, title, description, created_at
		FROM community_rules
		WHERE community_id = $1
		ORDER BY rule_number ASC;
	`
	rows, err := r.db.QueryContext(ctx, query, commID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []domain.CommunityRule
	for rows.Next() {
		var rule domain.CommunityRule
		if err := rows.Scan(&rule.ID, &rule.CommunityID, &rule.RuleNumber, &rule.Title, &rule.Description, &rule.CreatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (r *PostgresRepository) LinkPostToCommunity(ctx context.Context, postID, commID uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO community_posts (id, post_id, community_id, created_at)
		VALUES (gen_random_uuid(), $1, $2, NOW())
		ON CONFLICT (post_id) DO NOTHING;
	`, postID, commID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE communities SET posts_count = (
			SELECT COUNT(*) FROM community_posts WHERE community_id = $1
		) WHERE id = $1;
	`, commID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresRepository) GetCommunityPosts(ctx context.Context, commID uuid.UUID, limit, offset int, viewerID *uuid.UUID) ([]domain.Post, error) {
	if limit <= 0 {
		limit = 20
	}
	// Query posts linked to community, ordering pinned posts first
	query := `
		SELECT p.id, p.author_id, p.content, p.visibility, p.moderation_status,
		       p.replies_count, p.reposts_count, p.likes_count, p.views_count,
		       p.created_at, p.updated_at,
		       u.username, u.role, u.is_verified, u.verified_type,
		       COALESCE(prof.display_name, u.username), COALESCE(prof.avatar_url, ''), COALESCE(prof.name_style, ''),
		       cp.is_pinned,
		       c.id, c.name, c.slug
		FROM community_posts cp
		JOIN communities c ON cp.community_id = c.id
		JOIN posts p ON cp.post_id = p.id
		JOIN users u ON p.author_id = u.id
		LEFT JOIN user_profiles prof ON u.id = prof.user_id
		WHERE cp.community_id = $1 AND p.deleted_at IS NULL AND p.moderation_status = 'active'
	`
	var args []interface{}
	args = append(args, commID)

	// Filter out blocked authors if viewer present
	if viewerID != nil {
		query += fmt.Sprintf(`
			AND p.author_id NOT IN (
				SELECT blocked_id FROM blocks WHERE blocker_id = $%d
				UNION
				SELECT blocker_id FROM blocks WHERE blocked_id = $%d
			)
		`, len(args)+1, len(args)+1)
		args = append(args, *viewerID)
	}

	query += fmt.Sprintf(`
		ORDER BY cp.is_pinned DESC, p.created_at DESC
		LIMIT $%d OFFSET $%d;
	`, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []domain.Post
	for rows.Next() {
		var p domain.Post
		var author domain.User
		var isPinned bool
		var cID uuid.UUID
		var cName, cSlug string
		if err := rows.Scan(
			&p.ID, &p.AuthorID, &p.Content, &p.Visibility, &p.ModerationStatus,
			&p.RepliesCount, &p.RepostsCount, &p.LikesCount, &p.ViewsCount,
			&p.CreatedAt, &p.UpdatedAt,
			&author.Username, &author.Role, &author.IsVerified, &author.VerifiedType,
			&author.DisplayName, &author.AvatarURL, &author.NameStyle,
			&isPinned,
			&cID, &cName, &cSlug,
		); err != nil {
			return nil, err
		}
		author.ID = p.AuthorID
		p.Author = &author
		p.CommunityID = &cID
		p.CommunityName = cName
		p.CommunitySlug = cSlug

		// Load media objects for post
		mRows, _ := r.db.QueryContext(ctx, `
			SELECT m.id, m.url, m.media_type, m.width, m.height, m.duration_seconds
			FROM post_media pm
			JOIN media_objects m ON pm.media_id = m.id
			WHERE pm.post_id = $1
			ORDER BY pm.position ASC;
		`, p.ID)
		if mRows != nil {
			for mRows.Next() {
				var mo domain.MediaObject
				if err := mRows.Scan(&mo.ID, &mo.URL, &mo.MediaType, &mo.Width, &mo.Height, &mo.DurationSeconds); err == nil {
					p.Media = append(p.Media, mo)
				}
			}
			mRows.Close()
		}

		// Viewer engagement flags
		if viewerID != nil {
			_ = r.db.QueryRowContext(ctx, `
				SELECT EXISTS(SELECT 1 FROM post_likes WHERE post_id = $1 AND user_id = $2);
			`, p.ID, *viewerID).Scan(&p.IsLiked)

			_ = r.db.QueryRowContext(ctx, `
				SELECT EXISTS(SELECT 1 FROM post_reposts WHERE post_id = $1 AND user_id = $2);
			`, p.ID, *viewerID).Scan(&p.IsReposted)

			_ = r.db.QueryRowContext(ctx, `
				SELECT EXISTS(SELECT 1 FROM post_bookmarks WHERE post_id = $1 AND user_id = $2);
			`, p.ID, *viewerID).Scan(&p.IsBookmarked)
		}

		posts = append(posts, p)
	}

	return posts, nil
}

func (r *PostgresRepository) PinCommunityPost(ctx context.Context, commID, postID, modID uuid.UUID, isPinned bool) error {
	query := `
		UPDATE community_posts SET
			is_pinned = $3,
			pinned_at = CASE WHEN $3 THEN NOW() ELSE NULL END,
			pinned_by = CASE WHEN $3 THEN $4 ELSE NULL END
		WHERE community_id = $1 AND post_id = $2;
	`
	_, err := r.db.ExecContext(ctx, query, commID, postID, isPinned, modID)
	return err
}
