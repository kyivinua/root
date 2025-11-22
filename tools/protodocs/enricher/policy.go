package enricher

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// StaticPolicyEngine implements PolicyEngine with static configuration
type StaticPolicyEngine struct {
	mu       sync.RWMutex
	policies map[string]*TenantPolicy
	defaultPolicy *TenantPolicy
}

// NewStaticPolicyEngine creates a new static policy engine
func NewStaticPolicyEngine(configPath string, defaultTenant string) (*StaticPolicyEngine, error) {
	engine := &StaticPolicyEngine{
		policies: make(map[string]*TenantPolicy),
	}

	// Load policies from config file if provided
	if configPath != "" {
		if err := engine.LoadPoliciesFromFile(configPath); err != nil {
			return nil, fmt.Errorf("load policies: %w", err)
		}
	}

	// Set default policy if not loaded
	if engine.defaultPolicy == nil {
		engine.defaultPolicy = &TenantPolicy{
			TenantID:            defaultTenant,
			AllowEnrichment:     true,
			RequireApproval:     false,
			MaxTokens:           10000,
			AllowedVisibilities: []string{"PUBLIC", "PARTNER", "INTERNAL"},
			CustomRules:         make(map[string]interface{}),
		}
	}

	return engine, nil
}

// Evaluate evaluates the policy for a target and tenant
func (pe *StaticPolicyEngine) Evaluate(ctx context.Context, target EnrichmentTarget, tenant string) (*PolicyDecision, error) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	// Get tenant policy
	policy, ok := pe.policies[tenant]
	if !ok {
		policy = pe.defaultPolicy
	}

	decision := &PolicyDecision{
		Allowed:         policy.AllowEnrichment,
		RequiresApproval: policy.RequireApproval,
		Tenant:          tenant,
	}

	if !policy.AllowEnrichment {
		decision.Reason = fmt.Sprintf("enrichment disabled for tenant %s", tenant)
		return decision, nil
	}

	// Check visibility filters
	targetCtx := target.GetContext()
	if visibility, ok := targetCtx["visibility"].(string); ok {
		allowed := false
		for _, v := range policy.AllowedVisibilities {
			if v == visibility {
				allowed = true
				break
			}
		}
		if !allowed {
			decision.Allowed = false
			decision.Reason = fmt.Sprintf("visibility %s not allowed for tenant %s", visibility, tenant)
			return decision, nil
		}
	}

	// Check custom rules
	if customCheck, ok := policy.CustomRules["require_approval_for_public"].(bool); ok && customCheck {
		if visibility, ok := targetCtx["visibility"].(string); ok && visibility == "PUBLIC" {
			decision.RequiresApproval = true
			decision.Reason = "public visibility requires approval"
			return decision, nil
		}
	}

	decision.Reason = "policy check passed"
	return decision, nil
}

// GetPolicyForTenant returns the policy for a specific tenant
func (pe *StaticPolicyEngine) GetPolicyForTenant(tenant string) (*TenantPolicy, error) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	policy, ok := pe.policies[tenant]
	if !ok {
		return pe.defaultPolicy, nil
	}

	return policy, nil
}

// LoadPoliciesFromFile loads tenant policies from a JSON file
func (pe *StaticPolicyEngine) LoadPoliciesFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read policies file: %w", err)
	}

	var config struct {
		Default *TenantPolicy            `json:"default"`
		Tenants map[string]*TenantPolicy `json:"tenants"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("unmarshal policies: %w", err)
	}

	pe.mu.Lock()
	defer pe.mu.Unlock()

	if config.Default != nil {
		pe.defaultPolicy = config.Default
	}

	for tenantID, policy := range config.Tenants {
		policy.TenantID = tenantID
		pe.policies[tenantID] = policy
	}

	return nil
}

// AddPolicy adds or updates a policy for a tenant
func (pe *StaticPolicyEngine) AddPolicy(tenant string, policy *TenantPolicy) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	policy.TenantID = tenant
	pe.policies[tenant] = policy
}

// RemovePolicy removes a policy for a tenant
func (pe *StaticPolicyEngine) RemovePolicy(tenant string) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	delete(pe.policies, tenant)
}

// NoOpPolicyEngine is a policy engine that allows all enrichments
type NoOpPolicyEngine struct{}

// NewNoOpPolicyEngine creates a no-op policy engine
func NewNoOpPolicyEngine() *NoOpPolicyEngine {
	return &NoOpPolicyEngine{}
}

// Evaluate always allows enrichment
func (npe *NoOpPolicyEngine) Evaluate(ctx context.Context, target EnrichmentTarget, tenant string) (*PolicyDecision, error) {
	return &PolicyDecision{
		Allowed:         true,
		RequiresApproval: false,
		Reason:          "no-op policy allows all",
		Tenant:          tenant,
	}, nil
}

// GetPolicyForTenant returns a default policy
func (npe *NoOpPolicyEngine) GetPolicyForTenant(tenant string) (*TenantPolicy, error) {
	return &TenantPolicy{
		TenantID:            tenant,
		AllowEnrichment:     true,
		RequireApproval:     false,
		MaxTokens:           100000,
		AllowedVisibilities: []string{"PUBLIC", "PARTNER", "INTERNAL", "PRIVATE"},
		CustomRules:         make(map[string]interface{}),
	}, nil
}

// ApprovalWorkflowEngine implements approval workflow for enrichments
type ApprovalWorkflowEngine struct {
	basePolicyEngine PolicyEngine
	mu               sync.RWMutex
	pendingApprovals map[string]*ApprovalRequest
}

// ApprovalRequest represents a pending approval
type ApprovalRequest struct {
	RequestID   string
	TargetID    string
	Tenant      string
	RequestedBy string
	RequestedAt time.Time
	ApprovedBy  string
	ApprovedAt  *time.Time
	Rejected    bool
	RejectedBy  string
	RejectedAt  *time.Time
	Reason      string
}

// NewApprovalWorkflowEngine creates a new approval workflow engine
func NewApprovalWorkflowEngine(basePolicyEngine PolicyEngine) *ApprovalWorkflowEngine {
	return &ApprovalWorkflowEngine{
		basePolicyEngine: basePolicyEngine,
		pendingApprovals: make(map[string]*ApprovalRequest),
	}
}

// Evaluate evaluates policy and manages approval workflow
func (awe *ApprovalWorkflowEngine) Evaluate(ctx context.Context, target EnrichmentTarget, tenant string) (*PolicyDecision, error) {
	// First, check base policy
	decision, err := awe.basePolicyEngine.Evaluate(ctx, target, tenant)
	if err != nil {
		return nil, err
	}

	if !decision.Allowed {
		return decision, nil
	}

	// Check if there's a pending or approved request
	awe.mu.RLock()
	approval, exists := awe.pendingApprovals[target.GetIdentifier()]
	awe.mu.RUnlock()

	if exists {
		if approval.ApprovedAt != nil {
			decision.RequiresApproval = false
			decision.ApprovedBy = approval.ApprovedBy
			decision.ApprovedAt = approval.ApprovedAt
			return decision, nil
		}
		if approval.Rejected {
			decision.Allowed = false
			decision.Reason = fmt.Sprintf("approval rejected by %s: %s", approval.RejectedBy, approval.Reason)
			return decision, nil
		}
	}

	return decision, nil
}

// RequestApproval creates a new approval request
func (awe *ApprovalWorkflowEngine) RequestApproval(targetID, tenant, requestedBy string) (*ApprovalRequest, error) {
	awe.mu.Lock()
	defer awe.mu.Unlock()

	request := &ApprovalRequest{
		RequestID:   fmt.Sprintf("approval-%d", time.Now().UnixNano()),
		TargetID:    targetID,
		Tenant:      tenant,
		RequestedBy: requestedBy,
		RequestedAt: time.Now(),
	}

	awe.pendingApprovals[targetID] = request
	return request, nil
}

// ApproveRequest approves a pending request
func (awe *ApprovalWorkflowEngine) ApproveRequest(targetID, approvedBy string) error {
	awe.mu.Lock()
	defer awe.mu.Unlock()

	request, exists := awe.pendingApprovals[targetID]
	if !exists {
		return fmt.Errorf("no pending approval for target %s", targetID)
	}

	now := time.Now()
	request.ApprovedBy = approvedBy
	request.ApprovedAt = &now
	return nil
}

// RejectRequest rejects a pending request
func (awe *ApprovalWorkflowEngine) RejectRequest(targetID, rejectedBy, reason string) error {
	awe.mu.Lock()
	defer awe.mu.Unlock()

	request, exists := awe.pendingApprovals[targetID]
	if !exists {
		return fmt.Errorf("no pending approval for target %s", targetID)
	}

	now := time.Now()
	request.Rejected = true
	request.RejectedBy = rejectedBy
	request.RejectedAt = &now
	request.Reason = reason
	return nil
}

// GetPolicyForTenant delegates to base policy engine
func (awe *ApprovalWorkflowEngine) GetPolicyForTenant(tenant string) (*TenantPolicy, error) {
	return awe.basePolicyEngine.GetPolicyForTenant(tenant)
}
