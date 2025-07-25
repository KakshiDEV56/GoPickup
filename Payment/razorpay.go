package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go_pickup/config"
	"go_pickup/utils"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/razorpay/razorpay-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RazorpayWebhookHandler(c *gin.Context) {
    secret := os.Getenv("RAZORPAY_WEBHOOK_SECRET")

    // Read the raw request body using io.ReadAll instead of ioutil.ReadAll
    body, err := io.ReadAll(c.Request.Body)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
        return
    }

    // Validate Razorpay signature from header
    razorSignature := c.GetHeader("X-Razorpay-Signature")
    h := hmac.New(sha256.New, []byte(secret))
    h.Write(body)
    expectedSignature := hex.EncodeToString(h.Sum(nil))

    if razorSignature != expectedSignature {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
        return
    }

    // Process webhook data as needed
    c.JSON(http.StatusOK, gin.H{"status": "Webhook received and verified"})
}
// here i will just do a thing that i will after sometime create a middleware than make sure that the middleware will verify the otp rather than using it manually in the functions
func CreatePaymentOrder(c *gin.Context) {
	var req struct {
		ParcelID string `json:"parcel_id"`
		Amount int `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	parcelObjID, err := primitive.ObjectIDFromHex(req.ParcelID)
fmt.Println(err)
	// 1. Check if parcel status is 'delivered'
	parcelCollection := config.GetCollection("parcels")
	var parcel bson.M
	err = parcelCollection.FindOne(context.TODO(), bson.M{"_id": parcelObjID}).Decode(&parcel)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Parcel not found"})
		return
	}
	status, ok := parcel["status"].(string)
	if !ok || status != "delivered" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Parcel not delivered yet"})
		return
	}

	// 2. Generate Razorpay Order
	client := razorpay.NewClient(os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_KEY_SECRET"))
	amount := req.Amount
	data := map[string]interface{}{
		"amount":          amount,              // must be int (e.g. 50000 for ₹500)
		"currency":        "INR",               // must be valid currency
		"receipt":         "receipt_" + req.ParcelID,
		"payment_capture": 1,                   // 1 means auto capture
	}

	order, err := client.Order.Create(data, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Razorpay order"})
		return
	}

	// 3. Store order ID linked with parcel in DB for tracking
	paymentCollection := config.GetCollection("payments")
	_, _ = paymentCollection.InsertOne(context.TODO(), bson.M{
		"parcel_id": parcelObjID, // ✅ Use ObjectID here
		"order_id":  order["id"],
		"status":    "created",
		"amount":    amount,
	})

	// 4. Send Razorpay Payment Page details to frontend
	c.JSON(http.StatusOK, gin.H{
		"order_id":    order["id"],
		"amount":      amount,
		"key_id":      os.Getenv("RAZORPAY_KEY_ID"),
		"payment_url": os.Getenv("PAYMENT_URL"), // For hosted page use case
	})
}

//First i will verify the otp right that is being created 
func VerifyParcelOtp(c *gin.Context) {
    var req struct {
        ParcelID string `json:"parcel_id"`
        Otp      string `json:"otp"`
    }
	
    if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }
	
    isValid, err := utils.VerifyOtp(req.ParcelID, req.Otp)
    if err != nil || !isValid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid OTP"})
        return
    }
	
    // Update parcel status in DB
	objID, err := primitive.ObjectIDFromHex(req.ParcelID)
	fmt.Println(err)
    parcelCollection := config.GetCollection("parcels")
    _, err = parcelCollection.UpdateOne(
        context.TODO(),
        bson.M{"_id": objID},
        bson.M{"$set": bson.M{"status": "delivered"}},
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"status": "OTP verified, parcel delivered"})
}