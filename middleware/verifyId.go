package middleware

import (
	"net/http"

	"go_pickup/config" // replace with your actual config package path

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// VerifyClientId checks if the given agent ID and parcel ID exist in the DB.
func VerifyClientId(c *gin.Context) {
	// request payload
	type req struct {
		ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
		ParcelID  primitive.ObjectID `bson:"parcel_id" json:"parcel_id"`
	}

	var request req
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request payload",
			"error":   err.Error(),
		})
		return
	}

	// get collections
	agentsCol := config.GetCollection("agents")
	parcelsCol := config.GetCollection("parcels")

	// check if agent ID exists
	var agent bson.M
	err := agentsCol.FindOne(c, bson.M{"id": request.ID}).Decode(&agent)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Agent ID not found",
		})
		return
	}

	// check if parcel ID exists
	var parcel bson.M
	err = parcelsCol.FindOne(c, bson.M{"parcel_id": request.ParcelID}).Decode(&parcel)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Parcel ID not found",
		})
		return
	}

	// both found, proceed to next handler
	c.Set("clientID", request.ID)
	c.Set("parcelID", request.ParcelID)
	c.Next()
}

