package main

type generationProfile string

const profileLocal generationProfile = "local"

type generationOptions struct {
	profile generationProfile
}

func localGenerationOptions() generationOptions {
	return generationOptions{profile: profileLocal}
}

func (options generationOptions) digest() string {
	return ""
}
