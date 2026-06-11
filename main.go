package main

import (
	"log"
	"time"
	"github.com/gin-gonic/gin"
)

func main() {
	connectDB()
	defer db.Close()

	migrateDB()
	go startOrderWatcher()

	r := gin.Default()

	r.GET("/api/health", healthCheck)

	r.POST("/api/users", createUser)
	r.GET("/api/listings", listListings)
	r.POST("/api/listings", authRequired, createListing)
	r.GET("/api/listings/:id", getListing)
	r.POST("/api/orders", authRequired, createOrder)
    r.GET("/api/orders/:id", getOrder)
	r.POST("/api/messages", authRequired, createMessage)
	r.GET("/api/messages/:listing_id", listMessages)
	r.POST("/api/reviews", authRequired, createReview)
	r.GET("/api/users/:id/reviews", listUserReviews)
	r.GET("/api/users/:id/profile", getUserProfile)
	r.PATCH("/api/orders/:id/status", authRequired, updateOrderStatus)
	r.POST("/api/orders/:id/verify", verifyOrderPayment)
	r.POST("/api/uploads", uploadImage)
	r.Static("/uploads", "./uploads")
	r.POST("/api/orders/:id/claim", authRequired, recordClaim)
	r.POST("/api/orders/:id/refund", authRequired, recordRefund)
	r.POST("/api/listings/:id/bids", authRequired, createBid)
	r.GET("/api/listings/:id/bids", listBids)
	r.GET("/api/auth/nonce/:username", getAuthNonce)
	r.POST("/api/auth/login", loginWithSignature)



	log.Println("BCHBazaar API running on :8080")

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"ok": true,
		"app": "BCHBazaar",
	})
}

func startOrderWatcher() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		verifyPendingOrders()
		<-ticker.C
	}
}


func verifyPendingOrders() {
	rows, err := db.Query(`
		SELECT id
		FROM orders
		WHERE status = 'pending'
		AND contract_address IS NOT NULL
		AND contract_address != ''
		AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY id ASC
		LIMIT 20
	`)
	if err != nil {
		log.Println("watcher query error:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var orderID uint64
		if err := rows.Scan(&orderID); err != nil {
			log.Println("watcher scan error:", err)
			continue
		}

		if err := verifyOrderByID(orderID); err != nil {
			log.Println("verify order error:", orderID, err)
		}
	}
}

func verifyOrderByID(orderID uint64) error {
	var (
		contractAddress string
		amount         float64
		currency       string
		status         string
	)

	err := db.QueryRow(`
		SELECT contract_address, amount, currency, status
		FROM orders
		WHERE id = ?
	`, orderID).Scan(&contractAddress, &amount, &currency, &status)

	if err != nil {
		return err
	}

	if contractAddress == "" {
		return nil
	}

	if status != "pending" {
		return nil
	}

	paid, txid, err := checkPayment(contractAddress, amount, currency)
	if err != nil {
		return err
	}

	if !paid {
		return nil
	}

	_, err = db.Exec(`
		UPDATE orders
		SET status = 'paid', txid = ?
		WHERE id = ? AND status = 'pending'
	`, txid, orderID)

	return err
}

type TxidRequest struct {
	Txid string `json:"txid"`
}

