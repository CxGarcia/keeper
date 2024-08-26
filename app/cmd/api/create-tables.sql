CREATE TABLE IF NOT EXISTS `providers` (
    `id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
 	`name` text NOT NULL,
 	`base_url` text NOT NULL,
 	`api_key` text,
 	`model` text NOT NULL,
 	`created_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL,
 	`updated_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL
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
 	`created_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL,
 	`updated_at` text DEFAULT CURRENT_TIMESTAMP NOT NULL,
 	FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON UPDATE no action ON DELETE no action,
 	FOREIGN KEY (`selected_provider_id`) REFERENCES `providers`(`id`) ON UPDATE no action ON DELETE no action
);
