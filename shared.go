package app

type (
	Product struct {
		Id          int
		Name        string
		Description string
		Image       string
		Price       float32
	}
)

var Products = []Product{
	{
		Id:          0,
		Name:        "iPhone 12 Pro",
		Description: "Test",
		Image:       "https://images.unsplash.com/photo-1603921326210-6edd2d60ca68",
		Price:       999,
	},
	{
		Id:          1,
		Name:        "iPhone 12",
		Description: "Test",
		Image:       "https://images.unsplash.com/photo-1611472173362-3f53dbd65d80",
		Price:       699,
	},
	{
		Id:          2,
		Name:        "iPhone SE",
		Description: "399",
		Image:       "https://images.unsplash.com/photo-1529618160092-2f8ccc8e087b",
		Price:       399,
	},
	{
		Id:          3,
		Name:        "iPhone 11",
		Description: "599",
		Image:       "https://images.unsplash.com/photo-1574755393849-623942496936",
		Price:       599,
	},
}
