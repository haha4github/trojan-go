//go:build mysql || full || mini
// +build mysql full mini

package build

import (
	_ "github.com/haha4github/trojan-go/statistic/mysql"
	_ "github.com/haha4github/trojan-go/statistic/redis"
)
