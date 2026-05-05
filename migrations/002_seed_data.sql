-- Phase 1 seed data to bootstrap a tenant and an initial user
BEGIN;
INSERT INTO tenants (id, name, slug, default_language, supported_languages, settings, branding, created_at, updated_at)
  VALUES (uuid_generate_v4(), 'Acme Corp', 'acme', 'en', ARRAY['en'], '{}'::jsonb, '{}'::jsonb, NOW(), NOW());

-- Note: In a real environment, you would capture the generated UUID and insert a user with tenant_id referencing it.
COMMIT;
