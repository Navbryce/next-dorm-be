package util

import (
	"fmt"
	"github.com/navbryce/next-dorm-be/config"
)

func Avatar(seed string) string {
	return fmt.Sprintf("https://avatars.dicebear.com/api/bottts/%v.svg?size=%v", seed, config.AVATAR_SIZE)
}
