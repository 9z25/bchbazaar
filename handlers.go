package main

import (
	"net/http"
	"fmt"
	"github.com/gin-gonic/gin"
	"database/sql"
	"time"
	"github.com/google/uuid"
	"strings"
	"bytes"
	"encoding/json"
)

type CreateUserRequest struct {
	Username     string `json:"username"`
	BCHAddress   string `json:"bch_address"`
	TokenAddress string `json:"token_address"`
}

type CreateListingRequest struct {
	UserID      uint64  `json:"user_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
	ImageURL    string  `json:"image_url"`
	Category    string  `json:"category"`
}

func createUser(c *gin.Context) {
	var req CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if req.Username == "" || req.BCHAddress == "" || req.TokenAddress == "" {
		c.JSON(400, gin.H{"error": "username, bch_address, and token_address are required"})
		return
	}

	_, err := db.Exec(`
		INSERT INTO users
		(username, bch_address, token_address)
		VALUES (?, ?, ?)
	`, req.Username, req.BCHAddress, req.TokenAddress)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "user created"})
}

func createListing(c *gin.Context) {
	userID := c.MustGet("user_id").(uint64)

	var req CreateListingRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := db.Exec(`
		INSERT INTO listings
		(user_id, title, description, price, currency, category, image_url)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, userID, req.Title, req.Description, req.Price, req.Currency, req.Category, req.ImageURL)

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
			COALESCE(l.image_url, ''),
			l.created_at
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
			id          uint64
			userID      uint64
			username    string
			title       string
			description string
			price       float64
			currency    string
			imageURL    string
			createdAt   string
		)

		if err := rows.Scan(
			&id,
			&userID,
			&username,
			&title,
			&description,
			&price,
			&currency,
			&imageURL,
			&createdAt,
		); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		listings = append(listings, gin.H{
			"id":          id,
			"user_id":     userID,
			"seller":      username,
			"title":       title,
			"description": description,
			"price":       price,
			"currency":    currency,
			"image_url":   imageURL,
			"created_at":  createdAt,
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
			u.pusd_address,
			l.title,
			l.description,
			l.price_pusd,
			COALESCE(l.image_url, ''),
			l.created_at
		FROM listings l
		JOIN users u ON u.id = l.user_id
		WHERE l.id = ?
	`, id)

	var (
		listingID   uint64
		userID      uint64
		username    string
		pusdAddress string
		title       string
		description string
		pricePUSD   float64
		imageURL    string
		createdAt   string
	)

	err := row.Scan(
		&listingID,
		&userID,
		&username,
		&pusdAddress,
		&title,
		&description,
		&pricePUSD,
		&imageURL,
		&createdAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "listing not found"})
		return
	}

	listing["id"] = listingID
	listing["user_id"] = userID
	listing["seller"] = username
	listing["pusd_address"] = pusdAddress
	listing["title"] = title
	listing["description"] = description
	listing["price_pusd"] = pricePUSD
	listing["image_url"] = imageURL
	listing["created_at"] = createdAt

	c.JSON(http.StatusOK, listing)
}

const PUSDCategory = "2469acc5afa4b10cb5b5c04afb89c3a3ffd61c5da9c01e26d00951cae2a02544"
const MUSDCategory = "b38a33f750f84c5c169a6f23cb873e6e79605021585d4f3408789689ed87f366"

type CreateOrderRequest struct {
	ListingID        uint64 `json:"listing_id"`
	BuyerAddress    string `json:"buyer_address"`
	ContractAddress string `json:"contract_address"`
	SellerPKH       string `json:"seller_pkh"`
	BuyerPKH        string `json:"buyer_pkh"`
	RefundLocktime  uint64 `json:"refund_locktime"`
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
		amount       float64
		currency     string
		sellerUserID uint64
		bchAddress   string
		tokenAddress string
	)

	err := db.QueryRow(`
		SELECT
			l.price,
			l.currency,
			l.user_id,
			u.bch_address,
			u.token_address
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
	)

	if err != nil {
		c.JSON(404, gin.H{"error": "active listing not found"})
		return
	}

	if buyerUserID == sellerUserID {
		c.JSON(400, gin.H{"error": "cannot buy your own listing"})
		return
	}

	sellerAddress := bchAddress
	if currency != "BCH" {
		sellerAddress = tokenAddress
	}

	paymentAddress := req.ContractAddress
	expiresAt := time.Now().Add(24 * time.Hour)

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
			refund_locktime,
			amount,
			currency,
			expires_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		req.ListingID,
		buyerUserID,
		req.BuyerAddress,
		sellerAddress,
		paymentAddress,
		req.ContractAddress,
		req.SellerPKH,
		req.BuyerPKH,
		req.RefundLocktime,
		amount,
		currency,
		expiresAt,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	orderID, _ := result.LastInsertId()

	c.JSON(201, gin.H{
		"order_id":         orderID,
		"listing_id":       req.ListingID,
		"buyer_user_id":    buyerUserID,
		"amount":           amount,
		"currency":         currency,
		"seller_address":   sellerAddress,
		"payment_address":  paymentAddress,
		"contract_address": req.ContractAddress,
		"seller_pkh":       req.SellerPKH,
		"buyer_pkh":        req.BuyerPKH,
		"refund_locktime":  req.RefundLocktime,
		"payment_uri":      buildPaymentURI(paymentAddress, amount, currency),
		"status":           "pending",
		"expires_at":       expiresAt,
	})
}

func getOrder(c *gin.Context) {
	id := c.Param("id")

	var (
		orderID       uint64
		listingID     uint64
		buyerAddress  string
		sellerAddress string
		paymentAddress string
		amount        float64
		currency      string
		status        string
		txid          sql.NullString
		createdAt     string
		updatedAt     string
	)

	err := db.QueryRow(`
		SELECT id, listing_id, buyer_address, seller_address, payment_address, amount, currency, status, txid, created_at, updated_at
		FROM orders
		WHERE id = ?
	`, id).Scan(
		&orderID,
		&listingID,
		&buyerAddress,
		&sellerAddress,
		&paymentAddress,
		&amount,
		&currency,
		&status,
		&txid,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		c.JSON(404, gin.H{"error": "order not found"})
		return
	}

	c.JSON(200, gin.H{
		"id":              orderID,
		"listing_id":      listingID,
		"buyer_address":   buyerAddress,
		"seller_address":  sellerAddress,
		"payment_address": paymentAddress,
		"payment_uri":     buildPaymentURI(paymentAddress, amount, currency),
		"amount":          amount,
		"currency":        currency,
		"status":          status,
		"txid":            txid.String,
		"created_at":      createdAt,
		"updated_at":      updatedAt,
	})
}

func buildPUSDURI(address string, amount float64) string {
	baseUnits := int64(amount * 100)

	return  address +
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
	c.JSON(200, gin.H{"reviews": []any{}})
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
			"image_url":   imageURL,
			"status":      status,
			"created_at":  createdAt,
		})
	}

	c.JSON(200, gin.H{
		"id":             userID,
		"username":       username,
		"bch_address":    bchAddress,
		"token_address":  tokenAddress,
		"rating_avg":     ratingAvg,
		"review_count":   reviewCount,
		"listing_count":  listingCount,
		"listings":       listings,
	})
}
type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
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
		"completed": true,
		"cancelled": true,
	}

	if !allowed[req.Status] {
		c.JSON(400, gin.H{"error": "invalid status"})
		return
	}

	var (
		buyerUserID  uint64
		sellerUserID uint64
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

	if req.Status == "completed" && userID != buyerUserID {
		c.JSON(403, gin.H{"error": "only buyer can mark completed"})
		return
	}

	if req.Status == "cancelled" && userID != buyerUserID && userID != sellerUserID {
		c.JSON(403, gin.H{"error": "not allowed"})
		return
	}

	result, err := db.Exec(`
		UPDATE orders
		SET status = ?
		WHERE id = ?
	`, req.Status, id)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(404, gin.H{"error": "order not found"})
		return
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

	_, err = db.Exec(`
		UPDATE orders
		SET status = 'paid', txid = ?
		WHERE id = ? AND status = 'pending'
	`, txid, orderID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
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

	// PUSD/MUSD verification needs token tx inspection later.
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
	)

	err := db.QueryRow(`
		SELECT l.user_id, o.listing_id, o.status
		FROM orders o
		JOIN listings l ON l.id = o.listing_id
		WHERE o.id = ?
	`, id).Scan(&sellerUserID, &listingID, &status)

	if err != nil {
		c.JSON(404, gin.H{"error": "order not found"})
		return
	}

	if userID != sellerUserID {
		c.JSON(403, gin.H{"error": "only seller can record claim"})
		return
	}

	if status != "paid" && status != "shipped" && status != "completed" {
		c.JSON(400, gin.H{"error": "order is not claimable"})
		return
	}

	_, err = db.Exec(`
		UPDATE orders
		SET status = 'claimed', claim_txid = ?
		WHERE id = ?
	`, req.Txid, id)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	_, _ = db.Exec(`
		UPDATE listings
		SET status = 'sold'
		WHERE id = ?
	`, listingID)

	c.JSON(200, gin.H{
		"status":     "claimed",
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
		"nonce": nonce,
	})
}

func loginWithSignature(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var savedNonce string
	var savedAddress string

	err := db.QueryRow(`
		SELECT auth_nonce, bch_address
		FROM users
		WHERE username = ?
	`, req.Username).Scan(&savedNonce, &savedAddress)

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
	})
}

func cleanAddress(address string) string {
	return strings.TrimPrefix(address, "bitcoincash:")
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
		Address: address,
		Message: message,
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
		buyerUserID  uint64
		sellerUserID uint64
		status       string
	)

	err := db.QueryRow(`
		SELECT o.buyer_user_id, l.user_id, o.status
		FROM orders o
		JOIN listings l ON l.id = o.listing_id
		WHERE o.id = ?
	`, id).Scan(&buyerUserID, &sellerUserID, &status)

	if err != nil {
		c.JSON(404, gin.H{"error": "order not found"})
		return
	}

	if userID != buyerUserID && userID != sellerUserID {
		c.JSON(403, gin.H{"error": "not part of this order"})
		return
	}

	if status != "paid" && status != "shipped" && status != "completed" {
		c.JSON(400, gin.H{"error": "order is not disputable"})
		return
	}

	_, err = db.Exec(`
		UPDATE orders
		SET dispute_status = 'opened',
		    dispute_reason = ?
		WHERE id = ?
	`, req.Reason, id)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"status":         status,
		"dispute_status": "opened",
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

	if disputeStatus != "opened" {
		c.JSON(400, gin.H{"error": "dispute is not open"})
		return
	}

	_, err = db.Exec(`
		UPDATE orders
		SET status = 'claimed',
		    dispute_status = 'resolved',
		    moderator_decision = 'release',
		    moderator_txid = ?,
		    claim_txid = ?
		WHERE id = ?
	`, req.Txid, req.Txid, id)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"status":             "claimed",
		"dispute_status":     "resolved",
		"moderator_decision": "release",
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

	if disputeStatus != "opened" {
		c.JSON(400, gin.H{"error": "dispute is not open"})
		return
	}

	_, err = db.Exec(`
		UPDATE orders
		SET status = 'refunded',
		    dispute_status = 'resolved',
		    moderator_decision = 'refund',
		    moderator_txid = ?,
		    refund_txid = ?
		WHERE id = ?
	`, req.Txid, req.Txid, id)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"status":             "refunded",
		"dispute_status":     "resolved",
		"moderator_decision": "refund",
		"moderator_txid":     req.Txid,
	})
}