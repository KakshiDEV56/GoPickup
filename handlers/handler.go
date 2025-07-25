package handlers

import (
	"context"
	"fmt"
	services "go_pickup/auth"
	"go_pickup/cloudinary"
	"go_pickup/config"
	"go_pickup/email"
	"go_pickup/kafka"
	"go_pickup/models"
	"go_pickup/partner"
	"go_pickup/twilio"
	"go_pickup/utils"
	"math/rand"
	"net/http"
	"time"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//When the order is finally delievered it will send the otp to the user which is given to the agent who will put it
//and change the status to delievered
func GenerateAndSendOTP(c *gin.Context){
    // Get parcel ID from URL parameter or JSON body
    parcelID := c.Param("id")
    if parcelID == "" {
		//using the local struct is production grade or should i switch to some other method?
        var req struct {
            ParcelID string `json:"parcel_id binding:required"`
        }
        if err := c.ShouldBindJSON(&req); err != nil  {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Parcel ID is required"})
            return
        }
        parcelID = req.ParcelID
    }

    // Query the parcel collection to verify the parcel exists
    collection := config.GetCollection("parcels")
    objID, err := primitive.ObjectIDFromHex(parcelID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parcel ID"})
        return
    }

    var parcel models.Parcel
    err = collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&parcel)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Parcel not found"})
        return
    }

    // Generate OTP
    otp := fmt.Sprintf("%06d", rand.Intn(1000000))
    otpDoc := models.ParcelOTP{
        ParcelID:  objID,
        OTP:       otp,
        ExpiresAt: time.Now().Add(10 * time.Minute),
    }

    // Store OTP
    _, err = services.OtpCollection().InsertOne(context.TODO(), otpDoc)
    if err != nil {
		fmt.Println("OTP error ",err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store OTP"})
        return
    }

    // Send OTP via Twilio SMS to the receiver's phone
    message := fmt.Sprintf("Your parcel delivery OTP is: %s. Please provide it to the delivery agent to receive your parcel.", otp)
    twilio.SendSMS(parcel.ReceiverPhone, message)

    c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully"})

    // Send OTP via Twilio SMS


}
func UploadAgentProfile(c *gin.Context) {
	cld, ctx, err := cloudinary.InitCloudinary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cloudinary setup failed"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image file required"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to open file"})
		return
	}
	defer src.Close()

	// ✅ Generate a consistent publicID for this upload
	publicID := "user_" + uuid.New().String()

	// ✅ Upload to Cloudinary
	imageResp, err := cloudinary.UploadProfileImage(cld, ctx, src, publicID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "image upload failed"})
		return
	}
	imageURL := imageResp

	// ✅ Extract user ID from URL param
	userID := c.Param("id")
	if userID == "" {
		// Clean up uploaded image
		_, _ = cld.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID})
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user ID"})
		return
	}

	// ✅ Convert userID string to ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		// Clean up uploaded image
		_, _ = cld.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// ✅ Get MongoDB collection
	collection := config.GetCollection("agents")

	// ✅ Update user document with image URL
	_, err = collection.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"profile_image": imageURL}},
	)
	if err != nil {
		// Clean up uploaded image
		_, _ = cld.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save image URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"image_url": imageURL})
}

func UploadUserProfile(c *gin.Context) {
	cld, ctx, err := cloudinary.InitCloudinary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cloudinary setup failed"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image file required"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to open file"})
		return
	}
	defer src.Close()

	// ✅ Generate a consistent publicID for this upload
	publicID := "user_" + uuid.New().String()

	// ✅ Upload to Cloudinary
	imageResp, err := cloudinary.UploadProfileImage(cld, ctx, src, publicID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "image upload failed"})
		return
	}
	imageURL := imageResp

	// ✅ Extract user ID from URL param
	userID := c.Param("id")
	if userID == "" {
		// Clean up uploaded image
		_, _ = cld.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID})
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user ID"})
		return
	}

	// ✅ Convert userID string to ObjectID
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		// Clean up uploaded image
		_, _ = cld.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// ✅ Get MongoDB collection
	collection := config.GetCollection("users")

	// ✅ Update user document with image URL
	_, err = collection.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"profile_image": imageURL}},
	)
	if err != nil {
		// Clean up uploaded image
		_, _ = cld.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save image URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"image_url": imageURL})
}



// updating the status of the parcel pickup
func UpdateParcelStatusByAgent(c *gin.Context) {
	var req struct {
		ParcelID string `json:"parcel_id"`
		Status   string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	collection := config.GetCollection("parcels")
	objID, err := primitive.ObjectIDFromHex(req.ParcelID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parcel ID"})
		return
	}

	update := bson.M{"$set": bson.M{"status": req.Status}}
	_, err = collection.UpdateOne(context.Background(), bson.M{"_id": objID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Parcel status updated"})
}

func DeliveryPartnerRegistration(c *gin.Context) {
	var deliverypartner models.DeliveryPartner
	//fmt.Println(&deliverypartner)

	if err := c.ShouldBindJSON(&deliverypartner); err != nil {
		// DEBUG: print actual bind error
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid input",
			"details": err.Error(),
		})
		return
	}
	if err := partner.RegisterDeliveryPartner(deliverypartner); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Send SMS notification
	twilio.SendSMS(deliverypartner.Phone, "Congratulations! You have been successfully registered as a Delivery Partner with GoPickup. Welcome to the team! You can now log in and start accepting delivery assignments. If you have any questions, reply to this message or contact our support. Happy delivering!")
	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

func DeliveryPartnerLogin(c *gin.Context) {
	var credentials models.DeliveryPartner
	credentials.CreatedAt = time.Now()
	credentials.UpdatedAt = time.Now()
	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	partnerObj, err := partner.LoginDeliveryPartner(credentials.Email, credentials.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Generate JWT token (customize claims as needed)
	token, err := utils.GenerateJWT(partnerObj.ID, partnerObj.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func ParcelDetails(c *gin.Context) {

	collection := config.GetCollection("parcels")

	// Optionally, filter parcels by agent/user if you store that info
	// Example: agentID, _ := c.Get("user_id")
	// filter := bson.M{"agent_id": agentID}

	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch parcels"})
		return
	}
	defer cursor.Close(context.Background())

	var parcels []models.Parcel
	if err := cursor.All(context.Background(), &parcels); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not decode parcels"})
		return
	}

	c.JSON(http.StatusOK, parcels)
}

func ViewParcelStatus(c *gin.Context) {
	// i will just do the query from the database and gets the 
	var ParcelDetails models.Parcel
	var ParcelStatus =c.Param("status")
     if ParcelStatus==""{
		c.JSON(http.StatusBadGateway,gin.H{
			"message":"Status is empty ",
		})
		if err:=c.ShouldBindJSON(&ParcelDetails); err!=nil{
			c.JSON(404,gin.H{
				"message":"Invalid JSON Data ",
			})

		//just do the query from the data and check the status tag and change it to 
		// set what is there like the status could be not delivered yet ,delivered yet 

		}
		//now i just wanted to check and do the query of the 
	 }

	// in this i wanted to see if the parcel is accecpted by any driver to pickup and drop
	// also i wanted to see the  fare  i have to give to the person carrying the delivery box
}
func CreateParcel(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var parcel models.Parcel
	if err := c.BindJSON(&parcel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	parcel.ID = primitive.NewObjectID()
	parcel.UserID = userID
	parcel.Status = "Pending"
	parcel.CreatedAt = time.Now()
	parcel.UpdatedAt = time.Now()

	collection := config.GetCollection("parcels")
	_, err = collection.InsertOne(context.Background(), parcel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create parcel"})
		return
	}
	//Fetching the email data and then sending the notification via the email about the creation of parcel
	userCollection := config.GetCollection("users")
	var user models.User
	err = userCollection.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err == nil && user.Email != "" {
		go email.SendEmail(
			user.Email,
			"Your Parcel Has Been Created!",
			fmt.Sprintf(
				`Hello %s,

Your parcel booking has been successfully created on GoPickup!

**Pickup Address:** %s
**Drop Address:** %s
**Receiver Name:** %s
**Receiver Phone:** %s
**Status:** %s

You can track your parcel status anytime by logging into your GoPickup account.

Thank you for choosing GoPickup. If you have any questions, feel free to reply to this email or contact our support team.

Best regards,
The GoPickup Team
`, user.Name, parcel.PickupAddress, parcel.DropAddress, parcel.ReceiverName, parcel.ReceiverPhone, parcel.Status),
		)
	}
	c.JSON(http.StatusCreated, parcel)
}
// Handler: User requests password reset (sends OTP to email)
func ForgotAgentPasswordRequest(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}
	userCollection := config.GetCollection("agents")
	var user models.DeliveryPartner
	if err := userCollection.FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&user); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Delivery Partner  not found"})
		return
	}
	// Generate OTP
	otp := fmt.Sprintf("%06d", rand.Intn(1000000))
	otpDoc := models.ParcelOTP{
		ParcelID:  user.ID, // reuse ParcelOTP for user, or create UserOTP model
		OTP:       otp,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	_, err := services.OtpCollection().InsertOne(context.TODO(), otpDoc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store OTP"})
		return
	}
	// Send OTP via email
	go email.SendEmail(user.Email, "Your GoPickup Password Reset OTP", fmt.Sprintf("Your OTP for password reset is: %s. It expires in 10 minutes.", otp))
	c.JSON(http.StatusOK, gin.H{"message": "OTP sent to your email"})
}
func VerifyAgentPasswordResetOTP(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
		OTP   string `json:"otp" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and OTP required"})
		return
	}
	userCollection := config.GetCollection("agents")
	var user models.DeliveryPartner
	if err := userCollection.FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&user); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}
	// Find OTP
	var otpDoc models.ParcelOTP
	filter := bson.M{"parcel_id": user.ID, "otp": req.OTP}
	if err := services.OtpCollection().FindOne(context.TODO(), filter).Decode(&otpDoc); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid OTP"})
		return
	}
	if time.Now().After(otpDoc.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OTP expired"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "OTP verified. You can now reset your password."})
}

func ResetAgentPassword(c *gin.Context) {
	var req struct {
		Email       string `json:"email" binding:"required"`
		OTP         string `json:"otp" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email, OTP, and new password required"})
		return
	}
	userCollection := config.GetCollection("agents")
	var user models.User
	if err := userCollection.FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&user); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Delivery Partner  not found"})
		return
	}
	// Find OTP
	var otpDoc models.ParcelOTP
	filter := bson.M{"parcel_id": user.ID, "otp": req.OTP}
	if err := services.OtpCollection().FindOne(context.TODO(), filter).Decode(&otpDoc); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid OTP"})
		return
	}
	if time.Now().After(otpDoc.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OTP expired"})
		return
	}
	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword),bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	// Update password in DB
	_, err = userCollection.UpdateOne(context.Background(), bson.M{"_id": user.ID}, bson.M{"$set": bson.M{"password": string(hashedPassword)}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}
	// Send confirmation email
	go email.SendEmail(user.Email, "GoPickup Password Changed", "Your password has been changed successfully. You can now login with your new password.")
	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}
func ForgotPasswordRequest(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}
	userCollection := config.GetCollection("users")
	var user models.User
	if err := userCollection.FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&user); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	// Generate OTP
	otp := fmt.Sprintf("%06d", rand.Intn(1000000))
	otpDoc := models.ParcelOTP{
		ParcelID:  user.ID, // reuse ParcelOTP for user, or create UserOTP model
		OTP:       otp,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	_, err := services.OtpCollection().InsertOne(context.TODO(), otpDoc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store OTP"})
		return
	}
	// Send OTP via email
	go email.SendEmail(user.Email, "Your GoPickup Password Reset OTP", fmt.Sprintf("Your OTP for password reset is: %s. It expires in 10 minutes.", otp))
	c.JSON(http.StatusOK, gin.H{"message": "OTP sent to your email"})
}

// Handler: Verify OTP for password reset
func VerifyPasswordResetOTP(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
		OTP   string `json:"otp" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and OTP required"})
		return
	}
	userCollection := config.GetCollection("users")
	var user models.User
	if err := userCollection.FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&user); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	// Find OTP
	var otpDoc models.ParcelOTP
	filter := bson.M{"parcel_id": user.ID, "otp": req.OTP}
	if err := services.OtpCollection().FindOne(context.TODO(), filter).Decode(&otpDoc); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid OTP"})
		return
	}
	if time.Now().After(otpDoc.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OTP expired"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "OTP verified. You can now reset your password."})
}

// Handler: Reset password after OTP verification
func ResetPassword(c *gin.Context) {
	var req struct {
		Email       string `json:"email" binding:"required"`
		OTP         string `json:"otp" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email, OTP, and new password required"})
		return
	}
	userCollection := config.GetCollection("users")
	var user models.User
	if err := userCollection.FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&user); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	// Find OTP
	var otpDoc models.ParcelOTP
	filter := bson.M{"parcel_id": user.ID, "otp": req.OTP}
	if err := services.OtpCollection().FindOne(context.TODO(), filter).Decode(&otpDoc); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid OTP"})
		return
	}
	if time.Now().After(otpDoc.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OTP expired"})
		return
	}
	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword),bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	// Update password in DB
	_, err = userCollection.UpdateOne(context.Background(), bson.M{"_id": user.ID}, bson.M{"$set": bson.M{"password": string(hashedPassword)}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}
	// Send confirmation email
	go email.SendEmail(user.Email, "GoPickup Password Changed", "Your password has been changed successfully. You can now login with your new password.")
	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}
func Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if err := services.Register(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// sending email via the sendgrid
	go email.SendEmail(
		user.Email,
		"Welcome to GoPickup!",
		"Congratulations! You have successfully registered as a user on GoPickup. You can now log in and start using our services.",
	)

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

func Login(c *gin.Context) {
	var credentials models.User
	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	user, err := services.Login(credentials.Email, credentials.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	token, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func UpdateDriverLocation(c *gin.Context) {
    var loc models.LocationUpdate
    if err := c.ShouldBindJSON(&loc); err != nil {
        c.JSON(400, gin.H{"error": "invalid payload"})
        return
    }

    producer, err := kafka.NewProducer(kafka.ProducerConfig{
        Brokers:      []string{"localhost:9092"},
        Topic:        "location-updates",
        BatchSize:    10,
        BatchTimeout: 100 * time.Millisecond,
    })
    if err != nil {
        c.JSON(500, gin.H{"error": "producer init failed"})
        return
    }
    defer producer.Close()

    if err := producer.SendLocationUpdate(c, loc); err != nil {
        c.JSON(500, gin.H{"error": "send failed"})
        return
    }

    c.JSON(200, gin.H{"message": "location sent"})
}
func GetDriverLocation(c *gin.Context) {
    agentID := c.Param("agentID")
    parcelID := c.Param("parcelID")

    rdb := redis.NewClient(&redis.Options{Addr: "redis:6379"})
    key := fmt.Sprintf("location:%s:%s", agentID, parcelID)
    val, err := rdb.Get(context.Background(), key).Result()
    if err != nil {
        c.JSON(404, gin.H{"error": "location not found"})
        return
    }

    c.Data(200, "application/json", []byte(val))
}
