package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"time"
	"math"
)

type CreateUserRequest struct {
	Username     string `json:"username"`
	Email        string `json:"email"`
	BCHAddress   string `json:"bch_address"`
	TokenAddress string `json:"token_address"`
}

type CreateListingRequest struct {
	UserID          uint64  `json:"user_id"`
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	Price           float64 `json:"price"`
	Currency        string  `json:"currency"`
	ImageURL        string  `json:"image_url"`
	Category        string  `json:"category"`
	Location        string  `json:"location"`
	ModeratorUserID uint64  `json:"moderator_user_id"`
	ModeratorPKH    string  `json:"moderator_pkh"`
}

const MarketplaceFeeBps uint64 = 250 // 2.5%

const MarketplaceFeePKH = "0c67ad9176ee207e2c0020a11baf563db3595fbd"

const MarketplaceFeeAddress = "bitcoincash:qqxx0tv3wmhzql3vqqs2zxa02c7mxk2lh5q0n2yach"

// Replace this with the actual tokenaddr version of the same address
const MarketplaceFeeTokenAddress = "bitcoincash:zqxx0tv3wmhzql3vqqs2zxa02c7mxk2lh589q52m8y" 

const PUSDCategory = "2469acc5afa4b10cb5b5c04afb89c3a3ffd61c5da9c01e26d00951cae2a02544"

const TokenDust uint64 = 1000

func normalizeCurrency(currency string) string {
	return strings.ToUpper(strings.TrimSpace(currency))
}

func amountBaseUnits(currency string, amount float64) uint64 {
	switch currency {
	case "BCH":
		return uint64(math.Round(amount * 100000000))

	case "PUSD":
		return uint64(math.Round(amount * 100))

	default:
		return 0
	}
}

func feeAmountBaseUnits(currency string, amount float64) uint64 {
	baseAmount := amountBaseUnits(currency, amount)
	return (baseAmount * MarketplaceFeeBps) / 10000
}

func tokenCategoryForContract(currency string) string {
	if currency == "PUSD" {
		return reverseHexBytes(PUSDCategory)
	}

	return "0x"
}

func reverseHexBytes(hex string) string {
	if len(hex)%2 != 0 {
		return hex
	}

	out := ""
	for i := len(hex); i > 0; i -= 2 {
		out += hex[i-2 : i]
	}
	return out
}

func createUser(c *gin.Context) {
	var req CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Username == "" || req.BCHAddress == "" || req.TokenAddress == "" {
		c.JSON(400, gin.H{"error": "username, bch_address, and token_address are required"})
		return
	}

	if req.Email != "" && !strings.Contains(req.Email, "@") {
		c.JSON(400, gin.H{"error": "invalid email"})
		return
	}

	_, err := db.Exec(`
		INSERT INTO users
		(username, email, bch_address, token_address)
		VALUES (?, ?, ?, ?)
	`, req.Username, req.Email, req.BCHAddress, req.TokenAddress)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "user created"})
}

func listModerators(c *gin.Context) {
	rows, err := db.Query(`
		SELECT id, username, bch_address, token_address
		FROM users
		WHERE role IN ('moderator', 'admin')
		ORDER BY username ASC
	`)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	moderators := []gin.H{}
	for rows.Next() {
		var id uint64
		var username, bchAddress, tokenAddress string
		if err := rows.Scan(&id, &username, &bchAddress, &tokenAddress); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		moderators = append(moderators, gin.H{
			"id":            id,
			"username":      username,
			"bch_address":   bchAddress,
			"token_address": tokenAddress,
		})
	}

	c.JSON(200, moderators)
}

func listModeratorDisputes(c *gin.Context) {
	userID := c.MustGet("user_id").(uint64)
	role := c.MustGet("role").(string)
	if role != "moderator" && role != "admin" {
		c.JSON(403, gin.H{"error": "moderator role required"})
		return
	}

	rows, err := db.Query(`
		SELECT
			o.id,
			o.listing_id,
			l.title,
			COALESCE(b.username, ''),
			s.username,
			o.amount,
			o.currency,
			o.status,
			o.dispute_status,
			COALESCE(o.dispute_reason, ''),
			o.moderator_decision,
			COALESCE(o.moderator_txid, ''),
			o.created_at
		FROM orders o
		JOIN listings l ON l.id = o.listing_id
		JOIN users s ON s.id = l.user_id
		LEFT JOIN users b ON b.id = o.buyer_user_id
		WHERE o.moderator_user_id = ?
		AND o.dispute_status = 'open'
		ORDER BY o.id DESC
	`, userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	disputes := []gin.H{}
	for rows.Next() {
		var orderID, listingID uint64
		var listingTitle, buyerUsername, sellerUsername string
		var amount float64
		var currency, status, disputeStatus, disputeReason, moderatorDecision, moderatorTxid, createdAt string
		if err := rows.Scan(&orderID, &listingID, &listingTitle, &buyerUsername, &sellerUsername, &amount, &currency, &status, &disputeStatus, &disputeReason, &moderatorDecision, &moderatorTxid, &createdAt); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		disputes = append(disputes, gin.H{
			"id":                 orderID,
			"listing_id":         listingID,
			"listing_title":      listingTitle,
			"buyer_username":     buyerUsername,
			"seller_username":    sellerUsername,
			"amount":             amount,
			"currency":           currency,
			"status":             status,
			"dispute_status":     disputeStatus,
			"dispute_reason":     disputeReason,
			"moderator_decision": moderatorDecision,
			"moderator_txid":     moderatorTxid,
			"created_at":         createdAt,
		})
	}

	c.JSON(200, disputes)
}

func createListing(c *gin.Context) {
	userID := c.MustGet("user_id").(uint64)

	var req CreateListingRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var moderatorUserID any
	if req.ModeratorUserID != 0 {
		var role string
		err := db.QueryRow(`
			SELECT role
			FROM users
			WHERE id = ?
		`, req.ModeratorUserID).Scan(&role)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "moderator not found"})
			return
		}
		if role != "moderator" && role != "admin" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "selected user is not a moderator"})
			return
		}
		if req.ModeratorPKH == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "moderator_pkh is required"})
			return
		}
		moderatorUserID = req.ModeratorUserID
	}

	_, err := db.Exec(`
		INSERT INTO listings
		(user_id, title, description, price, currency, category, location, image_url, moderator_user_id, moderator_pkh)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, userID, req.Title, req.Description, req.Price, req.Currency, req.Category, req.Location, req.ImageURL, moderatorUserID, req.ModeratorPKH)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "listing created"})
}

func listListings(c *gin.Context) {
	currency := c.Query("currency")
	search := c.Query("search")
	sellerID := c.Query("seller")
	category := c.Query("category")

	query := `
		SELECT
			l.id,
			l.user_id,
			u.username,
			l.title,
			l.description,
			l.price,
			l.currency,
			COALESCE(l.category, ''),
			COALESCE(l.location, ''),
			COALESCE(l.image_url, ''),
			l.moderator_user_id,
			COALESCE(l.moderator_pkh, ''),
			l.created_at,
			l.status
		FROM listings l
		JOIN users u ON u.id = l.user_id
		WHERE 1=1
	`

	args := []any{}

	if currency != "" {
		query += " AND l.currency = ?"
		args = append(args, currency)
	}

	if sellerID != "" {
		query += " AND l.user_id = ?"
		args = append(args, sellerID)
	}

	if search != "" {
		query += " AND (l.title LIKE ? OR l.description LIKE ?)"
		like := "%" + search + "%"
		args = append(args, like, like)
	}
	if category != "" {
		query += " AND l.category = ?"
		args = append(args, category)
	}

	query += " ORDER BY l.id DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var listings []gin.H

	for rows.Next() {
		var (
			id            uint64
			userID        uint64
			username      string
			title         string
			description   string
			price         float64
			currency      string
			categoryValue string
			location      string
			imageURL      string
			moderatorID   sql.NullInt64
			moderatorPKH  string
			createdAt     string
			status        string
		)

		if err := rows.Scan(
			&id,
			&userID,
			&username,
			&title,
			&description,
			&price,
			&currency,
			&categoryValue,
			&location,
			&imageURL,
			&moderatorID,
			&moderatorPKH,
			&createdAt,
			&status,
		); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		listings = append(listings, gin.H{
			"id":                id,
			"user_id":           userID,
			"seller":            username,
			"title":             title,
			"description":       description,
			"price":             price,
			"currency":          currency,
			"category":          categoryValue,
			"location":          location,
			"image_url":         imageURL,
			"moderator_user_id": moderatorID.Int64,
			"moderator_pkh":     moderatorPKH,
			"created_at":        createdAt,
			"status":            status,
		})
	}

	c.JSON(200, listings)
}

func getListing(c *gin.Context) {
	id := c.Param("id")

	var listing gin.H = gin.H{}

	row := db.QueryRow(`
		SELECT 
			l.id,
			l.user_id,
			u.username,
			u.bch_address,
			u.token_address,
			l.title,
			l.description,
			l.price,
			l.currency,
			COALESCE(l.category, ''),
			COALESCE(l.location, ''),
			COALESCE(l.image_url, ''),
			l.moderator_user_id,
			COALESCE(l.moderator_pkh, ''),
			COALESCE(m.username, ''),
			l.created_at,
			l.status
		FROM listings l
		JOIN users u ON u.id = l.user_id
		LEFT JOIN users m ON m.id = l.moderator_user_id
		WHERE l.id = ?
	`, id)

	var (
		listingID    uint64
		userID       uint64
		username     string
		bchAddress   string
		tokenAddress string
		title        string
		description  string
		price        float64
		currency     string
		category     string
		location     string
		imageURL     string
		moderatorID  sql.NullInt64
		moderatorPKH string
		moderator    string
		createdAt    string
		status       string
	)

	err := row.Scan(
		&listingID,
		&userID,
		&username,
		&bchAddress,
		&tokenAddress,
		&title,
		&description,
		&price,
		&currency,
		&category,
		&location,
		&imageURL,
		&moderatorID,
		&moderatorPKH,
		&moderator,
		&createdAt,
		&status,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "listing not found"})
		return
	}

	listing["id"] = listingID
	listing["user_id"] = userID
	listing["seller"] = username
	listing["seller_bch_address"] = bchAddress
	listing["token_address"] = tokenAddress
	listing["seller_token_address"] = tokenAddress
	listing["title"] = title
	listing["description"] = description
	listing["price"] = price
	listing["price_pusd"] = price
	listing["currency"] = currency
	listing["category"] = category
	listing["location"] = location
	listing["image_url"] = imageURL
	listing["moderator_user_id"] = moderatorID.Int64
	listing["moderator_pkh"] = moderatorPKH
	listing["moderator"] = moderator
	listing["created_at"] = createdAt
	listing["status"] = status

	c.JSON(http.StatusOK, listing)
}

const MUSDCategory = "b38a33f750f84c5c169a6f23cb873e6e79605021585d4f3408789689ed87f366"

type CreateOrderRequest struct {
	ListingID        uint64 `json:"listing_id"`
	BuyerAddress    string `json:"buyer_address"`
	ContractAddress string `json:"contract_address"`

	SellerPKH      string `json:"seller_pkh"`
	BuyerPKH       string `json:"buyer_pkh"`
	ModeratorPKH   string `json:"moderator_pkh"`
	RefundLocktime uint64 `json:"refund_locktime"`

	FeePKH    string `json:"fee_pkh"`
	FeeAmount uint64 `json:"fee_amount"`

	ShippingName         string `json:"shipping_name"`
	ShippingAddress1     string `json:"shipping_address_1"`
	ShippingAddress2     string `json:"shipping_address_2"`
	ShippingCity         string `json:"shipping_city"`
	ShippingState        string `json:"shipping_state"`
	ShippingPostalCode   string `json:"shipping_postal_code"`
	ShippingCountry      string `json:"shipping_country"`
	ShippingInstructions string `json:"shipping_instructions"`
}

func createOrder(c *gin.Context) {
	buyerUserID := c.MustGet("user_id").(uint64)

	var req CreateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if req.ContractAddress == "" || req.SellerPKH == "" || req.BuyerPKH == "" || req.RefundLocktime == 0 {
		c.JSON(400, gin.H{
			"error": "contract_address, seller_pkh, buyer_pkh, and refund_locktime are required",
		})
		return
	}

	var (
		amount                 float64
		currency               string
		sellerUserID           uint64
		bchAddress             string
		tokenAddress           string
		listingModeratorUserID sql.NullInt64
		listingModeratorPKH    sql.NullString
	)

	

	err := db.QueryRow(`
		SELECT
			l.price,
			l.currency,
			l.user_id,
			u.bch_address,
			u.token_address,
			l.moderator_user_id,
			l.moderator_pkh
		FROM listings l
		JOIN users u ON u.id = l.user_id
		WHERE l.id = ?
		AND l.status = 'active'
	`, req.ListingID).Scan(
		&amount,
		&currency,
		&sellerUserID,
		&bchAddress,
		&tokenAddress,
		&listingModeratorUserID,
		&listingModeratorPKH,
	)

	if err != nil {
		c.JSON(404, gin.H{"error": "active listing not found"})
		return
	}

	if buyerUserID == sellerUserID {
		c.JSON(400, gin.H{"error": "cannot buy your own listing"})
		return
	}

	currency = normalizeCurrency(currency)

	escrowAmount := amountBaseUnits(currency, amount)
	feeAmount := feeAmountBaseUnits(currency, amount)
	tokenCategory := tokenCategoryForContract(currency)
	feePKH := MarketplaceFeePKH
	feeAddress := MarketplaceFeeAddress
    feeTokenAddress := MarketplaceFeeTokenAddress

	if req.FeePKH != "" && req.FeePKH != feePKH {
		c.JSON(400, gin.H{"error": "invalid fee pkh"})
		return
	}

	// Frontend fee_amount is only informational.
	// Backend is the source of truth.
	if req.FeeAmount != 0 && req.FeeAmount != feeAmount {
		c.JSON(400, gin.H{
			"error": "invalid fee amount",
			"frontend_fee_amount": req.FeeAmount,
			"backend_fee_amount": feeAmount,
		})
		return
	}

	sellerAddress := bchAddress
	if currency != "BCH" {
		sellerAddress = tokenAddress
	}

	paymentAddress := req.ContractAddress
	expiresAt := time.Now().Add(24 * time.Hour)

	var moderatorUserID any
	if listingModeratorUserID.Valid {
		moderatorUserID = listingModeratorUserID.Int64
	}

	result, err := db.Exec(`
		INSERT INTO orders
		(
			listing_id,
			buyer_user_id,
			buyer_address,
			seller_address,
			payment_address,
			contract_address,
			seller_pkh,
			buyer_pkh,
			moderator_user_id,
			moderator_pkh,
			refund_locktime,
			amount,
			currency,
			expires_at,
			shipping_name,
			shipping_address_1,
			shipping_address_2,
			shipping_city,
			shipping_state,
			shipping_postal_code,
			shipping_country,
			shipping_instructions,
			escrow_amount,
			fee_amount,
			fee_pkh,
			fee_address,
			fee_token_address,
			token_category,
			token_dust,
			contract_version
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		req.ListingID,
		buyerUserID,
		req.BuyerAddress,
		sellerAddress,
		paymentAddress,
		req.ContractAddress,
		req.SellerPKH,
		req.BuyerPKH,
		moderatorUserID,
		listingModeratorPKH.String,
		req.RefundLocktime,
		amount,
		currency,
		expiresAt,
		req.ShippingName,
		req.ShippingAddress1,
		req.ShippingAddress2,
		req.ShippingCity,
		req.ShippingState,
		req.ShippingPostalCode,
		req.ShippingCountry,
		req.ShippingInstructions,
		escrowAmount,
		feeAmount,
		feePKH,
		feeAddress,
		feeTokenAddress,
		tokenCategory,
		TokenDust,
		"v2-pusd-fee",
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	orderID, _ := result.LastInsertId()

	c.JSON(201, gin.H{
		"order_id":          orderID,
		"listing_id":        req.ListingID,
		"buyer_user_id":     buyerUserID,
		"amount":            amount,
		"currency":          currency,
		"seller_address":    sellerAddress,
		"payment_address":   paymentAddress,
		"contract_address":  req.ContractAddress,
		"seller_pkh":        req.SellerPKH,
		"buyer_pkh":         req.BuyerPKH,
		"moderator_user_id": listingModeratorUserID.Int64,
		"moderator_pkh":     listingModeratorPKH.String,
		"refund_locktime":   req.RefundLocktime,
		"payment_uri":       buildPaymentURI(paymentAddress, amount, currency),
		"status":            "pending",
		"expires_at":        expiresAt,
	})
}

func listOrders(c *gin.Context) {
	userID := c.MustGet("user_id").(uint64)
	buyerOnly := c.Query("buyer") == "true"
	sellerOnly := c.Query("seller") == "true"

	if buyerOnly == sellerOnly {
		c.JSON(400, gin.H{"error": "pass exactly one of buyer=true or seller=true"})
		return
	}

	query := `
		SELECT
			o.id,
			o.listing_id,
			l.title,
			COALESCE(b.username, ''),
			s.username,
			o.amount,
			o.currency,
			o.status,
			o.dispute_status,
			o.created_at,
			COALESCE(l.image_url, '') AS listing_image_url
		FROM orders o
		JOIN listings l ON l.id = o.listing_id
		JOIN users s ON s.id = l.user_id
		LEFT JOIN users b ON b.id = o.buyer_user_id
	`

	args := []any{userID}
	if buyerOnly {
		query += " WHERE o.buyer_user_id = ?"
	} else {
		query += " WHERE l.user_id = ?"
	}
	query += " ORDER BY o.id DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	orders := []gin.H{}
	for rows.Next() {
		var (
			orderID        uint64
			listingID      uint64
			listingTitle   string
			buyerUsername  string
			sellerUsername string
			amount         float64
			currency       string
			status         string
			disputeStatus  string
			createdAt      string
			listingImageURL string
		)

		if err := rows.Scan(
			&orderID,
			&listingID,
			&listingTitle,
			&buyerUsername,
			&sellerUsername,
			&amount,
			&currency,
			&status,
			&disputeStatus,
			&createdAt,
			&listingImageURL,
		); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		orders = append(orders, gin.H{
			"id":              orderID,
			"listing_id":      listingID,
			"listing_title":   listingTitle,
			"listing_image_url": listingImageURL,
			"buyer_username":  buyerUsername,
			"seller_username": sellerUsername,
			"amount":          amount,
			"currency":        currency,
			"status":          status,
			"dispute_status":  disputeStatus,
			"created_at":      createdAt,
		})
	}

	c.JSON(200, orders)
}

func getOrder(c *gin.Context) {
	id := c.Param("id")

	var (
		orderID              uint64
		listingID            uint64
		buyerUserID          sql.NullInt64
		sellerUserID         uint64
		buyerUsername        sql.NullString
		sellerUsername       string

		listingTitle         string
		listingDescription   string
		listingImageURL      string
		listingCategory      string
		listingLocation      string

		buyerAddress         string
		sellerAddress        string
		paymentAddress       string
		sellerPkh            string
		buyerPkh             string
		moderatorUserID      sql.NullInt64
		moderatorPkh         sql.NullString
		refundLocktime       uint64
		amount               float64
		currency             string
		status               string
		disputeStatus        string
		txid                 sql.NullString

		trackingNumber       sql.NullString
		shippingCarrier      sql.NullString
		shippedAt            sql.NullString
		createdAt            string
		updatedAt            string

		claimTxid            sql.NullString
		refundTxid           sql.NullString
		moderatorTxid        sql.NullString

		shippingName         sql.NullString
		shippingAddress1     sql.NullString
		shippingAddress2     sql.NullString
		shippingCity         sql.NullString
		shippingState        sql.NullString
		shippingPostalCode   sql.NullString
		shippingCountry      sql.NullString
		shippingInstructions sql.NullString

		feePKH               string
		feeAmount            uint64
		feeAddress           string
		feeTokenAddress      string
		tokenCategory        string
		tokenDust            uint64
		escrowAmount         uint64
		contractVersion      string
	)

	err := db.QueryRow(`
		SELECT
			o.id,
			o.listing_id,
			o.buyer_user_id,
			l.user_id AS seller_user_id,
			COALESCE(b.username, '') AS buyer_username,
			COALESCE(s.username, '') AS seller_username,

			COALESCE(l.title, '') AS listing_title,
			COALESCE(l.description, '') AS listing_description,
			COALESCE(l.image_url, '') AS listing_image_url,
			COALESCE(l.category, '') AS listing_category,
			COALESCE(l.location, '') AS listing_location,

			COALESCE(o.buyer_address, '') AS buyer_address,
			COALESCE(o.seller_address, '') AS seller_address,
			COALESCE(o.payment_address, '') AS payment_address,
			COALESCE(o.seller_pkh, '') AS seller_pkh,
			COALESCE(o.buyer_pkh, '') AS buyer_pkh,
			o.moderator_user_id,
			o.moderator_pkh,
			COALESCE(o.refund_locktime, 0) AS refund_locktime,
			o.amount,
			o.currency,
			o.status,
			COALESCE(o.dispute_status, 'none') AS dispute_status,
			o.txid,

			o.tracking_number,
			o.shipping_carrier,
			o.shipped_at,
			o.created_at,
			o.updated_at,

			o.claim_txid,
			o.refund_txid,
			o.moderator_txid,

			o.shipping_name,
			o.shipping_address_1,
			o.shipping_address_2,
			o.shipping_city,
			o.shipping_state,
			o.shipping_postal_code,
			o.shipping_country,
			o.shipping_instructions,

			COALESCE(o.fee_pkh, '') AS fee_pkh,
			COALESCE(o.fee_amount, 0) AS fee_amount,
			COALESCE(o.fee_address, '') AS fee_address,
			COALESCE(o.fee_token_address, '') AS fee_token_address,
			COALESCE(o.token_category, '') AS token_category,
			COALESCE(o.token_dust, 1000) AS token_dust,
			COALESCE(o.escrow_amount, 0) AS escrow_amount,
			COALESCE(o.contract_version, '') AS contract_version
		FROM orders o
		JOIN listings l ON l.id = o.listing_id
		JOIN users s ON s.id = l.user_id
		LEFT JOIN users b ON b.id = o.buyer_user_id
		WHERE o.id = ?
	`, id).Scan(
		&orderID,
		&listingID,
		&buyerUserID,
		&sellerUserID,
		&buyerUsername,
		&sellerUsername,

		&listingTitle,
		&listingDescription,
		&listingImageURL,
		&listingCategory,
		&listingLocation,

		&buyerAddress,
		&sellerAddress,
		&paymentAddress,
		&sellerPkh,
		&buyerPkh,
		&moderatorUserID,
		&moderatorPkh,
		&refundLocktime,
		&amount,
		&currency,
		&status,
		&disputeStatus,
		&txid,

		&trackingNumber,
		&shippingCarrier,
		&shippedAt,
		&createdAt,
		&updatedAt,

		&claimTxid,
		&refundTxid,
		&moderatorTxid,

		&shippingName,
		&shippingAddress1,
		&shippingAddress2,
		&shippingCity,
		&shippingState,
		&shippingPostalCode,
		&shippingCountry,
		&shippingInstructions,

		&feePKH,
		&feeAmount,
		&feeAddress,
		&feeTokenAddress,
		&tokenCategory,
		&tokenDust,
		&escrowAmount,
		&contractVersion,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(404, gin.H{"error": "order not found"})
			return
		}

		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uint64)
	role := c.MustGet("role").(string)

	buyerAllowed := buyerUserID.Valid && userID == uint64(buyerUserID.Int64)
	sellerAllowed := userID == sellerUserID
	moderatorAllowed := moderatorUserID.Valid && userID == uint64(moderatorUserID.Int64)
	adminAllowed := role == "admin"

	if !buyerAllowed && !sellerAllowed && !moderatorAllowed && !adminAllowed {
		c.JSON(403, gin.H{"error": "not allowed"})
		return
	}

	var buyerUserIDValue any
	if buyerUserID.Valid {
		buyerUserIDValue = buyerUserID.Int64
	}

	var moderatorUserIDValue any
	if moderatorUserID.Valid {
		moderatorUserIDValue = moderatorUserID.Int64
	}

	c.JSON(200, gin.H{
		"id":                    orderID,
		"listing_id":            listingID,
		"listing_title":         listingTitle,
		"listing_description":   listingDescription,
		"listing_image_url":     listingImageURL,
		"listing_category":      listingCategory,
		"listing_location":      listingLocation,

		"buyer_user_id":         buyerUserIDValue,
		"seller_user_id":        sellerUserID,
		"buyer_username":        buyerUsername.String,
		"seller_username":       sellerUsername,

		"buyer_address":         buyerAddress,
		"seller_address":        sellerAddress,
		"payment_address":       paymentAddress,
		"seller_pkh":            sellerPkh,
		"buyer_pkh":             buyerPkh,
		"moderator_user_id":     moderatorUserIDValue,
		"moderator_pkh":         moderatorPkh.String,
		"refund_locktime":       refundLocktime,

		"payment_uri":           buildPaymentURI(paymentAddress, amount, currency),
		"amount":                amount,
		"currency":              currency,
		"status":                status,
		"dispute_status":        disputeStatus,
		"txid":                  txid.String,

		"tracking_number":       trackingNumber.String,
		"shipping_carrier":      shippingCarrier.String,
		"shipped_at":            shippedAt.String,
		"created_at":            createdAt,
		"updated_at":            updatedAt,

		"claim_txid":            claimTxid.String,
		"refund_txid":           refundTxid.String,
		"moderator_txid":        moderatorTxid.String,

		"shipping_name":         shippingName.String,
		"shipping_address_1":    shippingAddress1.String,
		"shipping_address_2":    shippingAddress2.String,
		"shipping_city":         shippingCity.String,
		"shipping_state":        shippingState.String,
		"shipping_postal_code":  shippingPostalCode.String,
		"shipping_country":      shippingCountry.String,
		"shipping_instructions": shippingInstructions.String,

		"escrow_amount":         escrowAmount,
		"fee_amount":            feeAmount,
		"fee_pkh":               feePKH,
		"fee_address":           feeAddress,
		"fee_token_address":     feeTokenAddress,
		"token_category":        tokenCategory,
		"token_dust":            tokenDust,
		"contract_version":      contractVersion,
	})
}

func buildPUSDURI(address string, amount float64) string {
	baseUnits := int64(amount * 100)

	return address +
		"?c=" + PUSDCategory +
		"&ft=" + fmt.Sprintf("%d", baseUnits)
}

type CreateMessageRequest struct {
	ListingID uint64 `json:"listing_id"`
	Message   string `json:"message"`
}

func createMessage(c *gin.Context) {
	userID := c.MustGet("user_id").(uint64)

	var req CreateMessageRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if req.ListingID == 0 || req.Message == "" {
		c.JSON(400, gin.H{"error": "missing required fields"})
		return
	}

	var username string

	err := db.QueryRow(`
		SELECT username
		FROM users
		WHERE id = ?
	`, userID).Scan(&username)

	if err != nil {
		c.JSON(401, gin.H{"error": "user not found"})
		return
	}

	result, err := db.Exec(`
		INSERT INTO messages
		(listing_id, sender, message)
		VALUES (?, ?, ?)
	`, req.ListingID, username, req.Message)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()

	c.JSON(201, gin.H{
		"id":         id,
		"listing_id": req.ListingID,
		"sender":     username,
		"message":    req.Message,
	})
}

func listMessages(c *gin.Context) {
	listingID := c.Param("listing_id")

	rows, err := db.Query(`
		SELECT id, listing_id, sender, message, created_at
		FROM messages
		WHERE listing_id = ?
		ORDER BY id ASC
	`, listingID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var messages []gin.H

	for rows.Next() {
		var (
			id        uint64
			listingID uint64
			sender    string
			message   string
			createdAt string
		)

		if err := rows.Scan(&id, &listingID, &sender, &message, &createdAt); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		messages = append(messages, gin.H{
			"id":         id,
			"listing_id": listingID,
			"sender":     sender,
			"message":    message,
			"created_at": createdAt,
		})
	}

	c.JSON(200, messages)
}

type CreateReviewRequest struct {
	SellerID uint64 `json:"seller_id"`
	Rating   int    `json:"rating"`
	Comment  string `json:"comment"`
}

func createReview(c *gin.Context) {
	reviewerID := c.MustGet("user_id").(uint64)

	var req CreateReviewRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if req.SellerID == 0 || req.Rating < 1 || req.Rating > 5 {
		c.JSON(400, gin.H{
			"error": "seller_id and rating (1-5) are required",
		})
		return
	}

	if reviewerID == req.SellerID {
		c.JSON(400, gin.H{
			"error": "cannot review yourself",
		})
		return
	}

	result, err := db.Exec(`
		INSERT INTO reviews
		(seller_id, reviewer_id, rating, comment)
		VALUES (?, ?, ?, ?)
	`,
		req.SellerID,
		reviewerID,
		req.Rating,
		req.Comment,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()

	c.JSON(201, gin.H{
		"id":          id,
		"seller_id":   req.SellerID,
		"reviewer_id": reviewerID,
		"rating":      req.Rating,
		"comment":     req.Comment,
	})
}

func listUserReviews(c *gin.Context) {
	sellerID := c.Param("id")

	rows, err := db.Query(`
		SELECT
			r.id,
			r.seller_id,
			r.reviewer_id,
			COALESCE(u.username, ''),
			r.rating,
			COALESCE(r.comment, ''),
			r.created_at
		FROM reviews r
		LEFT JOIN users u ON u.id = r.reviewer_id
		WHERE r.seller_id = ?
		ORDER BY r.id DESC
	`, sellerID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	reviews := []gin.H{}

	for rows.Next() {
		var (
			id               uint64
			sellerIDValue    uint64
			reviewerID       uint64
			reviewerUsername string
			rating           int
			comment          string
			createdAt        string
		)

		if err := rows.Scan(
			&id,
			&sellerIDValue,
			&reviewerID,
			&reviewerUsername,
			&rating,
			&comment,
			&createdAt,
		); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		reviews = append(reviews, gin.H{
			"id":                id,
			"seller_id":         sellerIDValue,
			"reviewer_id":       reviewerID,
			"reviewer_username": reviewerUsername,
			"rating":            rating,
			"comment":           comment,
			"created_at":        createdAt,
		})
	}

	c.JSON(200, gin.H{"reviews": reviews})
}

func getUserProfile(c *gin.Context) {
	id := c.Param("id")

	var (
		userID       uint64
		username     string
		bchAddress   string
		tokenAddress string
		ratingAvg    float64
		reviewCount  uint64
		listingCount uint64
	)

	err := db.QueryRow(`
		SELECT
			u.id,
			u.username,
			u.bch_address,
			u.token_address,
			COALESCE(AVG(r.rating), 0),
			COUNT(DISTINCT r.id),
			COUNT(DISTINCT l.id)
		FROM users u
		LEFT JOIN reviews r ON r.seller_id = u.id
		LEFT JOIN listings l ON l.user_id = u.id
		WHERE u.id = ?
		GROUP BY u.id
	`, id).Scan(
		&userID,
		&username,
		&bchAddress,
		&tokenAddress,
		&ratingAvg,
		&reviewCount,
		&listingCount,
	)

	if err != nil {
		c.JSON(404, gin.H{"error": "user not found"})
		return
	}

	rows, err := db.Query(`
		SELECT
			id,
			title,
			description,
			price,
			currency,
			COALESCE(category, ''),
			COALESCE(location, ''),
			COALESCE(image_url, ''),
			status,
			created_at
		FROM listings
		WHERE user_id = ?
		ORDER BY id DESC
	`, id)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var listings []gin.H

	for rows.Next() {
		var (
			listingID   uint64
			title       string
			description string
			price       float64
			currency    string
			category    string
			location    string
			imageURL    string
			status      string
			createdAt   string
		)

		if err := rows.Scan(
			&listingID,
			&title,
			&description,
			&price,
			&currency,
			&category,
			&location,
			&imageURL,
			&status,
			&createdAt,
		); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		listings = append(listings, gin.H{
			"id":          listingID,
			"title":       title,
			"description": description,
			"price":       price,
			"currency":    currency,
			"category":    category,
			"location":    location,
			"image_url":   imageURL,
			"status":      status,
			"created_at":  createdAt,
		})
	}

	c.JSON(200, gin.H{
		"id":            userID,
		"username":      username,
		"bch_address":   bchAddress,
		"token_address": tokenAddress,
		"rating_avg":    ratingAvg,
		"review_count":  reviewCount,
		"listing_count": listingCount,
		"listings":      listings,
	})
}

type UpdateOrderStatusRequest struct {
	Status          string `json:"status"`
	TrackingNumber string `json:"tracking_number"`
	ShippingCarrier string `json:"shipping_carrier"`
}

func updateOrderStatus(c *gin.Context) {
	userID := c.MustGet("user_id").(uint64)
	id := c.Param("id")

	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	allowed := map[string]bool{
		"shipped":   true,
		"cancelled": true,
	}

	if !allowed[req.Status] {
		c.JSON(400, gin.H{"error": "invalid status"})
		return
	}

	var (
		buyerUserID   uint64
		sellerUserID  uint64
		currentStatus string
	)

	err := db.QueryRow(`
		SELECT 
			o.buyer_user_id,
			l.user_id,
			o.status
		FROM orders o
		JOIN listings l ON l.id = o.listing_id
		WHERE o.id = ?
	`, id).Scan(&buyerUserID, &sellerUserID, &currentStatus)

	if err != nil {
		c.JSON(404, gin.H{"error": "order not found"})
		return
	}
	

	if req.Status == "shipped" && userID != sellerUserID {
		c.JSON(403, gin.H{"error": "only seller can mark shipped"})
		return
	}


	if req.Status == "cancelled" && userID != buyerUserID && userID != sellerUserID {
		c.JSON(403, gin.H{"error": "not allowed"})
		return
	}

	if req.Status == "shipped" && currentStatus != "paid" {
		c.JSON(400, gin.H{"error": "order must be paid before marking shipped"})
		return
	}

	result, err := db.Exec(`
		UPDATE orders
		SET
			status = ?,
			tracking_number = ?,
			shipping_carrier = ?,
			shipped_at = CASE
				WHEN ? = 'shipped' AND shipped_at IS NULL
				THEN NOW()
				ELSE shipped_at
			END
		WHERE id = ?
	`,
		req.Status,
		req.TrackingNumber,
		req.ShippingCarrier,
		req.Status,
		id,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(404, gin.H{"error": "order not found"})
		return
	}

	if req.Status == "shipped" {
		sendOrderEmailAsync(id, "shipped")
	}

	c.JSON(200, gin.H{
		"id":     id,
		"status": req.Status,
	})
}

func buildPaymentURI(address string, amount float64, currency string) string {
	switch currency {
	case "BCH":
		return fmt.Sprintf("%s?amount=%.8f", address, amount)

	case "PUSD":
		baseUnits := int64(amount * 100)
		return fmt.Sprintf("%s?c=%s&ft=%d", address, PUSDCategory, baseUnits)

	case "MUSD":
		baseUnits := int64(amount * 100)
		return fmt.Sprintf("%s?c=%s&ft=%d", address, MUSDCategory, baseUnits)

	default:
		return ""
	}
}

func verifyOrderPayment(c *gin.Context) {
	id := c.Param("id")

	var (
		orderID         uint64
		contractAddress string
		amount          float64
		currency        string
		status          string
		expiresAt       time.Time
	)

	err := db.QueryRow(`
		SELECT id, contract_address, amount, currency, status, expires_at
		FROM orders
		WHERE id = ?
	`, id).Scan(
		&orderID,
		&contractAddress,
		&amount,
		&currency,
		&status,
		&expiresAt,
	)

	if err != nil {
		c.JSON(404, gin.H{"error": "order not found"})
		return
	}

	if contractAddress == "" {
		c.JSON(400, gin.H{"error": "order has no escrow contract address"})
		return
	}

	if status != "pending" {
		c.JSON(200, gin.H{
			"message": "order is not pending",
			"status":  status,
		})
		return
	}

	if time.Now().After(expiresAt) {
		_, _ = db.Exec(`
			UPDATE orders
			SET status = 'expired'
			WHERE id = ? AND status = 'pending'
		`, orderID)

		c.JSON(200, gin.H{
			"paid":   false,
			"status": "expired",
			"error":  "order expired",
		})
		return
	}

	paid, txid, err := checkPayment(contractAddress, amount, currency)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if !paid {
		c.JSON(200, gin.H{
			"paid":   false,
			"status": status,
		})
		return
	}

	result, err := db.Exec(`
		UPDATE orders
		SET status = 'paid', txid = ?
		WHERE id = ? AND status = 'pending'
	`, txid, orderID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		sendOrderEmailAsync(orderID, "paid")
	}

	c.JSON(200, gin.H{
		"paid":   true,
		"status": "paid",
		"txid":   txid,
	})
}

func checkPayment(address string, amount float64, currency string) (bool, string, error) {
	if currency == "BCH" {
		return checkBCHPayment(address, amount)
	}

	if currency == "PUSD" {
		return checkTokenPayment(address, PUSDCategory, int64(amount*100+0.5))
	}

	if currency == "MUSD" {
		return checkTokenPayment(address, MUSDCategory, int64(amount*100+0.5))
	}

	return false, "", nil
}

func uploadImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(400, gin.H{"error": "image is required"})
		return
	}

	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
	path := "./uploads/" + filename

	if err := c.SaveUploadedFile(file, path); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{
		"image_url": "/uploads/" + filename,
	})
}

func recordClaim(c *gin.Context) {
	userID := c.MustGet("user_id").(uint64)
	id := c.Param("id")

	var req TxidRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Txid == "" {
		c.JSON(400, gin.H{"error": "txid is required"})
		return
	}

	var (
		sellerUserID uint64
		listingID    uint64
		status       string
		claimTxid    sql.NullString
	)

	err := db.QueryRow(`
		SELECT
			l.user_id,
			o.listing_id,
			o.status,
			o.claim_txid
		FROM orders o
		JOIN listings l ON l.id = o.listing_id
		WHERE o.id = ?
	`, id).Scan(
		&sellerUserID,
		&listingID,
		&status,
		&claimTxid,
	)

	if err != nil {
		c.JSON(404, gin.H{"error": "order not found"})
		return
	}

	if userID != sellerUserID {
		c.JSON(403, gin.H{"error": "only seller can record claim"})
		return
	}

	if status != "shipped" {
		c.JSON(400, gin.H{"error": "order must be shipped before seller can claim"})
		return
	}

	if claimTxid.Valid && claimTxid.String != "" {
		c.JSON(400, gin.H{"error": "order has already been claimed"})
		return
	}

	result, err := db.Exec(`
		UPDATE orders
		SET
			status = 'completed',
			claim_txid = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		  AND status = 'shipped'
		  AND (claim_txid IS NULL OR claim_txid = '')
	`, req.Txid, id)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(400, gin.H{"error": "order is no longer claimable"})
		return
	}

	

	_, _ = db.Exec(`
		UPDATE listings
		SET status = 'sold'
		WHERE id = ?
	`, listingID)

	sendOrderEmailAsync(id, "completed")

	c.JSON(200, gin.H{
		"status":     "completed",
		"claim_txid": req.Txid,
	})
}


func recordRefund(c *gin.Context) {
	userID := c.MustGet("user_id").(uint64)
	id := c.Param("id")

	var req TxidRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Txid == "" {
		c.JSON(400, gin.H{"error": "txid is required"})
		return
	}

	var (
		buyerUserID uint64
		status      string
	)

	err := db.QueryRow(`
		SELECT buyer_user_id, status
		FROM orders
		WHERE id = ?
	`, id).Scan(&buyerUserID, &status)

	if err != nil {
		c.JSON(404, gin.H{"error": "order not found"})
		return
	}

	if userID != buyerUserID {
		c.JSON(403, gin.H{"error": "only buyer can record refund"})
		return
	}

	if status != "paid" && status != "expired" && status != "cancelled" {
		c.JSON(400, gin.H{"error": "order is not refundable"})
		return
	}

	_, err = db.Exec(`
		UPDATE orders
		SET status = 'refunded', refund_txid = ?
		WHERE id = ?
	`, req.Txid, id)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"status":      "refunded",
		"refund_txid": req.Txid,
	})
}

type CreateBidRequest struct {
	Amount float64 `json:"amount"`
}

func createBid(c *gin.Context) {
	userID := c.MustGet("user_id").(uint64)
	listingID := c.Param("id")

	var req CreateBidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if req.Amount <= 0 {
		c.JSON(400, gin.H{"error": "amount is required"})
		return
	}

	var bidderAddress string
	err := db.QueryRow(`
		SELECT bch_address
		FROM users
		WHERE id = ?
	`, userID).Scan(&bidderAddress)

	if err != nil {
		c.JSON(401, gin.H{"error": "user not found"})
		return
	}

	var (
		saleType      string
		currency      string
		startingPrice float64
		sellerUserID  uint64
	)

	err = db.QueryRow(`
		SELECT sale_type, currency, COALESCE(starting_price, price), user_id
		FROM listings
		WHERE id = ?
		AND status = 'active'
	`, listingID).Scan(&saleType, &currency, &startingPrice, &sellerUserID)

	if err != nil {
		c.JSON(404, gin.H{"error": "active listing not found"})
		return
	}

	if sellerUserID == userID {
		c.JSON(400, gin.H{"error": "cannot bid on your own listing"})
		return
	}

	if saleType != "auction" {
		c.JSON(400, gin.H{"error": "listing is not an auction"})
		return
	}

	var highestBid float64

	err = db.QueryRow(`
		SELECT COALESCE(MAX(amount), 0)
		FROM bids
		WHERE listing_id = ?
	`, listingID).Scan(&highestBid)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	minBid := startingPrice
	if highestBid > 0 {
		minBid = highestBid
	}

	if req.Amount <= minBid {
		c.JSON(400, gin.H{
			"error":       "bid must be higher than current price",
			"current_bid": minBid,
		})
		return
	}

	result, err := db.Exec(`
		INSERT INTO bids
		(listing_id, bidder_address, amount, currency)
		VALUES (?, ?, ?, ?)
	`, listingID, bidderAddress, req.Amount, currency)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	bidID, _ := result.LastInsertId()

	c.JSON(201, gin.H{
		"id":             bidID,
		"listing_id":     listingID,
		"bidder_user_id": userID,
		"bidder_address": bidderAddress,
		"amount":         req.Amount,
		"currency":       currency,
	})
}

func listBids(c *gin.Context) {
	listingID := c.Param("id")

	rows, err := db.Query(`
		SELECT id, listing_id, bidder_address, amount, currency, created_at
		FROM bids
		WHERE listing_id = ?
		ORDER BY amount DESC
	`, listingID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var bids []gin.H

	for rows.Next() {
		var (
			id            uint64
			listingIDNum  uint64
			bidderAddress string
			amount        float64
			currency      string
			createdAt     string
		)

		if err := rows.Scan(&id, &listingIDNum, &bidderAddress, &amount, &currency, &createdAt); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		bids = append(bids, gin.H{
			"id":             id,
			"listing_id":     listingIDNum,
			"bidder_address": bidderAddress,
			"amount":         amount,
			"currency":       currency,
			"created_at":     createdAt,
		})
	}

	c.JSON(200, bids)
}

type LoginRequest struct {
	Username  string `json:"username"`
	Address   string `json:"address"`
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

func getAuthNonce(c *gin.Context) {
	username := c.Param("username")
	nonce := uuid.NewString()

	message := "Login to BCHBazaar: " + nonce

	_, err := db.Exec(`
		UPDATE users
		SET auth_nonce = ?
		WHERE username = ?
	`, nonce, username)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": message,
		"nonce":   nonce,
	})
}

func loginWithSignature(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var userID uint64
	var savedNonce string
	var savedAddress string

	err := db.QueryRow(`
		SELECT id, auth_nonce, bch_address
		FROM users
		WHERE username = ?
	`, req.Username).Scan(&userID, &savedNonce, &savedAddress)

	if err != nil {
		c.JSON(401, gin.H{"error": "user not found"})
		return
	}

	expectedMessage := "Login to BCHBazaar: " + savedNonce

	if savedNonce == "" {
		c.JSON(401, gin.H{"error": "no login nonce found"})
		return
	}

	if req.Message != expectedMessage {
		c.JSON(401, gin.H{"error": "invalid login message"})
		return
	}

	if cleanAddress(req.Address) != cleanAddress(savedAddress) {
		c.JSON(401, gin.H{"error": "address does not match user"})
		return
	}

	if !verifySignature(
		req.Address,
		req.Message,
		req.Signature,
	) {
		c.JSON(401, gin.H{
			"error": "invalid signature",
		})
		return
	}

	sessionToken := uuid.NewString()

	_, err = db.Exec(`
		UPDATE users
		SET session_token = ?, auth_nonce = NULL
		WHERE username = ?
	`, sessionToken, req.Username)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"session_token": sessionToken,
		"user_id":       userID,
	})
}

func cleanAddress(address string) string {
	return strings.TrimPrefix(address, "bitcoincash:")
}

type TokenPaymentRequest struct {
	Address  string `json:"address"`
	Category string `json:"category"`
	Amount   int64  `json:"amount"`
}

type TokenPaymentResponse struct {
	Paid        bool   `json:"paid"`
	Txid        string `json:"txid"`
	TokenAmount string `json:"token_amount"`
	Required    string `json:"required"`
	Error       string `json:"error"`
}

func checkTokenPayment(address, category string, amount int64) (bool, string, error) {
	reqBody := TokenPaymentRequest{
		Address:  address,
		Category: category,
		Amount:   amount,
	}

	data, _ := json.Marshal(reqBody)

	resp, err := http.Post(
		"http://localhost:8788/token-payment",
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	var result TokenPaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, "", err
	}
	if result.Error != "" {
		return false, "", fmt.Errorf("%s", result.Error)
	}

	return result.Paid, result.Txid, nil
}

type VerifyRequest struct {
	Address   string `json:"address"`
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

type VerifyResponse struct {
	Valid bool `json:"valid"`
}

func verifySignature(address, message, signature string) bool {
	reqBody := VerifyRequest{
		Address:   address,
		Message:   message,
		Signature: signature,
	}

	data, _ := json.Marshal(reqBody)

	resp, err := http.Post(
		"http://localhost:8788/verify",
		"application/json",
		bytes.NewBuffer(data),
	)

	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var result VerifyResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false
	}

	return result.Valid
}

type OpenDisputeRequest struct {
	Reason string `json:"reason"`
}

type ModeratorTxRequest struct {
	Txid string `json:"txid"`
}

func openDispute(c *gin.Context) {
	userID := c.MustGet("user_id").(uint64)
	id := c.Param("id")

	var req OpenDisputeRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Reason == "" {
		c.JSON(400, gin.H{"error": "reason is required"})
		return
	}

	var (
		buyerUserID     uint64
		sellerUserID    uint64
		moderatorUserID sql.NullInt64
		status          string
		disputeStatus   string
	)

	err := db.QueryRow(`
		SELECT o.buyer_user_id, l.user_id, o.moderator_user_id, o.status, o.dispute_status
		FROM orders o
		JOIN listings l ON l.id = o.listing_id
		WHERE o.id = ?
	`, id).Scan(&buyerUserID, &sellerUserID, &moderatorUserID, &status, &disputeStatus)

	if err != nil {
		c.JSON(404, gin.H{"error": "order not found"})
		return
	}

	if userID != buyerUserID && userID != sellerUserID {
		c.JSON(403, gin.H{"error": "not part of this order"})
		return
	}

	if !moderatorUserID.Valid {
		c.JSON(400, gin.H{"error": "order has no assigned moderator"})
		return
	}

	if disputeStatus != "none" {
		c.JSON(400, gin.H{"error": "dispute is already open or resolved"})
		return
	}

	if status != "paid" && status != "shipped" && status != "completed" {
		c.JSON(400, gin.H{"error": "order is not disputable"})
		return
	}

	_, err = db.Exec(`
		UPDATE orders
		SET dispute_status = 'open',
		    dispute_reason = ?
		WHERE id = ? AND dispute_status = 'none'
	`, req.Reason, id)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"status":         status,
		"dispute_status": "open",
		"dispute_reason": req.Reason,
	})
}

func recordModeratorRelease(c *gin.Context) {
	userID := c.MustGet("user_id").(uint64)
	id := c.Param("id")

	var req ModeratorTxRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Txid == "" {
		c.JSON(400, gin.H{"error": "txid is required"})
		return
	}

	var moderatorUserID uint64
	var disputeStatus string

	err := db.QueryRow(`
		SELECT moderator_user_id, dispute_status
		FROM orders
		WHERE id = ?
	`, id).Scan(&moderatorUserID, &disputeStatus)

	if err != nil {
		c.JSON(404, gin.H{"error": "order not found"})
		return
	}

	if userID != moderatorUserID {
		c.JSON(403, gin.H{"error": "only moderator can record release"})
		return
	}

	if disputeStatus != "open" {
		c.JSON(400, gin.H{"error": "dispute is not open"})
		return
	}

	_, err = db.Exec(`
		UPDATE orders
		SET status = 'claimed',
		    dispute_status = 'resolved',
		    moderator_decision = 'release_to_seller',
		    moderator_txid = ?,
		    claim_txid = ?
		WHERE id = ? AND dispute_status = 'open'
	`, req.Txid, req.Txid, id)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"status":             "claimed",
		"dispute_status":     "resolved",
		"moderator_decision": "release_to_seller",
		"moderator_txid":     req.Txid,
	})
}

func recordModeratorRefund(c *gin.Context) {
	userID := c.MustGet("user_id").(uint64)
	id := c.Param("id")

	var req ModeratorTxRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Txid == "" {
		c.JSON(400, gin.H{"error": "txid is required"})
		return
	}

	var moderatorUserID uint64
	var disputeStatus string

	err := db.QueryRow(`
		SELECT moderator_user_id, dispute_status
		FROM orders
		WHERE id = ?
	`, id).Scan(&moderatorUserID, &disputeStatus)

	if err != nil {
		c.JSON(404, gin.H{"error": "order not found"})
		return
	}

	if userID != moderatorUserID {
		c.JSON(403, gin.H{"error": "only moderator can record refund"})
		return
	}

	if disputeStatus != "open" {
		c.JSON(400, gin.H{"error": "dispute is not open"})
		return
	}

	_, err = db.Exec(`
		UPDATE orders
		SET status = 'refunded',
		    dispute_status = 'resolved',
		    moderator_decision = 'refund_to_buyer',
		    moderator_txid = ?,
		    refund_txid = ?
		WHERE id = ? AND dispute_status = 'open'
	`, req.Txid, req.Txid, id)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"status":             "refunded",
		"dispute_status":     "resolved",
		"moderator_decision": "refund_to_buyer",
		"moderator_txid":     req.Txid,
	})
}

func getMe(c *gin.Context) {
	userID := c.GetUint64("user_id")

	var user struct {
		ID                  uint64  `json:"id"`
		Username            string  `json:"username"`
		Email               string  `json:"email"`
		BCHAddress          string  `json:"bch_address"`
		TokenAddress        string  `json:"token_address"`
		Role                string  `json:"role"`
		ModeratorActive     bool    `json:"moderator_active"`
		ModeratorBio        string  `json:"moderator_bio"`
		ModeratorFeePercent float64 `json:"moderator_fee_percent"`
	}

	err := db.QueryRow(`
		SELECT id, username, COALESCE(email, ''), bch_address, token_address, role,
			moderator_active, COALESCE(moderator_bio, ''), moderator_fee_percent
		FROM users
		WHERE id = ?
	`, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.BCHAddress,
		&user.TokenAddress,
		&user.Role,
		&user.ModeratorActive,
		&user.ModeratorBio,
		&user.ModeratorFeePercent,
	)

	if err != nil {
		c.JSON(404, gin.H{"error": "user not found"})
		return
	}

	c.JSON(200, user)
}

func updateMe(c *gin.Context) {
	userID := c.GetUint64("user_id")

	var req struct {
		Email               string  `json:"email"`
		ModeratorActive     bool    `json:"moderator_active"`
		ModeratorBio        string  `json:"moderator_bio"`
		ModeratorFeePercent float64 `json:"moderator_fee_percent"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Email != "" && !strings.Contains(req.Email, "@") {
		c.JSON(400, gin.H{"error": "invalid email"})
		return
	}

	if req.ModeratorFeePercent < 0 || req.ModeratorFeePercent > 10 {
		c.JSON(400, gin.H{"error": "moderator fee must be between 0 and 10 percent"})
		return
	}

	_, err := db.Exec(`
		UPDATE users
		SET email = ?,
			moderator_active = ?,
			moderator_bio = ?,
			moderator_fee_percent = ?
		WHERE id = ?
	`, req.Email, req.ModeratorActive, req.ModeratorBio, req.ModeratorFeePercent, userID)

	if err != nil {
		c.JSON(500, gin.H{"error": "failed to update profile"})
		return
	}

	c.JSON(200, gin.H{"ok": true})
}

type CreateConversationRequest struct {
	ListingID uint64 `json:"listing_id"`
}

type CreateConversationMessageRequest struct {
	Message string `json:"message"`
}

func createConversation(c *gin.Context) {
	userID := c.MustGet("user_id").(uint64)

	var req CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.ListingID == 0 {
		c.JSON(400, gin.H{"error": "listing_id is required"})
		return
	}

	var sellerUserID uint64
	err := db.QueryRow(`
		SELECT user_id
		FROM listings
		WHERE id = ?
	`, req.ListingID).Scan(&sellerUserID)

	if err != nil {
		c.JSON(404, gin.H{"error": "listing not found"})
		return
	}

	if userID == sellerUserID {
		c.JSON(400, gin.H{"error": "cannot message yourself"})
		return
	}

	var conversationID uint64
	err = db.QueryRow(`
		SELECT id
		FROM conversations
		WHERE listing_id = ?
		AND buyer_user_id = ?
		AND seller_user_id = ?
	`, req.ListingID, userID, sellerUserID).Scan(&conversationID)

	if err == nil {
		c.JSON(200, gin.H{"conversation_id": conversationID})
		return
	}

	result, err := db.Exec(`
		INSERT INTO conversations
		(listing_id, buyer_user_id, seller_user_id)
		VALUES (?, ?, ?)
	`, req.ListingID, userID, sellerUserID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(201, gin.H{"conversation_id": id})
}

func listConversations(c *gin.Context) {
	userID := c.MustGet("user_id").(uint64)

	rows, err := db.Query(`
		SELECT
			c.id,
			c.listing_id,
			l.title,
			c.buyer_user_id,
			b.username,
			c.seller_user_id,
			s.username,
			c.updated_at
		FROM conversations c
		JOIN listings l ON l.id = c.listing_id
		JOIN users b ON b.id = c.buyer_user_id
		JOIN users s ON s.id = c.seller_user_id
		WHERE c.buyer_user_id = ?
		OR c.seller_user_id = ?
		ORDER BY c.updated_at DESC
	`, userID, userID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	conversations := []gin.H{}

	for rows.Next() {
		var (
			id, listingID, buyerUserID, sellerUserID uint64
			title, buyerUsername, sellerUsername     string
			updatedAt                               string
		)

		if err := rows.Scan(
			&id,
			&listingID,
			&title,
			&buyerUserID,
			&buyerUsername,
			&sellerUserID,
			&sellerUsername,
			&updatedAt,
		); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		conversations = append(conversations, gin.H{
			"id":              id,
			"listing_id":      listingID,
			"listing_title":   title,
			"buyer_user_id":   buyerUserID,
			"buyer_username":  buyerUsername,
			"seller_user_id":  sellerUserID,
			"seller_username": sellerUsername,
			"updated_at":      updatedAt,
		})
	}

	c.JSON(200, conversations)
}

func listConversationMessages(c *gin.Context) {
	userID := c.MustGet("user_id").(uint64)
	conversationID := c.Param("id")

	if !userCanAccessConversation(userID, conversationID) {
		c.JSON(403, gin.H{"error": "not allowed"})
		return
	}

	_, _ = db.Exec(`
	UPDATE messages
	SET read_at = NOW()
	WHERE conversation_id = ?
	  AND read_at IS NULL
	  AND (sender_user_id IS NULL OR sender_user_id != ?)
`, conversationID, userID)

	rows, err := db.Query(`
		SELECT
			m.id,
			m.sender_user_id,
			u.username,
			m.message,
			m.created_at
		FROM messages m
		JOIN users u ON u.id = m.sender_user_id
		WHERE m.conversation_id = ?
		ORDER BY m.id ASC
	`, conversationID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	messages := []gin.H{}

	for rows.Next() {
		var (
			id, senderUserID uint64
			username         string
			message          string
			createdAt        string
		)

		if err := rows.Scan(&id, &senderUserID, &username, &message, &createdAt); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		messages = append(messages, gin.H{
			"id":             id,
			"sender_user_id": senderUserID,
			"sender":         username,
			"message":        message,
			"created_at":     createdAt,
		})
	}

	c.JSON(200, messages)
}

func createConversationMessage(c *gin.Context) {
	userID := c.MustGet("user_id").(uint64)
	conversationID := c.Param("id")

	if !userCanAccessConversation(userID, conversationID) {
		c.JSON(403, gin.H{"error": "not allowed"})
		return
	}

	var req CreateConversationMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Message) == "" {
		c.JSON(400, gin.H{"error": "message is required"})
		return
	}

	var listingID uint64
	err := db.QueryRow(`
		SELECT listing_id
		FROM conversations
		WHERE id = ?
	`, conversationID).Scan(&listingID)

	if err != nil {
		c.JSON(404, gin.H{"error": "conversation not found"})
		return
	}

	var username string
	_ = db.QueryRow(`SELECT username FROM users WHERE id = ?`, userID).Scan(&username)

	result, err := db.Exec(`
		INSERT INTO messages
		(listing_id, conversation_id, sender_user_id, sender, message)
		VALUES (?, ?, ?, ?, ?)
	`, listingID, conversationID, userID, username, req.Message)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	_, _ = db.Exec(`
		UPDATE conversations
		SET updated_at = NOW()
		WHERE id = ?
	`, conversationID)

	messageID, _ := result.LastInsertId()

	c.JSON(201, gin.H{
		"id":              messageID,
		"conversation_id": conversationID,
		"sender_user_id":  userID,
		"sender":          username,
		"message":         req.Message,
	})
}

func userCanAccessConversation(userID uint64, conversationID string) bool {
	var exists int

	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM conversations
		WHERE id = ?
		AND (buyer_user_id = ? OR seller_user_id = ?)
	`, conversationID, userID, userID).Scan(&exists)

	return err == nil && exists > 0
}


func getMyNotifications(c *gin.Context) {
	userID := c.MustGet("user_id").(uint64)

	var sellerActions int
	var unreadMessages int

	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM orders o
		JOIN listings l ON l.id = o.listing_id
		WHERE l.user_id = ?
		  AND (
		    o.status = 'paid'
		    OR (
		      o.status = 'shipped'
		      AND (o.claim_txid IS NULL OR o.claim_txid = '')
		    )
		  )
	`, userID).Scan(&sellerActions)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM messages m
		JOIN conversations c ON c.id = m.conversation_id
		WHERE m.read_at IS NULL
		  AND (m.sender_user_id IS NULL OR m.sender_user_id != ?)
		  AND (
		    c.buyer_user_id = ?
		    OR c.seller_user_id = ?
		  )
	`, userID, userID, userID).Scan(&unreadMessages)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"unread_messages": unreadMessages,
		"seller_actions":  sellerActions,
	})
}