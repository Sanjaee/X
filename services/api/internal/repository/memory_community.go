package repository

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/x-reconstruction/api/internal/domain"
)

// -------------------------------------------------------------
// Community & Community Relationships In-Memory Implementation
// -------------------------------------------------------------

func (r *MemoryRepository) CreateCommunity(ctx context.Context, c *domain.Community, rules []domain.CommunityRule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, existing := range r.communities {
		if strings.EqualFold(existing.Slug, c.Slug) {
			return ErrAlreadyExists
		}
	}

	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	now := time.Now().UTC()
	c.CreatedAt = now
	c.UpdatedAt = now
	c.MembersCount = 1

	r.communities[c.ID] = c

	// Creator is Owner member
	ownerMember := &domain.CommunityMember{
		ID:          uuid.New(),
		CommunityID: c.ID,
		UserID:      c.OwnerID,
		Role:        "owner",
		Status:      "active",
		JoinedAt:    now,
	}
	r.communityMembers[fmt.Sprintf("%s:%s", c.ID, c.OwnerID)] = ownerMember

	// Rules
	createdRules := make([]domain.CommunityRule, len(rules))
	for i, rule := range rules {
		rule.ID = uuid.New()
		rule.CommunityID = c.ID
		rule.RuleNumber = i + 1
		rule.CreatedAt = now
		createdRules[i] = rule
	}
	r.communityRules[c.ID] = createdRules

	return nil
}

func (r *MemoryRepository) GetCommunityByID(ctx context.Context, id uuid.UUID, viewerID *uuid.UUID) (*domain.Community, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.communities[id]
	if !ok {
		return nil, ErrNotFound
	}
	copyComm := *c
	if viewerID != nil {
		if mem, exists := r.communityMembers[fmt.Sprintf("%s:%s", c.ID, *viewerID)]; exists {
			copyComm.ViewerRole = mem.Role
			copyComm.ViewerStatus = mem.Status
		}
	}
	return &copyComm, nil
}

func (r *MemoryRepository) GetCommunityBySlug(ctx context.Context, slug string, viewerID *uuid.UUID) (*domain.Community, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, c := range r.communities {
		if strings.EqualFold(c.Slug, slug) {
			copyComm := *c
			if viewerID != nil {
				if mem, exists := r.communityMembers[fmt.Sprintf("%s:%s", c.ID, *viewerID)]; exists {
					copyComm.ViewerRole = mem.Role
					copyComm.ViewerStatus = mem.Status
				}
			}
			return &copyComm, nil
		}
	}
	return nil, ErrNotFound
}

func (r *MemoryRepository) GetCommunities(ctx context.Context, search string, limit, offset int, viewerID *uuid.UUID) ([]domain.Community, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matched []domain.Community
	q := strings.ToLower(strings.TrimSpace(search))

	for _, c := range r.communities {
		if c.Visibility == "unlisted" {
			// Only visible if viewer is member
			if viewerID == nil {
				continue
			}
			if _, isMem := r.communityMembers[fmt.Sprintf("%s:%s", c.ID, *viewerID)]; !isMem {
				continue
			}
		}
		if q != "" {
			nameLower := strings.ToLower(c.Name)
			descLower := strings.ToLower(c.Description)
			if !strings.Contains(nameLower, q) && !strings.Contains(descLower, q) {
				continue
			}
		}
		copyComm := *c
		if viewerID != nil {
			if mem, exists := r.communityMembers[fmt.Sprintf("%s:%s", c.ID, *viewerID)]; exists {
				copyComm.ViewerRole = mem.Role
				copyComm.ViewerStatus = mem.Status
			}
		}
		matched = append(matched, copyComm)
	}

	sort.Slice(matched, func(i, j int) bool {
		return matched[i].MembersCount > matched[j].MembersCount
	})

	total := len(matched)
	if offset >= total {
		return []domain.Community{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return matched[offset:end], total, nil
}

func (r *MemoryRepository) GetUserCommunities(ctx context.Context, userID uuid.UUID) ([]domain.Community, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var res []domain.Community
	for key, mem := range r.communityMembers {
		if strings.HasSuffix(key, ":"+userID.String()) && mem.Status == "active" {
			if c, ok := r.communities[mem.CommunityID]; ok {
				copyComm := *c
				copyComm.ViewerRole = mem.Role
				copyComm.ViewerStatus = mem.Status
				res = append(res, copyComm)
			}
		}
	}
	return res, nil
}

func (r *MemoryRepository) UpdateCommunity(ctx context.Context, c *domain.Community) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.communities[c.ID]
	if !ok {
		return ErrNotFound
	}
	existing.Name = c.Name
	existing.Description = c.Description
	existing.AvatarURL = c.AvatarURL
	existing.BannerURL = c.BannerURL
	existing.Visibility = c.Visibility
	existing.AccessModel = c.AccessModel
	existing.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *MemoryRepository) DeleteCommunity(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.communities, id)
	delete(r.communityRules, id)
	for key := range r.communityMembers {
		if strings.HasPrefix(key, id.String()+":") {
			delete(r.communityMembers, key)
		}
	}
	return nil
}

func (r *MemoryRepository) GetCommunityMember(ctx context.Context, commID, userID uuid.UUID) (*domain.CommunityMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	m, ok := r.communityMembers[fmt.Sprintf("%s:%s", commID, userID)]
	if !ok {
		return nil, ErrNotFound
	}
	copyMem := *m
	return &copyMem, nil
}

func (r *MemoryRepository) AddCommunityMember(ctx context.Context, m *domain.CommunityMember) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%s:%s", m.CommunityID, m.UserID)
	if _, exists := r.communityMembers[key]; exists {
		return ErrAlreadyExists
	}
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if m.JoinedAt.IsZero() {
		m.JoinedAt = time.Now().UTC()
	}
	r.communityMembers[key] = m

	if comm, ok := r.communities[m.CommunityID]; ok {
		comm.MembersCount++
	}
	return nil
}

func (r *MemoryRepository) RemoveCommunityMember(ctx context.Context, commID, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%s:%s", commID, userID)
	if _, exists := r.communityMembers[key]; !exists {
		return ErrNotFound
	}
	delete(r.communityMembers, key)

	if comm, ok := r.communities[commID]; ok && comm.MembersCount > 0 {
		comm.MembersCount--
	}
	return nil
}

func (r *MemoryRepository) UpdateCommunityMemberRole(ctx context.Context, commID, userID uuid.UUID, role string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%s:%s", commID, userID)
	m, ok := r.communityMembers[key]
	if !ok {
		return ErrNotFound
	}
	m.Role = role
	return nil
}

func (r *MemoryRepository) UpdateCommunityMemberStatus(ctx context.Context, commID, userID uuid.UUID, status string, suspendedUntil *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%s:%s", commID, userID)
	m, ok := r.communityMembers[key]
	if !ok {
		return ErrNotFound
	}
	m.Status = status
	m.SuspendedUntil = suspendedUntil
	return nil
}

func (r *MemoryRepository) GetCommunityMembers(ctx context.Context, commID uuid.UUID, limit, offset int) ([]domain.CommunityMember, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var members []domain.CommunityMember
	for key, m := range r.communityMembers {
		if strings.HasPrefix(key, commID.String()+":") {
			copyMem := *m
			if u, ok := r.users[m.UserID]; ok {
				copyMem.User = &domain.User{
					ID:          u.ID,
					Username:    u.Username,
					DisplayName: u.DisplayName,
					AvatarURL:   u.AvatarURL,
					IsVerified:  u.IsVerified,
				}
			}
			members = append(members, copyMem)
		}
	}

	sort.Slice(members, func(i, j int) bool {
		return members[i].JoinedAt.Before(members[j].JoinedAt)
	})

	total := len(members)
	if offset >= total {
		return []domain.CommunityMember{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return members[offset:end], total, nil
}

func (r *MemoryRepository) CreateCommunityJoinRequest(ctx context.Context, req *domain.CommunityJoinRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if req.ID == uuid.Nil {
		req.ID = uuid.New()
	}
	req.CreatedAt = time.Now().UTC()
	req.Status = "pending"
	r.communityJoinReqs[req.ID] = req
	return nil
}

func (r *MemoryRepository) GetCommunityJoinRequests(ctx context.Context, commID uuid.UUID, status string, limit, offset int) ([]domain.CommunityJoinRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var res []domain.CommunityJoinRequest
	for _, req := range r.communityJoinReqs {
		if req.CommunityID == commID && (status == "" || req.Status == status) {
			copyReq := *req
			if u, ok := r.users[req.UserID]; ok {
				copyReq.User = &domain.User{
					ID:          u.ID,
					Username:    u.Username,
					DisplayName: u.DisplayName,
					AvatarURL:   u.AvatarURL,
				}
			}
			res = append(res, copyReq)
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].CreatedAt.After(res[j].CreatedAt)
	})
	if offset >= len(res) {
		return []domain.CommunityJoinRequest{}, nil
	}
	end := offset + limit
	if end > len(res) {
		end = len(res)
	}
	return res[offset:end], nil
}

func (r *MemoryRepository) ReviewCommunityJoinRequest(ctx context.Context, reqID, reviewerID uuid.UUID, status, note string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	req, ok := r.communityJoinReqs[reqID]
	if !ok {
		return ErrNotFound
	}
	req.Status = status
	req.ReviewedBy = &reviewerID
	now := time.Now().UTC()
	req.ReviewedAt = &now

	if status == "approved" {
		memKey := fmt.Sprintf("%s:%s", req.CommunityID, req.UserID)
		if _, exists := r.communityMembers[memKey]; !exists {
			r.communityMembers[memKey] = &domain.CommunityMember{
				ID:          uuid.New(),
				CommunityID: req.CommunityID,
				UserID:      req.UserID,
				Role:        "member",
				Status:      "active",
				JoinedAt:    now,
			}
			if comm, ok := r.communities[req.CommunityID]; ok {
				comm.MembersCount++
			}
		}
	}
	return nil
}

func (r *MemoryRepository) CancelCommunityJoinRequest(ctx context.Context, commID, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id, req := range r.communityJoinReqs {
		if req.CommunityID == commID && req.UserID == userID && req.Status == "pending" {
			delete(r.communityJoinReqs, id)
			return nil
		}
	}
	return ErrNotFound
}

func (r *MemoryRepository) GetCommunityRules(ctx context.Context, commID uuid.UUID) ([]domain.CommunityRule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rules, ok := r.communityRules[commID]
	if !ok {
		return []domain.CommunityRule{}, nil
	}
	cpy := make([]domain.CommunityRule, len(rules))
	copy(cpy, rules)
	return cpy, nil
}

func (r *MemoryRepository) LinkPostToCommunity(ctx context.Context, postID, commID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.communityPosts[postID] = commID
	return nil
}

func (r *MemoryRepository) GetCommunityPosts(ctx context.Context, commID uuid.UUID, limit, offset int, viewerID *uuid.UUID) ([]domain.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var posts []domain.Post
	for pID, cID := range r.communityPosts {
		if cID == commID {
			p, ok := r.posts[pID]
			if !ok || p.DeletedAt != nil {
				continue
			}
			// Block filtering
			if viewerID != nil {
				bKey1 := fmt.Sprintf("%s:%s", *viewerID, p.AuthorID)
				bKey2 := fmt.Sprintf("%s:%s", p.AuthorID, *viewerID)
				if _, blocked := r.blocks[bKey1]; blocked {
					continue
				}
				if _, blocked := r.blocks[bKey2]; blocked {
					continue
				}
			}
			posts = append(posts, *p)
		}
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].CreatedAt.After(posts[j].CreatedAt)
	})

	if offset >= len(posts) {
		return []domain.Post{}, nil
	}
	end := offset + limit
	if end > len(posts) {
		end = len(posts)
	}
	return posts[offset:end], nil
}

func (r *MemoryRepository) PinCommunityPost(ctx context.Context, commID, postID, modID uuid.UUID, isPinned bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%s:%s", commID, postID)
	if isPinned {
		r.pinnedPosts[key] = true
	} else {
		delete(r.pinnedPosts, key)
	}
	return nil
}
