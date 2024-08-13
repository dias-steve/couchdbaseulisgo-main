package entities

type GrafanaCredentialsEntity struct {
	CredentialId    string `json:"credential_id"`
	UserId          string `json:"user_id"`
	UsernameGrafana string `json:"username_grafana"`
	PasswordGrafana string `json:"password_grafana"`
}

type GrafanaCredentialsDto struct {
	CredentialId    string `json:"credential_id"`
	UserId          string `json:"user_id"`
	UsernameGrafana string `json:"username_grafana"`
	PasswordGrafana string `json:"password_grafana"`
}
