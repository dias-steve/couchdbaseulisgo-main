package entities

type GrafanaCredentialsEntity struct {
	CredentialId    string `json:"credential_id"`
	UserId          string `json:"user_id"`
	UsernameGrafana string `json:"username_grafana"`
	PasswordGrafana string `json:"password_grafana"`
}

// Entitie shown out of the router
// For example, the password is not shown here for security reasons
type GrafanaCredentialsDto struct {
	CredentialId    string `json:"credential_id"`
	UserId          string `json:"user_id"`
	UsernameGrafana string `json:"username_grafana"`
}
