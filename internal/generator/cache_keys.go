package generator

import (
	"fmt"
	"os"
	"strings"
)

// generateCacheKeysForEntity generates Redis cache key constants for a specific entity
func (g *Generator) generateCacheKeysForEntity(schema, entity string) error {
	fmt.Printf("🔑 Generating cache keys for %s.%s\n", schema, entity)

	cacheFilePath := "common/cache/redis.go"

	// Check if cache file exists
	if _, err := os.Stat(cacheFilePath); os.IsNotExist(err) {
		return fmt.Errorf("cache file not found: %s", cacheFilePath)
	}

	// Read existing cache file
	content, err := os.ReadFile(cacheFilePath)
	if err != nil {
		return fmt.Errorf("read cache file error: %v", err)
	}

	contentStr := string(content)

	// Check if cache keys already exist for this entity
	if g.cacheKeysExist(contentStr, schema, entity) {
		fmt.Printf("✅ Cache keys already exist for %s.%s\n", schema, entity)
		return nil
	}

	// Generate new cache key constants
	newCacheKeys := g.generateCacheKeyConstants(schema, entity)

	// Find insertion point and update content
	updatedContent := g.insertCacheKeys(contentStr, newCacheKeys)

	// Write updated content
	if err := os.WriteFile(cacheFilePath, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("write cache file error: %v", err)
	}

	fmt.Printf("✅ Updated cache keys for %s.%s in: %s\n", schema, entity, cacheFilePath)
	return nil
}

// cacheKeysExist checks if cache keys already exist for this entity
func (g *Generator) cacheKeysExist(content, schema, entity string) bool {
	displayName := toPascalCase(entity)
	searchPattern := fmt.Sprintf("%sFindByID", displayName)
	return strings.Contains(content, searchPattern)
}

// generateCacheKeyConstants generates the cache key constant definitions for an entity
func (g *Generator) generateCacheKeyConstants(schema, entity string) string {
	var result strings.Builder

	displayName := toPascalCase(entity)

	// Generate cache keys for common repository patterns
	cacheKeys := []struct {
		suffix string
		desc   string
	}{
		{"FindByID", fmt.Sprintf("find %s by id", entity)},
		{"FindByName", fmt.Sprintf("find %s by name", entity)},
	}

	for _, key := range cacheKeys {
		constName := fmt.Sprintf("%s%s", displayName, key.suffix)
		redisKey := g.generateRedisKey(schema, entity, strings.ToLower(key.suffix))

		result.WriteString(fmt.Sprintf("\t// %s is a redis key for %s.\n", constName, key.desc))
		result.WriteString(fmt.Sprintf("\t%s = prefix + \":%s:%s\"\n", constName, schema, redisKey))
	}

	return result.String()
}

// generateRedisKey generates the Redis key pattern
func (g *Generator) generateRedisKey(schema, entity, action string) string {
	// Convert action to kebab-case if needed
	action = strings.ReplaceAll(action, "by", "-by-")
	return fmt.Sprintf("%s:%s:%%v", entity, action)
}

// insertCacheKeys inserts new cache keys before the last closing parenthesis
func (g *Generator) insertCacheKeys(content, newCacheKeys string) string {
	// Find the last closing parenthesis and newline before the final closing brace
	lastParen := strings.LastIndex(content, ")")
	if lastParen == -1 {
		return content
	}

	// Insert new cache keys before the last closing parenthesis
	before := content[:lastParen]
	after := content[lastParen:]

	return before + newCacheKeys + after
}
