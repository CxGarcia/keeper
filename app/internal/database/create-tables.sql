
CREATE TABLE IF NOT EXISTS `profiles` (
    `id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
    `name` text NOT NULL UNIQUE,
    `is_active` integer DEFAULT 0,
    `is_default` integer DEFAULT 0,
    `created_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL,
    `updated_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS `providers` (
    `id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
    `name` text NOT NULL UNIQUE,
    `base_url` text NOT NULL,
    `model` text NOT NULL,
    `created_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL,
    `updated_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS `profile_settings` (
    `id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
    `profile_id` integer NOT NULL,
    `provider_id` integer,
    `provider_key_id` integer,
    `force_selected_provider` integer DEFAULT 0,
    `created_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL,
    `updated_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (`profile_id`) REFERENCES `profiles`(`id`) ON UPDATE no action ON DELETE CASCADE,
    FOREIGN KEY (`provider_id`) REFERENCES `providers`(`id`) ON UPDATE no action ON DELETE SET NULL,
    FOREIGN KEY (`provider_key_id`) REFERENCES `provider_keys`(`id`) ON UPDATE no action ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS `provider_keys` (
    `id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
    `provider_id` integer NOT NULL,
    `name` text,
    `secret` text NOT NULL,
    `is_active` integer DEFAULT 1,
    `req_limit` integer DEFAULT 0,
    `usage_count` integer DEFAULT 0,
    `last_used_at` text,
    `created_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL,
    `updated_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (`provider_id`) REFERENCES `providers`(`id`) ON UPDATE no action ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS `runtime_info` (
    `id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
    `data` text NOT NULL,
    `created_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL,
    `updated_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX idx_profile_settings_profile_id ON profile_settings(profile_id);
CREATE INDEX idx_profile_settings_provider_id ON profile_settings(provider_id);
CREATE INDEX idx_profile_settings_provider_key_id ON profile_settings(provider_key_id);
CREATE INDEX idx_provider_keys_provider_id ON provider_keys(provider_id);
CREATE INDEX idx_runtime_info_created_at ON runtime_info(created_at);
