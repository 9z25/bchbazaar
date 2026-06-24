package main

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func connectDB() {
	var err error

	dsn := "bchbazaar:password@tcp(127.0.0.1:3306)/bchbazaar?parseTime=true"

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MariaDB")
}

func migrateDB() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(100) NOT NULL UNIQUE,
			email VARCHAR(255) NULL,

			bch_address VARCHAR(255) NOT NULL,
			token_address VARCHAR(255) NOT NULL,

			auth_nonce VARCHAR(100),
			session_token VARCHAR(255),

			role ENUM('user','moderator','admin') DEFAULT 'user',
			moderator_active BOOLEAN DEFAULT FALSE,
			moderator_bio TEXT NULL,
			moderator_fee_percent DECIMAL(5,2) DEFAULT 0,

			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

			INDEX idx_email (email),
			INDEX idx_session_token (session_token),
			INDEX idx_role (role),
			INDEX idx_moderator_active (moderator_active)
		);`,

		`CREATE TABLE IF NOT EXISTS listings (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,

			user_id BIGINT NOT NULL,

			title VARCHAR(255) NOT NULL,
			description TEXT NOT NULL,

			price DECIMAL(18,8) NOT NULL,
			currency VARCHAR(20) NOT NULL DEFAULT 'PUSD',

			category VARCHAR(50),
			location VARCHAR(100),
			image_url TEXT,

			moderator_user_id BIGINT,
			moderator_pkh VARCHAR(100),

			sale_type ENUM('fixed','auction') DEFAULT 'fixed',
			auction_ends_at TIMESTAMP NULL,
			starting_price DECIMAL(18,8) NULL,

			status ENUM('active','sold','hidden','deleted') DEFAULT 'active',

			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

			INDEX idx_user_id (user_id),
			INDEX idx_currency (currency),
			INDEX idx_category (category),
			INDEX idx_location (location),
			INDEX idx_moderator_user_id (moderator_user_id),
			INDEX idx_status (status),
			INDEX idx_sale_type (sale_type),

			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (moderator_user_id) REFERENCES users(id) ON DELETE SET NULL
		);`,

		`CREATE TABLE IF NOT EXISTS orders (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,

			listing_id BIGINT NOT NULL,
			buyer_user_id BIGINT,

			buyer_address VARCHAR(255),
			seller_address VARCHAR(255) NOT NULL,

			payment_address VARCHAR(255),
			contract_address VARCHAR(255),

			seller_pkh VARCHAR(100),
			buyer_pkh VARCHAR(100),

			moderator_user_id BIGINT,
			moderator_pkh VARCHAR(100),

			refund_locktime BIGINT,

			amount DECIMAL(18,8) NOT NULL,
			currency VARCHAR(20) NOT NULL,

			status ENUM(
				'pending',
				'paid',
				'shipped',
				'completed',
				'cancelled',
				'expired',
				'claimed',
				'refunded'
			) DEFAULT 'pending',

			dispute_status ENUM('none','open','resolved') DEFAULT 'none',
			dispute_reason TEXT,

			moderator_decision ENUM(
				'none',
				'release_to_seller',
				'refund_to_buyer'
			) DEFAULT 'none',

			txid VARCHAR(100),
			claim_txid VARCHAR(100),
			refund_txid VARCHAR(100),
			moderator_txid VARCHAR(100),

			expires_at TIMESTAMP NULL,

			tracking_number VARCHAR(100),
			shipping_carrier VARCHAR(50),
			shipped_at TIMESTAMP NULL,

			shipping_name VARCHAR(150) NULL,
			shipping_address_1 VARCHAR(255) NULL,
			shipping_address_2 VARCHAR(255) NULL,
			shipping_city VARCHAR(100) NULL,
			shipping_state VARCHAR(100) NULL,
			shipping_postal_code VARCHAR(50) NULL,
			shipping_country VARCHAR(100) NULL,
			shipping_instructions TEXT NULL,

			escrow_amount BIGINT UNSIGNED DEFAULT 0,
			fee_amount BIGINT UNSIGNED DEFAULT 0,
			fee_pkh VARCHAR(100),
			fee_address VARCHAR(255),
			fee_token_address VARCHAR(255),
			token_category VARCHAR(100),
			token_dust BIGINT UNSIGNED DEFAULT 1000,
			contract_version VARCHAR(50),

			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

			INDEX idx_listing_id (listing_id),
			INDEX idx_buyer_user_id (buyer_user_id),
			INDEX idx_moderator_user_id (moderator_user_id),
			INDEX idx_status (status),
			INDEX idx_dispute_status (dispute_status),
			INDEX idx_txid (txid),
			INDEX idx_contract_address (contract_address),
			INDEX idx_payment_address (payment_address),

			FOREIGN KEY (listing_id) REFERENCES listings(id) ON DELETE CASCADE,
			FOREIGN KEY (buyer_user_id) REFERENCES users(id) ON DELETE SET NULL,
			FOREIGN KEY (moderator_user_id) REFERENCES users(id) ON DELETE SET NULL
		);`,

		`CREATE TABLE IF NOT EXISTS messages (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,

			listing_id BIGINT NOT NULL,
			conversation_id BIGINT NULL,
			sender_user_id BIGINT NULL,

			sender VARCHAR(100) NOT NULL,
			message TEXT NOT NULL,

			read_at TIMESTAMP NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

			INDEX idx_listing_id (listing_id),
			INDEX idx_conversation_id (conversation_id),
			INDEX idx_sender_user_id (sender_user_id),
			INDEX idx_read_at (read_at),

			FOREIGN KEY (listing_id) REFERENCES listings(id) ON DELETE CASCADE,
			FOREIGN KEY (sender_user_id) REFERENCES users(id) ON DELETE SET NULL
		);`,

		`CREATE TABLE IF NOT EXISTS reviews (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,

			seller_id BIGINT NOT NULL,
			reviewer_id BIGINT NOT NULL,

			rating TINYINT NOT NULL,
			comment TEXT,

			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

			INDEX idx_seller_id (seller_id),
			INDEX idx_reviewer_id (reviewer_id),

			UNIQUE KEY uniq_review (seller_id, reviewer_id),

			CHECK (rating >= 1 AND rating <= 5),

			FOREIGN KEY (seller_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (reviewer_id) REFERENCES users(id) ON DELETE CASCADE
		);`,

		`CREATE TABLE IF NOT EXISTS bids (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,

			listing_id BIGINT NOT NULL,

			bidder_address VARCHAR(255) NOT NULL,
			amount DECIMAL(18,8) NOT NULL,
			currency VARCHAR(20) NOT NULL,

			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

			INDEX idx_listing_id (listing_id),
			INDEX idx_amount (amount),

			FOREIGN KEY (listing_id) REFERENCES listings(id) ON DELETE CASCADE
		);`,

		`CREATE TABLE IF NOT EXISTS conversations (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,

			listing_id BIGINT NOT NULL,
			buyer_user_id BIGINT NOT NULL,
			seller_user_id BIGINT NOT NULL,

			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

			UNIQUE KEY unique_conversation (listing_id, buyer_user_id, seller_user_id),
			INDEX idx_buyer_user_id (buyer_user_id),
			INDEX idx_seller_user_id (seller_user_id),
			INDEX idx_listing_id (listing_id),
			INDEX idx_updated_at (updated_at),

			FOREIGN KEY (listing_id) REFERENCES listings(id) ON DELETE CASCADE,
			FOREIGN KEY (buyer_user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (seller_user_id) REFERENCES users(id) ON DELETE CASCADE
		);`,

		`ALTER TABLE users ADD COLUMN IF NOT EXISTS email VARCHAR(255) NULL AFTER username;`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS role ENUM('user','moderator','admin') DEFAULT 'user' AFTER session_token;`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS moderator_active BOOLEAN DEFAULT FALSE AFTER role;`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS moderator_bio TEXT NULL AFTER moderator_active;`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS moderator_fee_percent DECIMAL(5,2) DEFAULT 0 AFTER moderator_bio;`,
		`ALTER TABLE users ADD INDEX IF NOT EXISTS idx_email (email);`,
		`ALTER TABLE users ADD INDEX IF NOT EXISTS idx_role (role);`,
		`ALTER TABLE users ADD INDEX IF NOT EXISTS idx_moderator_active (moderator_active);`,

		`ALTER TABLE listings ADD COLUMN IF NOT EXISTS location VARCHAR(100) AFTER category;`,
		`ALTER TABLE listings ADD COLUMN IF NOT EXISTS moderator_user_id BIGINT AFTER image_url;`,
		`ALTER TABLE listings ADD COLUMN IF NOT EXISTS moderator_pkh VARCHAR(100) AFTER moderator_user_id;`,
		`ALTER TABLE listings ADD COLUMN IF NOT EXISTS sale_type ENUM('fixed','auction') DEFAULT 'fixed' AFTER moderator_pkh;`,
		`ALTER TABLE listings ADD COLUMN IF NOT EXISTS auction_ends_at TIMESTAMP NULL AFTER sale_type;`,
		`ALTER TABLE listings ADD COLUMN IF NOT EXISTS starting_price DECIMAL(18,8) NULL AFTER auction_ends_at;`,
		`ALTER TABLE listings ADD COLUMN IF NOT EXISTS status ENUM('active','sold','hidden','deleted') DEFAULT 'active' AFTER starting_price;`,

		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS tracking_number VARCHAR(100) NULL;`,
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS shipping_carrier VARCHAR(50) NULL;`,
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS shipped_at TIMESTAMP NULL;`,

		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS shipping_name VARCHAR(150) NULL;`,
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS shipping_address_1 VARCHAR(255) NULL;`,
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS shipping_address_2 VARCHAR(255) NULL;`,
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS shipping_city VARCHAR(100) NULL;`,
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS shipping_state VARCHAR(100) NULL;`,
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS shipping_postal_code VARCHAR(50) NULL;`,
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS shipping_country VARCHAR(100) NULL;`,
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS shipping_instructions TEXT NULL;`,

		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS escrow_amount BIGINT UNSIGNED DEFAULT 0;`,
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS fee_amount BIGINT UNSIGNED DEFAULT 0;`,
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS fee_pkh VARCHAR(100) NULL;`,
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS fee_address VARCHAR(255) NULL;`,
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS fee_token_address VARCHAR(255) NULL;`,
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS token_category VARCHAR(100) NULL;`,
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS token_dust BIGINT UNSIGNED DEFAULT 1000;`,
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS contract_version VARCHAR(50) NULL;`,

		`ALTER TABLE orders MODIFY COLUMN dispute_status ENUM('none','opened','open','resolved') DEFAULT 'none';`,
		`UPDATE orders SET dispute_status = 'open' WHERE dispute_status = 'opened';`,
		`ALTER TABLE orders MODIFY COLUMN dispute_status ENUM('none','open','resolved') DEFAULT 'none';`,

		`ALTER TABLE orders MODIFY COLUMN moderator_decision ENUM('none','release','refund','release_to_seller','refund_to_buyer') DEFAULT 'none';`,
		`UPDATE orders SET moderator_decision = 'release_to_seller' WHERE moderator_decision = 'release';`,
		`UPDATE orders SET moderator_decision = 'refund_to_buyer' WHERE moderator_decision = 'refund';`,
		`ALTER TABLE orders MODIFY COLUMN moderator_decision ENUM('none','release_to_seller','refund_to_buyer') DEFAULT 'none';`,

		`ALTER TABLE messages ADD COLUMN IF NOT EXISTS conversation_id BIGINT NULL;`,
		`ALTER TABLE messages ADD COLUMN IF NOT EXISTS sender_user_id BIGINT NULL;`,
		`ALTER TABLE messages ADD COLUMN IF NOT EXISTS read_at TIMESTAMP NULL;`,
		`ALTER TABLE messages ADD INDEX IF NOT EXISTS idx_conversation_id (conversation_id);`,
		`ALTER TABLE messages ADD INDEX IF NOT EXISTS idx_sender_user_id (sender_user_id);`,
		`ALTER TABLE messages ADD INDEX IF NOT EXISTS idx_read_at (read_at);`,

		`ALTER TABLE conversations ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;`,
		`ALTER TABLE conversations ADD INDEX IF NOT EXISTS idx_updated_at (updated_at);`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("Database migrations complete")
}
