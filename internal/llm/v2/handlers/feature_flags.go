package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"agent-core/internal/llm/utils"
	"agent-core/internal/llm/v2/middleware"
)

// FeatureFlagsHandler - POST /v2/feature-flags
// Manages V2 feature flag settings
func FeatureFlagsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[FeatureFlags] Request received: %s", r.Method)

	if r.Method == http.MethodGet {
		// Get current settings
		featureFlags := middleware.GetFeatureFlags()
		stats := featureFlags.GetStats()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    stats,
		})
		return
	}

	if r.Method == http.MethodPost {
		// Update settings
		var req struct {
			V2Enabled         *bool    `json:"v2Enabled"`
			RolloutPercentage *float64 `json:"rolloutPercentage"`
			EnabledAgents     []string `json:"enabledAgents"`
			QuickRolloutStage string   `json:"quickRolloutStage"` // "testing", "pilot", "beta", "full", "off"
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[FeatureFlags] Failed to decode request: %v", err)
			utils.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		featureFlags := middleware.GetFeatureFlags()

		// Handle quick rollout stages
		if req.QuickRolloutStage != "" {
			featureFlags.QuickRollout(req.QuickRolloutStage)
		} else {
			// Manual configuration
			if req.V2Enabled != nil {
				featureFlags.SetV2Enabled(*req.V2Enabled)
			}
			if req.RolloutPercentage != nil {
				featureFlags.SetV2RolloutPercentage(*req.RolloutPercentage)
			}
			if req.EnabledAgents != nil {
				for _, agent := range req.EnabledAgents {
					featureFlags.EnableAgentV2(agent)
				}
			}
		}

		stats := featureFlags.GetStats()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Feature flags updated",
			"data":    stats,
		})
		return
	}

	utils.WriteJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// FeatureFlagsStatsHandler - GET /v2/feature-flags/stats
// Returns current feature flag statistics
func FeatureFlagsStatsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[FeatureFlagsStats] Request received")

	featureFlags := middleware.GetFeatureFlags()
	stats := featureFlags.GetStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// FeatureFlagsSimulateHandler - POST /v2/feature-flags/simulate
// Simulates rollout distribution for testing
func FeatureFlagsSimulateHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("[FeatureFlagsSimulate] Request received")

	var req struct {
		NumUsers int `json:"numUsers"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.NumUsers == 0 {
		req.NumUsers = 1000 // Default
	}

	featureFlags := middleware.GetFeatureFlags()
	simulation := featureFlags.SimulateRollout(req.NumUsers)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    simulation,
	})
}
