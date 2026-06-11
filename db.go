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

			bch_address VARCHAR(255) NOT NULL,
			token_address VARCHAR(255) NOT NULL,

			auth_nonce VARCHAR(100),
			session_token VARCHAR(255),

			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

			INDEX idx_session_token (session_token)
		);`,

		`CREATE TABLE IF NOT EXISTS listings (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,

			user_id BIGINT NOT NULL,

			title VARCHAR(255) NOT NULL,
			description TEXT NOT NULL,

			price DECIMAL(18,8) NOT NULL,
			currency VARCHAR(20) NOT NULL DEFAULT 'PUSD',

			category VARCHAR(50),
			image_url TEXT,

			sale_type ENUM('fixed','auction') DEFAULT 'fixed',
			auction_ends_at TIMESTAMP NULL,
			starting_price DECIMAL(18,8) NULL,

			status ENUM(
				'active',
				'sold',
				'hidden',
				'deleted'
			) DEFAULT 'active',

			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				ON UPDATE CURRENT_TIMESTAMP,

			INDEX idx_user_id (user_id),
			INDEX idx_currency (currency),
			INDEX idx_category (category),
			INDEX idx_status (status),
			INDEX idx_sale_type (sale_type),

			FOREIGN KEY (user_id)
			REFERENCES users(id)
			ON DELETE CASCADE
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

			txid VARCHAR(100),
			claim_txid VARCHAR(100),
			refund_txid VARCHAR(100),

			expires_at TIMESTAMP NULL,

			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
				ON UPDATE CURRENT_TIMESTAMP,

			INDEX idx_listing_id (listing_id),
			INDEX idx_buyer_user_id (buyer_user_id),
			INDEX idx_status (status),
			INDEX idx_txid (txid),
			INDEX idx_contract_address (contract_address),
			INDEX idx_payment_address (payment_address),

			FOREIGN KEY (listing_id)
			REFERENCES listings(id)
			ON DELETE CASCADE,

			FOREIGN KEY (buyer_user_id)
			REFERENCES users(id)
			ON DELETE SET NULL
		);`,

		`CREATE TABLE IF NOT EXISTS messages (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,

			listing_id BIGINT NOT NULL,

			sender VARCHAR(100) NOT NULL,
			message TEXT NOT NULL,

			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

			INDEX idx_listing_id (listing_id),

			FOREIGN KEY (listing_id)
			REFERENCES listings(id)
			ON DELETE CASCADE
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

			FOREIGN KEY (seller_id)
			REFERENCES users(id)
			ON DELETE CASCADE,

			FOREIGN KEY (reviewer_id)
			REFERENCES users(id)
			ON DELETE CASCADE
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

			FOREIGN KEY (listing_id)
			REFERENCES listings(id)
			ON DELETE CASCADE
		);`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("Database migrations complete")
}