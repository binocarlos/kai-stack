package store

import (
	"context"
	"fmt"

	"github.com/binocarlos/kai-stack/api/pkg/auth"
	"github.com/binocarlos/kai-stack/api/pkg/types"
	"gorm.io/gorm"
)

// ProfileRepository persists the app's canonical users (the profiles table).
type ProfileRepository struct {
	*Repository[types.Profile]
}

func NewProfileRepository(db *gorm.DB) *ProfileRepository {
	return &ProfileRepository{
		Repository: NewRepository[types.Profile](db),
	}
}

// GetOrCreate resolves a verified identity to the app's canonical profile,
// creating it on first sight. Roles are written only on insert (from the
// identity's default roles); for a returning user only email/updated_at are
// refreshed, so the app stays the source of truth for authorization and a
// provider can never escalate a user's roles on login.
func (r *ProfileRepository) GetOrCreate(ctx context.Context, id *auth.Identity) (*types.Profile, error) {
	roles := types.Roles(id.Roles)
	if roles == nil {
		roles = types.Roles{}
	}

	var profile types.Profile
	err := r.db.WithContext(ctx).Raw(`
INSERT INTO profiles (auth_provider, auth_subject, email, roles)
VALUES (?, ?, ?, ?::jsonb)
ON CONFLICT (auth_provider, auth_subject) DO UPDATE
SET email = EXCLUDED.email,
    updated_at = now()
RETURNING id, auth_provider, auth_subject, email, roles, created_at, updated_at`,
		id.Provider, id.Subject, id.Email, roles,
	).Scan(&profile).Error
	if err != nil {
		return nil, fmt.Errorf("failed to upsert profile: %w", err)
	}
	if profile.ID == "" {
		return nil, fmt.Errorf("profile upsert returned no row")
	}
	return &profile, nil
}
