//go:build mysql || full || mini
// +build mysql full mini

package build

import (
	_ "github.com/haha4github/trojan-go/statistic/redis"
	_ "github.com/p4gefau1t/trojan-go/statistic/mysql"
)
