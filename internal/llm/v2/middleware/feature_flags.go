package middleware

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
)

// FeatureFlags manages V2 feature rollout
type FeatureFlags struct {
	mu                  sync.RWMutex
	useV2Endpoints      bool
	v2RolloutPercentage float64
	enabledAgents       map[string]bool
	userOverrides       map[string]bool // userId -> useV2
}

var globalFlags *FeatureFlags
var once sync.Once

// InitFeatureFlags initializes the global feature flags
func InitFeatureFlags() *FeatureFlags {
	once.Do(func() {
		globalFlags = &FeatureFlags{
			useV2Endpoints:      false,
			v2RolloutPercentage: 0.0,
			enabledAgents:       make(map[string]bool),
			userOverrides:       make(map[string]bool),
		}
		log.Println("[FeatureFlags] Initialized with V2 disabled")
	})
	return globalFlags
}

// GetFeatureFlags returns the global feature flags instance
func GetFeatureFlags() *FeatureFlags {
	if globalFlags == nil {
		return InitFeatureFlags()
	}
	return globalFlags
}

// SetV2Enabled enables or disables V2 endpoints globally
func (ff *FeatureFlags) SetV2Enabled(enabled bool) {
	ff.mu.Lock()
	defer ff.mu.Unlock()
	ff.useV2Endpoints = enabled
	log.Printf("[FeatureFlags] V2 endpoints enabled: %v", enabled)
}

// SetV2RolloutPercentage sets the percentage of users to use V2
func (ff *FeatureFlags) SetV2RolloutPercentage(percentage float64) {
	ff.mu.Lock()
	defer ff.mu.Unlock()
	if percentage < 0 {
		percentage = 0
	}
	if percentage > 100 {
		percentage = 100
	}
	ff.v2RolloutPercentage = percentage
	log.Printf("[FeatureFlags] V2 rollout percentage: %.1f%%", percentage)
}

// EnableAgentV2 enables V2 for a specific agent
func (ff *FeatureFlags) EnableAgentV2(agentId string) {
	ff.mu.Lock()
	defer ff.mu.Unlock()
	ff.enabledAgents[agentId] = true
	log.Printf("[FeatureFlags] Enabled V2 for agent: %s", agentId)
}

// DisableAgentV2 disables V2 for a specific agent
func (ff *FeatureFlags) DisableAgentV2(agentId string) {
	ff.mu.Lock()
	defer ff.mu.Unlock()
	ff.enabledAgents[agentId] = false
	log.Printf("[FeatureFlags] Disabled V2 for agent: %s", agentId)
}

// SetUserOverride forces a specific user to use V2 or V1
func (ff *FeatureFlags) SetUserOverride(userId string, useV2 bool) {
	ff.mu.Lock()
	defer ff.mu.Unlock()
	ff.userOverrides[userId] = useV2
	log.Printf("[FeatureFlags] User override for %s: V2=%v", userId, useV2)
}

// RemoveUserOverride removes a user override
func (ff *FeatureFlags) RemoveUserOverride(userId string) {
	ff.mu.Lock()
	defer ff.mu.Unlock()
	delete(ff.userOverrides, userId)
	log.Printf("[FeatureFlags] Removed user override for %s", userId)
}

// ShouldUseV2 determines if a request should use V2 endpoints
func (ff *FeatureFlags) ShouldUseV2(userId string, agentId string) bool {
	ff.mu.RLock()
	defer ff.mu.RUnlock()

	// Check user override first
	if override, exists := ff.userOverrides[userId]; exists {
		log.Printf("[FeatureFlags] User override for %s: V2=%v", userId, override)
		return override
	}

	// If V2 is globally disabled, return false
	if !ff.useV2Endpoints {
		return false
	}

	// Check if agent is enabled for V2
	if agentEnabled, exists := ff.enabledAgents[agentId]; exists && !agentEnabled {
		log.Printf("[FeatureFlags] Agent %s not enabled for V2", agentId)
		return false
	}

	// Use rollout percentage for gradual rollout
	if ff.v2RolloutPercentage < 100 {
		userHash := hashUserId(userId)
		shouldUse := userHash < ff.v2RolloutPercentage
		log.Printf("[FeatureFlags] User %s rollout check: %.1f%% (hash: %.1f) -> V2=%v",
			userId, ff.v2RolloutPercentage, userHash, shouldUse)
		return shouldUse
	}

	// Full rollout
	return true
}

// hashUserId creates a consistent hash for a user ID (0-100)
func hashUserId(userId string) float64 {
	hasher := md5.New()
	hasher.Write([]byte(userId))
	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)

	// Use first 8 characters to generate a number between 0-100
	var hashValue uint64
	for i := 0; i < 8 && i < len(hashString); i++ {
		hashValue = hashValue*16 + uint64(hashString[i])
	}

	return float64(hashValue%10000) / 100.0
}

// GetStats returns current feature flag statistics
func (ff *FeatureFlags) GetStats() map[string]interface{} {
	ff.mu.RLock()
	defer ff.mu.RUnlock()

	enabledAgentsList := []string{}
	for agent, enabled := range ff.enabledAgents {
		if enabled {
			enabledAgentsList = append(enabledAgentsList, agent)
		}
	}

	return map[string]interface{}{
		"v2Enabled":          ff.useV2Endpoints,
		"rolloutPercentage":  ff.v2RolloutPercentage,
		"enabledAgents":      enabledAgentsList,
		"userOverridesCount": len(ff.userOverrides),
	}
}

// QuickRollout provides preset rollout configurations
func (ff *FeatureFlags) QuickRollout(stage string) {
	ff.mu.Lock()
	defer ff.mu.Unlock()

	switch stage {
	case "testing":
		ff.useV2Endpoints = true
		ff.v2RolloutPercentage = 1.0
		log.Println("[FeatureFlags] Quick rollout: TESTING (1%)")
	case "pilot":
		ff.useV2Endpoints = true
		ff.v2RolloutPercentage = 10.0
		log.Println("[FeatureFlags] Quick rollout: PILOT (10%)")
	case "beta":
		ff.useV2Endpoints = true
		ff.v2RolloutPercentage = 50.0
		log.Println("[FeatureFlags] Quick rollout: BETA (50%)")
	case "full":
		ff.useV2Endpoints = true
		ff.v2RolloutPercentage = 100.0
		log.Println("[FeatureFlags] Quick rollout: FULL (100%)")
	case "off":
		ff.useV2Endpoints = false
		ff.v2RolloutPercentage = 0.0
		log.Println("[FeatureFlags] Quick rollout: OFF")
	default:
		log.Printf("[FeatureFlags] Unknown rollout stage: %s", stage)
	}
}

// SimulateRollout simulates rollout for testing
func (ff *FeatureFlags) SimulateRollout(numUsers int) map[string]int {
	ff.mu.RLock()
	percentage := ff.v2RolloutPercentage
	ff.mu.RUnlock()

	v1Count := 0
	v2Count := 0

	for i := 0; i < numUsers; i++ {
		userId := generateTestUserId(i)
		if ff.ShouldUseV2(userId, "test") {
			v2Count++
		} else {
			v1Count++
		}
	}

	return map[string]int{
		"v1Users":          v1Count,
		"v2Users":          v2Count,
		"totalUsers":       numUsers,
		"v2Percentage":     int((float64(v2Count) / float64(numUsers)) * 100),
		"targetPercentage": int(percentage),
	}
}

// generateTestUserId generates a test user ID
func generateTestUserId(index int) string {
	return fmt.Sprintf("test-user-%d", index)
}
