CREATE TABLE IF NOT EXISTS `providers` (
    `id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
 	`name` text NOT NULL UNIQUE,
 	`base_url` text NOT NULL,
 	`model` text NOT NULL,
    `selected_key_id` integer,
 	`created_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL,
 	`updated_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (`selected_key_id`) REFERENCES `provider_keys`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
CREATE TABLE IF NOT EXISTS `users` (
 	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
 	`name` text NOT NULL,
 	`created_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL,
 	`updated_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL
);
 --> statement-breakpoint
CREATE TABLE IF NOT EXISTS `user_settings` (
 	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
 	`user_id` integer,
 	`selected_provider_id` integer,
    `force_selected_provider` integer DEFAULT 0,
    `req_limit` integer DEFAULT 0,
 	`created_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL,
 	`updated_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL,
 	FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON UPDATE no action ON DELETE no action,
 	FOREIGN KEY (`selected_provider_id`) REFERENCES `providers`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
CREATE TABLE IF NOT EXISTS `keys` (
   	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
   	`user_id` integer,
    `provider_id` integer,
   	`name` text,
    `secret` text NOT NULL,
   	`created_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL,
   	`updated_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL,
   	FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON UPDATE no action ON DELETE CASCADE,
    FOREIGN KEY (`provider_id`) REFERENCES `providers`(`id`) ON UPDATE no action ON DELETE CASCADE
)
