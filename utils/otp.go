package utils

import (
	"context"
	"fmt"
	"go_pickup/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func VerifyOtp(parcelID, otp string) (bool, error) {
    otpCollection := config.GetCollection("otp")
    objID, err:= primitive.ObjectIDFromHex(parcelID)
    fmt.Println(err)

    var result struct {
        OTP string `bson:"otp"`
    }
    
    err = otpCollection.FindOne(context.TODO(), bson.M{
        "parcel_id": objID, // ✅ Correct type
        "otp":       otp,
    }).Decode(&result)
    if err != nil {
        return false, err
    }
    return true, nil
}