#!/bin/bash

# Script to add V2 context fields to all request types
# This is a helper to speed up the migration process

echo "V2 Migration Helper Script"
echo "=========================="
echo ""
echo "This script helps identify request types that need V2 fields added."
echo ""

# Find all type files
echo "Scanning for type files..."
find internal/llm -name "types_*.go" -type f

echo ""
echo "To complete migration, add these fields to each Request struct:"
echo ""
echo "// V2 fields (optional)"
echo "AgentID       string               \`json:\"agentId,omitempty\"\`"
echo "SessionID     string               \`json:\"sessionId,omitempty\"\`"
echo "ComponentType string               \`json:\"componentType,omitempty\"\`"
echo "ComponentID   string               \`json:\"componentId,omitempty\"\`"
echo "Context       *types.V2Context     \`json:\"context,omitempty\"\`"
echo "Messages      []types.V2Message    \`json:\"messages,omitempty\"\`"
echo ""
echo "And add these methods:"
echo ""
echo "func (r TypeName) GetV2Context() *types.V2Context { return r.Context }"
echo "func (r TypeName) GetUserID() string { return r.UserID }"
echo "func (r TypeName) GetAgentID() string { return r.AgentID }"
echo ""
