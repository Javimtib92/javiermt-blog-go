package models

type Article struct {
	Slug      string
	Category  string
	Data FrontMatter
}

type FrontMatter struct {
	Thumbnail string
	Title string
}