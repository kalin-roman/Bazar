package user

type User struct {
	ID              int64 // For future instead of the ID I should change it on the UUID from the Supabase
	FullName        string
	Email           string
	Age             int64
	AddressDelivery string
	Avatar          string
}
