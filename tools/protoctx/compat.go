package protoctx

import (
	"fmt"
	"sort"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// ChangeSeverity represents the severity of a schema change.
type ChangeSeverity int

const (
	ChangeSeverityInfo ChangeSeverity = iota
	ChangeSeverityWarning
	ChangeSeverityBreaking
)

func (s ChangeSeverity) String() string {
	switch s {
	case ChangeSeverityInfo:
		return "INFO"
	case ChangeSeverityWarning:
		return "WARNING"
	case ChangeSeverityBreaking:
		return "BREAKING"
	default:
		return "UNKNOWN"
	}
}

// SchemaChange represents a detected schema change.
type SchemaChange struct {
	Severity ChangeSeverity `json:"severity"`
	Path     string         `json:"path"`      // FQN of the affected element
	Code     string         `json:"code"`      // Change type code
	Message  string         `json:"message"`   // Human-readable description
	OldValue string         `json:"old_value,omitempty"`
	NewValue string         `json:"new_value,omitempty"`
}

// CompatibilityReport contains the results of schema compatibility check.
type CompatibilityReport struct {
	Breaking []SchemaChange `json:"breaking,omitempty"`
	Warnings []SchemaChange `json:"warnings,omitempty"`
	Infos    []SchemaChange `json:"infos,omitempty"`
}

// HasBreaking returns true if there are any breaking changes.
func (r *CompatibilityReport) HasBreaking() bool {
	return len(r.Breaking) > 0
}

// HasWarnings returns true if there are any warnings.
func (r *CompatibilityReport) HasWarnings() bool {
	return len(r.Warnings) > 0
}

// TotalChanges returns the total number of changes detected.
func (r *CompatibilityReport) TotalChanges() int {
	return len(r.Breaking) + len(r.Warnings) + len(r.Infos)
}

// CompatibilityRules defines rules for compatibility checking.
type CompatibilityRules struct {
	// CheckFieldRemoval: treat field removal as breaking
	CheckFieldRemoval bool

	// CheckTypeChange: treat field type change as breaking
	CheckTypeChange bool

	// CheckLabelChange: treat label change (optional → required) as breaking
	CheckLabelChange bool

	// CheckEnumValueRemoval: treat enum value removal as breaking
	CheckEnumValueRemoval bool

	// CheckServiceRemoval: treat service/method removal as breaking
	CheckServiceRemoval bool

	// IgnorePackages: packages to ignore during comparison
	IgnorePackages []string
}

// DefaultCompatibilityRules returns the default compatibility rules.
func DefaultCompatibilityRules() CompatibilityRules {
	return CompatibilityRules{
		CheckFieldRemoval:     true,
		CheckTypeChange:       true,
		CheckLabelChange:      true,
		CheckEnumValueRemoval: true,
		CheckServiceRemoval:   true,
		IgnorePackages:        []string{"google.protobuf", "google.api"},
	}
}

// CompareSchemas compares two contexts and reports compatibility issues.
//
// This is the main entry point for schema compatibility checking.
// It compares all messages, enums, and services between old and new contexts.
func (c *Context) CompareSchemas(oldCtx *Context, rules CompatibilityRules) (*CompatibilityReport, error) {
	report := &CompatibilityReport{}

	// Compare messages
	c.compareMessages(oldCtx, rules, report)

	// Compare enums
	c.compareEnums(oldCtx, rules, report)

	// Compare services
	c.compareServices(oldCtx, rules, report)

	return report, nil
}

// compareMessages compares all messages between old and new contexts.
func (c *Context) compareMessages(oldCtx *Context, rules CompatibilityRules, report *CompatibilityReport) {
	oldMessages := oldCtx.ListMessages()
	newMessages := c.ListMessages()

	oldMap := make(map[string]protoreflect.MessageDescriptor)
	for _, md := range oldMessages {
		pkg := string(md.ParentFile().Package())
		if isIgnoredPackage(pkg, rules.IgnorePackages) {
			continue
		}
		oldMap[string(md.FullName())] = md
	}

	newMap := make(map[string]protoreflect.MessageDescriptor)
	for _, md := range newMessages {
		pkg := string(md.ParentFile().Package())
		if isIgnoredPackage(pkg, rules.IgnorePackages) {
			continue
		}
		newMap[string(md.FullName())] = md
	}

	// Check for removed messages
	for fqn := range oldMap {
		if _, exists := newMap[fqn]; !exists {
			report.Breaking = append(report.Breaking, SchemaChange{
				Severity: ChangeSeverityBreaking,
				Path:     fqn,
				Code:     "MESSAGE_REMOVED",
				Message:  "Message was removed",
			})
		}
	}

	// Check for modified messages
	for fqn, newMd := range newMap {
		oldMd, exists := oldMap[fqn]
		if !exists {
			// New message - not breaking
			report.Infos = append(report.Infos, SchemaChange{
				Severity: ChangeSeverityInfo,
				Path:     fqn,
				Code:     "MESSAGE_ADDED",
				Message:  "New message added",
			})
			continue
		}

		// Compare fields
		compareFields(oldMd, newMd, rules, report)
	}
}

// compareFields compares fields between two message descriptors.
func compareFields(oldMd, newMd protoreflect.MessageDescriptor, rules CompatibilityRules, report *CompatibilityReport) {
	oldFields := make(map[int32]protoreflect.FieldDescriptor)
	for i := 0; i < oldMd.Fields().Len(); i++ {
		fd := oldMd.Fields().Get(i)
		oldFields[int32(fd.Number())] = fd
	}

	newFields := make(map[int32]protoreflect.FieldDescriptor)
	for i := 0; i < newMd.Fields().Len(); i++ {
		fd := newMd.Fields().Get(i)
		newFields[int32(fd.Number())] = fd
	}

	fqn := string(newMd.FullName())

	// Check for removed fields
	if rules.CheckFieldRemoval {
		for num, oldFd := range oldFields {
			if _, exists := newFields[num]; !exists {
				report.Breaking = append(report.Breaking, SchemaChange{
					Severity: ChangeSeverityBreaking,
					Path:     fmt.Sprintf("%s.%s", fqn, oldFd.Name()),
					Code:     "FIELD_REMOVED",
					Message:  "Field was removed",
					OldValue: fmt.Sprintf("%s (number: %d)", oldFd.Name(), num),
				})
			}
		}
	}

	// Check for modified fields
	if rules.CheckTypeChange || rules.CheckLabelChange {
		for num, newFd := range newFields {
			oldFd, exists := oldFields[num]
			if !exists {
				// New field - not breaking
				continue
			}

			// Check type change
			if rules.CheckTypeChange && oldFd.Kind() != newFd.Kind() {
				report.Breaking = append(report.Breaking, SchemaChange{
					Severity: ChangeSeverityBreaking,
					Path:     fmt.Sprintf("%s.%s", fqn, newFd.Name()),
					Code:     "FIELD_TYPE_CHANGED",
					Message:  "Field type changed",
					OldValue: oldFd.Kind().String(),
					NewValue: newFd.Kind().String(),
				})
			}

			// Check label change (cardinality)
			if rules.CheckLabelChange && oldFd.Cardinality() != newFd.Cardinality() {
				report.Breaking = append(report.Breaking, SchemaChange{
					Severity: ChangeSeverityBreaking,
					Path:     fmt.Sprintf("%s.%s", fqn, newFd.Name()),
					Code:     "FIELD_LABEL_CHANGED",
					Message:  "Field label changed",
					OldValue: oldFd.Cardinality().String(),
					NewValue: newFd.Cardinality().String(),
				})
			}

			// Check if name changed for same number
			if string(oldFd.Name()) != string(newFd.Name()) {
				report.Warnings = append(report.Warnings, SchemaChange{
					Severity: ChangeSeverityWarning,
					Path:     fmt.Sprintf("%s (field %d)", fqn, num),
					Code:     "FIELD_NAME_CHANGED",
					Message:  "Field name changed (same number)",
					OldValue: string(oldFd.Name()),
					NewValue: string(newFd.Name()),
				})
			}
		}
	}
}

// compareEnums compares all enums between old and new contexts.
func (c *Context) compareEnums(oldCtx *Context, rules CompatibilityRules, report *CompatibilityReport) {
	oldEnums := oldCtx.ListEnums()
	newEnums := c.ListEnums()

	oldMap := make(map[string]protoreflect.EnumDescriptor)
	for _, ed := range oldEnums {
		pkg := string(ed.ParentFile().Package())
		if isIgnoredPackage(pkg, rules.IgnorePackages) {
			continue
		}
		oldMap[string(ed.FullName())] = ed
	}

	newMap := make(map[string]protoreflect.EnumDescriptor)
	for _, ed := range newEnums {
		pkg := string(ed.ParentFile().Package())
		if isIgnoredPackage(pkg, rules.IgnorePackages) {
			continue
		}
		newMap[string(ed.FullName())] = ed
	}

	// Check for removed enums
	for fqn := range oldMap {
		if _, exists := newMap[fqn]; !exists {
			report.Breaking = append(report.Breaking, SchemaChange{
				Severity: ChangeSeverityBreaking,
				Path:     fqn,
				Code:     "ENUM_REMOVED",
				Message:  "Enum was removed",
			})
		}
	}

	// Check for modified enums
	if rules.CheckEnumValueRemoval {
		for fqn, newEd := range newMap {
			oldEd, exists := oldMap[fqn]
			if !exists {
				// New enum - not breaking
				report.Infos = append(report.Infos, SchemaChange{
					Severity: ChangeSeverityInfo,
					Path:     fqn,
					Code:     "ENUM_ADDED",
					Message:  "New enum added",
				})
				continue
			}

			// Compare enum values
			compareEnumValues(oldEd, newEd, report)
		}
	}
}

// compareEnumValues compares enum values between two enum descriptors.
func compareEnumValues(oldEd, newEd protoreflect.EnumDescriptor, report *CompatibilityReport) {
	oldValues := make(map[int32]string)
	for i := 0; i < oldEd.Values().Len(); i++ {
		vd := oldEd.Values().Get(i)
		oldValues[int32(vd.Number())] = string(vd.Name())
	}

	newValues := make(map[int32]string)
	for i := 0; i < newEd.Values().Len(); i++ {
		vd := newEd.Values().Get(i)
		newValues[int32(vd.Number())] = string(vd.Name())
	}

	fqn := string(newEd.FullName())

	// Check for removed values
	for num, name := range oldValues {
		if _, exists := newValues[num]; !exists {
			report.Warnings = append(report.Warnings, SchemaChange{
				Severity: ChangeSeverityWarning,
				Path:     fmt.Sprintf("%s.%s", fqn, name),
				Code:     "ENUM_VALUE_REMOVED",
				Message:  "Enum value was removed",
				OldValue: fmt.Sprintf("%s = %d", name, num),
			})
		}
	}
}

// compareServices compares all services between old and new contexts.
func (c *Context) compareServices(oldCtx *Context, rules CompatibilityRules, report *CompatibilityReport) {
	if !rules.CheckServiceRemoval {
		return
	}

	oldServices := oldCtx.ListServices()
	newServices := c.ListServices()

	oldMap := make(map[string]protoreflect.ServiceDescriptor)
	for _, sd := range oldServices {
		pkg := string(sd.ParentFile().Package())
		if isIgnoredPackage(pkg, rules.IgnorePackages) {
			continue
		}
		oldMap[string(sd.FullName())] = sd
	}

	newMap := make(map[string]protoreflect.ServiceDescriptor)
	for _, sd := range newServices {
		pkg := string(sd.ParentFile().Package())
		if isIgnoredPackage(pkg, rules.IgnorePackages) {
			continue
		}
		newMap[string(sd.FullName())] = sd
	}

	// Check for removed services
	for fqn := range oldMap {
		if _, exists := newMap[fqn]; !exists {
			report.Breaking = append(report.Breaking, SchemaChange{
				Severity: ChangeSeverityBreaking,
				Path:     fqn,
				Code:     "SERVICE_REMOVED",
				Message:  "Service was removed",
			})
		}
	}

	// Check for modified services
	for fqn, newSd := range newMap {
		oldSd, exists := oldMap[fqn]
		if !exists {
			// New service - not breaking
			report.Infos = append(report.Infos, SchemaChange{
				Severity: ChangeSeverityInfo,
				Path:     fqn,
				Code:     "SERVICE_ADDED",
				Message:  "New service added",
			})
			continue
		}

		// Compare methods
		compareMethods(oldSd, newSd, report)
	}
}

// compareMethods compares methods between two service descriptors.
func compareMethods(oldSd, newSd protoreflect.ServiceDescriptor, report *CompatibilityReport) {
	oldMethods := make(map[string]protoreflect.MethodDescriptor)
	for i := 0; i < oldSd.Methods().Len(); i++ {
		md := oldSd.Methods().Get(i)
		oldMethods[string(md.Name())] = md
	}

	newMethods := make(map[string]protoreflect.MethodDescriptor)
	for i := 0; i < newSd.Methods().Len(); i++ {
		md := newSd.Methods().Get(i)
		newMethods[string(md.Name())] = md
	}

	fqn := string(newSd.FullName())

	// Check for removed methods
	for name := range oldMethods {
		if _, exists := newMethods[name]; !exists {
			report.Breaking = append(report.Breaking, SchemaChange{
				Severity: ChangeSeverityBreaking,
				Path:     fmt.Sprintf("%s.%s", fqn, name),
				Code:     "METHOD_REMOVED",
				Message:  "Method was removed",
			})
		}
	}

	// Check for modified methods
	for name, newMd := range newMethods {
		oldMd, exists := oldMethods[name]
		if !exists {
			// New method - not breaking
			continue
		}

		// Check input type change
		if string(oldMd.Input().FullName()) != string(newMd.Input().FullName()) {
			report.Breaking = append(report.Breaking, SchemaChange{
				Severity: ChangeSeverityBreaking,
				Path:     fmt.Sprintf("%s.%s", fqn, name),
				Code:     "METHOD_INPUT_CHANGED",
				Message:  "Method input type changed",
				OldValue: string(oldMd.Input().FullName()),
				NewValue: string(newMd.Input().FullName()),
			})
		}

		// Check output type change
		if string(oldMd.Output().FullName()) != string(newMd.Output().FullName()) {
			report.Breaking = append(report.Breaking, SchemaChange{
				Severity: ChangeSeverityBreaking,
				Path:     fmt.Sprintf("%s.%s", fqn, name),
				Code:     "METHOD_OUTPUT_CHANGED",
				Message:  "Method output type changed",
				OldValue: string(oldMd.Output().FullName()),
				NewValue: string(newMd.Output().FullName()),
			})
		}

		// Check streaming mode change
		if oldMd.IsStreamingClient() != newMd.IsStreamingClient() ||
			oldMd.IsStreamingServer() != newMd.IsStreamingServer() {
			report.Breaking = append(report.Breaking, SchemaChange{
				Severity: ChangeSeverityBreaking,
				Path:     fmt.Sprintf("%s.%s", fqn, name),
				Code:     "METHOD_STREAMING_CHANGED",
				Message:  "Method streaming mode changed",
			})
		}
	}
}

// isIgnoredPackage checks if a package should be ignored.
func isIgnoredPackage(pkg string, ignoreList []string) bool {
	for _, ignored := range ignoreList {
		if pkg == ignored {
			return true
		}
		// Check if pkg starts with ignored prefix (e.g., "google." matches "google.protobuf")
		if len(pkg) > len(ignored) && pkg[:len(ignored)+1] == ignored+"." {
			return true
		}
	}
	return false
}

// SortChanges sorts changes by severity (breaking first) and then by path.
func (r *CompatibilityReport) SortChanges() {
	sortChanges := func(changes []SchemaChange) {
		sort.Slice(changes, func(i, j int) bool {
			return changes[i].Path < changes[j].Path
		})
	}

	sortChanges(r.Breaking)
	sortChanges(r.Warnings)
	sortChanges(r.Infos)
}
