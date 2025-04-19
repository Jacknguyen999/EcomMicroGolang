package utils

import (
	"strings"
)

// ProductCategory represents a product category
type ProductCategory string

const (
	CategoryGaming      ProductCategory = "Gaming"
	CategoryOffice      ProductCategory = "Office"
	CategoryFitness     ProductCategory = "Fitness"
	CategoryElectronics ProductCategory = "Electronics"
	CategoryFashion     ProductCategory = "Fashion"
	CategoryHome        ProductCategory = "Home"
	CategoryOther       ProductCategory = "Other"
)

// CategoryKeywords maps categories to their keywords
var CategoryKeywords = map[ProductCategory][]string{
	CategoryGaming: {
		"gaming", "game", "player", "console", "controller", "keyboard", "mouse", "headset",
		"joystick", "playstation", "xbox", "nintendo", "steam", "gamer",
	},
	CategoryOffice: {
		"office", "work", "desk", "chair", "ergonomic", "professional", "business",
		"document", "printer", "scanner", "stationary", "pen", "notebook",
	},
	CategoryFitness: {
		"fitness", "workout", "exercise", "health", "track", "sport", "gym", "yoga",
		"running", "weight", "training", "muscle", "cardio", "athletic",
	},
	CategoryElectronics: {
		"laptop", "phone", "computer", "tablet", "watch", "electronic", "device",
		"smartphone", "gadget", "tech", "technology", "digital", "smart", "wireless",
	},
	CategoryFashion: {
		"fashion", "clothing", "apparel", "wear", "dress", "shirt", "pants", "shoes",
		"accessory", "jewelry", "watch", "bag", "handbag", "style",
	},
	CategoryHome: {
		"home", "house", "kitchen", "bathroom", "bedroom", "living", "furniture",
		"decor", "decoration", "appliance", "garden", "indoor", "outdoor",
	},
}

// GetProductCategories determines the categories a product belongs to based on its name and description
func GetProductCategories(name, description string) []ProductCategory {
	name = strings.ToLower(name)
	description = strings.ToLower(description)
	
	categories := make([]ProductCategory, 0)
	
	for category, keywords := range CategoryKeywords {
		for _, keyword := range keywords {
			if strings.Contains(name, keyword) || strings.Contains(description, keyword) {
				categories = append(categories, category)
				break
			}
		}
	}
	
	// If no categories were found, assign to "Other"
	if len(categories) == 0 {
		categories = append(categories, CategoryOther)
	}
	
	return categories
}

// GetRelatedCategories returns categories that are related to the given categories
func GetRelatedCategories(categories []ProductCategory) []ProductCategory {
	// Define category relationships (which categories are related to each other)
	categoryRelationships := map[ProductCategory][]ProductCategory{
		CategoryGaming:      {CategoryElectronics},
		CategoryElectronics: {CategoryGaming, CategoryOffice},
		CategoryOffice:      {CategoryElectronics, CategoryHome},
		CategoryFitness:     {CategoryFashion},
		CategoryFashion:     {CategoryFitness},
		CategoryHome:        {CategoryOffice},
		CategoryOther:       {},
	}
	
	relatedCats := make(map[ProductCategory]bool)
	
	// Add all input categories
	for _, cat := range categories {
		relatedCats[cat] = true
	}
	
	// Add related categories
	for _, cat := range categories {
		for _, relatedCat := range categoryRelationships[cat] {
			relatedCats[relatedCat] = true
		}
	}
	
	// Convert map to slice
	result := make([]ProductCategory, 0, len(relatedCats))
	for cat := range relatedCats {
		result = append(result, cat)
	}
	
	return result
}

// ProductMatchesCategories checks if a product matches any of the given categories
func ProductMatchesCategories(name, description string, categories []ProductCategory) bool {
	productCategories := GetProductCategories(name, description)
	
	for _, productCat := range productCategories {
		for _, targetCat := range categories {
			if productCat == targetCat {
				return true
			}
		}
	}
	
	return false
}
