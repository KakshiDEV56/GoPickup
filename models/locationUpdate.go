package models
// Struct embedding allows you to include one struct inside another,
// inheriting its fields. For example:

// type Agent struct {
// 	ID   string `json:"id"`
// 	Name string `json:"name"`
// 	// other agent fields...
// }

// // LocationUpdate embeds Agent, so it has all Agent fields plus its own.
// type LocationUpdate struct {
// 	Agent
// 	ParcelID   string    `json:"parcel_id"`
// 	Latitude   float64   `json:"latitude"`
// 	Longitude  float64   `json:"longitude"`
// 	Timestamp  time.Time `json:"timestamp"`
// 	Speed      float64   `json:"speed,omitempty"`
// 	Status     string    `json:"status,omitempty"`
// }

// Now, LocationUpdate has fields: ID, Name, ParcelID, Latitude, etc.
//type LocationUpdate struct {
//	AgentID    string    `json:"agent_id"`
//	ParcelID   string    `json:"parcel_id"`
//	Latitude   float64   `json:"latitude"`
//	Longitude  float64   `json:"longitude"`
//	Timestamp  time.Time `json:"timestamp"`
//	Speed      float64   `json:"speed,omitempty"`
//	Status     string    `json:"status,omitempty"`
//}