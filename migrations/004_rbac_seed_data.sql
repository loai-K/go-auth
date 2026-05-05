-- Phase 2: seed basic RBAC data
INSERT INTO roles (id, tenant_id, name, description) VALUES (uuid_generate_v4(), 'tenant1', 'admin', 'System administrator');
INSERT INTO roles (id, tenant_id, name, description) VALUES (uuid_generate_v4(), 'tenant1', 'user', 'End user');
COMMIT;
