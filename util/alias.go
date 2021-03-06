package util

import (
	"fmt"
	"github.com/navbryce/next-dorm-be/model"
	"math/rand"
)

var names = []string{
	"Dog",
	"Cat",
	"Frog",
	"Wreck",
	"Phone",
	"Dragon",
	"Number 3",
	"Piano",
	"Mask",
	"Aardvark",
	"Albatross",
	"Gazelle",
	"Sand",
}

func GenerateAnonymousUser() *model.AnonymousUser {
	return BuildAnonymousUserFromDisplayName(fmt.Sprintf("Anon %v", names[rand.Intn(len(names))]))
}

func BuildAnonymousUserFromDisplayName(displayName string) *model.AnonymousUser {
	return &model.AnonymousUser{
		DisplayName: displayName,
		AvatarUrl:   Avatar(displayName),
	}
}
