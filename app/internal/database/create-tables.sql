CREATE TABLE IF NOT EXISTS `providers` (
    `id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
    `name` text NOT NULL UNIQUE,
    `base_url` text NOT NULL,
    `model` text NOT NULL,
    `created_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL,
    `updated_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS `users` (
    `id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
    `name` text NOT NULL,
    `created_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL,
    `updated_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS `user_settings` (
    `id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
    `user_id` integer,
    `selected_provider_id` integer,
    `force_selected_provider` integer DEFAULT 0,
    `created_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL,
    `updated_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON UPDATE no action ON DELETE no action,
    FOREIGN KEY (`selected_provider_id`) REFERENCES `providers`(`id`) ON UPDATE no action ON DELETE no action
);

CREATE TABLE IF NOT EXISTS `provider_keys` (
    `id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
    `user_id` integer,
    `provider_id` integer,
    `name` text,
    `secret` text NOT NULL,
    `is_active` integer DEFAULT 1,
    `req_limit` integer DEFAULT 0,
    `usage_count` integer DEFAULT 0,
    `last_used_at` text,
    `created_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL,
    `updated_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON UPDATE no action ON DELETE CASCADE,
    FOREIGN KEY (`provider_id`) REFERENCES `providers`(`id`) ON UPDATE no action ON DELETE CASCADE
);

CREATE INDEX idx_user_settings_user_id ON user_settings(user_id);
CREATE INDEX idx_user_settings_selected_provider_id ON user_settings(selected_provider_id);
CREATE INDEX idx_provider_keys_user_id ON provider_keys(user_id);
CREATE INDEX idx_provider_keys_provider_id ON provider_keys(provider_id);
