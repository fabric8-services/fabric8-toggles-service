package featuretoggles

// UserFeature a feature with the user enablement
type UserFeature struct {
	Name            string
	Description     string
	Enabled         bool
	EnablementLevel string
	UserEnabled     bool
}

// GetETagData returns the field values to use to generate the ETag
func (f UserFeature) GetETagData() []interface{} {
	return []interface{}{f.Name, f.Enabled, f.EnablementLevel, f.UserEnabled}
}

// ZeroUserFeature to check if a feature is empty
var ZeroUserFeature UserFeature
