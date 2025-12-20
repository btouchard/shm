-- SPDX-License-Identifier: AGPL-3.0-or-later
-- Migration: Add applications table and link instances to applications

-- Create slugify function for app_name normalization
CREATE OR REPLACE FUNCTION slugify(text) RETURNS text AS $$
DECLARE
    result text;
BEGIN
    -- Convert to lowercase
    result := lower($1);

    -- Replace accented characters
    result := translate(result,
        'àáâãäåèéêëìíîïòóôõöùúûüýÿñç',
        'aaaaaaeeeeiiiiooooouuuuyync'
    );

    -- Replace non-alphanumeric chars with hyphens
    result := regexp_replace(result, '[^a-z0-9]+', '-', 'g');

    -- Remove leading/trailing hyphens
    result := trim(both '-' from result);

    -- Collapse multiple hyphens
    result := regexp_replace(result, '-+', '-', 'g');

    -- Fallback if empty
    IF result = '' THEN
        result := 'app';
    END IF;

    RETURN result;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Create applications table
CREATE TABLE applications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    app_slug VARCHAR(100) NOT NULL UNIQUE,
    app_name VARCHAR(100) NOT NULL,

    -- GitHub metrics (auto-fetched)
    github_url VARCHAR(500),
    github_stars INT DEFAULT 0,
    github_stars_updated_at TIMESTAMP WITH TIME ZONE,

    -- Custom logo (admin-set, nullable)
    logo_url VARCHAR(500),

    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for fast slug lookups
CREATE INDEX idx_applications_slug ON applications(app_slug);

-- Populate applications from existing instances
INSERT INTO applications (app_slug, app_name)
SELECT DISTINCT slugify(app_name), app_name
FROM instances
WHERE app_name IS NOT NULL AND app_name != ''
ON CONFLICT (app_slug) DO NOTHING;

-- Add application_id column to instances
ALTER TABLE instances
    ADD COLUMN application_id UUID;

-- Link existing instances to applications
UPDATE instances i
SET application_id = a.id
FROM applications a
WHERE slugify(i.app_name) = a.app_slug;

-- Add foreign key constraint
ALTER TABLE instances
    ADD CONSTRAINT fk_instances_application
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE RESTRICT;

-- Index for joins
CREATE INDEX idx_instances_application_id ON instances(application_id);

-- Add trigger to update updated_at on applications
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_applications_updated_at
    BEFORE UPDATE ON applications
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
