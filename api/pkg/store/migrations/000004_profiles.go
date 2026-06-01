package migrations

import "context"

// Profiles creates the app's canonical user table. Each row is owned by the
// app (id is our own UUID) and mapped to whichever auth provider proved the
// user's identity via the (auth_provider, auth_subject) pair. Keeping our own
// id - rather than referencing Supabase's auth.users directly - means we can
// add or swap auth providers later without rewriting everything that points at
// a user. roles is owned by the app (the provider never sets it after insert).
func Profiles(ctx context.Context, m Migrator) error {
	return m.ExecSQL(ctx, `
CREATE TABLE IF NOT EXISTS profiles (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    auth_provider varchar(50)  NOT NULL,
    auth_subject  varchar(255) NOT NULL,
    email         varchar(255),
    roles         jsonb NOT NULL DEFAULT '[]',
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE (auth_provider, auth_subject)
);
`)
}
