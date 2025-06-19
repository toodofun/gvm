package common

import (
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
	"sort"
)

const (
	latestVersion = "latest"
)

func MatchVersion(v string, versions []*version.Version) (*version.Version, error) {
	// 从小到大排序
	sort.Sort(version.Collection(versions))

	// 如果是 "latest"，则返回最新的正式版本（非预发布版本）
	if v == latestVersion {
		for i := len(versions) - 1; i >= 0; i-- {
			if versions[i].Prerelease() == "" {
				return versions[i], nil
			}
		}
		// 如果没有正式版本，返回最新版本
		return versions[len(versions)-1], nil
	}

	// 尝试解析版本字符串
	ver, err := version.NewVersion(v)
	if err != nil {
		logrus.Debugf("Failed to parse version %s: %s", v, err)
		return nil, fmt.Errorf("invalid version format: %s", v)
	}

	// 统计原始输入中点的数量来判断是精确匹配还是模糊匹配
	dotCount := 0
	for _, ch := range v {
		if ch == '.' {
			dotCount++
		}
	}

	// 如果原始输入的点数少于2个，进行模糊匹配
	if dotCount < 2 {
		var matched []*version.Version
		var exactMatch *version.Version
		inputSegments := ver.Segments()

		for _, vItem := range versions {
			segments := vItem.Segments()

			// 对于主版本匹配（如输入"1"），检查是否有精确匹配
			if dotCount == 0 && vItem.Equal(ver) {
				exactMatch = vItem
			}

			// 主版本匹配 (输入如 "1")
			if dotCount == 0 && len(segments) >= 1 && len(inputSegments) >= 1 &&
				segments[0] == inputSegments[0] {
				matched = append(matched, vItem)
			}
			// 主+次版本匹配 (输入如 "1.2")
			if dotCount == 1 && len(segments) >= 2 && len(inputSegments) >= 2 &&
				segments[0] == inputSegments[0] && segments[1] == inputSegments[1] {
				matched = append(matched, vItem)
			}
		}

		// 如果是主版本匹配且有精确匹配，优先返回精确匹配
		if dotCount == 0 && exactMatch != nil {
			return exactMatch, nil
		}

		if len(matched) > 0 {
			// 返回匹配中的最大正式版本，如果没有正式版本则返回最大版本
			sort.Sort(version.Collection(matched))
			for i := len(matched) - 1; i >= 0; i-- {
				if matched[i].Prerelease() == "" {
					return matched[i], nil
				}
			}
			// 如果没有正式版本，返回最新版本
			return matched[len(matched)-1], nil
		}
	} else {
		// 精准匹配 (输入如 "1.2.3")
		for _, vItem := range versions {
			if vItem.Equal(ver) {
				return vItem, nil
			}
		}
	}

	return nil, fmt.Errorf("version %s not found", v)
}
