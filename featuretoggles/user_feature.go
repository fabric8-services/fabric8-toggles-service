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
	return []interface{}{f.Name, f.Description, f.Enabled, f.EnablementLevel, f.UserEnabled}
}

// ZeroUserFeature to check if a feature is empty
var ZeroUserFeature UserFeature

// ByName implements sort.Interface for []UserFeature based on
// the Name field.
type ByName []UserFeature

func (s ByName) Len() int           { return len(s) }
func (s ByName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByName) Less(i, j int) bool { return s[i].Name < s[j].Name }
